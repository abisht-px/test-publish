package controlplane

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

var (
	dataServiceExpectedMetrics = map[string][]parser.VectorSelector{
		dataservices.Cassandra: {
			// cassandra_clientrequest_latency_seconds_sum{clientrequest="Read",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_latency_seconds_sum{clientrequest="Write",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_latency_seconds_count{clientrequest="Read",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_latency_seconds_count{clientrequest="Write",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_latency_seconds_count{clientrequest!~".*Write.*",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_latency_seconds_count{clientrequest=~".*Write.*",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_timeouts_count{clientrequest!~".*Write.*",pds_deployment_id=":deployment_id"}
			// cassandra_clientrequest_timeouts_count{clientrequest=~".*Write.*",pds_deployment_id=":deployment_id"}
			{
				Name: "cassandra_clientrequest_latency_seconds_sum",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "clientrequest", "Read"),
				},
			},
			{
				Name: "cassandra_clientrequest_latency_seconds_sum",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "clientrequest", "Write"),
				},
			},
			{
				Name: "cassandra_clientrequest_latency_seconds_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "clientrequest", "Read"),
				},
			},
			{
				Name: "cassandra_clientrequest_latency_seconds_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchNotEqual, "clientrequest", "Write"),
				},
			},
			{
				Name: "cassandra_clientrequest_latency_seconds_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchNotRegexp, "clientrequest", ".*Write.*"),
				},
			},
			{
				Name: "cassandra_clientrequest_latency_seconds_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchRegexp, "clientrequest", ".*Write.*"),
				},
			},
			{
				Name: "cassandra_clientrequest_timeouts_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchNotRegexp, "clientrequest", ".*Write.*"),
				},
			},
			{
				Name: "cassandra_clientrequest_timeouts_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchRegexp, "clientrequest", ".*Write.*"),
				},
			},
		},
		dataservices.Couchbase: {
			// cbnode_interestingstats_curr_items{pds_deployment_id=":deployment_id"}
			// cbnode_interestingstats_curr_items_tot{pds_deployment_id=":deployment_id"}
			// cbnode_interestingstats_couch_docs_actual_disk_size{pds_deployment_id=":deployment_id"}
			// cbnode_interestingstats_cmd_get{pds_deployment_id=":deployment_id"}
			// cbnode_interestingstats_get_hits{pds_deployment_id=":deployment_id"}
			// cbnode_interestingstats_ep_bg_fetched{pds_deployment_id=":deployment_id"}
			// cbnode_interestingstats_ops{pds_deployment_id=":deployment_id"})
			{Name: "cbnode_interestingstats_curr_items"},
			{Name: "cbnode_interestingstats_curr_items_tot"},
			{Name: "cbnode_interestingstats_couch_docs_actual_disk_size"},
			{Name: "cbnode_interestingstats_cmd_get"},
			{Name: "cbnode_interestingstats_get_hits"},
			{Name: "cbnode_interestingstats_ep_bg_fetched"},
			{Name: "cbnode_interestingstats_ops"},
		},
		dataservices.Consul: {
			// consul_members_clients{pds_deployment_id=":deployment_id"}
			// consul_members_servers{pds_deployment_id=":deployment_id"}
			// consul_state_services{pds_deployment_id=":deployment_id"}
			// consul_kvs_apply_count{pds_deployment_id=":deployment_id"}
			// consul_txn_apply_count{pds_deployment_id=":deployment_id"}
			// consul_rpc_request{pds_deployment_id=":deployment_id"}
			// consul_kvs_apply{quantile="0.5",pds_deployment_id=":deployment_id"}
			// consul_kvs_apply{quantile="0.9",pds_deployment_id=":deployment_id"}
			// consul_kvs_apply{quantile="0.99",pds_deployment_id=":deployment_id"}
			// consul_catalog_register_count{pds_deployment_id=":deployment_id"}
			// consul_catalog_deregister_count{pds_deployment_id=":deployment_id"}
			// consul_raft_apply{pds_deployment_id=":deployment_id"}
			// consul_raft_commitTime{quantile="0.5",pds_deployment_id=":deployment_id"}
			// consul_raft_commitTime{quantile="0.9",pds_deployment_id=":deployment_id"}
			// consul_raft_commitTime{quantile="0.99",pds_deployment_id=":deployment_id"}
			{Name: "consul_members_clients"},
			{Name: "consul_members_servers"},
			{Name: "consul_state_services"},
			{Name: "consul_kvs_apply_count"},
			{Name: "consul_txn_apply_count"},
			{Name: "consul_rpc_request"},
			{
				Name: "consul_kvs_apply",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "quantile", "0.5"),
				},
			},
			{
				Name: "consul_kvs_apply",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "quantile", "0.9"),
				},
			},
			{
				Name: "consul_kvs_apply",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "quantile", "0.99"),
				},
			},
			{Name: "consul_catalog_register_count"},
			{Name: "consul_catalog_deregister_count"},
			{Name: "consul_raft_apply"},
			{
				Name: "consul_raft_commitTime",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "quantile", "0.5"),
				},
			},
			{
				Name: "consul_raft_commitTime",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "quantile", "0.9"),
				},
			},
			{
				Name: "consul_raft_commitTime",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "quantile", "0.99"),
				},
			},
		},
		dataservices.Kafka: {
			// kafka_server_replicamanager_partitioncount{pds_deployment_id=":deployment_id"}
			// kafka_server_replicamanager_underreplicatedpartitions{pds_deployment_id=":deployment_id"}
			// kafka_server_brokertopicmetrics_bytesin_total{pds_deployment_id=":deployment_id",topic=""}
			// kafka_server_brokertopicmetrics_bytesout_total{pds_deployment_id=":deployment_id",topic=""}
			// kafka_server_brokertopicmetrics_totalproducerequests_total{pds_deployment_id=":deployment_id",topic=""}
			// kafka_server_brokertopicmetrics_failedproducerequests_total{pds_deployment_id=":deployment_id",topic=""}
			// kafka_server_brokertopicmetrics_totalfetchrequests_total{pds_deployment_id=":deployment_id",topic=""}
			// kafka_server_brokertopicmetrics_failedfetchrequests_total{pds_deployment_id=":deployment_id",topic=""}
			// kafka_server_brokertopicmetrics_messagesin_total{pds_deployment_id=":deployment_id",topic=""}
			{Name: "kafka_server_replicamanager_partitioncount"},
			{Name: "kafka_server_replicamanager_underreplicatedpartitions"},
			{
				Name: "kafka_server_brokertopicmetrics_bytesin_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "kafka_server_brokertopicmetrics_bytesout_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "kafka_server_brokertopicmetrics_totalproducerequests_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "kafka_server_brokertopicmetrics_failedproducerequests_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "kafka_server_brokertopicmetrics_totalfetchrequests_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "kafka_server_brokertopicmetrics_failedfetchrequests_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "kafka_server_brokertopicmetrics_messagesin_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
		},
		dataservices.MongoDB: {
			// mongodb_connections{state="active",pds_deployment_id=":deployment_id"}
			// mongodb_connections{state="current",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_metrics_document_total{state="deleted",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_metrics_document_total{state="inserted",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_metrics_document_total{state="returned",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_metrics_document_total{state="updated",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcounters{legacy_op_type="command",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcounters{legacy_op_type="delete",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcounters{legacy_op_type="getmore",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcounters{legacy_op_type="insert",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcounters{legacy_op_type="query",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcounters{legacy_op_type="update",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcountersRepl{legacy_op_type="command",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcountersRepl{legacy_op_type="delete",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcountersRepl{legacy_op_type="getmore",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcountersRepl{legacy_op_type="insert",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcountersRepl{legacy_op_type="query",pds_deployment_id=":deployment_id"}
			// mongodb_ss_opcountersRepl{legacy_op_type="update",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_op_latencies_latency_total{type="command",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_op_latencies_latency_total{type="read",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_op_latencies_latency_total{type="transactions",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_op_latencies_latency_total{type="write",pds_deployment_id=":deployment_id"}
			// mongodb_mongod_replset_member_replication_lag{pds_deployment_id=":deployment_id"}
			{
				Name: "mongodb_connections",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "active"),
				},
			},
			{
				Name: "mongodb_connections",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "current"),
				},
			},
			{
				Name: "mongodb_mongod_metrics_document_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "deleted"),
				},
			},
			{
				Name: "mongodb_mongod_metrics_document_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "inserted"),
				},
			},
			{
				Name: "mongodb_mongod_metrics_document_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "returned"),
				},
			},
			{
				Name: "mongodb_mongod_metrics_document_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "updated"),
				},
			},
			{
				Name: "mongodb_ss_opcounters",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "command"),
				},
			},
			{
				Name: "mongodb_ss_opcounters",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "delete"),
				},
			},
			{
				Name: "mongodb_ss_opcounters",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "getmore"),
				},
			},
			{
				Name: "mongodb_ss_opcounters",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "insert"),
				},
			},
			{
				Name: "mongodb_ss_opcounters",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "query"),
				},
			},
			{
				Name: "mongodb_ss_opcounters",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "update"),
				},
			},
			{
				Name: "mongodb_ss_opcountersRepl",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "command"),
				},
			},
			{
				Name: "mongodb_ss_opcountersRepl",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "delete"),
				},
			},
			{
				Name: "mongodb_ss_opcountersRepl",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "getmore"),
				},
			},
			{
				Name: "mongodb_ss_opcountersRepl",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "insert"),
				},
			},
			{
				Name: "mongodb_ss_opcountersRepl",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "query"),
				},
			},
			{
				Name: "mongodb_ss_opcountersRepl",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "legacy_op_type", "update"),
				},
			},
			{
				Name: "mongodb_mongod_op_latencies_latency_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "type", "command"),
				},
			},
			{
				Name: "mongodb_mongod_op_latencies_latency_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "type", "read"),
				},
			},
			{
				Name: "mongodb_mongod_op_latencies_latency_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "type", "transactions"),
				},
			},
			{
				Name: "mongodb_mongod_op_latencies_latency_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "type", "write"),
				},
			},
			{Name: "mongodb_mongod_replset_member_replication_lag"},
		},
		dataservices.MySQL: {
			// mysql_global_status_threads_connected{pds_deployment_id=":deployment_id"}
			// mysql_global_variables_max_connections{pds_deployment_id=":deployment_id"}
			// mysql_global_status_slow_queries{pds_deployment_id=":deployment_id"}
			// mysql_global_status_select_full_join{pds_deployment_id=":deployment_id"}
			// mysql_global_variables_innodb_open_files{pds_deployment_id=":deployment_id"}
			// mysql_global_variables_open_files_limit{pds_deployment_id=":deployment_id"}
			// mysql_global_status_innodb_buffer_pool_reads{pds_deployment_id=":deployment_id"}
			// mysql_global_status_innodb_buffer_pool_read_requests{pds_deployment_id=":deployment_id"}
			// mysql_global_status_table_open_cache_hits{pds_deployment_id=":deployment_id"}
			// mysql_global_status_table_open_cache_misses{pds_deployment_id=":deployment_id"}
			// mysql_global_status_connection_errors_total{pds_deployment_id=":deployment_id"}
			{Name: "mysql_global_status_threads_connected"},
			{Name: "mysql_global_variables_max_connections"},
			{Name: "mysql_global_status_slow_queries"},
			{Name: "mysql_global_status_select_full_join"},
			{Name: "mysql_global_variables_innodb_open_files"},
			{Name: "mysql_global_variables_open_files_limit"},
			{Name: "mysql_global_status_innodb_buffer_pool_reads"},
			{Name: "mysql_global_status_innodb_buffer_pool_read_requests"},
			{Name: "mysql_global_status_table_open_cache_hits"},
			{Name: "mysql_global_status_table_open_cache_misses"},
			{Name: "mysql_global_status_connection_errors_total"},
		},
		dataservices.ElasticSearch: {
			// elasticsearch_cluster_health_active_primary_shards{pds_deployment_id=":deployment_id"}
			// elasticsearch_cluster_health_active_shards{pds_deployment_id=":deployment_id"}
			// elasticsearch_cluster_health_relocating_shards{pds_deployment_id=":deployment_id"}
			// elasticsearch_cluster_health_initializing_shards{pds_deployment_id=":deployment_id"}
			// elasticsearch_cluster_health_unassigned_shards{pds_deployment_id=":deployment_id"}
			// elasticsearch_indices_search_query_time_seconds{pds_deployment_id=":deployment_id",topic=""}
			// elasticsearch_indices_search_fetch_time_seconds{pds_deployment_id=":deployment_id",topic=""}
			// elasticsearch_indices_indexing_index_time_seconds_total{pds_deployment_id=":deployment_id",topic=""}
			// elasticsearch_indices_refresh_time_seconds_total{pds_deployment_id=":deployment_id",topic=""}
			// elasticsearch_indices_flush_time_seconds{pds_deployment_id=":deployment_id",topic=""}
			// elasticsearch_process_max_files_descriptors{pds_deployment_id=":deployment_id"}
			// elasticsearch_process_open_files_count{pds_deployment_id=":deployment_id"}

			{Name: "elasticsearch_cluster_health_active_primary_shards"},
			{Name: "elasticsearch_cluster_health_active_shards"},
			{Name: "elasticsearch_cluster_health_relocating_shards"},
			{Name: "elasticsearch_cluster_health_initializing_shards"},
			{Name: "elasticsearch_cluster_health_unassigned_shards"},
			{
				Name: "elasticsearch_indices_search_query_time_seconds",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "elasticsearch_indices_search_fetch_time_seconds",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "elasticsearch_indices_indexing_index_time_seconds_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "elasticsearch_indices_refresh_time_seconds_total",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{
				Name: "elasticsearch_indices_flush_time_seconds",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "topic", ""),
				},
			},
			{Name: "elasticsearch_process_max_files_descriptors"},
			{Name: "elasticsearch_process_open_files_count"},
		},
		dataservices.Postgres: {
			// pg_stat_activity_count{state="active",pds_deployment_id=":deployment_id"}
			// pg_stat_activity_count{state="idle",pds_deployment_id=":deployment_id"}
			// pg_stat_activity_count{pds_deployment_id=":deployment_id"}
			// pg_stat_database_xact_commit{pds_deployment_id=":deployment_id"}
			// pg_stat_database_xact_rollback{pds_deployment_id=":deployment_id"}
			// pg_stat_database_tup_inserted{pds_deployment_id=":deployment_id"}
			// pg_stat_database_tup_updated{pds_deployment_id=":deployment_id"}
			// pg_stat_database_tup_deleted{pds_deployment_id=":deployment_id"}
			// pg_stat_database_tup_fetched{pds_deployment_id=":deployment_id"}
			// pg_stat_database_tup_returned{pds_deployment_id=":deployment_id"}
			// pg_stat_database_blks_read{pds_deployment_id=":deployment_id"}
			// pg_stat_database_blks_hit{pds_deployment_id=":deployment_id"}
			{
				Name: "pg_stat_activity_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "active"),
				},
			},
			{
				Name: "pg_stat_activity_count",
				LabelMatchers: []*labels.Matcher{
					parser.MustLabelMatcher(labels.MatchEqual, "state", "idle"),
				},
			},
			{Name: "pg_stat_activity_count"},
			{Name: "pg_stat_database_xact_commit"},
			{Name: "pg_stat_database_xact_rollback"},
			{Name: "pg_stat_database_tup_inserted"},
			{Name: "pg_stat_database_tup_updated"},
			{Name: "pg_stat_database_tup_deleted"},
			{Name: "pg_stat_database_tup_fetched"},
			{Name: "pg_stat_database_tup_returned"},
			{Name: "pg_stat_database_blks_read"},
			{Name: "pg_stat_database_blks_hit"},
		},
		dataservices.RabbitMQ: {
			// rabbitmq_global_messages_received_total{pds_deployment_id=":deployment_id"}
			// rabbitmq_global_messages_acknowledged_total{pds_deployment_id=":deployment_id"}
			// rabbitmq_connections{pds_deployment_id=":deployment_id"}
			// rabbitmq_consumers{pds_deployment_id=":deployment_id"}
			// rabbitmq_process_resident_memory_bytes{pds_deployment_id=":deployment_id"}
			// rabbitmq_resident_memory_limit_bytes{pds_deployment_id=":deployment_id"}
			{Name: "rabbitmq_global_messages_received_total"},
			{Name: "rabbitmq_global_messages_acknowledged_total"},
			{Name: "rabbitmq_connections"},
			{Name: "rabbitmq_consumers"},
			{Name: "rabbitmq_process_resident_memory_bytes"},
			{Name: "rabbitmq_resident_memory_limit_bytes"},
		},
		dataservices.Redis: {
			// redis_rejected_connections_total{pds_deployment_id=":deployment_id"}
			// redis_connected_clients{pds_deployment_id=":deployment_id"}
			// redis_config_maxclients{pds_deployment_id=":deployment_id"}
			// redis_slowlog_length{pds_deployment_id=":deployment_id"}
			// redis_commands_duration_seconds_total{pds_deployment_id=":deployment_id"}
			// redis_commands_processed_total{pds_deployment_id=":deployment_id"}
			// redis_keyspace_hits_total{pds_deployment_id=":deployment_id"}
			// redis_keyspace_misses_total{pds_deployment_id=":deployment_id"}
			// redis_expired_keys_total{pds_deployment_id=":deployment_id"}
			// redis_memory_used_bytes{pds_deployment_id=":deployment_id"}
			{Name: "redis_rejected_connections_total"},
			{Name: "redis_connected_clients"},
			{Name: "redis_config_maxclients"},
			{Name: "redis_slowlog_length"},
			{Name: "redis_commands_duration_seconds_total"},
			{Name: "redis_commands_processed_total"},
			{Name: "redis_keyspace_hits_total"},
			{Name: "redis_keyspace_misses_total"},
			{Name: "redis_expired_keys_total"},
			{Name: "redis_memory_used_bytes"},
		},
		dataservices.ZooKeeper: {
			// zookeeper_num_alive_connections{pds_deployment_id=":deployment_id"}
			// zookeeper_auth_failed_count{pds_deployment_id=":deployment_id"}
			// zookeeper_avg_latency{pds_deployment_id=":deployment_id"}
			// zookeeper_max_latency{pds_deployment_id=":deployment_id"}
			// zookeeper_min_latency{pds_deployment_id=":deployment_id"}
			// zookeeper_packets_received{pds_deployment_id=":deployment_id"}
			// zookeeper_packets_sent{pds_deployment_id=":deployment_id"}
			// zookeeper_outstanding_requests{pds_deployment_id=":deployment_id"}
			// zookeeper_open_file_descriptor_count{pds_deployment_id=":deployment_id"}
			// zookeeper_znode_count{pds_deployment_id=":deployment_id"}
			// zookeeper_ephemerals_count{pds_deployment_id=":deployment_id"}
			// zookeeper_max_client_response_size{pds_deployment_id=":deployment_id"}
			// zookeeper_min_client_response_size{pds_deployment_id=":deployment_id"}
			{Name: "zookeeper_num_alive_connections"},
			{Name: "zookeeper_auth_failed_count"},
			{Name: "zookeeper_avg_latency"},
			{Name: "zookeeper_max_latency"},
			{Name: "zookeeper_min_latency"},
			{Name: "zookeeper_packets_received"},
			{Name: "zookeeper_packets_sent"},
			{Name: "zookeeper_outstanding_requests"},
			{Name: "zookeeper_open_file_descriptor_count"},
			{Name: "zookeeper_znode_count"},
			{Name: "zookeeper_ephemerals_count"},
			{Name: "zookeeper_max_client_response_size"},
			{Name: "zookeeper_min_client_response_size"},
		},
	}
)

