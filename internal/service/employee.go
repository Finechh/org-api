package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"org-api/internal/model"
	"org-api/internal/repository"
)

type EmployeeService interface {
	Create(ctx context.Context, deptID uint, req model.CreateEmployeeRequest) (*model.Employee, error)
}

type employeeService struct {
	deptRepo repository.DepartmentRepository
	empRepo  repository.EmployeeRepository
}

func NewEmployeeService(dr repository.DepartmentRepository, er repository.EmployeeRepository) EmployeeService {
	return &employeeService{deptRepo: dr, empRepo: er}
}

func (s *employeeService) Create(ctx context.Context, deptID uint, req model.CreateEmployeeRequest) (*model.Employee, error) {
	if _, err := s.deptRepo.GetByID(ctx, deptID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: department not found", ErrNotFound)
		}
		return nil, err
	}

	fullName := strings.TrimSpace(req.FullName)
	position := strings.TrimSpace(req.Position)

	if fullName == "" || len(fullName) > 200 {
		return nil, fmt.Errorf("%w: full_name must be 1-200 characters", ErrBadRequest)
	}
	if position == "" || len(position) > 200 {
		return nil, fmt.Errorf("%w: position must be 1-200 characters", ErrBadRequest)
	}

	emp := &model.Employee{
		DepartmentID: deptID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      req.HiredAt,
	}
	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}
	return emp, nil
}
