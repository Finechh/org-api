package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"org-api/internal/model"
	"org-api/internal/repository"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrBadRequest = errors.New("bad request")
)

type DepartmentService interface {
	Create(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error)
	Update(ctx context.Context, id uint, req model.UpdateDepartmentRequest) (*model.Department, error)
	Delete(ctx context.Context, id uint, mode string, reassignTo *uint) error
	GetTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*model.DepartmentResponse, error)
}

type departmentService struct {
	deptRepo repository.DepartmentRepository
	empRepo  repository.EmployeeRepository
}

func NewDepartmentService(dr repository.DepartmentRepository, er repository.EmployeeRepository) DepartmentService {
	return &departmentService{deptRepo: dr, empRepo: er}
}

func (s *departmentService) Create(ctx context.Context, req model.CreateDepartmentRequest) (*model.Department, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 200 {
		return nil, fmt.Errorf("%w: name must be 1-200 characters", ErrBadRequest)
	}

	if req.ParentID != nil {
		if _, err := s.deptRepo.GetByID(ctx, *req.ParentID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, fmt.Errorf("%w: parent department not found", ErrNotFound)
			}
			return nil, err
		}
	}

	exists, err := s.deptRepo.ExistsWithNameUnderParent(ctx, name, req.ParentID, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: department with this name already exists under the same parent", ErrConflict)
	}

	dept := &model.Department{Name: name, ParentID: req.ParentID}
	if err := s.deptRepo.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) Update(ctx context.Context, id uint, req model.UpdateDepartmentRequest) (*model.Department, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > 200 {
			return nil, fmt.Errorf("%w: name must be 1-200 characters", ErrBadRequest)
		}
		dept.Name = name
	}

	if req.ParentID != nil {
		newParentID := req.ParentID
		if *newParentID == id {
			return nil, fmt.Errorf("%w: department cannot be its own parent", ErrConflict)
		}
		if err := s.checkCycle(ctx, id, *newParentID); err != nil {
			return nil, err
		}
		if _, err := s.deptRepo.GetByID(ctx, *newParentID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, fmt.Errorf("%w: parent department not found", ErrNotFound)
			}
			return nil, err
		}
		dept.ParentID = newParentID
	}

	exists, err := s.deptRepo.ExistsWithNameUnderParent(ctx, dept.Name, dept.ParentID, &id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: department with this name already exists under the same parent", ErrConflict)
	}

	if err := s.deptRepo.Update(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) Delete(ctx context.Context, id uint, mode string, reassignTo *uint) error {
	if _, err := s.deptRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	switch mode {
	case "cascade":
		return s.deptRepo.CascadeDelete(ctx, id)
	case "reassign":
		if reassignTo == nil {
			return fmt.Errorf("%w: reassign_to_department_id is required for reassign mode", ErrBadRequest)
		}
		if _, err := s.deptRepo.GetByID(ctx, *reassignTo); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf("%w: target department not found", ErrNotFound)
			}
			return err
		}
		return s.deptRepo.ReassignDelete(ctx, id, *reassignTo)
	default:
		return fmt.Errorf("%w: mode must be 'cascade' or 'reassign'", ErrBadRequest)
	}
}

func (s *departmentService) GetTree(ctx context.Context, id uint, depth int, includeEmployees bool) (*model.DepartmentResponse, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 5 {
		depth = 5
	}
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.buildResponse(ctx, dept, depth, includeEmployees)
}

func (s *departmentService) buildResponse(ctx context.Context, dept *model.Department, depth int, includeEmployees bool) (*model.DepartmentResponse, error) {
	resp := &model.DepartmentResponse{
		ID:        dept.ID,
		Name:      dept.Name,
		ParentID:  dept.ParentID,
		CreatedAt: dept.CreatedAt,
	}

	if includeEmployees {
		employees, err := s.empRepo.GetByDepartmentID(ctx, dept.ID)
		if err != nil {
			return nil, err
		}
		resp.Employees = employees
	}

	if depth > 1 {
		children, err := s.deptRepo.GetChildren(ctx, dept.ID)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			childResp, err := s.buildResponse(ctx, &child, depth-1, includeEmployees)
			if err != nil {
				return nil, err
			}
			resp.Children = append(resp.Children, *childResp)
		}
	}
	return resp, nil
}

func (s *departmentService) checkCycle(ctx context.Context, id uint, newParentID uint) error {
	descendantIDs, err := s.deptRepo.GetAllDescendantIDs(ctx, id)
	if err != nil {
		return err
	}
	for _, did := range descendantIDs {
		if did == newParentID {
			return fmt.Errorf("%w: moving department would create a cycle", ErrConflict)
		}
	}
	return nil
}
