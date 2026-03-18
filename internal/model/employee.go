package model

import "time"

type Employee struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DepartmentID uint      `gorm:"not null;index" json:"department_id"`
	FullName     string    `gorm:"not null;size:200" json:"full_name"`
	Position     string    `gorm:"not null;size:200" json:"position"`
	HiredAt      *string   `json:"hired_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}
