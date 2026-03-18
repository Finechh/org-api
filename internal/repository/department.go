package repository

import (
	"context"
	"errors"

	"org-api/internal/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type DepartmentRepository interface {
	Create(ctx context.Context, dept *model.Department) error
	GetByID(ctx context.Context, id uint) (*model.Department, error)
	Update(ctx context.Context, dept *model.Department) error
	Delete(ctx context.Context, id uint) error
	GetChildren(ctx context.Context, parentID uint) ([]model.Department, error)
	GetAllDescendantIDs(ctx context.Context, rootID uint) ([]uint, error)
	ExistsWithNameUnderParent(ctx context.Context, name string, parentID *uint, excludeID *uint) (bool, error)
	CascadeDelete(ctx context.Context, id uint) error
	ReassignDelete(ctx context.Context, id uint, targetID uint) error
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *departmentRepository) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	var dept model.Department
	if err := r.db.WithContext(ctx).First(&dept, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) Update(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *departmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Department{}, id).Error
}

func (r *departmentRepository) GetChildren(ctx context.Context, parentID uint) ([]model.Department, error) {
	var children []model.Department
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}

func (r *departmentRepository) GetAllDescendantIDs(ctx context.Context, rootID uint) ([]uint, error) {
	var ids []uint
	queue := []uint{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		ids = append(ids, cur)
		var childIDs []uint
		if err := r.db.WithContext(ctx).Model(&model.Department{}).Where("parent_id = ?", cur).Pluck("id", &childIDs).Error; err != nil {
			return nil, err
		}
		queue = append(queue, childIDs...)
	}
	return ids, nil
}

func (r *departmentRepository) ExistsWithNameUnderParent(ctx context.Context, name string, parentID *uint, excludeID *uint) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Department{}).Where("name = ?", name)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *departmentRepository) CascadeDelete(ctx context.Context, id uint) error {
	ids, err := r.GetAllDescendantIDs(ctx, id)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("department_id IN ?", ids).Delete(&model.Employee{}).Error; err != nil {
			return err
		}
		for i := len(ids) - 1; i >= 0; i-- {
			if err := tx.Delete(&model.Department{}, ids[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *departmentRepository) ReassignDelete(ctx context.Context, id uint, targetID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Employee{}).Where("department_id = ?", id).Update("department_id", targetID).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Department{}, id).Error
	})
}
