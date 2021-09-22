package migrations

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func CreateTargetClusters(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202103011500",
			Migrate: func(gdb *gorm.DB) error {
				type Image struct {
					Model
				}

				type TargetCluster struct {
					Model
					Name     string    `gorm:"column:name;type:text"`
					Token    string    `gorm:"column:token;type:text"`
					Images   []Image   `gorm:"many2many:target_cluster_images;foreignKey:ID;joinForeignKey:ID"`
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}

				type Deployment struct {
					Model
					TargetClusterName string    `gorm:"column:target_cluster;type:text"`
					TargetClusterID   uuid.UUID `gorm:"column:target_cluster_id;type:uuid"`
					DomainID          uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}

				type Environment struct {
					Model
					Name           string          `gorm:"column:name;type:text;not null"`
					TargetClusters []TargetCluster `gorm:"many2many:environment_target_clusters;foreignKey:ID;joinForeignKey:ID"`
					Clusters       string          `gorm:"column:clusters;type:text"`
					DomainID       uuid.UUID       `gorm:"column:domain_id;type:uuid;not null;index"`
				}

				// Create target_clusters table
				if err := gdb.AutoMigrate(&TargetCluster{}).Error; err != nil {
					return err
				}
				// Add target_cluster_id to Deployments
				if err := gdb.AutoMigrate(&Deployment{}).Error; err != nil {
					return err
				}
				// Add environment_target_clusters join table
				if err := gdb.AutoMigrate(&Environment{}).Error; err != nil {
					return err
				}

				// Pull domain_id from recent deployment
				var deployments []Deployment
				gdb.Order("created_at desc").Limit(1).Find(&deployments)
				if len(deployments) == 0 {
					fmt.Println("202103011500_create_target_clusters: Could not find a latest deployment to pull domain_id. Skipping object relation updates.")
				} else {
					domainID := deployments[0].DomainID

					//====================================================================
					// Create TargetCluster objects
					//====================================================================
					stellaK8sBlue := TargetCluster{
						Name:     "stella-k8s-blue",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaK8sBlue); resp.Error != nil {
						return resp.Error
					}

					stellaK8sRed := TargetCluster{
						Name:     "stella-k8s-red",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaK8sRed); resp.Error != nil {
						return resp.Error
					}

					stellaK8sBlack := TargetCluster{
						Name:     "stella-k8s-black",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaK8sBlack); resp.Error != nil {
						return resp.Error
					}

					stellaK8sGreen := TargetCluster{
						Name:     "stella-k8s-green",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaK8sGreen); resp.Error != nil {
						return resp.Error
					}

					stellaK8sSilver := TargetCluster{
						Name:     "stella-k8s-silver",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaK8sSilver); resp.Error != nil {
						return resp.Error
					}

					stellaWestusAzure1Green := TargetCluster{
						Name:     "stella-westus-azure1-green",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaWestusAzure1Green); resp.Error != nil {
						return resp.Error
					}

					stellaWestusAzure1Blue := TargetCluster{
						Name:     "stella-westus-azure1-blue",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaWestusAzure1Blue); resp.Error != nil {
						return resp.Error
					}

					stellaWestusStudioGreen := TargetCluster{
						Name:     "stella-westus-studio-green",
						DomainID: domainID,
					}
					if resp := gdb.Create(&stellaWestusStudioGreen); resp.Error != nil {
						return resp.Error
					}

					dwaWestusAzure1Teal := TargetCluster{
						Name:     "aks-dwa-westus-azure1-teal",
						DomainID: domainID,
					}
					if resp := gdb.Create(&dwaWestusAzure1Teal); resp.Error != nil {
						return resp.Error
					}

					dwaEastus2Azure1Teal := TargetCluster{
						Name:     "aks-dwa-eastus2-azure1-teal",
						DomainID: domainID,
					}
					if resp := gdb.Create(&dwaEastus2Azure1Teal); resp.Error != nil {
						return resp.Error
					}

					dwaWestusStudioTeal := TargetCluster{
						Name:     "aks-dwa-westus-studio-teal",
						DomainID: domainID,
					}
					if resp := gdb.Create(&dwaWestusStudioTeal); resp.Error != nil {
						return resp.Error
					}

					dwaEastus2StudioTeal := TargetCluster{
						Name:     "aks-dwa-eastus2-studio-teal",
						DomainID: domainID,
					}
					if resp := gdb.Create(&dwaEastus2StudioTeal); resp.Error != nil {
						return resp.Error
					}

					//====================================================================
					// Update Deployment target cluster fields
					//====================================================================
					var k8sBlueDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-k8s-blue").Find(&k8sBlueDeployments)
					for _, deployment := range k8sBlueDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaK8sBlue.ID,
							"target_cluster":    stellaK8sBlue.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var k8sRedDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-k8s-red").Find(&k8sRedDeployments)
					for _, deployment := range k8sRedDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaK8sRed.ID,
							"target_cluster":    stellaK8sRed.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var k8sBlackDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-k8s-black").Find(&k8sBlackDeployments)
					for _, deployment := range k8sBlackDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaK8sBlack.ID,
							"target_cluster":    stellaK8sBlack.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var k8sGreenDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-k8s-green").Find(&k8sGreenDeployments)
					for _, deployment := range k8sGreenDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaK8sGreen.ID,
							"target_cluster":    stellaK8sGreen.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var k8sSilverDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-k8s-silver").Find(&k8sSilverDeployments)
					for _, deployment := range k8sSilverDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaK8sSilver.ID,
							"target_cluster":    stellaK8sSilver.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var azure1GreenDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-westus-azure1-green").Find(&azure1GreenDeployments)
					for _, deployment := range azure1GreenDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaWestusAzure1Green.ID,
							"target_cluster":    stellaWestusAzure1Green.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var azure1BlueDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-westus-azure1-blue").Find(&azure1BlueDeployments)
					for _, deployment := range azure1BlueDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaWestusAzure1Blue.ID,
							"target_cluster":    stellaWestusAzure1Blue.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var studioGreenDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "stella-westus-studio-green").Find(&studioGreenDeployments)
					for _, deployment := range studioGreenDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": stellaWestusStudioGreen.ID,
							"target_cluster":    stellaWestusStudioGreen.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var eastus2Azure1TealDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "aks-dwa-eastus2-azure1-teal").Find(&eastus2Azure1TealDeployments)
					for _, deployment := range eastus2Azure1TealDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": dwaEastus2Azure1Teal.ID,
							"target_cluster":    dwaEastus2Azure1Teal.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var westusAzure1TealDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "aks-dwa-westus-azure1-teal").Find(&westusAzure1TealDeployments)
					for _, deployment := range westusAzure1TealDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": dwaWestusAzure1Teal.ID,
							"target_cluster":    dwaWestusAzure1Teal.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var westusStudioTealDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "aks-dwa-westus-studio-teal").Find(&westusStudioTealDeployments)
					for _, deployment := range westusStudioTealDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": dwaWestusStudioTeal.ID,
							"target_cluster":    dwaWestusStudioTeal.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					var eastus2StudioTealDeployments []Deployment
					gdb.Unscoped().Where("target_cluster = ?", "aks-dwa-eastus2-studio-teal").Find(&eastus2StudioTealDeployments)
					for _, deployment := range eastus2StudioTealDeployments {
						err := gdb.Unscoped().Model(&deployment).Updates(map[string]interface{}{
							"target_cluster_id": dwaEastus2StudioTeal.ID,
							"target_cluster":    dwaEastus2StudioTeal.Name,
						}).Error
						if err != nil {
							return err
						}
					}

					//====================================================================
					// Update Environment target clusters
					//====================================================================
					var environments []Environment
					gdb.Unscoped().Find(&environments)
					for _, environment := range environments {
						var targetClusters []TargetCluster

						if strings.Contains(environment.Clusters, "stella-k8s-blue") {
							targetClusters = append(targetClusters, stellaK8sBlue)
						}
						if strings.Contains(environment.Clusters, "stella-k8s-red") {
							targetClusters = append(targetClusters, stellaK8sRed)
						}
						if strings.Contains(environment.Clusters, "stella-k8s-black") {
							targetClusters = append(targetClusters, stellaK8sBlack)
						}
						if strings.Contains(environment.Clusters, "stella-k8s-green") {
							targetClusters = append(targetClusters, stellaK8sGreen)
						}
						if strings.Contains(environment.Clusters, "stella-k8s-silver") {
							targetClusters = append(targetClusters, stellaK8sSilver)
						}
						if strings.Contains(environment.Clusters, "stella-westus-azure1-green") {
							targetClusters = append(targetClusters, stellaWestusAzure1Green)
						}
						if strings.Contains(environment.Clusters, "stella-westus-azure1-blue") {
							targetClusters = append(targetClusters, stellaWestusAzure1Blue)
						}
						if strings.Contains(environment.Clusters, "stella-westus-studio-green") {
							targetClusters = append(targetClusters, stellaWestusStudioGreen)
						}
						if strings.Contains(environment.Clusters, "aks-dwa-westus-azure1-teal") {
							targetClusters = append(targetClusters, dwaWestusAzure1Teal)
						}
						if strings.Contains(environment.Clusters, "aks-dwa-eastus2-azure1-teal") {
							targetClusters = append(targetClusters, dwaEastus2Azure1Teal)
						}
						if strings.Contains(environment.Clusters, "aks-dwa-westus-studio-teal") {
							targetClusters = append(targetClusters, dwaWestusStudioTeal)
						}
						if strings.Contains(environment.Clusters, "aks-dwa-eastus2-studio-teal") {
							targetClusters = append(targetClusters, dwaEastus2StudioTeal)
						}

						environment.TargetClusters = targetClusters
						if resp := gdb.Unscoped().Save(&environment); resp.Error != nil {
							return resp.Error
						}
					}
				}

				// Drop old clusters col from Environments
				if err := gdb.Model(&Environment{}).DropColumn("clusters").Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				// Drop target_clusters table and target_clusters_images table
				if err := gdb.DropTable("target_clusters").Error; err != nil {
					return err
				}
				if err := gdb.DropTable("target_cluster_images").Error; err != nil {
					return err
				}

				// Drop target_cluster_id col from Deployments
				type Deployment struct{}
				if err := gdb.Model(&Deployment{}).DropColumn("target_cluster_id").Error; err != nil {
					return err
				}

				// Create clusters col in Environments and drop environment_target_clusters table
				type Environment struct {
					Clusters string `gorm:"column:clusters;type:text"`
				}
				if err := gdb.AutoMigrate(&Environment{}).Error; err != nil {
					return err
				}
				return gdb.DropTable("environment_target_clusters").Error
			},
		},
	})

	return m.Migrate()
}
