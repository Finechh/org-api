package repository

import (
	"context"
	"errors"

	"org-api/internal/model"

	"gorm.io/gorm"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) error
	GetByID(ctx context.Context, id uint) (*model.Employee, error)
	GetByDepartmentID(ctx context.Context, deptID uint) ([]model.Employee, error)
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, emp *model.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *employeeRepository) GetByDepartmentID(ctx context.Context, deptID uint) ([]model.Employee, error) {
	var employees []model.Employee
	if err := r.db.WithContext(ctx).Where("department_id = ?", deptID).Order("created_at asc").Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) GetByID(ctx context.Context, id uint) (*model.Employee, error) {
	var emp model.Employee
	if err := r.db.WithContext(ctx).First(&emp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}
