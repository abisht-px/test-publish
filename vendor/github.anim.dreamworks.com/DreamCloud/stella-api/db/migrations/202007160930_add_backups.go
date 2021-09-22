package migrations

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddBackups(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202007160930",
			Migrate: func(gdb *gorm.DB) error {
				type Backup struct {
					Model
					FileSize       uint      `gorm:"column:file_size;type:integer;not null"`
					BackupTime     time.Time `gorm:"column:backup_time;type:timestamp;not null"`
					DeploymentID   uuid.UUID `gorm:"column:deployment_id;type:uuid;not null;index"`
					DeploymentName string    `gorm:"column:deployment;type:text;not null"`
					State          string    `json:"state" mapstructure:"-"`
					DomainID       uuid.UUID `json:"domain_id"`
					ProjectID      uuid.UUID `json:"project_id"`
					UserID         uuid.UUID `json:"user_id"`
					UserLogin      string    `json:"user_login"`
				}
				if err := gdb.AutoMigrate(&Backup{}).Error; err != nil {
					return err
				}
				return gdb.DropTable("snapshots").Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Snapshot struct {
					Model
					FileSize       uint      `gorm:"column:file_size;type:integer;not null"`
					SnapshotTime   time.Time `gorm:"column:snapshot_time;type:timestamp;not null"`
					DeploymentID   uuid.UUID `gorm:"column:deployment_id;type:uuid;not null;index"`
					DeploymentName string    `gorm:"column:deployment;type:text;not null"`
					DomainID       uuid.UUID `json:"domain_id"`
					ProjectID      uuid.UUID `json:"project_id"`
					UserID         uuid.UUID `json:"user_id"`
					UserLogin      string    `json:"user_login"`
				}
				if err := gdb.AutoMigrate(&Snapshot{}).Error; err != nil {
					return err
				}
				return gdb.DropTable("backups").Error
			},
		},
	})

	return m.Migrate()
}