func (c *ControlPlane) MustVerifyMetrics(ctx context.Context, t *testing.T, deploymentID string) {
	deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := c.PDS.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	require.Contains(t, dataServiceExpectedMetrics, dataServiceType, "%s data service has no defined expected metrics")
	expectedMetrics := dataServiceExpectedMetrics[dataServiceType]

	var missingMetrics []model.LabelValue
	for _, expectedMetric := range expectedMetrics {
		// Add deployment ID to the metric label filter.
		pdsDeploymentIDMatch := parser.MustLabelMatcher(labels.MatchEqual, "pds_deployment_id", deploymentID)
		expectedMetric.LabelMatchers = append(expectedMetric.LabelMatchers, pdsDeploymentIDMatch)

		queryResult, _, err := c.Prometheus.Query(ctx, expectedMetric.String(), time.Now())
		require.NoError(t, err, "prometheus: query error")

		require.IsType(t, model.Vector{}, queryResult, "prometheus: wrong result model")
		queryResultMetrics := queryResult.(model.Vector)

		if len(queryResultMetrics) == 0 {
			missingMetrics = append(missingMetrics, model.LabelValue(expectedMetric.Name))
		}
	}

	require.Equalf(t, len(missingMetrics), 0, "%s: prometheus missing %d/%d metrics: %v", dataServiceType, len(missingMetrics), len(expectedMetrics), missingMetrics)
}
