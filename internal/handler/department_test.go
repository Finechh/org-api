package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"org-api/internal/handler"
	"org-api/internal/model"
	"org-api/internal/service"
	"org-api/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDepartmentService struct {
	createFn  func(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error)
	updateFn  func(ctx context.Context, id uint, req model.UpdateDepartmentRequest) (*model.Department, error)
	deleteFn  func(ctx context.Context, id uint, mode string, reassignTo *uint) error
	getTreeFn func(ctx context.Context, id uint, depth int, includeEmployees bool) (*model.DepartmentResponse, error)
}

func (m *mockDepartmentService) Create(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error) {
	return m.createFn(ctx, req)
}
func (m *mockDepartmentService) Update(ctx context.Context, id uint, req model.UpdateDepartmentRequest) (*model.Department, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockDepartmentService) Delete(ctx context.Context, id uint, mode string, reassignTo *uint) error {
	return m.deleteFn(ctx, id, mode, reassignTo)
}
func (m *mockDepartmentService) GetTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*model.DepartmentResponse, error) {
	return m.getTreeFn(ctx, id, depth, includeEmployees)
}

type mockEmployeeService struct {
	createFn func(ctx context.Context, deptID uint, req model.CreateEmployeeRequest) (*model.Employee, error)
}

func (m *mockEmployeeService) Create(ctx context.Context, deptID uint, req model.CreateEmployeeRequest) (*model.Employee, error) {
	return m.createFn(ctx, deptID, req)
}

func setupMux(deptSvc service.DepartmentService, empSvc service.EmployeeService) http.Handler {
	log := logger.New()
	deptH := handler.NewDepartmentHandler(deptSvc, log)
	empH := handler.NewEmployeeHandler(empSvc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments/", deptH.Create)
	mux.HandleFunc("GET /departments/{id}", deptH.Get)
	mux.HandleFunc("PATCH /departments/{id}", deptH.Update)
	mux.HandleFunc("DELETE /departments/{id}", deptH.Delete)
	mux.HandleFunc("POST /departments/{id}/employees/", empH.Create)
	return mux
}

func TestCreateDepartment_Success(t *testing.T) {
	deptSvc := &mockDepartmentService{
		createFn: func(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error) {
			return &model.Department{ID: 1, Name: req.Name}, nil
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"Engineering"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var dept model.Department
	require.NoError(t, json.NewDecoder(w.Body).Decode(&dept))
	assert.Equal(t, "Engineering", dept.Name)
}

func TestCreateDepartment_EmptyName(t *testing.T) {
	deptSvc := &mockDepartmentService{
		createFn: func(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error) {
			return nil, service.ErrBadRequest
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDepartment_DuplicateName(t *testing.T) {
	deptSvc := &mockDepartmentService{
		createFn: func(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error) {
			return nil, service.ErrConflict
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"HR"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateDepartment_ParentNotFound(t *testing.T) {
	deptSvc := &mockDepartmentService{
		createFn: func(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error) {
			return nil, service.ErrNotFound
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"Backend","parent_id":999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDepartment_WithEmployeesAndChildren(t *testing.T) {
	parentID := uint(1)
	deptSvc := &mockDepartmentService{
		getTreeFn: func(ctx context.Context, id uint, depth int, includeEmployees bool) (*model.DepartmentResponse, error) {
			return &model.DepartmentResponse{
				ID:   1,
				Name: "Root",
				Children: []model.DepartmentResponse{
					{ID: 2, Name: "Child", ParentID: &parentID},
				},
				Employees: []model.Employee{
					{ID: 1, FullName: "Alice Smith", Position: "Engineer"},
				},
			}, nil
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	req := httptest.NewRequest(http.MethodGet, "/departments/1?depth=2&include_employees=true", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp model.DepartmentResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "Root", resp.Name)
	assert.Len(t, resp.Children, 1)
	assert.Len(t, resp.Employees, 1)
}

func TestUpdateDepartment_CycleDetection(t *testing.T) {
	deptSvc := &mockDepartmentService{
		updateFn: func(ctx context.Context, id uint, req model.UpdateDepartmentRequest) (*model.Department, error) {
			return nil, service.ErrConflict
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	patchBody, _ := json.Marshal(map[string]any{"parent_id": 3})
	req := httptest.NewRequest(http.MethodPatch, "/departments/1", bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDeleteDepartment_Cascade(t *testing.T) {
	deptSvc := &mockDepartmentService{
		deleteFn: func(ctx context.Context, id uint, mode string, reassignTo *uint) error {
			return nil
		},
	}
	mux := setupMux(deptSvc, &mockEmployeeService{})

	req := httptest.NewRequest(http.MethodDelete, "/departments/1?mode=cascade", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCreateEmployee_DeptNotFound(t *testing.T) {
	empSvc := &mockEmployeeService{
		createFn: func(ctx context.Context, deptID uint, req model.CreateEmployeeRequest) (*model.Employee, error) {
			return nil, service.ErrNotFound
		},
	}
	mux := setupMux(&mockDepartmentService{}, empSvc)

	req := httptest.NewRequest(http.MethodPost, "/departments/999/employees/",
		bytes.NewBufferString(`{"full_name":"Ghost","position":"None"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
