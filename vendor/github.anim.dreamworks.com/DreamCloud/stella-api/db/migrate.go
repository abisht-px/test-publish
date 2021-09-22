package db

import (
	"github.anim.dreamworks.com/golang/logging"
	"github.com/jinzhu/gorm"

	"github.anim.dreamworks.com/DreamCloud/stella-api/db/migrations"
	"github.anim.dreamworks.com/DreamCloud/stella-api/db/seeding"
)

func runMigrations(db *gorm.DB) {
	migrateSchema(db)
	seedData(db)
}

func migrateSchema(db *gorm.DB) {
	var err error

	// 201907011638_create_resources.go
	if err = migrations.CreateResources(db); err != nil {
		logging.Fatalf("Database migration (201907011638_create_resources) failed. %v", err)
	}

	// 201910011533_add_endpoint_to_deployments.go
	if err = migrations.AddEndpointToDeployments(db); err != nil {
		logging.Fatalf("Database migration (201910011533_add_endpoint_to_deployments) failed. %v", err)
	}

	// 201910011605_add_cluster_id_to_environments.go
	if err = migrations.AddClusterIDToEnvironments(db); err != nil {
		logging.Fatalf("Database migration (201910011605_add_cluster_id_to_environments) failed. %v", err)
	}

	// 201910021345_create_templates.go
	if err = migrations.CreateTemplates(db); err != nil {
		logging.Fatalf("Database migration (201910021345_create_templates) failed. %v", err)
	}

	// 201910031626_add_sort_order_to_templates.go
	if err = migrations.AddSortOrderToTemplates(db); err != nil {
		logging.Fatalf("Database migration (201910031626_add_sort_order_to_templates) failed. %v", err)
	}

	// 201910081532_add_project_and_domain_ids.go
	if err = migrations.AddProjectAndDomainIDs(db); err != nil {
		logging.Fatalf("Database migration (201910081532_add_project_and_domain_ids) failed. %v", err)
	}

	// 201910221422_add_service_to_deployments.go
	if err = migrations.AddServiceToDeployments(db); err != nil {
		logging.Fatalf("Database migration (201910221422_add_service_to_deployments) failed. %v", err)
	}

	// 201910221423_add_storage_provider_to_deployments.go
	if err = migrations.AddStorageProviderToDeployments(db); err != nil {
		logging.Fatalf("Database migration (201910221423_add_storage_provider_to_deployments) failed. %v", err)
	}

	// 201910221424_add_short_name_to_database_types.go
	if err = migrations.AddShortNameToDatabaseTypes(db); err != nil {
		logging.Fatalf("Database migration (201910221424_add_short_name_to_database_types) failed. %v", err)
	}

	// 201910231045_update_image_environments.go
	if err = migrations.UpdateImageEnvironments(db); err != nil {
		logging.Fatalf("Database migration (201910231045_update_image_environments) failed. %v", err)
	}

	// 201910231545_add_configurations_to_templates.go
	if err = migrations.AddConfigurationsToTemplates(db); err != nil {
		logging.Fatalf("Database migration (201910231545_add_configurations_to_templates) failed. %v", err)
	}

	// 201911201555_add_connection_details_to_deployments.go
	if err = migrations.AddConnectionDetailsToDeployments(db); err != nil {
		logging.Fatalf("Database migration (201911201555_add_connection_details_to_deployments) failed. %v", err)
	}

	// 202001141038_add_user_details.go
	if err = migrations.AddUserDetails(db); err != nil {
		logging.Fatalf("Database migration (202001141038_add_user_details) failed. %v", err)
	}

	// 202002101721_add_origin_to_deployments.go
	if err = migrations.AddOriginToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202002101721_add_origin_to_deployments) failed. %v", err)
	}

	// 202002131410_add_user_details_to_tasks.go
	if err = migrations.AddUserDetailsToTasks(db); err != nil {
		logging.Fatalf("Database migration (202002131410_add_user_details_to_tasks) failed. %v", err)
	}

	// 202002191404_add_build_to_deployments.go
	if err = migrations.AddBuildToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202002191404_add_build_to_deployments) failed. %v", err)
	}

	// 202002191520_add_deployment_id_to_deployments.go
	if err = migrations.AddDeploymentIDToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202002191520_add_deployment_id_to_deployments) failed. %v", err)
	}

	// 202002211035_set_deployment_ids.go
	if err = migrations.SetDeploymentIDs(db); err != nil {
		logging.Fatalf("Database migration (202002211035_set_deployment_ids) failed. %v", err)
	}

	// 202004011615_add_registered_in_vault_to_deployments.go
	if err = migrations.AddRegisteredInVaultToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202004011615_add_registered_in_vault_to_deployments) failed. %v", err)
	}

	// 202005181500_update_environment_clusters.go
	if err = migrations.UpdateEnvironmentClusters(db); err != nil {
		logging.Fatalf("Database migration (202005181500_update_environment_clusters) failed. %v", err)
	}

	// 202007160930_add_backups.go
	if err = migrations.AddBackups(db); err != nil {
		logging.Fatalf("Database migration (202007160930_add_backups) failed. %v", err)
	}

	// 202008130930_add_backup_schedule_to_deployments.go
	if err = migrations.AddBackupScheduleToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202008130930_add_backup_schedule_to_deployments) failed. %v", err)
	}

	// 202008210930_create_backup_types.go
	if err = migrations.CreateBackupTypes(db); err != nil {
		logging.Fatalf("Database migration (202008210930_create_backup_types) failed. %v", err)
	}

	// 202008261630_add_type_to_backups.go
	if err = migrations.AndTypeAndLevelToBackups(db); err != nil {
		logging.Fatalf("Database migration (202008261630_add_type_to_backups) failed. %v", err)
	}

	// 202009141530_add_endpoint_and_schedule_to_backups.go
	if err = migrations.AndEndpointAndScheduleToBackups(db); err != nil {
		logging.Fatalf("Database migration (202009141530_add_endpoint_and_schedule_to_backups) failed. %v", err)
	}

	// 20209211330_add_backup_id_to_backups.go
	if err = migrations.AddBackupIDToBackups(db); err != nil {
		logging.Fatalf("Database migration (20209211330_add_backup_id_to_backups) failed. %v", err)
	}

	// 202009221430_add_options_to_database_types.go
	if err = migrations.AddOptionsToDatabaseTypes(db); err != nil {
		logging.Fatalf("Database migration (202009221430_add_options_to_database_types) failed. %v", err)
	}

	// 202011091430_add_job_history_limit_to_backups.go
	if err = migrations.AddJobHistoryLimitToBackups(db); err != nil {
		logging.Fatalf("Database migration (202011091430_add_job_history_limit_to_backups) failed. %v", err)
	}

	// 202011111030_add_target_clusters_to_deployments.go
	if err = migrations.AddTargetClustersToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202011111030_add_target_clusters_to_deployments) failed. %v", err)
	}

	// 202012161230_add_backup_limits_to_deployments.go
	if err = migrations.AddBackupLimitsToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202012161230_add_backup_limits_to_deployments) failed. %v", err)
	}

	// 202002020830_create_deployment_groups.go
	if err = migrations.CreateDeploymentGroups(db); err != nil {
		logging.Fatalf("Database migration (202002020830_create_deployment_groups) failed. %v", err)
	}

	// 202102241300_add_deployment_groups_to_deployments.go
	if err = migrations.AddDeploymentGroupsToDeployments(db); err != nil {
		logging.Fatalf("Database migration (202102241300_add_deployment_groups_to_deployments) failed. %v", err)
	}

	// 202102251600_update_target_clusters.go
	if err = migrations.UpdateTargetClusters(db); err != nil {
		logging.Fatalf("Database migration (202102251600_update_target_clusters) failed. %v", err)
	}

	// 202103011500_create_target_clusters.go
	if err = migrations.CreateTargetClusters(db); err != nil {
		logging.Fatalf("Database migration (202103011500_create_target_clusters) failed. %v", err)
	}

	// 202108171540_add_pipeline_parameters_to_deployment.go
	if err = migrations.AddPipelineParametersToDeployment(db); err != nil {
		logging.Fatalf("Database migration (202108171540_add_pipeline_parameters_to_deployment) failed. %v", err)
	}

	// 202108271100_add_api_server_to_target_clusters.go
	if err = migrations.AddAPIServerToTargetClusters(db); err != nil {
		logging.Fatalf("Database migration (202108271100_add_api_server_to_target_clusters) failed. %v", err)
	}

	// 202109171540_add_servicetype_parameter_to_deployment.go
	if err = migrations.AddServiceTypeParameterToDeployment(db); err != nil {
		logging.Fatalf("Database migration (202109171540_add_servicetype_parameter_to_deployment) failed. %v", err)
	}
}

func seedData(db *gorm.DB) {
	// 202108170000_database_types.go
	if err := seeding.InsertDefaultDbTypes(db); err != nil {
		logging.Fatalf("Database migration (202108170000_database_types) failed. %v", err)
	}
}
