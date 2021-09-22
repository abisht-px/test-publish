package migrations

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Model struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primary_key;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;"`
}

type TaskModel struct {
	ID        uint       `gorm:"primary_key"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;"`
}

// BeforeCreate sets a new UUID on model creation.
// This hook is auto called by GORM on creation of the Model struct.
func (model *Model) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New())
}
