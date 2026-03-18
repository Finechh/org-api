package model

import "time"

type Department struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `gorm:"not null;size:200" json:"name"`
	ParentID  *uint       `gorm:"index" json:"parent_id"`
	CreatedAt time.Time   `json:"created_at"`
	Parent    *Department `gorm:"foreignKey:ParentID" json:"-"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID" json:"employees,omitempty"`
}

type DepartmentResponse struct {
	ID        uint                  `json:"id"`
	Name      string                `json:"name"`
	ParentID  *uint                 `json:"parent_id"`
	CreatedAt time.Time             `json:"created_at"`
	Employees []Employee            `json:"employees,omitempty"`
	Children  []DepartmentResponse  `json:"children,omitempty"`
}

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateDepartmentRequest struct {
	Name     *string `json:"name"`
	ParentID *uint   `json:"parent_id"`
}

type DeleteDepartmentQuery struct {
	Mode                   string `json:"mode"`
	ReassignToDepartmentID *uint  `json:"reassign_to_department_id"`
}
