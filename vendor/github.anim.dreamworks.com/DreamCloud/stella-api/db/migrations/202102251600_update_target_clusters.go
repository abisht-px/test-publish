package migrations

import (
	"strings"

	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func UpdateTargetClusters(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202102251600",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Model
					TargetClusters       []string `gorm:"-"`
					TargetClustersString string   `gorm:"column:target_clusters;type:text"`
					TargetCluster        string   `gorm:"column:target_cluster;type:text"`
				}

				// Add new TargetCluster field
				if err := gdb.AutoMigrate(&Deployment{}).Error; err != nil {
					return err
				}

				// Set TargetCluster to first element in TargetClusters Array
				var deployments []Deployment
				gdb.Unscoped().Find(&deployments)
				for _, deployment := range deployments {
					// Skip if the deployment doesn't have any target clusters
					if len(deployment.TargetClustersString) == 0 {
						continue
					}
					// Deserialize target cluster string into array
					deployment.TargetClusters = strings.Split(deployment.TargetClustersString, ";;")
					if len(deployment.TargetClusters) == 0 {
						continue
					}
					// Set new TargetCluster field to first element in array
					resp := gdb.Model(&deployment).Update(map[string]interface{}{
						"target_cluster": deployment.TargetClusters[0],
					})
					if resp.Error != nil {
						return resp.Error
					}
				}

				// Drop target_clusters field on deployments
				return gdb.Model(&Deployment{}).DropColumn("target_clusters").Error
			},
		},
	})

	return m.Migrate()
}
