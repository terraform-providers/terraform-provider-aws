## 2.68.0 (Unreleased)
## 2.67.0 (June 19, 2020)

NOTES:

* data-source/aws_secretsmanager_secret: The `rotation_*` attributes have been deprecated in favor of the new `aws_secretsmanager_secret_rotation` data source. ([#9487](https://github.com/terraform-providers/terraform-provider-aws/issues/9487))
* resource/aws_secretsmanager_secret: The `rotation_*` arguments have been deprecated in favor of the new `aws_secretsmanager_secret_rotation` resource. The new resource, which fixes workflow issues, prevents the existing resource from showing changes when attempting to remove rotation. Removal must be done with the new resource or manually. ([#9487](https://github.com/terraform-providers/terraform-provider-aws/issues/9487))

FEATURES:

* **New Data Source:** `aws_ebs_volumes` ([#13751](https://github.com/terraform-providers/terraform-provider-aws/issues/13751))
* **New Data Source:** `aws_ec2_local_gateway_virtual_interface` ([#13770](https://github.com/terraform-providers/terraform-provider-aws/issues/13770))
* **New Data Source:** `aws_ec2_local_gateway_virtual_interface_group` ([#13767](https://github.com/terraform-providers/terraform-provider-aws/issues/13767))
* **New Data Source:** `aws_ec2_local_gateway_virtual_interface_groups` ([#13767](https://github.com/terraform-providers/terraform-provider-aws/issues/13767))
* **New Data Source:** `aws_ecr_authorization_token` ([#12395](https://github.com/terraform-providers/terraform-provider-aws/issues/12395))
* **New Data Source:** `aws_outposts_outpost` ([#13777](https://github.com/terraform-providers/terraform-provider-aws/issues/13777))
* **New Data Source:** `aws_outposts_outpost_instance_type` ([#13824](https://github.com/terraform-providers/terraform-provider-aws/issues/13824))
* **New Data Source:** `aws_outposts_outpost_instance_types` ([#13824](https://github.com/terraform-providers/terraform-provider-aws/issues/13824))
* **New Data Source:** `aws_outposts_outposts` ([#13777](https://github.com/terraform-providers/terraform-provider-aws/issues/13777))
* **New Data Source:** `aws_outposts_site` ([#13825](https://github.com/terraform-providers/terraform-provider-aws/issues/13825))
* **New Data Source:** `aws_outposts_sites` ([#13825](https://github.com/terraform-providers/terraform-provider-aws/issues/13825))
* **New Data Source:** `aws_secretsmanager_secret_rotation` ([#9487](https://github.com/terraform-providers/terraform-provider-aws/issues/9487))
* **New Data Source:** `aws_wafv2_web_acl` ([#12791](https://github.com/terraform-providers/terraform-provider-aws/issues/12791))
* **New Guide:** [Resource Tagging](https://terraform.io/docs/providers/aws/guides/resource-tagging.html) ([#13779](https://github.com/terraform-providers/terraform-provider-aws/issues/13779))
* **New Resource:** `aws_ec2_local_gateway_route` ([#13768](https://github.com/terraform-providers/terraform-provider-aws/issues/13768))
* **New Resource:** `aws_ec2_local_gateway_route_table_vpc_association` ([#13765](https://github.com/terraform-providers/terraform-provider-aws/issues/13765))
* **New Resource:** `aws_ec2_tag` ([#8457](https://github.com/terraform-providers/terraform-provider-aws/issues/8457))
* **New Resource:** `aws_secretsmanager_secret_rotation` ([#9487](https://github.com/terraform-providers/terraform-provider-aws/issues/9487))
* **New Resource:** `aws_wafv2_web_acl` ([#12688](https://github.com/terraform-providers/terraform-provider-aws/issues/12688))
* **New Resource:** `aws_wafv2_web_acl_association` ([#12698](https://github.com/terraform-providers/terraform-provider-aws/issues/12698))

ENHANCEMENTS:

* data-source/aws_vpc_dhcp_options: Add `arn` attribute ([#13606](https://github.com/terraform-providers/terraform-provider-aws/issues/13606))
* data-source/aws_efs_mount_target: Add `availability_zone_name`, `availability_zone_id`, `mount_target_dns_name`, and `owner_id` attributes ([#13650](https://github.com/terraform-providers/terraform-provider-aws/issues/13650))
* data-source/aws_internet_gateway: Add `arn` attribute ([#13614](https://github.com/terraform-providers/terraform-provider-aws/issues/13614))
* data-source/aws_lambda_function: Add `file_system_config` attribute ([#13782](https://github.com/terraform-providers/terraform-provider-aws/issues/13782))
* data-source/aws_workspaces_bundle: Add `name` and `owner` as search crieria ([#12497](https://github.com/terraform-providers/terraform-provider-aws/issues/12497))
* resource/aws_ec2_capacity_reservation: Add `arn` attribute and support for Linux SQL Server values in `instance_platform` argument ([#13607](https://github.com/terraform-providers/terraform-provider-aws/issues/13607))
* resource/aws_cloudwatch_log_stream: Support resource import ([#13697](https://github.com/terraform-providers/terraform-provider-aws/issues/13697))
* resource/aws_default_vpc_dhcp_options: Add `arn` attribute ([#13606](https://github.com/terraform-providers/terraform-provider-aws/issues/13606))
* resource/aws_devicefarm_project: Support resource import ([#13020](https://github.com/terraform-providers/terraform-provider-aws/issues/13020))
* resource/aws_ec2_traffic_mirror_target: Add `arn` attribute ([#13649](https://github.com/terraform-providers/terraform-provider-aws/issues/13649))
* resource/aws_ecs_capacity_provider: Implement API deletion support ([#13740](https://github.com/terraform-providers/terraform-provider-aws/issues/13740))
* resource/aws_efs_mount_target: Add `availability_zone_name`, `availability_zone_id`, `mount_target_dns_name`, and `owner_id` attributes ([#13650](https://github.com/terraform-providers/terraform-provider-aws/issues/13650))
* resource/aws_efs_mount_target: Add plan time validation for `ip_address` argument ([#13650](https://github.com/terraform-providers/terraform-provider-aws/issues/13650))
* resource_aws_elasticsearch_domain: Add support for elasticsearch warm storage ([#12668](https://github.com/terraform-providers/terraform-provider-aws/issues/12668))
* resource_aws_elasticsearch_domain: Add plan time validation for `cloudwatch_log_group_arn`, `role_arn`, and `ebs_options.volume_type` ([#12668](https://github.com/terraform-providers/terraform-provider-aws/issues/12668))
* resource/aws_flow_log: Add `arn` attribute ([#13656](https://github.com/terraform-providers/terraform-provider-aws/issues/13656))
* resource/aws_glue_catalog_table: Add `arn` attribute ([#13456](https://github.com/terraform-providers/terraform-provider-aws/issues/13456))
* resource/aws_guardduty_detector: Add `tags` argument and `arn` attribute ([#13756](https://github.com/terraform-providers/terraform-provider-aws/issues/13756))
* resource/aws_guardduty_ipset: Add `tags` argument and `arn` attribute ([#13756](https://github.com/terraform-providers/terraform-provider-aws/issues/13756))
* resource/aws_guardduty_threatintelset: Add `tags` argument and `arn` attribute ([#13756](https://github.com/terraform-providers/terraform-provider-aws/issues/13756))
* resource/aws_internet_gateway: Add `arn` attribute ([#13614](https://github.com/terraform-providers/terraform-provider-aws/issues/13614))
* resource/aws_key_pair: Add `arn` attribute ([#13648](https://github.com/terraform-providers/terraform-provider-aws/issues/13648))
* resource/aws_lambda_function: Add `file_system_config` attribute ([#13782](https://github.com/terraform-providers/terraform-provider-aws/issues/13782))
* resource/aws_service_discovery_http_namespace: Add `tags` argument ([#13750](https://github.com/terraform-providers/terraform-provider-aws/issues/13750))
* resource/aws_service_discovery_private_dns_namespace: Add `tags` argument ([#13750](https://github.com/terraform-providers/terraform-provider-aws/issues/13750))
* resource/aws_service_discovery_public_dns_namespace: Add `tags` argument ([#13750](https://github.com/terraform-providers/terraform-provider-aws/issues/13750))
* resource/aws_service_discovery_service: Add `tags` argument ([#13750](https://github.com/terraform-providers/terraform-provider-aws/issues/13750))
* resource/aws_vpc_dhcp_options: Add `arn` attribute ([#13606](https://github.com/terraform-providers/terraform-provider-aws/issues/13606))
* resource/aws_wafv2_ip_set: Increase `addresses` maximum amount of items limit from 50 to 10,000 to match AWS limits ([#13698](https://github.com/terraform-providers/terraform-provider-aws/issues/13698))

BUG FIXES:

* resource/aws_cloudfront_origin_access_identity: Properly trigger resource recreation when deleted outside Terraform ([#13081](https://github.com/terraform-providers/terraform-provider-aws/issues/13081))
* resource/aws_cognito_user_pool_client: Mark `client_secret` as sensitive ([#13680](https://github.com/terraform-providers/terraform-provider-aws/issues/13680))
* resource/aws_route: Ensure retries occur on read after creation if route not found for EC2 eventual consistency ([#13747](https://github.com/terraform-providers/terraform-provider-aws/issues/13747))

## 2.66.0 (June 12, 2020)

FEATURES:

* **New Data Source:** `aws_wafv2_rule_group` ([#12790](https://github.com/terraform-providers/terraform-provider-aws/issues/12790))
* **New Resource:** `aws_wafv2_rule_group` ([#12677](https://github.com/terraform-providers/terraform-provider-aws/issues/12677))

BUG FIXES:

* resource/aws_autoscaling_group: Allow `on_demand_base_capacity` to be set to 0 [[#13623](https://github.com/terraform-providers/terraform-provider-aws/issues/13623)] 
* resource/aws_autoscaling_group: Add `Computed` field to `instances_distribution` and it's sub-fields `on_demand_allocation_strategy`, `on_demand_base_capacity`, `on_demand_percentage_above_base_capacity`, and `spot_allocation_strategy`  ([#13623](https://github.com/terraform-providers/terraform-provider-aws/issues/13623))
* resource/aws_autoscaling_group: Remove `Default` field from `instances_distribution` sub-fields `on_demand_allocation_strategy`, `on_demand_percentage_above_base_capacity`, and `spot_allocation_strategy`  ([#13623](https://github.com/terraform-providers/terraform-provider-aws/issues/13623))
* resource/aws_batch_job_definition: Prevent differences when no `command` is specified in container properties ([#13634](https://github.com/terraform-providers/terraform-provider-aws/issues/13634))
* resource/aws_instance: Continue supporting empty string (`""`) `private_ip` argument ([#13640](https://github.com/terraform-providers/terraform-provider-aws/issues/13640))

## 2.65.0 (June 04, 2020)

ENHANCEMENTS:

* resource/aws_acm_certificate: Add `status` attribute ([#13513](https://github.com/terraform-providers/terraform-provider-aws/issues/13513))
* resource/aws_directory_servicedirectory: Add `availability_zones` attribute to `vpc_settings` block ([#12654](https://github.com/terraform-providers/terraform-provider-aws/issues/12654))
* resource/aws_directory_servicedirectory: Add `availability_zones` attribute to `connect_settings` block ([#12654](https://github.com/terraform-providers/terraform-provider-aws/issues/12654))
* resource/aws_directory_servicedirectory: Add plan time validation to `customer_dns_ips` in `connect_settings` block ([#12654](https://github.com/terraform-providers/terraform-provider-aws/issues/12654))
* resource/aws_ec2_client_vpn_endpoint: Add `arn` attribute and plan time validation to `root_certificate_chain_arn` (in `authentication_options` block), `client_cidr_block`, and `server_certificate_arn` [[#13601](https://github.com/terraform-providers/terraform-provider-aws/issues/13601)] 
* resource/aws_instance: Add plan time validation to `volume_type`(in `ebs_block_device` and `root_block_device` blocks), `private_ip`, `ipv6_addresses`, and `tenancy` ([#13033](https://github.com/terraform-providers/terraform-provider-aws/issues/13033))
* resource/aws_lb_listener_rule: Add support for multiple, weighted target groups in `forward` rules ([#12574](https://github.com/terraform-providers/terraform-provider-aws/issues/12574))
* resource/aws_lb_listener: Add support for multiple, weighted target groups in default actions ([#12574](https://github.com/terraform-providers/terraform-provider-aws/issues/12574))
* resource/aws_workspaces_ip_group: Add plan-time validation for `rules.source` ([#13178](https://github.com/terraform-providers/terraform-provider-aws/issues/13178))

BUG FIXES:

* resource/aws_acm_certificate: Detect `AMAZON_ISSUED` type `validation_method` value directly from API response instead of custom logic ([#13513](https://github.com/terraform-providers/terraform-provider-aws/issues/13513))
* resource/aws_acm_certificate: Increase deletion retries from 10 minutes to 20 minutes (better support API Gateway Custom Domain deletion) ([#13513](https://github.com/terraform-providers/terraform-provider-aws/issues/13513))
* resource/aws_apigatewayv2_stage: Prevent perpetual plan differences with `default_route_settings.logging_level` argument for HTTP APIs ([#12904](https://github.com/terraform-providers/terraform-provider-aws/issues/12904))
* resource/aws_appmesh_route: Allow configuration of `spec` `http_route` `action` `weighted_target` `weight` argument to be 0 ([#13539](https://github.com/terraform-providers/terraform-provider-aws/issues/13539))
* resource/aws_autoscaling_group: Prevent crash with `tags` argument containing boolean values in Terraform 0.11 and earlier ([#13604](https://github.com/terraform-providers/terraform-provider-aws/issues/13604))
* resource/aws_dynamodb_table: Prevent multiple replica creation/deletion errors ([#13523](https://github.com/terraform-providers/terraform-provider-aws/issues/13523))
* resource/aws_instance: Prevent perpetual plan differences, forcing replacement, with `ebs_block_device` configuration blocks [[#13589](https://github.com/terraform-providers/terraform-provider-aws/issues/13589)] 
* resource/aws_kinesis_firehose_delivery_stream: Correctly set `kinesis_source_configuration` during import to prevent resource recreation ([#13536](https://github.com/terraform-providers/terraform-provider-aws/issues/13536))
* resource/aws_ses_configuration_set: Prevent `Provider produced inconsistent result after apply` errors during creation or import [[#12024](https://github.com/terraform-providers/terraform-provider-aws/issues/12024)] 
* resource/aws_workspaces_ip_group: Remove resource from state if deleted outside of Terraform ([#13178](https://github.com/terraform-providers/terraform-provider-aws/issues/13178))

## 2.64.0 (May 28, 2020)

ENHANCEMENTS:

* data-source/aws_directory_service_directory: `connect_settings` `connect_ips` attribute now set ([#13395](https://github.com/terraform-providers/terraform-provider-aws/issues/13395))
* resource/aws_directory_service_directory: `connect_settings` `connect_ips` attribute now set ([#13395](https://github.com/terraform-providers/terraform-provider-aws/issues/13395))
* resource/aws_iot_topic_rule: Add `step_functions` configuration block ([#13520](https://github.com/terraform-providers/terraform-provider-aws/issues/13520))
* resource/aws_ses_event_destination: Support resource import ([#13464](https://github.com/terraform-providers/terraform-provider-aws/issues/13464))

BUG FIXES:

* data-source/aws_elasticsearch_domain: `processing` is now correctly set ([#13397](https://github.com/terraform-providers/terraform-provider-aws/issues/13397))
* resource/aws_acm_certificate: Update pending DNS validation record creation time from 1 minute to 5 minutes (better support for certificates with high amount of Subject Alternative Names) ([#12371](https://github.com/terraform-providers/terraform-provider-aws/issues/12371))
* resource/aws_api_gateway_method_settings: `settings` now properly set ([#13403](https://github.com/terraform-providers/terraform-provider-aws/issues/13403))
* resource/aws_autoscaling_group: Ignore ordering differences for `tags` argument (prevent unexpected differences from version 2.63.0) ([#13515](https://github.com/terraform-providers/terraform-provider-aws/issues/13515))
* resource/aws_codebuild_project: Enable drift detection for `environment_variable` argument ([#6427](https://github.com/terraform-providers/terraform-provider-aws/issues/6427))
* resource/aws_codebuild_project: Prevent `inconsistent final plan` errors with `source` configuration block ([#10615](https://github.com/terraform-providers/terraform-provider-aws/issues/10615))
* resource/aws_ecs_task_definition: Ensure `efs_volume_configuration` changes are properly detected ([#12571](https://github.com/terraform-providers/terraform-provider-aws/issues/12571)] / [[#12751](https://github.com/terraform-providers/terraform-provider-aws/issues/12751))
* resource/aws_lb_cookie_stickiness_policy: `cookie_expiration_policy` now properly set ([#13418](https://github.com/terraform-providers/terraform-provider-aws/issues/13418))
* resource/aws_lightsail_instance: `ram_size` now properly set ([#13430](https://github.com/terraform-providers/terraform-provider-aws/issues/13430))
* resource/aws_load_balancer_backend_server_policy: `instance_port` now properly set ([#13418](https://github.com/terraform-providers/terraform-provider-aws/issues/13418))
* resource/aws_load_balancer_listener_policy: `load_balancer_port` now properly set ([#13418](https://github.com/terraform-providers/terraform-provider-aws/issues/13418))
* resource/aws_opsworks_application: `environment` `secure` now properly set ([#13435](https://github.com/terraform-providers/terraform-provider-aws/issues/13435))
* resource/aws_security_group_rule: Correctly set `description` after state refresh when `source_security_group_id` refers to a security group across accounts ([#13364](https://github.com/terraform-providers/terraform-provider-aws/issues/13364))
* resource/aws_ses_active_receipt_rule_set: Recreate resource when destroyed outside of Terraform ([#9086](https://github.com/terraform-providers/terraform-provider-aws/issues/9086))
* resource/aws_ses_event_destination: Correctly refresh entire resource state (prevent unexpected differences from version 2.63.0 and properly perform drift detection) ([#13464](https://github.com/terraform-providers/terraform-provider-aws/issues/13464))
* resource/aws_ses_receipt_rule: Recreate resource when destroyed outside of Terraform ([#9086](https://github.com/terraform-providers/terraform-provider-aws/issues/9086))
* resource/aws_sns_topic: Attributes of type `schema.TypeInt` are now correctly set ([#13437](https://github.com/terraform-providers/terraform-provider-aws/issues/13437))

## 2.63.0 (May 22, 2020)

FEATURES:

* **New Data Source:** `aws_efs_access_point` ([#11965](https://github.com/terraform-providers/terraform-provider-aws/issues/11965))
* **New Data Source:** `aws_wafv2_ip_set` ([#12788](https://github.com/terraform-providers/terraform-provider-aws/issues/12788))
* **New Data Source:** `aws_wafv2_regex_pattern_set` ([#12789](https://github.com/terraform-providers/terraform-provider-aws/issues/12789))
* **New Resource:** `aws_efs_access_point` ([#11965](https://github.com/terraform-providers/terraform-provider-aws/issues/11965))
* **New Resource:** `aws_efs_file_system_policy` ([#11960](https://github.com/terraform-providers/terraform-provider-aws/issues/11960))
* **New Resource:** `aws_wafv2_ip_set` ([#12119](https://github.com/terraform-providers/terraform-provider-aws/issues/12119))
* **New Resource:** `aws_wafv2_regex_pattern_set` ([#12284](https://github.com/terraform-providers/terraform-provider-aws/issues/12284))

ENHANCEMENTS:

* resource/aws_ssm_document: Add `document_version` attribute ([#13438](https://github.com/terraform-providers/terraform-provider-aws/issues/13438))
* data-source/aws_ram_resource_share: Add `owning_account_id` attribute ([#13402](https://github.com/terraform-providers/terraform-provider-aws/issues/13402))
* data-source/aws_lb: Add `ip_address_type` attribute ([#13400](https://github.com/terraform-providers/terraform-provider-aws/issues/13400))
* data-source/aws_lb_target_group: Add `load_balancing_algorithm_type` attribute ([#13400](https://github.com/terraform-providers/terraform-provider-aws/issues/13400))
* data-source/aws_rds_cluster: `backtrack_window` attribute now available ([#13362](https://github.com/terraform-providers/terraform-provider-aws/issues/13362))
* resource/aws_codebuild_webhook: Support `COMMIT_MESSAGE` value in filter types ([#13436](https://github.com/terraform-providers/terraform-provider-aws/issues/13436))
* resource/aws_cognito_identity_pool_roles_attachment: Add import support ([#13440](https://github.com/terraform-providers/terraform-provider-aws/issues/13440))
* resource/aws_ecs_service: Add `force_new_deployment` argument ([#13376](https://github.com/terraform-providers/terraform-provider-aws/issues/13376))
* resource/aws_ecs_service: Support in-place updates for `ordered_placement_strategy` and `placement_constraints` ([#13376](https://github.com/terraform-providers/terraform-provider-aws/issues/13376))
* resource/aws_eks_node_group: Add `force_update_version` argument ([#13414](https://github.com/terraform-providers/terraform-provider-aws/issues/13414))
* resource/aws_glue_connection: Add `arn` argument ([#13404](https://github.com/terraform-providers/terraform-provider-aws/issues/13404))
* resource/aws_iot_topic_rule: Add `tags` argument ([#13293](https://github.com/terraform-providers/terraform-provider-aws/issues/13293))

BUG FIXES:

* resource/aws_ssm_activation: `expired` now properly set ([#13438](https://github.com/terraform-providers/terraform-provider-aws/issues/13438))
* resource/aws_redshift_security_group: The resource is now importable ([#13431](https://github.com/terraform-providers/terraform-provider-aws/issues/13431))
* resource/cloudwatch_log_metric_filter: `metric_transformation` `default_value` now properly set ([#13411](https://github.com/terraform-providers/terraform-provider-aws/issues/13411))
* data-source/aws_db_instance: `auto_minor_version_upgrade` attribute now properly set ([#13362](https://github.com/terraform-providers/terraform-provider-aws/issues/13362))
* resource/aws_autoscaling_group: `tags` `propagate_at_launch` attribute now properly set ([#13360](https://github.com/terraform-providers/terraform-provider-aws/issues/13360))
* resource/aws_eks_node_group: Only pass `release_version` value during `UpdateNodegroupVersion` if changed ([#13407](https://github.com/terraform-providers/terraform-provider-aws/issues/13407))
* resource/aws_network_acl: Fix issue with updating subnet associations returning `InvalidAssociationID.NotFound` ([#13382](https://github.com/terraform-providers/terraform-provider-aws/issues/13382))

## 2.62.0 (May 15, 2020)

FEATURES:

* **New Resource:** `aws_workspaces_workspace` ([#11608](https://github.com/terraform-providers/terraform-provider-aws/issues/11608))

ENHANCEMENTS:

* resource/aws_appsync_resolver: Add `cache_config` configuration block ([#12747](https://github.com/terraform-providers/terraform-provider-aws/issues/12747))
* resource/aws_codebuild_project: Support `git_submodules_config` with `GITHUB` and `GITHUB_ENTERPRISE` source types ([#13285](https://github.com/terraform-providers/terraform-provider-aws/issues/13285))
* resource/aws_codebuild_project: Support `SECRETS_MANAGER` environment variable type ([#12572](https://github.com/terraform-providers/terraform-provider-aws/issues/12572))
* resource/aws_datasync_task: Support `ONLY_FILES_TRANSFERRED` value in `verify_mode` argument ([#12897](https://github.com/terraform-providers/terraform-provider-aws/issues/12897))
* resource/aws_iot_topic_rule: Add `dynamodbv2` configuration block ([#7469](https://github.com/terraform-providers/terraform-provider-aws/issues/7469))
* resource/aws_iot_topic_rule: Add `iot_analytics` configuration block ([#9859](https://github.com/terraform-providers/terraform-provider-aws/issues/9859))
* resource/aws_iot_topic_rule: Add `iot_events` configuration block ([#9890](https://github.com/terraform-providers/terraform-provider-aws/issues/9890))
* resource/aws_iot_topic_rule: Add `operation` argument to `dynamodb` configuration block ([#12714](https://github.com/terraform-providers/terraform-provider-aws/issues/12714))
* resource/aws_iot_topic_rule: Add `qos` argument `republish` configuration block ([#12869](https://github.com/terraform-providers/terraform-provider-aws/issues/12869))

BUG FIXES:

* resource/aws_codebuild_project: Allow empty value (`""`) environment variables ([#11572](https://github.com/terraform-providers/terraform-provider-aws/issues/11572))
* resource/aws_security_group_rule: Prevent recreation when `source_security_group_id` refers to a security group across accounts ([#11809](https://github.com/terraform-providers/terraform-provider-aws/issues/11809))

## 2.61.0 (May 08, 2020)

FEATURES:

* **New Data Source:** `aws_ec2_coip_pool` ([#12852](https://github.com/terraform-providers/terraform-provider-aws/issues/12852))
* **New Data Source:** `aws_ec2_coip_pools` ([#12852](https://github.com/terraform-providers/terraform-provider-aws/issues/12852))
* **New Data Source:** `aws_ec2_local_gateway` ([#12764](https://github.com/terraform-providers/terraform-provider-aws/issues/12764))
* **New Data Source:** `aws_ec2_local_gateways` ([#12764](https://github.com/terraform-providers/terraform-provider-aws/issues/12764))
* **New Data Source:** `aws_ec2_local_gateway_route_table` ([#13002](https://github.com/terraform-providers/terraform-provider-aws/issues/13002))
* **New Data Source:** `aws_ec2_local_gateway_route_tables` ([#13002](https://github.com/terraform-providers/terraform-provider-aws/issues/13002))
* **New Resource:** `aws_ec2_transit_gateway_peering_attachment_accepter` ([#11185](https://github.com/terraform-providers/terraform-provider-aws/issues/11185))

ENHANCEMENTS:

* data-source/aws_ebs_volume: Add `multi_attach_enabled` attribute ([#13108](https://github.com/terraform-providers/terraform-provider-aws/issues/13108))
* data-source/aws_efs_file_system: Add `size_in_bytes` attribute ([#13125](https://github.com/terraform-providers/terraform-provider-aws/issues/13125))
* data-source/aws_eip: Add `customer_owned_ip` and `customer_owned_ipv4_pool` attributes ([#12862](https://github.com/terraform-providers/terraform-provider-aws/issues/12862))
* data-source/aws_launch_template: add `partition_number` attribute ([#11655](https://github.com/terraform-providers/terraform-provider-aws/issues/11655))
* resource/aws_api_gateway_deployment: Add `triggers` argument ([#13054](https://github.com/terraform-providers/terraform-provider-aws/issues/13054))
* resource/aws_apigatewayv2_deployment: Add `triggers` argument ([#13055](https://github.com/terraform-providers/terraform-provider-aws/issues/13055))
* resource/aws_ebs_volume: Add `multi_attach_enabled` attribute ([#13108](https://github.com/terraform-providers/terraform-provider-aws/issues/13108))
* resource/aws_eip: Add `customer_owned_ip` attribute and `customer_owned_ipv4_pool` argument ([#12862](https://github.com/terraform-providers/terraform-provider-aws/issues/12862))
* resource/aws_glue_connection: Support `KAFKA` for `connection_type` argument ([#13141](https://github.com/terraform-providers/terraform-provider-aws/issues/13141))
* resource/aws_launch_template: add `partition_number` attribute ([#11655](https://github.com/terraform-providers/terraform-provider-aws/issues/11655))
* resource/aws_launch_template: add `plan time validation to `volume_type`, `spot_instance_type`, `ipv6_addresses`, `ipv4_addresses`, `private_ip_address` ([#11655](https://github.com/terraform-providers/terraform-provider-aws/issues/11655))
* resource/aws_workspaces_directory: Add output attributes for `workspace_security_group_id`, `iam_role_id`, `registration_code`, `directory_name`, `directory_type`, `customer_user_name`, `alias`, `ip_group_ids` and `dns_ip_addresses` ([#13089](https://github.com/terraform-providers/terraform-provider-aws/issues/13089))

BUG FIXES:

* resource/aws_workspaces_directory: Fixes error when removing tags ([#13089](https://github.com/terraform-providers/terraform-provider-aws/issues/13089))

## 2.60.0 (May 01, 2020)

NOTES:

* provider: Region validation now automatically supports the new `eu-south-1` (Europe (Milan)) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g. `data.aws_availability_zones.current: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). ([#12970](https://github.com/terraform-providers/terraform-provider-aws/issues/12970))
* provider: Ignore tags functionality across all data sources and resources (except `aws_autoscaling_group`) via the provider-level `ignore_tags` configuration block has been enabled and this functionality is no longer considered in preview. ([#13039](https://github.com/terraform-providers/terraform-provider-aws/issues/13039))

FEATURES:

* **New Data Source:** `aws_backup_plan` ([#13035](https://github.com/terraform-providers/terraform-provider-aws/issues/13035))
* **New Data Source:** `aws_backup_selection` ([#13035](https://github.com/terraform-providers/terraform-provider-aws/issues/13035))
* **New Data Source:** `aws_backup_vault` ([#13035](https://github.com/terraform-providers/terraform-provider-aws/issues/13035))
* **New Data Source:** `aws_ec2_transit_gateway_peering_attachment` ([#11162](https://github.com/terraform-providers/terraform-provider-aws/issues/11162))
* **New Resource:** `aws_ec2_transit_gateway_peering_attachment` ([#11162](https://github.com/terraform-providers/terraform-provider-aws/issues/11162))
* **New Resource:** `aws_guardduty_organization_admin_account` ([#13034](https://github.com/terraform-providers/terraform-provider-aws/issues/13034))
* **New Resource:** `aws_guardduty_organization_configuration` ([#13034](https://github.com/terraform-providers/terraform-provider-aws/issues/13034))

ENHANCEMENTS:

* data-source/aws_cloudtrail_service_account: Support `eu-south-1` region ([#13061](https://github.com/terraform-providers/terraform-provider-aws/issues/13061))
* data-source/aws_ebs_volume: Add `outpost_arn` attribute ([#12439](https://github.com/terraform-providers/terraform-provider-aws/issues/12439))
* data-source/aws_elastic_beanstalk_hosted_zone: Support `eu-south-1` region ([#13061](https://github.com/terraform-providers/terraform-provider-aws/issues/13061))
* data-source/aws_elb_hosted_zone_id: Add `us-gov-east-1` and `us-gov-west-1` region values ([#12976](https://github.com/terraform-providers/terraform-provider-aws/issues/12976))
* data-source/aws_elb_hosted_zone_id: Support `eu-south-1` region ([#13061](https://github.com/terraform-providers/terraform-provider-aws/issues/13061))
* data-source/aws_elb_service_account: Support `eu-south-1` region ([#13061](https://github.com/terraform-providers/terraform-provider-aws/issues/13061))
* data-source/aws_instance: Add `outpost_arn` attribute ([#12330](https://github.com/terraform-providers/terraform-provider-aws/issues/12330))
* data-source/aws_network_interface: Add `outpost_arn` attribute ([#12440](https://github.com/terraform-providers/terraform-provider-aws/issues/12440))
* data-source/aws_s3_bucket: Support `eu-south-1` region for `hosted_zone_id` attribute ([#13061](https://github.com/terraform-providers/terraform-provider-aws/issues/13061))
* data-source/aws_subnet: Add `outposts_arn` attribute ([#12097](https://github.com/terraform-providers/terraform-provider-aws/issues/12097))
* provider: Support automatic region validation for `eu-south-1` ([#12970](https://github.com/terraform-providers/terraform-provider-aws/issues/12970))
* provider: Implement ignore tags functionality across all data sources and resources (except `aws_autoscaling_group`) ([#13039](https://github.com/terraform-providers/terraform-provider-aws/issues/13039))
* resource/aws_api_gateway_stage: Ignore `NotFoundException` error on destroy ([#12826](https://github.com/terraform-providers/terraform-provider-aws/issues/12826))
* resource/aws_db_snapshot: Support import ([#12978](https://github.com/terraform-providers/terraform-provider-aws/issues/12978))
* resource/aws_default_route_table: Add plan-time validation to `cidr_block` and `ipv6_cidr_block` arguments ([#12858](https://github.com/terraform-providers/terraform-provider-aws/issues/12858))
* resource/aws_default_route_table: Support import ([#13030](https://github.com/terraform-providers/terraform-provider-aws/issues/13030))
* resource/aws_dms_endpoint: Add `kafka_settings` configuration block and `kafka` to `engine_name` argument validation ([#12835](https://github.com/terraform-providers/terraform-provider-aws/issues/12835))
* resource/aws_ebs_volume: Add `outpost_arn` argument ([#12439](https://github.com/terraform-providers/terraform-provider-aws/issues/12439))
* resource/aws_elasticsearch_domain: Support customizable update timeout ([#12916](https://github.com/terraform-providers/terraform-provider-aws/issues/12916))
* resource/aws_glue_connection: Support `MONGODB` for `connection_type` argument ([#13011](https://github.com/terraform-providers/terraform-provider-aws/issues/13011))
* resource/aws_key_pair: Support tag-on-create ([#12962](https://github.com/terraform-providers/terraform-provider-aws/issues/12962))
* resource/aws_instance: Add `outpost_arn` attribute ([#12330](https://github.com/terraform-providers/terraform-provider-aws/issues/12330))
* resource/aws_mq_broker: Support import ([#11841](https://github.com/terraform-providers/terraform-provider-aws/issues/11841))
* resource/aws_network_interface: Add `outpost_arn` attribute ([#12440](https://github.com/terraform-providers/terraform-provider-aws/issues/12440))
* resource/aws_placement_group: Support tag-on-create ([#12963](https://github.com/terraform-providers/terraform-provider-aws/issues/12963))
* resource/aws_route_table: Add plan-time validation to `cidr_block` and `ipv6_cidr_block` arguments ([#12858](https://github.com/terraform-providers/terraform-provider-aws/issues/12858))
* resource/aws_route53_health_check: Support plan-time validation for `reference_name` argument ([#12873](https://github.com/terraform-providers/terraform-provider-aws/issues/12873))
* resource/aws_s3_bucket: Support `eu-south-1` region for `hosted_zone_id` attribute ([#13061](https://github.com/terraform-providers/terraform-provider-aws/issues/13061))
* resource/aws_spot_fleet_request: Add `launch_template_config` configuration block (Support EC2 Launch Templates) ([#12732](https://github.com/terraform-providers/terraform-provider-aws/issues/12732))
* resource/aws_spot_fleet_request: Support import ([#12767](https://github.com/terraform-providers/terraform-provider-aws/issues/12767))
* resource/aws_storagegateway_gateway: Add `gateway_vpc_endpoint` argument ([#9966](https://github.com/terraform-providers/terraform-provider-aws/issues/9966))
* resource/aws_storagegateway_smb_file_share: Add `path` attribute ([#12623](https://github.com/terraform-providers/terraform-provider-aws/issues/12623))
* resource/aws_subnet: Add `outposts_arn` argument ([#12097](https://github.com/terraform-providers/terraform-provider-aws/issues/12097))
* resource/aws_wafregional_xss_match_set: Add plan-time validation for `xss_match_tuple` configuration block arguments ([#13024](https://github.com/terraform-providers/terraform-provider-aws/issues/13024))

BUG FIXES:

* data-source/aws_api_gateway_rest_api: Prevent error with VPC Endpoint configured APIs ([#12825](https://github.com/terraform-providers/terraform-provider-aws/issues/12825))
* resource/aws_appautoscaling_scheduled_action: Prevent error on refresh with multiple resources using the same scheduled action name ([#12699](https://github.com/terraform-providers/terraform-provider-aws/issues/12699))
* resource/aws_batch_job_queue: Prevent panic when `ComputeEnvironmentOrder` is updated outside Terraform ([#12632](https://github.com/terraform-providers/terraform-provider-aws/issues/12632))
* resource/aws_default_route_table: Proper tag on resource creation ([#12858](https://github.com/terraform-providers/terraform-provider-aws/issues/12858))
* resource/aws_efs_file_system: Prevent panic with empty `lifecycle_policy` configuration block ([#12640](https://github.com/terraform-providers/terraform-provider-aws/issues/12640))
* resource/aws_fsx_windows_file_system: Prevent panic when update includes `self_managed_active_directory` settings ([#12630](https://github.com/terraform-providers/terraform-provider-aws/issues/12630))
* resource/aws_glue_catalog_table: Prevent various panics with empty configuration blocks ([#12611](https://github.com/terraform-providers/terraform-provider-aws/issues/12611))
* resource/aws_kinesis_firehose_delivery_stream: Prevent panic with empty `processing_configuration` configuration block ([#12613](https://github.com/terraform-providers/terraform-provider-aws/issues/12613))
* resource/aws_kms_external_key: Prevent `MalformedPolicyDocumentException` errors on creation by retrying for up to 2 minutes to wait for IAM change propagation ([#12863](https://github.com/terraform-providers/terraform-provider-aws/issues/12863))
* resource/aws_kms_key: Prevent `MalformedPolicyDocumentException` errors on creation by retrying for up to 2 minutes to wait for IAM change propagation ([#12863](https://github.com/terraform-providers/terraform-provider-aws/issues/12863))
* resource/aws_lb_listener: Prevent panics on creation and refresh when API throttled ([#12617](https://github.com/terraform-providers/terraform-provider-aws/issues/12617))
* resource/aws_route53_zone: Prevent panic with APIs missing `ChangeInfo` during creation (best effort fix for LocalStack) ([#12634](https://github.com/terraform-providers/terraform-provider-aws/issues/12634))
* resource/aws_storagegateway_gateway: Perform multiple connectivity checks after activation to wait if the underlying server (e.g. EC2 Instance) is automatically rebooted ([#12772](https://github.com/terraform-providers/terraform-provider-aws/issues/12772))
* resource/aws_storagegateway_gateway: Retry 504 status code on activation ([#12773](https://github.com/terraform-providers/terraform-provider-aws/issues/12773))
* resource/aws_wafregional_xss_match_set: Prevent crash with `xss_match_tuple` configuration block since version 2.59.0 ([#13024](https://github.com/terraform-providers/terraform-provider-aws/issues/13024))

## 2.59.0 (April 23, 2020)

NOTES:

* provider: Region validation now automatically supports the new `af-south-1` (Africa (Cape Town)) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g. `data.aws_availability_zones.current: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). ([#12715](https://github.com/terraform-providers/terraform-provider-aws/issues/12715))
* resource/aws_iam_user: The additional `force_destroy` behavior for handling signing certificates requires two additional IAM permissions (`iam:ListSigningCertificates` and `iam:DeleteSigningCertificate`). Restrictive IAM permissions for Terraform runs may require updates. ([#10542](https://github.com/terraform-providers/terraform-provider-aws/issues/10542))
* resource/aws_rds_cluster: Due to recent API support for Aurora MySQL 5.7 and PostgreSQL Global Clusters which implemented the engine mode as `provisioned` instead of the previous `global` for Aurora MySQL 5.6, the resource now requires the `DescribeGlobalClusters` API call. Restrictive IAM permissions may require updates. ([#12867](https://github.com/terraform-providers/terraform-provider-aws/issues/12867))

FEATURES:

* **New Resource:** `aws_apigatewayv2_api_mapping` ([#9461](https://github.com/terraform-providers/terraform-provider-aws/issues/9461))
* **New Resource:** `aws_apigatewayv2_vpc_link` ([#12577](https://github.com/terraform-providers/terraform-provider-aws/issues/12577))

ENHANCEMENTS:

* data_source/aws_acm_certificate: Add `tags` output ([#11659](https://github.com/terraform-providers/terraform-provider-aws/issues/11659))
* data-source/aws_cloudtrail_service_account: Support `af-south-1` region ([#12967](https://github.com/terraform-providers/terraform-provider-aws/issues/12967))
* data-source/aws_elastic_beanstalk_hosted_zone: Support `af-south-1` region ([#12967](https://github.com/terraform-providers/terraform-provider-aws/issues/12967))
* data-source/aws_elb_hosted_zone_id: Support `af-south-1` region ([#12967](https://github.com/terraform-providers/terraform-provider-aws/issues/12967))
* data-source/aws_elb_service_account: Support `af-south-1` region ([#12967](https://github.com/terraform-providers/terraform-provider-aws/issues/12967))
* data-source/aws_s3_bucket: Support `af-south-1` region for `hosted_zone_id` attribute ([#12967](https://github.com/terraform-providers/terraform-provider-aws/issues/12967))
* provider: Support automatic region validation for `af-south-1` ([#12715](https://github.com/terraform-providers/terraform-provider-aws/issues/12715))
* resource/aws_apigatewayv2_api: Add `cors_configuration`, `credentials_arn`, `route_key` and `target` attributes ([#12452](https://github.com/terraform-providers/terraform-provider-aws/issues/12452))
* resource/aws_appsync_graphql_api: Add `log_config` configuration block `exclude_verbose_content` argument ([#12884](https://github.com/terraform-providers/terraform-provider-aws/issues/12884))
* resource/aws_config_configuration_recorder: Prevent error during deletion operation when resource is missing ([#12734](https://github.com/terraform-providers/terraform-provider-aws/issues/12734))
* resource/aws_default_network_acl: Support import ([#12924](https://github.com/terraform-providers/terraform-provider-aws/issues/12924))
* resource/aws_lambda_alias: Suppress differences for equivalent `function_name` argument values of name versus ARN ([#12902](https://github.com/terraform-providers/terraform-provider-aws/issues/12902))
* resource/aws_network_acl_rule: Support import ([#12921](https://github.com/terraform-providers/terraform-provider-aws/issues/12921))
* resource/aws_route: Add plan-time validation for `destination_cidr_block` and `destination_ipv6_cidr_block` arguments ([#12890](https://github.com/terraform-providers/terraform-provider-aws/issues/12890))
* resource/aws_s3_bucket: Support `af-south-1` region for `hosted_zone_id` attribute ([#12967](https://github.com/terraform-providers/terraform-provider-aws/issues/12967))
* resource/aws_service_discovery_private_dns_namespace: Support import ([#12929](https://github.com/terraform-providers/terraform-provider-aws/issues/12929))
* resource/aws_ssm_activation: Support import ([#12933](https://github.com/terraform-providers/terraform-provider-aws/issues/12933))
* resource/aws_ssm_maintenance_window_target: Add plan-time validation to `resource_type` argument ([#11783](https://github.com/terraform-providers/terraform-provider-aws/issues/11783))
* resource/aws_ssm_maintenance_window_target: Support import ([#12935](https://github.com/terraform-providers/terraform-provider-aws/issues/12935))
* resource/aws_volume_attachment: Support import ([#12948](https://github.com/terraform-providers/terraform-provider-aws/issues/12948))
* resource/aws_waf_ipset: Add plan-time validation for `ip_set_descriptors` configuration block arguments ([#12775](https://github.com/terraform-providers/terraform-provider-aws/issues/12775))
* resource/aws_waf_sql_injection_match_set: Support import ([#11657](https://github.com/terraform-providers/terraform-provider-aws/issues/11657))
* resource/aws_waf_xss_match_set: Add plan-time validation for `xss_match_tuples` configuration block arguments ([#12777](https://github.com/terraform-providers/terraform-provider-aws/issues/12777))
* resource/aws_wafregional_web_acl: Add plan-time validation to various arguments ([#12793](https://github.com/terraform-providers/terraform-provider-aws/issues/12793))

BUG FIXES:

* data-source/aws_launch_template: Prevent type error with `network_interfaces` `associate_public_ip_address` attribute ([#12936](https://github.com/terraform-providers/terraform-provider-aws/issues/12936))
* resource/aws_glue_security_configuration: Prevent empty string KMS Key ARN in S3 Encryption settings ([#12898](https://github.com/terraform-providers/terraform-provider-aws/issues/12898))
* resource/aws_iam_user: Ensure `force_destroy` argument removes signing certificates when enabled ([#10542](https://github.com/terraform-providers/terraform-provider-aws/issues/10542))
* resource/aws_rds_cluster: Prevent unexpected `global_cluster_identifier` differences and deletion error with `aurora-mysql` and `aurora-postgresql` Global Cluster members ([#12867](https://github.com/terraform-providers/terraform-provider-aws/issues/12867))
* resource/aws_route: Prevent not found after creation error with `destination_ipv6_cidr_block` set to `::0/0` ([#12890](https://github.com/terraform-providers/terraform-provider-aws/issues/12890))

## 2.58.0 (April 16, 2020)

FEATURES:

* **New Data Source:** `aws_regions` ([#12269](https://github.com/terraform-providers/terraform-provider-aws/issues/12269))
* **New Resource:** `aws_apigatewayv2_deployment` ([#9245](https://github.com/terraform-providers/terraform-provider-aws/issues/9245))
* **New Resource:** `aws_apigatewayv2_domain_name` ([#9391](https://github.com/terraform-providers/terraform-provider-aws/issues/9391))
* **New Resource:** `aws_apigatewayv2_integration_response` ([#9365](https://github.com/terraform-providers/terraform-provider-aws/issues/9365))
* **New Resource:** `aws_apigatewayv2_route` ([#8881](https://github.com/terraform-providers/terraform-provider-aws/issues/8881))
* **New Resource:** `aws_apigatewayv2_route_response` ([#9373](https://github.com/terraform-providers/terraform-provider-aws/issues/9373))
* **New Resource:** `aws_apigatewayv2_stage` ([#9232](https://github.com/terraform-providers/terraform-provider-aws/issues/9232))
* **New Resource:** `aws_dms_event_subscription` ([#7170](https://github.com/terraform-providers/terraform-provider-aws/issues/7170))

ENHANCEMENTS:

* data-source/aws_dynamodb_table: Add `replica` attribute (initial support for Global Tables V2 (version 2019.11.21)) ([#12342](https://github.com/terraform-providers/terraform-provider-aws/issues/12342))
* data-source/aws_instance: Exports `volume_name` for `root_block_device` ([#12620](https://github.com/terraform-providers/terraform-provider-aws/issues/12620))
* resource/aws_backup_plan: Add `rule` configuration block `copy_action` configuration block (support cross region copy) ([#11923](https://github.com/terraform-providers/terraform-provider-aws/issues/11923))
* resource/aws_cognito_identity_provider: Support plan-time validation for `idp_identifiers`, `provider_name`, and `provider_type` arguments ([#10705](https://github.com/terraform-providers/terraform-provider-aws/issues/10705))
* resource/aws_dms_endpoint: Add `elasticsearch_settings` configuration block and `elasticsearch` to `engine_name` validation (support Elasticsearch endpoints) ([#11792](https://github.com/terraform-providers/terraform-provider-aws/issues/11792))
* resource/aws_dms_endpoint: Add `kinesis_settings` configuration block and `kinesis` to `engine_name` validation (support Kinesis endpoints) ([#8633](https://github.com/terraform-providers/terraform-provider-aws/issues/8633))
* resource/aws_dynamodb_table: Add `replica` configuration block (initial support for Global Tables V2 (version 2019.11.21)) ([#12342](https://github.com/terraform-providers/terraform-provider-aws/issues/12342))
* resource/aws_ec2_client_vpn_endpoint: Allow two `authentication_options`  configuration blocks ([#12819](https://github.com/terraform-providers/terraform-provider-aws/issues/12819))
* resource/aws_instance: Allow changing root volume size without re-creating resource ([#12620](https://github.com/terraform-providers/terraform-provider-aws/issues/12620))
* resource/aws_instance: Exports `volume_name` for `root_block_device` ([#12620](https://github.com/terraform-providers/terraform-provider-aws/issues/12620))

BUG FIXES:

* resource/aws_dlm_lifecycle_policy: Ensure plan-time validation for `times` argument only allows 24 hour format ([#12800](https://github.com/terraform-providers/terraform-provider-aws/issues/12800))

## 2.57.0 (April 09, 2020)

BREAKING CHANGES:

* provider: The configuration for the preview ignore tags functionality has been updated to include a wrapping configuration block. For example:

```hcl
provider "aws" {
  ignore_tags {
    keys = ["TagKey1"]
  }
}
```

FEATURES:

* **New Data Source:** `aws_cloudfront_distribution` ([#6468](https://github.com/terraform-providers/terraform-provider-aws/issues/6468))
* **New Resource:** `aws_apigatewayv2_authorizer` ([#9228](https://github.com/terraform-providers/terraform-provider-aws/issues/9228))
* **New Resource:** `aws_apigatewayv2_integration` ([#8949](https://github.com/terraform-providers/terraform-provider-aws/issues/8949))
* **New Resource:** `aws_apigatewayv2_model` ([#8912](https://github.com/terraform-providers/terraform-provider-aws/issues/8912))

ENHANCEMENTS:

* data-source/aws_lambda_layer_version: Support plan-time validation for `compatible_runtime` argument `dotnetcore3.1` value (support .NET Core 3.1) ([#12712](https://github.com/terraform-providers/terraform-provider-aws/issues/12712))
* resource/aws_cloudhsm_v2_cluster: Support tag-on-create ([#11683](https://github.com/terraform-providers/terraform-provider-aws/issues/11683))
* resource/aws_docdb_cluster: Add `deletion_protection` argument ([#12650](https://github.com/terraform-providers/terraform-provider-aws/issues/12650))
* resource/aws_egress_only_internet_gateway: Add `tags` argument ([#11568](https://github.com/terraform-providers/terraform-provider-aws/issues/11568))
* resource/aws_lambda_function: Support plan-time validation for `runtime` argument `dotnetcore3.1` value (support .NET Core 3.1) ([#12712](https://github.com/terraform-providers/terraform-provider-aws/issues/12712))
* resource/aws_lambda_layer_version: Support plan-time validation for `compatible_runtimes` argument `dotnetcore3.1` value (support .NET Core 3.1) ([#12712](https://github.com/terraform-providers/terraform-provider-aws/issues/12712))
* resource/aws_rds_global_cluster: Add `aurora-postgresql` to `engine` argument plan-time validation ([#12401](https://github.com/terraform-providers/terraform-provider-aws/issues/12401))
* resource/aws_redshift_snapshot_copy_grant: Support resource import ([#10350](https://github.com/terraform-providers/terraform-provider-aws/issues/10350))
* resource/aws_spot_fleet_request: Add `tags` argument (support tagging of Spot Fleet Request itself) ([#12295](https://github.com/terraform-providers/terraform-provider-aws/issues/12295))
* resource/aws_spot_fleet_request: Support plan-time validation for `launch_specification` configuration block `ebs_block_device` `volume_type`, `iam_instance_profile_arn`, `placement_tenancy`, and `root_block_device` `volume_type` arguments ([#12295](https://github.com/terraform-providers/terraform-provider-aws/issues/12295))
* resource/aws_spot_fleet_request: Support plan-time validation for `allocation_strategy`, `instance_interruption_behaviour`, and `target_group_arns` arguments ([#12295](https://github.com/terraform-providers/terraform-provider-aws/issues/12295))
* service/ec2: Prevent eventual consistency errors tagging resources on creation ([#12735](https://github.com/terraform-providers/terraform-provider-aws/issues/12735))

BUG FIXES:

* resource/aws_appautoscaling_policy: Fix error when importing DynamoDB Table Index policy ([#11232](https://github.com/terraform-providers/terraform-provider-aws/issues/11232))
* resource/aws_db_instance: Allow creating read replica into RAM shared Subnet with VPC Security Group ([#12700](https://github.com/terraform-providers/terraform-provider-aws/issues/12700))
* resource/aws_kms_key: Prevent eventual consistency related errors on creation ([#12738](https://github.com/terraform-providers/terraform-provider-aws/issues/12738))
* resource/aws_lb_target_group: Automatically propose resource recreation for TCP `protocol` Target Groups when `health_check` configuration block `interval`, `protocol`, or `timeout` argument values are updated ([#4568](https://github.com/terraform-providers/terraform-provider-aws/issues/4568))

## 2.56.0 (April 03, 2020)

NOTES:

* resource/aws_emr_cluster: The bug fix in this release will potentially re-create EMR Clusters with multiple bootstrap actions, since bootstrap actions cannot be modified in place. To avoid re-creation, temporarily add the [`ignore_changes` lifecycle configuration argument](https://www.terraform.io/docs/configuration/resources.html#ignore_changes) and/or update the order in your Terraform configuration.

ENHANCEMENTS:

* data-source/aws_launch_template: Add `hibernation_options` attribute ([#12492](https://github.com/terraform-providers/terraform-provider-aws/issues/12492))
* resource/aws_codepipeline: Adds cross-region action support ([#12549](https://github.com/terraform-providers/terraform-provider-aws/issues/12549))
* resource/aws_dx_connection: Support `2Gbps` and `5Gbps` values in plan-time validation for `bandwidth` argument ([#12559](https://github.com/terraform-providers/terraform-provider-aws/issues/12559))
* resource/aws_dx_lag: Support `2Gbps` and `5Gbps` values in plan-time validation for `bandwidth` argument ([#12559](https://github.com/terraform-providers/terraform-provider-aws/issues/12559))
* resource/aws_elastic_transcoder_preset: Support plan-time validation for `role` argument ([#12575](https://github.com/terraform-providers/terraform-provider-aws/issues/12575))
* resource/aws_kms_grant: Support resource import ([#11991](https://github.com/terraform-providers/terraform-provider-aws/issues/11991))
* resource/aws_launch_template: Add `hibernation_options` configuration block ([#12492](https://github.com/terraform-providers/terraform-provider-aws/issues/12492))

BUG FIXES:

* resource/aws_codedeploy_deployment_group: Fix `blue_green_deployment_config` updates for ECS ([#11885](https://github.com/terraform-providers/terraform-provider-aws/issues/11885))
* resource/aws_emr_cluster: Now properly sets the order when multiple bootstrap actions are defined
* resource/aws_kms_grant: Remove resource from Terraform state instead of error if removed outside Terraform ([#12560](https://github.com/terraform-providers/terraform-provider-aws/issues/12560))
* resource/aws_s3_bucket: Prevent various panics with empty configuration blocks ([#12614](https://github.com/terraform-providers/terraform-provider-aws/issues/12614))
* resource/aws_volume_attachment: Ensure any error is shown while waiting for volume to detach ([#12596](https://github.com/terraform-providers/terraform-provider-aws/issues/12596))

## 2.55.0 (March 27, 2020)

FEATURES:

* **New Resource:** `aws_ec2_availability_zone_group` ([#12400](https://github.com/terraform-providers/terraform-provider-aws/issues/12400))

ENHANCEMENTS:

* data-source/aws_availability_zone: Add `all_availability_zones` and `filter` arguments ([#12400](https://github.com/terraform-providers/terraform-provider-aws/issues/12400))
* data-source/aws_availability_zone: Add `group_name`, `network_border_group`, and `opt_in_status` attributes ([#12400](https://github.com/terraform-providers/terraform-provider-aws/issues/12400))
* data-source/aws_availability_zones: Add `all_availability_zones` and `filter` arguments ([#12400](https://github.com/terraform-providers/terraform-provider-aws/issues/12400))
* data-source/aws_availability_zones: Add `group_names` attribute ([#12400](https://github.com/terraform-providers/terraform-provider-aws/issues/12400))
* data-source/aws_ec2_transit_gateway_dx_gateway_attachement: Add `filter` and `tags` arguments ([#12516](https://github.com/terraform-providers/terraform-provider-aws/issues/12516))
* data-source/aws_ec2_transit_gateway_vpn_attachment: Add `filter` and `tags` arguments ([#12415](https://github.com/terraform-providers/terraform-provider-aws/issues/12415))
* data-source/aws_instance: Add `metadata_options` attribute ([#12491](https://github.com/terraform-providers/terraform-provider-aws/issues/12491))
* data-source/aws_launch_template: Add `filter` and `tags` arguments ([#12403](https://github.com/terraform-providers/terraform-provider-aws/issues/12403))
* data-source/aws_launch_template: Add `metadata_options` attribute ([#12491](https://github.com/terraform-providers/terraform-provider-aws/issues/12491))
* data-source/aws_prefix_list: Add `filter` argument ([#12416](https://github.com/terraform-providers/terraform-provider-aws/issues/12416))
* data-source/aws_vpc_endpoint_service: Add `filter` and `tags` arguments ([#12404](https://github.com/terraform-providers/terraform-provider-aws/issues/12404))
* resource/aws_athena_workgroup: Add `force_destroy` argument ([#12254](https://github.com/terraform-providers/terraform-provider-aws/issues/12254))
* resource/aws_cloudwatch_log_metric_filter: Support resource import ([#11992](https://github.com/terraform-providers/terraform-provider-aws/issues/11992))
* resource/aws_flow_log: Add `max_aggregation_interval` argument ([#12483](https://github.com/terraform-providers/terraform-provider-aws/issues/12483))
* resource/aws_instance: Add `metadata_options` configuration block (support IMDSv2) ([#12491](https://github.com/terraform-providers/terraform-provider-aws/issues/12491))
* resource/aws_launch_template: Add `metadata_options` configuration block (support IMDSv2) ([#12491](https://github.com/terraform-providers/terraform-provider-aws/issues/12491))
* resource/aws_msk_cluster: Add `logging_info` configuration block (support CloudWatch, Firehose, and S3 logging) ([#12215](https://github.com/terraform-providers/terraform-provider-aws/issues/12215))
* resource/aws_mq_configuration: Support plan-time validation for `engine_type` argument ([#11843](https://github.com/terraform-providers/terraform-provider-aws/issues/11843))
* resource/aws_route53_health_check: A dd plan-time validation to `insufficient_data_health_status` ([#12305](https://github.com/terraform-providers/terraform-provider-aws/issues/12305))
* resource/aws_storagegateway_nfs_file_share: Add `path` attribute ([#12530](https://github.com/terraform-providers/terraform-provider-aws/issues/12530))

BUG FIXES:

* resource/aws_db_instance: Allow restoring from snapshot into RAM shared Subnet with VPC Security Group ([#12447](https://github.com/terraform-providers/terraform-provider-aws/issues/12447))
* resource/aws_mq_configuration: Remove extraneous `ListTags` API call during refresh ([#11843](https://github.com/terraform-providers/terraform-provider-aws/issues/11843))
* resource/aws_neptune_cluster_instance: Add missing `configuring-log-exports` as allowed pending state ([#12079](https://github.com/terraform-providers/terraform-provider-aws/issues/12079))
* resource/aws_route53_health_check: Do not recreate health check when using compressed ipv6 address ([#12305](https://github.com/terraform-providers/terraform-provider-aws/issues/12305))

## 2.54.0 (March 19, 2020)

FEATURES:

* **New Resource:** `aws_kinesis_video_stream` ([#8291](https://github.com/terraform-providers/terraform-provider-aws/issues/8291))
* **New Resource:** `aws_securityhub_member` ([#6975](https://github.com/terraform-providers/terraform-provider-aws/issues/6975))

ENHANCEMENTS:

* data-source/aws_iam_role: Add `tags` attribute ([#12349](https://github.com/terraform-providers/terraform-provider-aws/issues/12349))
* data-source/aws_lb: Add `drop_invalid_header_fields` attribute ([#11257](https://github.com/terraform-providers/terraform-provider-aws/issues/11257))
* provider: Support AWS shared configuration file `duration_seconds` setting for assume role ([#12359](https://github.com/terraform-providers/terraform-provider-aws/issues/12359))
* resource/aws_backup_plan: Support resource import ([#12381](https://github.com/terraform-providers/terraform-provider-aws/issues/12381))
* resource/aws_cognito_user_pool: Add `email_configuration` configuration block `from_email_address` argument ([#11607](https://github.com/terraform-providers/terraform-provider-aws/issues/11607))
* resource/aws_cognito_user_pool: Add `username_configuration` configuration block (Support case insensitive usernames) ([#12317](https://github.com/terraform-providers/terraform-provider-aws/issues/12317))
* resource/aws_cognito_user_pool_client: Add `analytics_configuration` configuration block (Support Pinpoint analytics) ([#11762](https://github.com/terraform-providers/terraform-provider-aws/issues/11762))
* resource/aws_cognito_user_pool_client: Add `prevent_user_existence_errors` argument ([#11604](https://github.com/terraform-providers/terraform-provider-aws/issues/11604))
* resource/aws_dlm_lifecycle_policy: Support plan-time validation for 1 hour schedules in `policy_details` `schedule` `create_rule` `interval` argument ([#12327](https://github.com/terraform-providers/terraform-provider-aws/issues/12327))
* resource/aws_inspector_assessment_template: Add `tags` argument ([#12375](https://github.com/terraform-providers/terraform-provider-aws/issues/12375))
* resource/aws_inspector_assessment_template: Support resource import ([#12375](https://github.com/terraform-providers/terraform-provider-aws/issues/12375))
* resource/aws_lambda_function: Support plan-time validation for `handler` argument ([#12411](https://github.com/terraform-providers/terraform-provider-aws/issues/12411))
* resource/aws_lb: Add `drop_invalid_header_fields` argument ([#11257](https://github.com/terraform-providers/terraform-provider-aws/issues/11257))
* resource/aws_nat_gateway: Support tag-on-create ([#12347](https://github.com/terraform-providers/terraform-provider-aws/issues/12347))
* resource/aws_opsworks_application: Support resource import ([#12383](https://github.com/terraform-providers/terraform-provider-aws/issues/12383))
* resource/aws_opsworks_application: Add plan-time validation to `data_source_arn` and `data_source_type` arguments and `app_source` configuration block `type` argument ([#12383](https://github.com/terraform-providers/terraform-provider-aws/issues/12383))
* resource/aws_opsworks_custom_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_ganglia_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_haproxy_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_java_app_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_memcached_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_mysql_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_nodejs_app_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_php_app_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_rails_app_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_opsworks_static_web_layer: Add `tags` argument, `arn` attribute, and plan-time validation to `custom_instance_profile_arn` argument ([#11667](https://github.com/terraform-providers/terraform-provider-aws/issues/11667))
* resource/aws_vpc_dhcp_options_association: Support resource import ([#7252](https://github.com/terraform-providers/terraform-provider-aws/issues/7252))

BUG FIXES:

* resource/aws_api_gateway_rest_api: Ignore ordering differences for `endpoint_configuration` configuration block `vpc_endpoint_ids` argument ([#12350](https://github.com/terraform-providers/terraform-provider-aws/issues/12350))
* resource/aws_backup_selection: Automatically retry on additional IAM Role eventual consistency error ([#10687](https://github.com/terraform-providers/terraform-provider-aws/issues/10687))
* resource/aws_backup_vault: Remove resource from Terraform state when deleted outside Terraform ([#11845](https://github.com/terraform-providers/terraform-provider-aws/issues/11845))
* resource/aws_cognito_user_pool_client: Ignore ordering differences for `callback_urls`, `logout_urls`, and `supported_identity_providers` arguments ([#12388](https://github.com/terraform-providers/terraform-provider-aws/issues/12388))
* resource/aws_ebs_snapshot_copy: Return API errors instead of panic if unable to read snapshot ([#12283](https://github.com/terraform-providers/terraform-provider-aws/issues/12283))
* resource/aws_kinesis_stream: Ensure `kms_key_id` argument in-place updates complete successfully ([#12008](https://github.com/terraform-providers/terraform-provider-aws/issues/12008))
* resource/aws_lambda_alias: Propose resource recreation for `function_name` argument updates ([#11170](https://github.com/terraform-providers/terraform-provider-aws/issues/11170))
* resource/aws_opsworks_application: Mark `app_source` configuration block `ssh_key` argument as sensitive ([#11984](https://github.com/terraform-providers/terraform-provider-aws/issues/11984))
* resource/aws_opsworks_stack: Mark `custom_cookbooks_source` configuration block `ssh_key` argument as sensitive ([#11984](https://github.com/terraform-providers/terraform-provider-aws/issues/11984))
* resource/aws_s3_bucket: Retry `NoSuchBucket` error when setting tags during resource creation ([#12418](https://github.com/terraform-providers/terraform-provider-aws/issues/12418))

## 2.53.0 (March 12, 2020)

NOTES:

* resource/aws_cognito_user_pool: The addition of Software Token MFA support required the use of new `GetUserPoolMfaConfig` and `SetUserPoolMfaConfig` API calls. Restrictive IAM permissions for Terraform may require updates.  ([#12358](https://github.com/terraform-providers/terraform-provider-aws/issues/12358))

FEATURES:

* **New Resource:** `aws_apigatewayv2_api` ([#8842](https://github.com/terraform-providers/terraform-provider-aws/issues/8842))


ENHANCEMENTS:

* resource/aws_appsync_graphql_api: Add `xray_enabled` argument ([#11972](https://github.com/terraform-providers/terraform-provider-aws/issues/11972))
* resource/aws_cloud9_environment_ec2: Add `tags` argument ([#12132](https://github.com/terraform-providers/terraform-provider-aws/issues/12132))
* resource/aws_cognito_user_pool: Add `software_token_mfa_configuration` configuration block (Support Time-based One-Time Password (TOTP) Multi-Factor Authentication) ([#12358](https://github.com/terraform-providers/terraform-provider-aws/issues/12358))
* resource/aws_ec2_traffic_mirror_filter: Add `tags` argument ([#12133](https://github.com/terraform-providers/terraform-provider-aws/issues/12133))
* resource/aws_ec2_traffic_mirror_session: Add `tags` argument ([#12134](https://github.com/terraform-providers/terraform-provider-aws/issues/12134))
* resource/aws_ec2_traffic_mirror_target: Add `tags` argument and `network_load_balancer_arn` plan-time validation ([#12135](https://github.com/terraform-providers/terraform-provider-aws/issues/12135))
* resource/aws_flow_log: Add `tags` argument ([#12273](https://github.com/terraform-providers/terraform-provider-aws/issues/12273))
* resource/aws_flow_log: Add `iam_role_arn` and `log_destination` plan-time validation ([#12273](https://github.com/terraform-providers/terraform-provider-aws/issues/12273))
* resource/aws_globalaccelerator_accelerator: Add `tags` argument ([#12309](https://github.com/terraform-providers/terraform-provider-aws/issues/12309))
* resource/aws_vpc_endpoint: Support tag-on-create ([#12288](https://github.com/terraform-providers/terraform-provider-aws/issues/12288))
* resource/aws_vpc_endpoint_service: Support tag-on-create and add `network_load_balancer_arns` plan-time validation ([#12290](https://github.com/terraform-providers/terraform-provider-aws/issues/12290))

BUG FIXES:

* resource/aws_vpn_gateway: Automatically retry on `DetachVpnGateway` calls receiving `InvalidParameterValue: This call cannot be completed because there are pending VPNs or Virtual Interfaces` ([#11720](https://github.com/terraform-providers/terraform-provider-aws/issues/11720))
* resource/aws_vpn_gateway_attachment: Automatically retry on `DetachVpnGateway` calls receiving `InvalidParameterValue: This call cannot be completed because there are pending VPNs or Virtual Interfaces` ([#11720](https://github.com/terraform-providers/terraform-provider-aws/issues/11720))

## 2.52.0 (March 06, 2020)

FEATURES:

* **New Data Source:** `aws_ec2_instance_type_offering` ([#12139](https://github.com/terraform-providers/terraform-provider-aws/issues/12139))
* **New Data Source:** `aws_ec2_instance_type_offerings` ([#12139](https://github.com/terraform-providers/terraform-provider-aws/issues/12139))

ENHANCEMENTS:

* resource/aws_eks_cluster: Add `encryption_config` configuration block ([#12280](https://github.com/terraform-providers/terraform-provider-aws/issues/12280))
* resource/aws_globalaccelerator_accelerator: Add `dns_name` and `hosted_zone_id` attributes ([#11670](https://github.com/terraform-providers/terraform-provider-aws/issues/11670))
* resource/aws_lb_target_group: Add `load_balancing_algorithm_type` argument (support Least Outstanding Requests algorithm for Application Load Balancers) ([#11141](https://github.com/terraform-providers/terraform-provider-aws/issues/11141))
* resource/aws_s3_bucket: Add `grant` to implement ACL policy grants ([#3728](https://github.com/terraform-providers/terraform-provider-aws/issues/3728))

BUG FIXES:

* resource/aws_iam_service_linked_role: Allow `aws_service_name` argument validation to accept values in AWS partitions outside AWS Commercial and AWS GovCloud (US) ([#11919](https://github.com/terraform-providers/terraform-provider-aws/issues/11919))
* resource/aws_lambda_function_event_invoke_config: Retry on additional IAM eventual consistency error with SNS Topic destinations ([#12171](https://github.com/terraform-providers/terraform-provider-aws/issues/12171))
* resource/aws_media_store_container: Prevent `ValidationException` error on creation when no tags are configured ([#12170](https://github.com/terraform-providers/terraform-provider-aws/issues/12170))

## 2.51.0 (February 28, 2020)

FEATURES:

* **New Data Source:** `aws_sfn_activity` ([#11080](https://github.com/terraform-providers/terraform-provider-aws/issues/11080))
* **New Data Source:** `aws_sfn_state_machine` ([#10932](https://github.com/terraform-providers/terraform-provider-aws/issues/10932))
* **New Resource:** `aws_ec2_traffic_mirror_filter` ([#9372](https://github.com/terraform-providers/terraform-provider-aws/issues/9372))
* **New Resource:** `aws_ec2_traffic_mirror_filter_rule` ([#9372](https://github.com/terraform-providers/terraform-provider-aws/issues/9372))
* **New Resource:** `aws_ec2_traffic_mirror_session` ([#9372](https://github.com/terraform-providers/terraform-provider-aws/issues/9372))
* **New Resource:** `aws_ec2_traffic_mirror_target` ([#9372](https://github.com/terraform-providers/terraform-provider-aws/issues/9372))
* **New Resource:** `aws_s3_access_point` ([#11276](https://github.com/terraform-providers/terraform-provider-aws/issues/11276))

ENHANCEMENTS:

* data-source/aws_lambda_layer_version: Support plan-time validation for `compatible_runtime` argument `ruby2.7` value ([#12116](https://github.com/terraform-providers/terraform-provider-aws/issues/12116))
* resource/aws_dx_hosted_private_virtual_interface: Add `amazon_side_asn` attribute ([#11415](https://github.com/terraform-providers/terraform-provider-aws/issues/11415))
* resource/aws_dx_hosted_public_virtual_interface: Add `amazon_side_asn` attribute ([#11415](https://github.com/terraform-providers/terraform-provider-aws/issues/11415))
* resource/aws_dx_hosted_transit_virtual_interface: Add `amazon_side_asn` attribute ([#11415](https://github.com/terraform-providers/terraform-provider-aws/issues/11415))
* resource/aws_dx_private_virtual_interface: Add `amazon_side_asn` attribute ([#11415](https://github.com/terraform-providers/terraform-provider-aws/issues/11415))
* resource/aws_dx_public_virtual_interface: Add `amazon_side_asn` attribute ([#11415](https://github.com/terraform-providers/terraform-provider-aws/issues/11415))
* resource/aws_dx_transit_virtual_interface: Add `amazon_side_asn` attribute ([#11415](https://github.com/terraform-providers/terraform-provider-aws/issues/11415))
* resource/aws_glue_job: Add `notification_property` configuration block ([#12115](https://github.com/terraform-providers/terraform-provider-aws/issues/12115))
* resource/aws_lambda_event_source_mapping: Add `bisect_batch_on_function_error`, `maximum_record_age_in_seconds`, `maximum_retry_attempts`, and `parallelization_factor` arguments ([#11100](https://github.com/terraform-providers/terraform-provider-aws/issues/11100))
* resource/aws_lambda_event_source_mapping: Add `destination_config` configuration block ([#11100](https://github.com/terraform-providers/terraform-provider-aws/issues/11100))
* resource/aws_lambda_function: Support plan-time validation for `runtime` argument `ruby2.7` value ([#12116](https://github.com/terraform-providers/terraform-provider-aws/issues/12116))
* resource/aws_lambda_layer_version: Support plan-time validation for `compatible_runtimes` argument `ruby2.7` value ([#12116](https://github.com/terraform-providers/terraform-provider-aws/issues/12116))
* resource/aws_msk_cluster: Support in-place updates to `enhanced_monitoring` and `number_of_broker_nodes` arguments ([#11451](https://github.com/terraform-providers/terraform-provider-aws/issues/11451))
* resource/aws_msk_cluster: Add `open_monitoring` configuration block (support Prometheus monitoring configuration) ([#11451](https://github.com/terraform-providers/terraform-provider-aws/issues/11451))

BUG FIXES:

* resource/aws_workspaces_directory: Prevent panic and remove resource from Terraform state if removed outside Terraform ([#11837](https://github.com/terraform-providers/terraform-provider-aws/issues/11837))

## 2.50.0 (February 20, 2020)

NOTES:

* resource/aws_lambda_function: The `publish` argument now will also publish versions for configuration updates. This is accomplished via a separate `PublishVersion` API call, where before the publishing only occurred via the `Publish` parameter of the `UpdateFunctionCode` API call. Restrictive IAM permissions for Terraform may require updates. ([#11211](https://github.com/terraform-providers/terraform-provider-aws/issues/11211))
* resource/aws_ram_resource_share_accepter: The `status` attribute now reflects the status of the RAM Resource Share and not the RAM Resource Share Invitation (which expires after 7 days). ([#11562](https://github.com/terraform-providers/terraform-provider-aws/issues/11562))

FEATURES:

* **New Data Source:** `aws_lambda_alias` ([#9490](https://github.com/terraform-providers/terraform-provider-aws/issues/9490))

ENHANCEMENTS:

* resource/aws_appmesh_route: Add `priority` and `header` attributes to support route priorities and HTTP header-based routing ([#10402](https://github.com/terraform-providers/terraform-provider-aws/issues/10402))
* resource/aws_iam_access_key: Add `ses_smtp_password_v4` attribute (add per-region SigV4 support) ([#11144](https://github.com/terraform-providers/terraform-provider-aws/issues/11144))
* resource/aws_security_group: Support import of `name_prefix` argument ([#12052](https://github.com/terraform-providers/terraform-provider-aws/issues/12052))
* resource/aws_transfer_server: Add `host_key` argument and `host_key_fingerprint` attribute ([#8913](https://github.com/terraform-providers/terraform-provider-aws/issues/8913))

BUG FIXES:

* resource/aws_lambda_function: If `publish` argument is enabled, also publish new versions on function configuration-only updates in addition to function code updates ([#11211](https://github.com/terraform-providers/terraform-provider-aws/issues/11211))
* resource/aws_lambda_permission: Fix error when Lambda permission is deleted out-of-band ([#11924](https://github.com/terraform-providers/terraform-provider-aws/issues/11924))
* resource/aws_ram_resource_share_accepter: Fix read operations after the RAM Resource Share Invitation is no longer present after 7 days ([#11562](https://github.com/terraform-providers/terraform-provider-aws/issues/11562))

## 2.49.0 (February 14, 2020)

FEATURES:

* **New Resource:** `aws_codestarnotifications_notification_rule` ([#10991](https://github.com/terraform-providers/terraform-provider-aws/issues/10991))
* **New Resource:** `aws_s3_bucket_analytics_configuration` ([#11874](https://github.com/terraform-providers/terraform-provider-aws/issues/11874))

ENHANCEMENTS:

* data-source/aws_api_gateway_rest_api: Add `api_key_source`, `arn`, `binary_media_types`, `description`, `endpoint_configuration`, `execution_arn`, `minimum_compression_size`, `policy`, and `tags` attributes ([#10971](https://github.com/terraform-providers/terraform-provider-aws/issues/10971))
* resource/aws_db_instance: Support `agent` value in `enable_cloudwatch_logs_exports` argument plan-time validation (Support MSSQL agent log) ([#11472](https://github.com/terraform-providers/terraform-provider-aws/issues/11472))
* resource/aws_db_instance: Add `delete_automated_backups` argument ([#8461](https://github.com/terraform-providers/terraform-provider-aws/issues/8461))
* resource/aws_gamelift_fleet: Add `tags` argument ([#11559](https://github.com/terraform-providers/terraform-provider-aws/issues/11559))
* resource/aws_instance: Add `hibernation` argument ([#6961](https://github.com/terraform-providers/terraform-provider-aws/issues/6961))
* resource/aws_launch_template: Add `cpu_options` configuration block (support disabling multithreading) ([#6552](https://github.com/terraform-providers/terraform-provider-aws/issues/6552))
* resource/aws_neptune_cluster: Add `enable_cloudwatch_logs_exports` argument (support audit logging) ([#11949](https://github.com/terraform-providers/terraform-provider-aws/issues/11949))
* resource/aws_neptune_cluster: Add `deletion_protection` argument ([#11731](https://github.com/terraform-providers/terraform-provider-aws/issues/11731))
* resource/aws_rds_global_database: Support `aurora-mysql` value in `engine` argument plan-time validation (Support Aurora MySQL 5.7) ([#11790](https://github.com/terraform-providers/terraform-provider-aws/issues/11790))

BUG FIXES:

* data-source/aws_route53_zone: Fixes regression from version 2.48.0 when filtering using `tags` ([#11953](https://github.com/terraform-providers/terraform-provider-aws/issues/11953))
* resource/aws_batch_job_definition: Prevent extraneous differences with `container_properties` argument missing `environment`, `mountPoints`, `ulimits`, and `volumes` configuration ([#12000](https://github.com/terraform-providers/terraform-provider-aws/issues/12000))
* resource/aws_cognito_user_pool: Allow `admin_create_user_config` configuration block `unused_account_validity_days` argument to be omitted ([#12001](https://github.com/terraform-providers/terraform-provider-aws/issues/12001))
* resource/aws_launch_configuration: Fixes regression from version 2.23.0 with instance store AMIs returning an unexpected error ([#9810](https://github.com/terraform-providers/terraform-provider-aws/issues/9810))
* resource/aws_launch_configuration: Fixes regression from version 2.23.0 to allow missing EC2 Image during root block device lookup ([#12009](https://github.com/terraform-providers/terraform-provider-aws/issues/12009))
* resource/aws_route53_record: The artificial, hardcoded five minute timeouts for creation and deletions have been removed in preference of the default AWS Go SDK retrying logic ([#11895](https://github.com/terraform-providers/terraform-provider-aws/issues/11895))

## 2.48.0 (February 07, 2020)

NOTES:

* resource/aws_organizations_policy_attachment: The underlying API calls have switched from `ListPoliciesForTarget` to `ListTargetsForPolicy`. Restrictive IAM Policies for Terraform execution may require updates. ([#11612](https://github.com/terraform-providers/terraform-provider-aws/issues/11612))

FEATURES:

* **New Data Source:** `aws_ssm_patch_baseline` ([#9486](https://github.com/terraform-providers/terraform-provider-aws/issues/9486))
* **New Resource:** `aws_datasync_location_smb` ([#10381](https://github.com/terraform-providers/terraform-provider-aws/issues/10381))

ENHANCEMENTS:

* resource/aws_batch_job_definition: Support resource import ([#11407](https://github.com/terraform-providers/terraform-provider-aws/issues/11407))
* resource/aws_codebuild_project: Add `source` and `secondary_source` configuration block `git_submodules_config` configuration block ([#10952](https://github.com/terraform-providers/terraform-provider-aws/issues/10952))
* resource/aws_codebuild_project: Add `source_version` argument ([#9877](https://github.com/terraform-providers/terraform-provider-aws/issues/9877))
* resource/aws_elasticache_cluster: Add `computed` flag for `port` property and set to true ([#10017](https://github.com/terraform-providers/terraform-provider-aws/issues/10017))
* resource/aws_fsx_lustre_file_system: Lower minimum `storage_capacity` argument validation to 1200 to match API updates ([#11847](https://github.com/terraform-providers/terraform-provider-aws/issues/11847))
* resource/aws_organizations_policy: Support `type` argument `TAG_POLICY` value in plan-time validation ([#11612](https://github.com/terraform-providers/terraform-provider-aws/issues/11612))
* resource/aws_organizations_policy_attachment: Support tag policies ([#11612](https://github.com/terraform-providers/terraform-provider-aws/issues/11612))

BUG FIXES:

* resource/aws_appautoscaling_target: Prevent state removal of resource immediately after creation due to eventual consistency ([#11819](https://github.com/terraform-providers/terraform-provider-aws/issues/11819))
* resource/aws_appautoscaling_target: Automatically retry creation on `ValidationException: ECS service doesn't exist` for ECS eventual consistency ([#11693](https://github.com/terraform-providers/terraform-provider-aws/issues/11693))
* resource/aws_batch_job_definition: Properly set `container_properties` and `name` into Terraform state and perform drift detection ([#11488](https://github.com/terraform-providers/terraform-provider-aws/issues/11488))
* resource/aws_cloudformation_stack_set: Wait for update operation completion (default timeout of 30 minutes) and report any errors ([#11726](https://github.com/terraform-providers/terraform-provider-aws/issues/11726))
* resource/aws_cloudwatch_log_stream: Prevent state removal of resource immediately after creation due to eventual consistency ([#11617](https://github.com/terraform-providers/terraform-provider-aws/issues/11617))
* resource/aws_codedeploy_deployment_group: Fixes unexpected behaviour when removing block attributes ([#11648](https://github.com/terraform-providers/terraform-provider-aws/issues/11648))
* resource/aws_default_security_group: Ensure `description` attribute is written into Terraform state ([#11650](https://github.com/terraform-providers/terraform-provider-aws/issues/11650))
* resource/aws_dynamodb_table: Skip `ResourceNotFoundException` error during deletion ([#11692](https://github.com/terraform-providers/terraform-provider-aws/issues/11692))
* resource/aws_ec2_client_vpn_endpoint: Ensure `dns_servers` attribute is refreshed in Terraform state ([#11889](https://github.com/terraform-providers/terraform-provider-aws/issues/11889))
* resource/aws_ecs_cluster: Delay check of ECS Cluster status during creation for ECS eventual consistency ([#11701](https://github.com/terraform-providers/terraform-provider-aws/issues/11701))
* resource/aws_kinesis_firehose_delivery_stream: Allow processors to be cleared from extended S3 configuration ([#11649](https://github.com/terraform-providers/terraform-provider-aws/issues/11649))
* resource/aws_network_acl_rule: Trigger resource recreation instead of error when same number rule (but opposite ingress/egress) is removed ([#11544](https://github.com/terraform-providers/terraform-provider-aws/issues/11544))
* resource/aws_placement_group: Additional handling for creation and deletion eventual consistency ([#11671](https://github.com/terraform-providers/terraform-provider-aws/issues/11671))
* resource/aws_s3_bucket: Retry read after creation for 404 status code and prevent 2 minute delay for triggering recreation on existing resources deleted outside Terraform ([#11894](https://github.com/terraform-providers/terraform-provider-aws/issues/11894))

## 2.47.0 (January 30, 2020)

NOTES:

* resource/aws_efs_file_system: Tagging API calls have been refactored to the AWS standardized `TagResource` and `UntagResource` API calls (from `CreateTags` and `DeleteTags` respectively). Restrictive IAM Policies for Terraform execution may require updates. ([#11654](https://github.com/terraform-providers/terraform-provider-aws/issues/11654))

ENHANCEMENTS:

* data-source/aws_api_gateway_vpc_link: Add `description`, `status`, `status_message`, `tags`, and `target_arns` attributes ([#10822](https://github.com/terraform-providers/terraform-provider-aws/issues/10822))
* data-source/aws_dynamodb_table: Add `server_side_encryption` `kms_key_arn` attribute ([#11081](https://github.com/terraform-providers/terraform-provider-aws/issues/11081))
* data-source/aws_efs_file_system: Add `lifecycle_policy`, `provisioned_throughput_in_mibps`, and `throughput_mode` attributes ([#11647](https://github.com/terraform-providers/terraform-provider-aws/issues/11647))
* data-source/aws_kms_key: Add `customer_master_key_spec` attribute ([#11062](https://github.com/terraform-providers/terraform-provider-aws/issues/11062))
* resource/aws_dynamodb_table: Add `server_side_encryption` configuration block `kms_key_arn` argument (support customer managed CMKs for server-side encryption) ([#11081](https://github.com/terraform-providers/terraform-provider-aws/issues/11081))
* resource/aws_dynamodb_table: Support in-place updates for `server_side_encryption` configurations ([#11081](https://github.com/terraform-providers/terraform-provider-aws/issues/11081))
* resource/aws_elasticsearch_domain: Add `domain_endpoint_options` configuration block (support enforcing HTTPS) ([#10430](https://github.com/terraform-providers/terraform-provider-aws/issues/10430))
* resource/aws_gamelift_fleet: Add `fleet_type` argument (support Spot Fleets) ([#8234](https://github.com/terraform-providers/terraform-provider-aws/issues/8234))
* resource/aws_kms_key: Add `customer_master_key_spec` argument and plan-time validation support for `key_usage` value `SIGN_VERIFY` (support asymmetric keys) ([#11062](https://github.com/terraform-providers/terraform-provider-aws/issues/11062))
* resource/aws_sagemaker_notebook_instance: Add `direct_internet_access` argument ([#8618](https://github.com/terraform-providers/terraform-provider-aws/issues/8618))
* resource/aws_ssm_activation: Add `automation_target_parameter_name` argument ([#11755](https://github.com/terraform-providers/terraform-provider-aws/issues/11755))
* resource/aws_ssm_document: Add `target_type` argument ([#11479](https://github.com/terraform-providers/terraform-provider-aws/issues/11479))
* resource/aws_ssm_maintenance_window: Add `description` argument ([#11478](https://github.com/terraform-providers/terraform-provider-aws/issues/11478))
* resource/aws_storagegateway_gateway: Add `cloudwatch_log_group_arn` argument ([#10939](https://github.com/terraform-providers/terraform-provider-aws/issues/10939))

BUG FIXES:

* data-source/aws_api_gateway_rest_api: Fixes `root_resource_id` not being set on correctly when REST API contains more than 25 resources ([#11705](https://github.com/terraform-providers/terraform-provider-aws/issues/11705))
* resource/aws_cloudwatch_log_subscription_filter: Perform eventual consistency retries on update ([#11739](https://github.com/terraform-providers/terraform-provider-aws/issues/11739))
* resource/aws_cognito_user_pool: Deprecate `unused_account_validity_days` argument and add support for `temporary_password_validity_days` argument ([#10890](https://github.com/terraform-providers/terraform-provider-aws/issues/10890))
* resource/aws_elasticsearch_domain: Automatically retry resource creation on additional error messages relating to eventual consistency ([#11663](https://github.com/terraform-providers/terraform-provider-aws/issues/11663))
* resource/aws_elasticsearch_domain: Ensure in-place version upgrade is fully successful before returning ([#11793](https://github.com/terraform-providers/terraform-provider-aws/issues/11793))
* resource/aws_emr_instance_group: Wait for `RUNNING` status on creation ([#11688](https://github.com/terraform-providers/terraform-provider-aws/issues/11688))
* resource/aws_ssm_activation: Properly trigger resource recreation when deleted outside Terraform ([#11658](https://github.com/terraform-providers/terraform-provider-aws/issues/11658))
* resource/aws_ssm_parameter: Prevent `KeyId` error when switching `type` value from `SecureString` to `String` ([#10819](https://github.com/terraform-providers/terraform-provider-aws/issues/10819))
* service/efs: Generate proper `dns_name` attribute hostname suffix in AWS China, AWS C2S, and AWS SC2S partitions ([#11746](https://github.com/terraform-providers/terraform-provider-aws/issues/11746))

## 2.46.0 (January 23, 2020)

NOTES:

* provider: Terraform AWS Provider version 2.45.0 included AWS Go SDK version 1.28.0, which contained a regression in error handling behavior across many services that either prevented or incorrectly modified error messages from being surfaced by the API. Other than confusing errors in certain cases, this also affected automatic retry logic in a few resources. This release contains an AWS Go SDK update which should resolve these issues.

ENHANCEMENTS:

* data-source/aws_api_gateway_api_key: Add `created_date`, `description`, `enabled`, `last_updated_date`, and `tags` attributes ([#10821](https://github.com/terraform-providers/terraform-provider-aws/issues/10821))
* data-source/aws_cloudwatch_log_group: Add `kms_key_id`, `retention_in_days`, and `tags` attributes ([#10755](https://github.com/terraform-providers/terraform-provider-aws/issues/10755))
* data-source/aws_db_instance: Add `multi_az` attribute ([#10795](https://github.com/terraform-providers/terraform-provider-aws/issues/10795))
* data-source/aws_sqs_queue: Add `tags` attribute ([#10820](https://github.com/terraform-providers/terraform-provider-aws/issues/10820))
* resource/aws_acm_certificate: Support tag-on-create ([#11073](https://github.com/terraform-providers/terraform-provider-aws/issues/11073))
* resource/aws_api_gateway_rest_api: Add `endpoint_configuration` configuration block `vpc_endpoint_ids` argument ([#10627](https://github.com/terraform-providers/terraform-provider-aws/issues/10627))
* resource/aws_cloudfront_distribution: Validate `origin_group` configuration block `member` argument contains max 2 items ([#10357](https://github.com/terraform-providers/terraform-provider-aws/issues/10357))
* resource/aws_cognito_user_pool_client: Support plan-time validation values of `ALLOW_*` variations for `explicit_auth_flows` argument ([#10976](https://github.com/terraform-providers/terraform-provider-aws/issues/10976))
* resource/aws_ecs_task_definition: Add `volume` configuration block `efs_volume_configuration` configuration block (support preview EFS volume configuration) ([#11707](https://github.com/terraform-providers/terraform-provider-aws/issues/11707))
* resource/aws_ecs_task_definition: Add plan-time validation for `execution_role_arn` argument, `placement_constraints` configuration block `type` argument, and `task_role_arn` argument ([#11707](https://github.com/terraform-providers/terraform-provider-aws/issues/11707))
* resource/aws_egress_only_internet_gateway: Support resource import ([#11071](https://github.com/terraform-providers/terraform-provider-aws/issues/11071))
* resource/aws_key_pair: Add `tags` argument and `key_pair_id` attribute ([#11481](https://github.com/terraform-providers/terraform-provider-aws/issues/11481))
* resource/aws_network_interface: Add `mac_address` attribute ([#10633](https://github.com/terraform-providers/terraform-provider-aws/issues/10633))
* resource/aws_organization_organization: Support plan-time validation value of `TAG_POLICY` in `enabled_policy_types` argument ([#11535](https://github.com/terraform-providers/terraform-provider-aws/issues/11535))
* resource/aws_placement_group: Add `tags` argument and `placement_group_id` attribute ([#11482](https://github.com/terraform-providers/terraform-provider-aws/issues/11482))
* resource/aws_rds_cluster_endpoint: Add `tags` argument ([#11074](https://github.com/terraform-providers/terraform-provider-aws/issues/11074))

BUG FIXES:

* data-source/aws_acmpca_certificate_authority: Properly set `not_after` and `not_before` values into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* provider: Upgrade AWS Go SDK dependency to fix missing/incorrect API error messages and missing retries regression introduced in Terraform AWS Provider version 2.45.0 ([#11727](https://github.com/terraform-providers/terraform-provider-aws/issues/11727))
* resource/aws_acmpca_certificate_authority: Properly set `not_after` and `not_before` values into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_api_gateway_account: Update retryable error message handling for recent API update ([#11735](https://github.com/terraform-providers/terraform-provider-aws/issues/11735))
* resource_aws_cognito_resource_server: Increase `scope` max limit to match API ([#10505](https://github.com/terraform-providers/terraform-provider-aws/issues/10505))
* resource_aws_cognito_user_pool_client: Increase `allowed_oauth_scopes` max limit to match API ([#10505](https://github.com/terraform-providers/terraform-provider-aws/issues/10505))
* resource/aws_dms_certificate: Properly set `certificate_wallet` value into Terraform state ([#11496](https://github.com/terraform-providers/terraform-provider-aws/issues/11496))
* resource/aws_ec2_client_vpn_endpoint: Properly set `status` value into Terraform state ([#11497](https://github.com/terraform-providers/terraform-provider-aws/issues/11497))
* resource/aws_ecs_task_definition: Properly refresh `ipc_mode` and `pid_mode` attributes in Terraform state for drift detection ([#11707](https://github.com/terraform-providers/terraform-provider-aws/issues/11707))
* resource/aws_emr_security_configuration: Properly set `creation_date` value into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_iam_service_linked_role: Properly set `create_date` value into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_iot_topic_rule: Trigger resource recreation on `name` argument updates ([#10366](https://github.com/terraform-providers/terraform-provider-aws/issues/10366))
* resource/aws_lambda_event_source_mapping: Properly set `last_modified` value into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_organizations_account: Properly set `joined_timestamp` value into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_redshift_cluster: Handle `available, prep-for-resize` pending status during creation and update ([#10530](https://github.com/terraform-providers/terraform-provider-aws/issues/10530))
* resource/aws_ssm_activation: Properly set `expiration_date` value into the Terraform state and perform drift detection when configured ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_ssm_document: Properly set `created_date` value into the Terraform state ([#11491](https://github.com/terraform-providers/terraform-provider-aws/issues/11491))
* resource/aws_waf_sql_injection_match_set: Properly set `sql_injection_match_tuples` value into Terraform state ([#11498](https://github.com/terraform-providers/terraform-provider-aws/issues/11498))

## 2.45.0 (January 17, 2020)

ENHANCEMENTS:

* resource/aws_codepipeline_webhook: Support in-place `tags` updates ([#11387](https://github.com/terraform-providers/terraform-provider-aws/issues/11387))
* resource/aws_db_parameter_group: Support resetting parameter group values ([#11540](https://github.com/terraform-providers/terraform-provider-aws/issues/11540))
* resource/aws_docdb_cluster: Support `profiler` CloudWatch export type ([#11051](https://github.com/terraform-providers/terraform-provider-aws/issues/11051))
* resource/aws_gamelift_alias: Add `tags` argument ([#11486](https://github.com/terraform-providers/terraform-provider-aws/issues/11486))
* resource/aws_gamelift_build: Add `tags` argument and `arn` attribute ([#11486](https://github.com/terraform-providers/terraform-provider-aws/issues/11486))
* resource/aws_gamelift_fleet - Add support for instance_role_arn ([#11553](https://github.com/terraform-providers/terraform-provider-aws/issues/11553))
* resource/aws_gamelift_game_session_queue: Add `tags` argument ([#11486](https://github.com/terraform-providers/terraform-provider-aws/issues/11486))
* resource/aws_neptune_parameter_group: Support tag-on-create ([#11245](https://github.com/terraform-providers/terraform-provider-aws/issues/11245))
* resource/aws_pinpoint_app: Add plan-time validation for `limit` configuration block `daily`, `maximum_duration`, `messages_per_second` and `total` arguments ([#11368](https://github.com/terraform-providers/terraform-provider-aws/issues/11368))
* resource/aws_rds_cluster: Allow enabling Aurora Serverless HTTP endpoint (Data API) with `enable_http_endpoint` ([#11048](https://github.com/terraform-providers/terraform-provider-aws/issues/11048))
* resource/aws_rds_cluster_parameter_group: Support resetting parameter group values ([#11540](https://github.com/terraform-providers/terraform-provider-aws/issues/11540))
* resource/aws_ssm_document: Add support for "Package" document type ([#11492](https://github.com/terraform-providers/terraform-provider-aws/issues/11492))
* resource/aws_vpc_peering_connection_accepter: Support resource import ([#4486](https://github.com/terraform-providers/terraform-provider-aws/issues/4486))

BUG FIXES:

* resource/aws_autoscaling_group: Prevent indefinite wait for desired capacity to be available when instance_weight specified and >=1 ([#11357](https://github.com/terraform-providers/terraform-provider-aws/issues/11357))
* resource/aws_cloudwatch_event_rule: Retry deletion on CloudWatch Events Target deletion eventual consistency ([#11475](https://github.com/terraform-providers/terraform-provider-aws/issues/11475))
* resource/aws_cloudwatch_event_target: Return failed entry error code and message if provided in `RemoveTargets` response ([#11475](https://github.com/terraform-providers/terraform-provider-aws/issues/11475))
* resource/aws_codepipeline_webhook: Properly trigger resource recreation when `authentication_configuration` configuration block `allowed_ip_range` and `secret_token` arguments change ([#11387](https://github.com/terraform-providers/terraform-provider-aws/issues/11387))
* resource/aws_emr_cluster: Prevent perpetual difference with `ec2_attributes` configuration block `emr_managed_master_security_group`, `emr_managed_slave_security_group`, and `service_access_security_groups` arguments when omitted (support EMR Managed Security Groups) ([#5493](https://github.com/terraform-providers/terraform-provider-aws/issues/5493))
* resource/aws_opsworks_permission: Prevent `Unable to change own permission level` error during self updates ([#11379](https://github.com/terraform-providers/terraform-provider-aws/issues/11379))

## 2.44.0 (January 09, 2020)

FEATURES:

* **New Data Source:** `aws_directory_service_directory` ([#11282](https://github.com/terraform-providers/terraform-provider-aws/issues/11282))
* **New Resource:** `aws_workspaces_directory` ([#11023](https://github.com/terraform-providers/terraform-provider-aws/issues/11023))

ENHANCEMENTS:

* data-source/aws_launch_configuration: Add `arn` attribute ([#11416](https://github.com/terraform-providers/terraform-provider-aws/issues/11416))
* data-source/aws_eks_cluster: Add `vpc_config` list `public_access_cidrs` attribute ([#11442](https://github.com/terraform-providers/terraform-provider-aws/issues/11442))
* resource/aws_ami_launch_permission: Support resource import ([#11437](https://github.com/terraform-providers/terraform-provider-aws/issues/11437))
* resource/aws_api_gateway_authorizer: Support resource import ([#11436](https://github.com/terraform-providers/terraform-provider-aws/issues/11436))
* resource/aws_api_gateway_authorizer: Add plan time validation for `provider_arns` argument ([#11436](https://github.com/terraform-providers/terraform-provider-aws/issues/11436))
* resource/aws_api_gateway_usage_plan_key: Support resource import ([#11439](https://github.com/terraform-providers/terraform-provider-aws/issues/11439))
* resource/aws_batch_compute_environment: Add `compute_environment_name_prefix` argument and make `compute_environment_name` argument optional (support full name generation) ([#10682](https://github.com/terraform-providers/terraform-provider-aws/issues/10682))
* resource/aws_batch_compute_environment: Add `compute_resources` configuration block `allocation_strategy` argument ([#10894](https://github.com/terraform-providers/terraform-provider-aws/issues/10894))
* resource/aws_batch_job_queue: Support resource import ([#11406](https://github.com/terraform-providers/terraform-provider-aws/issues/11406))
* resource/aws_cloudformation_stack: Prevent difference with `Transform` templates showing processed template (support SAM templates) ([#9006](https://github.com/terraform-providers/terraform-provider-aws/issues/9006))
* resource/aws_cloudwatch_event_rule: Support tag-on-create ([#11346](https://github.com/terraform-providers/terraform-provider-aws/issues/11346))
* resource/aws_db_instance: Remove `identifier_prefix` 16 character truncation for `sqlserver` engine ([#9040](https://github.com/terraform-providers/terraform-provider-aws/issues/9040))
* resource/aws_ecs_service: Add plan time validation for `launch_type`, `load_balancer` configuration block `target_group_arn` and `container_port`, and `placement_constraints` configuration block `type` arguments ([#11423](https://github.com/terraform-providers/terraform-provider-aws/issues/11423))
* resource/aws_eks_cluster: Add `vpc_config` configuration block `public_access_cidrs` argument ([#11442](https://github.com/terraform-providers/terraform-provider-aws/issues/11442))
* resource/aws_elasticache_cluster: Add `arn` attribute ([#11243](https://github.com/terraform-providers/terraform-provider-aws/issues/11243))
* resource/aws_launch_configuration: Add `arn` attribute ([#11416](https://github.com/terraform-providers/terraform-provider-aws/issues/11416))
* resource/aws_lb: Add plan-time validation for `ip_address_type` and `load_balancer_type` arguments ([#11419](https://github.com/terraform-providers/terraform-provider-aws/issues/11419))
* resource/aws_rds_cluster_instance: Allow updating `ca_cert_identifier` for `aws_rds_cluster_instance` ([#10954](https://github.com/terraform-providers/terraform-provider-aws/issues/10954))
* resource/aws_wafregional_xss_match_set: Support resource import ([#11432](https://github.com/terraform-providers/terraform-provider-aws/issues/11432))

BUG FIXES:

* provider: Allow `aws` account ID in ARN validation (support ARNs such as AWS Managed IAM Policies) ([#11450](https://github.com/terraform-providers/terraform-provider-aws/issues/11450))
* provider: Support AWS C2S/SC2S Regional ARNs in ARN validation ([#11471](https://github.com/terraform-providers/terraform-provider-aws/issues/11471))
* resource/aws_api_gateway_usage_plan_key: Ensure Terraform performs drift detection of `key_type` argument ([#11439](https://github.com/terraform-providers/terraform-provider-aws/issues/11439))
* resource/aws_appautoscaling_policy: Prevent potential state removal of resource immediately after creation due to eventual consistency ([#11222](https://github.com/terraform-providers/terraform-provider-aws/issues/11222))
* resource/aws_cloudwatch_dashboard: Trigger resource recreation on `dashboard_name` updates (prevent dangling resource) ([#9784](https://github.com/terraform-providers/terraform-provider-aws/issues/9784))
* resource/aws_cloudwatch_event_rule: Improved handling of `is_enabled` argument ([#11346](https://github.com/terraform-providers/terraform-provider-aws/issues/11346))
* resource/aws_ecs_service: Automatically retry IAM Service Linked Role assume role error on creation due to asynchronous creation of role on first usage and IAM eventual consistency ([#11423](https://github.com/terraform-providers/terraform-provider-aws/issues/11423))
* resource/aws_iam_instance: Allows for instance profiles to be changed when instances are in stopped state ([#11104](https://github.com/terraform-providers/terraform-provider-aws/issues/11104))
* resource/aws_opsworks_stack: Ensure `tags` are refreshed in Terraform state during read for drift detection ([#11373](https://github.com/terraform-providers/terraform-provider-aws/issues/11373))
* resource/aws_rds_cluster_instance: Prevent `is already being deleted` error on deletion and wait for deletion completion ([#11468](https://github.com/terraform-providers/terraform-provider-aws/issues/11468))

## 2.43.0 (December 19, 2019)

NOTES:

* This will be the last planned release until early January. Enjoy the rest of your year!

FEATURES:

* **New Data Source:** `aws_organizations_organizational_units` ([#10395](https://github.com/terraform-providers/terraform-provider-aws/issues/10395))
* **New Resource:** `aws_accessanalyzer_analyzer` ([#11169](https://github.com/terraform-providers/terraform-provider-aws/issues/11169))
* **New Resource:** `aws_lambda_function_event_invoke_config` ([#11165](https://github.com/terraform-providers/terraform-provider-aws/issues/11165))

ENHANCEMENTS:

* data-source/aws_elb: Add `arn` attribute ([#11345](https://github.com/terraform-providers/terraform-provider-aws/issues/11345))
* resource/aws_batch_compute_environment: Support resource import ([#11299](https://github.com/terraform-providers/terraform-provider-aws/issues/11299))
* resource/aws_codebuild_project: Add `queued_timeout` argument ([#11261](https://github.com/terraform-providers/terraform-provider-aws/issues/11261))
* resource/aws_fsx_windows_file_system: Support `storage_capacity` minimum value of `32` in validation to match recent updates to the API ([#11272](https://github.com/terraform-providers/terraform-provider-aws/issues/11272))
* resource/aws_opsworks_custom_layer: Add `encrypted` `ebs_volume` configuration ([#7110](https://github.com/terraform-providers/terraform-provider-aws/issues/7110))

BUG FIXES:

* resource/aws_datasync_agent: Trigger resource recreation on updated `InvalidRequestException` error for agents deleted outside Terraform ([#11005](https://github.com/terraform-providers/terraform-provider-aws/issues/11005))
* resource/aws_ecs_cluster: Fixes intermittent failures on update when cluster dependencies are updating ([#11310](https://github.com/terraform-providers/terraform-provider-aws/issues/11310))
* resource/aws_ecs_cluster: Fixes bug where ECS cluster capacity providers are updated but default provider strategy is not changed ([#11316](https://github.com/terraform-providers/terraform-provider-aws/issues/11316))
* resource/aws_globalaccelerator_endpoint_group: Allow `traffic_dial_percentage` to be set to `0` ([#11253](https://github.com/terraform-providers/terraform-provider-aws/issues/11253))
* resource/aws_lb_listener_rule: Fixes regression from version 2.42.0 when updating a rule without modifying condition ([#11364](https://github.com/terraform-providers/terraform-provider-aws/issues/11364))
* resource/aws_ssm_activation: Ensure `tags` are refreshed into Terraform state during read for drift detection ([#11290](https://github.com/terraform-providers/terraform-provider-aws/issues/11290))

## 2.42.0 (December 13, 2019)

FEATURES:

* **New Resource:** `aws_ecs_capacity_provider` ([#11151](https://github.com/terraform-providers/terraform-provider-aws/issues/11151))
* **New Resource:** `aws_media_convert_queue` ([#10041](https://github.com/terraform-providers/terraform-provider-aws/issues/10041))
* **New Resource:** `aws_workspaces_ip_group` ([#10904](https://github.com/terraform-providers/terraform-provider-aws/issues/10904))

ENHANCEMENTS:

* resource/aws_apigateway_usage_plan: Add `tags` argument and `arn` attribute ([#10566](https://github.com/terraform-providers/terraform-provider-aws/issues/10566))
* resource/aws_codebuild_project: Add `ARM_CONTAINER` as valid `environment` configuration block `compute_type` argument value ([#11206](https://github.com/terraform-providers/terraform-provider-aws/issues/11206))
* resource/aws_ecs_cluster: Add `capacity_providers` argument and `default_capacity_provider_strategy` configuration block (support ECS Capacity Providers) ([#11151](https://github.com/terraform-providers/terraform-provider-aws/issues/11151))
* resource/aws_ecs_service: Add `capacity_provider_strategy` configuration block (support ECS Capacity Providers) ([#11151](https://github.com/terraform-providers/terraform-provider-aws/issues/11151))
* resource/aws_emr_cluster: Add `step_concurrency_level` argument ([#11196](https://github.com/terraform-providers/terraform-provider-aws/issues/11196))
* resource/aws_lb_listener_rule: Support ALB advanced routing rules ([#8268](https://github.com/terraform-providers/terraform-provider-aws/issues/8268))

BUG FIXES:

* provider: Prevent crash in planning IAM Policy equivalency checking with invalid `Resource` declarations (e.g. a list of list of strings) ([#11107](https://github.com/terraform-providers/terraform-provider-aws/issues/11107))
* resource/aws_eks_cluster: Handle additional `InvalidParameterException: Error in role params` error during creation for IAM eventual consistency ([#11127](https://github.com/terraform-providers/terraform-provider-aws/issues/11127))
* resource/aws_iam_role: Ignore additional `NoSuchEntity` errors on deletion ([#11125](https://github.com/terraform-providers/terraform-provider-aws/issues/11125))
* resource/aws_network_interface: Prevent extraneous `ModifyNetworkInterfaceAttribute` API call during update ([#11277](https://github.com/terraform-providers/terraform-provider-aws/issues/11277))
* resource/aws_security_group: Support ampersand (`&`) in `ingress` and `egress` configuration block `description` argument value validation ([#9528](https://github.com/terraform-providers/terraform-provider-aws/issues/9528))
* resource/aws_security_group_rule: Support ampersand (`&`) in `description` argument value validation ([#9528](https://github.com/terraform-providers/terraform-provider-aws/issues/9528))

## 2.41.0 (December 04, 2019)

FEATURES:

* **New Resource:** `aws_eks_fargate_profile` ([#11111](https://github.com/terraform-providers/terraform-provider-aws/issues/11111))
* **New Resource:** `aws_lambda_provisioned_concurrency_config` ([#11129](https://github.com/terraform-providers/terraform-provider-aws/issues/11129))

ENHANCEMENTS:

* data-source/aws_route_table: adds attributes `gateway_id` and `associations.gateway_id` ([#11122](https://github.com/terraform-providers/terraform-provider-aws/issues/11122))
* resource/aws_autoscaling_group: Add `max_instance_lifetime` argument ([#10951](https://github.com/terraform-providers/terraform-provider-aws/issues/10951))
* resource/aws_autoscaling_group: Add `mixed_instances_policy` `launch_template` `override` configuration block `weighted_capacity` argument ([#11004](https://github.com/terraform-providers/terraform-provider-aws/issues/11004))
* resource/aws_codebuild_project: Add Linux GPU worker ([#11035](https://github.com/terraform-providers/terraform-provider-aws/issues/11035))
* resource/aws_docdb_cluster_instance: Add support for `ca_cert_identifier` parameter ([#11041](https://github.com/terraform-providers/terraform-provider-aws/issues/11041))
* resource/aws_emr_cluster: Outputs EMR cluster ARN ([#11078](https://github.com/terraform-providers/terraform-provider-aws/issues/11078))
* resource/aws_iam_access_key: Remove deprecation from `secret` and mark `secret` and `ses_smtp_password` to sensitive ([#10908](https://github.com/terraform-providers/terraform-provider-aws/issues/10908))
* resource/aws_iam_user: Delete a user's virtual MFA devices when `force_destroy` is enabled ([#11040](https://github.com/terraform-providers/terraform-provider-aws/issues/11040))
* resource/aws_route_table_association: adds attribute `gateway_id` ([#11122](https://github.com/terraform-providers/terraform-provider-aws/issues/11122))

BUG FIXES:

* resource/aws_batch_compute_environment: Forces new resource when `launch_template` contents are changed ([#11057](https://github.com/terraform-providers/terraform-provider-aws/issues/11057))
* resource/aws_datasync_location_s3: Automatically retry creation for IAM errors due to eventual consistency ([#10984](https://github.com/terraform-providers/terraform-provider-aws/issues/10984))
* resource/aws_launch_template: Only set associate_public_ip_address on network interfaces if it's explicitly set to avoid problems with multiple network interfaces ([#10157](https://github.com/terraform-providers/terraform-provider-aws/issues/10157))

## 2.40.0 (November 26, 2019)

NOTES:

* resource/aws_datasync_task: The DataSync API and SDK have removed `BEST_EFFORT` as a valid value for the `options` configuration block `posix_permissions` argument. The value has been removed from the validation in this resource to match those changes. ([#10985](https://github.com/terraform-providers/terraform-provider-aws/issues/10985))

FEATURES:

* **New Resource:** `aws_dx_hosted_transit_virtual_interface` ([#8523](https://github.com/terraform-providers/terraform-provider-aws/issues/8523))
* **New Resource:** `aws_dx_hosted_transit_virtual_interface_accepter` ([#8523](https://github.com/terraform-providers/terraform-provider-aws/issues/8523))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `vpc_config` nested block `cluster_security_group_id` attribute ([#11002](https://github.com/terraform-providers/terraform-provider-aws/issues/11002))
* resource/aws_cloudwatch_metric_alarm: Add `threshold_metric_id` argument (support Anomaly Detection metrics) ([#9828](https://github.com/terraform-providers/terraform-provider-aws/issues/9828))
* resource/aws_codebuild_project: Add support for BUILD_GENERAL1_2XLARGE CodeBuild compute type [GH11015]
* resource/aws_dx_private_virtual_interface: Support tagging-on-create ([#9572](https://github.com/terraform-providers/terraform-provider-aws/issues/9572))
* resource/aws_dx_private_virtual_interface: Validate Virtual Interface type on import ([#9572](https://github.com/terraform-providers/terraform-provider-aws/issues/9572))
* resource/aws_dx_public_virtual_interface: Validate Virtual Interface type on import ([#9572](https://github.com/terraform-providers/terraform-provider-aws/issues/9572))
* resource/aws_ebs_snapshot: Support tagging-on-create and in-place `tags` updates ([#10935](https://github.com/terraform-providers/terraform-provider-aws/issues/10935))
* resource/aws_ebs_snapshot_copy: Support tagging-on-create and in-place `tags` updates ([#10936](https://github.com/terraform-providers/terraform-provider-aws/issues/10936))
* resource/aws_eks_cluster: Add `vpc_config` configuration block `cluster_security_group_id` attribute ([#11002](https://github.com/terraform-providers/terraform-provider-aws/issues/11002))
* resource/aws_lambda_function: Support waiting for function creation and configuration updates ([#11016](https://github.com/terraform-providers/terraform-provider-aws/issues/11016))

BUG FIXES:

* data-source/aws_iam_group: Ensure `users` attribute populates fully when group contains more than 100 users ([#10993](https://github.com/terraform-providers/terraform-provider-aws/issues/10993))
* resource/aws_default_route_table: Return helpful not found error on resource creation instead of generic `Provider produced inconsistent result after apply` error when given invalid `default_route_table_id` argument value ([#10981](https://github.com/terraform-providers/terraform-provider-aws/issues/10981))
* resource/aws_default_route_table: Propose resource recreation for missing Default Route Table on refresh instead of returning an error ([#10981](https://github.com/terraform-providers/terraform-provider-aws/issues/10981))

## 2.39.0 (November 21, 2019)

FEATURES:

* **New Data Source:** `aws_guardduty_detector` ([#10463](https://github.com/terraform-providers/terraform-provider-aws/issues/10463))
* **New Resource:** `aws_glue_workflow` ([#10891](https://github.com/terraform-providers/terraform-provider-aws/issues/10891))

ENHANCEMENTS:

* provider: Support for EC2 Metadata secure tokens ([#10940](https://github.com/terraform-providers/terraform-provider-aws/issues/10940))
* resource/aws_glue_job: Add `number_of_workers` and `worker_type` arguments ([#9115](https://github.com/terraform-providers/terraform-provider-aws/issues/9115))
* resource/aws_glue_job: Add `tags` argument and `arn` attribute ([#10968](https://github.com/terraform-providers/terraform-provider-aws/issues/10968))
* resource/aws_glue_trigger: Add `workflow_name` argument ([#9762](https://github.com/terraform-providers/terraform-provider-aws/issues/9762))
* resource/aws_glue_trigger: Add `actions` configuration block `crawler_name` argument ([#10190](https://github.com/terraform-providers/terraform-provider-aws/issues/10190))
* resource/aws_glue_trigger: Add `predicate` `conditions` configuration block `crawler_name` and `crawl_state` arguments ([#10190](https://github.com/terraform-providers/terraform-provider-aws/issues/10190))
* resource/aws_glue_trigger: Add `tags` argument and `arn` attribute ([#10967](https://github.com/terraform-providers/terraform-provider-aws/issues/10967))
* resource/aws_iam_group_policy: Add IAM Policy JSON difference suppression and validation to `policy` argument ([#9660](https://github.com/terraform-providers/terraform-provider-aws/issues/9660))
* resource/aws_lambda_event_source_mapping: Add `maximum_batching_window_in_seconds` argument ([#10051](https://github.com/terraform-providers/terraform-provider-aws/issues/10051))
* resource/aws_lambda_function: Support `java11`, `nodejs12.x`, and `python3.8` as valid `runtime` argument values in validation ([#10938](https://github.com/terraform-providers/terraform-provider-aws/issues/10938))
* resource/aws_lambda_layer_version: Support `java11`, `nodejs12.x`, and `python3.8` as valid `compatible_runtimes` argument values in validation ([#10938](https://github.com/terraform-providers/terraform-provider-aws/issues/10938))
* resource/aws_resourcegroups_group: Add `tags` argument ([#10640](https://github.com/terraform-providers/terraform-provider-aws/issues/10640))

BUG FIXES:

* data_source/aws_instance: Fixes a bug where multiple EBS volumes would get collapsed and only one would return ([#10045](https://github.com/terraform-providers/terraform-provider-aws/issues/10045))
* resource/aws_appmesh_virtual_node: Allow FQDN values in `service_discovery` `aws_cloud_map` configuration block `namespace_name` and `service_name` argument validations ([#9788](https://github.com/terraform-providers/terraform-provider-aws/issues/9788))
* resource/aws_batch_compute_environment: Propose resource recreation when updating `compute_resources` configuration block `tags` argument ([#10937](https://github.com/terraform-providers/terraform-provider-aws/issues/10937))
* resource/aws_iam_instance_profile: Remove requirement to specify a role, as it is not required by the API ([#10525](https://github.com/terraform-providers/terraform-provider-aws/issues/10525))
* resource/aws_opsworks_application: Fixes issue where `terraform apply` continuously suggests applying changes to `ssh_key` or `password` in `app_source` property ([#10175](https://github.com/terraform-providers/terraform-provider-aws/issues/10175))
* resource/aws_opsworks_stack: Fixes issue where `terraform apply` continuously suggests applying changes to `ssh_key` or `password` in `custom_cookbooks_source` property ([#10175](https://github.com/terraform-providers/terraform-provider-aws/issues/10175))

## 2.38.0 (November 18, 2019)

FEATURES:

* **New Resource:** `aws_eks_node_group` ([#10916](https://github.com/terraform-providers/terraform-provider-aws/issues/10916))

## 2.37.0 (November 18, 2019)

ENHANCEMENTS:

* resource/aws_api_gateway_rest_api: Add `tags` argument and `arn` attribute ([#10581](https://github.com/terraform-providers/terraform-provider-aws/issues/10581))
* resource/aws_db_instance: Add `ca_cert_identifier` argument ([#10490](https://github.com/terraform-providers/terraform-provider-aws/issues/10490))
* resource/aws_dlm_lifecycle_policy: Add `tags` argument and `arn` attribute ([#10864](https://github.com/terraform-providers/terraform-provider-aws/issues/10864))
* resource/aws_efs_file_system: Add `AFTER_7_DAYS` as a valid `lifecycle_policy` configuration block `transition_to_ia` argument value ([#10825](https://github.com/terraform-providers/terraform-provider-aws/issues/10825))
* resource/aws_glue_crawler: Add `tags` argument ([#10805](https://github.com/terraform-providers/terraform-provider-aws/issues/10805))
* resource/aws_s3_bucket_inventory: Add `IntelligentTieringAccessTier` as valid value for `optional_fields` argument ([#10746](https://github.com/terraform-providers/terraform-provider-aws/issues/10746))
* resource/aws_waf_geo_match_set: Support resource import and add `arn` attribute ([#10480](https://github.com/terraform-providers/terraform-provider-aws/issues/10480))
* resource/aws_waf_regex_match_set: Support resource import and add `arn` attribute ([#10481](https://github.com/terraform-providers/terraform-provider-aws/issues/10481))
* resource/aws_waf_regex_pattern_set: Support resource import and add `arn` attribute ([#10482](https://github.com/terraform-providers/terraform-provider-aws/issues/10482))
* resource/aws_waf_size_constraint_set: Support resource import and add `arn` attribute ([#10484](https://github.com/terraform-providers/terraform-provider-aws/issues/10484))
* resource/aws_waf_xss_match_set: Support resource import and add `arn` attribute ([#10485](https://github.com/terraform-providers/terraform-provider-aws/issues/10485))
* resource/aws_wafregional_rate_based_rule: Add `tags` argument and `arn` attribute ([#10897](https://github.com/terraform-providers/terraform-provider-aws/issues/10897))
* resource/aws_wafregional_rule_group: Add `tags` argument and `arn` attribute ([#10896](https://github.com/terraform-providers/terraform-provider-aws/issues/10896))
* resource/aws_wafregional_rule: Add `tags` argument and `arn` attribute ([#10895](https://github.com/terraform-providers/terraform-provider-aws/issues/10895))
* resource/aws_wafregional_web_acl: Add `tags` argument ([#10889](https://github.com/terraform-providers/terraform-provider-aws/issues/10889))
* resource/aws_wafregional_web_acl_association: Support resource import ([#10538](https://github.com/terraform-providers/terraform-provider-aws/issues/10538))
* resource/aws_cloudtrail: support Tag on create ([#10818](https://github.com/terraform-providers/terraform-provider-aws/issues/10818))

BUG FIXES:

* data-source/aws_iam_policy_document: Prevent panic when combining single principal identifier with multiple principal identifiers ([#10780](https://github.com/terraform-providers/terraform-provider-aws/issues/10780))
* data-source/aws_iam_policy_document: Prevent losing identifier elements when combining single and multiple principals identifiers ([#10844](https://github.com/terraform-providers/terraform-provider-aws/issues/10844))
* resource/aws_servicequotas_service_quota: Remove resource from Terraform state on `NoSuchResourceException` error ([#10735](https://github.com/terraform-providers/terraform-provider-aws/issues/10735))

## 2.36.0 (November 14, 2019)

ENHANCEMENTS:

* data-source/aws_iam_group: Add `users` attribute ([#7132](https://github.com/terraform-providers/terraform-provider-aws/issues/7132))
* resource/aws_apigateway_stage: Add `arn` attribute ([#10570](https://github.com/terraform-providers/terraform-provider-aws/issues/10570))
* resource/aws_s3_bucket: Retry reading tags on `NoSuchBucket` errors due to eventual inconsistency ([#10863](https://github.com/terraform-providers/terraform-provider-aws/issues/10863))
* resource/aws_waf_rule: Add `arn` attribute ([#10798](https://github.com/terraform-providers/terraform-provider-aws/issues/10798))
* resource/aws_waf_rule_group: Add `arn` attribute ([#10799](https://github.com/terraform-providers/terraform-provider-aws/issues/10799))

## 2.35.0 (November 07, 2019)

NOTES:

* provider: New `ignore_tag_prefixes` and `ignore_tags` arguments are being tested as a public preview for ignoring tags across all resources under a provider. Support for the functionality must be added to individual resources in the codebase and is only implemented for the `aws_subnet` and `aws_vpc` resources at this time. Until a general availability announcement, no compatibility promises are made with these provider arguments and their functionality. ([#10418](https://github.com/terraform-providers/terraform-provider-aws/issues/10418))

FEATURES:

* **New Data Source:** `aws_qldb_ledger` ([#10394](https://github.com/terraform-providers/terraform-provider-aws/issues/10394))
* **New Resource:** `aws_qldb_ledger` ([#10394](https://github.com/terraform-providers/terraform-provider-aws/issues/10394))

ENHANCEMENTS:

* data-source/aws_db_cluster_snapshot: Add `tags` attribute ([#10488](https://github.com/terraform-providers/terraform-provider-aws/issues/10488))
* data-source/aws_db_instance: Add `tags` attribute ([#10550](https://github.com/terraform-providers/terraform-provider-aws/issues/10550))
* data-source/aws_vpc_endpoint: Add `filter` and `tags` arguments ([#10503](https://github.com/terraform-providers/terraform-provider-aws/issues/10503))
* provider: Add `ignore_tag_prefixes` and `ignore_tags` arguments (in public preview, see note above) ([#10418](https://github.com/terraform-providers/terraform-provider-aws/issues/10418))
* resource/aws_acmpca_certificate_authority: Support tagging on creation ([#10736](https://github.com/terraform-providers/terraform-provider-aws/issues/10736))
* resource/aws_api_gateway_api_key: Add `tags` argument and `arn` attribute ([#10568](https://github.com/terraform-providers/terraform-provider-aws/issues/10568))
* resource/aws_api_gateway_client_certificate: Add `tags` argument and `arn` attribute ([#10569](https://github.com/terraform-providers/terraform-provider-aws/issues/10569))
* resource/aws_api_gateway_domain_name: Add `tags` argument and `arn` attribute ([#10567](https://github.com/terraform-providers/terraform-provider-aws/issues/10567))
* resource/aws_api_gateway_vpc_link: Add `tags` argument and `arn` attribute ([#10561](https://github.com/terraform-providers/terraform-provider-aws/issues/10561))
* resource/aws_cloudwatch_log_group: Support tagging on creation ([#10753](https://github.com/terraform-providers/terraform-provider-aws/issues/10753))
* resource/aws_db_cluster_snapshot: Add `tags` argument ([#10488](https://github.com/terraform-providers/terraform-provider-aws/issues/10488))
* resource/aws_ec2_fleet: Support in-place `tags` updates ([#10761](https://github.com/terraform-providers/terraform-provider-aws/issues/10761))
* resource/aws_launch_template: Support tagging on creation ([#10759](https://github.com/terraform-providers/terraform-provider-aws/issues/10759))
* resource/aws_mq_broker: Support in-place `security_groups` updates ([#10442](https://github.com/terraform-providers/terraform-provider-aws/issues/10442))
* resource/aws_storagegateway_cached_iscsi_volume: Add `tags` argument ([#10613](https://github.com/terraform-providers/terraform-provider-aws/issues/10613))
* resource/aws_storagegateway_gateway: Add `tags` argument ([#10588](https://github.com/terraform-providers/terraform-provider-aws/issues/10588))
* resource/aws_storagegateway_nfs_file_share: Add `tags` argument ([#10722](https://github.com/terraform-providers/terraform-provider-aws/issues/10722))
* resource/aws_subnet: Support provider-wide ignore tags (in public preview, see note above) ([#10418](https://github.com/terraform-providers/terraform-provider-aws/issues/10418))
* resource/aws_swf_domain: Add `tags` argument and `arn` attribute ([#10763](https://github.com/terraform-providers/terraform-provider-aws/issues/10763))
* resource/aws_vpc: Support provider-wide ignore tags (in public preview, see note above) ([#10418](https://github.com/terraform-providers/terraform-provider-aws/issues/10418))
* resource/aws_waf_rate_based_rule: Add `tags` argument and `arn` attribute ([#10479](https://github.com/terraform-providers/terraform-provider-aws/issues/10479))

BUG FIXES:

* data-source/aws_route53_resolver_rule: Do not retrieve tags for rules shared with the AWS account that owns the data source ([#10348](https://github.com/terraform-providers/terraform-provider-aws/issues/10348))
* resource/aws_api_gateway_authorizer: Set `authorizer_result_ttl_in_seconds` argument default to 300 to match API default which properly allows setting to 0 for disabling caching ([#9605](https://github.com/terraform-providers/terraform-provider-aws/issues/9605))
* resource/aws_autoscaling_group: Batch ELB attachments and detachments by 10 to prevent API and rate limiting errors ([#10445](https://github.com/terraform-providers/terraform-provider-aws/issues/10445))
* resource/aws_s3_bucket_public_access_block: Remove from Terraform state when S3 Bucket is already destroyed ([#10534](https://github.com/terraform-providers/terraform-provider-aws/issues/10534))
* resource/aws_ssm_maintenance_window_task: Prevent crashes with empty configuration blocks ([#10713](https://github.com/terraform-providers/terraform-provider-aws/issues/10713))

## 2.34.0 (October 31, 2019)

ENHANCEMENTS:

* resource/aws_ecr_repository: Add `image_scanning_configuration` configuration block (support image scanning on push) ([#10671](https://github.com/terraform-providers/terraform-provider-aws/issues/10671))
* resource/aws_elasticache_replication_group: Add `kms_key_id` argument (support KMS encryption) ([#10380](https://github.com/terraform-providers/terraform-provider-aws/issues/10380))
* resource/aws_flow_log: Add `log_format` argument ([#10374](https://github.com/terraform-providers/terraform-provider-aws/issues/10374))
* resource/aws_glue_job: Add `glue_version` argument ([#10237](https://github.com/terraform-providers/terraform-provider-aws/issues/10237))
* resource/aws_storagegateway_smb_file_share: Add `tags` argument ([#10620](https://github.com/terraform-providers/terraform-provider-aws/issues/10620))

BUG FIXES:

* resource/aws_backup_plan: Correctly handle changes to `recovery_point_tags` arguments ([#10641](https://github.com/terraform-providers/terraform-provider-aws/issues/10641))
* resource/aws_backup_plan: Prevent `diffs didn't match` errors with `rule` configuration blocks ([#10641](https://github.com/terraform-providers/terraform-provider-aws/issues/10641))
* resource/aws_cloudhsm_v2_cluster: Ensure multiple tag configurations are applied correctly ([#10309](https://github.com/terraform-providers/terraform-provider-aws/issues/10309))
* resource/aws_cloudhsm_v2_cluster: Perform drift detection with tags ([#10309](https://github.com/terraform-providers/terraform-provider-aws/issues/10309))
* resource/aws_dx_gateway_association: Fix backwards compatibility issue with missing `dx_gateway_association_id` attribute ([#8776](https://github.com/terraform-providers/terraform-provider-aws/issues/8776))
* resource/aws_s3_bucket: Bypass `MethodNotAllowed` errors for Object Lock Configuration on read (support AWS C2S) ([#10657](https://github.com/terraform-providers/terraform-provider-aws/issues/10657))

## 2.33.0 (October 17, 2019)

FEATURES:

* **New Data Source:** `aws_waf_rate_based_rule` ([#10124](https://github.com/terraform-providers/terraform-provider-aws/issues/10124))
* **New Data Source:** `aws_wafregional_rate_based_rule` ([#10125](https://github.com/terraform-providers/terraform-provider-aws/issues/10125))
* **New Resource:** `aws_quicksight_user` ([#10401](https://github.com/terraform-providers/terraform-provider-aws/issues/10401))

ENHANCEMENTS:

* resource/aws_glue_classifier: Add `csv_classifier` configuration block (support CSV classifiers) ([#9824](https://github.com/terraform-providers/terraform-provider-aws/issues/9824))
* resource/aws_waf_byte_match_set: Support resource import ([#10477](https://github.com/terraform-providers/terraform-provider-aws/issues/10477))
* resource/aws_waf_rate_based_rule: Support resource import ([#10475](https://github.com/terraform-providers/terraform-provider-aws/issues/10475))
* resource/aws_waf_rule: Add `tags` argument ([#10408](https://github.com/terraform-providers/terraform-provider-aws/issues/10408))
* resource/aws_waf_rule_group: Add `tags` argument ([#10408](https://github.com/terraform-providers/terraform-provider-aws/issues/10408))
* resource/aws_waf_web_acl: Add `tags` argument ([#10408](https://github.com/terraform-providers/terraform-provider-aws/issues/10408))

BUG FIXES:

* resource/aws_gamelift_fleet: Increase default deletion timeout to 20 minutes to match service timing ([#10443](https://github.com/terraform-providers/terraform-provider-aws/issues/10443))

## 2.32.0 (October 10, 2019)

NOTES:

* provider: The underlying Terraform codebase dependency for the provider SDK and acceptance testing framework has been migrated from `github.com/hashicorp/terraform` to `github.com/hashicorp/terraform-plugin-sdk`. They are functionality equivalent and this should only impact codebase development to switch imports. For more information see the [Terraform Plugin SDK page in the Extending Terraform documentation](https://www.terraform.io/docs/extend/plugin-sdk.html). ([#10367](https://github.com/terraform-providers/terraform-provider-aws/issues/10367))

ENHANCEMENTS:

* resource/aws_emr_instance_group: Add `configurations_json` argument ([#10426](https://github.com/terraform-providers/terraform-provider-aws/issues/10426))

BUG FIXES:

* provider: Fix session handling to correctly validate and use assume_role credentials ([#10379](https://github.com/terraform-providers/terraform-provider-aws/issues/10379))
* resource/aws_autoscaling_group: Batch ALB/NLB attachments and detachments by 10 to prevent API and rate limiting errors ([#10435](https://github.com/terraform-providers/terraform-provider-aws/issues/10435))
* resource/aws_emr_instance_group: Remove terminated instance groups from the Terraform state ([#10425](https://github.com/terraform-providers/terraform-provider-aws/issues/10425))
* resource/aws_s3_bucket: Prevent infinite deletion recursion with `force_destroy` argument and object keys with empty "directory" prefixes present since version 2.29.0 ([#10388](https://github.com/terraform-providers/terraform-provider-aws/issues/10388))
* resource/aws_vpc_endpoint_route_table_association: Fix resource import support ([#10454](https://github.com/terraform-providers/terraform-provider-aws/issues/10454))

## 2.31.0 (October 03, 2019)

NOTES:

* resource/aws_lambda_function: Environments using Lambda functions with VPC configurations should upgrade their Terraform AWS Provider to this version or later to appropriately handle the networking changes introduced by the [improved VPC networking for AWS Lambda functions](https://aws.amazon.com/blogs/compute/announcing-improved-vpc-networking-for-aws-lambda-functions/) deployment. These changes prevent proper deletion of EC2 Subnets and Security Groups for accounts and regions updated to the new Lambda networking infrastructure in older versions of the Terraform AWS Provider. Additional information and configuration workarounds for prior versions can be found in [this GitHub issue](https://github.com/terraform-providers/terraform-provider-aws/issues/10329).

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `tags` attribute ([#10307](https://github.com/terraform-providers/terraform-provider-aws/issues/10307))
* resource/aws_efs_filesystem: Support tag-on-create ([#10254](https://github.com/terraform-providers/terraform-provider-aws/issues/10254))
* resource/aws_eks_cluster: Add `tags` argument ([#10307](https://github.com/terraform-providers/terraform-provider-aws/issues/10307))
* resource/aws_mq_broker: Add `encryption_options` configuration block (support AWS and customer managed KMS CMKs) ([#10276](https://github.com/terraform-providers/terraform-provider-aws/issues/10276))

BUG FIXES:

* provider: Upstream AWS Go SDK fix for parsing AWS shared credentials files missing right-hand values ([#10310](https://github.com/terraform-providers/terraform-provider-aws/issues/10310))
* resource/aws_lb_listener_certificate: Retry `CertificateNotFound` errors on creation for eventual consistency ([#10294](https://github.com/terraform-providers/terraform-provider-aws/issues/10294))
* resource/aws_s3_bucket_object: Fix object deletion for non-versioned objects ([#10352](https://github.com/terraform-providers/terraform-provider-aws/issues/10352))
* resource/aws_security_group: Handle updated ENI description and longer deletion timeframe for new Lambda Hyperplane ENIs ([#10114](https://github.com/terraform-providers/terraform-provider-aws/issues/10114)] / [[#10347](https://github.com/terraform-providers/terraform-provider-aws/issues/10347))
* resource/aws_subnet: Handle updated ENI description and longer deletion timeframe for new Lambda Hyperplane ENIs ([#10114](https://github.com/terraform-providers/terraform-provider-aws/issues/10114)] / [[#10347](https://github.com/terraform-providers/terraform-provider-aws/issues/10347))
* resource/aws_vpc_peering_connection: Ensure `allow_remote_vpc_dns_resolution` usage works with inter-region peering ([#7627](https://github.com/terraform-providers/terraform-provider-aws/issues/7627))
* resource/aws_vpc_peering_connection_accepter: Ensure `allow_remote_vpc_dns_resolution` usage works with inter-region peering ([#7627](https://github.com/terraform-providers/terraform-provider-aws/issues/7627))
* resource/aws_vpc_peering_connection_options: Ensure `allow_remote_vpc_dns_resolution` usage works with inter-region peering ([#7627](https://github.com/terraform-providers/terraform-provider-aws/issues/7627))
* resource/aws_waf_rate_based_rule: Upstream AWS Go SDK fix to allow `rate_limit` arguments between 100 and 1999 ([#10310](https://github.com/terraform-providers/terraform-provider-aws/issues/10310))
* resource/aws_wafregional_rate_based_rule: Upstream AWS Go SDK fix to allow `rate_limit` arguments between 100 and 1999 ([#10310](https://github.com/terraform-providers/terraform-provider-aws/issues/10310))
* resource/aws_wafregional_web_acl_association: Ensure missing resource triggers state removal ([#10216](https://github.com/terraform-providers/terraform-provider-aws/issues/10216))
* service/waf: Prevent incorrect `Error getting WAF change token` errors for API calls that should be retried or specially handled ([#10242](https://github.com/terraform-providers/terraform-provider-aws/issues/10242))
* service/wafregional: Prevent incorrect `Error getting WAF regional change token` errors for API calls that should be retried or specially handled ([#10242](https://github.com/terraform-providers/terraform-provider-aws/issues/10242))

## 2.30.0 (September 26, 2019)

NOTES:

* provider: The default development, testing, and building of the Terraform AWS Provider binary is now done with Go 1.13. This version of Go now requires macOS 10.11 El Capitan or later and FreeBSD 11.2 or later. Support for previous versions of those operating systems has been discontinued. ([#10206](https://github.com/terraform-providers/terraform-provider-aws/issues/10206))
* provider: The actual Terraform version running the provider will now be included the AWS Go SDK `User-Agent` headers for Terraform 0.12 and later. Terraform 0.11 and earlier will use `Terraform/0.11+compatible` as this information was not accessible in those versions. Previously, the Terraform version in the `User-Agent` header was based on the github.com/hashicorp/terraform dependency in the provider codebase. ([#9570](https://github.com/terraform-providers/terraform-provider-aws/issues/9570))

ENHANCEMENTS:

* data-source/aws_cloudtrail_service_account: Support `cn-north-1` region ([#10134](https://github.com/terraform-providers/terraform-provider-aws/issues/10134))
* data-source/aws_elastic_beanstalk_hosted_zone: Support `ap-east-1`, `ap-northeast-3`, `us-gov-east-1` and `us-gov-west-1` regions ([#10134](https://github.com/terraform-providers/terraform-provider-aws/issues/10134))
* data-source/aws_elb_hosted_zone_id: Support `cn-northwest-1` region  ([#10134](https://github.com/terraform-providers/terraform-provider-aws/issues/10134))
* data-source/aws_redshift_service_account: Support `ap-northeast-3`, `cn-north-1`, `eu-north-1` and `me-south-1` regions ([#10134](https://github.com/terraform-providers/terraform-provider-aws/issues/10134))
* provider: Use real Terraform version in User-Agent header ([#9570](https://github.com/terraform-providers/terraform-provider-aws/issues/9570))
* resource/aws_appsync_graphql_api: Add `additional_authentication_providers` configuration blocks ([#8587](https://github.com/terraform-providers/terraform-provider-aws/issues/8587))
* resource/aws_elastic_beanstalk_environment: Add `endpoint_url` attribute ([#10015](https://github.com/terraform-providers/terraform-provider-aws/issues/10015))
* resource/aws_lightsail_static_ip_attachment: Add `ip_address` attribute ([#10109](https://github.com/terraform-providers/terraform-provider-aws/issues/10109))
* resource/aws_opsworks_stack: Switch legacy Opsworks client User-Agent to real Terraform version ([#10246](https://github.com/terraform-providers/terraform-provider-aws/issues/10246))
* resource/aws_sns_topic_policy: Support resource import ([#10163](https://github.com/terraform-providers/terraform-provider-aws/issues/10163))
* resource/aws_sqs_queue: Support tag-on-create in AWS Commercial regions ([#10156](https://github.com/terraform-providers/terraform-provider-aws/issues/10156))

BUG FIXES:

* data-source/aws_elb_hosted_zone_id: Correct value for `cn-north-1` region ([#10134](https://github.com/terraform-providers/terraform-provider-aws/issues/10134))
* resource/aws_ec2_client_vpn_endpoint: Ensure missing resource triggers state removal ([#10187](https://github.com/terraform-providers/terraform-provider-aws/issues/10187))
* resource/aws_instance: Prevent panic when updating `credit_specification` to empty configuration block ([#10212](https://github.com/terraform-providers/terraform-provider-aws/issues/10212))
* resource/aws_security_group: Ensure deletion errors are properly raised ([#10165](https://github.com/terraform-providers/terraform-provider-aws/issues/10165))
* resource/aws_spot_fleet_request: Ensure `launch_specification` configuration block `placement_group` argument is passed through to the API when it is specified ([#10103](https://github.com/terraform-providers/terraform-provider-aws/issues/10103))

## 2.29.0 (September 20, 2019)

ENHANCEMENTS:
* data-source/aws_s3_bucket_object: Add `object_lock_legal_hold_status`, `object_lock_mode` and `object_lock_retain_until_date` attributes ([#9942](https://github.com/terraform-providers/terraform-provider-aws/issues/9942))
* resource/aws_glue_job: Add ability to specify python version for pythonshell in glue jobs ([#9409](https://github.com/terraform-providers/terraform-provider-aws/issues/9409))
* resource/aws_s3_bucket_object: Add `force_destroy`, `object_lock_legal_hold_status`, `object_lock_mode` and `object_lock_retain_until_date` attributes ([#9942](https://github.com/terraform-providers/terraform-provider-aws/issues/9942))
* resource/aws_ssm_association: Add import support ([#10055](https://github.com/terraform-providers/terraform-provider-aws/issues/10055))
* resource/aws_waf_rate_based_rule: Update rate based rule limit for WAF ([#9946](https://github.com/terraform-providers/terraform-provider-aws/pull/9946))
* resource/aws_wafregional_rate_based_rule: Update rate based rule limit for WAF ([#9946](https://github.com/terraform-providers/terraform-provider-aws/pull/9946))

BUG FIXES:

* resource/aws_ecs_task_definition: Fix a crash if `containers_definition` argument JSON defines `environment` without `name` value ([#10074](https://github.com/terraform-providers/terraform-provider-aws/issues/10074))

## 2.28.1 (September 12, 2019)

BUG FIXES:

* Revert "resource/aws_cloudfront_distribution: Fix `active_trusted_signers` attribute for Terraform 0.12" ([#10093](https://github.com/terraform-providers/terraform-provider-aws/issues/10093))

## 2.28.0 (September 12, 2019)

NOTES:

* resource/aws_cloudfront_distribution: This attribute implemented a legacy Terraform library (flatmap), which does not work with Terraform 0.12's data types and whose only usage was on this single attribute across all Terraform Providers. The attribute now implements (in the closest approximation to the previous implementation) the nested object data into the Terraform state in all Terraform versions. Any references to nested attributes such as `active_trusted_signers.enabled` will need to be updated to `active_trusted_signers.0.enabled`. ([#10013](https://github.com/terraform-providers/terraform-provider-aws/issues/10013))

FEATURES:

* **New Data Source:** `aws_route53_resolver_rule` ([#9805](https://github.com/terraform-providers/terraform-provider-aws/issues/9805))
* **New Data Source:** `aws_route53_resolver_rules` ([#9805](https://github.com/terraform-providers/terraform-provider-aws/issues/9805))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `identity` attribute (support getting OIDC issuer URL) ([#10006](https://github.com/terraform-providers/terraform-provider-aws/issues/10006))
* resource/aws_eks_cluster: Add `identity` attribute (support getting OIDC issuer URL) ([#10006](https://github.com/terraform-providers/terraform-provider-aws/issues/10006))
* resource/aws_elasticache_cluster: Support `cluster_id` validation up to 50 characters ([#9941](https://github.com/terraform-providers/terraform-provider-aws/issues/9941))
* resource/aws_elasticache_replication_group: Support `replication_group_id` validation up to 40 characters ([#9941](https://github.com/terraform-providers/terraform-provider-aws/issues/9941))

BUG FIXES:

* resource/aws_instance: Final retries after timeouts creating and updating instance and getting instance password data
* resource/aws_cloudfront_distribution: Support accessing `active_trusted_signers` attribute `items` in Terraform 0.12 ([#10013](https://github.com/terraform-providers/terraform-provider-aws/issues/10013))
* resource/aws_cognito_user_pool: Fix perpetual diffs on `sms_verification_message` ([#9758](https://github.com/terraform-providers/terraform-provider-aws/issues/9758))
* resource/aws_elasticsearch_domain: Final retries after timeouts creating, updating, and deleting domains ([#9892](https://github.com/terraform-providers/terraform-provider-aws/issues/9892))
* resource/aws_elasticsearch_domain_policy: Final retries after timeouts upserting and deleting domain policies ([#9892](https://github.com/terraform-providers/terraform-provider-aws/issues/9892))
* resource/aws_iam_policy_attachment: Revert a change causing errors with policies not being found during attachment ([#10063](https://github.com/terraform-providers/terraform-provider-aws/issues/10063))
* resource/aws_lightsail_instance: Fixes an issue where 2-character lightsail instance names didn't get validated properly ([#10046](https://github.com/terraform-providers/terraform-provider-aws/issues/10046))


## 2.27.0 (September 05, 2019)

ENHANCEMENTS:

* data-source/aws_ecs_cluster: Add `setting` attribute ([#9720](https://github.com/terraform-providers/terraform-provider-aws/issues/9720))
* provider: Support AWS C2S and SC2S endpoints ([#9998](https://github.com/terraform-providers/terraform-provider-aws/issues/9998))
* resource/aws_ecs_cluster: Add `setting` configuration blocks (support enabling Container Insights) ([#9720](https://github.com/terraform-providers/terraform-provider-aws/issues/9720))
* resource/aws_kinesis_firehose_delivery_stream: Add `server_side_encryption` configuration block (support Server Side Encryption) ([#6523](https://github.com/terraform-providers/terraform-provider-aws/issues/6523))

BUG FIXES:

* resource/aws_s3_bucket: Include any system tags that Terraform ignores when setting S3 bucket tags ([#7342](https://github.com/terraform-providers/terraform-provider-aws/issues/7342))

## 2.26.0 (August 29, 2019)

FEATURES:

* **New Data Source:** `aws_elasticsearch_domain` ([#1867](https://github.com/terraform-providers/terraform-provider-aws/issues/1867))

BUG FIXES:

* resource/aws_ec2_capacity_reservation: Fixes error handling when an EC2 Capacity Reservation is deleted manually but is still in state ([#9862](https://github.com/terraform-providers/terraform-provider-aws/issues/9862))
* resource/aws_s3_bucket: Final retries after timeouts creating, updating and updating replication configuration for s3 buckets ([#9861](https://github.com/terraform-providers/terraform-provider-aws/issues/9861))
* resource/aws_s3_bucket_inventory: Final retries after timeout reading and putting bucket inventories ([#9861](https://github.com/terraform-providers/terraform-provider-aws/issues/9861))
* resource/aws_s3_bucket_metric: Final retry after timeout putting bucket metric ([#9861](https://github.com/terraform-providers/terraform-provider-aws/issues/9861))
* resource/aws_s3_bucket_notification: Final retry after timeout putting notification ([#9861](https://github.com/terraform-providers/terraform-provider-aws/issues/9861))
* resource/aws_s3_bucket_policy: Final retry after timeout putting policy ([#9861](https://github.com/terraform-providers/terraform-provider-aws/issues/9861))
* resource/aws_s3_bucket_public_access_block: Final retries after timeouts creating and reading blocks ([#9861](https://github.com/terraform-providers/terraform-provider-aws/issues/9861))

## 2.25.0 (August 23, 2019)

ENHANCEMENTS:

* resource/aws_rds_cluster: Support `postgresql` in plan time validation for `enabled_cloudwatch_logs_exports` argument ([#9740](https://github.com/terraform-providers/terraform-provider-aws/issues/9740))

BUG FIXES:

* resource/aws_cloudwatch_event_target: Add default setting for ecs_target task_count ([#9773](https://github.com/terraform-providers/terraform-provider-aws/issues/9773))
* resource/aws_cloudwatch_log_subscription_filter: Prevent difference when omitting default `distribution` argument value of `ByLogStream` ([#9265](https://github.com/terraform-providers/terraform-provider-aws/issues/9265))
* resource/aws_db_instance: Fix enabling Enhanced Monitoring on update to handle IAM eventual consistency ([#9747](https://github.com/terraform-providers/terraform-provider-aws/issues/9747))
* resource/aws_elb: Final retries after timeouts creating and updating ELBs ([#9765](https://github.com/terraform-providers/terraform-provider-aws/issues/9765))
* resource/aws_elb_attachment: Final retry after timout creating ELB attachment ([#9765](https://github.com/terraform-providers/terraform-provider-aws/issues/9765))
* resource/aws_iam_instance_profile: Final retry after timeout adding role to profile ([#9766](https://github.com/terraform-providers/terraform-provider-aws/issues/9766))
* resource/aws_iam_policy: Final retry after timeout reading policy ([#9766](https://github.com/terraform-providers/terraform-provider-aws/issues/9766))
* resource/aws_iam_role: Final retries after timeouts creating and deleting IAM roles ([#9766](https://github.com/terraform-providers/terraform-provider-aws/issues/9766))
* resource/aws_iam_user: Final retry after timeout deleting user login profile ([#9766](https://github.com/terraform-providers/terraform-provider-aws/issues/9766))
* resource/aws_inspector_assessment_target: Final retry after timeout deleting target ([#9767](https://github.com/terraform-providers/terraform-provider-aws/issues/9767))
* resource/aws_internet_gateway: Final retries after timeouts creating, attaching, and deleting gateways ([#9779](https://github.com/terraform-providers/terraform-provider-aws/issues/9779))
* resource/aws_iot_thing_type: Final retry after timeout deleting IOT thing type ([#9780](https://github.com/terraform-providers/terraform-provider-aws/issues/9780))
* resource/aws_kinesis_firehose_delivery_stream: Prevent differences with disabled `data_format_conversion_configuration` and `processing_configuration` after changes outside Terraform ([#9103](https://github.com/terraform-providers/terraform-provider-aws/issues/9103))
* resource/aws_launch_configuration: Final retry after timeout creating launch configuration ([#9781](https://github.com/terraform-providers/terraform-provider-aws/issues/9781))
* resource/aws_lb: Final retry after timeout waiting for network interfaces to detach ([#9787](https://github.com/terraform-providers/terraform-provider-aws/issues/9787))
* resource/aws_lb_listener_certificate: Final retry after timeout reading listener certificate ([#9787](https://github.com/terraform-providers/terraform-provider-aws/issues/9787))
* resource/aws_lb_listener_rule: Final retries after timeout reading and creating listener rules ([#9787](https://github.com/terraform-providers/terraform-provider-aws/issues/9787))
* resource/aws_msk_cluster: Final retries after timeouts creating and deleting clusters ([#9793](https://github.com/terraform-providers/terraform-provider-aws/issues/9793))
* resource/aws_network_acl: Final retry after timeout deleting ACLs ([#9830](https://github.com/terraform-providers/terraform-provider-aws/issues/9830))
* resource/aws_network_acl_rule: Final retry after timeout creating ACL rules ([#9830](https://github.com/terraform-providers/terraform-provider-aws/issues/9830))
* resource/aws_network_acl_rule: Remove resource from Terraform state on `InvalidNetworkAclID.NotFound` errors ([#9710](https://github.com/terraform-providers/terraform-provider-aws/issues/9710))
* resource/aws_opsworks_stack: Final retry after timeout creating stack ([#9818](https://github.com/terraform-providers/terraform-provider-aws/issues/9818))
* resource/aws_rds_cluster_instance: Ensure `monitoring_interval` and `monitoring_role_arn` attributes are always written to the Terraform state ([#9748](https://github.com/terraform-providers/terraform-provider-aws/issues/9748))
* resource/aws_redshift_cluster: Final retry after timeout deleting cluster ([#9796](https://github.com/terraform-providers/terraform-provider-aws/issues/9796))
* resource/aws_redshift_snapshot_copy_grant: Final retries after timeouts finding and deleting grants ([#9796](https://github.com/terraform-providers/terraform-provider-aws/issues/9796))
* resource/aws_route: Final retry after timeout creating route ([#9797](https://github.com/terraform-providers/terraform-provider-aws/issues/9797))
* resource/aws_route_table: Final retry after timeout updating route table ([#9797](https://github.com/terraform-providers/terraform-provider-aws/issues/9797))
* resource/aws_route_table_association: Final retry after timeout creating route table association ([#9797](https://github.com/terraform-providers/terraform-provider-aws/issues/9797))
* resource/aws_s3_bucket_object: Allow using SSE-S3 encryption with `etag` argument ([#9442](https://github.com/terraform-providers/terraform-provider-aws/issues/9442))
* resource/aws_sagemaker_model: Final retry after timeout deleting model ([#9799](https://github.com/terraform-providers/terraform-provider-aws/issues/9799))
* resource/aws_sagemaker_notebook_instance: Final retry after timeout updating instance ([#9799](https://github.com/terraform-providers/terraform-provider-aws/issues/9799))
* resource/aws_security_group: Final retry after timeout deleting security group ([#9812](https://github.com/terraform-providers/terraform-provider-aws/issues/9812))
* resource/aws_security_group_rule: Final retry after timeout creating security group rule ([#9812](https://github.com/terraform-providers/terraform-provider-aws/issues/9812))
* resource/aws_sqs_queue: Final retry after timeout creating queue ([#9813](https://github.com/terraform-providers/terraform-provider-aws/issues/9813))
* resource/aws_sqs_queue_policy: Final retry after timeout updating queue policy ([#9813](https://github.com/terraform-providers/terraform-provider-aws/issues/9813))
* resource/aws_transfer_server: Final retry after timeout waiting for transfer server deletion ([#9815](https://github.com/terraform-providers/terraform-provider-aws/issues/9815))
* resource/aws_wafregional_web_acl_association: Final retry after timeout creating association ([#9820](https://github.com/terraform-providers/terraform-provider-aws/issues/9820))
* service/dynamodb: Final retries after timeouts setting dynamodb tags ([#9821](https://github.com/terraform-providers/terraform-provider-aws/issues/9821))
* service/sagemaker: Final retries after timeouts setting sagemaker tags ([#9821](https://github.com/terraform-providers/terraform-provider-aws/issues/9821))
* service/waf: Final retry after timeout getting change token ([#9826](https://github.com/terraform-providers/terraform-provider-aws/issues/9826))
* service/wafregional: Final retry after timeout getting change token ([#9826](https://github.com/terraform-providers/terraform-provider-aws/issues/9826))

## 2.24.0 (August 15, 2019)

FEATURES:

* **New Resource:** `aws_config_organization_custom_rule` ([#9716](https://github.com/terraform-providers/terraform-provider-aws/issues/9716))
* **New Resource:** `aws_config_organization_managed_rule` ([#9716](https://github.com/terraform-providers/terraform-provider-aws/issues/9716))
* **New Resource:** `aws_fsx_lustre_file_system` ([#7074](https://github.com/terraform-providers/terraform-provider-aws/issues/7074)] / [[#9761](https://github.com/terraform-providers/terraform-provider-aws/issues/9761))
* **New Resource:** `aws_fsx_windows_file_system` ([#7074](https://github.com/terraform-providers/terraform-provider-aws/issues/7074)] / [[#9761](https://github.com/terraform-providers/terraform-provider-aws/issues/9761))
* **New Resource:** `aws_ram_resource_share_accepter` ([#8259](https://github.com/terraform-providers/terraform-provider-aws/issues/8259))

ENHANCEMENTS:

* resource/aws_codebuild_project: Add `artifacts` configuration block `artifact_identifier` argument ([#9652](https://github.com/terraform-providers/terraform-provider-aws/issues/9652))
* resource/aws_codebuild_project: Add plan time validation for `artifacts` and `secondary_artifacts` configuration blocks `packaging` argument ([#9652](https://github.com/terraform-providers/terraform-provider-aws/issues/9652))
* resource/aws_rds_cluster: Add `multimaster` to `engine_mode` argument validation (support Aurora Multi-Master Clusters) ([#9691](https://github.com/terraform-providers/terraform-provider-aws/issues/9691))
* resource/aws_rds_cluster_instance: Allow `aurora-mysql` (MySQL 5.7) engine to enable Performance Insights ([#9635](https://github.com/terraform-providers/terraform-provider-aws/issues/9635))
* resource/aws_wafregional_regex_match_set: Support resource import ([#9699](https://github.com/terraform-providers/terraform-provider-aws/issues/9699))
* resource/aws_wafregional_regex_pattern_set: Support resource import ([#9712](https://github.com/terraform-providers/terraform-provider-aws/issues/9712))
* resource/aws_wafregional_size_constraint_set: Support resource import ([#9713](https://github.com/terraform-providers/terraform-provider-aws/issues/9713))
* resource/aws_wafregional_sql_injection_match_set: Support resource import ([#9717](https://github.com/terraform-providers/terraform-provider-aws/issues/9717))

BUG FIXES:

* resource/aws_acm_certificate_validation: Final retries after timeouts creating and checking validation for ACM certificates ([#9661](https://github.com/terraform-providers/terraform-provider-aws/issues/9661))
* resource/aws_ami: Final retry after timeout reading AMI ([#9674](https://github.com/terraform-providers/terraform-provider-aws/issues/9674))
* resource/aws_cloudhsm_v2_cluster: Final retries after timeouts creating, updating, and deleting CloudHSM clusters ([#9675](https://github.com/terraform-providers/terraform-provider-aws/issues/9675))
* resource/aws_cloudhsm_v2_hsm: Final retries after timeouts creating and deleting CloudHSM modules ([#9675](https://github.com/terraform-providers/terraform-provider-aws/issues/9675))
* resource/aws_cloudtrail: Final retries after timeouts creating and updating cloudtrails ([#9678](https://github.com/terraform-providers/terraform-provider-aws/issues/9678))
* resource/aws_codebuild_project: Final retries after timeouts creating and updating codebuild projects ([#9682](https://github.com/terraform-providers/terraform-provider-aws/issues/9682))
* resource/aws_codebuild_project: Properly perform drift detection and updates for `secondary_artifacts` configuration block arguments (except `name` which will require a separate fix) ([#9652](https://github.com/terraform-providers/terraform-provider-aws/issues/9652))
* resource/aws_codedeploy_deployment_group: Final retries after timeouts creating and updating deployment groups ([#9682](https://github.com/terraform-providers/terraform-provider-aws/issues/9682))
* resource/aws_codepipeline: Final retry after timeout creating codepipeline ([#9682](https://github.com/terraform-providers/terraform-provider-aws/issues/9682))
* resource/aws_cognito_user_pool: Final retries after timeouts creating and updating Cognito user pools ([#9684](https://github.com/terraform-providers/terraform-provider-aws/issues/9684))
* resource/aws_db_instance: Fix enabling Performance Insights on update without Performance Insights KMS Key ID ([#9745](https://github.com/terraform-providers/terraform-provider-aws/issues/9745))
* resource/aws_dms_endpoint: Final retry after timeout creating DMS endpoint ([#9695](https://github.com/terraform-providers/terraform-provider-aws/issues/9695))
* resource/aws_docdb_cluster_instance: Final retries after timeouts creating and updating DocDB cluster instances ([#9696](https://github.com/terraform-providers/terraform-provider-aws/issues/9696))
* resource/aws_docdb_cluster_parameter_group: Final retry after timeout deleting DocDB cluster parameter groups ([#9696](https://github.com/terraform-providers/terraform-provider-aws/issues/9696))
* resource/aws_docdb_subnet_group: Final retry after timeout deleting DocDB subnet groups ([#9696](https://github.com/terraform-providers/terraform-provider-aws/issues/9696))
* resource/aws_dynamodb_table: Final retries after timeouts creating, updating, and deleting DynamoDB tables ([#9697](https://github.com/terraform-providers/terraform-provider-aws/issues/9697))
* resource/aws_ebs_snapshot: Final retries after timeouts creating, deleting or waiting for available EBS snapshots ([#9698](https://github.com/terraform-providers/terraform-provider-aws/issues/9698))
* resource/aws_ebs_snapshot_copy: Final retry after timeout deleting EBS snapshot copies ([#9698](https://github.com/terraform-providers/terraform-provider-aws/issues/9698))
* resource/aws_ecs_cluster: Final retries after timeouts reading and deleting ECS cluster ([#9704](https://github.com/terraform-providers/terraform-provider-aws/issues/9704))
* resource/aws_ecs_service: Final retries after timeouts creating, updating, and deleting ECS service ([#9704](https://github.com/terraform-providers/terraform-provider-aws/issues/9704))
* resource/aws_eip: Final retries after timeouts reading, updating, and deleting EIPs ([#9728](https://github.com/terraform-providers/terraform-provider-aws/issues/9728))
* resource/aws_eip_association: Final retry after timeout creating EIP association ([#9728](https://github.com/terraform-providers/terraform-provider-aws/issues/9728))
* resource/aws_eks_cluster: Final retry after timeout creating EKS cluster ([#9729](https://github.com/terraform-providers/terraform-provider-aws/issues/9729))
* resource/aws_elastic_beanstalk_application: Final retries after timeouts reading and deleting beanstalk applications ([#9731](https://github.com/terraform-providers/terraform-provider-aws/issues/9731))
* resource/aws_gamelift_build: Final retry after timeout creating gamelift build ([#9752](https://github.com/terraform-providers/terraform-provider-aws/issues/9752))
* resource/aws_gamelift fleet: Final retry after timeout deleting gamelift fleet ([#9752](https://github.com/terraform-providers/terraform-provider-aws/issues/9752))
* resource/aws_glue_crawler: Final retry after timeout creating glue crawler ([#9753](https://github.com/terraform-providers/terraform-provider-aws/issues/9753))
* resource/aws_guardduty_member: Final retry after timeout waiting for email invitation ([#9757](https://github.com/terraform-providers/terraform-provider-aws/issues/9757))
* resource/aws_lb_target_group_attachment: Perform drift detection on attachments using target health description (trigger resource recreation for manually deregistered attachments) ([#9610](https://github.com/terraform-providers/terraform-provider-aws/issues/9610))
* resource/aws_vpn_gateway: Retry after timeouts attaching and deleting VPN gateways, and retrying attachment after pending VPN errors ([#9641](https://github.com/terraform-providers/terraform-provider-aws/issues/9641))

## 2.23.0 (August 07, 2019)

FEATURES:

* **New Data Source:** `aws_s3_bucket_objects` ([#6968](https://github.com/terraform-providers/terraform-provider-aws/issues/6968))
* **New Resource:** `aws_dx_transit_virtual_interface` ([#8522](https://github.com/terraform-providers/terraform-provider-aws/issues/8522))
* **New Resource:** `aws_redshift_snapshot_schedule` ([#8064](https://github.com/terraform-providers/terraform-provider-aws/issues/8064))
* **New Resource:** `aws_redshift_snapshot_schedule_association` ([#8064](https://github.com/terraform-providers/terraform-provider-aws/issues/8064))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `status` attribute ([#9582](https://github.com/terraform-providers/terraform-provider-aws/issues/9582))
* data-source/aws_instance: Add `ebs_block_device` and `root_block_device` configuration block `encryption` and `kms_key_id` attributes ([#4861](https://github.com/terraform-providers/terraform-provider-aws/issues/4861)] / [[#7757](https://github.com/terraform-providers/terraform-provider-aws/issues/7757))
* data-source/aws_partition: Add `dns_suffix` attribute (e.g. `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China) ([#5602](https://github.com/terraform-providers/terraform-provider-aws/issues/5602))
* resource/aws_acm_certificate: Support `options` configuration block `certificate_transparency_logging_preference` argument ([#9413](https://github.com/terraform-providers/terraform-provider-aws/issues/9413))
* resource/aws_acm_certificate: Add `certificate_authority_arn` argument (support issuance of ACM private certificates) ([#6666](https://github.com/terraform-providers/terraform-provider-aws/issues/6666))
* resource/aws_cognito_identity_pool: Add `tags` argument ([#9639](https://github.com/terraform-providers/terraform-provider-aws/issues/9639))
* resource/aws_ecr_repository: Add `image_tag_mutability` argument (support immutable image tags) ([#9557](https://github.com/terraform-providers/terraform-provider-aws/issues/9557))
* resource/aws_efs_file_system: Add `lifecycle_policy` configuration block (support transition to IA storage after 14, 30, 60, or 90 days) ([#9636](https://github.com/terraform-providers/terraform-provider-aws/issues/9636))
* resource/aws_eks_cluster: Add `status` attribute ([#9582](https://github.com/terraform-providers/terraform-provider-aws/issues/9582))
* resource/aws_glue_crawler: Add `catalog_target` configuration block ([#9430](https://github.com/terraform-providers/terraform-provider-aws/issues/9430))
* resource/aws_instance: Add `ebs_block_device` and `root_block_device` configuration block `encryption` and `kms_key_id` arguments (support encryption on launch) ([#4861](https://github.com/terraform-providers/terraform-provider-aws/issues/4861)] / [[#7757](https://github.com/terraform-providers/terraform-provider-aws/issues/7757))
* resource/aws_iot_certificate: Mark `csr` argument as optional and add `certificate_pem`, `public_key`, and `private_key` attributes (support creating key and certificate) ([#9283](https://github.com/terraform-providers/terraform-provider-aws/issues/9283))
* resource/aws_lambda_permission: Support resource import ([#9369](https://github.com/terraform-providers/terraform-provider-aws/issues/9369))
* resource/aws_launch_configuration: Add `root_block_device` configuration block `encrypted` argument (support encryption on launch) ([#7759](https://github.com/terraform-providers/terraform-provider-aws/issues/7759))
* resource/aws_s3_bucket_object: Added plan time validation for `bucket` and `key` arguments ([#9591](https://github.com/terraform-providers/terraform-provider-aws/issues/9591))
* resource/aws_spot_fleet_request: Add `ebs_block_device` and `root_block_device` configuration block `kms_key_id` argument (support encryption on launch) ([#9599](https://github.com/terraform-providers/terraform-provider-aws/issues/9599))
* resource/aws_wafregional_geo_match_set: Support resource import ([#9620](https://github.com/terraform-providers/terraform-provider-aws/issues/9620))
* resource/aws_wafregional_rate_based_rule: Support resource import ([#9621](https://github.com/terraform-providers/terraform-provider-aws/issues/9621))

BUG FIXES:

* provider: Environment credentials have precedence over shared config credentials even if the `AWS_PROFILE` environment credentials are present. Explicitly configure the provider `profile` to override this behavior. The AWS Go SDK change that was released in version 2.21.0 of the Terraform AWS Provider has been mostly reverted. ([#9555](https://github.com/terraform-providers/terraform-provider-aws/issues/9555))
* resource/aws_acm_certificate: Wait for presence of `DomainValidationOptions` when requesting ACM certificates (previously the API would always immediately return this information during creation) ([#9598](https://github.com/terraform-providers/terraform-provider-aws/issues/9598))
* resource/aws_autoscaling_group: Final retries after timeouts creating, draining, and deleting ASGs and autoscaling helpers ([#9649](https://github.com/terraform-providers/terraform-provider-aws/issues/9649))
* resource/aws_cloud9_environment_ec2: Final retries after timeouts creating and deleting Cloud9 environments ([#9629](https://github.com/terraform-providers/terraform-provider-aws/issues/9629))
* resource/aws_cloudfront_distribution: Ensure deployment timeout matches documentation at 90 minutes ([#9642](https://github.com/terraform-providers/terraform-provider-aws/issues/9642))
* resource/aws_datasync_agent: Final retries after timeouts creating datasync agent ([#9608](https://github.com/terraform-providers/terraform-provider-aws/issues/9608))
* resource/aws_datasync_task: Final retry after timeout error creating datasync task ([#9608](https://github.com/terraform-providers/terraform-provider-aws/issues/9608))
* resource/aws_dax_cluster: Final retries after timeouts when creating and deleting Dax clusters ([#9630](https://github.com/terraform-providers/terraform-provider-aws/issues/9630))
* resource/aws_egress_only_internet_gateway: Final retry after timeout when reading gateway ([#9638](https://github.com/terraform-providers/terraform-provider-aws/issues/9638))
* resource/aws_lambda_event_source_mapping: Final retries after timeout when creating, updating, and deleting event source mappings ([#9553](https://github.com/terraform-providers/terraform-provider-aws/issues/9553))
* resource/aws_lambda_function: Final retry when creating lambda function ([#9553](https://github.com/terraform-providers/terraform-provider-aws/issues/9553))
* resource/aws_lambda_permission: Final retries when creating, reading, and deleting lambda permissions ([#9553](https://github.com/terraform-providers/terraform-provider-aws/issues/9553))
* resource/aws_media_package_channel: Final retries after timeouts deleting media package channels ([#9633](https://github.com/terraform-providers/terraform-provider-aws/issues/9633))
* resource/aws_media_store_container: Final retries after timeouts deleting media store containers ([#9633](https://github.com/terraform-providers/terraform-provider-aws/issues/9633))
* resource/aws_organizations_organizational_unit: Final retry after timeout when creating organizational unit ([#9631](https://github.com/terraform-providers/terraform-provider-aws/issues/9631))
* resource/aws_organizations_policy: Final retry after timeout creating policy ([#9631](https://github.com/terraform-providers/terraform-provider-aws/issues/9631))
* resource/aws_organizations_policy_attachment: Final retry after timeout creating policy attachment ([#9631](https://github.com/terraform-providers/terraform-provider-aws/issues/9631))
* resource/aws_secretsmanager_secret: Final retries after timeouts creating and updating secrets ([#9632](https://github.com/terraform-providers/terraform-provider-aws/issues/9632))
* resource/aws_sns_platform_application: Final retry after timeout error updating SNS platform application ([#9607](https://github.com/terraform-providers/terraform-provider-aws/issues/9607))
* resource/aws_vpc: Final retry after timeout deleting VPC ([#9644](https://github.com/terraform-providers/terraform-provider-aws/issues/9644))

## 2.22.0 (August 01, 2019)

NOTES:

* provider: Region validation now automatically supports the new `me-south-1` Middle East (Bahrain) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [previous new region announcement blog post](https://aws.amazon.com/blogs/aws/now-open-aws-asia-pacific-hong-kong-region/). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g. `data.aws_availability_zones.current: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). ([#9538](https://github.com/terraform-providers/terraform-provider-aws/issues/9538))

FEATURES:

* **New Resource:** `aws_codebuild_source_credential` ([#7631](https://github.com/terraform-providers/terraform-provider-aws/issues/7631))
* **New Resource:** `aws_fms_admin_account` ([#4310](https://github.com/terraform-providers/terraform-provider-aws/issues/4310))

ENHANCEMENTS:

* data-source/aws_cloudtrail_service_account: Support `me-south-1` region ([#9547](https://github.com/terraform-providers/terraform-provider-aws/issues/9547))
* data-source/aws_elastic_beanstalk_hosted_zone: Support `me-south-1` region ([#9547](https://github.com/terraform-providers/terraform-provider-aws/issues/9547))
* data-source/aws_elb_hosted_zone_id: Support `me-south-1` region ([#9547](https://github.com/terraform-providers/terraform-provider-aws/issues/9547))
* data-source/aws_elb_service_account: Support `me-south-1` region ([#9547](https://github.com/terraform-providers/terraform-provider-aws/issues/9547))
* data-source/aws_s3_bucket: Support `me-south-1` region for `hosted_zone_id` attribute ([#9547](https://github.com/terraform-providers/terraform-provider-aws/issues/9547))
* provider: Support automatic region validation for `me-south-1` ([#9538](https://github.com/terraform-providers/terraform-provider-aws/issues/9538))
* resource/aws_codebuild_project: Add `override_artifact_name` argument to `artifacts` and `secondary_artifacts` configuration blocks ([#7824](https://github.com/terraform-providers/terraform-provider-aws/issues/7824))
* resource/aws_config_aggregate_authorization: Add `tags` argument ([#9561](https://github.com/terraform-providers/terraform-provider-aws/issues/9561))
* resource/aws_config_config_rule: Add `tags` argument ([#9561](https://github.com/terraform-providers/terraform-provider-aws/issues/9561))
* resource/aws_config_configuration_aggregator: Add `tags` argument ([#9561](https://github.com/terraform-providers/terraform-provider-aws/issues/9561))
* resource/aws_ec2_client_vpn_endpoint: Add `split_tunnel` argument ([#9566](https://github.com/terraform-providers/terraform-provider-aws/issues/9566))
* resource/aws_ecs_service: Allow multiple `load_balancer` configuration blocks (support for multiple target groups) ([#9411](https://github.com/terraform-providers/terraform-provider-aws/issues/9411))
* resource/aws_pinpoint_app: Add `tags` argument ([#9460](https://github.com/terraform-providers/terraform-provider-aws/issues/9460))
* resource/aws_route_table_association: Support resource import ([#6999](https://github.com/terraform-providers/terraform-provider-aws/issues/6999))
* resource/aws_route_table_association: Allow in-place updates of `subnet_id` argument ([#6999](https://github.com/terraform-providers/terraform-provider-aws/issues/6999))
* resource/aws_s3_bucket: Support `me-south-1` region for `hosted_zone_id` attribute ([#9547](https://github.com/terraform-providers/terraform-provider-aws/issues/9547))

BUG FIXES:

* resource/aws_codebuild_project: Properly perform drift detection and updates for `artifacts` configuration block arguments ([#9559](https://github.com/terraform-providers/terraform-provider-aws/issues/9559))
* resource/aws_ec2_client_vpn_endpoint: Remove hardcoded one minute timeout during resource creation ([#9558](https://github.com/terraform-providers/terraform-provider-aws/issues/9558))
* resource/aws_route53_record: Prevent error when removing `weighted_routing_policy` ([#9565](https://github.com/terraform-providers/terraform-provider-aws/issues/9565))
* resource/aws_storagegateway_cached_iscsi_volume: Retry after timeout deleting volume ([#9536](https://github.com/terraform-providers/terraform-provider-aws/issues/9536))
* resource/aws_storagegateway_cached_iscsi_volume: Fix errors deleting volumes when volumes don't exist ([#9543](https://github.com/terraform-providers/terraform-provider-aws/issues/9543))
* resource/aws_storagegateway_gateway: Retry after timeouts creating gateway ([#9536](https://github.com/terraform-providers/terraform-provider-aws/issues/9536))

## 2.21.1 (July 26, 2019)

BUG FIXES:

* resource/aws_autoscaling_group: Revert change from version 2.21.0 to `load_balancers` and `target_group_arns` arguments that removes attachments when using the `aws_autoscaling_attachment` resource (https://github.com/terraform-providers/terraform-provider-aws/issues/9513) ([#9518](https://github.com/terraform-providers/terraform-provider-aws/issues/9518))

## 2.21.0 (July 25, 2019)

NOTES:

* provider: After this update, the AWS Go SDK will prefer credentials found via the `AWS_PROFILE` environment variable when both the `AWS_PROFILE` environment variable and the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables are statically defined. Previously the SDK would ignore the `AWS_PROFILE` environment variable, if static environment credentials were also specified. This is listed as a bug fix in the AWS Go SDK release notes. ([#9428](https://github.com/terraform-providers/terraform-provider-aws/issues/9428))

FEATURES:
* **New Data Source**: `aws_organizations_organization` ([#9419](https://github.com/terraform-providers/terraform-provider-aws/issues/9419))
* **New Data Source**: `aws_waf_ipset` ([#9481](https://github.com/terraform-providers/terraform-provider-aws/issues/9481))
* **New Data Source**: `aws_wafregional_ipset` ([#9484](https://github.com/terraform-providers/terraform-provider-aws/issues/9484))

ENHANCEMENTS:

* provider: Add support for assuming role via web identity token via the `AWS_WEB_IDENTITY_TOKEN_FILE` and `AWS_ROLE_ARN` environment variables ([#9428](https://github.com/terraform-providers/terraform-provider-aws/issues/9428))
* resource/aws_cloudwatch_event_target: Support resource import ([#9431](https://github.com/terraform-providers/terraform-provider-aws/issues/9431))
* resource/aws_s3_bucket_object: Add `metadata` argument ([#1945](https://github.com/terraform-providers/terraform-provider-aws/issues/1945))
* resource/aws_wafregional_ipset: Support resource import ([#9424](https://github.com/terraform-providers/terraform-provider-aws/issues/9424))

BUG FIXES:

* provider: Load credentials via the `AWS_PROFILE` environment variable (if available) when `AWS_PROFILE` is defined along with `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` ([#9428](https://github.com/terraform-providers/terraform-provider-aws/issues/9428))
* resource/aws_autoscaling_group: Always perform drift detection with `load_balancers` and `target_group_arns` arguments ([#9478](https://github.com/terraform-providers/terraform-provider-aws/issues/9478))
* resource/aws_cloudfront_distribution: Prevent `DistributionAlreadyExists` errors during concurrent distribution creation ([#9470](https://github.com/terraform-providers/terraform-provider-aws/issues/9470))
* resource/aws_cognito_user_pool_client: Properly update name value ([#9437](https://github.com/terraform-providers/terraform-provider-aws/issues/9437))
* resource/aws_config_config_rule: Retries after timeouts when creating and deleting config rules ([#9438](https://github.com/terraform-providers/terraform-provider-aws/issues/9438))
* resource/aws_config_delivery_channel: Retries after timeouts when creating and deleting config delivery channels ([#9438](https://github.com/terraform-providers/terraform-provider-aws/issues/9438))
* resource/aws_customer_gateway: Final retry after timeout deleting customer gateway ([#9421](https://github.com/terraform-providers/terraform-provider-aws/issues/9421))
* resource/aws_db_instance: Redact `MasterUserPassword` from user interface when displaying `InvalidParameterValue` error during resource creation ([#9446](https://github.com/terraform-providers/terraform-provider-aws/issues/9446))
* resource/aws_db_instance: Retries after timeouts creating DB instances ([#9477](https://github.com/terraform-providers/terraform-provider-aws/issues/9477))
* resource/aws_db_option_group: Retry after timeout deleting DB option group ([#9477](https://github.com/terraform-providers/terraform-provider-aws/issues/9477))
* resource/aws_db_parameter_group: Retry after timeout deleting DB parameter group ([#9477](https://github.com/terraform-providers/terraform-provider-aws/issues/9477))
* resource/aws_kms_grant: Final retries after timeouts when creating, finding, and revoking grants ([#9415](https://github.com/terraform-providers/terraform-provider-aws/issues/9415))
* resource/aws_kms_key: Final retries after timeouts when creating keys and updating key rotation status ([#9415](https://github.com/terraform-providers/terraform-provider-aws/issues/9415))
* resource/aws_rds_cluster: Properly update `master_password` during snapshot restore ([#9505](https://github.com/terraform-providers/terraform-provider-aws/issues/9505))
* resource/aws_s3_bucket: Ensure `website_endpoint` and `website_domain` attributes have correct DNS suffix in AWS China ([#9444](https://github.com/terraform-providers/terraform-provider-aws/issues/9444))
* resource/aws_ses_domain_identity_verification: Retry after timeout when creating SES domain identity verification ([#9417](https://github.com/terraform-providers/terraform-provider-aws/issues/9417))
* resource/aws_sfn_activity: Retry after timeout deleting SFN activity ([#9498](https://github.com/terraform-providers/terraform-provider-aws/issues/9498))
* resource/aws_sfn_state_machine: Retry after timeouts deleting and creating SFN state machines ([#9498](https://github.com/terraform-providers/terraform-provider-aws/issues/9498))

## 2.20.0 (July 19, 2019)

NOTES:

* resource/aws_ssm_maintenance_window_task: The `logging_info` and `task_parameters` configuration blocks have been deprecated in favor of a new `task_invocation_parameters` configuration block to match the API ([#7823](https://github.com/terraform-providers/terraform-provider-aws/issues/7823))

FEATURES:
* **New Data Source** `aws_waf_rule` ([#9318](https://github.com/terraform-providers/terraform-provider-aws/issues/9318))
* **New Data Source:** `aws_waf_web_acl` ([#9320](https://github.com/terraform-providers/terraform-provider-aws/issues/9320))
* **New Data Source** `aws_wafregional_rule` ([#9319](https://github.com/terraform-providers/terraform-provider-aws/issues/9319))
* **New Data Source:** `aws_wafregional_web_acl` ([#9321](https://github.com/terraform-providers/terraform-provider-aws/issues/9321))
* **New Resource:** `aws_quicksight_group` ([#8233](https://github.com/terraform-providers/terraform-provider-aws/issues/8233))
* **New Resource:** `aws_servicequotas_service_quota` ([#9192](https://github.com/terraform-providers/terraform-provider-aws/issues/9192))

ENHANCEMENTS:

* data-source/aws_route53_zone: Add `linked_service_principal` and `linked_service_description` attributes ([#9390](https://github.com/terraform-providers/terraform-provider-aws/issues/9390))
* provider: Support for assuming role using credential process from the shared AWS configuration file ([#9305](https://github.com/terraform-providers/terraform-provider-aws/issues/9305))
* resource/aws_api_gateway_domain_name: Add `security_policy` argument ([#9128](https://github.com/terraform-providers/terraform-provider-aws/issues/9128))
* resource/aws_athena_named_query: Add `workgroup` argument ([#9383](https://github.com/terraform-providers/terraform-provider-aws/issues/9383))
* resource/aws_autoscaling_lifecycle_hook: Support resource import ([#9336](https://github.com/terraform-providers/terraform-provider-aws/issues/9336))
* resource/aws_elasticsearch_domain: Add `cluster_config` configuration block `zone_awareness_config` configuration block (support three Availability Zone awareness) ([#9398](https://github.com/terraform-providers/terraform-provider-aws/issues/9398))
* resource/aws_emr_cluster: Add `master_instance_group` configuration block `instance_count` argument (support multiple master nodes) ([#9235](https://github.com/terraform-providers/terraform-provider-aws/issues/9235))
* resource/aws_media_store_container: Add `tags` argument ([#9379](https://github.com/terraform-providers/terraform-provider-aws/issues/9379))
* resource/aws_rds_cluster: Support `scaling_configuration` configuration block `timeout_action` argument ([#9374](https://github.com/terraform-providers/terraform-provider-aws/issues/9374))
* resource/aws_s3_bucket_object: Allow empty object ([#7544](https://github.com/terraform-providers/terraform-provider-aws/issues/7544))
* resource/aws_ssm_maintenance_window_task: Support resource import and in-place updates ([#7823](https://github.com/terraform-providers/terraform-provider-aws/issues/7823))
* resource/aws_ssm_maintenance_window_task: Add `task_invocation_parameters` configuration block and deprecate `logging_info` and `task_parameters` configuration blocks to match API ([#7823](https://github.com/terraform-providers/terraform-provider-aws/issues/7823))

BUG FIXES:

* resource/aws_appautoscaling_policy: Properly support importing of dynamodb policies ([#8397](https://github.com/terraform-providers/terraform-provider-aws/issues/8397))
* resource/aws_cloudwatch_event_permissions: Clean up error handling when reading event permissions ([#9065](https://github.com/terraform-providers/terraform-provider-aws/issues/9065))
* resource/aws_cloudwatch_event_rule: Retry error handling when creating and updating event rules ([#9065](https://github.com/terraform-providers/terraform-provider-aws/issues/9065))
* resource/aws_cloudwatch_log_destination: Clean up error handling when putting log destination ([#9065](https://github.com/terraform-providers/terraform-provider-aws/issues/9065))
* resource/aws_cloudwatch_log_subscription_filter: Clean up error handling when creating log subscription filter ([#9065](https://github.com/terraform-providers/terraform-provider-aws/issues/9065))
* resource/aws_cognito_identity_provider: Properly pass all attributes during update ([#9396](https://github.com/terraform-providers/terraform-provider-aws/issues/9396))
* resource/aws_db_event_subscription: Handle `SubscriptionNotFound` errors during refresh and deletion ([#9371](https://github.com/terraform-providers/terraform-provider-aws/issues/9371))
* resource/aws_s3_account_public_access_block: Retry after timeout when reading s3 account public access block ([#9387](https://github.com/terraform-providers/terraform-provider-aws/issues/9387))
* resource/aws_ssm_maintenance_window_task: Bypass `DoesNotExistException` error on deletion ([#7823](https://github.com/terraform-providers/terraform-provider-aws/issues/7823))
* resource/aws_ssm_maintenance_window_task: Prevent `task_parameters` ordering differences ([#9364](https://github.com/terraform-providers/terraform-provider-aws/issues/9364))

## 2.19.0 (July 11, 2019)

FEATURES:

* **New Data Source:** `aws_msk_configuration` ([#9088](https://github.com/terraform-providers/terraform-provider-aws/issues/9088))
* **New Resource:** `aws_athena_workgroup` ([#9290](https://github.com/terraform-providers/terraform-provider-aws/issues/9290))
* **New Resource:** `aws_datapipeline_pipeline` ([#9267](https://github.com/terraform-providers/terraform-provider-aws/issues/9267))
* **New Resource:** `aws_directory_service_log_subscription` ([#9261](https://github.com/terraform-providers/terraform-provider-aws/issues/9261))

ENHANCEMENTS:

* resource/aws_acmpca_certificate_authority: Support validation for `ROOT` certificate authority type ([#9292](https://github.com/terraform-providers/terraform-provider-aws/issues/9292))
* resource/aws_appmesh_virtual_node: Add `aws_cloud_map` configuration block under `spec` and `service_discovery` ([#9271](https://github.com/terraform-providers/terraform-provider-aws/issues/9271))
* resource/aws_appmesh_virtual_router: Add `tags` argument ([#9249](https://github.com/terraform-providers/terraform-provider-aws/issues/9249))
* resource/aws_appmesh_virtual_service: Add `tags` argument ([#9252](https://github.com/terraform-providers/terraform-provider-aws/issues/9252))
* resource/aws_codebuild_project: Add `environment` configuration block `registry_credential` configuration block (support Secrets Manager registry credentials) ([#9168](https://github.com/terraform-providers/terraform-provider-aws/issues/9168))
* resource/aws_codebuild_project: Add `logs_config` configuration block (support CloudWatch and S3 logging configuration) ([#7534](https://github.com/terraform-providers/terraform-provider-aws/issues/7534))
* resource/aws_ebs_snapshot: Support customizable create/delete timeouts and increase defaults to 10 minutes ([#9157](https://github.com/terraform-providers/terraform-provider-aws/issues/9157))
* resource/aws_lightsail_instance: Add validation for `name` argument ([#8667](https://github.com/terraform-providers/terraform-provider-aws/issues/8667))
* resource/aws_lightsail_instance: Add `tags` argument ([#9273](https://github.com/terraform-providers/terraform-provider-aws/issues/9273))
* resource/aws_organizations_account: Add `tags` argument ([#9202](https://github.com/terraform-providers/terraform-provider-aws/issues/9202))
* resource/aws_service_discovery_service: Add `namespace_id` argument (Support HTTP namespaces) ([#7341](https://github.com/terraform-providers/terraform-provider-aws/issues/7341))
* resource/aws_ssm_document: Support resource import ([#9313](https://github.com/terraform-providers/terraform-provider-aws/issues/9313))
* resource/aws_waf_rule_group: Support resource import ([#9254](https://github.com/terraform-providers/terraform-provider-aws/issues/9254))
* resource/aws_wafregional_byte_match_set: Support resource import ([#9258](https://github.com/terraform-providers/terraform-provider-aws/issues/9258))
* resource/aws_wafregional_rule: Support resource import ([#9239](https://github.com/terraform-providers/terraform-provider-aws/issues/9239))
* resource/aws_wafregional_rule_group: Support resource import ([#9240](https://github.com/terraform-providers/terraform-provider-aws/issues/9240))
* resource/aws_wafregional_web_acl: Support resource import ([#9248](https://github.com/terraform-providers/terraform-provider-aws/issues/9248))

BUG FIXES:

* resource/aws_backup_selection: Retry creation for IAM eventual consistency error ([#9298](https://github.com/terraform-providers/terraform-provider-aws/issues/9298))
* resource/aws_db_event_subscription: Prevent `Unable to find RDS Event Subscription` error during deletion and refresh ([#9274](https://github.com/terraform-providers/terraform-provider-aws/issues/9274))
* resource/aws_iam_policy_attachment: Bypass `NoSuchEntity` error when detaching groups, roles, and users (support group, role (when `force_detach_policies` is enabled), and user renames (when `force_destroy` is enabled)) ([#9278](https://github.com/terraform-providers/terraform-provider-aws/issues/9278))
* resource/aws_s3_bucket: Properly handle the creation of tags defined in `lifecycle_rule` when no prefix argument is specified ([#7162](https://github.com/terraform-providers/terraform-provider-aws/issues/7162))
* resource/aws_ssm_document: Ensure `content` attribute is always refreshed ([#9313](https://github.com/terraform-providers/terraform-provider-aws/issues/9313))
* resource/aws_transfer_user: Final retry after timeout waiting for deletion of transfer user ([#9241](https://github.com/terraform-providers/terraform-provider-aws/issues/9241))
* service/organizations: Automatically retry API calls on `ConcurrentModificationException` error ([#9195](https://github.com/terraform-providers/terraform-provider-aws/issues/9195))

## 2.18.0 (July 05, 2019)

FEATURES:

* **New Data Source:** `aws_servicequotas_service` ([#9177](https://github.com/terraform-providers/terraform-provider-aws/issues/9177))
* **New Data Source:** `aws_servicequotas_service_quota` ([#9177](https://github.com/terraform-providers/terraform-provider-aws/issues/9177))

ENHANCEMENTS:

* resource/aws_appmesh_route: Add `tags` argument ([#9206](https://github.com/terraform-providers/terraform-provider-aws/issues/9206))
* resource/aws_appmesh_virtual_node: Add `tags` argument ([#9207](https://github.com/terraform-providers/terraform-provider-aws/issues/9207))
* resource/aws_codecommit_repository: Add `tags` argument ([#9215](https://github.com/terraform-providers/terraform-provider-aws/issues/9215))
* resource/aws_ec2_transit_gateway_route: Add `blackhole` argument ([#9224](https://github.com/terraform-providers/terraform-provider-aws/issues/9224))
* resource/aws_iam_group_policy: Support resource import ([#9217](https://github.com/terraform-providers/terraform-provider-aws/issues/9217))

BUG FIXES:

* resource/aws_db_instance: Properly include `allow_major_version_upgrade` value when creating an RDS instance from a replica  or snapshot to allow RDS to perform a major version upgrade if necessary ([#9178](https://github.com/terraform-providers/terraform-provider-aws/issues/9178))
* resource/aws_db_instance: Prevent `InvalidParameterCombination: No modifications were requested` error when updating only `allow_major_version_upgrade` argument ([#9193](https://github.com/terraform-providers/terraform-provider-aws/issues/9193))
* resource/aws_emr_cluster: Skip refreshing the `kerberos_attributes` configuration block `ad_domain_join_user` argument from the API as it does not contain the real configuration value ([#8559](https://github.com/terraform-providers/terraform-provider-aws/issues/8559))

## 2.17.0 (June 28, 2019)

FEATURES:

* **New Data Source:** `aws_ebs_default_kms_key` ([#8884](https://github.com/terraform-providers/terraform-provider-aws/issues/8884))
* **New Data Source:** `aws_ebs_encryption_by_default` ([#8884](https://github.com/terraform-providers/terraform-provider-aws/issues/8884))
* **New Resource:** `aws_appsync_function` ([#8502](https://github.com/terraform-providers/terraform-provider-aws/issues/8502))

ENHANCEMENTS:

* data-source/aws_acm_certificate: Add `key_types` argument (support searching for non-default key algorithm certificates such as RSA 4096 bit) ([#8553](https://github.com/terraform-providers/terraform-provider-aws/issues/8553))
* data-source/aws_ssm_parameter: Add `version` attribute ([#9127](https://github.com/terraform-providers/terraform-provider-aws/issues/9127))
* resource/aws_appmesh_mesh: Add `tags` argument ([#8111](https://github.com/terraform-providers/terraform-provider-aws/issues/8111))
* resource/aws_appsync_resolver: Add `kind` argument and `pipeline_config` configuration block ([#8502](https://github.com/terraform-providers/terraform-provider-aws/issues/8502))
* resource/aws_db_instance: Add `max_allocated_storage` argument (support Storage Autoscaling) ([#9087](https://github.com/terraform-providers/terraform-provider-aws/issues/9087))
* resource/aws_directory_service_directory: Tag on create (support tag limiting IAM policies) ([#7937](https://github.com/terraform-providers/terraform-provider-aws/issues/7937))
* resource/aws_dms_endpoint: Support `db2` in `engine_name` validation ([#9097](https://github.com/terraform-providers/terraform-provider-aws/issues/9097))
* resource/aws_kinesis_firehose_delivery_stream: Tag on create (support tag limiting IAM policies) ([#7981](https://github.com/terraform-providers/terraform-provider-aws/issues/7981))
* resource/aws_lb_listener: Support `TCP_UDP` and `UDP` in `protocol` validation ([#9111](https://github.com/terraform-providers/terraform-provider-aws/issues/9111))
* resource/aws_lb_target_group: Support `TCP_UDP` and `UDP` in `protocol` validation ([#9111](https://github.com/terraform-providers/terraform-provider-aws/issues/9111))
* resource/aws_route53_healthcheck: Add validation for `request_interval` argument ([#9158](https://github.com/terraform-providers/terraform-provider-aws/issues/9158))
* resource/aws_ssm_parameter: Add `version` attribute ([#9127](https://github.com/terraform-providers/terraform-provider-aws/issues/9127))

BUG FIXES:
* resource/aws_api_gateway_account: Fix error handling during update ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_api_gateway_base_path_mapping: Fix error handling during create ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_api_gateway_domain_name: Remove unnecessary retry during delete ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_api_gateway_gateway_response: Remove unnecessary retry during delete ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_api_gateway_model: Remove unnecessary retry during delete ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_api_gateway_usage_plan: Remove unnecessary retry during delete ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_api_gateway_usage_plan_key: Remove unnecessary retry during delete ([#9068](https://github.com/terraform-providers/terraform-provider-aws/issues/9068))
* resource/aws_db_snapshot: Prevent not found error when deleted outside Terraform ([#9099](https://github.com/terraform-providers/terraform-provider-aws/issues/9099))
* resource/aws_ebs_snapshot_copy: Prevent error when resource is deleted outside Terraform ([#9106](https://github.com/terraform-providers/terraform-provider-aws/issues/9106))
* resource/aws_ecr_repository: Final retries when reading and deleting ECR repositories ([#9079](https://github.com/terraform-providers/terraform-provider-aws/issues/9079))
* resource/aws_ecr_repository_policy: Final retries when creating and updating ECR repository policies ([#9079](https://github.com/terraform-providers/terraform-provider-aws/issues/9079))
* resource/aws_lb_target_group: Properly validate up to `120` seconds for `health_check` configuration block `timeout` argument ([#9152](https://github.com/terraform-providers/terraform-provider-aws/issues/9152))
* resource/aws_spot_fleet_request: Add final retry when creating spot fleet request ([#9078](https://github.com/terraform-providers/terraform-provider-aws/issues/9078))
* resource/aws_spot_instance_request: Add final retry when creating spot instance request ([#9078](https://github.com/terraform-providers/terraform-provider-aws/issues/9078))
* resource/aws_ssm_maintenance_window_target: Prevent `InvalidParameter` error on resource creation when optional `name` or `description` were missing ([#9165](https://github.com/terraform-providers/terraform-provider-aws/issues/9165))

## 2.16.0 (June 20, 2019)

FEATURES:

* **New Resource:** `aws_globalaccelerator_endpoint_group` ([#8328](https://github.com/terraform-providers/terraform-provider-aws/issues/8328))
* **New Resource:** `aws_ebs_default_kms_key` ([#8771](https://github.com/terraform-providers/terraform-provider-aws/issues/8771))
* **New Resource:** `aws_ebs_encryption_by_default` ([#8771](https://github.com/terraform-providers/terraform-provider-aws/issues/8771))
* **New Resource:** `aws_ses_identity_policy` ([#5128](https://github.com/terraform-providers/terraform-provider-aws/issues/5128))

ENHANCEMENTS:

* data-source/aws_vpc_endpoint: Add `owner_id` and `tags` attributes ([#8674](https://github.com/terraform-providers/terraform-provider-aws/issues/8674))
* data-source/aws_vpc_endpoint: Add `requester_managed` attribute ([#8396](https://github.com/terraform-providers/terraform-provider-aws/issues/8396))
* data-source/aws_vpc_endpoint_service: Add `manages_vpc_endpoints` attribute ([#8396](https://github.com/terraform-providers/terraform-provider-aws/issues/8396))
* data-source/aws_vpc_endpoint_service: Add `service_id` and `tags` attributes ([#8674](https://github.com/terraform-providers/terraform-provider-aws/issues/8674))
* provider: Support for chaining assume IAM role from AWS shared configuration files ([#8987](https://github.com/terraform-providers/terraform-provider-aws/issues/8987))
* resource/aws_backup_vault: Support resource import ([#9041](https://github.com/terraform-providers/terraform-provider-aws/issues/9041))
* resource/aws_codepipeline: Add `tags` argument ([#8993](https://github.com/terraform-providers/terraform-provider-aws/issues/8993))
* resource/aws_codepipeline_webhook: Add `tags` argument ([#8993](https://github.com/terraform-providers/terraform-provider-aws/issues/8993))
* resource/aws_ecs_task_definition: Add `proxy_configuration` configuration block (support AppMesh proxying) ([#8780](https://github.com/terraform-providers/terraform-provider-aws/issues/8780))
* resource/aws_instance: Prevent panic when `credit_specification` configuration block is missing arguments ([#9003](https://github.com/terraform-providers/terraform-provider-aws/issues/9003))
* resource/aws_organizations_organization: Add `non_master_accounts` attribute ([#8926](https://github.com/terraform-providers/terraform-provider-aws/issues/8926))
* resource/aws_secretsmanager_secret: Tag on create (support tag limiting IAM policies) ([#9023](https://github.com/terraform-providers/terraform-provider-aws/issues/9023))
* resource/aws_vpc_endpoint: Add `owner_id` attribute ([#8674](https://github.com/terraform-providers/terraform-provider-aws/issues/8674))
* resource/aws_vpc_endpoint: Add `requester_managed` attribute ([#8396](https://github.com/terraform-providers/terraform-provider-aws/issues/8396))
* resource/aws_vpc_endpoint: Add `tags` argument ([#8674](https://github.com/terraform-providers/terraform-provider-aws/issues/8674))
* resource/aws_vpc_endpoint_service: Add `manages_vpc_endpoints` attribute ([#8396](https://github.com/terraform-providers/terraform-provider-aws/issues/8396))
* resource/aws_vpc_endpoint_service: Add `tags` argument ([#8674](https://github.com/terraform-providers/terraform-provider-aws/issues/8674))

BUG FIXES:

* provider: Fix AWS shared configuration file credential source not assuming a role with environment and ECS credentials ([#8987](https://github.com/terraform-providers/terraform-provider-aws/issues/8987))
* provider: Properly configure Route 53 service client in AWS GovCloud (US) ([#9010](https://github.com/terraform-providers/terraform-provider-aws/issues/9010))
* provider: Properly configure Route 53 service client in AWS China ([#9060](https://github.com/terraform-providers/terraform-provider-aws/issues/9060))
* resource/aws_api_gateway_resource: Removes an extraneous retry when deleting API gateway resource ([#9054](https://github.com/terraform-providers/terraform-provider-aws/issues/9054))
* resource/aws_appautoscaling_policy: Retries after timeouts in creating and reading policies ([#9039](https://github.com/terraform-providers/terraform-provider-aws/issues/9039))
* resource/aws_appautoscaling_scheduled_action: Retry after timeout putting scheduled actions ([#9039](https://github.com/terraform-providers/terraform-provider-aws/issues/9039))
* resource/aws_cloudwatch_event_permission: Prevent not found error when deleted outside Terraform ([#9044](https://github.com/terraform-providers/terraform-provider-aws/issues/9044))
* resource/aws_dx_gateway: Fix resource import with associations ([#8970](https://github.com/terraform-providers/terraform-provider-aws/issues/8970))
* resource/aws_elasticache_parameter_group: Final retry deleting parameter group ([#9013](https://github.com/terraform-providers/terraform-provider-aws/issues/9013))
* resource/aws_elasticache_replication_group: Final retry deleting replication group ([#9013](https://github.com/terraform-providers/terraform-provider-aws/issues/9013))
* resource/aws_elasticache_subnet_group: Final retry deleting subnet group ([#9013](https://github.com/terraform-providers/terraform-provider-aws/issues/9013))
* resource/aws_emr_cluster: Final retry after timeout error when deleting EMR cluster ([#9053](https://github.com/terraform-providers/terraform-provider-aws/issues/9053))
* resource/aws_kinesis_firehose_delivery_stream: Add final retries when creating and updating firehose delivery streams ([#9017](https://github.com/terraform-providers/terraform-provider-aws/issues/9017))
* resource/aws_neptune_cluster: Final retries when creating, updating, and deleting Neptune clusters ([#9036](https://github.com/terraform-providers/terraform-provider-aws/issues/9036))
* resource/aws_neptune_cluster_instance: Final retries creating and updating cluster instances ([#9036](https://github.com/terraform-providers/terraform-provider-aws/issues/9036))
* resource/aws_neptune_parameter_group: Final retries updating and deleting parameter groups ([#9036](https://github.com/terraform-providers/terraform-provider-aws/issues/9036))
* resource/aws_opsworks_permission: Improves error handing when setting opsworks permissions ([#9055](https://github.com/terraform-providers/terraform-provider-aws/issues/9055))
* resource/aws_rds_cluster: Final retries after timeout creating and updating cluster ([#8994](https://github.com/terraform-providers/terraform-provider-aws/issues/8994))
* resource/aws_rds_cluster_instance: Final retry after timeout creating cluster instance ([#8994](https://github.com/terraform-providers/terraform-provider-aws/issues/8994))
* resource/aws_rds_global_cluster: Final retry after timeout deleting global cluster ([#8994](https://github.com/terraform-providers/terraform-provider-aws/issues/8994))
* resource/aws_ses_receipt_rule_set: Prevent missing Terraform state for newly created resources ([#9045](https://github.com/terraform-providers/terraform-provider-aws/issues/9045))
* resource/aws_ssm_document: Final retries when creating and deleting SSM documents ([#8992](https://github.com/terraform-providers/terraform-provider-aws/issues/8992))
* resource/aws_ssm_resource_data_sync: Final retry when creating SSM resource data sync ([#8992](https://github.com/terraform-providers/terraform-provider-aws/issues/8992))


## 2.15.0 (June 13, 2019)

FEATURES:

* **New Data Source:** `aws_customer_gateway` ([#8977](https://github.com/terraform-providers/terraform-provider-aws/issues/8977))

ENHANCEMENTS:

* resource/aws_cognito_user_pool: Add `email_sending_account` attribute to the email configuration block ([#8626](https://github.com/terraform-providers/terraform-provider-aws/issues/8626))
* resource/aws_redshift_cluster: Add `arn` attribute ([#8894](https://github.com/terraform-providers/terraform-provider-aws/issues/8894))
* resource/aws_redshift_event_subscription: Add `arn` attribute and support `tags` updates ([#8894](https://github.com/terraform-providers/terraform-provider-aws/issues/8894))
* resource/aws_redshift_parameter_group: Add `arn` attribute and `tags` argument ([#8894](https://github.com/terraform-providers/terraform-provider-aws/issues/8894))
* resource/aws_redshift_snapshot_copy_grant: Add `arn` attribute and support `tags` updates ([#8894](https://github.com/terraform-providers/terraform-provider-aws/issues/8894))
* resource/aws_redshift_subnet_group: Add `arn` attribute ([#8894](https://github.com/terraform-providers/terraform-provider-aws/issues/8894))
* resource/aws_ses_identity_notification_topic: Add `include_original_headers` argument ([#7293](https://github.com/terraform-providers/terraform-provider-aws/issues/7293))

BUG FIXES:

* resource/aws_api_gateway_resource: Final retry for deleting api gateway resource ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_appautoscaling_target: Final retry for registering autoscaling target ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_default_vpc_dhcp_options: Add pagination to get the default DHCP options correctly ([#8907](https://github.com/terraform-providers/terraform-provider-aws/issues/8907))
* resource/aws_docdb_cluster: Retries after timeout errors for docdb cluster operations ([#8986](https://github.com/terraform-providers/terraform-provider-aws/issues/8986))
* resource/aws_dx_connection_association: Final retry for deleting dx connection association ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_dynamodb_table_item: add a nil check when building the table item ID ([#8900](https://github.com/terraform-providers/terraform-provider-aws/issues/8900))
* resource/aws_eks_cluster: Increase default creation timeout to 30 minutes ([#8909](https://github.com/terraform-providers/terraform-provider-aws/issues/8909))
* resource/aws_elasticache_cluster: Final retry when deleting elasticache cluster ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_elasticache_replication_group: Implement passthrough state migration for upstream `missing MigrateState function` error in Terraform 0.12 ([#8887](https://github.com/terraform-providers/terraform-provider-aws/issues/8887))
* resource/aws_elasticache_security_group: Final retry for deleting elasticache security group ([#8981](https://github.com/terraform-providers/terraform-provider-aws/issues/8981))
* resource/aws_iam_server_certificate: Final retry for deleting IAM server cert ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_iam_service_linked_role: Automatically suppress Application Autoscaling `custom_suffix` differences ([#8931](https://github.com/terraform-providers/terraform-provider-aws/issues/8931))
* resource/aws_kinesis_analytics_application: Final retries for kinesis applications ([#8984](https://github.com/terraform-providers/terraform-provider-aws/issues/8984))
* resource/aws_sns_topic_subscription: Final retry for SNS topic subscription ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_ssm_activation: Final retry for creating SSM activation ([#8893](https://github.com/terraform-providers/terraform-provider-aws/issues/8893))
* resource/aws_vpc_dhcp_options: Add final retry to deleting DHCP options ([#8907](https://github.com/terraform-providers/terraform-provider-aws/issues/8907))

## 2.14.0 (June 06, 2019)

FEATURES:

* **New Data Source:** `aws_ec2_transit_gateway_dx_gateway_attachment` ([#8678](https://github.com/terraform-providers/terraform-provider-aws/issues/8678))

ENHANCEMENTS:

* data-source/aws_msk_cluster: Add `bootstrap_brokers_tls` attribute ([#8850](https://github.com/terraform-providers/terraform-provider-aws/issues/8850))
* resource/aws_codebuild_webhook: Add `filter_groups` configuration blocks ([#8110](https://github.com/terraform-providers/terraform-provider-aws/issues/8110))
* resource/aws_msk_cluster: Add `client_authentication`, `configuration_info`, and `encryption_in_transit` configuration blocks ([#8850](https://github.com/terraform-providers/terraform-provider-aws/issues/8850))
* resource/aws_msk_cluster: Add `bootstrap_brokers_tls` and `current_version` attributes ([#8850](https://github.com/terraform-providers/terraform-provider-aws/issues/8850))
* resource/aws_msk_cluster: Support `broker_node_group_into` configuration block `ebs_volume_size` argument updates ([#8850](https://github.com/terraform-providers/terraform-provider-aws/issues/8850))
* resource/aws_msk_cluster: Support tagging on creation ([#8850](https://github.com/terraform-providers/terraform-provider-aws/issues/8850))
* resource/aws_subnet: Use customizable timeouts for pending creation and waiting for `DependencyViolation` errors on deletion ([#6322](https://github.com/terraform-providers/terraform-provider-aws/issues/6322))

BUG FIXES:

* resource/aws_acmpca_certificate_authority: Add retry after timeout when creating CA ([#8856](https://github.com/terraform-providers/terraform-provider-aws/issues/8856))
* resource/aws_launch_template: Add a nil check for `spot_options` to avoiding panicking if options are empty ([#8844](https://github.com/terraform-providers/terraform-provider-aws/issues/8844))
* resource/aws_s3_bucket_metric: Add a nil check for `filter` to avoid panicking if empty ([#8852](https://github.com/terraform-providers/terraform-provider-aws/issues/8852))
* resource/aws_subnet: Bump default timeout for deletion from 10 to 20 minutes to better handle ELBv2 ENI deletions ([#6322](https://github.com/terraform-providers/terraform-provider-aws/issues/6322))

## 2.13.0 (May 31, 2019)

FEATURES:

* **New Resource:** `aws_ec2_transit_gateway_vpc_attachment_accepter` ([#8679](https://github.com/terraform-providers/terraform-provider-aws/issues/8679))

ENHANCEMENTS:

* resource/aws_iot_role_alias: Add `arn` attribute ([#8812](https://github.com/terraform-providers/terraform-provider-aws/issues/8812))

BUG FIXES:

* data-source/aws_rds_cluster: Add missing `hosted_zone_id` attribute ([#8799](https://github.com/terraform-providers/terraform-provider-aws/issues/8799))
* resource/aws_dms_endpoint: Fix casing for `mongodb_settings` configuration block `auth_type`, `auth_mechanism`, and `nesting_level` argument defaults ([#8008](https://github.com/terraform-providers/terraform-provider-aws/issues/8008)] / [[#8795](https://github.com/terraform-providers/terraform-provider-aws/issues/8795))

## 2.12.0 (May 24, 2019)

NOTES:

* resource/aws_dx_gateway_association: The `vpn_gateway_id` attribute is being deprecated in favor of the new `associated_gateway_id` attribute to support transit gateway associations ([#8528](https://github.com/terraform-providers/terraform-provider-aws/issues/8528))
* resource/aws_dx_gateway_association_proposal: The `vpn_gateway_id` attribute is being deprecated in favor of the new `associated_gateway_id` attribute to support transit gateway associations ([#8528](https://github.com/terraform-providers/terraform-provider-aws/issues/8528))

FEATURES:

* **New Data Source:** `aws_msk_cluster` ([#8743](https://github.com/terraform-providers/terraform-provider-aws/issues/8743))
* **New Resource:** `aws_msk_cluster` ([#8635](https://github.com/terraform-providers/terraform-provider-aws/issues/8635))
* **New Resource:** `aws_msk_configuration` ([#8740](https://github.com/terraform-providers/terraform-provider-aws/issues/8740))

ENHANCEMENTS:

* resource/aws_codebuild_project: Add `cache` configuration block `modes` argument and add `type` argument validation of `local` (support local cache) ([#8215](https://github.com/terraform-providers/terraform-provider-aws/issues/8215))
* resource/aws_db_instance: Add `performance_insights_enabled`, `performance_insights_kms_key_id`, and `performance_insights_retention_period` arguments ([#6453](https://github.com/terraform-providers/terraform-provider-aws/issues/6453))
* resource/aws_dx_gateway_association: New attributes `associated_gateway_owner_account_id` and `proposal_id` added to support the acceptance of a Direct Connect gateway association proposal and create a cross-account Direct Connect gateway association. New exported attributes `associated_gateway_type` and `dx_gateway_owner_account_id` ([#8528](https://github.com/terraform-providers/terraform-provider-aws/issues/8528))
* resource/aws_dx_gateway_association_proposal: New attribute `associated_gateway_id` replaces deprecated `vpn_gateway_id` attribute to support transit gateway associations. New exported attributes `associated_gateway_owner_account_id` and `associated_gateway_type` ([#8528](https://github.com/terraform-providers/terraform-provider-aws/issues/8528))
* resource/aws_ec2_transit_gateway: Handle deletion of transit gateways with DirectConnect Attachments ([#8752](https://github.com/terraform-providers/terraform-provider-aws/issues/8752))
* resource/aws_ec2_transit_gateway_route: Add retry to read method after timeout ([#8726](https://github.com/terraform-providers/terraform-provider-aws/issues/8726))
* resource/aws_kinesis_stream: Add `enforce_consumer_deletion` argument ([#8682](https://github.com/terraform-providers/terraform-provider-aws/issues/8682))
* resource/aws_ssm_maintenance_window_target: Add support for name and description for maintenance window targets ([#8671](https://github.com/terraform-providers/terraform-provider-aws/issues/8671))

BUG FIXES:

* resource/aws_iam_group: Prevent state removal during name attribute update ([#8707](https://github.com/terraform-providers/terraform-provider-aws/issues/8707))

## 2.11.0 (May 17, 2019)

NOTES:

* resource/aws_emr_cluster: The `instance_group` configuration block, `master_instance_type` argument, `core_instance_count` argument, and `core_instance_type` argument have been deprecated in favor of new `master_instance_group` and `core_instance_group` configuration blocks. The older, conflicting configurations were problematic in update scenarios. Task instance groups can be managed with the `aws_emr_instance_group` resource. Upgrade instructions can be found in the new Version 3 Upgrade Guide. ([#8459](https://github.com/terraform-providers/terraform-provider-aws/issues/8459))
* resource/aws_emr_instance_group: The addition of `autoscaling_policy` and `bid_price` arguments allow for the migration of task instance groups from the deprecated `instance_group` configuration block in the `aws_emr_cluster` resource. Import instructions for `aws_emr_instance_group` can be found in the resource documentation. ([#8078](https://github.com/terraform-providers/terraform-provider-aws/issues/8078))

FEATURES:

* **New Data Source:** `aws_ecr_image` ([#8403](https://github.com/terraform-providers/terraform-provider-aws/issues/8403))
* **New Data Source:** `aws_ram_resource_share` ([#8491](https://github.com/terraform-providers/terraform-provider-aws/issues/8491))
* **New Guide:** Version 3 Upgrade Guide ([#8459](https://github.com/terraform-providers/terraform-provider-aws/issues/8459))
* **New Resource:** `aws_ses_email_identity` ([#6575](https://github.com/terraform-providers/terraform-provider-aws/issues/6575))
* **New Resource:** `aws_shield_protection` ([#7721](https://github.com/terraform-providers/terraform-provider-aws/issues/7721))

ENHANCEMENTS:

* resource/aws_autoscaling_schedule: Support resource import ([#8300](https://github.com/terraform-providers/terraform-provider-aws/issues/8300))
* resource/aws_backup_selection: Support resource import ([#8546](https://github.com/terraform-providers/terraform-provider-aws/issues/8546))
* resource/aws_dynamodb_table: Support tagging on creation (where available) ([#8469](https://github.com/terraform-providers/terraform-provider-aws/issues/8469))
* resource/aws_elastic_beanstalk_application: Add `tags` argument and `arn` attribute ([#8614](https://github.com/terraform-providers/terraform-provider-aws/issues/8614))
* resource/aws_elastic_beanstalk_application_version: Add `tags` argument and `arn` attribute ([#8614](https://github.com/terraform-providers/terraform-provider-aws/issues/8614))
* resource/aws_emr_cluster: Add `master_instance_group` and `core_instance_group` configuration blocks (deprecates other instance group configuration methods) ([#8459](https://github.com/terraform-providers/terraform-provider-aws/issues/8459))
* resource/aws_emr_instance_group: Add support for `autoscaling_policy`, `bid_price`, and resource import ([#8078](https://github.com/terraform-providers/terraform-provider-aws/issues/8078))
* resource/aws_kinesis_analytics_application: Add `tags` argument ([#8643](https://github.com/terraform-providers/terraform-provider-aws/issues/8643))
* resource/aws_lambda_function: Support `nodejs10.x` in `runtime` validation ([#8622](https://github.com/terraform-providers/terraform-provider-aws/issues/8622))
* resource/aws_organizations_account: Add parent_id argument (support moving accounts) ([#8583](https://github.com/terraform-providers/terraform-provider-aws/issues/8583))
* resource/aws_sfn_activity: Support tagging on creation ([#8395](https://github.com/terraform-providers/terraform-provider-aws/issues/8395))
* resource/aws_sfn_state_machine: Support tagging on creation ([#8395](https://github.com/terraform-providers/terraform-provider-aws/issues/8395))
* resource/aws_sns_topic: Add `tags` argument ([#8468](https://github.com/terraform-providers/terraform-provider-aws/issues/8468))

BUG FIXES:

* resource/aws_backup_selection: Properly trigger resource recreation with `selection_tag` updates ([#8546](https://github.com/terraform-providers/terraform-provider-aws/issues/8546))
* resource/aws_cloudwatch_event_rule: Ignore `UnknownOperationException` error on reading tags (fixes ap-east-1 support) ([#8659](https://github.com/terraform-providers/terraform-provider-aws/issues/8659))
* resource/aws_ssm_parameter: Remove `Tier` from `PutParameter` if unsupported (fixes AWS China support) ([#8664](https://github.com/terraform-providers/terraform-provider-aws/issues/8664))
* resource/aws_vpn_gateway: Handle `attaching` and `detaching` attachment status ([#8576](https://github.com/terraform-providers/terraform-provider-aws/issues/8576))
* resource/aws_vpn_gateway_attachment: Handle `attaching` and `detaching` attachment status ([#8576](https://github.com/terraform-providers/terraform-provider-aws/issues/8576))

## 2.10.0 (May 10, 2019)

FEATURES:

* **New Data Source:** `aws_lambda_layer_version` ([#8577](https://github.com/terraform-providers/terraform-provider-aws/issues/8577))
* **New Resource:** `aws_organizations_organizational_unit` ([#4207](https://github.com/terraform-providers/terraform-provider-aws/issues/4207))
* **New Resource:** `aws_xray_sampling_rule` ([#8535](https://github.com/terraform-providers/terraform-provider-aws/issues/8535))

ENHANCEMENTS:

* data-source/aws_rds_cluster: Add `resource_id` attribute ([#8317](https://github.com/terraform-providers/terraform-provider-aws/issues/8317))
* resource/aws_appsync_graphql_api: Add `tags` argument ([#8567](https://github.com/terraform-providers/terraform-provider-aws/issues/8567))
* resource/aws_cloudfront_distribution: Validate `*_cache_behavior` `forwarded_values` `cookies` configuration block `forward` argument ([#8563](https://github.com/terraform-providers/terraform-provider-aws/issues/8563))
* resource/aws_glue_job: Add `pythonshell` job support by adding the `max_capacity` argument and deprecating the `allocated_capacity` argument ([#7340](https://github.com/terraform-providers/terraform-provider-aws/issues/7340))
* resource/aws_lambda_alias: Support resource import ([#8513](https://github.com/terraform-providers/terraform-provider-aws/issues/8513))
* resource/aws_organizations_organization: Add `roots` attribute ([#8399](https://github.com/terraform-providers/terraform-provider-aws/issues/8399))
* resource/aws_organizations_organization: Add `accounts` attribute ([#8581](https://github.com/terraform-providers/terraform-provider-aws/issues/8581))
* resource/aws_organizations_organization: Add `enabled_policy_types` argument ([#8588](https://github.com/terraform-providers/terraform-provider-aws/issues/8588))
* resource/aws_rds_cluster: Add `copy_tags_to_snapshot` argument ([#8544](https://github.com/terraform-providers/terraform-provider-aws/issues/8544))
* resource/aws_ssm_parameter: Add `tier` argument (support `Advanced` parameters) ([#8525](https://github.com/terraform-providers/terraform-provider-aws/issues/8525))

## 2.9.0 (May 06, 2019)

FEATURES:

* **New Resource:** aws_db_instance_role_association ([#8466](https://github.com/terraform-providers/terraform-provider-aws/issues/8466))

ENHANCEMENTS:

* data-source/aws_availability_zones: Add blacklisted_names and blacklisted_zone_ids arguments ([#8463](https://github.com/terraform-providers/terraform-provider-aws/issues/8463))
* resource/aws_elb_attachment: Retry ELB attachment on `InvalidTarget` error ([#8483](https://github.com/terraform-providers/terraform-provider-aws/issues/8483))
* resource/aws_sfn_state_machine: Bypass `UnknownOperationException` error for `ListTagsForResource` API call (additional LocalStack support) ([#8467](https://github.com/terraform-providers/terraform-provider-aws/issues/8467))
* resource/aws_ssm_activation: Add `tags` argument ([#8426](https://github.com/terraform-providers/terraform-provider-aws/issues/8426))
* resource/aws_ssm_document: Add `tags` argument ([#8426](https://github.com/terraform-providers/terraform-provider-aws/issues/8426))
* resource/aws_ssm_maintenance_window: Add `tags` argument ([#8426](https://github.com/terraform-providers/terraform-provider-aws/issues/8426))
* resource/aws_ssm_patch_baseline: Add `tags` argument ([#8426](https://github.com/terraform-providers/terraform-provider-aws/issues/8426))

## 2.8.0 (April 26, 2019)

NOTES:

* provider: Region validation now automatically supports the new `ap-east-1` Asia Pacific (Hong Kong) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [announcement blog post](https://aws.amazon.com/blogs/aws/now-open-aws-asia-pacific-hong-kong-region/). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g. `provider.aws: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g. `data.aws_availability_zones.current: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`).

FEATURES:

* **New Resource:** `aws_dx_gateway_association_proposal` ([#8320](https://github.com/terraform-providers/terraform-provider-aws/issues/8320))

ENHANCEMENTS:

* data-source/aws_cloudtrail_service_account: Support new `ap-east-1` region ([#8437](https://github.com/terraform-providers/terraform-provider-aws/issues/8437))
* data-source/aws_dx_gateway: Add `owner_account_id` attribute ([#8320](https://github.com/terraform-providers/terraform-provider-aws/issues/8320))
* data-source/aws_eks_cluster: Add `enabled_cluster_log_types` attribute ([#8402](https://github.com/terraform-providers/terraform-provider-aws/issues/8402))
* data-source/aws_elb_hosted_zone_id: Support new `ap-east-1` region ([#8437](https://github.com/terraform-providers/terraform-provider-aws/issues/8437))
* data-source/aws_elb_service_account: Support new `ap-east-1` region ([#8437](https://github.com/terraform-providers/terraform-provider-aws/issues/8437))
* data-source/aws_redshift_service_account: Support new `ap-east-1` region ([#8437](https://github.com/terraform-providers/terraform-provider-aws/issues/8437))
* data-source/aws_s3_bucket: Support new `ap-east-1` region in `hosted_zone_id` attribute ([#8437](https://github.com/terraform-providers/terraform-provider-aws/issues/8437))
* provider: Support automatic region validation for `ap-east-1` ([#8440](https://github.com/terraform-providers/terraform-provider-aws/issues/8440))
* resource/aws_dx_gateway: Add `owner_account_id` attribute ([#8320](https://github.com/terraform-providers/terraform-provider-aws/issues/8320))
* resource/aws_dx_gateway_association: Support resource import ([#8222](https://github.com/terraform-providers/terraform-provider-aws/issues/8222))
* resource/aws_dx_gateway_association: Add `allowed_prefixes` argument ([#8222](https://github.com/terraform-providers/terraform-provider-aws/issues/8222))
* resource/aws_lb: Support Network Load Balancer (NLB) access logs ([#8282](https://github.com/terraform-providers/terraform-provider-aws/issues/8282))
* resource/aws_s3_bucket: Support new `ap-east-1` region in `hosted_zone_id` attribute ([#8437](https://github.com/terraform-providers/terraform-provider-aws/issues/8437))
* resource/aws_transfer_user: Support `user_name` containing uppercase letters, hyphens, and underscores ([#8304](https://github.com/terraform-providers/terraform-provider-aws/issues/8304))

BUG FIXES:

* resource/aws_eks_cluster: Ignore ordering and properly handle removals with `enabled_cluster_log_types` argument ([#8402](https://github.com/terraform-providers/terraform-provider-aws/issues/8402))
* resource/aws_emr_cluster: Increase deletion timeout from 10 to 20 minutes to match AWS documentation ([#8428](https://github.com/terraform-providers/terraform-provider-aws/issues/8428))
* resource/aws_lb: Prevent difference when `subnet_mapping` configuration block `allocation_id` argument was omitted ([#8282](https://github.com/terraform-providers/terraform-provider-aws/issues/8282))
* resource/aws_lb: Properly disable access logs with `access_logs` configuration block `enabled` argument set to `false` ([#8282](https://github.com/terraform-providers/terraform-provider-aws/issues/8282))
* resource/aws_network_interface: Refresh private_ips_count into Terraform state and allow for updates from 0 to greater than 0 ([#8353](https://github.com/terraform-providers/terraform-provider-aws/issues/8353))
* resource/aws_vpc: Set ipv6_association_id and ipv6_cidr_block attributes as updated for assign_generated_ipv6_cidr_block updates ([#6721](https://github.com/terraform-providers/terraform-provider-aws/issues/6721))

## 2.7.0 (April 18, 2019)

NOTES:

* provider: This release includes only a Terraform SDK upgrade with compatibility for Terraform v0.12. The provider remains backwards compatible with Terraform v0.11 and this update should have no significant changes in behavior for the provider. Please report any unexpected behavior in new GitHub issues (Terraform core: https://github.com/hashicorp/terraform/issues or Terraform AWS Provider: https://github.com/terraform-providers/terraform-provider-aws/issues) ([#8366](https://github.com/terraform-providers/terraform-provider-aws/issues/8366))

## 2.6.0 (April 10, 2019)

NOTES:

* resource/aws_route53_record: Remove deprecation from `allow_overwrite` argument as there are some use cases where it is helpful. We discourage its usage in most environments (in preference of using `terraform import`) as it can easily cause conflicting management of the same Route53 Record. ([#8274](https://github.com/terraform-providers/terraform-provider-aws/issues/8274))

FEATURES:

* **New Resource:** `aws_worklink_website_certificate_authority_association` ([#7459](https://github.com/terraform-providers/terraform-provider-aws/issues/7459))

ENHANCEMENTS:

* resource/aws_appmesh_mesh: Add `spec` configuration block (support egress filter rules) ([#8119](https://github.com/terraform-providers/terraform-provider-aws/issues/8119))
* resource/aws_appmesh_route: Add `spec` configuration block `tcp_route` configuration block (support TCP routing) ([#8119](https://github.com/terraform-providers/terraform-provider-aws/issues/8119))
* resource/aws_appmesh_virtual_node: Add `spec` configuration block `logging` configuration block (support access logging) ([#8119](https://github.com/terraform-providers/terraform-provider-aws/issues/8119))
* resource/aws_eks_cluster: Add `enabled_cluster_log_types` argument (support EKS control plane logging) ([#8216](https://github.com/terraform-providers/terraform-provider-aws/issues/8216))
* resource/aws_iam_user_group_membership: Support resource import ([#6976](https://github.com/terraform-providers/terraform-provider-aws/issues/6976))
* resource/aws_launch_template: Add `elastic_inference_accelerator` configuration block ([#8247](https://github.com/terraform-providers/terraform-provider-aws/issues/8247))
* resource/aws_redshift_cluster: Add configurable timeouts ([#8241](https://github.com/terraform-providers/terraform-provider-aws/issues/8241))
* resource/aws_transfer_server: Add `endpoint_details` configuration block and `endpoint_type` argument (support Private Link) ([#8121](https://github.com/terraform-providers/terraform-provider-aws/issues/8121))
* resource/aws_wafregional_web_acl_association: Support additional `resource_arn` types (e.g. API Gateway) ([#7205](https://github.com/terraform-providers/terraform-provider-aws/issues/7205))

BUG FIXES:

* data-source/aws_lb_target_group: Add missing schema attributes ([#8213](https://github.com/terraform-providers/terraform-provider-aws/issues/8213))
* resource/aws_appautoscaling_policy: Retry creation on `ObjectNotFound` errors for eventual consistency ([#8273](https://github.com/terraform-providers/terraform-provider-aws/issues/8273))
* resource/aws_backup_plan: Prevent the sending of empty lifecycle attributes ([#8236](https://github.com/terraform-providers/terraform-provider-aws/issues/8236))
* resource/aws_glue_catalog_table: Properly trigger resource recreation when deleted outside Terraform ([#8174](https://github.com/terraform-providers/terraform-provider-aws/issues/8174))
* resource/aws_secretsmanager_secret: Handle additional scheduled for deletion error message on immediate secret recreation ([#8219](https://github.com/terraform-providers/terraform-provider-aws/issues/8219))
* provider: Prevent panic when setting `endpoints` configuration ([#8226](https://github.com/terraform-providers/terraform-provider-aws/issues/8226))

## 2.5.0 (April 05, 2019)

FEATURES:

* **New Guide:** [`Custom Service Endpoints`](https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html) ([#8092](https://github.com/terraform-providers/terraform-provider-aws/issues/8092))
* **New Resource:** `aws_backup_selection` ([#7382](https://github.com/terraform-providers/terraform-provider-aws/issues/7382))
* **New Resource:** `aws_sagemaker_endpoint` ([#2479](https://github.com/terraform-providers/terraform-provider-aws/issues/2479))
* **New Resource:** `aws_sagemaker_notebook_instance_lifecycle_configuration` ([#7585](https://github.com/terraform-providers/terraform-provider-aws/issues/7585))

ENHANCEMENTS:

* provider: Support customization of all service endpoints ([#8096](https://github.com/terraform-providers/terraform-provider-aws/issues/8096))
* resource/aws_acmpca_certificate_authority: Add `permanent_deletion_time_in_days` argument ([#7366](https://github.com/terraform-providers/terraform-provider-aws/issues/7366))
* resource/aws_budgets_budget: Add `notification` configuration block (support notifications) ([#4523](https://github.com/terraform-providers/terraform-provider-aws/issues/4523))
* resource/aws_cloudfront_distribution: Add `wait_for_deployment` argument ([#8116](https://github.com/terraform-providers/terraform-provider-aws/issues/8116))
* resource/aws_cloudwatch_log_subscription_filter: Support resource import ([#8147](https://github.com/terraform-providers/terraform-provider-aws/issues/8147))
* resource/aws_cloudwatch_metric_alarm: Add `tags` argument ([#8168](https://github.com/terraform-providers/terraform-provider-aws/issues/8168))
* resource/aws_ecr_repository: Tag on creation (support tag limiting IAM policies) ([#8198](https://github.com/terraform-providers/terraform-provider-aws/issues/8198))
* resource/aws_lb_target_group: Add `enabled` argument to target group health checks ([#7570](https://github.com/terraform-providers/terraform-provider-aws/issues/7570))
* resource/aws_sagemaker_notebook_instance: Add `lifecycle_config_name` argument ([#7586](https://github.com/terraform-providers/terraform-provider-aws/issues/7586))
* service/ec2: Automatically retry `CreateVpnConnection` and `CreateVpnGateway` requests for concurrency errors ([#8161](https://github.com/terraform-providers/terraform-provider-aws/issues/8161))

BUG FIXES:

* resource/aws_cloudfront_distribution: Ignore attribute ordering of cache behavior forwarded values `headers` and cookies `whitelisted_names` arguments ([#8150](https://github.com/terraform-providers/terraform-provider-aws/issues/8150))

## 2.4.0 (March 29, 2019)

NOTES:

* service/ec2: Due to an upcoming update to the EC2 service, both the `aws_instance` data source and resource will no longer make the EC2 API call `DescribeInstanceCreditSpecifications` unless they are in the T2 or T3 instance families. Previously, the EC2 service would allow this API call for instance families that did not support credit specifications, however the upcoming update will return an error and prevent Terraform runs from completing.

FEATURES:

* **New Data Source:** `aws_ec2_transit_gateway_vpn_attachment` ([#8071](https://github.com/terraform-providers/terraform-provider-aws/issues/8071))
* **New Resource:** `aws_appsync_resolver` ([#6451](https://github.com/terraform-providers/terraform-provider-aws/issues/6451))
* **New Resource:** `aws_kms_ciphertext` ([#6993](https://github.com/terraform-providers/terraform-provider-aws/issues/6993))
* **New Resource:** `aws_kms_external_key` ([#8066](https://github.com/terraform-providers/terraform-provider-aws/issues/8066))
* **New Resource:** `aws_sagemaker_endpoint_configuration` ([#2477](https://github.com/terraform-providers/terraform-provider-aws/issues/2477))

ENHANCEMENTS:

* data-source/aws_instance: Only call `DescribeInstanceCreditSpecifications` for T2 and T3 Instance Families ([#8107](https://github.com/terraform-providers/terraform-provider-aws/issues/8107))
* data-source/aws_kms_ciphertext: Hide `plaintext` in logs and user interface ([#6100](https://github.com/terraform-providers/terraform-provider-aws/issues/6100))
* resource/aws_appsync_graphql_api: Add `schema` argument ([#4840](https://github.com/terraform-providers/terraform-provider-aws/issues/4840))
* resource/aws_batch_compute_environment: Add `compute_resources` configuration block `launch_template` configuration block (support EC2 Launch Templates) ([#8026](https://github.com/terraform-providers/terraform-provider-aws/issues/8026))
* resource/aws_cloudwatch_event_rule: Add `tags` argument ([#8076](https://github.com/terraform-providers/terraform-provider-aws/issues/8076))
* resource/aws_instance: Only call `DescribeInstanceCreditSpecifications` for T2 and T3 Instance Families ([#8107](https://github.com/terraform-providers/terraform-provider-aws/issues/8107))
* resource/aws_ram_principal_association: Validate `principal` as AWS Account ID or ARN ([#8048](https://github.com/terraform-providers/terraform-provider-aws/issues/8048))
* resource/aws_s3_bucket: Support `DEEP_ARCHIVE` in storage class validations ([#8109](https://github.com/terraform-providers/terraform-provider-aws/issues/8109))
* resource/aws_s3_bucket_object: Support `DEEP_ARCHIVE` in `storage_class` validation ([#8109](https://github.com/terraform-providers/terraform-provider-aws/issues/8109))
* resource/aws_vpn_connection: Add `transit_gateway_attachment_id` attribute ([#8070](https://github.com/terraform-providers/terraform-provider-aws/issues/8070))

BUG FIXES:

* resource/aws_cloudwatch_metric_alarm: Prevent `ValidationError` when updating `metric_query` alarms ([#8085](https://github.com/terraform-providers/terraform-provider-aws/issues/8085))

## 2.3.0 (March 21, 2019)

BREAKING CHANGES:

* service/appmesh: Changes to support AppMesh General Availability (GA) release. The AppMesh resources were added while the API was under Public Preview and the GA release earlier this month introduced breaking changes which prevented further AWS Go SDK updates. This is a very atypical situation for the Terraform AWS Provider as most AWS API and SDK changes are additive. To prevent this situation in the future we may introduce separate Terraform AWS Provider(s) specifically for Public Preview or Beta APIs. If or when this occurs, it will be separately announced. The maintainers will continue following Terraform Provider compatibility promises outlined on the [HashiCorp Blog](https://www.hashicorp.com/blog/hashicorp-terraform-provider-versioning) and [Extending Terraform documentation](https://www.terraform.io/docs/extend/best-practices/versioning.html) as best as possible except other existing Public Preview resources. ([#7659](https://github.com/terraform-providers/terraform-provider-aws/issues/7659))
* resource/aws_appmesh_virtual_node: Replace `backends` configuration block(s) with `backend` configuration blocks ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* resource/aws_appmesh_virtual_node: Replace `service_discovery` configuration block `service_name` argument with `service_discovery` configuration block `dns` configuration block ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* resource/aws_appmesh_virtual_router: Remove `spec` configuration block `service_names` argument ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))

FEATURES:

* **New Data Source:** `aws_transfer_server` ([#7977](https://github.com/terraform-providers/terraform-provider-aws/issues/7977))
* **New Resource:** `aws_appmesh_virtual_service` ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* **New Resource:** `aws_backup_plan` ([#7350](https://github.com/terraform-providers/terraform-provider-aws/issues/7350))
* **New Resource:** `aws_cloudformation_stack_set` ([#8020](https://github.com/terraform-providers/terraform-provider-aws/issues/8020))
* **New Resource:** `aws_cloudformation_stack_set_instance` ([#8020](https://github.com/terraform-providers/terraform-provider-aws/issues/8020))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `vpc_config` configuration block `endpoint_private_access` and `endpoint_public_access` attributes ([#8024](https://github.com/terraform-providers/terraform-provider-aws/issues/8024))
* data-source/aws_instance: Add `get_user_data` argument and `user_data_base64` attribute ([#8001](https://github.com/terraform-providers/terraform-provider-aws/issues/8001))
* provider: Support custom endpoint for `ses` ([#7986](https://github.com/terraform-providers/terraform-provider-aws/issues/7986))
* provider: Support custom endpoints for `firehose` and `redshift` ([#8007](https://github.com/terraform-providers/terraform-provider-aws/issues/8007))
* resource/aws_api_gateway_deployment: Allow `stage_name` argument to be optional ([#6459](https://github.com/terraform-providers/terraform-provider-aws/issues/6459))
* resource/aws_appautoscaling_policy: Support resource import ([#8032](https://github.com/terraform-providers/terraform-provider-aws/issues/8032))
* resource/aws_appautoscaling_target: Support resource import ([#8032](https://github.com/terraform-providers/terraform-provider-aws/issues/8032))
* resource/aws_appmesh_route: Support resource import ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* resource/aws_appmesh_virtual_node: Support resource import ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* resource/aws_appmesh_virtual_router: Support resource import ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* resource/aws_appmesh_virtual_router: Add `spec` configuration block `listener` configuration block ([#7858](https://github.com/terraform-providers/terraform-provider-aws/issues/7858))
* resource/aws_cloudfront_distribution: Add `origin_group` configuration block (support Origin Groups and failover) ([#7202](https://github.com/terraform-providers/terraform-provider-aws/issues/7202))
* resource/aws_codebuild_project: Add `project_environment` configuration block `image_pull_credentials_type` argument (support cross-account images) ([#7458](https://github.com/terraform-providers/terraform-provider-aws/issues/7458))
* resource/aws_ecr_repository_policy: Support resource import ([#7974](https://github.com/terraform-providers/terraform-provider-aws/issues/7974))
* resource/aws_eks_cluster: Add `vpc_config` configuration block `endpoint_private_access` and `endpoint_public_access` arguments (support disabling public access) ([#8024](https://github.com/terraform-providers/terraform-provider-aws/issues/8024))
* resource/aws_iam_access_key: Support `status` updates (support disabling/enabling access keys) ([#7961](https://github.com/terraform-providers/terraform-provider-aws/issues/7961))
* resource/aws_kinesis_analytics_application: Support resource import ([#8027](https://github.com/terraform-providers/terraform-provider-aws/issues/8027))
* resource/aws_media_package_channel: Add `tags` argument ([#7984](https://github.com/terraform-providers/terraform-provider-aws/issues/7984))
* resource/aws_route53_zone_association: Support resource import ([#7966](https://github.com/terraform-providers/terraform-provider-aws/issues/7966))
* resource/aws_s3_bucket_inventory: Support plan-time validation of `optional_fields` values `ObjectLockRetainUntilDate`, `ObjectLockMode` and `ObjectLockLegalHoldStatus` ([#7952](https://github.com/terraform-providers/terraform-provider-aws/issues/7952))
* resource/aws_ssm_association: Add `compliance_severity` argument ([#7852](https://github.com/terraform-providers/terraform-provider-aws/issues/7852))
* resource/aws_ssm_association: Add `max_concurrency` and `max_errors` arguments ([#7970](https://github.com/terraform-providers/terraform-provider-aws/issues/7970))

BUG FIXES:

* resource/aws_appautoscaling_policy: Recreate resource for `resource_id` updates ([#7982](https://github.com/terraform-providers/terraform-provider-aws/issues/7982))
* resource/aws_appautoscaling_policy: Ignore `ObjectNotFoundException` on deletion ([#7982](https://github.com/terraform-providers/terraform-provider-aws/issues/7982))
* resource/aws_route53_zone_association: Properly trigger resource recreation on all updates ([#7966](https://github.com/terraform-providers/terraform-provider-aws/issues/7966))

## 2.2.0 (March 15, 2019)

FEATURES:

* **New Resource:** `aws_globalaccelerator_listener` ([#7003](https://github.com/terraform-providers/terraform-provider-aws/issues/7003))
* **New Resource:** `aws_guardduty_invite_accepter` ([#4610](https://github.com/terraform-providers/terraform-provider-aws/issues/4610))
* **New Resource:** `aws_route53_resolver_rule` ([#7799](https://github.com/terraform-providers/terraform-provider-aws/issues/7799))
* **New Resource:** `aws_route53_resolver_rule_association` ([#7799](https://github.com/terraform-providers/terraform-provider-aws/issues/7799))

ENHANCEMENTS:

* data-source/aws_eip: Add `private_dns` and `public_dns` attributes ([#7349](https://github.com/terraform-providers/terraform-provider-aws/issues/7349))
* resource/aws_backup_vault: Support `tags` updates ([#7933](https://github.com/terraform-providers/terraform-provider-aws/issues/7933))
* resource/aws_dx_bgp_peer: Add `aws_device` and `bgp_peer_id` attributes ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_dx_connection: Add `aws_device` and `has_logical_redundancy` attributes ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_dx_hosted_private_virtual_interface: Add `aws_device` attribute ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_dx_hosted_public_virtual_interface: Add `aws_device` attribute ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_dx_lag: Add `has_logical_redundancy` attribute ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_dx_private_virtual_interface: Add `aws_device` attribute ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_dx_public_virtual_interface: Add `aws_device` attribute ([#7131](https://github.com/terraform-providers/terraform-provider-aws/issues/7131))
* resource/aws_eip: Add `private_dns` and `public_dns` attributes ([#7349](https://github.com/terraform-providers/terraform-provider-aws/issues/7349))
* resource/aws_glue_crawler: Add `arn` attribute ([#7948](https://github.com/terraform-providers/terraform-provider-aws/issues/7948))
* resource/aws_ssm_patch_baseline: Support resource import ([#7838](https://github.com/terraform-providers/terraform-provider-aws/issues/7838))

BUG FIXES:

* resource/aws_cloudfront_distribution: Ensure retain_on_delete disables the CloudFront Distribution before exiting ([#7875](https://github.com/terraform-providers/terraform-provider-aws/issues/7875))
* resource/aws_cloudwatch_log_metric_filter: Serialize create, update, and delete operations on the same CloudWatch Log Group to prevent `OperationAbortedException` errors ([#7880](https://github.com/terraform-providers/terraform-provider-aws/issues/7880))
* resource/aws_codebuild_webhook: Only pass BranchFilter configuration if non-empty ([#7841](https://github.com/terraform-providers/terraform-provider-aws/issues/7841))
* resource/aws_ec2_transit_gateway_vpc_attachment: Prevent errors with Resource Access Manager shared EC2 Transit Gateways ([#7513](https://github.com/terraform-providers/terraform-provider-aws/issues/7513))
* resource/aws_ecr_repository_policy: Properly read `policy` into the Terraform state ([#7853](https://github.com/terraform-providers/terraform-provider-aws/issues/7853))
* resource/aws_iam_role_policy_attachment: Prevent `NoSuchEntity` errors from race conditions ([#7855](https://github.com/terraform-providers/terraform-provider-aws/issues/7855))
* resource/aws_kms_alias: Prevent state removal of resource immediately after creation due to eventual consistency ([#7891](https://github.com/terraform-providers/terraform-provider-aws/issues/7891))
* resource/aws_s3_bucket: Prevent empty `replication_configuration` `rules` `filter` crash ([#7887](https://github.com/terraform-providers/terraform-provider-aws/issues/7887))
* resource/aws_s3_bucket: Continue supporting empty string (`""`) `bucket` argument ([#7881](https://github.com/terraform-providers/terraform-provider-aws/issues/7881))
* resource/aws_s3_bucket: Prevent `NoSuchBucket` errors when putting lifecycle configuration on resource creation ([#7930](https://github.com/terraform-providers/terraform-provider-aws/issues/7930))
* resource/aws_ses_domain_mail_from: Prevent crash with deleted SES Domain Identity ([#7883](https://github.com/terraform-providers/terraform-provider-aws/issues/7883))

## 2.1.0 (March 07, 2019)

FEATURES:

* **New Resource:** `aws_route53_resolver_endpoint` ([#6563](https://github.com/terraform-providers/terraform-provider-aws/issues/6563))

ENHANCEMENTS:

* data-source/aws_elastic_beanstalk_hosted_zone: Add eu-north-1 region support ([#7829](https://github.com/terraform-providers/terraform-provider-aws/issues/7829))
* data-source/aws_redshift_service_account: Add us-gov-east-1 and us-gov-west-1 region mappings ([#7635](https://github.com/terraform-providers/terraform-provider-aws/issues/7635))
* data-source/aws_s3_bucket: Add `bucket_regional_domain_name` attribute ([#7765](https://github.com/terraform-providers/terraform-provider-aws/issues/7765))
* resource/aws_autoscaling_group: Support new `mixed_instances_policy` `instance_distribution` `spot_max_price` ability to unset with empty string ([#7821](https://github.com/terraform-providers/terraform-provider-aws/issues/7821))
* resource/aws_dlm_lifecycle_policy: Add validation support for 2, 3, 4, 6, and 8 in `policy_details` `schedule` `create_rule` `interval` argument (support shorter intervals) ([#7751](https://github.com/terraform-providers/terraform-provider-aws/issues/7751))
* resource/aws_ec2_client_vpn_endpoint: Add `tags` argument ([#7619](https://github.com/terraform-providers/terraform-provider-aws/issues/7619))
* resource/aws_ecs_service: Support plan time validation of new `health_check_grace_period_seconds` max of `2147483647` ([#7806](https://github.com/terraform-providers/terraform-provider-aws/issues/7806))
* resource/aws_lb_target_group: Add `lambda_multi_value_headers_enabled` argument ([#7648](https://github.com/terraform-providers/terraform-provider-aws/issues/7648))
* resource/aws_ram_resource_share: Add `arn` attribute ([#7634](https://github.com/terraform-providers/terraform-provider-aws/issues/7634))
* resource/aws_s3_bucket: Add plan time length validation for `bucket` and `bucket_prefix` arguments ([#7778](https://github.com/terraform-providers/terraform-provider-aws/issues/7778))

BUG FIXES:

* resource/aws_autoscaling_group: Allow configuration of `mixed_instances_policy` `instance_distribution` `on_demand_base_capacity` argument to 0 ([#7821](https://github.com/terraform-providers/terraform-provider-aws/issues/7821))
* resource/aws_cloudfront_distribution: Remove problematic `viewer_certificate` configuration block argument `ConflictsWith` usage from version 2.0.0 ([#7794](https://github.com/terraform-providers/terraform-provider-aws/issues/7794))
* resource/aws_cloudfront_distribution: Skip disabling distributions on deletion for previously disabled distributions ([#7794](https://github.com/terraform-providers/terraform-provider-aws/issues/7794))
* resource/aws_cloudfront_distribution: Retry on `PreconditionFailed` error messages after disabling distribution on deletion ([#7794](https://github.com/terraform-providers/terraform-provider-aws/issues/7794))
* resource/aws_cloudfront_distribution: Wait for creation and update deployments to complete ([#7794](https://github.com/terraform-providers/terraform-provider-aws/issues/7794))
* resource/aws_cloudfront_distribution: Prevent one minute timeout error for creation and update errors due to throttling ([#7809](https://github.com/terraform-providers/terraform-provider-aws/issues/7809))
* resource/aws_db_instance: Properly set `engine_version` with `snapshot_identifier` ([#7738](https://github.com/terraform-providers/terraform-provider-aws/issues/7738))
* resource/aws_dynamodb_table: Prevent perpetual plan differences with `ttl` configuration block `enabled` argument set to `false` ([#3960](https://github.com/terraform-providers/terraform-provider-aws/issues/3960))
* resource/aws_ecs_service: Ensure `placement_strategy` removal in version 2.0.0 does not force recreation ([#7790](https://github.com/terraform-providers/terraform-provider-aws/issues/7790))
* resource/aws_guardduty_detector: Prevent GuardDuty member accounts with unconfigured `finding_publishing_frequency` from triggering update errors ([#7804](https://github.com/terraform-providers/terraform-provider-aws/issues/7804))
* resource/aws_launch_configuration: Prevent `ResourceInUse` errors caused by eventual consistency during deletion ([#7819](https://github.com/terraform-providers/terraform-provider-aws/issues/7819))
* resource/aws_s3_bucket_notification: Prevent crash with empty filters configuration ([#7791](https://github.com/terraform-providers/terraform-provider-aws/issues/7791))

## 2.0.0 (February 27, 2019)

NOTES:

* Full documentation about this update, including Terraform provider version pinning and configuration examples, can be found in the [Terraform AWS Provider Version 2 Upgrade Guide](https://www.terraform.io/docs/providers/aws/guides/version-2-upgrade.html)

BREAKING CHANGES:

* provider: Return error on AWS Account ID lookup failure during initialization (unless `skip_requesting_account_id = true`) ([#7737](https://github.com/terraform-providers/terraform-provider-aws/issues/7737))
* data-source/aws_ami: Require `owners` argument ([#5576](https://github.com/terraform-providers/terraform-provider-aws/issues/5576))
* data-source/aws_ami_ids: Require `owners` argument ([#5576](https://github.com/terraform-providers/terraform-provider-aws/issues/5576))
* data-source/aws_iam_role: Remove deprecated attributes ([#7696](https://github.com/terraform-providers/terraform-provider-aws/issues/7696))
* data-source/aws_kms_secret: Remove data source (replaced with `aws_kms_secrets` data source) ([#7657](https://github.com/terraform-providers/terraform-provider-aws/issues/7657))
* data-source/aws_lambda_function: Returns unqualified (no `:QUALIFIER` or `:VERSION` suffix) value in `arn` attribute by default and qualified (includes `:QUALIFIER` or `:VERSION` suffix) value in `qualified_arn` attribute. Previously the `arn` attribute included `:$LATEST` suffix by default which was not compatible with many other resources. To restore the previous default behavior, set the `qualifier` argument to `$LATEST` and reference the `qualified_arn` attribute. ([#7663](https://github.com/terraform-providers/terraform-provider-aws/issues/7663))
* data-source/aws_region: Remove deprecated `current` argument ([#7697](https://github.com/terraform-providers/terraform-provider-aws/issues/7697))
* resource/aws_api_gateway_api_key: Remove deprecated `stage_key` configuration block ([#7698](https://github.com/terraform-providers/terraform-provider-aws/issues/7698))
* resource/aws_api_gateway_integration: Remove deprecated `request_parameters_in_json` argument (replaced with `request_parameters` argument) ([#7699](https://github.com/terraform-providers/terraform-provider-aws/issues/7699))
* resource/aws_api-gateway_integration_response: Remove deprecated `response_parameters_in_json` argument (replaced with `response_parameters` argument) ([#7700](https://github.com/terraform-providers/terraform-provider-aws/issues/7700))
* resource/aws_api_gateway_method: Remove deprecated `request_parameters_in_json` argument (replaced with `request_parameters` argument) ([#7701](https://github.com/terraform-providers/terraform-provider-aws/issues/7701))
* resource/aws_api_gateway_method_response: Remove deprecated `response_parameters_in_json` argument (replaced with `response_parameters` argument) ([#7704](https://github.com/terraform-providers/terraform-provider-aws/issues/7704))
* resource/aws_appautoscaling_policy: Remove deprecated arguments (replaced with `step_scaling_policy_configuration` configuration block) ([#7706](https://github.com/terraform-providers/terraform-provider-aws/issues/7706))
* resource/aws_autoscaling_policy: Remove deprecated `min_adjustment_step` argument (replaced with `min_adjustment_magnitude` argument) ([#7707](https://github.com/terraform-providers/terraform-provider-aws/issues/7707))
* resource/aws_batch_compute_environment: Remove deprecated `ecc_cluster_arn` attribute (replaced with `ecs_cluster_arn` attribute) ([#7708](https://github.com/terraform-providers/terraform-provider-aws/issues/7708))
* resource/aws_cloudfront_distribution: Remove deprecated `cache_behaviors` configuration block (replaced with `ordered_cache_behaviors` configuration block) ([#7710](https://github.com/terraform-providers/terraform-provider-aws/issues/7710))
* resource/aws_cognito_user_pool: Ensure only `email_verification_message` argument or `verification_message_template` configuration block `email_message` argument is configured ([#2425](https://github.com/terraform-providers/terraform-provider-aws/issues/2425))
* resource/aws_cognito_user_pool: Ensure only `email_verification_subject` argument or `verification_message_template` configuration block `email_subject` argument is configured ([#2425](https://github.com/terraform-providers/terraform-provider-aws/issues/2425))
* resource/aws_cognito_user_pool: Ensure only `sms_verification_message` argument or `verification_message_template` configuration block `sms_message` argument is defined ([#2425](https://github.com/terraform-providers/terraform-provider-aws/issues/2425))
* resource/aws_dx_lag: Remove deprecated `number_of_connections` argument and delete unmanaged connection during resource creation ([#7711](https://github.com/terraform-providers/terraform-provider-aws/issues/7711))
* resource/aws_ecs_service: Remove deprecated `placement_strategy` configuration block (replaced with `ordered_placement_strategy` configuration block) ([#7712](https://github.com/terraform-providers/terraform-provider-aws/issues/7712))
* resource/aws_efs_file_system: Remove deprecated `reference_name` argument (replaced with `creation_token` argument) ([#7713](https://github.com/terraform-providers/terraform-provider-aws/issues/7713))
* resource/aws_elasticache_cluster: Remove deprecated `availability_zones` argument (replaced with `preferred_availability_zones` argument) ([#7714](https://github.com/terraform-providers/terraform-provider-aws/issues/7714))
* resource/aws_instance: Remove deprecated top-level `network_interface_id` attribute ([#1193](https://github.com/terraform-providers/terraform-provider-aws/issues/1193)] / [[#7715](https://github.com/terraform-providers/terraform-provider-aws/issues/7715))
* resource/aws_instance: Remove hardcoded AWS China prevention of tagging on creation ([#7654](https://github.com/terraform-providers/terraform-provider-aws/issues/7654))
* resource/aws_lambda_function: Setting `reserved_concurrent_executions` to `0` will now disable Lambda Function invocations, causing downtime for the Lambda Function. Previously `reserved_concurrent_executions` accepted `0` and below for unreserved concurrency, which means it was not previously possible to disable invocations. The argument now differentiates between a new value for unreserved concurrency (`-1`) and disabling Lambda invocations (`0`). If previously configuring this value to `0` for unreserved concurrency, update the configured value to `-1` or the resource will disable Lambda Function invocations on update. If previously unconfigured, the argument does not require any changes. See the [Lambda User Guide](https://docs.aws.amazon.com/lambda/latest/dg/concurrent-executions.html) for more information about concurrency.
* resource/aws_lambda_layer_version: Swap `arn` and `layer_arn` attribute values ([#7664](https://github.com/terraform-providers/terraform-provider-aws/issues/7664))
* resource/aws_network_acl: Remove deprecated `subnet_id` argument (replaced with `subnet_ids` argument) ([#7716](https://github.com/terraform-providers/terraform-provider-aws/issues/7716))
* resource/aws_redshift_cluster: Remove deprecated `bucket_name`, `enable_logging`, and `s3_key_prefix` arguments (replaced with `logging` configuration block) ([#7717](https://github.com/terraform-providers/terraform-provider-aws/issues/7717))
* resource/aws_route_table: Resource import no longer imports `aws_route`, `aws_route_table_association`, and `aws_main_route_table_association` resources into the Terraform state ([#5657](https://github.com/terraform-providers/terraform-provider-aws/issues/5657))
* resource/aws_route53_record: Require import for existing records (use deprecated `allow_overwrite` argument to temporarily reinstate the old behavior) ([#7734](https://github.com/terraform-providers/terraform-provider-aws/issues/7734))
* resource/aws_route53_zone: Remove deprecated `vpc_id` and `vpc_region` arguments (replaced with `vpc` configuration block) ([#7695](https://github.com/terraform-providers/terraform-provider-aws/issues/7695))
* resource/aws_s3_bucket_object: Remove hardcoded AWS China prevention of tagging ([#7654](https://github.com/terraform-providers/terraform-provider-aws/issues/7654))
* resource/aws_wafregional_byte_match_set: Remove deprecated `byte_match_tuple` configuration block (replaced with `byte_match_tuples` configuration block) ([#7718](https://github.com/terraform-providers/terraform-provider-aws/issues/7718))

ENHANCEMENTS:

* data-source/aws_lambda_function: Add `tags` attribute ([#7663](https://github.com/terraform-providers/terraform-provider-aws/issues/7663))
* resource/aws_dx_lag: Delete unmanaged connection during resource creation ([#7711](https://github.com/terraform-providers/terraform-provider-aws/issues/7711))
* resource/aws_lambda_function: Disable Lambda Function invocations by setting `reserved_concurrent_executions` to `0` ([#3806](https://github.com/terraform-providers/terraform-provider-aws/issues/3806))

BUG FIXES:

* data-source/aws_lambda_function: Properly return error for missing function ([#7663](https://github.com/terraform-providers/terraform-provider-aws/issues/7663))
* resource/aws_appautoscaling_policy: Properly read `step_scaling_policy_configuration` into Terraform state ([#7706](https://github.com/terraform-providers/terraform-provider-aws/issues/7706))
* resource/aws_cloudfront_distribution: Adjust TypeSet and TypeList attributes for better difference handling ([#7732](https://github.com/terraform-providers/terraform-provider-aws/issues/7732))
* resource/aws_redshift_cluster: Properly read logging into Terraform state ([#7717](https://github.com/terraform-providers/terraform-provider-aws/issues/7717))
* resource/aws_route53_record: Existing Route 53 Records are no longer silently overwritten during resource creation by default (use `terraform import` or deprecated `allow_overwrite` argument) ([#7734](https://github.com/terraform-providers/terraform-provider-aws/issues/7734))
* resource/aws_vpn_connection: Remove configurability of read-only `customer_gateway_configuration`, `routes`, and `vgw_telemetry` attributes ([#7636](https://github.com/terraform-providers/terraform-provider-aws/issues/7636))

## 1.60.0 (February 22, 2019)

ENHANCEMENTS:

* resource/aws_route53_record: Add validation for alias `name` and `zone_id` arguments ([#7606](https://github.com/terraform-providers/terraform-provider-aws/issues/7606))

BUG FIXES:

* resource/aws_db_instance: Prevent snapshot restore error with `allocated_storage` and io1 `storage_type` ([#5800](https://github.com/terraform-providers/terraform-provider-aws/issues/5800)] / [[#7426](https://github.com/terraform-providers/terraform-provider-aws/issues/7426))
* resource/aws_flow_log: Prevent crash on unsuccessful flow log creation ([#7528](https://github.com/terraform-providers/terraform-provider-aws/issues/7528))
* resource/aws_kinesis_analytics_application: Retry input Kinesis Firehose Delivery Stream `InvalidArgumentException` permission errors for IAM eventual consistency during creation ([#7578](https://github.com/terraform-providers/terraform-provider-aws/issues/7578))
* resource/aws_launch_template: Properly read and write `description` ([#7569](https://github.com/terraform-providers/terraform-provider-aws/issues/7569))
* resource/aws_rds_cluster: Remove requirement for `master_password` and `master_username` when creating via `global_cluster_identifier` ([#7213](https://github.com/terraform-providers/terraform-provider-aws/issues/7213))
* resource/aws_vpc_endpoint: Suppress equivalent `policy` differences ([#7645](https://github.com/terraform-providers/terraform-provider-aws/issues/7645))

## 1.59.0 (February 14, 2019)

FEATURES:

* **New Data Source:** `aws_cur_report_definition` ([#7432](https://github.com/terraform-providers/terraform-provider-aws/issues/7432))
* **New Resource:** `aws_cur_report_definition` ([#7432](https://github.com/terraform-providers/terraform-provider-aws/issues/7432))
* **New Resource:** `aws_docdb_cluster_instance` ([#7143](https://github.com/terraform-providers/terraform-provider-aws/issues/7143))
* **New Resource:** `aws_ram_principal_association` ([#7219](https://github.com/terraform-providers/terraform-provider-aws/issues/7219)] / [[#7563](https://github.com/terraform-providers/terraform-provider-aws/issues/7563))
* **New Resource:** `aws_ram_resource_association` ([#7449](https://github.com/terraform-providers/terraform-provider-aws/issues/7449))

ENHANCEMENTS:

* data-source/aws_dynamodb_table: Add `billing_mode` and `point_in_time_recovery` attributes ([#7497](https://github.com/terraform-providers/terraform-provider-aws/issues/7497))
* resource/aws_appmesh_virtual_node: Add support for listener health checks ([#7446](https://github.com/terraform-providers/terraform-provider-aws/issues/7446))
* resource/aws_cloudwatch_metric_alarm: Add `metric_query` argument (support math expressions) ([#6833](https://github.com/terraform-providers/terraform-provider-aws/issues/6833))
* resource/aws_cognito_user_pool: Add `user_pool_add_ons` argument (support advanced security mode) ([#7361](https://github.com/terraform-providers/terraform-provider-aws/issues/7361))
* resource/aws_directory_service_directory: Set `security_group_id` attribute when type is `ADConnector` ([#7487](https://github.com/terraform-providers/terraform-provider-aws/issues/7487))
* resource/aws_dynamodb_table: Use `update` customizable timeout for individual Global Secondary Index updates and increase default `update` timeout from 10 minutes to 60 minutes ([#7453](https://github.com/terraform-providers/terraform-provider-aws/issues/7453))
* resource/aws_dms_endpoint: Add `docdb` for `engine_name` validation ([#7491](https://github.com/terraform-providers/terraform-provider-aws/issues/7491))
* resource/aws_ses_identity_notification_topic: Support resource import ([#7343](https://github.com/terraform-providers/terraform-provider-aws/issues/7343))
* resource/aws_waf_web_acl: Add `arn` attribute and `logging_configuration` argument ([#6059](https://github.com/terraform-providers/terraform-provider-aws/issues/6059))
* resource/aws_wafregional_web_acl: Add `arn` attribute and `logging_configuration` argument ([#7480](https://github.com/terraform-providers/terraform-provider-aws/issues/7480))

BUG FIXES:

* resource/aws_api_gateway_rest_api: Prevent timeout errors with large amounts of concurrent resource deletions ([#7554](https://github.com/terraform-providers/terraform-provider-aws/issues/7554))
* resource/aws_dynamodb_table: Prevent error when updating `billing_mode` from `PAY_PER_REQUEST` to `PROVISIONED` for Tables with Global Secondary Indexes ([#7453](https://github.com/terraform-providers/terraform-provider-aws/issues/7453))
* resource/aws_ec2_transit_gateway_vpc_attachment: Allow `pendingAcceptance` as available state during resource creation (support shared Transit Gateways with manual acceptance) ([#7489](https://github.com/terraform-providers/terraform-provider-aws/issues/7489))
* resource/aws_iam_user_login_profile: Properly return all errors during resource creation ([#7519](https://github.com/terraform-providers/terraform-provider-aws/issues/7519))
* resource/aws_iot_topic_rule: Allow optional `range_key_field` and `range_key_value` arguments ([#7471](https://github.com/terraform-providers/terraform-provider-aws/issues/7471))
* resource/aws_kinesis_analytics_application: Correctly manage multiple outputs ([#7535](https://github.com/terraform-providers/terraform-provider-aws/issues/7535))
* resource/aws_ssm_maintenance_window_task: Prevent erroneous `name` and `description` validation errors on resource creation ([#7186](https://github.com/terraform-providers/terraform-provider-aws/issues/7186))
* resource/aws_ssm_resource_data_sync: Properly trigger resource recreation for argument updates under the `s3_destination` configuration block ([#7490](https://github.com/terraform-providers/terraform-provider-aws/issues/7490))

## 1.58.0 (February 08, 2019)

FEATURES:

* **New Data Source:** `aws_eks_cluster_auth` ([#7438](https://github.com/terraform-providers/terraform-provider-aws/issues/7438))
* **New Resource:** `aws_backup_vault` ([#7207](https://github.com/terraform-providers/terraform-provider-aws/issues/7207))
* **New Resource:** `aws_docdb_cluster` ([#7122](https://github.com/terraform-providers/terraform-provider-aws/issues/7122))
* **New Resource:** `aws_docdb_cluster_snapshot` ([#7123](https://github.com/terraform-providers/terraform-provider-aws/issues/7123))
* **New Resource:** `aws_ec2_client_vpn_endpoint` ([#7009](https://github.com/terraform-providers/terraform-provider-aws/issues/7009))
* **New Resource:** `aws_ec2_client_vpn_network_association` ([#7030](https://github.com/terraform-providers/terraform-provider-aws/issues/7030))
* **New Resource:** `aws_sagemaker_model` ([#2478](https://github.com/terraform-providers/terraform-provider-aws/issues/2478))

ENHANCEMENTS:

* resource/aws_s3_bucket: Limit no of `expiration` & `noncurrent_version_expiration` blocks to 1 ([#7462](https://github.com/terraform-providers/terraform-provider-aws/issues/7462))

BUG FIXES:

* resource/aws_dynamodb_table: Prevent error when updating `billing_mode` from `PAY_PER_REQUEST` to `PROVISIONED` ([#7363](https://github.com/terraform-providers/terraform-provider-aws/issues/7363))

## 1.57.0 (January 26, 2019)

NOTES:

* provider: Fairly light release this week due to Terraform 0.12 upgrade and Terraform AWS Provider 2.0 related activities happening behind the scenes, which will continue through at least next week. For additional tracking information about those efforts, see the v2.0.0 milestone in GitHub.

ENHANCEMENTS

* data-source/aws_mq_broker: Add `tags` attribute ([#7193](https://github.com/terraform-providers/terraform-provider-aws/issues/7193))
* provider: Switch codebase dependency management from `govendor` to Go modules ([#7165](https://github.com/terraform-providers/terraform-provider-aws/issues/7165))
* resource/aws_autoscaling_policy: Support resource import ([#7195](https://github.com/terraform-providers/terraform-provider-aws/issues/7195))
* resource/aws_lb_listener: Add `TLS` to `protocol` argument validation ([#7338](https://github.com/terraform-providers/terraform-provider-aws/issues/7338))
* resource/aws_lb_target_group: Add `TLS` to `protocol` argument validation ([#7338](https://github.com/terraform-providers/terraform-provider-aws/issues/7338))
* resource/aws_mq_broker: Add `tags` argument ([#7193](https://github.com/terraform-providers/terraform-provider-aws/issues/7193))
* resource/aws_mq_configuration: Add `tags` argument ([#7193](https://github.com/terraform-providers/terraform-provider-aws/issues/7193))

BUG FIXES

* resource/aws_autoscaling_policy: Properly read `step_adjustment` into Terraform state ([#7336](https://github.com/terraform-providers/terraform-provider-aws/issues/7336))
* resource/aws_emr_cluster: Fix regression with `instance_group` differences when using `name` ([#7324](https://github.com/terraform-providers/terraform-provider-aws/issues/7324))
* resource/aws_iot_topic_rule: Prevent panic with missing SQS UseBase64 attribute in API response ([#7337](https://github.com/terraform-providers/terraform-provider-aws/issues/7337))
* resource/aws_lambda_permission: Retry for Lambda function eventual consistency on creation ([#7327](https://github.com/terraform-providers/terraform-provider-aws/issues/7327))
* resource/aws_rds_cluster_parameter_group: Prevent missing DBClusterParameterGroupName error on creation with generated names and `name_prefix` ([#7326](https://github.com/terraform-providers/terraform-provider-aws/issues/7326))
* resource/aws_s3_bucket_object: Delete S3 objects with leading '/' in the key name ([#7268](https://github.com/terraform-providers/terraform-provider-aws/issues/7268))

## 1.56.0 (January 16, 2019)

NOTES

* resource/aws_db_option_group: The Terraform resource is now able to perform drift detection on and import the state of the `option` attribute. The caveat is that while RDS returns only modified options, it will return all option settings whether modified or not from their default values. To workaround this option settings issue, we pass in the options from the Terraform configuration and ignore default values that are not present in the configuration. Some Terraform configurations may require minor updates to match the expected values.

FEATURES

* **New Data Source:** `aws_elastic_beanstalk_application` ([#7144](https://github.com/terraform-providers/terraform-provider-aws/issues/7144))
* **New Resource:** `aws_docdb_subnet_group` ([#7106](https://github.com/terraform-providers/terraform-provider-aws/issues/7106))
* **New Resource:** `aws_globalaccelerator_accelerator` ([#7002](https://github.com/terraform-providers/terraform-provider-aws/issues/7002))
* **New Resource:** `aws_lambda_layer_version` ([#6782](https://github.com/terraform-providers/terraform-provider-aws/issues/6782))
* **New Resource:** `aws_ram_resource_share` ([#6528](https://github.com/terraform-providers/terraform-provider-aws/issues/6528))
* **New Resource:** `aws_sagemaker_notebook_instance` ([#7139](https://github.com/terraform-providers/terraform-provider-aws/issues/7139))

ENHANCEMENTS

* data-source/aws_lambda_function: Add `layers` attribute ([#7126](https://github.com/terraform-providers/terraform-provider-aws/issues/7126))
* resource/aws_emr_cluster: Support resource import ([#4488](https://github.com/terraform-providers/terraform-provider-aws/issues/4488)] / [[#6498](https://github.com/terraform-providers/terraform-provider-aws/issues/6498))
* resource/aws_emr_cluster: Support `instance_group` `autoscaling_policy` updates ([#6498](https://github.com/terraform-providers/terraform-provider-aws/issues/6498))
* resource/aws_inspector_assessment_target: Allow omitting resource_group_arn argument (support matching all EC2 instances) ([#7112](https://github.com/terraform-providers/terraform-provider-aws/issues/7112))
* resource/aws_inspector_assessment_target: Support resource import ([#7112](https://github.com/terraform-providers/terraform-provider-aws/issues/7112))
* resource/aws_lambda_function: Add `layers` argument ([#7126](https://github.com/terraform-providers/terraform-provider-aws/issues/7126))
* resource/aws_s3_bucket: Add `object_lock_configuration` argument (support S3 Object Lock) ([#6964](https://github.com/terraform-providers/terraform-provider-aws/issues/6964))

BUG FIXES

* resource/aws_acm_certificate: Prevent crash with empty `SubjectAlternativeNames` (e.g. `IMPORTED` type certificates with IP address `CommonName`) ([#7127](https://github.com/terraform-providers/terraform-provider-aws/issues/7127))
* resource/aws_api_gateway_method_settings: Prevent crash when using `cache_data_encrypted` ([#7133](https://github.com/terraform-providers/terraform-provider-aws/issues/7133))
* resource/aws_db_option_group: Read `option` attribute into Terraform state for drift detection ([#7125](https://github.com/terraform-providers/terraform-provider-aws/issues/7125))
* resource/aws_db_option_group: Skip erroneous `ModifyOptionGroup` error when no option updates are being performed ([#7125](https://github.com/terraform-providers/terraform-provider-aws/issues/7125))
* resource/aws_ec2_transit_gateway_route: Prevent crash with externally removed attachment ([#7117](https://github.com/terraform-providers/terraform-provider-aws/issues/7117))
* resource/aws_emr_cluster: Properly read `core_instance_count`, `master_instance_type`, and `termination_policies` into Terraform state ([#4488](https://github.com/terraform-providers/terraform-provider-aws/issues/4488)] / [[#6498](https://github.com/terraform-providers/terraform-provider-aws/issues/6498))
* resource/aws_inspector_assessment_target: Properly read resource_group_arn attribute into Terraform state ([#7112](https://github.com/terraform-providers/terraform-provider-aws/issues/7112))
* resource/aws_launch_template: Prevent crashes with empty configuration blocks for top-level attributes ([#7134](https://github.com/terraform-providers/terraform-provider-aws/issues/7134))
* resource/aws_s3_bucket_object: Prevent creating new S3 object version when updating ACL or tags ([#7138](https://github.com/terraform-providers/terraform-provider-aws/issues/7138))
* resource/aws_s3_bucket_object: Prevent `NoSuchKey` errors reading tags on resource creation ([#7138](https://github.com/terraform-providers/terraform-provider-aws/issues/7138))
* service/applicationautoscaling: Use proper endpoint information in AWS GovCloud (US) ([#7142](https://github.com/terraform-providers/terraform-provider-aws/issues/7142))
* service/servicediscovery: Return full error messaging for failed operations ([#7118](https://github.com/terraform-providers/terraform-provider-aws/issues/7118))

## 1.55.0 (January 10, 2019)

FEATURES

* **New Resource:** `aws_docdb_cluster_parameter_group` ([#7090](https://github.com/terraform-providers/terraform-provider-aws/issues/7090))
* **New Resource:** `aws_media_package_channel` ([#6957](https://github.com/terraform-providers/terraform-provider-aws/issues/6957))
* **New Resource:** `aws_resourcegroups_group` ([#6217](https://github.com/terraform-providers/terraform-provider-aws/issues/6217))

ENHANCEMENTS

* resource/aws_app_cookie_stickiness_policy: Support resource import ([#7080](https://github.com/terraform-providers/terraform-provider-aws/issues/7080))
* resource/aws_elasticsearch_domain: Support in-place updates of `elasticsearch_version` ([#6243](https://github.com/terraform-providers/terraform-provider-aws/issues/6243))
* resource/aws_kinesis_firehose_delivery_stream: Add `extended_s3_configuration` `error_output_prefix` argument ([#7026](https://github.com/terraform-providers/terraform-provider-aws/issues/7026))
* resource/aws_sfn_activity: Add `tags` argument ([#7024](https://github.com/terraform-providers/terraform-provider-aws/issues/7024))
* resource/aws_sfn_state_machine: Add `tags` argument ([#7024](https://github.com/terraform-providers/terraform-provider-aws/issues/7024))
* resource/aws_ssm_maintenance_window: Add `end_date`, `schedule_timezone`, and `start_date` arguments ([#7040](https://github.com/terraform-providers/terraform-provider-aws/issues/7040))

BUG FIXES

* resource/aws_batch_job_queue: Properly read `compute_environments` into Terraform state ([#7079](https://github.com/terraform-providers/terraform-provider-aws/issues/7079))
* resource/aws_dynamodb_table: Prevent `BillingMode` `ValidationError` on table creation ([#7064](https://github.com/terraform-providers/terraform-provider-aws/issues/7064))
* resource/aws_iam_group_policy: Skip `NoSuchEntity` errors on resource deletion without refresh ([#7071](https://github.com/terraform-providers/terraform-provider-aws/issues/7071))
* resource/aws_iam_role_policy: Skip `NoSuchEntity` errors on resource deletion without refresh ([#7070](https://github.com/terraform-providers/terraform-provider-aws/issues/7070))
* resource/aws_iam_policy: Present more human readable error message during resource deletion ([#7072](https://github.com/terraform-providers/terraform-provider-aws/issues/7072))
* resource/aws_iam_user_policy: Skip `NoSuchEntity` errors on resource deletion without refresh ([#7069](https://github.com/terraform-providers/terraform-provider-aws/issues/7069))
* resource/aws_instance: Skip `InvalidInstanceID.NotFound` error on resource deletion ([#6978](https://github.com/terraform-providers/terraform-provider-aws/issues/6978))
* resource/aws_kinesis_analytics_application: Retry Lambda permission `InvalidArgumentException` errors for IAM eventual consistency ([#7039](https://github.com/terraform-providers/terraform-provider-aws/issues/7039))

## 1.54.0 (December 21, 2018)

NOTES

* This will be the last release until early January. Enjoy the rest of your year!

FEATURES

* **New Data Source:** `aws_autoscaling_group` ([#6849](https://github.com/terraform-providers/terraform-provider-aws/issues/6849))
* **New Resource:** `aws_licensemanager_association` ([#6926](https://github.com/terraform-providers/terraform-provider-aws/issues/6926))
* **New Resource:** `aws_s3_bucket_public_access_block` ([#6607](https://github.com/terraform-providers/terraform-provider-aws/issues/6607))
* **New Resource:** `aws_securityhub_product_subscription` ([#6921](https://github.com/terraform-providers/terraform-provider-aws/issues/6921))

ENHANCEMENTS

* resource/aws_acm_certificate: Add `certificate_body`, `certificate_chain`, and `private_key` arguments (Support importing/uploading certificate into ACM) ([#5453](https://github.com/terraform-providers/terraform-provider-aws/issues/5453))
* resource/aws_launch_template: Add `license_specification` argument ([#6926](https://github.com/terraform-providers/terraform-provider-aws/issues/6926))
* resource/aws_redshift_cluster: Support in-place updates for adding or removing KMS encryption ([#6865](https://github.com/terraform-providers/terraform-provider-aws/issues/6865))
* resource/aws_transfer_server: Add `force_destroy` argument ([#6935](https://github.com/terraform-providers/terraform-provider-aws/issues/6935))

BUG FIXES

* resource/aws_acm_certificate: Prevent error using Terraform resource import with certificates missing domain validation options ([#5472](https://github.com/terraform-providers/terraform-provider-aws/issues/5472))

## 1.53.0 (December 20, 2018)

FEATURES

* **New Resource:** `aws_licensemanager_license_configuration` ([#6835](https://github.com/terraform-providers/terraform-provider-aws/issues/6835))
* **New Resource:** `aws_rds_global_cluster` ([#6861](https://github.com/terraform-providers/terraform-provider-aws/issues/6861))
* **New Resource:** `aws_s3_account_public_access_block` ([#6851](https://github.com/terraform-providers/terraform-provider-aws/issues/6851))
* **New Resource:** `aws_securityhub_standards_subscription` ([#6862](https://github.com/terraform-providers/terraform-provider-aws/issues/6862))
* **New Resource:** `aws_service_discovery_http_namespace` ([#6864](https://github.com/terraform-providers/terraform-provider-aws/issues/6864))
* **New Resource:** `aws_transfer_ssh_key` ([#6932](https://github.com/terraform-providers/terraform-provider-aws/issues/6932))
* **New Resource:** `aws_transfer_user` ([#6850](https://github.com/terraform-providers/terraform-provider-aws/issues/6850))

ENHANCEMENTS

* data-source/aws_cloudtrail_service_account: Support `us-gov-east-1` and `us-gov-west-1` regions ([#6893](https://github.com/terraform-providers/terraform-provider-aws/issues/6893))
* data-source/aws_ecr_repository: Add `tags` attribute ([#6911](https://github.com/terraform-providers/terraform-provider-aws/issues/6911))
* resource/aws_codebuild_project: Support `source` `report_build_status` for GitHub Enterprise ([#6929](https://github.com/terraform-providers/terraform-provider-aws/issues/6929))
* resource/aws_db_snapshot: Add `tags` argument ([#6881](https://github.com/terraform-providers/terraform-provider-aws/issues/6881))
* resource/aws_ecr_repository: Add `tags` argument ([#6911](https://github.com/terraform-providers/terraform-provider-aws/issues/6911))
* resource/aws_ecs_service: Add `platform_version` argument (ECS Fargate Platform Version Support) ([#6510](https://github.com/terraform-providers/terraform-provider-aws/issues/6510))
* resource/aws_guardduty_detector: Add `finding_publishing_frequency` argument ([#6922](https://github.com/terraform-providers/terraform-provider-aws/issues/6922))
* resource/aws_lb_target_group: Add `lambda` as supported `target_type` with omitting `port`, `protocol`, and `vpc_id` arguments (support Lambda target groups) ([#6719](https://github.com/terraform-providers/terraform-provider-aws/issues/6719))
* resource/aws_rds_cluster: Allow `global` in `engine_mode` validation and add `global_cluster_identifier` argument ([#6861](https://github.com/terraform-providers/terraform-provider-aws/issues/6861))

BUG FIXES

* resource/aws_cloudwatch_log_stream: Trigger resource recreation on `ResourceNotFoundException` error ([#6776](https://github.com/terraform-providers/terraform-provider-aws/issues/6776))
* resource/aws_s3_bucket: Skip `MethodNotAllowed` error during read for missing S3 Bucket Acceleration functionality (`eu-north-1` region support) ([#6873](https://github.com/terraform-providers/terraform-provider-aws/issues/6873))
* resource/aws_transfer_server: Prevent error when no `tags` are assigned on resource creation ([#6883](https://github.com/terraform-providers/terraform-provider-aws/issues/6883))

## 1.52.0 (December 13, 2018)

FEATURES

* **New Data Source:** `aws_api_gateway_vpc_link` ([#6763](https://github.com/terraform-providers/terraform-provider-aws/issues/6763))
* **New Resource:** `aws_appmesh_route` ([#6766](https://github.com/terraform-providers/terraform-provider-aws/issues/6766))
* **New Resource:** `aws_appmesh_virtual_node` ([#6764](https://github.com/terraform-providers/terraform-provider-aws/issues/6764))
* **New Resource:** `aws_rds_cluster_endpoint` ([#6576](https://github.com/terraform-providers/terraform-provider-aws/issues/6576))
* **New Resource:** `aws_securityhub_account` ([#6839](https://github.com/terraform-providers/terraform-provider-aws/issues/6839))
* **New Resource:** `aws_transfer_server` ([#6639](https://github.com/terraform-providers/terraform-provider-aws/issues/6639))

ENHANCEMENTS

* data-source/aws_cloudtrail_service_account: Support `ap-northeast-3` and `eu-north-1` regions ([#6836](https://github.com/terraform-providers/terraform-provider-aws/issues/6836))
* data-source/aws_elb_hosted_zone_id: Support `ap-northeast-3` and `eu-north-1` regions ([#6836](https://github.com/terraform-providers/terraform-provider-aws/issues/6836))
* data-source/aws_elb_service_account: Support `ap-northeast-3`, `eu-north-1`, and `us-gov-east-1` regions ([#6836](https://github.com/terraform-providers/terraform-provider-aws/issues/6836))
* data-source/aws_instance: Add `host_id` attribute ([#6767](https://github.com/terraform-providers/terraform-provider-aws/issues/6767))
* data-source/aws_ip_ranges: Add `url` argument ([#6756](https://github.com/terraform-providers/terraform-provider-aws/issues/6756))
* data-source/aws_s3_bucket: Support `ap-northeast-3` and `eu-north-1` regions for `hosted_zone_id` ([#6836](https://github.com/terraform-providers/terraform-provider-aws/issues/6836))
* provider: Support automatic region validation for `eu-north-1` ([#6815](https://github.com/terraform-providers/terraform-provider-aws/issues/6815))
* resource/aws_acm_certificate: Automatically trim trailing period from `domain_name` and `subject_alternative_names` arguments ([#6844](https://github.com/terraform-providers/terraform-provider-aws/issues/6844))
* resource/aws_db_instance: Allow `postgresql` and `upgrade` values for `enabled_cloudwatch_logs_exports` (e.g. Postgres specific log exports) ([#6829](https://github.com/terraform-providers/terraform-provider-aws/issues/6829))
* resource/aws_dynamodb_table: Allow `global_secondary_index` configuration block `read_capacity` and `write_capacity` to be omitted with `billing_mode` set to `PAY_PER_REQUEST` (support on-demand billing with GSIs) ([#6737](https://github.com/terraform-providers/terraform-provider-aws/issues/6737))
* resource/aws_eks_cluster: Support `version` update ([#6843](https://github.com/terraform-providers/terraform-provider-aws/issues/6843))
* resource/aws_glue_crawler: Add `security_configuration` argument ([#6797](https://github.com/terraform-providers/terraform-provider-aws/issues/6797))
* resource/aws_iam_user_ssh_key: Support resource import ([#6727](https://github.com/terraform-providers/terraform-provider-aws/issues/6727))
* resource/aws_instance: Add `host_id` argument ([#6767](https://github.com/terraform-providers/terraform-provider-aws/issues/6767))
* resource/aws_s3_bucket: Support `ap-northeast-3` and `eu-north-1` regions for `hosted_zone_id` ([#6836](https://github.com/terraform-providers/terraform-provider-aws/issues/6836))
* resource/aws_ssm_maintenance_window: Support resource import ([#6747](https://github.com/terraform-providers/terraform-provider-aws/issues/6747))

BUG FIXES

* data-source/aws_lb_listener: Add missing `default_action` attributes from resource ([#6830](https://github.com/terraform-providers/terraform-provider-aws/issues/6830))
* resource/aws_cloudwatch_log_subscription_filter: Ignore `ResourceNotFound` error on deletion ([#6760](https://github.com/terraform-providers/terraform-provider-aws/issues/6760))
* resource/aws_ec2_transit_gateway_route: Trigger resource recreation with deleted/deleting route state ([#6817](https://github.com/terraform-providers/terraform-provider-aws/issues/6817))
* resource/aws_elasticache_parameter_group: Handle API reset issues with `reserved-memory` parameter updates ([#6752](https://github.com/terraform-providers/terraform-provider-aws/issues/6752))
* resource/aws_lambda_permission: Ignore `ResourceNotFoundException` error on deletion ([#6770](https://github.com/terraform-providers/terraform-provider-aws/issues/6770))
* resource/aws_lb_listener: Properly return an error when there are issues setting `default_action` attributes in Terraform state ([#6830](https://github.com/terraform-providers/terraform-provider-aws/issues/6830))
* resource/aws_route53_record: Prevent scanning entire zone for missing record ([#6753](https://github.com/terraform-providers/terraform-provider-aws/issues/6753))
* resource/aws_ssm_document: Properly batch large `permissions` updates for API limits ([#6735](https://github.com/terraform-providers/terraform-provider-aws/issues/6735))

## 1.51.0 (December 05, 2018)

FEATURES

* **New Resource:** `aws_appmesh_mesh` ([#6708](https://github.com/terraform-providers/terraform-provider-aws/issues/6708))
* **New Resource:** `aws_appmesh_virtual_router` ([#6720](https://github.com/terraform-providers/terraform-provider-aws/issues/6720))

ENHANCEMENTS

* data-source/aws_availability_zone: Add `zone_id` attribute ([#6686](https://github.com/terraform-providers/terraform-provider-aws/issues/6686))
* data-source/aws_availability_zones: Add `zone_ids` attribute ([#6686](https://github.com/terraform-providers/terraform-provider-aws/issues/6686))
* data-source/aws_iam_policy_document: Provide error if duplicate `sid` are configured across statements ([#6675](https://github.com/terraform-providers/terraform-provider-aws/issues/6675))
* data-source/aws_iam_policy_document: Add `version` argument ([#6699](https://github.com/terraform-providers/terraform-provider-aws/issues/6699))
* data-source/aws_internet_gateway: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* data-source/aws_route_table: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* data-source/aws_subnet: Add `availability_zone_id` argument and `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* data-source/aws_subnet: Use API provided `arn` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* data-source/aws_vpc: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* data-source/aws_vpc_dhcp_options: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_cloudtrail: Add `is_organization_trail` argument ([#6580](https://github.com/terraform-providers/terraform-provider-aws/issues/6580))
* resource/aws_codedeploy_config: Add `compute_platform` and `traffic_routing_config` arguments (support Lambda) ([#6644](https://github.com/terraform-providers/terraform-provider-aws/issues/6644))
* resource/aws_default_network_acl: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_default_route_table: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_default_subnet: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_default_vpc: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_default_vpc_dhcp_options: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_dynamodb_table: Add `billing_mode` argument (support on-demand capacity) ([#6648](https://github.com/terraform-providers/terraform-provider-aws/issues/6648))
* resource/aws_ecs_service: Add `propagate_tags` argument ([#6603](https://github.com/terraform-providers/terraform-provider-aws/issues/6603))
* resource/aws_ecs_task_definition: Support resource import ([#6723](https://github.com/terraform-providers/terraform-provider-aws/issues/6723))
* resource/aws_iam_group_policy_attachment: Support resource import ([#6625](https://github.com/terraform-providers/terraform-provider-aws/issues/6625))
* resource/aws_iam_user_policy_attachment: Support resource import ([#6487](https://github.com/terraform-providers/terraform-provider-aws/issues/6487))
* resource/aws_internet_gateway: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_lambda_alias: Add `invoke_arn` attribute ([#6329](https://github.com/terraform-providers/terraform-provider-aws/issues/6329))
* resource/aws_lambda_function: Support `provided` in `runtime` validation ([#6676](https://github.com/terraform-providers/terraform-provider-aws/issues/6676))
* resource/aws_lambda_function: Support `python3.7` in `runtime` validation ([#6583](https://github.com/terraform-providers/terraform-provider-aws/issues/6583))
* resource/aws_lambda_function: Support `ruby2.5` in `runtime` validation ([#6657](https://github.com/terraform-providers/terraform-provider-aws/issues/6657))
* resource/aws_network_acl: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_rds_cluster_instance: Add `copy_tags_to_snapshot` argument ([#6582](https://github.com/terraform-providers/terraform-provider-aws/issues/6582))
* resource/aws_route_table: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_s3_bucket: Support `INTELLIGENT_TIERING` in storage class validations ([#6589](https://github.com/terraform-providers/terraform-provider-aws/issues/6589))
* resource/aws_s3_bucket: Support replication rule destination storage class `GLACIER` ([#6613](https://github.com/terraform-providers/terraform-provider-aws/issues/6613))
* resource/aws_s3_bucket_inventory: Support destination bucket `Parquet` in `format` validation ([#6729](https://github.com/terraform-providers/terraform-provider-aws/issues/6729))
* resource/aws_s3_bucket_object: Support `GLACIER` in `storage_class` validation ([#6610](https://github.com/terraform-providers/terraform-provider-aws/issues/6610))
* resource/aws_s3_bucket_object: Support `INTELLIGENT_TIERING` in `storage_class` validation ([#6589](https://github.com/terraform-providers/terraform-provider-aws/issues/6589))
* resource/aws_ses_event_destination: Support multiple `cloudwatch_destination` configuration blocks ([#6690](https://github.com/terraform-providers/terraform-provider-aws/issues/6690))
* resource/aws_subnet: Add `availability_zone_id` argument and `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_subnet: Use API provided `arn` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_vpc: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))
* resource/aws_vpc_dhcp_options: Add `owner_id` attribute ([#6642](https://github.com/terraform-providers/terraform-provider-aws/issues/6642))

BUG FIXES:

* resource/aws_db_instance: Allow `configuring-iam-database-auth` as pending state ([#6597](https://github.com/terraform-providers/terraform-provider-aws/issues/6597))
* resource/aws_ec2_transit_gateway_vpc_attachment: Prevent error when Transit Gateway does not have default route table ([#6665](https://github.com/terraform-providers/terraform-provider-aws/issues/6665))
* resource/aws_iam_user_ssh_key: Properly trigger resource recreation with `encoding` and `public_key` updates ([#6718](https://github.com/terraform-providers/terraform-provider-aws/issues/6718))
* resource/aws_iot_topic_rule: Omit sending empty string `cloudwatch_metric` configuration block `metric_timestamp` argument to AWS ([#6618](https://github.com/terraform-providers/terraform-provider-aws/issues/6618))

## 1.50.0 (November 29, 2018)

ENHANCEMENTS

* resource/aws_codedeploy_app: Support `ECS` `compute_platform` ([#6647](https://github.com/terraform-providers/terraform-provider-aws/issues/6647))
* resource/aws_codedeploy_deployment_group: Add `ecs_service` argument and `load_balancer_info` configuration block `target_group_pair_info` argument (Support ECS Blue/Green Deployment) ([#6647](https://github.com/terraform-providers/terraform-provider-aws/issues/6647))
* resource/aws_ecs_service: Add `deployment_controller` argument ([#6647](https://github.com/terraform-providers/terraform-provider-aws/issues/6647))

## 1.49.0 (November 27, 2018)

FEATURES

* **New Data Source:** `aws_ec2_transit_gateway` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Data Source:** `aws_ec2_transit_gateway_route_table` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Data Source:** `aws_ec2_transit_gateway_vpc_attachment` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Resource:** `aws_ec2_transit_gateway` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Resource:** `aws_ec2_transit_gateway_route` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Resource:** `aws_ec2_transit_gateway_route_table` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Resource:** `aws_ec2_transit_gateway_route_table_association` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Resource:** `aws_ec2_transit_gateway_route_table_propagation` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* **New Resource:** `aws_ec2_transit_gateway_vpc_attachment` ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))

ENHANCEMENTS

* data-source/aws_route: Add `transit_gateway_id` attribute ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* data-source/aws_route_table: Add `route` attribute block `transit_gateway_id` attribute ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* resource/aws_default_route_table: Add `route` configuration block `transit_gateway_id` argument ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* resource/aws_route: Add `transit_gateway_id` argument ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* resource/aws_route_table: Add `route` configuration block `transit_gateway_id` argument ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))
* resource/aws_vpn_connection: Add `transit_gateway_id` argument, mark `vpn_gateway_id` as optional ([#6605](https://github.com/terraform-providers/terraform-provider-aws/issues/6605))

## 1.48.0 (November 26, 2018)

FEATURES

* **New Resource:** `aws_datasync_agent` ([#6591](https://github.com/terraform-providers/terraform-provider-aws/issues/6591))
* **New Resource:** `aws_datasync_location_efs` ([#6591](https://github.com/terraform-providers/terraform-provider-aws/issues/6591))
* **New Resource:** `aws_datasync_location_nfs` ([#6591](https://github.com/terraform-providers/terraform-provider-aws/issues/6591))
* **New Resource:** `aws_datasync_location_s3` ([#6591](https://github.com/terraform-providers/terraform-provider-aws/issues/6591))
* **New Resource:** `aws_datasync_task` ([#6591](https://github.com/terraform-providers/terraform-provider-aws/issues/6591))

## 1.47.0 (November 26, 2018)

FEATURES:

* **New Data Source:** `aws_route53_delegation_set` ([#6152](https://github.com/terraform-providers/terraform-provider-aws/issues/6152))
* **New Data Source:** `aws_ssm_document` ([#6479](https://github.com/terraform-providers/terraform-provider-aws/issues/6479))

ENHANCEMENTS:

* resource/aws_ecs_service: Add `enable_ecs_managed_tags` argument ([#6544](https://github.com/terraform-providers/terraform-provider-aws/issues/6544))
* resource/aws_kinesis_firehose_delivery_stream: Add `tags` argument ([#6548](https://github.com/terraform-providers/terraform-provider-aws/issues/6548))
* resource/aws_organizations_organization: Add `aws_service_access_principals` argument ([#6581](https://github.com/terraform-providers/terraform-provider-aws/issues/6581))
* resource/aws_s3_bucket_policy: Support resource import ([#6543](https://github.com/terraform-providers/terraform-provider-aws/issues/6543))
* resource/aws_vpc: Support plan-time validation for `cidr_block` block size ([#6577](https://github.com/terraform-providers/terraform-provider-aws/issues/6577))

BUG FIXES:

* resource/aws_elastic_transcoder_preset: Properly read `video_codec_options` into Terraform state ([#6545](https://github.com/terraform-providers/terraform-provider-aws/issues/6545))
* resource/aws_subnet: Always set `ipv6_cidr_block_association_id` and `ipv6_cidr_block` attributes in Terraform state ([#6533](https://github.com/terraform-providers/terraform-provider-aws/issues/6533))

## 1.46.0 (November 20, 2018)

FEATURES:

* **New Data Source:** `aws_api_gateway_api_key` ([#6449](https://github.com/terraform-providers/terraform-provider-aws/issues/6449))

ENHANCEMENTS:

* data-source/aws_eip: Add `association_id`, `domain`, `instance_id`, `network_interface_id`, `network_interface_owner_id`, `private_ip`, and `public_ipv4_pool` attributes ([#6463](https://github.com/terraform-providers/terraform-provider-aws/issues/6463)] / [[#6518](https://github.com/terraform-providers/terraform-provider-aws/issues/6518))
* resource/aws_ecs_cluster: Add `tags` argument ([#6486](https://github.com/terraform-providers/terraform-provider-aws/issues/6486))
* resource/aws_ecs_service: Add `tags` argument ([#6486](https://github.com/terraform-providers/terraform-provider-aws/issues/6486))
* resource/aws_ecs_task_definition: Add `tags` argument ([#6486](https://github.com/terraform-providers/terraform-provider-aws/issues/6486))
* resource/aws_ecs_task_definition: Add `ipc_mode` and `pid_mode` arguments ([#6515](https://github.com/terraform-providers/terraform-provider-aws/issues/6515))
* resource/aws_eip: Add `public_ipv4_pool` argument ([#6518](https://github.com/terraform-providers/terraform-provider-aws/issues/6518))
* resource/aws_iam_role: Add `tags` argument ([#6499](https://github.com/terraform-providers/terraform-provider-aws/issues/6499))
* resource/aws_iam_user: Add `tags` argument ([#6497](https://github.com/terraform-providers/terraform-provider-aws/issues/6497))
* resource/aws_sns_topic: Add `kms_master_key_id` argument (support server-side encryption) ([#6502](https://github.com/terraform-providers/terraform-provider-aws/issues/6502))

BUG FIXES:

* resource/aws_kinesis_analytics_application: Properly handle `processing_configuration` argument ([#6495](https://github.com/terraform-providers/terraform-provider-aws/issues/6495))

## 1.45.0 (November 15, 2018)

ENHANCEMENTS:

* resource/aws_autoscaling_group: Mixed Instances Policy support ([#6465](https://github.com/terraform-providers/terraform-provider-aws/issues/6465))

## 1.44.0 (November 14, 2018)

FEATURES:

* **New Resource:** `aws_gamelift_game_session_queue` ([#6335](https://github.com/terraform-providers/terraform-provider-aws/issues/6335))
* **New Resource:** `aws_glacier_vault_lock` ([#6432](https://github.com/terraform-providers/terraform-provider-aws/issues/6432))

ENHANCEMENTS:

* data-source/aws_eip: Add `filter` argument ([#3525](https://github.com/terraform-providers/terraform-provider-aws/issues/3525))
* data-source/aws_eip: Add `tags` argument ([#3505](https://github.com/terraform-providers/terraform-provider-aws/issues/3505))
* data-source/aws_eip: Support EC2-Classic Elastic IPs ([#3522](https://github.com/terraform-providers/terraform-provider-aws/issues/3522))
* resource/aws_codebuild_project: Support `source` `report_build_status` for Bitbucket ([#6426](https://github.com/terraform-providers/terraform-provider-aws/issues/6426))
* resource/aws_dlm_lifecycle_policy: Add `copy_tags` argument ([#6445](https://github.com/terraform-providers/terraform-provider-aws/issues/6445))
* resource/aws_ebs_snapshot: Allow retries for `SnapshotCreationPerVolumeRateExceeded` errors on creation ([#6414](https://github.com/terraform-providers/terraform-provider-aws/issues/6414))
* resource/aws_ebs_volume: Switch to tagging on creation ([#6396](https://github.com/terraform-providers/terraform-provider-aws/issues/6396))
* resource/aws_elastic_transcoder_pipeline: Support resource import ([#6388](https://github.com/terraform-providers/terraform-provider-aws/issues/6388))
* resource/aws_elastic_transcoder_preset: Support resource import ([#6388](https://github.com/terraform-providers/terraform-provider-aws/issues/6388))
* resource/aws_lambda_event_source_mapping: Add `starting_position_timestamp` argument ([#6437](https://github.com/terraform-providers/terraform-provider-aws/issues/6437))
* resource/aws_route53_health_check: Provide plan-time validation for `type` ([#6460](https://github.com/terraform-providers/terraform-provider-aws/issues/6460))
* resource/aws_ses_receipt_rule: Support resource import ([#6237](https://github.com/terraform-providers/terraform-provider-aws/issues/6237))
* resource/aws_ssm_maintenance_window_task: Add `description` and `name` arguments ([#5762](https://github.com/terraform-providers/terraform-provider-aws/issues/5762))

BUG FIXES:

* data-source/aws_ebs_snapshot: Fix `most_recent` ordering ([#6414](https://github.com/terraform-providers/terraform-provider-aws/issues/6414))
* resource/aws_cloudwatch_log_metric_filter: Properly leave `default_value` empty when unset ([#5933](https://github.com/terraform-providers/terraform-provider-aws/issues/5933))
* resource/aws_route53_health_check: Properly read `child_healthchecks` into Terraform state ([#6460](https://github.com/terraform-providers/terraform-provider-aws/issues/6460))
* resource/aws_security_group_rule: Support all non-zero `from_port` and `to_port` configurations with `protocol` ALL/-1 ([#6423](https://github.com/terraform-providers/terraform-provider-aws/issues/6423))
* resource/aws_sns_platform_application: Properly trigger resource recreation when deleted outside Terraform ([#6436](https://github.com/terraform-providers/terraform-provider-aws/issues/6436))
* service/ec2: Allow `tags` and `volume_tags` updates to retry based on SDK retries instead of time bounds for EC2 throttling ([#3586](https://github.com/terraform-providers/terraform-provider-aws/issues/3586))

## 1.43.2 (November 10, 2018)

BUG FIXES:

* resource/aws_security_group_rule: Prevent crash when reading rules from groups containing an `ALL`/`-1` `protocol` rule ([#6419](https://github.com/terraform-providers/terraform-provider-aws/issues/6419))

## 1.43.1 (November 09, 2018)

BUG FIXES:

* resource/aws_cloudwatch_metric_alarm: Accept EC2 automate reboot ARN ([#6405](https://github.com/terraform-providers/terraform-provider-aws/issues/6405))
* resource/aws_lambda_function: Handle slower code uploads on creation with configurable timeout ([#6409](https://github.com/terraform-providers/terraform-provider-aws/issues/6409))
* resource/aws_rds_cluster: Prevent `InvalidParameterCombination` error with `engine_version` and `snapshot_identifier` on creation ([#6391](https://github.com/terraform-providers/terraform-provider-aws/issues/6391))
* resource/aws_security_group_rule: Properly handle updating description when `protocol` is -1/ALL ([#6407](https://github.com/terraform-providers/terraform-provider-aws/issues/6407))
* resource/aws_vpc: Always set `assign_generated_ipv6_cidr_block`, `ipv6_association_id`, and `ipv6_cidr_block` attributes in Terraform state ([#2103](https://github.com/terraform-providers/terraform-provider-aws/issues/2103))
* resource/aws_vpc: Always wait for IPv6 CIDR block association on resource creation if `assign_generated_ipv6_cidr_block` is set ([#6394](https://github.com/terraform-providers/terraform-provider-aws/issues/6394))
* service/ec2: Properly ignore sending existing tags during updates ([#5108](https://github.com/terraform-providers/terraform-provider-aws/issues/5108)] / [[#6370](https://github.com/terraform-providers/terraform-provider-aws/issues/6370))

## 1.43.0 (November 07, 2018)

NOTES:

* resource/aws_lb_listener: This resource will now sort the API response based on action ordering. If necessary, sorting your configuration based on `order` should resolve any plan difference.
* resource/aws_lb_listener_rule: This resource will now sort the API response based on action ordering. If necessary, sorting your configuration based on `order` should resolve any plan difference.

FEATURES:

* **New Resource:** `aws_dlm_lifecycle_policy` ([#5558](https://github.com/terraform-providers/terraform-provider-aws/issues/5558))
* **New Resource:** `aws_kinesis_analytics_application` ([#5456](https://github.com/terraform-providers/terraform-provider-aws/issues/5456))

ENHANCEMENTS:

* data-source/aws_efs_file_system: Add `arn` attribute ([#6371](https://github.com/terraform-providers/terraform-provider-aws/issues/6371))
* data-source/aws_efs_mount_target: Add `file_system_arn` attribute ([#6371](https://github.com/terraform-providers/terraform-provider-aws/issues/6371))
* data-source/aws_mq_broker: Add `logs` attribute ([#6122](https://github.com/terraform-providers/terraform-provider-aws/issues/6122))
* resource/aws_efs_file_system: Add `arn` attribute ([#6371](https://github.com/terraform-providers/terraform-provider-aws/issues/6371))
* resource/aws_efs_mount_target: Add `file_system_arn` attribute ([#6371](https://github.com/terraform-providers/terraform-provider-aws/issues/6371))
* resource/aws_launch_configuration: Add `capacity_reservation_specification` argument ([#6325](https://github.com/terraform-providers/terraform-provider-aws/issues/6325))
* resource/aws_mq_broker: Add `logs` argument ([#6122](https://github.com/terraform-providers/terraform-provider-aws/issues/6122))

BUG FIXES:

* resource/aws_ecs_service: Continue supporting replica `deployment_minimum_healthy_percent = 0` and `deployment_maximum_percent = 100` ([#6316](https://github.com/terraform-providers/terraform-provider-aws/issues/6316))
* resource/aws_flow_log: Automatically trim `:*` suffix from `log_destination` argument ([#6377](https://github.com/terraform-providers/terraform-provider-aws/issues/6377))
* resource/aws_iam_user: Delete SSH keys with `force_delete` ([#6337](https://github.com/terraform-providers/terraform-provider-aws/issues/6337))
* resource/aws_lb_listener: Prevent panics with actions deleted outside Terraform ([#6319](https://github.com/terraform-providers/terraform-provider-aws/issues/6319))
* resource/aws_lb_listener_rule: Prevent panics with actions deleted outside Terraform ([#6319](https://github.com/terraform-providers/terraform-provider-aws/issues/6319))
* resource/aws_opsworks_application: Properly recreate resource on `short_name` updates ([#6359](https://github.com/terraform-providers/terraform-provider-aws/issues/6359))
* resource/aws_s3_bucket: Prevent `MalformedXML` error when using cross-region replication V1 with an empty `prefix` ([#6344](https://github.com/terraform-providers/terraform-provider-aws/issues/6344))

## 1.42.0 (October 31, 2018)

NOTES:

* resource/aws_route53_zone: The `vpc_id` and `vpc_region` arguments have been deprecated in favor of `vpc` configuration block(s). To upgrade, wrap existing `vpc_id` and `vpc_region` arguments with `vpc { ... }`. Since `vpc` is an exclusive set of VPC associations, you may need to define other `vpc` configuration blocks to match the infrastructure, or use lifecycle configuration `ignore_changes` to suppress the plan difference.
* resource/aws_route53_zone_association: Due to the multiple VPC association support now available in the `aws_route53_zone` resource, we recommend removing usage of this resource unless necessary for ordering. To remove this resource from management (without disassociating VPCs), you can use `terraform state rm`. If necessary to keep this resource for ordering, you can use the lifecycle `ignore_changes` in the `aws_route53_zone` resource to suppress plan differences.

FEATURES:

* **New Resource:** `aws_ec2_capacity_reservation` ([#6291](https://github.com/terraform-providers/terraform-provider-aws/issues/6291))
* **New Resource:** `aws_glue_security_configuration` ([#6288](https://github.com/terraform-providers/terraform-provider-aws/issues/6288))
* **New Resource:** `aws_iot_policy_attachment` ([#5864](https://github.com/terraform-providers/terraform-provider-aws/issues/5864))
* **New Resource:** `aws_iot_thing_principal_attachment` ([#5868](https://github.com/terraform-providers/terraform-provider-aws/issues/5868))
* **New Resource:** `aws_pinpoint_apns_sandbox_channel` ([#6233](https://github.com/terraform-providers/terraform-provider-aws/issues/6233))
* **New Resource:** `aws_pinpoint_apns_voip_channel` ([#6234](https://github.com/terraform-providers/terraform-provider-aws/issues/6234))
* **New Resource:** `aws_pinpoint_apns_voip_sandbox_channel` ([#6235](https://github.com/terraform-providers/terraform-provider-aws/issues/6235))

ENHANCEMENTS:

* data-source/aws_iot_endpoint: Add `endpoint_type` argument ([#6215](https://github.com/terraform-providers/terraform-provider-aws/issues/6215))
* data-source/aws_nat_gateway: Support `tags` as argument and attribute ([#6231](https://github.com/terraform-providers/terraform-provider-aws/issues/6231))
* resource/aws_budgets_budget: Support resource import ([#6226](https://github.com/terraform-providers/terraform-provider-aws/issues/6226))
* resource/aws_cloudwatch_event_permission: Add `condition` argument (support Organizations access) ([#6261](https://github.com/terraform-providers/terraform-provider-aws/issues/6261))
* resource/aws_codepipeline_webhook: Support resource import ([#6202](https://github.com/terraform-providers/terraform-provider-aws/issues/6202))
* resource/aws_cognito_user_pool_domain: Add `certificate_arn` argument (support custom domains) ([#6185](https://github.com/terraform-providers/terraform-provider-aws/issues/6185))
* resource/aws_dx_hosted_private_virtual_interface: Add `mtu` argument and `jumbo_frame_capable` attribute ([#6142](https://github.com/terraform-providers/terraform-provider-aws/issues/6142))
* resource/aws_dx_private_virtual_interface: Add `mtu` argument and `jumbo_frame_capable` attribute ([#6141](https://github.com/terraform-providers/terraform-provider-aws/issues/6141))
* resource/aws_ecs_service: Support `deployment_minimum_healthy_percent` for `DAEMON` strategy ([#6150](https://github.com/terraform-providers/terraform-provider-aws/issues/6150))
* resource/aws_flow_log: Add `log_destination` and `log_destination_type` arguments (support sending to S3) ([#5509](https://github.com/terraform-providers/terraform-provider-aws/issues/5509))
* resource/aws_glue_job: Add `security_configuration` argument ([#6232](https://github.com/terraform-providers/terraform-provider-aws/issues/6232))
* resource/aws_lb_target_group: Improve `name` and `name_prefix` argument plan-time validation ([#6168](https://github.com/terraform-providers/terraform-provider-aws/issues/6168))
* resource/aws_s3_bucket: Support S3 Cross-Region Replication filtering based on S3 object tags ([#6095](https://github.com/terraform-providers/terraform-provider-aws/issues/6095))
* resource/aws_secretsmanager_secret: Add `name_prefix` argument ([#6277](https://github.com/terraform-providers/terraform-provider-aws/issues/6277))
* resource/aws_secretsmanager_secret: Add plan-time validation for `name` argument ([#6277](https://github.com/terraform-providers/terraform-provider-aws/issues/6277))
* resource/aws_route53_zone: Add `vpc` argument, deprecate `vpc_id` and `vpc_region` arguments (support multiple VPC associations) ([#6299](https://github.com/terraform-providers/terraform-provider-aws/issues/6299))
* resource/aws_waf_rule: Support resource import ([#6247](https://github.com/terraform-providers/terraform-provider-aws/issues/6247))

BUG FIXES:

* data-source/aws_network_interface: Properly handle reading `private_ip` into Terraform state ([#6284](https://github.com/terraform-providers/terraform-provider-aws/issues/6284))
* resource/aws_ami_launch_permission: Prevent panic reading public permissions ([#6224](https://github.com/terraform-providers/terraform-provider-aws/issues/6224))
* resource/aws_budgets_budget: Properly read `time_period_start` and `time_period_end` into Terraform state ([#6226](https://github.com/terraform-providers/terraform-provider-aws/issues/6226))
* resource/aws_cloudwatch_metric_alarm: Allow EC2 Automate ARNs with `alarm_actions` ([#6206](https://github.com/terraform-providers/terraform-provider-aws/issues/6206))
* resource/aws_dx_gateway: Allow legacy `amazon_side_asn` in plan-time validation ([#6253](https://github.com/terraform-providers/terraform-provider-aws/issues/6253))
* resource/aws_egress_only_internet_gateway: Improve eventual consistency logic during creation ([#6190](https://github.com/terraform-providers/terraform-provider-aws/issues/6190))
* resource/aws_glue_crawler: Suppress `role` difference when using ARN ([#6293](https://github.com/terraform-providers/terraform-provider-aws/issues/6293))
* resource/aws_iam_role_policy: Properly handle reading attributes into Terraform state after creation and update ([#6304](https://github.com/terraform-providers/terraform-provider-aws/issues/6304))
* resource/aws_kinesis_firehose_delivery_stream: Properly recreate resource when updating `elasticsearch_configuration` `s3_backup_mode` ([#6305](https://github.com/terraform-providers/terraform-provider-aws/issues/6305))
* resource/aws_nat_gateway: Remove `network_interface_id`, `private_ip`, and `public_ip` as configurable (they continue to be available as read-only attributes) ([#6225](https://github.com/terraform-providers/terraform-provider-aws/issues/6225))
* resource/aws_network_acl: Properly handle ICMP code and type with IPv6 ICMP (protocol 58) ([#6264](https://github.com/terraform-providers/terraform-provider-aws/issues/6264))
* resource/aws_network_acl_rule: Suppress `protocol` differences between name and number ([#2454](https://github.com/terraform-providers/terraform-provider-aws/issues/2454))
* resource/aws_network_acl_rule: Properly handle ICMP code and type with IPv6 ICMP (protocol 58) ([#6263](https://github.com/terraform-providers/terraform-provider-aws/issues/6263))
* resource/aws_rds_cluster_parameter_group: Properly read `parameter` `apply_method` into Terraform state ([#6295](https://github.com/terraform-providers/terraform-provider-aws/issues/6295))

## 1.41.0 (October 18, 2018)

FEATURES:

* **New Data Source:** `aws_cloudhsm_v2_cluster` ([#4125](https://github.com/terraform-providers/terraform-provider-aws/issues/4125))
* **New Resource:** `aws_cloudhsm_v2_cluster` ([#4125](https://github.com/terraform-providers/terraform-provider-aws/issues/4125))
* **New Resource:** `aws_cloudhsm_v2_hsm` ([#4125](https://github.com/terraform-providers/terraform-provider-aws/issues/4125))
* **New Resource:** `aws_codepipeline_webhook` ([#5875](https://github.com/terraform-providers/terraform-provider-aws/issues/5875))
* **New Resource:** `aws_pinpoint_apns_channel` ([#6194](https://github.com/terraform-providers/terraform-provider-aws/issues/6194))
* **New Resource:** `aws_redshift_event_subscription` ([#6146](https://github.com/terraform-providers/terraform-provider-aws/issues/6146))

ENHANCEMENTS:

* resource/aws_appsync_datasource: Support resource import ([#6139](https://github.com/terraform-providers/terraform-provider-aws/issues/6139))
* resource/aws_appsync_datasource: Support `HTTP` `type` and add `http_config` argument ([#6139](https://github.com/terraform-providers/terraform-provider-aws/issues/6139))
* resource/aws_appsync_datasource: Make `dynamodb_config` and `elasticsearch_config` `region` configuration optional based on resource current region ([#6139](https://github.com/terraform-providers/terraform-provider-aws/issues/6139))
* resource/aws_appsync_graphql_api: Add `log_config` argument ([#6138](https://github.com/terraform-providers/terraform-provider-aws/issues/6138))
* resource/aws_appsync_graphql_api: Add `openid_connect_config` argument ([#6138](https://github.com/terraform-providers/terraform-provider-aws/issues/6138))
* resource/aws_appsync_graphql_api: Add `uris` attribute ([#6138](https://github.com/terraform-providers/terraform-provider-aws/issues/6138))
* resource/aws_appsync_graphql_api: Make `user_pool_config` `aws_region` configuration optional based on resource current region ([#6138](https://github.com/terraform-providers/terraform-provider-aws/issues/6138))
* resource/aws_athena_database: Add `encryption_configuration` argument ([#6117](https://github.com/terraform-providers/terraform-provider-aws/issues/6117))
* resource/aws_cloudwatch_metric_alarm: Validate `alarm_actions` ([#6151](https://github.com/terraform-providers/terraform-provider-aws/issues/6151))
* resource/aws_codebuild_project: Support `NO_SOURCE` in `source` `type` ([#6140](https://github.com/terraform-providers/terraform-provider-aws/issues/6140))
* resource/aws_db_instance: Directly restore snapshot with `parameter_group_name` set ([#6200](https://github.com/terraform-providers/terraform-provider-aws/issues/6200))
* resource/aws_dx_connection: Add `jumbo_frame_capable` attribute ([#6143](https://github.com/terraform-providers/terraform-provider-aws/issues/6143))
* resource/aws_dynamodb_table: Prevent error `UnknownOperationException: Tagging is not currently supported in DynamoDB Local` ([#6149](https://github.com/terraform-providers/terraform-provider-aws/issues/6149))
* resource/aws_lb_listener: Allow `default_action` `order` to be based on Terraform configuration ordering ([#6124](https://github.com/terraform-providers/terraform-provider-aws/issues/6124))
* resource/aws_lb_listener_rule: Allow `action` `order` to be based on Terraform configuration ordering ([#6124](https://github.com/terraform-providers/terraform-provider-aws/issues/6124))
* resource/aws_rds_cluster: Directly restore snapshot with `db_cluster_parameter_group_name` set ([#6200](https://github.com/terraform-providers/terraform-provider-aws/issues/6200))

BUG FIXES:

* resource/aws_appsync_graphql_api: Properly handle updates by passing all parameters ([#6138](https://github.com/terraform-providers/terraform-provider-aws/issues/6138))
* resource/aws_ecs_service: Properly handle `random` placement strategy ([#6176](https://github.com/terraform-providers/terraform-provider-aws/issues/6176))
* resource/aws_lb_listener: Prevent unconfigured `default_action` `order` from showing difference ([#6119](https://github.com/terraform-providers/terraform-provider-aws/issues/6119))
* resource/aws_lb_listener_rule: Prevent unconfigured `action` `order` from showing difference ([#6119](https://github.com/terraform-providers/terraform-provider-aws/issues/6119))
* resource/aws_lb_listener_rule: Retry read for eventual consistency after resource creation ([#6154](https://github.com/terraform-providers/terraform-provider-aws/issues/6154))

## 1.40.0 (October 10, 2018)

FEATURES:

* **New Data Source:** `aws_launch_template` ([#6064](https://github.com/terraform-providers/terraform-provider-aws/issues/6064))
* **New Data Source:** `aws_workspaces_bundle` ([#3243](https://github.com/terraform-providers/terraform-provider-aws/issues/3243))
* **New Guide:** [`AWS IAM Policy Documents`](https://www.terraform.io/docs/providers/aws/guides/iam-policy-documents.html) ([#6016](https://github.com/terraform-providers/terraform-provider-aws/issues/6016))
* **New Resource:** `aws_ebs_snapshot_copy` ([#3086](https://github.com/terraform-providers/terraform-provider-aws/issues/3086))
* **New Resource:** `aws_pinpoint_adm_channel` ([#6038](https://github.com/terraform-providers/terraform-provider-aws/issues/6038))
* **New Resource:** `aws_pinpoint_baidu_channel` ([#6111](https://github.com/terraform-providers/terraform-provider-aws/issues/6111))
* **New Resource:** `aws_pinpoint_email_channel` ([#6110](https://github.com/terraform-providers/terraform-provider-aws/issues/6110))
* **New Resource:** `aws_pinpoint_event_stream` ([#6069](https://github.com/terraform-providers/terraform-provider-aws/issues/6069))
* **New Resource:** `aws_pinpoint_gcm_channel` ([#6089](https://github.com/terraform-providers/terraform-provider-aws/issues/6089))
* **New Resource:** `aws_pinpoint_sms_channel` ([#6088](https://github.com/terraform-providers/terraform-provider-aws/issues/6088))
* **New Resource:** `aws_redshift_snapshot_copy_grant` ([#5134](https://github.com/terraform-providers/terraform-provider-aws/issues/5134))

ENHANCEMENTS:

* data-source/aws_iam_policy_document: Make `statement` argument optional ([#6052](https://github.com/terraform-providers/terraform-provider-aws/issues/6052))
* data-source/aws_secretsmanager_secret: Add `policy` attribute ([#6091](https://github.com/terraform-providers/terraform-provider-aws/issues/6091))
* data-source/aws_secretsmanager_secret_version: Add `secret_binary` attribute ([#6070](https://github.com/terraform-providers/terraform-provider-aws/issues/6070))
* resource/aws_codebuild_project: Add `environment` `certificate` argument ([#6087](https://github.com/terraform-providers/terraform-provider-aws/issues/6087))
* resource/aws_ecr_repository: Add configurable `delete` timeout ([#3910](https://github.com/terraform-providers/terraform-provider-aws/issues/3910))
* resource/aws_elastic_beanstalk_environment: Add `platform_arn` argument (support custom platforms) ([#6093](https://github.com/terraform-providers/terraform-provider-aws/issues/6093))
* resource/aws_lb_listener: Support Cognito and OIDC authentication ([#6094](https://github.com/terraform-providers/terraform-provider-aws/issues/6094))
* resource/aws_lb_listener_rule: Support Cognito and OIDC authentication ([#6094](https://github.com/terraform-providers/terraform-provider-aws/issues/6094))
* resource/aws_mq_broker: Add `instances` `ip_address` attribute ([#6103](https://github.com/terraform-providers/terraform-provider-aws/issues/6103))
* resource/aws_rds_cluster: Support `engine_version` updates ([#5010](https://github.com/terraform-providers/terraform-provider-aws/issues/5010))
* resource/aws_s3_bucket: Add replication `access_control_translation` and `account_id` arguments (support cross-account replication ownership) ([#3577](https://github.com/terraform-providers/terraform-provider-aws/issues/3577))
* resource/aws_secretsmanager_secret_version: Add `secret_binary` argument ([#6070](https://github.com/terraform-providers/terraform-provider-aws/issues/6070))
* resource/aws_security_group_rule: Support resource import ([#6027](https://github.com/terraform-providers/terraform-provider-aws/issues/6027))

BUG FIXES:

* resource/aws_appautoscaling_policy: Properly handle negative values in step scaling metric intervals ([#3480](https://github.com/terraform-providers/terraform-provider-aws/issues/3480))
* resource/aws_appsync_datasource: Properly pass all attributes during update ([#5814](https://github.com/terraform-providers/terraform-provider-aws/issues/5814))
* resource/aws_batch_job_queue: Prevent error during read of non-existent Job Queue ([#6085](https://github.com/terraform-providers/terraform-provider-aws/issues/6085))
* resource/aws_ecr_repository: Retry read for eventual consistency after resource creation ([#3910](https://github.com/terraform-providers/terraform-provider-aws/issues/3910))
* resource/aws_ecs_service: Properly remove non-existent services from Terraform state ([#6039](https://github.com/terraform-providers/terraform-provider-aws/issues/6039))
* resource/aws_iam_instance_profile: Retry for eventual consistency when adding a role ([#6079](https://github.com/terraform-providers/terraform-provider-aws/issues/6079))
* resource/aws_lb_listener: Retry read for eventual consistency after resource creation ([#5167](https://github.com/terraform-providers/terraform-provider-aws/issues/5167))

## 1.39.0 (October 03, 2018)

FEATURES:

* **New Resource:** `aws_ec2_fleet` ([#5960](https://github.com/terraform-providers/terraform-provider-aws/issues/5960))
* **New Resource:** `aws_pinpoint_app` ([#5956](https://github.com/terraform-providers/terraform-provider-aws/issues/5956))

ENHANCEMENTS:

* resource/aws_cloudwatch_event_target: Support additional ECS target arguments ([#5982](https://github.com/terraform-providers/terraform-provider-aws/issues/5982))
* resource/aws_codedeploy_app: Support resource import ([#6025](https://github.com/terraform-providers/terraform-provider-aws/issues/6025))
* resource/aws_codedeploy_deployment_config: Support resource import ([#6025](https://github.com/terraform-providers/terraform-provider-aws/issues/6025))
* resource/aws_codedeploy_deployment_group: Support resource import ([#6025](https://github.com/terraform-providers/terraform-provider-aws/issues/6025))
* resource/aws_db_instance: Add `deletion_protection` argument ([#6011](https://github.com/terraform-providers/terraform-provider-aws/issues/6011))
* resource/aws_dx_connection: Support 50Mbps, 100Mbps, 200Mbps, 300Mbps, 400Mbps, 500Mbps as valid `bandwidth` values ([#6057](https://github.com/terraform-providers/terraform-provider-aws/issues/6057))
* resource/aws_dx_lag: Support 50Mbps, 100Mbps, 200Mbps, 300Mbps, 400Mbps, 500Mbps as valid `connections_bandwidth` values ([#6057](https://github.com/terraform-providers/terraform-provider-aws/issues/6057))
* resource/aws_elasticsearch_domain: Add `node_to_node_encryption` argument ([#5997](https://github.com/terraform-providers/terraform-provider-aws/issues/5997))
* resource/aws_rds_cluster: Add `deletion_protection` argument ([#6010](https://github.com/terraform-providers/terraform-provider-aws/issues/6010))
* resource/aws_sns_topic_subscription: Add `delivery_policy` argument ([#3289](https://github.com/terraform-providers/terraform-provider-aws/issues/3289))
* resource/aws_spot_fleet_request: Add `instance_pools_to_use_count` argument ([#5955](https://github.com/terraform-providers/terraform-provider-aws/issues/5955))

BUG FIXES:

* resource/aws_api_gateway_deployment: Do not delete stage if it is in use by another deployment ([#3896](https://github.com/terraform-providers/terraform-provider-aws/issues/3896))
* resource/aws_codedeploy_deployment_group: Include autoscaling groups when updating blue green config ([#5827](https://github.com/terraform-providers/terraform-provider-aws/issues/5827))
* resource/aws_codedeploy_deployment_group: Properly read `autoscaling_groups` into Terraform state ([#6025](https://github.com/terraform-providers/terraform-provider-aws/issues/6025))
* resource/aws_ecs_task_definition: Properly handle task scoped docker volume configurations ([#5907](https://github.com/terraform-providers/terraform-provider-aws/issues/5907))
* resource/aws_network_interface_sg_attachment: Properly handle `InvalidNetworkInterfaceID.NotFound` errors ([#6048](https://github.com/terraform-providers/terraform-provider-aws/issues/6048))
* resource/aws_rds_cluster: Properly handle `kms_key_id` when restoring from snapshot ([#6012](https://github.com/terraform-providers/terraform-provider-aws/issues/6012))
* resource/aws_s3_bucket_object: Mark `version_id` as recomputed on `etag` updates ([#3861](https://github.com/terraform-providers/terraform-provider-aws/issues/3861))
* resource/aws_security_group: Prevent `InvalidNetworkInterfaceID.NotFound` errors when deleting lingering network interfaces ([#6037](https://github.com/terraform-providers/terraform-provider-aws/issues/6037))
* resource/aws_sns_topic_subscription: Properly read all attributes into Terraform state on reads ([#6023](https://github.com/terraform-providers/terraform-provider-aws/issues/6023))
* resource/aws_sns_topic_subscription: Properly handle `filter_policy` removal ([#6023](https://github.com/terraform-providers/terraform-provider-aws/issues/6023))
* resource/aws_subnet: Prevent `InvalidNetworkInterfaceID.NotFound` errors when deleting lingering network interfaces ([#6037](https://github.com/terraform-providers/terraform-provider-aws/issues/6037))

## 1.38.0 (September 26, 2018)

FEATURES:

* **New Data Source:** `aws_db_event_categories` ([#5514](https://github.com/terraform-providers/terraform-provider-aws/issues/5514))

ENHANCEMENTS:

* data-source/aws_autoscaling_groups: Add `arns` attribute ([#5766](https://github.com/terraform-providers/terraform-provider-aws/issues/5766))
* resource/aws_ami: Support resource import ([#5990](https://github.com/terraform-providers/terraform-provider-aws/issues/5990))
* resource/aws_codebuild_project: Add `secondary_artifacts` and `secondary_sources` arguments ([#5939](https://github.com/terraform-providers/terraform-provider-aws/issues/5939))
* resource/aws_codebuild_project: Add `arn` attribute ([#5973](https://github.com/terraform-providers/terraform-provider-aws/issues/5973))
* resource/aws_launch_template: Support `credit_specification` configuration of T3 instance types ([#5922](https://github.com/terraform-providers/terraform-provider-aws/issues/5922))
* resource/aws_launch_template: Allow `network_interface` `ipv6_address_count` configuration ([#5771](https://github.com/terraform-providers/terraform-provider-aws/issues/5771))
* resource/aws_rds_cluster: Support `parallelquery` `engine_mode` argument ([#5980](https://github.com/terraform-providers/terraform-provider-aws/issues/5980))

BUG FIXES:

* data-source/aws_ami: Prevent panics with AMIs in failed image state ([#5968](https://github.com/terraform-providers/terraform-provider-aws/issues/5968))
* resource/aws_db_instance: Properly set `backup_retention_period = 0` with `snapshot_identifier` ([#5970](https://github.com/terraform-providers/terraform-provider-aws/issues/5970))
* resource/aws_dms_replication_instance: Properly handle `engine_version` updates ([#5948](https://github.com/terraform-providers/terraform-provider-aws/issues/5948))
* resource/aws_launch_template: Prevent `Auto Scaling only supports the 'one-time' Spot instance type with no duration.` error when using `instance_market_options` and AutoScaling Groups ([#5957](https://github.com/terraform-providers/terraform-provider-aws/issues/5957))
* resource/aws_launch_template: Properly recreate existing resource when deleted ([#5967](https://github.com/terraform-providers/terraform-provider-aws/issues/5967))
* resource/aws_launch_template: Continue accepting string `"true"` and `"false"` values for `ebs_optimized` argument ([#5995](https://github.com/terraform-providers/terraform-provider-aws/issues/5995))
* resource/aws_load_balancer_policy: Properly handle resource when ELB is deleted ([#5972](https://github.com/terraform-providers/terraform-provider-aws/issues/5972))
* resource/aws_rds_cluster_instance: Properly handle `publicly_accessible` updates ([#5991](https://github.com/terraform-providers/terraform-provider-aws/issues/5991))
* resource/aws_security_group: Properly handle lingering ENIs from Lambda and similar services ([#4884](https://github.com/terraform-providers/terraform-provider-aws/issues/4884))
* resource/aws_subnet: Properly handle lingering ENIs from Lambda and similar services ([#4884](https://github.com/terraform-providers/terraform-provider-aws/issues/4884))

## 1.37.0 (September 19, 2018)

FEATURES:

* **New Resource:** `aws_dx_bgp_peer` ([#5886](https://github.com/terraform-providers/terraform-provider-aws/issues/5886))

ENHANCEMENTS:

* data-source/aws_ami_ids: Add `sort_ascending` argument ([#5912](https://github.com/terraform-providers/terraform-provider-aws/issues/5912))
* resource/aws_iam_role_policy_attachment: Support resource import ([#5910](https://github.com/terraform-providers/terraform-provider-aws/issues/5910))
* resource/aws_s3_bucket_inventory: Allow SSE-S3 encryption ([#5870](https://github.com/terraform-providers/terraform-provider-aws/issues/5870))
* resource/aws_security_group: Add `prefix_list_ids` argument for `ingress` rules ([#5916](https://github.com/terraform-providers/terraform-provider-aws/issues/5916))

BUG FIXES:

* resource/aws_config_config_rule: Prevent panic when specifying empty `scope` ([#5852](https://github.com/terraform-providers/terraform-provider-aws/issues/5852))
* resource/aws_iam_policy: Ensure `description` is properly read into Terraform state during resource creation ([#5884](https://github.com/terraform-providers/terraform-provider-aws/issues/5884))
* resource/aws_instance: Properly handle `credit_specifications` with T3 instance types ([#5805](https://github.com/terraform-providers/terraform-provider-aws/issues/5805))
* resource/aws_launch_template: Fix handling of `network_interface` `ipv6_addresses` ([#5883](https://github.com/terraform-providers/terraform-provider-aws/issues/5883))
* resource/aws_redshift_cluster: Properly disable logging when using `logging` nested argument ([#5895](https://github.com/terraform-providers/terraform-provider-aws/issues/5895))
* resource/aws_s3_bucket: Prevent panics with various API read failures ([#5842](https://github.com/terraform-providers/terraform-provider-aws/issues/5842))
* resource/aws_s3_bucket: Prevent `NoSuchBucket` error on deletion ([#5842](https://github.com/terraform-providers/terraform-provider-aws/issues/5842))
* resource/aws_wafregional_byte_match_set: Properly read `byte_match_tuple` into Terraform state ([#5902](https://github.com/terraform-providers/terraform-provider-aws/issues/5902))

## 1.36.0 (September 13, 2018)

FEATURES:

* **New Resource:** `aws_cloudfront_public_key` ([#5737](https://github.com/terraform-providers/terraform-provider-aws/issues/5737))

ENHANCEMENTS:

* data-source/aws_db_instance: Add `enabled_cloudwatch_logs_exports` attribute ([#5801](https://github.com/terraform-providers/terraform-provider-aws/issues/5801))
* resource/aws_api_gateway_stage: Add `xray_tracing_enabled` argument ([#5817](https://github.com/terraform-providers/terraform-provider-aws/issues/5817))
* resource/aws_cloudfront_distribution: Add `lambda_function_association` `include_body` argument ([#5681](https://github.com/terraform-providers/terraform-provider-aws/issues/5681))
* resource/aws_db_instance: Add `domain` and `domain_iam_role_name` arguments (support for domain joining RDS instances) ([#5378](https://github.com/terraform-providers/terraform-provider-aws/issues/5378))
* resource/aws_ecs_task_definition: Suppress `container_definition` differences for equivalent port and host mappings ([#5833](https://github.com/terraform-providers/terraform-provider-aws/issues/5833))
* resource/aws_ecs_task_definition: Add docker volume configuration ([#5727](https://github.com/terraform-providers/terraform-provider-aws/issues/5727))
* resource/aws_iam_user: Allow empty string (`""`) value for `permissions_boundary` argument ([#5859](https://github.com/terraform-providers/terraform-provider-aws/issues/5859))
* resource/aws_iot_topic_rule: Add `firehose` `separator` argument ([#5734](https://github.com/terraform-providers/terraform-provider-aws/issues/5734))
* resource/aws_launch_template: Allow `network_interface` `ipv4_address_count` configuration ([#5830](https://github.com/terraform-providers/terraform-provider-aws/issues/5830))
* resource/aws_ssm_document: Add support for `Session` `document_type` ([#5850](https://github.com/terraform-providers/terraform-provider-aws/issues/5850))

BUG FIXES:

* resource/aws_iam_policy: Ensure `description` is available as an attribute when empty ([#5815](https://github.com/terraform-providers/terraform-provider-aws/issues/5815))
* resource/aws_iam_user: Remove extraneous `DeleteUserPermissionsBoundary` API call during deletion ([#5857](https://github.com/terraform-providers/terraform-provider-aws/issues/5857))
* resource/aws_lambda_function: Retry on `InvalidParameterValueException` errors relating to KMS-backed environment variables ([#5849](https://github.com/terraform-providers/terraform-provider-aws/issues/5849))
* resource/aws_launch_template: Ensure `ebs_optimized` argument accepts "unspecified" value ([#5627](https://github.com/terraform-providers/terraform-provider-aws/issues/5627))

## 1.35.0 (September 06, 2018)

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `platform_version` attribute ([#5797](https://github.com/terraform-providers/terraform-provider-aws/issues/5797))
* resource/aws_eks_cluster: Add `platform_version` attribute ([#5797](https://github.com/terraform-providers/terraform-provider-aws/issues/5797))
* resource/aws_lambda_function: Allow empty lists for `vpc_config` `security_group_ids` and `subnet_ids` arguments to unconfigure VPC ([#1341](https://github.com/terraform-providers/terraform-provider-aws/issues/1341))
* resource/aws_iam_role: Allow empty string (`""`) value for `permissions_boundary` argument ([#5740](https://github.com/terraform-providers/terraform-provider-aws/issues/5740))

BUG FIXES:

* resource/aws_ecr_repository: Use `RepositoryUri` instead of our building our own URI for the `repository_url` attribute (AWS China fix) ([#5748](https://github.com/terraform-providers/terraform-provider-aws/issues/5748))
* resource/aws_lambda_function: Properly handle `vpc_config` removal ([#5798](https://github.com/terraform-providers/terraform-provider-aws/issues/5798))
* resource/aws_redshift_cluster: Properly force new resource when updating `availability_zone` argument ([#5758](https://github.com/terraform-providers/terraform-provider-aws/issues/5758))

## 1.34.0 (August 30, 2018)

NOTES:

* provider: This is the first release tested against and built with Go 1.11, which required `go fmt` changes to the code. If you are building a custom version of this provider or running tests using the repository Make targets (e.g. `make build`) when using a previous version of Go, you will receive errors. You can use the underlying `go` commands (e.g. `go build`) to workaround the `go fmt` check in the Make targets until you are able to upgrade Go.

ENHANCEMENTS:

* provider: `NO_PROXY` environment variable can accept CIDR notation and port
* data-source/aws_ip_ranges: Add `ipv6_cidr_blocks` attribute ([#5675](https://github.com/terraform-providers/terraform-provider-aws/issues/5675))
* resource/aws_codebuild_project: Add `artifacts` `encryption_disabled` argument ([#5678](https://github.com/terraform-providers/terraform-provider-aws/issues/5678))
* resource/aws_route: Support route import ([#5687](https://github.com/terraform-providers/terraform-provider-aws/issues/5687))

BUG FIXES:

* data-source/aws_rds_cluster: Prevent error setting `engine_mode` and `scaling_configuration` ([#5660](https://github.com/terraform-providers/terraform-provider-aws/issues/5660))
* resource/aws_autoscaling_group: Retry creation for eventual consistency with launch template IAM instance profile ([#5633](https://github.com/terraform-providers/terraform-provider-aws/issues/5633))
* resource/aws_dax_cluster: Properly recreate cluster when updating `server_side_encryption` ([#5664](https://github.com/terraform-providers/terraform-provider-aws/issues/5664))
* resource/aws_db_instance: Prevent double apply when using `replicate_source_db` parameters that require `ModifyDBInstance` during resource creation ([#5672](https://github.com/terraform-providers/terraform-provider-aws/issues/5672))
* resource/aws_db_instance: Prevent `pending-reboot` parameter group status on creation with `parameter_group_name` ([#5672](https://github.com/terraform-providers/terraform-provider-aws/issues/5672))
* resource/aws_lambda_event_source_mapping: Prevent perpetual difference when using function name with `function_name` (argument accepts both name and ARN) ([#5454](https://github.com/terraform-providers/terraform-provider-aws/issues/5454))
* resource/aws_launch_template: Prevent encrypted flag cannot be specified error with `block_device_mappings` `ebs` argument ([#5632](https://github.com/terraform-providers/terraform-provider-aws/issues/5632))
* resource/aws_key_pair: Ensure `fingerprint` attribute is saved in Terraform state during creation ([#5732](https://github.com/terraform-providers/terraform-provider-aws/issues/5732))
* resource/aws_ssm_association: Properly handle updates when multiple arguments are used ([#5537](https://github.com/terraform-providers/terraform-provider-aws/issues/5537))
* resource/aws_ssm_document: Properly handle deletion of privately shared documents ([#5668](https://github.com/terraform-providers/terraform-provider-aws/issues/5668))
* resource/aws_ssm_document: Properly update `permissions.account_ids` ([#5685](https://github.com/terraform-providers/terraform-provider-aws/issues/5685))

## 1.33.0 (August 22, 2018)

FEATURES:

* **New Data Source:** `aws_api_gateway_resource` ([#5629](https://github.com/terraform-providers/terraform-provider-aws/issues/5629))

ENHANCEMENTS:

* data-source/aws_storagegateway_local_disk: Add `disk_node` argument ([#5595](https://github.com/terraform-providers/terraform-provider-aws/issues/5595))
* resource/aws_api_gateway_base_path_mapping: Support resource import ([#5566](https://github.com/terraform-providers/terraform-provider-aws/issues/5566))
* resource/aws_api_gateway_gateway_response: Support resource import ([#5567](https://github.com/terraform-providers/terraform-provider-aws/issues/5567))
* resource/aws_api_gateway_integration: Support resource import ([#5568](https://github.com/terraform-providers/terraform-provider-aws/issues/5568))
* resource/aws_api_gateway_integration_response: Support resource import ([#5569](https://github.com/terraform-providers/terraform-provider-aws/issues/5569))
* resource/aws_api_gateway_method: Support resource import ([#5571](https://github.com/terraform-providers/terraform-provider-aws/issues/5571))
* resource/aws_api_gateway_method_response: Support resource import ([#5570](https://github.com/terraform-providers/terraform-provider-aws/issues/5570))
* resource/aws_api_gateway_model: Support resource import ([#5572](https://github.com/terraform-providers/terraform-provider-aws/issues/5572))
* resource/aws_api_gateway_request_validator: Support resource import ([#5573](https://github.com/terraform-providers/terraform-provider-aws/issues/5573))
* resource/aws_api_gateway_resource: Support resource import ([#5574](https://github.com/terraform-providers/terraform-provider-aws/issues/5574))
* resource/aws_api_gateway_rest_api: Support resource import ([#5564](https://github.com/terraform-providers/terraform-provider-aws/issues/5564))
* resource/aws_api_gateway_stage: Support resource import ([#5575](https://github.com/terraform-providers/terraform-provider-aws/issues/5575))
* resource/aws_dax_cluster: Add `server_side_encryption` argument (support encryption at rest) ([#5508](https://github.com/terraform-providers/terraform-provider-aws/issues/5508))
* resource/aws_ecs_service: Add retries for target group attachment ([#3535](https://github.com/terraform-providers/terraform-provider-aws/issues/3535))
* resource/aws_lb_listener: Add support for 'redirect' and 'fixed-response' actions ([#5430](https://github.com/terraform-providers/terraform-provider-aws/issues/5430))
* resource/aws_lb_listener_rule: Add support for 'redirect' and 'fixed-response' actions ([#5430](https://github.com/terraform-providers/terraform-provider-aws/issues/5430))
* resource/aws_rds_cluster: Add `scaling_configuration` argument ([#5531](https://github.com/terraform-providers/terraform-provider-aws/issues/5531))
* resource/aws_secretsmanager_secret: Support `ForceDeleteWithoutRecovery` (via `recovery_window_in_days = 0`) and secret recreation after immediate deletion ([#5583](https://github.com/terraform-providers/terraform-provider-aws/issues/5583))

BUG FIXES:

* provider: Disable AWS SDK retries faster by default for `connection refused` errors ([#5614](https://github.com/terraform-providers/terraform-provider-aws/issues/5614))
* resource/aws_api_gateway_integration: Properly read `integration_http_method` into Terraform state ([#5568](https://github.com/terraform-providers/terraform-provider-aws/issues/5568))
* resource/aws_api_gateway_integration_response: Properly read `content_handling` into Terraform state ([#5569](https://github.com/terraform-providers/terraform-provider-aws/issues/5569))
* resource/aws_api_gateway_integration_response: Properly read `response_templates` into Terraform state ([#5569](https://github.com/terraform-providers/terraform-provider-aws/issues/5569))
* resource/aws_cloudfront_distribution: Import into `ordered_cache_behavior` instead of deprecated `cache_behavior` ([#5586](https://github.com/terraform-providers/terraform-provider-aws/issues/5586))
* resource/aws_db_instance: Prevent error when using `snapshot_identifier` with `multi_az` enabled and sqlserver `engine` ([#5613](https://github.com/terraform-providers/terraform-provider-aws/issues/5613))
* resource/aws_db_instance: Prevent double apply when using `snapshot_identifier` parameters that require `ModifyDBInstance` during resource creation ([#5613](https://github.com/terraform-providers/terraform-provider-aws/issues/5613)] / [[#5621](https://github.com/terraform-providers/terraform-provider-aws/issues/5621))
* resource/aws_db_instance: Prevent `is already being deleted` error on deletion and wait for deletion completion ([#5624](https://github.com/terraform-providers/terraform-provider-aws/issues/5624))
* resource/aws_ecs_task_definition: Treat `INACTIVE` task definitions as removed ([#5565](https://github.com/terraform-providers/terraform-provider-aws/issues/5565))
* resource/aws_elasticache_cluster: Allow `availability_zone` to be specified with `replication_group_id` ([#5585](https://github.com/terraform-providers/terraform-provider-aws/issues/5585))
* resource/aws_instance: Ignore change of `user_data` from omission to empty string ([#5467](https://github.com/terraform-providers/terraform-provider-aws/issues/5467))
* resource/aws_service_discovery_public_dns_namespace: Prevent creation error with names longer than 34 characters ([#5610](https://github.com/terraform-providers/terraform-provider-aws/issues/5610))
* resource/aws_waf_ipset: Properly handle updates and deletions over 1000 IP set descriptors ([#5588](https://github.com/terraform-providers/terraform-provider-aws/issues/5588))
* resource/aws_wafregional_ipset: Properly handle updates and deletions over 1000 IP set descriptors ([#5588](https://github.com/terraform-providers/terraform-provider-aws/issues/5588))

## 1.32.0 (August 16, 2018)

FEATURES:

* **New Resource:** `aws_neptune_cluster_snapshot` ([#5492](https://github.com/terraform-providers/terraform-provider-aws/issues/5492))
* **New Resource:** `aws_storagegateway_cached_iscsi_volume` ([#5476](https://github.com/terraform-providers/terraform-provider-aws/issues/5476))

ENHANCEMENTS:

* data-source/aws_secretsmanager_secret_version: Add `arn` attribute ([#5488](https://github.com/terraform-providers/terraform-provider-aws/issues/5488))
* data-source/aws_subnet: Add `arn` attribute ([#5486](https://github.com/terraform-providers/terraform-provider-aws/issues/5486))
* resource/aws_cloudwatch_metric_alarm: Add `arn` attribute ([#5487](https://github.com/terraform-providers/terraform-provider-aws/issues/5487))
* resource/aws_db_instance: Allow `alert`, `listener`, and `trace` for `enabled_cloudwatch_logs_exports` (e.g. Oracle specific log exports) ([#5494](https://github.com/terraform-providers/terraform-provider-aws/issues/5494))
* resource/aws_emr_cluster: Support `st1` type EBS volumes ([#5534](https://github.com/terraform-providers/terraform-provider-aws/issues/5534))
* resource/aws_neptune_event_subscription: Support resource import ([#5491](https://github.com/terraform-providers/terraform-provider-aws/issues/5491))
* resource/aws_rds_cluster: Add `engine_mode` argument (support RDS Aurora Serverless) ([#5507](https://github.com/terraform-providers/terraform-provider-aws/issues/5507))
* resource/aws_rds_cluster: Allow `aurora` (MySQL 5.6) `engine_type` to enable Performance Insights ([#5468](https://github.com/terraform-providers/terraform-provider-aws/issues/5468))
* resource/aws_secretsmanager_secret_version: Add `arn` attribute ([#5488](https://github.com/terraform-providers/terraform-provider-aws/issues/5488))
* resource/aws_subnet: Add `arn` attribute ([#5486](https://github.com/terraform-providers/terraform-provider-aws/issues/5486))

BUG FIXES:

* storagegateway: Retry API calls on busy gateway proxy connection errors ([#5476](https://github.com/terraform-providers/terraform-provider-aws/issues/5476))
* resource/aws_cloudtrail: Increase IAM retry threshold from 15 seconds to 1 minute ([#5499](https://github.com/terraform-providers/terraform-provider-aws/issues/5499))
* resource/aws_cognito_user_pool: Properly pass all attributes during update (prevent perpetual flip-flop apply) ([#3458](https://github.com/terraform-providers/terraform-provider-aws/issues/3458))
* resource/aws_cognito_user_pool_client: Properly pass all attributes during update (prevent perpetual flip-flop apply) ([#5478](https://github.com/terraform-providers/terraform-provider-aws/issues/5478))
* resource/aws_db_instance: During S3 restore, lower retry threshold for IAM eventual consistency from 5 minutes to 2 minutes and retry on additional error ([#5536](https://github.com/terraform-providers/terraform-provider-aws/issues/5536))
* resource/aws_dynamodb_table: Allow simultaneous region deletion retry of 5 minutes to better handle global table deletions ([#5518](https://github.com/terraform-providers/terraform-provider-aws/issues/5518))
* resource/aws_glue_crawler: Additional IAM eventual consistency retry logic for create and update ([#5502](https://github.com/terraform-providers/terraform-provider-aws/issues/5502))
* resource/aws_iam_role: Remove extraneous `DeleteRolePermissionsBoundary` API call when deleting IAM role ([#5544](https://github.com/terraform-providers/terraform-provider-aws/issues/5544))
* resource/aws_kinesis_firehose_delivery_stream: Retry on additional IAM eventual consistency error with ElasticSearch destinations ([#5541](https://github.com/terraform-providers/terraform-provider-aws/issues/5541))
* resource/aws_storagegateway_cache: Prevent resource recreation due to disk identifier changes after creation ([#5476](https://github.com/terraform-providers/terraform-provider-aws/issues/5476))

## 1.31.0 (August 09, 2018)

FEATURES:

* **New Data Source:** `aws_db_cluster_snapshot` ([#4526](https://github.com/terraform-providers/terraform-provider-aws/issues/4526))
* **New Resource:** `aws_db_cluster_snapshot` ([#4526](https://github.com/terraform-providers/terraform-provider-aws/issues/4526))
* **New Resource:** `aws_neptune_event_subscription` ([#5480](https://github.com/terraform-providers/terraform-provider-aws/issues/5480))
* **New Resource:** `aws_storagegateway_cache` ([#5282](https://github.com/terraform-providers/terraform-provider-aws/issues/5282))
* **New Resource:** `aws_storagegateway_smb_file_share` ([#5276](https://github.com/terraform-providers/terraform-provider-aws/issues/5276))

ENHANCEMENTS:

* provider: Allow provider configuration AssumeRoleARN and sts:GetCallerIdentity credential validation call to shortcut account ID and partition lookup ([#5177](https://github.com/terraform-providers/terraform-provider-aws/issues/5177))
* provider: Improved output for multiple error handler ([#5442](https://github.com/terraform-providers/terraform-provider-aws/issues/5442))
* data-source/aws_instance: Add `arn` attribute ([#5432](https://github.com/terraform-providers/terraform-provider-aws/issues/5432))
* resource/aws_elasticsearch_domain: Support `ES_APPLICATION_LOGS` `log_type` in plan-time validation ([#5474](https://github.com/terraform-providers/terraform-provider-aws/issues/5474))
* resource/aws_instance: Add `arn` attribute ([#5432](https://github.com/terraform-providers/terraform-provider-aws/issues/5432))
* resource/aws_storagegateway_gateway: Add `smb_active_directory_settings` and `smb_guest_password` arguments ([#5269](https://github.com/terraform-providers/terraform-provider-aws/issues/5269))

BUG FIXES:

* provider: Prefer `USERPROFILE` over `HOMEPATH` for home directory expansion on Windows ([#5443](https://github.com/terraform-providers/terraform-provider-aws/issues/5443))
* resource/aws_ami_copy: Prevent `ena_support` attribute incorrectly reporting force new resource ([#5433](https://github.com/terraform-providers/terraform-provider-aws/issues/5433))
* resource/aws_ami_from_instance: Prevent `ena_support` attribute incorrectly reporting force new resource ([#5433](https://github.com/terraform-providers/terraform-provider-aws/issues/5433))
* resource/aws_elasticsearch_domain: Prevent crash when missing `AutomatedSnapshotStartHour` in API response ([#5451](https://github.com/terraform-providers/terraform-provider-aws/issues/5451))
* resource/aws_elasticsearch_domain: Suppress plan differences for `dedicated_master_count` and `dedicated_master_type` when `dedicated_master_enabled` is disabled ([#5423](https://github.com/terraform-providers/terraform-provider-aws/issues/5423))
* resource/aws_rds_cluster: Prevent error when restoring cluster from snapshot with tagging enabled ([#5479](https://github.com/terraform-providers/terraform-provider-aws/issues/5479))
* resource/aws_ssm_maintenance_window: Properly recreate resource when deleted outside Terraform ([#5416](https://github.com/terraform-providers/terraform-provider-aws/issues/5416))
* resource/aws_ssm_patch_baseline: Properly recreate resource when deleted outside Terraform ([#5438](https://github.com/terraform-providers/terraform-provider-aws/issues/5438))
* resource/aws_vpn_gateway: Allow legacy `amazon_side_asn` in plan-time validation (ASNs 10124 and 17493) ([#5441](https://github.com/terraform-providers/terraform-provider-aws/issues/5441))

## 1.30.0 (August 02, 2018)

FEATURES:

* **New Data Source:** `aws_storagegateway_local_disk` ([#5279](https://github.com/terraform-providers/terraform-provider-aws/issues/5279))
* **New Resource:** `aws_macie_member_account_association` ([#5283](https://github.com/terraform-providers/terraform-provider-aws/issues/5283))
* **New Resource:** `aws_neptune_cluster_instance` ([#5376](https://github.com/terraform-providers/terraform-provider-aws/issues/5376))
* **New Resource:** `aws_storagegateway_nfs_file_share` ([#5255](https://github.com/terraform-providers/terraform-provider-aws/issues/5255))
* **New Resource:** `aws_storagegateway_upload_buffer` ([#5284](https://github.com/terraform-providers/terraform-provider-aws/issues/5284))
* **New Resource:** `aws_storagegateway_working_storage` ([#5285](https://github.com/terraform-providers/terraform-provider-aws/issues/5285))

ENHANCEMENTS:

* data-source/aws_rds_cluster: Add `arn` attribute ([#5221](https://github.com/terraform-providers/terraform-provider-aws/issues/5221))
* resource/aws_ami: Add `ena_support` argument ([#5395](https://github.com/terraform-providers/terraform-provider-aws/issues/5395))
* resource/aws_api_gateway_domain_name: Support resource import ([#5368](https://github.com/terraform-providers/terraform-provider-aws/issues/5368))
* resource/aws_efs_file_system: Add `provisioned_throughput_in_mibps` and `throughput_mode` arguments ([#5210](https://github.com/terraform-providers/terraform-provider-aws/issues/5210))
* resource/aws_elasticsearch_domain: Add `cognito_options` arguments (support Cognito authentication) ([#5346](https://github.com/terraform-providers/terraform-provider-aws/issues/5346))
* resource/aws_glue_crawler: Add `dynamodb_target` argument ([#5152](https://github.com/terraform-providers/terraform-provider-aws/issues/5152))
* resource/aws_iam_role: Add `permissions_boundary` argument ([#5184](https://github.com/terraform-providers/terraform-provider-aws/issues/5184))
* resource/aws_iam_user: Add `permissions_boundary` argument ([#5183](https://github.com/terraform-providers/terraform-provider-aws/issues/5183))
* resource/aws_neptune_cluster: Support resource import ([#5227](https://github.com/terraform-providers/terraform-provider-aws/issues/5227))
* resource/aws_rds_cluster: Add `arn` attribute ([#5221](https://github.com/terraform-providers/terraform-provider-aws/issues/5221))
* resource/aws_ssm_patch_baseline: Add `AMAZON_LINUX_2` and `SUSE` to `operating_system` plan time validation ([#5371](https://github.com/terraform-providers/terraform-provider-aws/issues/5371))

BUG FIXES:

* resource/aws_codebuild_project: Handle additional IAM retry condition during update ([#5238](https://github.com/terraform-providers/terraform-provider-aws/issues/5238))
* resource/aws_codebuild_project: Remove extraneous UpdateProject API call after CreateProject API call ([#5238](https://github.com/terraform-providers/terraform-provider-aws/issues/5238))
* resource/aws_db_instance: Prevent error when restoring database from snapshot with tagging enabled ([#5370](https://github.com/terraform-providers/terraform-provider-aws/issues/5370))
* resource/aws_db_option_group: Prevent error when creating options with new IAM role ([#5389](https://github.com/terraform-providers/terraform-provider-aws/issues/5389))
* resource/aws_eip: Properly handle if multiple EIPs are returned during API read ([#5331](https://github.com/terraform-providers/terraform-provider-aws/issues/5331))
* resource/aws_emr_cluster: Add `configurations_json` argument (handles drift detection as compared to `configurations` argument) ([#5191](https://github.com/terraform-providers/terraform-provider-aws/issues/5191))
* resource/aws_emr_cluster: Ensure `keep_job_flow_alive_when_no_step = false` automatically terminates cluster ([#5415](https://github.com/terraform-providers/terraform-provider-aws/issues/5415))
* resource/aws_lambda_event_source_mapping: Properly read `enabled` into Terraform state ([#5292](https://github.com/terraform-providers/terraform-provider-aws/issues/5292))
* resource/aws_launch_template: Exclude `network_interfaces` `associate_public_ip_address` when conflicting `network_interface_id` is set ([#5314](https://github.com/terraform-providers/terraform-provider-aws/issues/5314))
* resource/aws_launch_template: Set `latest_version` as re-computed on updates (prevent need for double apply) ([#5250](https://github.com/terraform-providers/terraform-provider-aws/issues/5250))
* resource/aws_lb_listener: Prevent crash from new `fixed-response` and `redirect` actions ([#5367](https://github.com/terraform-providers/terraform-provider-aws/issues/5367))
* resource/aws_lb_listener_rule: Prevent crash from new `fixed-response` and `redirect` actions ([#5367](https://github.com/terraform-providers/terraform-provider-aws/issues/5367))
* resource/aws_vpn_gateway: Allow legacy `amazon_side_asn` in plan-time validation (ASNs 7224 and 9059) ([#5291](https://github.com/terraform-providers/terraform-provider-aws/issues/5291))
* resource/aws_waf_web_acl: Properly read `rules` into Terraform state ([#5342](https://github.com/terraform-providers/terraform-provider-aws/issues/5342))
* resource/aws_waf_web_acl: Properly update `rules` ([#5380](https://github.com/terraform-providers/terraform-provider-aws/issues/5380))
* resource/aws_wafregional_rate_based_rule: Fix `rate_limit` updates ([#5356](https://github.com/terraform-providers/terraform-provider-aws/issues/5356))
* resource/aws_wafregional_web_acl: Properly read `rules` into Terraform state ([#5342](https://github.com/terraform-providers/terraform-provider-aws/issues/5342))

## 1.29.0 (July 26, 2018)

NOTES:

* data-source/aws_kms_secret: This data source has been deprecated and will be removed in the next major version. This is required to support the upcoming Terraform 0.12. A new `aws_kms_secrets` data source is available that allows for the same multiple KMS secret decryption functionality, but requires different attribute references. Full migration information is available in the [AWS Provider Version 2 Upgrade Guide](https://www.terraform.io/docs/providers/aws/guides/version-2-upgrade.html#data-source-aws_kms_secret).

FEATURES:

* **New Data Source:** `aws_kms_secrets` ([#5195](https://github.com/terraform-providers/terraform-provider-aws/issues/5195))
* **New Data Source:** `aws_network_interfaces` ([#5324](https://github.com/terraform-providers/terraform-provider-aws/issues/5324))
* **New Guide:** [`AWS Provider Version 2 Upgrade`](https://www.terraform.io/docs/providers/aws/guides/version-2-upgrade.html) ([#5195](https://github.com/terraform-providers/terraform-provider-aws/issues/5195))

ENHANCEMENTS:

* data-source/aws_iam_role: Add `permissions_boundary` attribute ([#5186](https://github.com/terraform-providers/terraform-provider-aws/issues/5186))
* data-source/aws_vpc: Add `arn` attribute ([#5300](https://github.com/terraform-providers/terraform-provider-aws/issues/5300))
* resource/aws_default_vpc: Add `arn` attribute ([#5300](https://github.com/terraform-providers/terraform-provider-aws/issues/5300))
* resource/aws_instance: Add `cpu_core_count` and `cpu_threads_per_core` arguments ([#5159](https://github.com/terraform-providers/terraform-provider-aws/issues/5159))
* resource/aws_lambda_permission: Add `event_source_token` argument (support Alexa Skills) ([#5264](https://github.com/terraform-providers/terraform-provider-aws/issues/5264))
* resource/aws_launch_template: Add `arn` attribute ([#5306](https://github.com/terraform-providers/terraform-provider-aws/issues/5306))
* resource/aws_secretsmanager_secret: Add `policy` argument ([#5290](https://github.com/terraform-providers/terraform-provider-aws/issues/5290))
* resource/aws_vpc: Add `arn` attribute ([#5300](https://github.com/terraform-providers/terraform-provider-aws/issues/5300))
* resource/aws_waf_web_acl: Support resource import ([#5337](https://github.com/terraform-providers/terraform-provider-aws/issues/5337))

BUG FIXES:

* data-source/aws_vpc_endpoint_service: Perform client side filtering to workaround server side filtering issues in AWS China and AWS GovCloud (US) ([#4592](https://github.com/terraform-providers/terraform-provider-aws/issues/4592))
* resource/aws_kinesis_firehose_delivery_stream: Force new resource for `kinesis_source_configuration` argument changes ([#5332](https://github.com/terraform-providers/terraform-provider-aws/issues/5332))
* resource/aws_route53_record: Prevent DomainLabelEmpty errors when expanding record names with trailing period ([#5312](https://github.com/terraform-providers/terraform-provider-aws/issues/5312))
* resource/aws_ses_identity_notification_topic: Prevent panic when API returns no attributes ([#5327](https://github.com/terraform-providers/terraform-provider-aws/issues/5327))
* resource/aws_ssm_parameter: Reduce DescribeParameters API calls by switching filtering logic ([#5325](https://github.com/terraform-providers/terraform-provider-aws/issues/5325))

## 1.28.0 (July 18, 2018)

FEATURES:

* **New Resource:** `aws_macie_s3_bucket_association` ([#5201](https://github.com/terraform-providers/terraform-provider-aws/issues/5201))
* **New Resource:** `aws_neptune_cluster` ([#5050](https://github.com/terraform-providers/terraform-provider-aws/issues/5050))
* **New Resource:** `aws_storagegateway_gateway` ([#5208](https://github.com/terraform-providers/terraform-provider-aws/issues/5208))

ENHANCEMENTS:

* data-source/aws_iam_user: Add `permissions_boundary` attribute ([#5187](https://github.com/terraform-providers/terraform-provider-aws/issues/5187))
* resource/aws_api_gateway_integration: Add `timeout_milliseconds` argument ([#5199](https://github.com/terraform-providers/terraform-provider-aws/issues/5199))
* resource/aws_cloudwatch_log_group: Allow `tags` handling in AWS GovCloud (US) and AWS China ([#5175](https://github.com/terraform-providers/terraform-provider-aws/issues/5175))
* resource/aws_codebuild_project: Add `report_build_status` argument under `source` (support report build status for GitHub source type) ([#5156](https://github.com/terraform-providers/terraform-provider-aws/issues/5156))
* resource/aws_launch_template: Ignore `credit_specification` when not using T2 `instance_type` ([#5190](https://github.com/terraform-providers/terraform-provider-aws/issues/5190))
* resource/aws_rds_cluster_instance: Add `arn` attribute ([#5220](https://github.com/terraform-providers/terraform-provider-aws/issues/5220))
* resource/aws_route: Print more useful error message when missing valid target type ([#5198](https://github.com/terraform-providers/terraform-provider-aws/issues/5198))
* resource/aws_vpc_endpoint: Add configurable timeouts ([#3418](https://github.com/terraform-providers/terraform-provider-aws/issues/3418))
* resource/aws_vpc_endpoint_subnet_association: Add configurable timeouts ([#3418](https://github.com/terraform-providers/terraform-provider-aws/issues/3418))

BUG FIXES:

* resource/aws_glue_crawler: Prevent error when deleted outside Terraform ([#5158](https://github.com/terraform-providers/terraform-provider-aws/issues/5158))
* resource/aws_vpc_endpoint_subnet_association: Add mutex to prevent errors with concurrent `ModifyVpcEndpoint` calls ([#3418](https://github.com/terraform-providers/terraform-provider-aws/issues/3418))

## 1.27.0 (July 11, 2018)

NOTES:

* resource/aws_codebuild_project: The `service_role` argument is now required to match the API behavior and provide plan time validation. Additional details from AWS Support can be found in: https://github.com/terraform-providers/terraform-provider-aws/pull/4826
* resource/aws_wafregional_byte_match_set: The `byte_match_tuple` argument name has been deprecated in preference of a new  `byte_match_tuples` argument name, for consistency with the `aws_waf_byte_match_set` resource to reduce any confusion working between the two resources and to denote its multiple value support. Its behavior is exactly the same as the old argument. Simply changing the argument name (adding the `s`) to configurations should upgrade without other changes.

FEATURES:

* **New Resource:** `aws_appsync_api_key` ([#3827](https://github.com/terraform-providers/terraform-provider-aws/issues/3827))
* **New Resource:** `aws_swf_domain` ([#2803](https://github.com/terraform-providers/terraform-provider-aws/issues/2803))

ENHANCEMENTS:

* data-source/aws_region: Add `description` attribute ([#5077](https://github.com/terraform-providers/terraform-provider-aws/issues/5077))
* data-source/aws_vpc: Add `cidr_block_associations` attribute ([#5098](https://github.com/terraform-providers/terraform-provider-aws/issues/5098))
* resource/aws_cloudwatch_metric_alarm: Add `datapoints_to_alarm` and `evaluation_period` plan time validation ([#5095](https://github.com/terraform-providers/terraform-provider-aws/issues/5095))
* resource/aws_db_parameter_group: Clarify naming validation error messages ([#5090](https://github.com/terraform-providers/terraform-provider-aws/issues/5090))
* resource/aws_glue_connection: Add `physical_connection_requirements` argument `availability_zone` (currently required by the API) ([#5039](https://github.com/terraform-providers/terraform-provider-aws/issues/5039))
* resource/aws_instance: Ignore `credit_specifications` when not using T2 `instance_type` ([#5114](https://github.com/terraform-providers/terraform-provider-aws/issues/5114))
* resource/aws_instance: Allow AWS GovCloud (US) to perform tagging on creation ([#5106](https://github.com/terraform-providers/terraform-provider-aws/issues/5106))
* resource/aws_lambda_function: Support `dotnetcore2.1` in `runtime` validation ([#5150](https://github.com/terraform-providers/terraform-provider-aws/issues/5150))
* resource/aws_route_table: Ignore propagated routes during resource import ([#5100](https://github.com/terraform-providers/terraform-provider-aws/issues/5100))
* resource/aws_security_group: Authorize and revoke only changed individual `ingress`/`egress` rules despite their configuration grouping (e.g. replacing an individual element in a multiple element `cidr_blocks` list) ([#4726](https://github.com/terraform-providers/terraform-provider-aws/issues/4726))
* resource/aws_ses_receipt_rule: Add plan time validation for `s3_action` argument `position` ([#5092](https://github.com/terraform-providers/terraform-provider-aws/issues/5092))
* resource/aws_vpc_ipv4_cidr_block_association: Support resource import ([#5069](https://github.com/terraform-providers/terraform-provider-aws/issues/5069))
* resource/aws_waf_web_acl: Add `rules` `override_action` argument and support `GROUP` type ([#5053](https://github.com/terraform-providers/terraform-provider-aws/issues/5053))
* resource/aws_wafregional_web_acl: Add `rules` `override_action` argument and support `GROUP` type ([#5053](https://github.com/terraform-providers/terraform-provider-aws/issues/5053))

BUG FIXES:

* resource/aws_codebuild_project: Prevent panic when empty `vpc_config` block is configured ([#5070](https://github.com/terraform-providers/terraform-provider-aws/issues/5070))
* resource/aws_codebuild_project: Mark `service_role` as required ([#4826](https://github.com/terraform-providers/terraform-provider-aws/issues/4826))
* resource/aws_glue_catalog_database: Properly return error when missing colon during import ([#5123](https://github.com/terraform-providers/terraform-provider-aws/issues/5123))
* resource/aws_glue_catalog_database: Prevent error when deleted outside Terraform ([#5141](https://github.com/terraform-providers/terraform-provider-aws/issues/5141))
* resource/aws_instance: Allow AWS China to perform volume tagging post-creation on first apply ([#5106](https://github.com/terraform-providers/terraform-provider-aws/issues/5106))
* resource/aws_kms_grant: Properly return error when listing KMS grants ([#5063](https://github.com/terraform-providers/terraform-provider-aws/issues/5063))
* resource/aws_rds_cluster_instance: Support `configuring-log-exports` status ([#5124](https://github.com/terraform-providers/terraform-provider-aws/issues/5124))
* resource/aws_s3_bucket: Prevent extraneous ACL update during resource creation ([#5107](https://github.com/terraform-providers/terraform-provider-aws/issues/5107))
* resource/aws_wafregional_byte_match_set: Deprecate `byte_match_tuple` argument for `byte_match_tuples` ([#5043](https://github.com/terraform-providers/terraform-provider-aws/issues/5043))

## 1.26.0 (July 04, 2018)

FEATURES:

* **New Data Source:** `aws_launch_configuration` ([#3624](https://github.com/terraform-providers/terraform-provider-aws/issues/3624))
* **New Data Source:** `aws_pricing_product` ([#5057](https://github.com/terraform-providers/terraform-provider-aws/issues/5057))
* **New Resource:** `aws_s3_bucket_inventory` ([#5019](https://github.com/terraform-providers/terraform-provider-aws/issues/5019))
* **New Resource:** `aws_vpc_ipv4_cidr_block_association` ([#3723](https://github.com/terraform-providers/terraform-provider-aws/issues/3723))

ENHANCEMENTS:

* data-source/aws_elasticache_replication_group: Add `member_clusters` attribute ([#5056](https://github.com/terraform-providers/terraform-provider-aws/issues/5056))
* data-source/aws_instances: Add `instance_state_names` argument (support non-`running` instances) ([#4950](https://github.com/terraform-providers/terraform-provider-aws/issues/4950))
* data-source/aws_route_tables: Add `filter` argument ([#5035](https://github.com/terraform-providers/terraform-provider-aws/issues/5035))
* data-source/aws_subnet_ids: Add `filter` argument ([#5038](https://github.com/terraform-providers/terraform-provider-aws/issues/5038))
* resource/aws_eip_association: Support resource import ([#5006](https://github.com/terraform-providers/terraform-provider-aws/issues/5006))
* resource/aws_elasticache_replication_group: Add `member_clusters` attribute ([#5056](https://github.com/terraform-providers/terraform-provider-aws/issues/5056))
* resource/aws_lambda_alias: Add `routing_config` argument (support traffic shifting) ([#3316](https://github.com/terraform-providers/terraform-provider-aws/issues/3316))
* resource/aws_lambda_event_source_mapping: Make `starting_position` optional and allow `batch_size` to support default of 10 for SQS ([#5024](https://github.com/terraform-providers/terraform-provider-aws/issues/5024))
* resource/aws_network_acl_rule: Add plan time conflict validation with `cidr_block` and `ipv6_cidr_block` ([#3951](https://github.com/terraform-providers/terraform-provider-aws/issues/3951))
* resource/aws_spot_fleet_request: Add `fleet_type` argument ([#5032](https://github.com/terraform-providers/terraform-provider-aws/issues/5032))
* resource/aws_ssm_document: Add `tags` argument (support tagging) ([#5020](https://github.com/terraform-providers/terraform-provider-aws/issues/5020))

BUG FIXES:

* resource/aws_codebuild_project: Prevent panic with missing environment variable type ([#5052](https://github.com/terraform-providers/terraform-provider-aws/issues/5052))
* resource/aws_kms_alias: Fix perpetual plan when `target_key_id` is ARN ([#4010](https://github.com/terraform-providers/terraform-provider-aws/issues/4010))

## 1.25.0 (June 27, 2018)

NOTES:

* resource/aws_instance: Starting around June 21, 2018, the EC2 API began responding with an empty string value for user data for some instances instead of a completely empty response. In Terraform, it would show as a difference of `user_data: "da39a3ee5e6b4b0d3255bfef95601890afd80709" => "" (forces new resource)` if the `user_data` argument was not defined in the Terraform configuration for the resource. This release ignores that difference as equivalent.

FEATURES:

* **New Data Source:** `aws_codecommit_repository` ([#4934](https://github.com/terraform-providers/terraform-provider-aws/issues/4934))
* **New Data Source:** `aws_dx_gateway` ([#4988](https://github.com/terraform-providers/terraform-provider-aws/issues/4988))
* **New Data Source:** `aws_network_acls` ([#4966](https://github.com/terraform-providers/terraform-provider-aws/issues/4966))
* **New Data Source:** `aws_route_tables` ([#4841](https://github.com/terraform-providers/terraform-provider-aws/issues/4841))
* **New Data Source:** `aws_security_groups` ([#2947](https://github.com/terraform-providers/terraform-provider-aws/issues/2947))
* **New Resource:** `aws_dx_hosted_private_virtual_interface` ([#3255](https://github.com/terraform-providers/terraform-provider-aws/issues/3255))
* **New Resource:** `aws_dx_hosted_private_virtual_interface_accepter` ([#3255](https://github.com/terraform-providers/terraform-provider-aws/issues/3255))
* **New Resource:** `aws_dx_hosted_public_virtual_interface` ([#3254](https://github.com/terraform-providers/terraform-provider-aws/issues/3254))
* **New Resource:** `aws_dx_hosted_public_virtual_interface_accepter` ([#3254](https://github.com/terraform-providers/terraform-provider-aws/issues/3254))
* **New Resource:** `aws_dx_private_virtual_interface` ([#3253](https://github.com/terraform-providers/terraform-provider-aws/issues/3253))
* **New Resource:** `aws_dx_public_virtual_interface` ([#3252](https://github.com/terraform-providers/terraform-provider-aws/issues/3252))
* **New Resource:** `aws_media_store_container_policy` ([#3507](https://github.com/terraform-providers/terraform-provider-aws/issues/3507))

ENHANCEMENTS:

* provider: Support custom endpoint for `autoscaling` ([#4970](https://github.com/terraform-providers/terraform-provider-aws/issues/4970))
* resource/aws_codebuild_project: Support `WINDOWS_CONTAINER` as valid environment type ([#4960](https://github.com/terraform-providers/terraform-provider-aws/issues/4960))
* resource/aws_codebuild_project: Support resource import ([#4976](https://github.com/terraform-providers/terraform-provider-aws/issues/4976))
* resource/aws_ecs_service: Add `scheduling_strategy` argument (support `DAEMON` scheduling strategy) ([#4825](https://github.com/terraform-providers/terraform-provider-aws/issues/4825))
* resource/aws_iam_instance_profile: Add `create_date` attribute ([#4932](https://github.com/terraform-providers/terraform-provider-aws/issues/4932))
* resource/aws_media_store_container: Support resource import ([#3501](https://github.com/terraform-providers/terraform-provider-aws/issues/3501))
* resource/aws_network_acl: Add full mapping of protocol names to protocol numbers ([#4956](https://github.com/terraform-providers/terraform-provider-aws/issues/4956))
* resource/aws_network_acl_rule: Add full mapping of protocol names to protocol numbers ([#4956](https://github.com/terraform-providers/terraform-provider-aws/issues/4956))
* resource/aws_sqs_queue: Add .fifo suffix for FIFO queues using `name_prefix` ([#4929](https://github.com/terraform-providers/terraform-provider-aws/issues/4929))
* resource/aws_vpc: Support update of `instance_tenancy` from `dedicated` to `default` ([#2514](https://github.com/terraform-providers/terraform-provider-aws/issues/2514))
* resource/aws_waf_ipset: Support resource import ([#4979](https://github.com/terraform-providers/terraform-provider-aws/issues/4979))
* resource/aws_wafregional_web_acl: Add rule `type` argument (support rate limited rules) ([#4307](https://github.com/terraform-providers/terraform-provider-aws/issues/4307)] / [[#4978](https://github.com/terraform-providers/terraform-provider-aws/issues/4978))

BUG FIXES:

* data-source/aws_rds_cluster: Prevent panic with new CloudWatch logs support (`enabled_cloudwatch_logs_exports`) introduced in 1.23.0 ([#4927](https://github.com/terraform-providers/terraform-provider-aws/issues/4927))
* resource/aws_codebuild_webhook: Prevent panic when webhook is missing during read ([#4917](https://github.com/terraform-providers/terraform-provider-aws/issues/4917))
* resource/aws_db_instance: Properly raise any `ListTagsForResource` error instead of presenting a perpetual difference with `tags` ([#4943](https://github.com/terraform-providers/terraform-provider-aws/issues/4943))
* resource/aws_instance: Prevent extraneous ModifyInstanceAttribute call for `disable_api_termination` on resource creation ([#4941](https://github.com/terraform-providers/terraform-provider-aws/issues/4941))
* resource/aws_instance: Ignore empty string SHA (`da39a3ee5e6b4b0d3255bfef95601890afd80709`) `user_data` difference due to EC2 API response changes ([#4991](https://github.com/terraform-providers/terraform-provider-aws/issues/4991))
* resource/aws_launch_template: Prevent error when using `valid_until` ([#4952](https://github.com/terraform-providers/terraform-provider-aws/issues/4952))
* resource/aws_route: Properly force resource recreation when updating `route_table_id` ([#4946](https://github.com/terraform-providers/terraform-provider-aws/issues/4946))
* resource/aws_route53_zone: Further prevent HostedZoneAlreadyExists with specified caller reference errors ([#4903](https://github.com/terraform-providers/terraform-provider-aws/issues/4903))
* resource/aws_ses_receipt_rule: Prevent error with `s3_action` when `kms_key_arn` is not specified ([#4965](https://github.com/terraform-providers/terraform-provider-aws/issues/4965))

## 1.24.0 (June 21, 2018)

FEATURES:

* **New Data Source:** `aws_cloudformation_export` ([#2180](https://github.com/terraform-providers/terraform-provider-aws/issues/2180))
* **New Data Source:** `aws_vpc_dhcp_options` ([#4878](https://github.com/terraform-providers/terraform-provider-aws/issues/4878))
* **New Resource:** `aws_dx_gateway` ([#4896](https://github.com/terraform-providers/terraform-provider-aws/issues/4896))
* **New Resource:** `aws_dx_gateway_association` ([#4896](https://github.com/terraform-providers/terraform-provider-aws/issues/4896))
* **New Resource:** `aws_glue_crawler` ([#4484](https://github.com/terraform-providers/terraform-provider-aws/issues/4484))
* **New Resource:** `aws_neptune_cluster_parameter_group` ([#4860](https://github.com/terraform-providers/terraform-provider-aws/issues/4860))
* **New Resource:** `aws_neptune_subnet_group` ([#4782](https://github.com/terraform-providers/terraform-provider-aws/issues/4782))

ENHANCEMENTS:

* resource/aws_api_gateway_rest_api: Support `PRIVATE` endpoint type ([#4888](https://github.com/terraform-providers/terraform-provider-aws/issues/4888))
* resource/aws_codedeploy_app: Add `compute_platform` argument ([#4811](https://github.com/terraform-providers/terraform-provider-aws/issues/4811))
* resource/aws_kinesis_firehose_delivery_stream: Support extended S3 destination `data_format_conversion_configuration` ([#4842](https://github.com/terraform-providers/terraform-provider-aws/issues/4842))
* resource/aws_kms_grant: Support ARN for `key_id` argument (external CMKs) ([#4886](https://github.com/terraform-providers/terraform-provider-aws/issues/4886))
* resource/aws_neptune_parameter_group: Add `tags` argument and `arn` attribute ([#4873](https://github.com/terraform-providers/terraform-provider-aws/issues/4873))
* resource/aws_rds_cluster: Add `enabled_cloudwatch_logs_exports` argument ([#4875](https://github.com/terraform-providers/terraform-provider-aws/issues/4875))

BUG FIXES:

* resource/aws_batch_job_definition: Force resource recreation on retry_strategy attempts updates ([#4854](https://github.com/terraform-providers/terraform-provider-aws/issues/4854))
* resource/aws_cognito_user_pool_client: Prevent panic with updating `refresh_token_validity` ([#4868](https://github.com/terraform-providers/terraform-provider-aws/issues/4868))
* resource/aws_instance: Prevent extraneous ModifyInstanceCreditSpecification call on resource creation ([#4898](https://github.com/terraform-providers/terraform-provider-aws/issues/4898))
* resource/aws_s3_bucket: Properly detect `cors_rule` drift when it is deleted outside Terraform ([#4887](https://github.com/terraform-providers/terraform-provider-aws/issues/4887))
* resource/aws_vpn_gateway_attachment: Fix error handling for missing VPN gateway ([#4895](https://github.com/terraform-providers/terraform-provider-aws/issues/4895))

## 1.23.0 (June 14, 2018)

NOTES:

* resource/aws_elasticache_cluster: The `availability_zones` argument has been deprecated in favor of a new `preferred_availability_zones` argument to allow specifying the same Availability Zone more than once in larger Memcached clusters that also need to specifically set Availability Zones. The argument is still optional and the API will continue to automatically choose Availability Zones for nodes if not specified. The new argument will also continue to match the APIs required behavior that the length of the list must be the same as `num_cache_nodes`. Migration will require recreating the resource or using the resource [lifecycle configuration](https://www.terraform.io/docs/configuration/resources.html#lifecycle) of `ignore_changes = ["availability_zones"]` to prevent recreation. See the resource documentation for additional details.

FEATURES:

* **New Data Source:** `aws_vpcs` ([#4736](https://github.com/terraform-providers/terraform-provider-aws/issues/4736))
* **New Resource:** `aws_neptune_parameter_group` ([#4724](https://github.com/terraform-providers/terraform-provider-aws/issues/4724))

ENHANCEMENTS:

* resource/aws_db_instance: Display input arguments when receiving InvalidParameterValue error on resource creation ([#4803](https://github.com/terraform-providers/terraform-provider-aws/issues/4803))
* resource/aws_elasticache_cluster: Migrate from `availability_zones` TypeSet attribute to `preferred_availability_zones` TypeList attribute (allow duplicate Availability Zone elements) ([#4741](https://github.com/terraform-providers/terraform-provider-aws/issues/4741))
* resource/aws_launch_template: Add `tags` argument (support tagging the resource itself) ([#4763](https://github.com/terraform-providers/terraform-provider-aws/issues/4763))
* resource/aws_launch_template: Add plan time validation for tag_specifications `resource_type` ([#4765](https://github.com/terraform-providers/terraform-provider-aws/issues/4765))
* resource/aws_waf_ipset: Add `arn` attribute ([#4784](https://github.com/terraform-providers/terraform-provider-aws/issues/4784))
* resource/aws_wafregional_ipset: Add `arn` attribute ([#4816](https://github.com/terraform-providers/terraform-provider-aws/issues/4816))

BUG FIXES:

* resource/aws_codebuild_webhook: Properly export `secret` (the CodeBuild API only provides its value during resource creation) ([#4775](https://github.com/terraform-providers/terraform-provider-aws/issues/4775))
* resource/aws_codecommit_repository: Prevent error and trigger recreation when not found during read ([#4761](https://github.com/terraform-providers/terraform-provider-aws/issues/4761))
* resource/aws_eks_cluster: Properly export `arn` attribute ([#4766](https://github.com/terraform-providers/terraform-provider-aws/issues/4766)] / [[#4767](https://github.com/terraform-providers/terraform-provider-aws/issues/4767))
* resource/aws_elasticsearch_domain: Skip EBS options update/refresh if EBS is not enabled ([#4802](https://github.com/terraform-providers/terraform-provider-aws/issues/4802))

## 1.22.0 (June 05, 2018)

FEATURES:

* **New Data Source:** `aws_ecs_service` ([#3617](https://github.com/terraform-providers/terraform-provider-aws/issues/3617))
* **New Data Source:** `aws_eks_cluster` ([#4749](https://github.com/terraform-providers/terraform-provider-aws/issues/4749))
* **New Guide:** EKS Getting Started
* **New Resource:** `aws_config_aggregate_authorization` ([#4263](https://github.com/terraform-providers/terraform-provider-aws/issues/4263))
* **New Resource:** `aws_config_configuration_aggregator` ([#4262](https://github.com/terraform-providers/terraform-provider-aws/issues/4262))
* **New Resource:** `aws_eks_cluster` ([#4749](https://github.com/terraform-providers/terraform-provider-aws/issues/4749))

ENHANCEMENTS:

* provider: Support custom endpoint for EFS ([#4716](https://github.com/terraform-providers/terraform-provider-aws/issues/4716))
* resource/aws_api_gateway_method: Add `authorization_scopes` argument ([#4533](https://github.com/terraform-providers/terraform-provider-aws/issues/4533))
* resource/aws_api_gateway_rest_api: Add `api_key_source` argument ([#4717](https://github.com/terraform-providers/terraform-provider-aws/issues/4717))
* resource/aws_cloudfront_distribution: Allow create and update retries on InvalidViewerCertificate for eventual consistency with ACM/IAM services ([#4698](https://github.com/terraform-providers/terraform-provider-aws/issues/4698))
* resource/aws_cognito_identity_pool: Add `arn` attribute ([#4719](https://github.com/terraform-providers/terraform-provider-aws/issues/4719))
* resource/aws_cognito_user_pool: Add `endpoint` attribute ([#4718](https://github.com/terraform-providers/terraform-provider-aws/issues/4718))

BUG FIXES:

* resource/aws_service_discovery_private_dns_namespace: Prevent creation error with names longer than 34 characters ([#4702](https://github.com/terraform-providers/terraform-provider-aws/issues/4702))
* resource/aws_vpn_connection: Allow period in `tunnel[1-2]_preshared_key` validation ([#4731](https://github.com/terraform-providers/terraform-provider-aws/issues/4731))

## 1.21.0 (May 31, 2018)

FEATURES:

* **New Data Source:** `aws_route` ([#4529](https://github.com/terraform-providers/terraform-provider-aws/issues/4529))
* **New Resource:** `aws_codebuild_webhook` ([#4473](https://github.com/terraform-providers/terraform-provider-aws/issues/4473))
* **New Resource:** `aws_cognito_identity_provider` ([#3601](https://github.com/terraform-providers/terraform-provider-aws/issues/3601))
* **New Resource:** `aws_cognito_resource_server` ([#4530](https://github.com/terraform-providers/terraform-provider-aws/issues/4530))
* **New Resource:** `aws_glue_classifier` ([#4472](https://github.com/terraform-providers/terraform-provider-aws/issues/4472))

ENHANCEMENTS:

* provider: Support custom endpoint for SSM ([#4670](https://github.com/terraform-providers/terraform-provider-aws/issues/4670))
* resource/aws_codebuild_project: Add `badge_enabled` argument and `badge_url` attribute ([#3504](https://github.com/terraform-providers/terraform-provider-aws/issues/3504))
* resource/aws_codebuild_project: Add `environment_variable` argument `type` (support parameter store environment variables) ([#2811](https://github.com/terraform-providers/terraform-provider-aws/issues/2811)] / [[#4021](https://github.com/terraform-providers/terraform-provider-aws/issues/4021))
* resource/aws_codebuild_project: Add `source` argument `git_clone_depth` and `insecure_ssl` ([#3929](https://github.com/terraform-providers/terraform-provider-aws/issues/3929))
* resource/aws_elasticache_replication_group: Support `number_cache_nodes` updates ([#4504](https://github.com/terraform-providers/terraform-provider-aws/issues/4504))
* resource/aws_lb_target_group: Add `slow_start` argument ([#4661](https://github.com/terraform-providers/terraform-provider-aws/issues/4661))
* resource/aws_redshift_cluster: Add `dns_name` attribute ([#4582](https://github.com/terraform-providers/terraform-provider-aws/issues/4582))
* resource/aws_s3_bucket: Add `bucket_regional_domain_name` attribute ([#4556](https://github.com/terraform-providers/terraform-provider-aws/issues/4556))

BUG FIXES:

* data-source/aws_lambda_function: Qualifiers explicitly set are now honoured ([#4654](https://github.com/terraform-providers/terraform-provider-aws/issues/4654))
* resource/aws_batch_job_definition: Properly force new resource when updating timeout `attempt_duration_seconds` argument ([#4697](https://github.com/terraform-providers/terraform-provider-aws/issues/4697))
* resource/aws_budgets_budget: Force new resource when updating `name` ([#4656](https://github.com/terraform-providers/terraform-provider-aws/issues/4656))
* resource/aws_dms_endpoint: Additionally specify MongoDB connection info in the top-level API namespace to prevent issues connecting ([#4636](https://github.com/terraform-providers/terraform-provider-aws/issues/4636))
* resource/aws_rds_cluster: Prevent additional retry error during S3 import for IAM/S3 eventual consistency ([#4683](https://github.com/terraform-providers/terraform-provider-aws/issues/4683))
* resource/aws_sns_sms_preferences: Properly add SNS preferences to website docs ([#4694](https://github.com/terraform-providers/terraform-provider-aws/issues/4694))

## 1.20.0 (May 23, 2018)

NOTES:

* resource/aws_guardduty_member: Terraform will now try to properly detect if a member account has been invited based on its relationship status (`Disabled`/`Enabled`/`Invited`) and appropriately flag the new `invite` argument for update. You will want to set `invite = true` in your Terraform configuration if you previously handled the invitation process for a member, otherwise the resource will attempt to disassociate the member upon updating the provider to this version.

FEATURES:

* **New Data Source:** `aws_glue_script` ([#4481](https://github.com/terraform-providers/terraform-provider-aws/issues/4481))
* **New Resource:** `aws_glue_trigger` ([#4464](https://github.com/terraform-providers/terraform-provider-aws/issues/4464))

ENHANCEMENTS:

* resource/aws_api_gateway_domain_name: Add `endpoint_configuration` argument, `regional_certificate_arn` argument, `regional_certificate_name` argument, `regional_domain_name` attribute, and `regional_zone_id` attribute (support regional domain names) ([#2866](https://github.com/terraform-providers/terraform-provider-aws/issues/2866))
* resource/aws_api_gateway_rest_api: Add `endpoint_configuration` argument (support regional endpoint type) ([#2866](https://github.com/terraform-providers/terraform-provider-aws/issues/2866))
* resource/aws_appautoscaling_policy: Add retry logic for rate exceeded errors during read, update and delete ([#4594](https://github.com/terraform-providers/terraform-provider-aws/issues/4594))
* resource/aws_ecs_service: Add `container_name` and `container_port` arguments for `service_registry` (support bridge and host network mode for service registry) ([#4623](https://github.com/terraform-providers/terraform-provider-aws/issues/4623))
* resource/aws_emr_cluster: Add `additional_info` argument ([#4590](https://github.com/terraform-providers/terraform-provider-aws/issues/4590))
* resource/aws_guardduty_member: Support member account invitation on creation ([#4357](https://github.com/terraform-providers/terraform-provider-aws/issues/4357))
* resource/aws_guardduty_member: Support `invite` argument updates (invite or disassociate on update) ([#4604](https://github.com/terraform-providers/terraform-provider-aws/issues/4604))
* resource/aws_ssm_patch_baseline: Add `approval_rule` `enable_non_security` argument ([#4546](https://github.com/terraform-providers/terraform-provider-aws/issues/4546))

BUG FIXES:

* resource/aws_api_gateway_rest_api: Prevent error with `policy` containing special characters (e.g. forward slashes in CIDRs) ([#4606](https://github.com/terraform-providers/terraform-provider-aws/issues/4606))
* resource/aws_cloudwatch_event_rule: Prevent multiple names on creation ([#4579](https://github.com/terraform-providers/terraform-provider-aws/issues/4579))
* resource/aws_dynamodb_table: Prevent error with APIs that do not support point in time recovery (e.g. AWS China) ([#4573](https://github.com/terraform-providers/terraform-provider-aws/issues/4573))
* resource/aws_glue_catalog_table: Prevent multiple potential panic scenarios ([#4621](https://github.com/terraform-providers/terraform-provider-aws/issues/4621))
* resource/aws_kinesis_stream: Handle tag additions/removals of more than 10 tags ([#4574](https://github.com/terraform-providers/terraform-provider-aws/issues/4574))
* resource/aws_kinesis_stream: Prevent perpetual `encryption_type` difference with APIs that do not support encryption (e.g. AWS China) ([#4575](https://github.com/terraform-providers/terraform-provider-aws/issues/4575))
* resource/aws_s3_bucket: Prevent panic from CORS reading errors ([#4603](https://github.com/terraform-providers/terraform-provider-aws/issues/4603))
* resource/aws_spot_fleet_request: Prevent empty `iam_instance_profile_arn` from overwriting `iam_instance_profile` ([#4591](https://github.com/terraform-providers/terraform-provider-aws/issues/4591))

## 1.19.0 (May 16, 2018)

NOTES:

* data-source/aws_iam_policy_document: Please note there is a behavior change in the rendering of `principal`/`not_principal` in the case of `type = "AWS"` and `identifiers = ["*"]`. This will now render as `Principal": {"AWS": "*"}` instead of `"Principal": "*"`. This change is required for IAM role trust policy support as well as differentiating between anonymous access versus AWS access in policies. To keep the old behavior of anonymous access, use `type = "*"` and `identifiers = ["*"]`, which will continue to render as `"Principal": "*"`. For additional information, see the [`aws_iam_policy_document` documentation](https://www.terraform.io/docs/providers/aws/d/iam_policy_document.html).

FEATURES:

* **New Data Source:** `aws_arn` ([#3996](https://github.com/terraform-providers/terraform-provider-aws/issues/3996))
* **New Data Source:** `aws_lambda_invocation` ([#4222](https://github.com/terraform-providers/terraform-provider-aws/issues/4222))
* **New Resource:** `aws_sns_sms_preferences` ([#3858](https://github.com/terraform-providers/terraform-provider-aws/issues/3858))

ENHANCEMENTS:

* data-source/aws_iam_policy_document: Allow rendering of `"Principal": {"AWS": "*"}` (required for IAM role trust policies) ([#4248](https://github.com/terraform-providers/terraform-provider-aws/issues/4248))
* resource/aws_api_gateway_rest_api: Add `execution_arn` attribute ([#3968](https://github.com/terraform-providers/terraform-provider-aws/issues/3968))
* resource/aws_db_event_subscription: Add `name_prefix` argument ([#2754](https://github.com/terraform-providers/terraform-provider-aws/issues/2754))
* resource/aws_dms_endpoint: Add `azuredb` for `engine_name` validation ([#4506](https://github.com/terraform-providers/terraform-provider-aws/issues/4506))
* resource/aws_rds_cluster: Add `backtrack_window` argument and wait for updates to complete ([#4524](https://github.com/terraform-providers/terraform-provider-aws/issues/4524))
* resource/aws_spot_fleet_request: Add `launch_specification` `iam_instance_profile_arn` argument ([#4511](https://github.com/terraform-providers/terraform-provider-aws/issues/4511))

BUG FIXES:

* data-source/aws_autoscaling_groups: Use pagination function for DescribeTags filtering ([#4535](https://github.com/terraform-providers/terraform-provider-aws/issues/4535))
* resource/aws_elb: Ensure `bucket_prefix` for access logging can be updated to `""` ([#4383](https://github.com/terraform-providers/terraform-provider-aws/issues/4383))
* resource/aws_kinesis_firehose_delivery_stream: Retry on Elasticsearch destination IAM role errors and update IAM errors ([#4518](https://github.com/terraform-providers/terraform-provider-aws/issues/4518))
* resource/aws_launch_template: Allow `network_interfaces` `device_index` to be set to 0 ([#4367](https://github.com/terraform-providers/terraform-provider-aws/issues/4367))
* resource/aws_lb: Ensure `bucket_prefix` for access logging can be updated to `""` ([#4383](https://github.com/terraform-providers/terraform-provider-aws/issues/4383))
* resource/aws_lb: Ensure `access_logs` is properly set into Terraform state ([#4517](https://github.com/terraform-providers/terraform-provider-aws/issues/4517))
* resource/aws_security_group: Fix rule description handling when gathering multiple rules with same permissions ([#4416](https://github.com/terraform-providers/terraform-provider-aws/issues/4416))

## 1.18.0 (May 10, 2018)

FEATURES:

* **New Data Source:** `aws_acmpca_certificate_authority` ([#4458](https://github.com/terraform-providers/terraform-provider-aws/issues/4458))
* **New Resource:** `aws_acmpca_certificate_authority` ([#4458](https://github.com/terraform-providers/terraform-provider-aws/issues/4458))
* **New Resource:** `aws_glue_catalog_table` ([#4368](https://github.com/terraform-providers/terraform-provider-aws/issues/4368))

ENHANCEMENTS:

* provider: Lower retry threshold for DNS resolution failures ([#4459](https://github.com/terraform-providers/terraform-provider-aws/issues/4459))
* resource/aws_dms_endpoint: Support `s3` `engine_name` and add `s3_settings` argument ([#1685](https://github.com/terraform-providers/terraform-provider-aws/issues/1685)] and [[#4447](https://github.com/terraform-providers/terraform-provider-aws/issues/4447))
* resource/aws_glue_job: Add `timeout` argument ([#4460](https://github.com/terraform-providers/terraform-provider-aws/issues/4460))
* resource/aws_lb_target_group: Add `proxy_protocol_v2` argument ([#4365](https://github.com/terraform-providers/terraform-provider-aws/issues/4365))
* resource/aws_spot_fleet_request: Mark `spot_price` optional (defaults to on-demand price) ([#4424](https://github.com/terraform-providers/terraform-provider-aws/issues/4424))
* resource/aws_spot_fleet_request: Add plan time validation for `valid_from` and `valid_until` arguments ([#4463](https://github.com/terraform-providers/terraform-provider-aws/issues/4463))
* resource/aws_spot_instance_request: Mark `spot_price` optional (defaults to on-demand price) ([#4424](https://github.com/terraform-providers/terraform-provider-aws/issues/4424))

BUG FIXES:

* data-source/aws_autoscaling_groups: Correctly paginate through over 50 results ([#4433](https://github.com/terraform-providers/terraform-provider-aws/issues/4433))
* resource/aws_elastic_beanstalk_environment: Correctly handle `cname_prefix` attribute in China partition ([#4485](https://github.com/terraform-providers/terraform-provider-aws/issues/4485))
* resource/aws_glue_job: Remove `allocated_capacity` and `max_concurrent_runs` upper plan time validation limits ([#4461](https://github.com/terraform-providers/terraform-provider-aws/issues/4461))
* resource/aws_instance: Fix `root_device_mapping` matching of expected root device name with multiple block devices. ([#4489](https://github.com/terraform-providers/terraform-provider-aws/issues/4489))
* resource/aws_launch_template: Prevent `parameter iops is not supported for gp2 volumes` error ([#4344](https://github.com/terraform-providers/terraform-provider-aws/issues/4344))
* resource/aws_launch_template: Prevent `'iamInstanceProfile.name' may not be used in combination with 'iamInstanceProfile.arn'` error ([#4344](https://github.com/terraform-providers/terraform-provider-aws/issues/4344))
* resource/aws_launch_template: Prevent `parameter groupName cannot be used with the parameter subnet` error ([#4344](https://github.com/terraform-providers/terraform-provider-aws/issues/4344))
* resource/aws_launch_template: Separate usage of `ipv4_address_count`/`ipv6_address_count` from `ipv4_addresses`/`ipv6_addresses` ([#4344](https://github.com/terraform-providers/terraform-provider-aws/issues/4344))
* resource/aws_redshift_cluster: Properly send all required parameters when resizing ([#3127](https://github.com/terraform-providers/terraform-provider-aws/issues/3127))
* resource/aws_s3_bucket: Prevent crash from empty string CORS arguments ([#4465](https://github.com/terraform-providers/terraform-provider-aws/issues/4465))
* resource/aws_ssm_document: Add missing account ID to `arn` attribute ([#4436](https://github.com/terraform-providers/terraform-provider-aws/issues/4436))

## 1.17.0 (May 02, 2018)

NOTES:

* resource/aws_ecs_service: Please note the `placement_strategy` argument (an unordered list) has been marked deprecated in favor of the `ordered_placement_strategy` argument (an ordered list based on the Terraform configuration ordering).

FEATURES:

* **New Data Source:** `aws_mq_broker` ([#3163](https://github.com/terraform-providers/terraform-provider-aws/issues/3163))
* **New Resource:** `aws_budgets_budget` ([#1879](https://github.com/terraform-providers/terraform-provider-aws/issues/1879))
* **New Resource:** `aws_iam_user_group_membership` ([#3365](https://github.com/terraform-providers/terraform-provider-aws/issues/3365))
* **New Resource:** `aws_vpc_peering_connection_options` ([#3909](https://github.com/terraform-providers/terraform-provider-aws/issues/3909))

ENHANCEMENTS:

* data-source/aws_route53_zone: Add `name_servers` attribute ([#4336](https://github.com/terraform-providers/terraform-provider-aws/issues/4336))
* resource/aws_api_gateway_stage: Add `access_log_settings` argument (Support access logging) ([#4369](https://github.com/terraform-providers/terraform-provider-aws/issues/4369))
* resource/aws_autoscaling_group: Add `launch_template` argument ([#4305](https://github.com/terraform-providers/terraform-provider-aws/issues/4305))
* resource/aws_batch_job_definition: Add `timeout` argument ([#4386](https://github.com/terraform-providers/terraform-provider-aws/issues/4386))
* resource/aws_cloudwatch_event_rule: Add `name_prefix` argument ([#2752](https://github.com/terraform-providers/terraform-provider-aws/issues/2752))
* resource/aws_cloudwatch_event_rule: Make `name` optional (Terraform can generate unique ID) ([#2752](https://github.com/terraform-providers/terraform-provider-aws/issues/2752))
* resource/aws_codedeploy_deployment_group: Add `ec2_tag_set` argument (tag group support) ([#4324](https://github.com/terraform-providers/terraform-provider-aws/issues/4324))
* resource/aws_default_subnet: Allow `map_public_ip_on_launch` updates ([#4396](https://github.com/terraform-providers/terraform-provider-aws/issues/4396))
* resource/aws_dms_endpoint: Support `mongodb` engine_name and `mongodb_settings` argument ([#4406](https://github.com/terraform-providers/terraform-provider-aws/issues/4406))
* resource/aws_dynamodb_table: Add `point_in_time_recovery` argument ([#4063](https://github.com/terraform-providers/terraform-provider-aws/issues/4063))
* resource/aws_ecs_service: Add `ordered_placement_strategy` argument, deprecate `placement_strategy` argument ([#4390](https://github.com/terraform-providers/terraform-provider-aws/issues/4390))
* resource/aws_ecs_service: Allow `health_check_grace_period_seconds` up to 7200 seconds ([#4420](https://github.com/terraform-providers/terraform-provider-aws/issues/4420))
* resource/aws_lambda_permission: Add `statement_id_prefix` argument ([#2743](https://github.com/terraform-providers/terraform-provider-aws/issues/2743))
* resource/aws_lambda_permission: Make `statement_id` optional (Terraform can generate unique ID) ([#2743](https://github.com/terraform-providers/terraform-provider-aws/issues/2743))
* resource/aws_rds_cluster: Add `s3_import` argument (Support MySQL Backup Restore from S3) ([#4366](https://github.com/terraform-providers/terraform-provider-aws/issues/4366))
* resource/aws_vpc_peering_connection: Support configurable timeouts ([#3909](https://github.com/terraform-providers/terraform-provider-aws/issues/3909))

BUG FIXES:

* data-source/aws_instance: Bypass `UnsupportedOperation` errors with `DescribeInstanceCreditSpecifications` call ([#4362](https://github.com/terraform-providers/terraform-provider-aws/issues/4362))
* resource/aws_iam_group_policy: Properly handle generated policy name updates ([#4379](https://github.com/terraform-providers/terraform-provider-aws/issues/4379))
* resource/aws_instance: Bypass `UnsupportedOperation` errors with `DescribeInstanceCreditSpecifications` call ([#4362](https://github.com/terraform-providers/terraform-provider-aws/issues/4362))
* resource/aws_launch_template: Appropriately set `security_groups` in network interfaces ([#4364](https://github.com/terraform-providers/terraform-provider-aws/issues/4364))
* resource/aws_rds_cluster: Add retries for IAM eventual consistency ([#4371](https://github.com/terraform-providers/terraform-provider-aws/issues/4371))
* resource/aws_rds_cluster_instance: Add retries for IAM eventual consistency ([#4370](https://github.com/terraform-providers/terraform-provider-aws/issues/4370))
* resource/aws_route53_zone: Add domain name to CallerReference to prevent creation issues with count greater than one ([#4341](https://github.com/terraform-providers/terraform-provider-aws/issues/4341))

## 1.16.0 (April 25, 2018)

FEATURES:

* **New Data Source:** `aws_batch_compute_environment` ([#4270](https://github.com/terraform-providers/terraform-provider-aws/issues/4270))
* **New Data Source:** `aws_batch_job_queue` ([#4288](https://github.com/terraform-providers/terraform-provider-aws/issues/4288))
* **New Data Source:** `aws_iot_endpoint` ([#4303](https://github.com/terraform-providers/terraform-provider-aws/issues/4303))
* **New Data Source:** `aws_lambda_function` ([#2984](https://github.com/terraform-providers/terraform-provider-aws/issues/2984))
* **New Data Source:** `aws_redshift_cluster` ([#2603](https://github.com/terraform-providers/terraform-provider-aws/issues/2603))
* **New Data Source:** `aws_secretsmanager_secret` ([#4272](https://github.com/terraform-providers/terraform-provider-aws/issues/4272))
* **New Data Source:** `aws_secretsmanager_secret_version` ([#4272](https://github.com/terraform-providers/terraform-provider-aws/issues/4272))
* **New Resource:** `aws_dax_parameter_group` ([#4299](https://github.com/terraform-providers/terraform-provider-aws/issues/4299))
* **New Resource:** `aws_dax_subnet_group` ([#4302](https://github.com/terraform-providers/terraform-provider-aws/issues/4302))
* **New Resource:** `aws_organizations_policy` ([#4249](https://github.com/terraform-providers/terraform-provider-aws/issues/4249))
* **New Resource:** `aws_organizations_policy_attachment` ([#4253](https://github.com/terraform-providers/terraform-provider-aws/issues/4253))
* **New Resource:** `aws_secretsmanager_secret` ([#4272](https://github.com/terraform-providers/terraform-provider-aws/issues/4272))
* **New Resource:** `aws_secretsmanager_secret_version` ([#4272](https://github.com/terraform-providers/terraform-provider-aws/issues/4272))

ENHANCEMENTS:

* data-source/aws_cognito_user_pools: Add `arns` attribute ([#4256](https://github.com/terraform-providers/terraform-provider-aws/issues/4256))
* data-source/aws_ecs_cluster Return error on multiple clusters ([#4286](https://github.com/terraform-providers/terraform-provider-aws/issues/4286))
* data-source/aws_iam_instance_profile: Add `role_arn` and `role_name` attributes ([#4300](https://github.com/terraform-providers/terraform-provider-aws/issues/4300))
* data-source/aws_instance: Add `disable_api_termination` attribute ([#4314](https://github.com/terraform-providers/terraform-provider-aws/issues/4314))
* resource/aws_api_gateway_rest_api: Add `policy` argument ([#4211](https://github.com/terraform-providers/terraform-provider-aws/issues/4211))
* resource/aws_api_gateway_stage: Add `tags` argument ([#2858](https://github.com/terraform-providers/terraform-provider-aws/issues/2858))
* resource/aws_api_gateway_stage: Add `execution_arn` and `invoke_url` attributes ([#3469](https://github.com/terraform-providers/terraform-provider-aws/issues/3469))
* resource/aws_api_gateway_vpc_link: Support import ([#4306](https://github.com/terraform-providers/terraform-provider-aws/issues/4306))
* resource/aws_cloudwatch_event_target: Add `batch_target` argument ([#4312](https://github.com/terraform-providers/terraform-provider-aws/issues/4312))
* resource/aws_cloudwatch_event_target: Add `kinesis_target` and `sqs_target` arguments ([#4323](https://github.com/terraform-providers/terraform-provider-aws/issues/4323))
* resource/aws_cognito_user_pool: Support `user_migration` in `lambda_config` ([#4301](https://github.com/terraform-providers/terraform-provider-aws/issues/4301))
* resource/aws_db_instance: Add `s3_import` argument ([#2728](https://github.com/terraform-providers/terraform-provider-aws/issues/2728))
* resource/aws_elastic_beanstalk_application: Add `appversion_lifecycle` argument ([#1907](https://github.com/terraform-providers/terraform-provider-aws/issues/1907))
* resource/aws_instance: Add `credit_specification` argument (e.g. t2.unlimited support) ([#2619](https://github.com/terraform-providers/terraform-provider-aws/issues/2619))
* resource/aws_kinesis_firehose_delivery_stream: Support Redshift `processing_configuration` ([#4251](https://github.com/terraform-providers/terraform-provider-aws/issues/4251))
* resource/aws_launch_configuration: Add `user_data_base64` argument ([#4257](https://github.com/terraform-providers/terraform-provider-aws/issues/4257))
* resource/aws_s3_bucket: Add support for `ONEZONE_IA` storage class ([#4287](https://github.com/terraform-providers/terraform-provider-aws/issues/4287))
* resource/aws_s3_bucket_object: Add support for `ONEZONE_IA` storage class ([#4287](https://github.com/terraform-providers/terraform-provider-aws/issues/4287))
* resource/aws_spot_instance_request: Add `valid_from` and `valid_until` arguments ([#4018](https://github.com/terraform-providers/terraform-provider-aws/issues/4018))
* resource/aws_ssm_patch_baseline: Support `CENTOS` `operating_system` argument ([#4268](https://github.com/terraform-providers/terraform-provider-aws/issues/4268))

BUG FIXES:

* data-source/aws_iam_policy_document: Prevent crash with multiple value principal identifiers ([#4277](https://github.com/terraform-providers/terraform-provider-aws/issues/4277))
* data-source/aws_lb_listener: Ensure attributes are properly set when not used as arguments ([#4317](https://github.com/terraform-providers/terraform-provider-aws/issues/4317))
* resource/aws_codebuild_project: Mark auth resource attribute as sensitive ([#4284](https://github.com/terraform-providers/terraform-provider-aws/issues/4284))
* resource/aws_cognito_user_pool_client: Fix import to include user pool ID ([#3762](https://github.com/terraform-providers/terraform-provider-aws/issues/3762))
* resource/aws_elasticache_cluster: Remove extraneous plan-time validation for `node_type` and `subnet_group_name` ([#4333](https://github.com/terraform-providers/terraform-provider-aws/issues/4333))
* resource/aws_launch_template: Allow dashes in `name` and `name_prefix` arguments ([#4321](https://github.com/terraform-providers/terraform-provider-aws/issues/4321))
* resource/aws_launch_template: Properly set `block_device_mappings` EBS information into Terraform state ([#4321](https://github.com/terraform-providers/terraform-provider-aws/issues/4321))
* resource/aws_launch_template: Properly pass `block_device_mappings` information to EC2 API ([#4321](https://github.com/terraform-providers/terraform-provider-aws/issues/4321))
* resource/aws_s3_bucket: Prevent panic on lifecycle rule reading errors ([#4282](https://github.com/terraform-providers/terraform-provider-aws/issues/4282))

## 1.15.0 (April 18, 2018)

NOTES:

* resource/aws_cloudfront_distribution: Please note the `cache_behavior` argument (an unordered list) has been marked deprecated in favor of the `ordered_cache_behavior` argument (an ordered list based on the Terraform configuration ordering). This is to support proper cache behavior precedence within a CloudFront distribution.

FEATURES:

* **New Data Source:** `aws_api_gateway_rest_api` ([#4172](https://github.com/terraform-providers/terraform-provider-aws/issues/4172))
* **New Data Source:** `aws_cloudwatch_log_group` ([#4167](https://github.com/terraform-providers/terraform-provider-aws/issues/4167))
* **New Data Source:** `aws_cognito_user_pools` ([#4212](https://github.com/terraform-providers/terraform-provider-aws/issues/4212))
* **New Data Source:** `aws_sqs_queue` ([#2311](https://github.com/terraform-providers/terraform-provider-aws/issues/2311))
* **New Resource:** `aws_directory_service_conditional_forwarder` ([#4071](https://github.com/terraform-providers/terraform-provider-aws/issues/4071))
* **New Resource:** `aws_glue_connection` ([#4016](https://github.com/terraform-providers/terraform-provider-aws/issues/4016))
* **New Resource:** `aws_glue_job` ([#4028](https://github.com/terraform-providers/terraform-provider-aws/issues/4028))
* **New Resource:** `aws_iam_service_linked_role` ([#2985](https://github.com/terraform-providers/terraform-provider-aws/issues/2985))
* **New Resource:** `aws_launch_template` ([#2927](https://github.com/terraform-providers/terraform-provider-aws/issues/2927))
* **New Resource:** `aws_ses_domain_identity_verification` ([#4108](https://github.com/terraform-providers/terraform-provider-aws/issues/4108))

ENHANCEMENTS:

* data-source/aws_iam_server_certificate: Filter by `path_prefix` ([#3801](https://github.com/terraform-providers/terraform-provider-aws/issues/3801))
* resource/aws_api_gateway_integration: Support VPC connection ([#3428](https://github.com/terraform-providers/terraform-provider-aws/issues/3428))
* resource/aws_cloudfront_distribution: Added `ordered_cache_behavior` argument, deprecate `cache_behavior` ([#4117](https://github.com/terraform-providers/terraform-provider-aws/issues/4117))
* resource/aws_db_instance: Support `enabled_cloudwatch_logs_exports` argument ([#4111](https://github.com/terraform-providers/terraform-provider-aws/issues/4111))
* resource/aws_db_option_group: Support option version argument ([#2590](https://github.com/terraform-providers/terraform-provider-aws/issues/2590))
* resource/aws_ecs_service: Support ServiceRegistries ([#3906](https://github.com/terraform-providers/terraform-provider-aws/issues/3906))
* resource/aws_iam_service_linked_role: Support `custom_suffix` and `description` arguments ([#4188](https://github.com/terraform-providers/terraform-provider-aws/issues/4188))
* resource/aws_service_discovery_service: Support `health_check_custom_config` argument ([#4083](https://github.com/terraform-providers/terraform-provider-aws/issues/4083))
* resource/aws_spot_fleet_request: Support configurable delete timeout ([#3940](https://github.com/terraform-providers/terraform-provider-aws/issues/3940))
* resource/aws_spot_instance_request: Support optionally fetching password data ([#4189](https://github.com/terraform-providers/terraform-provider-aws/issues/4189))
* resource/aws_waf_rate_based_rule: Support `RegexMatch` predicate type ([#4069](https://github.com/terraform-providers/terraform-provider-aws/issues/4069))
* resource/aws_waf_rule: Support `RegexMatch` predicate type ([#4069](https://github.com/terraform-providers/terraform-provider-aws/issues/4069))
* resource/aws_wafregional_rate_based_rule: Support `RegexMatch` predicate type ([#4069](https://github.com/terraform-providers/terraform-provider-aws/issues/4069))

BUG FIXES:

* resource/aws_athena_database: Handle database names with uppercase and underscores ([#4133](https://github.com/terraform-providers/terraform-provider-aws/issues/4133))
* resource/aws_codebuild_project: Retry UpdateProject for IAM eventual consistency ([#4238](https://github.com/terraform-providers/terraform-provider-aws/issues/4238))
* resource/aws_codedeploy_deployment_config: Force new resource for `minimum_healthy_hosts` updates ([#4194](https://github.com/terraform-providers/terraform-provider-aws/issues/4194))
* resource/aws_cognito_user_group: Fix `role_arn` updates ([#4237](https://github.com/terraform-providers/terraform-provider-aws/issues/4237))
* resource/aws_elasticache_replication_group: Increase default create timeout to 60 minutes ([#4093](https://github.com/terraform-providers/terraform-provider-aws/issues/4093))
* resource/aws_emr_cluster: Force new resource if any of the `ec2_attributes` change ([#4218](https://github.com/terraform-providers/terraform-provider-aws/issues/4218))
* resource/aws_iam_role: Suppress `NoSuchEntity` errors while detaching policies from role during deletion ([#4209](https://github.com/terraform-providers/terraform-provider-aws/issues/4209))
* resource/aws_lb: Force new resource if any of the `subnet_mapping` attributes change ([#4086](https://github.com/terraform-providers/terraform-provider-aws/issues/4086))
* resource/aws_rds_cluster: Properly handle `engine_version` with `snapshot_identifier` ([#4215](https://github.com/terraform-providers/terraform-provider-aws/issues/4215))
* resource/aws_route53_record: Improved handling of non-alphanumeric record names ([#4183](https://github.com/terraform-providers/terraform-provider-aws/issues/4183))
* resource/aws_spot_instance_request: Fix `instance_interuption_behaviour` hibernate and stop handling with placement ([#1986](https://github.com/terraform-providers/terraform-provider-aws/issues/1986))
* resource/aws_vpc_dhcp_options: Handle plural and non-plural `InvalidDhcpOptionsID.NotFound` errors ([#4136](https://github.com/terraform-providers/terraform-provider-aws/issues/4136))

## 1.14.1 (April 11, 2018)

ENHANCEMENTS:

* resource/aws_db_event_subscription: Add `arn` attribute ([#4151](https://github.com/terraform-providers/terraform-provider-aws/issues/4151))
* resource/aws_db_event_subscription: Support configurable timeouts ([#4151](https://github.com/terraform-providers/terraform-provider-aws/issues/4151))

BUG FIXES:

* resource/aws_codebuild_project: Properly handle setting cache type `NO_CACHE` ([#4134](https://github.com/terraform-providers/terraform-provider-aws/issues/4134))
* resource/aws_db_event_subscription: Fix `tag` ARN handling ([#4151](https://github.com/terraform-providers/terraform-provider-aws/issues/4151))
* resource/aws_dynamodb_table_item: Trigger destructive update if range_key has changed ([#3821](https://github.com/terraform-providers/terraform-provider-aws/issues/3821))
* resource/aws_elb: Return any errors when updating listeners ([#4159](https://github.com/terraform-providers/terraform-provider-aws/issues/4159))
* resource/aws_emr_cluster: Prevent crash with missing StateChangeReason ([#4165](https://github.com/terraform-providers/terraform-provider-aws/issues/4165))
* resource/aws_iam_user: Retry user login profile deletion on `EntityTemporarilyUnmodifiable` ([#4143](https://github.com/terraform-providers/terraform-provider-aws/issues/4143))
* resource/aws_kinesis_firehose_delivery_stream: Prevent crash with missing CloudWatch logging options ([#4148](https://github.com/terraform-providers/terraform-provider-aws/issues/4148))
* resource/aws_lambda_alias: Force new resource on `name` change ([#4106](https://github.com/terraform-providers/terraform-provider-aws/issues/4106))
* resource/aws_lambda_function: Prevent perpetual difference when removing `dead_letter_config` ([#2684](https://github.com/terraform-providers/terraform-provider-aws/issues/2684))
* resource/aws_launch_configuration: Properly read `security_groups`, `user_data`, and `vpc_classic_link_security_groups` attributes into Terraform state ([#2800](https://github.com/terraform-providers/terraform-provider-aws/issues/2800))
* resource/aws_network_acl: Prevent error on deletion with already deleted subnets ([#4119](https://github.com/terraform-providers/terraform-provider-aws/issues/4119))
* resource/aws_network_acl: Prevent error on update with removing associations for already deleted subnets ([#4119](https://github.com/terraform-providers/terraform-provider-aws/issues/4119))
* resource/aws_rds_cluster: Properly handle `engine_version` during regular creation ([#4139](https://github.com/terraform-providers/terraform-provider-aws/issues/4139))
* resource/aws_rds_cluster: Set `port` updates to force new resource ([#4144](https://github.com/terraform-providers/terraform-provider-aws/issues/4144))
* resource/aws_route53_zone: Suppress `name` difference with trailing period ([#3982](https://github.com/terraform-providers/terraform-provider-aws/issues/3982))
* resource/aws_vpc_peering_connection: Allow active pending state during deletion for eventual consistency ([#4140](https://github.com/terraform-providers/terraform-provider-aws/issues/4140))

## 1.14.0 (April 06, 2018)

NOTES:

* resource/aws_organizations_account: As noted in the resource documentation, resource deletion from Terraform will _not_ automatically close AWS accounts due to the behavior of the AWS Organizations service. There are also various manual steps required by AWS before the account can be removed from an organization and made into a standalone account, then manually closed if desired.

FEATURES:

* **New Resource:** `aws_organizations_account` ([#3524](https://github.com/terraform-providers/terraform-provider-aws/issues/3524))
* **New Resource:** `aws_ses_identity_notification_topic` ([#2640](https://github.com/terraform-providers/terraform-provider-aws/issues/2640))

ENHANCEMENTS:

* provider: Fallback to SDK default credential chain if credentials not found using provider credential chain ([#2883](https://github.com/terraform-providers/terraform-provider-aws/issues/2883))
* data-source/aws_iam_role: Add `max_session_duration` attribute ([#4092](https://github.com/terraform-providers/terraform-provider-aws/issues/4092))
* resource/aws_cloudfront_distribution: Add cache_behavior `field_level_encryption_id` attribute ([#4102](https://github.com/terraform-providers/terraform-provider-aws/issues/4102))
* resource/aws_codebuild_project: Support `cache` configuration ([#2860](https://github.com/terraform-providers/terraform-provider-aws/issues/2860))
* resource/aws_elasticache_replication_group: Support Cluster Mode Enabled online shard reconfiguration ([#3932](https://github.com/terraform-providers/terraform-provider-aws/issues/3932))
* resource/aws_elasticache_replication_group: Configurable create, update, and delete timeouts ([#3932](https://github.com/terraform-providers/terraform-provider-aws/issues/3932))
* resource/aws_iam_role: Add `max_session_duration` argument ([#3977](https://github.com/terraform-providers/terraform-provider-aws/issues/3977))
* resource/aws_kinesis_firehose_delivery_stream: Add Elasticsearch destination processing configuration support ([#3621](https://github.com/terraform-providers/terraform-provider-aws/issues/3621))
* resource/aws_kinesis_firehose_delivery_stream: Add Extended S3 destination backup mode support ([#2987](https://github.com/terraform-providers/terraform-provider-aws/issues/2987))
* resource/aws_kinesis_firehose_delivery_stream: Add Splunk destination processing configuration support ([#3944](https://github.com/terraform-providers/terraform-provider-aws/issues/3944))
* resource/aws_lambda_function: Support `nodejs8.10` runtime ([#4020](https://github.com/terraform-providers/terraform-provider-aws/issues/4020))
* resource/aws_launch_configuration: Add support for `ebs_block_device.*.no_device` ([#4070](https://github.com/terraform-providers/terraform-provider-aws/issues/4070))
* resource/aws_ssm_maintenance_window_target: Make resource updatable ([#4074](https://github.com/terraform-providers/terraform-provider-aws/issues/4074))
* resource/aws_wafregional_rule: Validate all predicate types ([#4046](https://github.com/terraform-providers/terraform-provider-aws/issues/4046))

BUG FIXES:

* resource/aws_cognito_user_pool: Trim `custom:` prefix of `developer_only_attribute = false` schema attributes ([#4041](https://github.com/terraform-providers/terraform-provider-aws/issues/4041))
* resource/aws_cognito_user_pool: Fix `email_message_by_link` max length validation ([#4051](https://github.com/terraform-providers/terraform-provider-aws/issues/4051))
* resource/aws_elasticache_replication_group: Properly set `cluster_mode` in state ([#3932](https://github.com/terraform-providers/terraform-provider-aws/issues/3932))
* resource/aws_iam_user_login_profile: Changed password generation to use `crypto/rand` ([#3989](https://github.com/terraform-providers/terraform-provider-aws/issues/3989))
* resource/aws_kinesis_firehose_delivery_stream: Prevent additional crash scenarios with optional configurations ([#4047](https://github.com/terraform-providers/terraform-provider-aws/issues/4047))
* resource/aws_lambda_function: IAM retry for "The role defined for the function cannot be assumed by Lambda" on update ([#3988](https://github.com/terraform-providers/terraform-provider-aws/issues/3988))
* resource/aws_lb: Suppress differences for non-applicable attributes ([#4032](https://github.com/terraform-providers/terraform-provider-aws/issues/4032))
* resource/aws_rds_cluster_instance: Prevent crash on importing non-cluster instances ([#3961](https://github.com/terraform-providers/terraform-provider-aws/issues/3961))
* resource/aws_route53_record: Fix ListResourceRecordSet pagination ([#3900](https://github.com/terraform-providers/terraform-provider-aws/issues/3900))

## 1.13.0 (March 28, 2018)

NOTES:

This release is happening outside the normal release schedule to accommodate a crash fix for the `aws_lb_target_group` resource. It appears an ELBv2 service update rolling out currently is the root cause. The potential for this crash has been present since the initial resource in Terraform 0.7.7 and all versions of the AWS provider up to v1.13.0.

FEATURES:

* **New Resource:** `aws_appsync_datasource` ([#2758](https://github.com/terraform-providers/terraform-provider-aws/issues/2758))
* **New Resource:** `aws_waf_regex_match_set` ([#3947](https://github.com/terraform-providers/terraform-provider-aws/issues/3947))
* **New Resource:** `aws_waf_regex_pattern_set` ([#3913](https://github.com/terraform-providers/terraform-provider-aws/issues/3913))
* **New Resource:** `aws_waf_rule_group` ([#3898](https://github.com/terraform-providers/terraform-provider-aws/issues/3898))
* **New Resource:** `aws_wafregional_geo_match_set` ([#3915](https://github.com/terraform-providers/terraform-provider-aws/issues/3915))
* **New Resource:** `aws_wafregional_rate_based_rule` ([#3871](https://github.com/terraform-providers/terraform-provider-aws/issues/3871))
* **New Resource:** `aws_wafregional_regex_match_set` ([#3950](https://github.com/terraform-providers/terraform-provider-aws/issues/3950))
* **New Resource:** `aws_wafregional_regex_pattern_set` ([#3933](https://github.com/terraform-providers/terraform-provider-aws/issues/3933))
* **New Resource:** `aws_wafregional_rule_group` ([#3948](https://github.com/terraform-providers/terraform-provider-aws/issues/3948))

ENHANCEMENTS:

* provider: Support custom Elasticsearch endpoint ([#3941](https://github.com/terraform-providers/terraform-provider-aws/issues/3941))
* resource/aws_appsync_graphql_api: Support import ([#3500](https://github.com/terraform-providers/terraform-provider-aws/issues/3500))
* resource/aws_elasticache_cluster: Allow port to be optional ([#3835](https://github.com/terraform-providers/terraform-provider-aws/issues/3835))
* resource/aws_elasticache_cluster: Add `replication_group_id` argument ([#3869](https://github.com/terraform-providers/terraform-provider-aws/issues/3869))
* resource/aws_elasticache_replication_group: Allow port to be optional ([#3835](https://github.com/terraform-providers/terraform-provider-aws/issues/3835))

BUG FIXES:

* resource/aws_autoscaling_group: Fix updating of `service_linked_role` ([#3942](https://github.com/terraform-providers/terraform-provider-aws/issues/3942))
* resource/aws_autoscaling_group: Properly set empty `enabled_metrics` in the state during read ([#3899](https://github.com/terraform-providers/terraform-provider-aws/issues/3899))
* resource/aws_autoscaling_policy: Fix conditional logic based on `policy_type` ([#3739](https://github.com/terraform-providers/terraform-provider-aws/issues/3739))
* resource/aws_batch_compute_environment: Correctly set `compute_resources` in state ([#3824](https://github.com/terraform-providers/terraform-provider-aws/issues/3824))
* resource/aws_cognito_user_pool: Correctly set `schema` in state ([#3789](https://github.com/terraform-providers/terraform-provider-aws/issues/3789))
* resource/aws_iam_user_login_profile: Fix `password_length` validation function regression from 1.12.0 ([#3919](https://github.com/terraform-providers/terraform-provider-aws/issues/3919))
* resource/aws_lb: Store correct state for http2 and ensure attributes are set on create ([#3854](https://github.com/terraform-providers/terraform-provider-aws/issues/3854))
* resource/aws_lb: Correctly set `subnet_mappings` in state ([#3822](https://github.com/terraform-providers/terraform-provider-aws/issues/3822))
* resource/aws_lb_listener: Retry CertificateNotFound errors on update for IAM eventual consistency ([#3901](https://github.com/terraform-providers/terraform-provider-aws/issues/3901))
* resource/aws_lb_target_group: Prevent crash from missing matcher during read ([#3954](https://github.com/terraform-providers/terraform-provider-aws/issues/3954))
* resource/aws_security_group: Retry read on creation for EC2 eventual consistency ([#3892](https://github.com/terraform-providers/terraform-provider-aws/issues/3892))


## 1.12.0 (March 23, 2018)

NOTES:

* provider: For resources implementing the IAM policy equivalence library (https://github.com/jen20/awspolicyequivalence/) on an attribute via `suppressEquivalentAwsPolicyDiffs`, the dependency has been updated, which should mark additional IAM policies as equivalent. ([#3832](https://github.com/terraform-providers/terraform-provider-aws/issues/3832))

FEATURES:

* **New Resource:** `aws_kms_grant` ([#3038](https://github.com/terraform-providers/terraform-provider-aws/issues/3038))
* **New Resource:** `aws_waf_geo_match_set` ([#3275](https://github.com/terraform-providers/terraform-provider-aws/issues/3275))
* **New Resource:** `aws_wafregional_rule` ([#3756](https://github.com/terraform-providers/terraform-provider-aws/issues/3756))
* **New Resource:** `aws_wafregional_size_constraint_set` ([#3796](https://github.com/terraform-providers/terraform-provider-aws/issues/3796))
* **New Resource:** `aws_wafregional_sql_injection_match_set` ([#1013](https://github.com/terraform-providers/terraform-provider-aws/issues/1013))
* **New Resource:** `aws_wafregional_web_acl` ([#3754](https://github.com/terraform-providers/terraform-provider-aws/issues/3754))
* **New Resource:** `aws_wafregional_web_acl_association` ([#3755](https://github.com/terraform-providers/terraform-provider-aws/issues/3755))
* **New Resource:** `aws_wafregional_xss_match_set` ([#1014](https://github.com/terraform-providers/terraform-provider-aws/issues/1014))

ENHANCEMENTS:

* provider: Treat IAM policies with account ID principals as equivalent to IAM account root ARN ([#3832](https://github.com/terraform-providers/terraform-provider-aws/issues/3832))
* provider: Treat additional IAM policy scenarios with empty principal trees as equivalent ([#3832](https://github.com/terraform-providers/terraform-provider-aws/issues/3832))
* resource/aws_acm_certificate: Retry on ResourceInUseException during deletion for eventual consistency ([#3868](https://github.com/terraform-providers/terraform-provider-aws/issues/3868))
* resource/aws_api_gateway_rest_api: Add support for content encoding ([#3642](https://github.com/terraform-providers/terraform-provider-aws/issues/3642))
* resource/aws_autoscaling_group: Add `service_linked_role_arn` argument ([#3812](https://github.com/terraform-providers/terraform-provider-aws/issues/3812))
* resource/aws_cloudfront_distribution: Validate origin `domain_name` and `origin_id` at plan time ([#3767](https://github.com/terraform-providers/terraform-provider-aws/issues/3767))
* resource/aws_eip: Support configurable timeouts ([#3769](https://github.com/terraform-providers/terraform-provider-aws/issues/3769))
* resource/aws_elasticache_cluster: Support plan time validation of az_mode ([#3857](https://github.com/terraform-providers/terraform-provider-aws/issues/3857))
* resource/aws_elasticache_cluster: Support plan time validation of node_type requiring VPC for cache.t2 instances ([#3857](https://github.com/terraform-providers/terraform-provider-aws/issues/3857))
* resource/aws_elasticache_cluster: Support plan time validation of num_cache_nodes > 1 for redis ([#3857](https://github.com/terraform-providers/terraform-provider-aws/issues/3857))
* resource/aws_elasticache_cluster: ForceNew on node_type changes for memcached engine ([#3857](https://github.com/terraform-providers/terraform-provider-aws/issues/3857))
* resource/aws_elasticache_cluster: ForceNew on engine_version downgrades ([#3857](https://github.com/terraform-providers/terraform-provider-aws/issues/3857))
* resource/aws_emr_cluster: Add step support ([#3673](https://github.com/terraform-providers/terraform-provider-aws/issues/3673))
* resource/aws_instance: Support optionally fetching encrypted Windows password data ([#2219](https://github.com/terraform-providers/terraform-provider-aws/issues/2219))
* resource/aws_launch_configuration: Validate `user_data` length during plan ([#2973](https://github.com/terraform-providers/terraform-provider-aws/issues/2973))
* resource/aws_lb_target_group: Validate health check threshold for TCP protocol during plan ([#3782](https://github.com/terraform-providers/terraform-provider-aws/issues/3782))
* resource/aws_security_group: Add arn attribute ([#3751](https://github.com/terraform-providers/terraform-provider-aws/issues/3751))
* resource/aws_ses_domain_identity: Support trailing period in domain name ([#3840](https://github.com/terraform-providers/terraform-provider-aws/issues/3840))
* resource/aws_sqs_queue: Support lack of ListQueueTags for all non-standard AWS implementations ([#3794](https://github.com/terraform-providers/terraform-provider-aws/issues/3794))
* resource/aws_ssm_document: Add `document_format` argument to support YAML ([#3814](https://github.com/terraform-providers/terraform-provider-aws/issues/3814))
* resource/aws_s3_bucket_object: New `content_base64` argument allows uploading raw binary data created in-memory, rather than reading from disk as with `source`. ([#3788](https://github.com/terraform-providers/terraform-provider-aws/issues/3788))

BUG FIXES:

* resource/aws_api_gateway_client_certificate: Export `*_date` fields correctly ([#3805](https://github.com/terraform-providers/terraform-provider-aws/issues/3805))
* resource/aws_cognito_user_pool: Detect `auto_verified_attributes` changes ([#3786](https://github.com/terraform-providers/terraform-provider-aws/issues/3786))
* resource/aws_cognito_user_pool_client: Fix `callback_urls` updates ([#3404](https://github.com/terraform-providers/terraform-provider-aws/issues/3404))
* resource/aws_db_instance: Support `incompatible-parameters` and `storage-full` state ([#3708](https://github.com/terraform-providers/terraform-provider-aws/issues/3708))
* resource/aws_dynamodb_table: Update and validate attributes correctly ([#3194](https://github.com/terraform-providers/terraform-provider-aws/issues/3194))
* resource/aws_ecs_task_definition: Correctly read `volume` attribute into Terraform state ([#3823](https://github.com/terraform-providers/terraform-provider-aws/issues/3823))
* resource/aws_kinesis_firehose_delivery_stream: Prevent crash on malformed ID for import ([#3834](https://github.com/terraform-providers/terraform-provider-aws/issues/3834))
* resource/aws_lambda_function: Only retry IAM eventual consistency errors for one minute ([#3765](https://github.com/terraform-providers/terraform-provider-aws/issues/3765))
* resource/aws_ssm_association: Prevent AssociationDoesNotExist error ([#3776](https://github.com/terraform-providers/terraform-provider-aws/issues/3776))
* resource/aws_vpc_endpoint: Prevent perpertual diff in non-standard partitions ([#3317](https://github.com/terraform-providers/terraform-provider-aws/issues/3317))

## 1.11.0 (March 09, 2018)

FEATURES:

* **New Data Source:** `aws_kms_key` ([#2224](https://github.com/terraform-providers/terraform-provider-aws/issues/2224))
* **New Resource:** `aws_organizations_organization` ([#903](https://github.com/terraform-providers/terraform-provider-aws/issues/903))
* **New Resource:** `aws_iot_thing` ([#3521](https://github.com/terraform-providers/terraform-provider-aws/issues/3521))

ENHANCEMENTS:

* resource/aws_api_gateway_authorizer: Support COGNITO_USER_POOLS type ([#3156](https://github.com/terraform-providers/terraform-provider-aws/issues/3156))
* resource/aws_cloud9_environment_ec2: Retry creation for IAM eventual consistency ([#3651](https://github.com/terraform-providers/terraform-provider-aws/issues/3651))
* resource/aws_cloudfront_distribution: Make `default_ttl`, `max_ttl`, and `min_ttl` arguments optional ([#3571](https://github.com/terraform-providers/terraform-provider-aws/issues/3571))
* resource/aws_dms_endpoint: Add aurora-postgresql as a target ([#2615](https://github.com/terraform-providers/terraform-provider-aws/issues/2615))
* resource/aws_dynamodb_table: Support Server Side Encryption ([#3303](https://github.com/terraform-providers/terraform-provider-aws/issues/3303))
* resource/aws_elastic_beanstalk_environment: Support modifying `tags` ([#3513](https://github.com/terraform-providers/terraform-provider-aws/issues/3513))
* resource/aws_emr_cluster: Add Kerberos support ([#3553](https://github.com/terraform-providers/terraform-provider-aws/issues/3553))
* resource/aws_iam_account_alias: Improve error messages to include API errors ([#3590](https://github.com/terraform-providers/terraform-provider-aws/issues/3590))
* resource/aws_iam_user_policy: Add support for import ([#3198](https://github.com/terraform-providers/terraform-provider-aws/issues/3198))
* resource/aws_lb: Add `enable_cross_zone_load_balancing` argument for NLBs ([#3537](https://github.com/terraform-providers/terraform-provider-aws/issues/3537))
* resource/aws_lb: Add `enable_http2` argument for ALBs ([#3609](https://github.com/terraform-providers/terraform-provider-aws/issues/3609))
* resource/aws_route: Add configurable timeouts ([#3639](https://github.com/terraform-providers/terraform-provider-aws/issues/3639))
* resource/aws_security_group: Add configurable timeouts ([#3599](https://github.com/terraform-providers/terraform-provider-aws/issues/3599))
* resource/aws_spot_fleet_request: Add `load_balancers` and `target_group_arns` arguments ([#2564](https://github.com/terraform-providers/terraform-provider-aws/issues/2564))
* resource/aws_ssm_parameter: Add `allowed_pattern`, `description`, and `tags` arguments ([#1520](https://github.com/terraform-providers/terraform-provider-aws/issues/1520))
* resource/aws_ssm_parameter: Allow `key_id` updates ([#1520](https://github.com/terraform-providers/terraform-provider-aws/issues/1520))

BUG FIXES:

* data-source/aws_db_instance: Prevent crash with EC2 Classic ([#3619](https://github.com/terraform-providers/terraform-provider-aws/issues/3619))
* data-source/aws_vpc_endpoint_service: Fix aws-us-gov partition handling ([#3514](https://github.com/terraform-providers/terraform-provider-aws/issues/3514))
* resource/aws_api_gateway_vpc_link: Ensure `target_arns` is properly read ([#3569](https://github.com/terraform-providers/terraform-provider-aws/issues/3569))
* resource/aws_batch_compute_environment: Fix `state` updates ([#3508](https://github.com/terraform-providers/terraform-provider-aws/issues/3508))
* resource/aws_ebs_snapshot: Prevent crash with outside snapshot deletion ([#3462](https://github.com/terraform-providers/terraform-provider-aws/issues/3462))
* resource/aws_ecs_service: Prevent crash when importing non-existent service ([#3672](https://github.com/terraform-providers/terraform-provider-aws/issues/3672))
* resource/aws_eip_association: Prevent deletion error InvalidAssociationID.NotFound ([#3653](https://github.com/terraform-providers/terraform-provider-aws/issues/3653))
* resource/aws_instance: Ensure at least one security group is being attached when modifying vpc_security_group_ids ([#2850](https://github.com/terraform-providers/terraform-provider-aws/issues/2850))
* resource/aws_lambda_function: Allow PutFunctionConcurrency retries on creation ([#3570](https://github.com/terraform-providers/terraform-provider-aws/issues/3570))
* resource/aws_spot_instance_request: Retry for 1 minute instead of 15 seconds for IAM eventual consistency ([#3561](https://github.com/terraform-providers/terraform-provider-aws/issues/3561))
* resource/aws_ssm_activation: Prevent crash with expiration_date ([#3597](https://github.com/terraform-providers/terraform-provider-aws/issues/3597))

## 1.10.0 (February 24, 2018)

NOTES:

* resource/aws_dx_lag: `number_of_connections` was deprecated and will be removed in future major version. Use `aws_dx_connection` and `aws_dx_connection_association` resources instead. Default connections will be removed as part of LAG creation automatically in future major version. ([#3367](https://github.com/terraform-providers/terraform-provider-aws/issues/3367))

FEATURES:

* **New Data Source:** `aws_inspector_rules_packages` ([#3175](https://github.com/terraform-providers/terraform-provider-aws/issues/3175))
* **New Resource:** `aws_api_gateway_vpc_link` ([#2512](https://github.com/terraform-providers/terraform-provider-aws/issues/2512))
* **New Resource:** `aws_appsync_graphql_api` ([#2494](https://github.com/terraform-providers/terraform-provider-aws/issues/2494))
* **New Resource:** `aws_dax_cluster` ([#2884](https://github.com/terraform-providers/terraform-provider-aws/issues/2884))
* **New Resource:** `aws_gamelift_alias` ([#3353](https://github.com/terraform-providers/terraform-provider-aws/issues/3353))
* **New Resource:** `aws_gamelift_fleet` ([#3327](https://github.com/terraform-providers/terraform-provider-aws/issues/3327))
* **New Resource:** `aws_lb_listener_certificate` ([#2686](https://github.com/terraform-providers/terraform-provider-aws/issues/2686))
* **New Resource:** `aws_s3_bucket_metric` ([#916](https://github.com/terraform-providers/terraform-provider-aws/issues/916))
* **New Resource:** `aws_ses_domain_mail_from` ([#2029](https://github.com/terraform-providers/terraform-provider-aws/issues/2029))
* **New Resource:** `aws_iot_thing_type` ([#3302](https://github.com/terraform-providers/terraform-provider-aws/issues/3302))

ENHANCEMENTS:

* data-source/aws_kms_alias: Always return `target_key_arn` ([#3304](https://github.com/terraform-providers/terraform-provider-aws/issues/3304))
* resource/aws_autoscaling_policy: Add support for `target_tracking_configuration` ([#2611](https://github.com/terraform-providers/terraform-provider-aws/issues/2611))
* resource/aws_codebuild_project: Support VPC configuration ([#2547](https://github.com/terraform-providers/terraform-provider-aws/issues/2547)] [[#3324](https://github.com/terraform-providers/terraform-provider-aws/issues/3324))
* resource/aws_cloudtrail: Add `event_selector` argument ([#2258](https://github.com/terraform-providers/terraform-provider-aws/issues/2258))
* resource/aws_codedeploy_deployment_group: Validate DeploymentReady and InstanceReady `trigger_events` ([#3412](https://github.com/terraform-providers/terraform-provider-aws/issues/3412))
* resource/aws_db_parameter_group: Validate underscore `name` during plan ([#3396](https://github.com/terraform-providers/terraform-provider-aws/issues/3396))
* resource/aws_directory_service_directory Add `edition` argument ([#3421](https://github.com/terraform-providers/terraform-provider-aws/issues/3421))
* resource/aws_directory_service_directory Validate `size` argument ([#3453](https://github.com/terraform-providers/terraform-provider-aws/issues/3453))
* resource/aws_dx_connection: Add support for tagging ([#2990](https://github.com/terraform-providers/terraform-provider-aws/issues/2990))
* resource/aws_dx_connection: Add support for import ([#2992](https://github.com/terraform-providers/terraform-provider-aws/issues/2992))
* resource/aws_dx_lag: Add support for tagging ([#2990](https://github.com/terraform-providers/terraform-provider-aws/issues/2990))
* resource/aws_dx_lag: Add support for import ([#2992](https://github.com/terraform-providers/terraform-provider-aws/issues/2992))
* resource/aws_emr_cluster: Add `autoscaling_policy` argument ([#2877](https://github.com/terraform-providers/terraform-provider-aws/issues/2877))
* resource/aws_emr_cluster: Add `scale_down_behavior` argument ([#3063](https://github.com/terraform-providers/terraform-provider-aws/issues/3063))
* resource/aws_instance: Expose reason of `shutting-down` state during creation ([#3371](https://github.com/terraform-providers/terraform-provider-aws/issues/3371))
* resource/aws_instance: Include size of user_data in validation error message ([#2971](https://github.com/terraform-providers/terraform-provider-aws/issues/2971))
* resource/aws_instance: Remove extra API call on creation for SGs ([#3426](https://github.com/terraform-providers/terraform-provider-aws/issues/3426))
* resource/aws_lambda_function: Recompute `version` and `qualified_arn` attributes on publish ([#3032](https://github.com/terraform-providers/terraform-provider-aws/issues/3032))
* resource/aws_lb_target_group: Allow stickiness block set to false with TCP ([#2954](https://github.com/terraform-providers/terraform-provider-aws/issues/2954))
* resource/aws_lb_listener_rule: Validate `priority` over 50000 ([#3379](https://github.com/terraform-providers/terraform-provider-aws/issues/3379))
* resource/aws_lb_listener_rule: Make `priority` argument optional ([#3219](https://github.com/terraform-providers/terraform-provider-aws/issues/3219))
* resource/aws_rds_cluster: Add `hosted_zone_id` attribute ([#3267](https://github.com/terraform-providers/terraform-provider-aws/issues/3267))
* resource/aws_rds_cluster: Add support for `source_region` (encrypted cross-region replicas) ([#3415](https://github.com/terraform-providers/terraform-provider-aws/issues/3415))
* resource/aws_rds_cluster_instance: Support `availability_zone` ([#2812](https://github.com/terraform-providers/terraform-provider-aws/issues/2812))
* resource/aws_rds_cluster_parameter_group: Validate underscore `name` during plan ([#3396](https://github.com/terraform-providers/terraform-provider-aws/issues/3396))
* resource/aws_route53_record Add `allow_overwrite` argument ([#2926](https://github.com/terraform-providers/terraform-provider-aws/issues/2926))
* resource/aws_s3_bucket Ssupport for SSE-KMS replication configuration ([#2625](https://github.com/terraform-providers/terraform-provider-aws/issues/2625))
* resource/aws_spot_fleet_request: Validate `iam_fleet_role` as ARN during plan ([#3431](https://github.com/terraform-providers/terraform-provider-aws/issues/3431))
* resource/aws_sqs_queue: Validate `name` during plan ([#2837](https://github.com/terraform-providers/terraform-provider-aws/issues/2837))
* resource/aws_ssm_association: Allow updating `targets` ([#2807](https://github.com/terraform-providers/terraform-provider-aws/issues/2807))
* resource/aws_service_discovery_service: Support routing policy and update the type of DNS record ([#3273](https://github.com/terraform-providers/terraform-provider-aws/issues/3273))

BUG FIXES:

* data-source/aws_elb_service_account: Correct GovCloud region ([#3315](https://github.com/terraform-providers/terraform-provider-aws/issues/3315))
* resource/aws_acm_certificate_validation: Prevent crash on `validation_record_fqdns` ([#3336](https://github.com/terraform-providers/terraform-provider-aws/issues/3336))
* resource/aws_acm_certificate_validation: Fix `validation_record_fqdns` handling with combined root and wildcard requests ([#3366](https://github.com/terraform-providers/terraform-provider-aws/issues/3366))
* resource/aws_autoscaling_policy: `cooldown` with zero value not set correctly ([#2809](https://github.com/terraform-providers/terraform-provider-aws/issues/2809))
* resource/aws_cloudtrail: Now respects initial `include_global_service_events = false` ([#2817](https://github.com/terraform-providers/terraform-provider-aws/issues/2817))
* resource/aws_dynamodb_table: Retry deletion on ResourceInUseException ([#3355](https://github.com/terraform-providers/terraform-provider-aws/issues/3355))
* resource/aws_dx_lag: `number_of_connections` deprecated (made Optional). Omitting field may now prevent spurious diffs. ([#3367](https://github.com/terraform-providers/terraform-provider-aws/issues/3367))
* resource/aws_ecs_service: Retry DescribeServices after creation ([#3387](https://github.com/terraform-providers/terraform-provider-aws/issues/3387))
* resource/aws_ecs_service: Fix reading `load_balancer` into state ([#3502](https://github.com/terraform-providers/terraform-provider-aws/issues/3502))
* resource/aws_elasticsearch_domain: Retry creation on `ValidationException` ([#3375](https://github.com/terraform-providers/terraform-provider-aws/issues/3375))
* resource/aws_iam_user_ssh_key: Correctly set status after creation ([#3390](https://github.com/terraform-providers/terraform-provider-aws/issues/3390))
* resource/aws_instance: Bump deletion timeout to 20mins ([#3452](https://github.com/terraform-providers/terraform-provider-aws/issues/3452))
* resource/aws_kinesis_firehose_delivery_stream: Retry on additional IAM eventual consistency errors ([#3381](https://github.com/terraform-providers/terraform-provider-aws/issues/3381))
* resource/aws_route53_record: Trim trailing dot during import ([#3321](https://github.com/terraform-providers/terraform-provider-aws/issues/3321))
* resource/aws_s3_bucket: Prevent crashes on location and replication read retry timeouts ([#3338](https://github.com/terraform-providers/terraform-provider-aws/issues/3338))
* resource/aws_s3_bucket: Always set replication_configuration in state ([#3349](https://github.com/terraform-providers/terraform-provider-aws/issues/3349))
* resource/aws_security_group: Allow empty rule description ([#2846](https://github.com/terraform-providers/terraform-provider-aws/issues/2846))
* resource/aws_sns_topic: Fix exit after updating first attribute ([#3360](https://github.com/terraform-providers/terraform-provider-aws/issues/3360))
* resource/aws_spot_instance_request: Bump delete timeout to 20mins ([#3435](https://github.com/terraform-providers/terraform-provider-aws/issues/3435))
* resource/aws_sqs_queue: Skip SQS ListQueueTags in aws-us-gov partition ([#3376](https://github.com/terraform-providers/terraform-provider-aws/issues/3376))
* resource/aws_vpc_endpoint: Treat pending as expected state during deletion ([#3370](https://github.com/terraform-providers/terraform-provider-aws/issues/3370))
* resource/aws_vpc_peering_connection: Treat `pending-acceptance` as expected during deletion ([#3393](https://github.com/terraform-providers/terraform-provider-aws/issues/3393))
* resource/aws_cognito_user_pool_client: support `USER_PASSWORD_AUTH` for explicit_auth_flows ([#3417](https://github.com/terraform-providers/terraform-provider-aws/issues/3417))

## 1.9.0 (February 09, 2018)

NOTES:

* data-source/aws_region: `current` field is deprecated and the data source defaults to the provider region if no endpoint or name is specified ([#3157](https://github.com/terraform-providers/terraform-provider-aws/issues/3157))
* data-source/aws_iam_policy_document: Statements are now de-duplicated per `Sid`s ([#2890](https://github.com/terraform-providers/terraform-provider-aws/issues/2890))

FEATURES:

* **New Data Source:** `aws_elastic_beanstalk_hosted_zone` ([#3208](https://github.com/terraform-providers/terraform-provider-aws/issues/3208))
* **New Data Source:** `aws_iam_policy` ([#1999](https://github.com/terraform-providers/terraform-provider-aws/issues/1999))
* **New Resource:** `aws_acm_certificate` ([#2813](https://github.com/terraform-providers/terraform-provider-aws/issues/2813))
* **New Resource:** `aws_acm_certificate_validation` ([#2813](https://github.com/terraform-providers/terraform-provider-aws/issues/2813))
* **New Resource:** `aws_api_gateway_documentation_version` ([#3287](https://github.com/terraform-providers/terraform-provider-aws/issues/3287))
* **New Resource:** `aws_cloud9_environment_ec2` ([#3291](https://github.com/terraform-providers/terraform-provider-aws/issues/3291))
* **New Resource:** `aws_cognito_user_group` ([#3010](https://github.com/terraform-providers/terraform-provider-aws/issues/3010))
* **New Resource:** `aws_dynamodb_table_item` ([#3238](https://github.com/terraform-providers/terraform-provider-aws/issues/3238))
* **New Resource:** `aws_guardduty_ipset` ([#3161](https://github.com/terraform-providers/terraform-provider-aws/issues/3161))
* **New Resource:** `aws_guardduty_threatintelset` ([#3200](https://github.com/terraform-providers/terraform-provider-aws/issues/3200))
* **New Resource:** `aws_iot_topic_rule` ([#1858](https://github.com/terraform-providers/terraform-provider-aws/issues/1858))
* **New Resource:** `aws_sns_platform_application` ([#1101](https://github.com/terraform-providers/terraform-provider-aws/issues/1101)] [[#3283](https://github.com/terraform-providers/terraform-provider-aws/issues/3283))
* **New Resource:** `aws_vpc_endpoint_service_allowed_principal` ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* **New Resource:** `aws_vpc_endpoint_service_connection_notification` ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* **New Resource:** `aws_vpc_endpoint_service` ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* **New Resource:** `aws_vpc_endpoint_subnet_association` ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))

ENHANCEMENTS:

* provider: Automatically determine AWS partition from configured region ([#3173](https://github.com/terraform-providers/terraform-provider-aws/issues/3173))
* provider: Automatically validate new regions from AWS SDK ([#3159](https://github.com/terraform-providers/terraform-provider-aws/issues/3159))
* data-source/aws_acm_certificate Add `most_recent` attribute for filtering ([#1837](https://github.com/terraform-providers/terraform-provider-aws/issues/1837))
* data-source/aws_iam_policy_document: Support layering via source_json and override_json attributes ([#2890](https://github.com/terraform-providers/terraform-provider-aws/issues/2890))
* data-source/aws_lb_listener: Support load_balancer_arn and port arguments ([#2886](https://github.com/terraform-providers/terraform-provider-aws/issues/2886))
* data-source/aws_network_interface: Add filter attribute ([#2851](https://github.com/terraform-providers/terraform-provider-aws/issues/2851))
* data-source/aws_region: Remove EC2 API call and default to current if no endpoint or name specified ([#3157](https://github.com/terraform-providers/terraform-provider-aws/issues/3157))
* data-source/aws_vpc_endpoint: Support AWS PrivateLink ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* data-source/aws_vpc_endpoint_service: Support AWS PrivateLink ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* resource/aws_athena_named_query: Support import ([#3231](https://github.com/terraform-providers/terraform-provider-aws/issues/3231))
* resource/aws_dynamodb_table: Add custom creation timeout ([#3195](https://github.com/terraform-providers/terraform-provider-aws/issues/3195))
* resource/aws_dynamodb_table: Validate attribute types ([#3188](https://github.com/terraform-providers/terraform-provider-aws/issues/3188))
* resource/aws_ecr_lifecycle_policy: Support import ([#3246](https://github.com/terraform-providers/terraform-provider-aws/issues/3246))
* resource/aws_ecs_service: Support import ([#2764](https://github.com/terraform-providers/terraform-provider-aws/issues/2764))
* resource/aws_ecs_service: Add public_assign_ip argument for Fargate services ([#2559](https://github.com/terraform-providers/terraform-provider-aws/issues/2559))
* resource/aws_kinesis_firehose_delivery_stream: Add splunk configuration ([#3117](https://github.com/terraform-providers/terraform-provider-aws/issues/3117))
* resource/aws_mq_broker: Validate user password ([#3164](https://github.com/terraform-providers/terraform-provider-aws/issues/3164))
* resource/aws_service_discovery_public_dns_namespace: Support import ([#3229](https://github.com/terraform-providers/terraform-provider-aws/issues/3229))
* resource/aws_service_discovery_service: Support import ([#3227](https://github.com/terraform-providers/terraform-provider-aws/issues/3227))
* resource/aws_rds_cluster: Add support for Aurora MySQL 5.7 ([#3278](https://github.com/terraform-providers/terraform-provider-aws/issues/3278))
* resource/aws_sns_topic: Add support for delivery status ([#2872](https://github.com/terraform-providers/terraform-provider-aws/issues/2872))
* resource/aws_sns_topic: Add support for name prefixes and fully generated names ([#2753](https://github.com/terraform-providers/terraform-provider-aws/issues/2753))
* resource/aws_sns_topic_subscription: Support filter policy ([#2806](https://github.com/terraform-providers/terraform-provider-aws/issues/2806))
* resource/aws_ssm_resource_data_sync: Support import ([#3232](https://github.com/terraform-providers/terraform-provider-aws/issues/3232))
* resource/aws_vpc_endpoint: Support AWS PrivateLink ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* resource/aws_vpc_endpoint_service: Support AWS PrivateLink ([#2515](https://github.com/terraform-providers/terraform-provider-aws/issues/2515))
* resource/aws_vpn_gateway: Add support for Amazon side private ASN ([#1888](https://github.com/terraform-providers/terraform-provider-aws/issues/1888))

BUG FIXES:

* data-source/aws_kms_alias: Prevent crash on aliases without target key ([#3203](https://github.com/terraform-providers/terraform-provider-aws/issues/3203))
* data-source/aws_ssm_parameter: Fix wrong arn attribute for full path parameter names ([#3211](https://github.com/terraform-providers/terraform-provider-aws/issues/3211))
* resource/aws_instance: Fix perpertual diff on default VPC instances using vpc_security_group_ids ([#2338](https://github.com/terraform-providers/terraform-provider-aws/issues/2338))
* resource/aws_codebuild_project: Prevent crash when using source auth configuration ([#3271](https://github.com/terraform-providers/terraform-provider-aws/issues/3271))
* resource/aws_cognito_identity_pool_roles_attachment: Fix validation for Token types ([#2894](https://github.com/terraform-providers/terraform-provider-aws/issues/2894))
* resource/aws_db_parameter_group: fix permanent diff when specifying parameters with database-default values ([#3182](https://github.com/terraform-providers/terraform-provider-aws/issues/3182))
* resource/aws_ecs_service: Retry only on ECS and IAM related InvalidParameterException ([#3240](https://github.com/terraform-providers/terraform-provider-aws/issues/3240))
* resource/aws_kinesis_firehose_delivery_stream: Prevent crashes on empty CloudWatchLoggingOptions ([#3301](https://github.com/terraform-providers/terraform-provider-aws/issues/3301))
* resource/aws_kinesis_firehose_delivery_stream: Fix extended_s3_configuration kms_key_arn handling from AWS API ([#3301](https://github.com/terraform-providers/terraform-provider-aws/issues/3301))
* resource/aws_kinesis_stream: Retry deletion on `LimitExceededException` ([#3108](https://github.com/terraform-providers/terraform-provider-aws/issues/3108))
* resource/aws_route53_record: Fix dualstack alias name regression trimming too many characters ([#3187](https://github.com/terraform-providers/terraform-provider-aws/issues/3187))
* resource/aws_ses_template: Send only specified attributes for update ([#3214](https://github.com/terraform-providers/terraform-provider-aws/issues/3214))
* resource/aws_dynamodb_table: Allow disabling stream with empty `stream_view_type` ([#3197](https://github.com/terraform-providers/terraform-provider-aws/issues/3197)] [[#3224](https://github.com/terraform-providers/terraform-provider-aws/issues/3224))
* resource/aws_dx_connection_association: Retry disassociation ([#3212](https://github.com/terraform-providers/terraform-provider-aws/issues/3212))
* resource/aws_volume_attachment: Allow updating `skip_destroy` and `force_detach` ([#2810](https://github.com/terraform-providers/terraform-provider-aws/issues/2810))

## 1.8.0 (January 29, 2018)

FEATURES:

* **New Resource:** `aws_dynamodb_global_table` ([#2517](https://github.com/terraform-providers/terraform-provider-aws/issues/2517))
* **New Resource:** `aws_gamelift_build` ([#2843](https://github.com/terraform-providers/terraform-provider-aws/issues/2843))

ENHANCEMENTS:

* provider: `cn-northwest-1` region is now supported ([#3142](https://github.com/terraform-providers/terraform-provider-aws/issues/3142))
* data-source/aws_kms_alias: Add target_key_arn attribute ([#2551](https://github.com/terraform-providers/terraform-provider-aws/issues/2551))
* resource/aws_api_gateway_integration: Allow update of content_handling attributes ([#3123](https://github.com/terraform-providers/terraform-provider-aws/issues/3123))
* resource/aws_appautoscaling_target: Support updating max_capacity, min_capacity, and role_arn attributes ([#2950](https://github.com/terraform-providers/terraform-provider-aws/issues/2950))
* resource/aws_cloudwatch_log_subscription_filter: Add support for distribution ([#3046](https://github.com/terraform-providers/terraform-provider-aws/issues/3046))
* resource/aws_cognito_user_pool: support pre_token_generation in lambda_config ([#3093](https://github.com/terraform-providers/terraform-provider-aws/issues/3093))
* resource/aws_elasticsearch_domain: Add support for encrypt_at_rest ([#2632](https://github.com/terraform-providers/terraform-provider-aws/issues/2632))
* resource/aws_emr_cluster: Support CustomAmiId ([#2766](https://github.com/terraform-providers/terraform-provider-aws/issues/2766))
* resource/aws_kms_alias: Add target_key_arn attribute ([#3096](https://github.com/terraform-providers/terraform-provider-aws/issues/3096))
* resource/aws_route: Allow adding IPv6 routes to instances and network interfaces ([#2265](https://github.com/terraform-providers/terraform-provider-aws/issues/2265))
* resource/aws_sqs_queue: Retry queue creation on QueueDeletedRecently error ([#3113](https://github.com/terraform-providers/terraform-provider-aws/issues/3113))
* resource/aws_vpn_connection: Add inside CIDR and pre-shared key attributes ([#1862](https://github.com/terraform-providers/terraform-provider-aws/issues/1862))

BUG FIXES:

* resource/aws_appautoscaling_policy: Support additional predefined metric types in validation [[#3122](https://github.com/terraform-providers/terraform-provider-aws/issues/3122)]]
* resource/aws_dynamodb_table: Recognize changes in `non_key_attributes` ([#3136](https://github.com/terraform-providers/terraform-provider-aws/issues/3136))
* resource/aws_ebs_snapshot: Fix `kms_key_id` attribute handling ([#3085](https://github.com/terraform-providers/terraform-provider-aws/issues/3085))
* resource/aws_eip_association: Retry association for pending instances ([#3072](https://github.com/terraform-providers/terraform-provider-aws/issues/3072))
* resource/aws_elastic_beanstalk_application: Prevent crash on reading missing application ([#3171](https://github.com/terraform-providers/terraform-provider-aws/issues/3171))
* resource/aws_kinesis_firehose_delivery_stream: Prevent panic on missing S3 configuration prefix ([#3073](https://github.com/terraform-providers/terraform-provider-aws/issues/3073))
* resource/aws_lambda_function: Retry updates for IAM eventual consistency ([#3116](https://github.com/terraform-providers/terraform-provider-aws/issues/3116))
* resource/aws_route53_record: Suppress uppercase alias name diff ([#3119](https://github.com/terraform-providers/terraform-provider-aws/issues/3119))
* resource/aws_sqs_queue_policy: Prevent missing policy error on read ([#2739](https://github.com/terraform-providers/terraform-provider-aws/issues/2739))
* resource/aws_rds_cluster: Retry deletion on InvalidDBClusterStateFault ([#3028](https://github.com/terraform-providers/terraform-provider-aws/issues/3028))

## 1.7.1 (January 19, 2018)

BUG FIXES:

* data-source/aws_db_snapshot: Prevent crash on unfinished snapshots ([#2960](https://github.com/terraform-providers/terraform-provider-aws/issues/2960))
* resource/aws_cloudfront_distribution: Retry deletion on DistributionNotDisabled ([#3034](https://github.com/terraform-providers/terraform-provider-aws/issues/3034))
* resource/aws_codebuild_project: Prevent crash on empty source buildspec and location ([#3011](https://github.com/terraform-providers/terraform-provider-aws/issues/3011))
* resource/aws_codepipeline: Prevent crash on empty artifacts ([#2998](https://github.com/terraform-providers/terraform-provider-aws/issues/2998))
* resource/aws_appautoscaling_policy: Match correct policy when multiple policies with same name and service ([#3012](https://github.com/terraform-providers/terraform-provider-aws/issues/3012))
* resource/aws_eip: Do not disassociate EIP on tags-only update ([#2975](https://github.com/terraform-providers/terraform-provider-aws/issues/2975))
* resource/aws_elastic_beanstalk_application: Retry DescribeApplication after creation ([#3064](https://github.com/terraform-providers/terraform-provider-aws/issues/3064))
* resource/aws_emr_cluster: Retry creation on `ValidationException` (IAM) ([#3027](https://github.com/terraform-providers/terraform-provider-aws/issues/3027))
* resource/aws_emr_cluster: Retry creation on `AccessDeniedException` (IAM) ([#3050](https://github.com/terraform-providers/terraform-provider-aws/issues/3050))
* resource/aws_iam_instance_profile: Allow cleanup during destruction without refresh ([#2983](https://github.com/terraform-providers/terraform-provider-aws/issues/2983))
* resource/aws_iam_role: Prevent missing attached policy results ([#2857](https://github.com/terraform-providers/terraform-provider-aws/issues/2857))
* resource/aws_iam_user: Prevent state removal during name attribute update ([#2979](https://github.com/terraform-providers/terraform-provider-aws/issues/2979))
* resource/aws_iam_user: Allow path attribute update ([#2940](https://github.com/terraform-providers/terraform-provider-aws/issues/2940))
* resource/aws_iam_user_policy: Fix updates with generated policy names and validate JSON ([#3031](https://github.com/terraform-providers/terraform-provider-aws/issues/3031))
* resource/aws_instance: Retry IAM instance profile (re)association for eventual consistency on update ([#3055](https://github.com/terraform-providers/terraform-provider-aws/issues/3055))
* resource/aws_lambda_function: Make EC2 rate limit errors retryable on update ([#2964](https://github.com/terraform-providers/terraform-provider-aws/issues/2964))
* resource/aws_lambda_function: Retry creation on EC2 throttle error ([#3062](https://github.com/terraform-providers/terraform-provider-aws/issues/3062))
* resource/aws_lb_target_group: Allow a blank health check path, for TCP healthchecks ([#2980](https://github.com/terraform-providers/terraform-provider-aws/issues/2980))
* resource/aws_sns_topic_subscription: Prevent crash on subscription attribute update ([#2967](https://github.com/terraform-providers/terraform-provider-aws/issues/2967))
* resource/aws_kinesis_firehose_delivery_stream: Fix import for S3 destinations ([#2970](https://github.com/terraform-providers/terraform-provider-aws/issues/2970))
* resource/aws_kinesis_firehose_delivery_stream: Prevent crash on empty Redshift's S3 Backup Description ([#2970](https://github.com/terraform-providers/terraform-provider-aws/issues/2970))
* resource/aws_kinesis_firehose_delivery_stream: Detect drifts in `processing_configuration` ([#2970](https://github.com/terraform-providers/terraform-provider-aws/issues/2970))
* resource/aws_kinesis_firehose_delivery_stream: Prevent crash on empty CloudWatch logging opts ([#3052](https://github.com/terraform-providers/terraform-provider-aws/issues/3052))

## 1.7.0 (January 12, 2018)

FEATURES:

* **New Resource:** `aws_api_gateway_documentation_part` ([#2893](https://github.com/terraform-providers/terraform-provider-aws/issues/2893))
* **New Resource:** `aws_cloudwatch_event_permission` ([#2888](https://github.com/terraform-providers/terraform-provider-aws/issues/2888))
* **New Resource:** `aws_cognito_user_pool_client` ([#1803](https://github.com/terraform-providers/terraform-provider-aws/issues/1803))
* **New Resource:** `aws_cognito_user_pool_domain` ([#2325](https://github.com/terraform-providers/terraform-provider-aws/issues/2325))
* **New Resource:** `aws_glue_catalog_database` ([#2175](https://github.com/terraform-providers/terraform-provider-aws/issues/2175))
* **New Resource:** `aws_guardduty_detector` ([#2524](https://github.com/terraform-providers/terraform-provider-aws/issues/2524))
* **New Resource:** `aws_guardduty_member` ([#2911](https://github.com/terraform-providers/terraform-provider-aws/issues/2911))
* **New Resource:** `aws_route53_query_log` ([#2770](https://github.com/terraform-providers/terraform-provider-aws/issues/2770))
* **New Resource:** `aws_service_discovery_service` ([#2613](https://github.com/terraform-providers/terraform-provider-aws/issues/2613))

ENHANCEMENTS:

* provider: `eu-west-3` is now supported ([#2707](https://github.com/terraform-providers/terraform-provider-aws/issues/2707))
* provider: Endpoints can now be specified for ACM, ECR, ECS, STS and Route 53 ([#2795](https://github.com/terraform-providers/terraform-provider-aws/issues/2795))
* provider: Endpoints can now be specified for API Gateway and Lambda ([#2641](https://github.com/terraform-providers/terraform-provider-aws/issues/2641))
* data-source/aws_iam_server_certificate: Add support for retrieving public key ([#2749](https://github.com/terraform-providers/terraform-provider-aws/issues/2749))
* data-source/aws_vpc_peering_connection: Add support for cross-region VPC peering ([#2508](https://github.com/terraform-providers/terraform-provider-aws/issues/2508))
* data-source/aws_ssm_parameter: Support returning raw encrypted SecureString value ([#2777](https://github.com/terraform-providers/terraform-provider-aws/issues/2777))
* resource/aws_kinesis_firehose_delivery_stream: Import is now supported ([#2082](https://github.com/terraform-providers/terraform-provider-aws/issues/2082))
* resource/aws_cognito_user_pool: The ARN for the pool is now computed and exposed as an attribute ([#2723](https://github.com/terraform-providers/terraform-provider-aws/issues/2723))
* resource/aws_directory_service_directory: Add `security_group_id` field ([#2688](https://github.com/terraform-providers/terraform-provider-aws/issues/2688))
* resource/aws_rds_cluster_instance: Support Performance Insights ([#2331](https://github.com/terraform-providers/terraform-provider-aws/issues/2331))
* resource/aws_rds_cluster_instance: Set `db_subnet_group_name` in state on read if available ([#2606](https://github.com/terraform-providers/terraform-provider-aws/issues/2606))
* resource/aws_eip: Tagging is now supported ([#2768](https://github.com/terraform-providers/terraform-provider-aws/issues/2768))
* resource/aws_codepipeline: ARN is now exposed as an attribute ([#2773](https://github.com/terraform-providers/terraform-provider-aws/issues/2773))
* resource/aws_appautoscaling_scheduled_action: `min_capacity` argument is now honoured ([#2794](https://github.com/terraform-providers/terraform-provider-aws/issues/2794))
* resource/aws_rds_cluster: Clusters in the `resetting-master-credentials` state no longer cause an error ([#2791](https://github.com/terraform-providers/terraform-provider-aws/issues/2791))
* resource/aws_cloudwatch_metric_alarm: Support optional datapoints_to_alarm configuration ([#2609](https://github.com/terraform-providers/terraform-provider-aws/issues/2609))
* resource/aws_ses_event_destination: Add support for SNS destinations ([#1737](https://github.com/terraform-providers/terraform-provider-aws/issues/1737))
* resource/aws_iam_role: Delete inline policies when `force_detach_policies = true` ([#2388](https://github.com/terraform-providers/terraform-provider-aws/issues/2388))
* resource/aws_lb_target_group: Improve `health_check` validation ([#2580](https://github.com/terraform-providers/terraform-provider-aws/issues/2580))
* resource/aws_ecs_service: Add `health_check_grace_period_seconds` attribute ([#2788](https://github.com/terraform-providers/terraform-provider-aws/issues/2788))
* resource/aws_vpc_peering_connection: Add support for cross-region VPC peering ([#2508](https://github.com/terraform-providers/terraform-provider-aws/issues/2508))
* resource/aws_vpc_peering_connection_accepter: Add support for cross-region VPC peering ([#2508](https://github.com/terraform-providers/terraform-provider-aws/issues/2508))
* resource/aws_elasticsearch_domain: export kibana endpoint ([#2804](https://github.com/terraform-providers/terraform-provider-aws/issues/2804))
* resource/aws_ssm_association: Allow for multiple targets ([#2297](https://github.com/terraform-providers/terraform-provider-aws/issues/2297))
* resource/aws_instance: Add computed field for volume_id of block device ([#1489](https://github.com/terraform-providers/terraform-provider-aws/issues/1489))
* resource/aws_api_gateway_integration: Allow update of URI attributes ([#2834](https://github.com/terraform-providers/terraform-provider-aws/issues/2834))
* resource/aws_ecs_cluster: Support resource import ([#2762](https://github.com/terraform-providers/terraform-provider-aws/issues/2762))

BUG FIXES:

* resource/aws_cognito_user_pool: Update Cognito email message length to 20,000 ([#2692](https://github.com/terraform-providers/terraform-provider-aws/issues/2692))
* resource/aws_volume_attachment: Changing device name without changing volume or instance ID now correctly produces a diff ([#2720](https://github.com/terraform-providers/terraform-provider-aws/issues/2720))
* resource/aws_s3_bucket_object: Object tagging is now supported in GovCloud ([#2665](https://github.com/terraform-providers/terraform-provider-aws/issues/2665))
* resource/aws_elasticsearch_domain: Fixed a crash when no Cloudwatch log group is configured ([#2787](https://github.com/terraform-providers/terraform-provider-aws/issues/2787))
* resource/aws_s3_bucket_policy: Set the resource ID after successful creation ([#2820](https://github.com/terraform-providers/terraform-provider-aws/issues/2820))
* resource/aws_db_event_subscription: Set the source type when updating categories ([#2833](https://github.com/terraform-providers/terraform-provider-aws/issues/2833))
* resource/aws_db_parameter_group: Remove group from state if it's gone ([#2868](https://github.com/terraform-providers/terraform-provider-aws/issues/2868))
* resource/aws_appautoscaling_target: Make `role_arn` optional & computed ([#2889](https://github.com/terraform-providers/terraform-provider-aws/issues/2889))
* resource/aws_ssm_maintenance_window: Respect `enabled` during updates ([#2818](https://github.com/terraform-providers/terraform-provider-aws/issues/2818))
* resource/aws_lb_target_group: Fix max prefix length check ([#2790](https://github.com/terraform-providers/terraform-provider-aws/issues/2790))
* resource/aws_config_delivery_channel: Retry deletion ([#2910](https://github.com/terraform-providers/terraform-provider-aws/issues/2910))
* resource/aws_lb+aws_elb: Fix regression with undefined `name` ([#2939](https://github.com/terraform-providers/terraform-provider-aws/issues/2939))
* resource/aws_lb_target_group: Fix validation rules for LB's healthcheck ([#2906](https://github.com/terraform-providers/terraform-provider-aws/issues/2906))
* provider: Fix regression affecting empty Optional+Computed fields ([#2348](https://github.com/terraform-providers/terraform-provider-aws/issues/2348))

## 1.6.0 (December 18, 2017)

FEATURES:

* **New Data Source:** `aws_network_interface` ([#2316](https://github.com/terraform-providers/terraform-provider-aws/issues/2316))
* **New Data Source:** `aws_elb` ([#2004](https://github.com/terraform-providers/terraform-provider-aws/issues/2004))
* **New Resource:** `aws_dx_connection_association` ([#2360](https://github.com/terraform-providers/terraform-provider-aws/issues/2360))
* **New Resource:** `aws_appautoscaling_scheduled_action` ([#2231](https://github.com/terraform-providers/terraform-provider-aws/issues/2231))
* **New Resource:** `aws_cloudwatch_log_resource_policy` ([#2243](https://github.com/terraform-providers/terraform-provider-aws/issues/2243))
* **New Resource:** `aws_media_store_container` ([#2448](https://github.com/terraform-providers/terraform-provider-aws/issues/2448))
* **New Resource:** `aws_service_discovery_public_dns_namespace` ([#2569](https://github.com/terraform-providers/terraform-provider-aws/issues/2569))
* **New Resource:** `aws_service_discovery_private_dns_namespace` ([#2589](https://github.com/terraform-providers/terraform-provider-aws/issues/2589))

IMPROVEMENTS:

* resource/aws_ssm_association: Add `association_name` ([#2257](https://github.com/terraform-providers/terraform-provider-aws/issues/2257))
* resource/aws_ecs_service: Add `network_configuration` ([#2299](https://github.com/terraform-providers/terraform-provider-aws/issues/2299))
* resource/aws_lambda_function: Add `reserved_concurrent_executions` ([#2504](https://github.com/terraform-providers/terraform-provider-aws/issues/2504))
* resource/aws_ecs_service: Add `launch_type` (Fargate support) ([#2483](https://github.com/terraform-providers/terraform-provider-aws/issues/2483))
* resource/aws_ecs_task_definition: Add `cpu`, `memory`, `execution_role_arn` & `requires_compatibilities` (Fargate support) ([#2483](https://github.com/terraform-providers/terraform-provider-aws/issues/2483))
* resource/aws_ecs_cluster: Add arn attribute ([#2552](https://github.com/terraform-providers/terraform-provider-aws/issues/2552))
* resource/aws_elasticache_security_group: Add import support ([#2277](https://github.com/terraform-providers/terraform-provider-aws/issues/2277))
* resource/aws_sqs_queue_policy: Support import by queue URL ([#2544](https://github.com/terraform-providers/terraform-provider-aws/issues/2544))
* resource/aws_elasticsearch_domain: Add `log_publishing_options` ([#2285](https://github.com/terraform-providers/terraform-provider-aws/issues/2285))
* resource/aws_athena_database: Add `force_destroy` field ([#2363](https://github.com/terraform-providers/terraform-provider-aws/issues/2363))
* resource/aws_elasticache_replication_group: Add support for Redis auth, in-transit and at-rest encryption ([#2090](https://github.com/terraform-providers/terraform-provider-aws/issues/2090))
* resource/aws_s3_bucket: Add `server_side_encryption_configuration` block ([#2472](https://github.com/terraform-providers/terraform-provider-aws/issues/2472))

BUG FIXES:

* data-source/aws_instance: Set `placement_group` if available ([#2400](https://github.com/terraform-providers/terraform-provider-aws/issues/2400))
* resource/aws_elasticache_parameter_group: Add StateFunc to make name lowercase ([#2426](https://github.com/terraform-providers/terraform-provider-aws/issues/2426))
* resource/aws_elasticache_replication_group: Modify validation, make replication_group_id lowercase ([#2432](https://github.com/terraform-providers/terraform-provider-aws/issues/2432))
* resource/aws_db_instance: Treat `storage-optimization` as valid state ([#2409](https://github.com/terraform-providers/terraform-provider-aws/issues/2409))
* resource/aws_dynamodb_table: Ensure `ttl` is properly read ([#2452](https://github.com/terraform-providers/terraform-provider-aws/issues/2452))
* resource/aws_lb_target_group: fixes to behavior based on protocol type ([#2380](https://github.com/terraform-providers/terraform-provider-aws/issues/2380))
* resource/aws_mq_broker: Fix crash in hashing function ([#2598](https://github.com/terraform-providers/terraform-provider-aws/issues/2598))
* resource/aws_ebs_volume_attachment: Allow attachments to instances which are stopped ([#1444](https://github.com/terraform-providers/terraform-provider-aws/issues/1444))
* resource/aws_ssm_parameter: Path names with a leading '/' no longer generate incorrect ARNs ([#2604](https://github.com/terraform-providers/terraform-provider-aws/issues/2604))

## 1.5.0 (November 29, 2017)

FEATURES:

* **New Resource:** `aws_mq_broker` ([#2466](https://github.com/terraform-providers/terraform-provider-aws/issues/2466))
* **New Resource:** `aws_mq_configuration` ([#2466](https://github.com/terraform-providers/terraform-provider-aws/issues/2466))

## 1.4.0 (November 29, 2017)

BUG FIXES:

* resource/aws_cognito_user_pool: Fix `email_subject_by_link` ([#2395](https://github.com/terraform-providers/terraform-provider-aws/issues/2395))
* resource/aws_api_gateway_method_response: Fix conflict exception in API gateway method response ([#2393](https://github.com/terraform-providers/terraform-provider-aws/issues/2393))
* resource/aws_api_gateway_method: Fix typo `authorization_type` -> `authorization` ([#2430](https://github.com/terraform-providers/terraform-provider-aws/issues/2430))

IMPROVEMENTS:

* data-source/aws_nat_gateway: Add missing address attributes to the schema ([#2209](https://github.com/terraform-providers/terraform-provider-aws/issues/2209))
* resource/aws_ssm_maintenance_window_target: Change MaxItems of targets ([#2361](https://github.com/terraform-providers/terraform-provider-aws/issues/2361))
* resource/aws_sfn_state_machine: Support Update State machine call ([#2349](https://github.com/terraform-providers/terraform-provider-aws/issues/2349))
* resource/aws_instance: Set placement_group in state on read if available ([#2398](https://github.com/terraform-providers/terraform-provider-aws/issues/2398))

## 1.3.1 (November 20, 2017)

BUG FIXES:

* resource/aws_ecs_task_definition: Fix equivalency comparator ([#2339](https://github.com/terraform-providers/terraform-provider-aws/issues/2339))
* resource/aws_batch_job_queue: Return errors correctly if deletion fails ([#2322](https://github.com/terraform-providers/terraform-provider-aws/issues/2322))
* resource/aws_security_group_rule: Parse `description` correctly ([#1959](https://github.com/terraform-providers/terraform-provider-aws/issues/1959))
* Fixed Cognito Lambda Config Validation for optional ARN configurations ([#2370](https://github.com/terraform-providers/terraform-provider-aws/issues/2370))
* resource/aws_cognito_identity_pool_roles_attachment: Fix typo "authenticated" -> "unauthenticated" ([#2358](https://github.com/terraform-providers/terraform-provider-aws/issues/2358))

## 1.3.0 (November 16, 2017)

NOTES:

* resource/aws_redshift_cluster: Field `enable_logging`, `bucket_name` and `s3_key_prefix` were deprecated in favour of a new `logging` block ([#2230](https://github.com/terraform-providers/terraform-provider-aws/issues/2230))
* resource/aws_lb_target_group: We no longer provide defaults for `health_check`'s `path` nor `matcher` in order to support network load balancers where these arguments aren't valid. Creating _new_ ALB will therefore require you to specify these two arguments. Existing deployments are unaffected. ([#2251](https://github.com/terraform-providers/terraform-provider-aws/issues/2251))

FEATURES:

* **New Data Source:** `aws_rds_cluster` ([#2070](https://github.com/terraform-providers/terraform-provider-aws/issues/2070))
* **New Data Source:** `aws_elasticache_replication_group` ([#2124](https://github.com/terraform-providers/terraform-provider-aws/issues/2124))
* **New Data Source:** `aws_instances` ([#2266](https://github.com/terraform-providers/terraform-provider-aws/issues/2266))
* **New Resource:** `aws_ses_template` ([#2003](https://github.com/terraform-providers/terraform-provider-aws/issues/2003))
* **New Resource:** `aws_dx_lag` ([#2154](https://github.com/terraform-providers/terraform-provider-aws/issues/2154))
* **New Resource:** `aws_dx_connection` ([#2173](https://github.com/terraform-providers/terraform-provider-aws/issues/2173))
* **New Resource:** `aws_athena_database` ([#1922](https://github.com/terraform-providers/terraform-provider-aws/issues/1922))
* **New Resource:** `aws_athena_named_query` ([#1893](https://github.com/terraform-providers/terraform-provider-aws/issues/1893))
* **New Resource:** `aws_ssm_resource_data_sync` ([#1895](https://github.com/terraform-providers/terraform-provider-aws/issues/1895))
* **New Resource:** `aws_cognito_user_pool` ([#1419](https://github.com/terraform-providers/terraform-provider-aws/issues/1419))

IMPROVEMENTS:

* provider: Add support for assuming roles via profiles defined in `~/.aws/config` ([#1608](https://github.com/terraform-providers/terraform-provider-aws/issues/1608))
* data-source/efs_file_system: Added dns_name ([#2105](https://github.com/terraform-providers/terraform-provider-aws/issues/2105))
* data-source/aws_ssm_parameter: Add `arn` attribute ([#2273](https://github.com/terraform-providers/terraform-provider-aws/issues/2273))
* data-source/aws_ebs_volume: Add `arn` attribute ([#2271](https://github.com/terraform-providers/terraform-provider-aws/issues/2271))
* resource/aws_batch_job_queue: Add validation for `name` ([#2159](https://github.com/terraform-providers/terraform-provider-aws/issues/2159))
* resource/aws_batch_compute_environment: Improve validation for `compute_environment_name` ([#2159](https://github.com/terraform-providers/terraform-provider-aws/issues/2159))
* resource/aws_ssm_parameter: Add support for import ([#2234](https://github.com/terraform-providers/terraform-provider-aws/issues/2234))
* resource/aws_redshift_cluster: Add support for `snapshot_copy` ([#2238](https://github.com/terraform-providers/terraform-provider-aws/issues/2238))
* resource/aws_ecs_task_definition: Print `container_definitions` as JSON instead of checksum ([#1195](https://github.com/terraform-providers/terraform-provider-aws/issues/1195))
* resource/aws_ssm_parameter: Add `arn` attribute ([#2273](https://github.com/terraform-providers/terraform-provider-aws/issues/2273))
* resource/aws_elb: Add listener `ssl_certificate_id` ARN validation ([#2276](https://github.com/terraform-providers/terraform-provider-aws/issues/2276))
* resource/aws_cloudformation_stack: Support updating `tags` ([#2262](https://github.com/terraform-providers/terraform-provider-aws/issues/2262))
* resource/aws_elb: Add `arn` attribute ([#2272](https://github.com/terraform-providers/terraform-provider-aws/issues/2272))
* resource/aws_ebs_volume: Add `arn` attribute ([#2271](https://github.com/terraform-providers/terraform-provider-aws/issues/2271))

BUG FIXES:

* resource/aws_appautoscaling_policy: Retry putting policy on invalid token ([#2135](https://github.com/terraform-providers/terraform-provider-aws/issues/2135))
* resource/aws_batch_compute_environment: `compute_environment_name` allows hyphens ([#2126](https://github.com/terraform-providers/terraform-provider-aws/issues/2126))
* resource/aws_batch_job_definition: `name` allows hyphens ([#2126](https://github.com/terraform-providers/terraform-provider-aws/issues/2126))
* resource/aws_elasticache_parameter_group: Raise timeout for retry on pending changes ([#2134](https://github.com/terraform-providers/terraform-provider-aws/issues/2134))
* resource/aws_kms_key: Retry GetKeyRotationStatus on NotFoundException ([#2133](https://github.com/terraform-providers/terraform-provider-aws/issues/2133))
* resource/aws_lb_target_group: Fix issue that prevented using `aws_lb_target_group` with
  Network type load balancers ([#2251](https://github.com/terraform-providers/terraform-provider-aws/issues/2251))
* resource/aws_lb: mark subnets as `ForceNew` for network load balancers ([#2310](https://github.com/terraform-providers/terraform-provider-aws/issues/2310))
* resource/aws_redshift_cluster: Make master_username ForceNew ([#2202](https://github.com/terraform-providers/terraform-provider-aws/issues/2202))
* resource/aws_cloudwatch_log_metric_filter: Fix pattern length check ([#2107](https://github.com/terraform-providers/terraform-provider-aws/issues/2107))
* resource/aws_cloudwatch_log_group: Use ID as name ([#2190](https://github.com/terraform-providers/terraform-provider-aws/issues/2190))
* resource/aws_elasticsearch_domain: Added ForceNew to vpc_options ([#2157](https://github.com/terraform-providers/terraform-provider-aws/issues/2157))
* resource/aws_redshift_cluster: Make snapshot identifiers `ForceNew` ([#2212](https://github.com/terraform-providers/terraform-provider-aws/issues/2212))
* resource/aws_elasticsearch_domain_policy: Fix typo in err code ([#2249](https://github.com/terraform-providers/terraform-provider-aws/issues/2249))
* resource/aws_appautoscaling_policy: Retry PutScalingPolicy on rate exceeded message ([#2275](https://github.com/terraform-providers/terraform-provider-aws/issues/2275))
* resource/aws_dynamodb_table: Retry creation on `LimitExceededException` w/ different error message ([#2274](https://github.com/terraform-providers/terraform-provider-aws/issues/2274))

## 1.2.0 (October 31, 2017)

INTERNAL:

* Remove `id` fields from schema definitions ([#1626](https://github.com/terraform-providers/terraform-provider-aws/issues/1626))

FEATURES:

* **New Resource:** `aws_servicecatalog_portfolio` ([#1694](https://github.com/terraform-providers/terraform-provider-aws/issues/1694))
* **New Resource:** `aws_ses_domain_dkim` ([#1786](https://github.com/terraform-providers/terraform-provider-aws/issues/1786))
* **New Resource:** `aws_cognito_identity_pool_roles_attachment` ([#863](https://github.com/terraform-providers/terraform-provider-aws/issues/863))
* **New Resource:** `aws_ecr_lifecycle_policy` ([#2096](https://github.com/terraform-providers/terraform-provider-aws/issues/2096))
* **New Data Source:** `aws_nat_gateway` ([#1294](https://github.com/terraform-providers/terraform-provider-aws/issues/1294))
* **New Data Source:** `aws_dynamodb_table` ([#2062](https://github.com/terraform-providers/terraform-provider-aws/issues/2062))
* **New Data Source:** `aws_cloudtrail_service_account` ([#1774](https://github.com/terraform-providers/terraform-provider-aws/issues/1774))

IMPROVEMENTS:

* resource/aws_ami: Support configurable timeouts ([#1811](https://github.com/terraform-providers/terraform-provider-aws/issues/1811))
* resource/ami_copy: Support configurable timeouts ([#1811](https://github.com/terraform-providers/terraform-provider-aws/issues/1811))
* resource/ami_from_instance: Support configurable timeouts ([#1811](https://github.com/terraform-providers/terraform-provider-aws/issues/1811))
* data-source/aws_security_group: add description ([#1943](https://github.com/terraform-providers/terraform-provider-aws/issues/1943))
* resource/aws_cloudfront_distribution: Change the default minimum_protocol_version to TLSv1 ([#1856](https://github.com/terraform-providers/terraform-provider-aws/issues/1856))
* resource/aws_sns_topic: Support SMS in protocols ([#1813](https://github.com/terraform-providers/terraform-provider-aws/issues/1813))
* resource/aws_spot_fleet_request: Add support for `tags` ([#2042](https://github.com/terraform-providers/terraform-provider-aws/issues/2042))
* resource/aws_kinesis_firehose_delivery_stream: Add `s3_backup_mode` option ([#1830](https://github.com/terraform-providers/terraform-provider-aws/issues/1830))
* resource/aws_elasticsearch_domain: Support VPC configuration ([#1958](https://github.com/terraform-providers/terraform-provider-aws/issues/1958))
* resource/aws_alb_target_group: Add support for `target_type` ([#1589](https://github.com/terraform-providers/terraform-provider-aws/issues/1589))
* resource/aws_sqs_queue: Add support for `tags` ([#1987](https://github.com/terraform-providers/terraform-provider-aws/issues/1987))
* resource/aws_security_group: Add `revoke_rules_on_delete` option to force a security group to revoke
  rules before deleting the grou ([#2074](https://github.com/terraform-providers/terraform-provider-aws/issues/2074))
* resource/aws_cloudwatch_log_metric_filter: Add support for DefaultValue ([#1578](https://github.com/terraform-providers/terraform-provider-aws/issues/1578))
* resource/aws_emr_cluster: Expose error on `TERMINATED_WITH_ERRORS` ([#2081](https://github.com/terraform-providers/terraform-provider-aws/issues/2081))

BUG FIXES:

* resource/aws_elasticache_parameter_group: Add missing return to retry logic ([#1891](https://github.com/terraform-providers/terraform-provider-aws/issues/1891))
* resource/aws_batch_job_queue: Wait for update completion when disabling ([#1892](https://github.com/terraform-providers/terraform-provider-aws/issues/1892))
* resource/aws_snapshot_create_volume_permission: Raise creation timeout to 10mins ([#1894](https://github.com/terraform-providers/terraform-provider-aws/issues/1894))
* resource/aws_snapshot_create_volume_permission: Raise creation timeout to 20mins ([#2049](https://github.com/terraform-providers/terraform-provider-aws/issues/2049))
* resource/aws_kms_alias: Retry creation on `NotFoundException` ([#1896](https://github.com/terraform-providers/terraform-provider-aws/issues/1896))
* resource/aws_kms_key: Retry reading tags on `NotFoundException` ([#1900](https://github.com/terraform-providers/terraform-provider-aws/issues/1900))
* resource/aws_db_snapshot: Raise creation timeout to 20mins ([#1905](https://github.com/terraform-providers/terraform-provider-aws/issues/1905))
* resource/aws_lb: Allow assigning EIP to network LB ([#1956](https://github.com/terraform-providers/terraform-provider-aws/issues/1956))
* resource/aws_s3_bucket: Retry tagging on OperationAborted ([#2008](https://github.com/terraform-providers/terraform-provider-aws/issues/2008))
* resource/aws_cognito_identity_pool: Fixed refresh of providers ([#2015](https://github.com/terraform-providers/terraform-provider-aws/issues/2015))
* resource/aws_elasticache_replication_group: Raise creation timeout to 50mins ([#2048](https://github.com/terraform-providers/terraform-provider-aws/issues/2048))
* resource/aws_api_gateway_usag_plan: Fixed setting of rate_limit ([#2076](https://github.com/terraform-providers/terraform-provider-aws/issues/2076))
* resource/aws_elastic_beanstalk_application: Expose error leading to failed deletion ([#2080](https://github.com/terraform-providers/terraform-provider-aws/issues/2080))
* resource/aws_s3_bucket: Accept query strings in redirect hosts ([#2059](https://github.com/terraform-providers/terraform-provider-aws/issues/2059))

## 1.1.0 (October 16, 2017)

NOTES:

* resource/aws_alb_* & data-source/aws_alb_*: In order to support network LBs, ALBs were renamed to `aws_lb_*` due to the way APIs "new" (non-Classic) load balancers are structured in AWS. All existing ALB functionality remains untouched and new resources work the same way. `aws_alb_*` resources are still in place as "aliases", but documentation will only mention `aws_lb_*`.
`aws_alb_*` aliases will be removed in future major version. ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* Deprecated:
  * data-source/aws_alb
  * data-source/aws_alb_listener
  * data-source/aws_alb_target_group
  * resource/aws_alb
  * resource/aws_alb_listener
  * resource/aws_alb_listener_rule
  * resource/aws_alb_target_group
  * resource/aws_alb_target_group_attachment

FEATURES:

* **New Resource:** `aws_batch_job_definition` ([#1710](https://github.com/terraform-providers/terraform-provider-aws/issues/1710))
* **New Resource:** `aws_batch_job_queue` ([#1710](https://github.com/terraform-providers/terraform-provider-aws/issues/1710))
* **New Resource:** `aws_lb` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Resource:** `aws_lb_listener` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Resource:** `aws_lb_listener_rule` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Resource:** `aws_lb_target_group` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Resource:** `aws_lb_target_group_attachment` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Data Source:** `aws_lb` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Data Source:** `aws_lb_listener` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Data Source:** `aws_lb_target_group` ([#1806](https://github.com/terraform-providers/terraform-provider-aws/issues/1806))
* **New Data Source:** `aws_iam_user` ([#1805](https://github.com/terraform-providers/terraform-provider-aws/issues/1805))
* **New Data Source:** `aws_s3_bucket` ([#1505](https://github.com/terraform-providers/terraform-provider-aws/issues/1505))

IMPROVEMENTS:

* data-source/aws_redshift_service_account: Add `arn` attribute ([#1775](https://github.com/terraform-providers/terraform-provider-aws/issues/1775))
* data-source/aws_vpc_endpoint: Expose `prefix_list_id` ([#1733](https://github.com/terraform-providers/terraform-provider-aws/issues/1733))
* resource/aws_kinesis_stream: Add support for encryption ([#1139](https://github.com/terraform-providers/terraform-provider-aws/issues/1139))
* resource/aws_cloudwatch_log_group: Add support for encryption via `kms_key_id` ([#1751](https://github.com/terraform-providers/terraform-provider-aws/issues/1751))
* resource/aws_spot_instance_request: Add support for `instance_interruption_behaviour` ([#1735](https://github.com/terraform-providers/terraform-provider-aws/issues/1735))
* resource/aws_ses_event_destination: Add support for `open` & `click` event types ([#1773](https://github.com/terraform-providers/terraform-provider-aws/issues/1773))
* resource/aws_efs_file_system: Expose `dns_name` ([#1825](https://github.com/terraform-providers/terraform-provider-aws/issues/1825))
* resource/aws_security_group+aws_security_group_rule: Add support for rule description ([#1587](https://github.com/terraform-providers/terraform-provider-aws/issues/1587))
* resource/aws_emr_cluster: enable configuration of ebs root volume size ([#1375](https://github.com/terraform-providers/terraform-provider-aws/issues/1375))
* resource/aws_ami: Add `root_snapshot_id` attribute ([#1572](https://github.com/terraform-providers/terraform-provider-aws/issues/1572))
* resource/aws_vpn_connection: Mark preshared keys as sensitive ([#1850](https://github.com/terraform-providers/terraform-provider-aws/issues/1850))
* resource/aws_codedeploy_deployment_group: Support blue/green and in-place deployments with traffic control ([#1162](https://github.com/terraform-providers/terraform-provider-aws/issues/1162))
* resource/aws_elb: Update ELB idle timeout to 4000s ([#1861](https://github.com/terraform-providers/terraform-provider-aws/issues/1861))
* resource/aws_spot_fleet_request: Add support for instance_interruption_behaviour ([#1847](https://github.com/terraform-providers/terraform-provider-aws/issues/1847))
* resource/aws_kinesis_firehose_delivery_stream: Specify kinesis stream as the source of a aws_kinesis_firehose_delivery_stream ([#1605](https://github.com/terraform-providers/terraform-provider-aws/issues/1605))
* resource/aws_kinesis_firehose_delivery_stream: Output complete error when creation fails ([#1881](https://github.com/terraform-providers/terraform-provider-aws/issues/1881))

BUG FIXES:

* data-source/aws_db_instance: Make `db_instance_arn` expose ARN instead of identifier (use `db_cluster_identifier` for identifier) ([#1766](https://github.com/terraform-providers/terraform-provider-aws/issues/1766))
* data-source/aws_db_snapshot: Expose `storage_type` (was not exposed) ([#1833](https://github.com/terraform-providers/terraform-provider-aws/issues/1833))
* data-source/aws_ami: Update the `tags` structure for easier referencing ([#1706](https://github.com/terraform-providers/terraform-provider-aws/issues/1706))
* data-source/aws_ebs_snapshot: Update the `tags` structure for easier referencing ([#1706](https://github.com/terraform-providers/terraform-provider-aws/issues/1706))
* data-source/aws_ebs_volume: Update the `tags` structure for easier referencing ([#1706](https://github.com/terraform-providers/terraform-provider-aws/issues/1706))
* data-source/aws_instance: Update the `tags` structure for easier referencing ([#1706](https://github.com/terraform-providers/terraform-provider-aws/issues/1706))
* resource/aws_spot_instance_request: Handle `closed` request correctly ([#1903](https://github.com/terraform-providers/terraform-provider-aws/issues/1903))
* resource/aws_cloudtrail: Raise update retry timeout ([#1820](https://github.com/terraform-providers/terraform-provider-aws/issues/1820))
* resource/aws_elasticache_parameter_group: Retry resetting group on pending changes ([#1821](https://github.com/terraform-providers/terraform-provider-aws/issues/1821))
* resource/aws_kms_key: Retry getting rotation status ([#1818](https://github.com/terraform-providers/terraform-provider-aws/issues/1818))
* resource/aws_kms_key: Retry getting key policy ([#1854](https://github.com/terraform-providers/terraform-provider-aws/issues/1854))
* resource/aws_vpn_connection: Raise timeout to 40mins ([#1819](https://github.com/terraform-providers/terraform-provider-aws/issues/1819))
* resource/aws_kinesis_firehose_delivery_stream: Fix crash caused by missing `processing_configuration` ([#1738](https://github.com/terraform-providers/terraform-provider-aws/issues/1738))
* resource/aws_rds_cluster_instance: Treat `configuring-enhanced-monitoring` as pending state ([#1744](https://github.com/terraform-providers/terraform-provider-aws/issues/1744))
* resource/aws_rds_cluster_instance: Treat more states as pending ([#1790](https://github.com/terraform-providers/terraform-provider-aws/issues/1790))
* resource/aws_route_table: Increase number of not-found checks/retries after creation ([#1791](https://github.com/terraform-providers/terraform-provider-aws/issues/1791))
* resource/aws_batch_compute_environment: Fix ARN attribute name/value (`ecc_cluster_arn` -> `ecs_cluster_arn`) ([#1809](https://github.com/terraform-providers/terraform-provider-aws/issues/1809))
* resource/aws_kinesis_stream: Retry creation of the stream on `LimitExceededException` (handle throttling) ([#1339](https://github.com/terraform-providers/terraform-provider-aws/issues/1339))
* resource/aws_vpn_connection_route: Treat route in state `deleted` as deleted ([#1848](https://github.com/terraform-providers/terraform-provider-aws/issues/1848))
* resource/aws_eip: Avoid disassociating if there's no association ([#1683](https://github.com/terraform-providers/terraform-provider-aws/issues/1683))
* resource/aws_elasticache_cluster: Allow scaling up cluster by modifying `az_mode` (avoid recreation) ([#1758](https://github.com/terraform-providers/terraform-provider-aws/issues/1758))
* resource/aws_lambda_function: Fix Lambda Function Updates When Published ([#1797](https://github.com/terraform-providers/terraform-provider-aws/issues/1797))
* resource/aws_appautoscaling_*: Use dimension to uniquely identify target/policy ([#1808](https://github.com/terraform-providers/terraform-provider-aws/issues/1808))
* resource/aws_vpn_connection_route: Wait until route is available/deleted ([#1849](https://github.com/terraform-providers/terraform-provider-aws/issues/1849))
* resource/aws_cloudfront_distribution: Ignore `minimum_protocol_version` if default certificate is used ([#1785](https://github.com/terraform-providers/terraform-provider-aws/issues/1785))
* resource/aws_security_group: Using `self = false` with `cidr_blocks` should be allowed ([#1839](https://github.com/terraform-providers/terraform-provider-aws/issues/1839))
* resource/aws_instance: Check VPC array size to avoid crashes on Eucalyptus Cloud ([#1882](https://github.com/terraform-providers/terraform-provider-aws/issues/1882))

## 1.0.0 (September 27, 2017)

NOTES:

* resource/aws_appautoscaling_policy: Nest step scaling policy fields, deprecate 1st level fields ([#1620](https://github.com/terraform-providers/terraform-provider-aws/issues/1620))

FEATURES:

* **New Resource:** `aws_waf_rate_based_rule` ([#1606](https://github.com/terraform-providers/terraform-provider-aws/issues/1606))
* **New Resource:** `aws_batch_compute_environment` ([#1048](https://github.com/terraform-providers/terraform-provider-aws/issues/1048))

IMPROVEMENTS:

* provider: Expand shared_credentials_file ([#1511](https://github.com/terraform-providers/terraform-provider-aws/issues/1511))
* provider: Add support for Task Roles when running on ECS or CodeBuild ([#1425](https://github.com/terraform-providers/terraform-provider-aws/issues/1425))
* resource/aws_instance: New `user_data_base64` attribute that allows non-UTF8 data (such as gzip) to be assigned to user-data without corruption ([#850](https://github.com/terraform-providers/terraform-provider-aws/issues/850))
* data-source/aws_vpc: Expose enable_dns_* in aws_vpc data_source ([#1373](https://github.com/terraform-providers/terraform-provider-aws/issues/1373))
* resource/aws_appautoscaling_policy: Add support for DynamoDB ([#1650](https://github.com/terraform-providers/terraform-provider-aws/issues/1650))
* resource/aws_directory_service_directory: Add support for `tags` ([#1398](https://github.com/terraform-providers/terraform-provider-aws/issues/1398))
* resource/aws_rds_cluster: Allow setting of rds cluster engine ([#1415](https://github.com/terraform-providers/terraform-provider-aws/issues/1415))
* resource/aws_ssm_association: now supports update for `parameters`, `schedule_expression`,`output_location` ([#1421](https://github.com/terraform-providers/terraform-provider-aws/issues/1421))
* resource/aws_ssm_patch_baseline: now supports update for multiple attributes ([#1421](https://github.com/terraform-providers/terraform-provider-aws/issues/1421))
* resource/aws_cloudformation_stack: Add support for Import ([#1432](https://github.com/terraform-providers/terraform-provider-aws/issues/1432))
* resource/aws_rds_cluster_instance: Expose availability_zone attribute ([#1439](https://github.com/terraform-providers/terraform-provider-aws/issues/1439))
* resource/aws_efs_file_system: Add support for encryption ([#1420](https://github.com/terraform-providers/terraform-provider-aws/issues/1420))
* resource/aws_db_parameter_group: Allow underscores in names ([#1460](https://github.com/terraform-providers/terraform-provider-aws/issues/1460))
* resource/aws_elasticsearch_domain: Assign tags right after creation ([#1399](https://github.com/terraform-providers/terraform-provider-aws/issues/1399))
* resource/aws_route53_record: Allow CAA record type ([#1467](https://github.com/terraform-providers/terraform-provider-aws/issues/1467))
* resource/aws_codebuild_project: Allowed for BITBUCKET source type ([#1468](https://github.com/terraform-providers/terraform-provider-aws/issues/1468))
* resource/aws_emr_cluster: Add `instance_group` parameter for EMR clusters ([#1071](https://github.com/terraform-providers/terraform-provider-aws/issues/1071))
* resource/aws_alb_listener_rule: Populate `listener_arn` field ([#1303](https://github.com/terraform-providers/terraform-provider-aws/issues/1303))
* resource/aws_api_gateway_rest_api: Add a body property to API Gateway RestAPI for Swagger import support ([#1197](https://github.com/terraform-providers/terraform-provider-aws/issues/1197))
* resource/aws_opsworks_stack: Add support for tags ([#1523](https://github.com/terraform-providers/terraform-provider-aws/issues/1523))
* Add retries for AppScaling policies throttling exceptions ([#1430](https://github.com/terraform-providers/terraform-provider-aws/issues/1430))
* resource/aws_ssm_patch_baseline: Add compliance level to patch approval rules ([#1531](https://github.com/terraform-providers/terraform-provider-aws/issues/1531))
* resource/aws_ssm_activation: Export ssm activation activation_code ([#1570](https://github.com/terraform-providers/terraform-provider-aws/issues/1570))
* resource/aws_network_interface: Added private_dns_name to network_interface ([#1599](https://github.com/terraform-providers/terraform-provider-aws/issues/1599))
* data-source/aws_redshift_service_account: updated with latest redshift service account ID's ([#1614](https://github.com/terraform-providers/terraform-provider-aws/issues/1614))
* resource/aws_ssm_parameter: Refresh from state on 404 ([#1436](https://github.com/terraform-providers/terraform-provider-aws/issues/1436))
* resource/aws_api_gateway_rest_api: Allow binary media types to be updated ([#1600](https://github.com/terraform-providers/terraform-provider-aws/issues/1600))
* resource/aws_waf_rule: Make `predicates`' `data_id` required (it always was on the API's side, it's just reflected in the schema) ([#1606](https://github.com/terraform-providers/terraform-provider-aws/issues/1606))
* resource/aws_waf_web_acl: Introduce new `type` field in `rules` to allow referencing `RATE_BASED` type ([#1606](https://github.com/terraform-providers/terraform-provider-aws/issues/1606))
* resource/aws_ssm_association: Migrate the schema to use association_id ([#1579](https://github.com/terraform-providers/terraform-provider-aws/issues/1579))
* resource/aws_ssm_document: Added name validation ([#1638](https://github.com/terraform-providers/terraform-provider-aws/issues/1638))
* resource/aws_nat_gateway: Add tags support ([#1625](https://github.com/terraform-providers/terraform-provider-aws/issues/1625))
* resource/aws_route53_record: Add support for Route53 multi-value answer routing policy ([#1686](https://github.com/terraform-providers/terraform-provider-aws/issues/1686))
* resource/aws_instance: Read iops only when volume type is io1 ([#1573](https://github.com/terraform-providers/terraform-provider-aws/issues/1573))
* resource/aws_rds_cluster(+_instance) Allow specifying the engine ([#1591](https://github.com/terraform-providers/terraform-provider-aws/issues/1591))
* resource/aws_cloudwatch_event_target: Add Input transformer for Cloudwatch Events ([#1343](https://github.com/terraform-providers/terraform-provider-aws/issues/1343))
* resource/aws_directory_service_directory: Support Import functionality ([#1732](https://github.com/terraform-providers/terraform-provider-aws/issues/1732))

BUG FIXES:

* resource/aws_instance: Fix `associate_public_ip_address` ([#1340](https://github.com/terraform-providers/terraform-provider-aws/issues/1340))
* resource/aws_instance: Fix import in EC2 Classic ([#1453](https://github.com/terraform-providers/terraform-provider-aws/issues/1453))
* resource/aws_emr_cluster: Avoid spurious diff of `log_uri` ([#1374](https://github.com/terraform-providers/terraform-provider-aws/issues/1374))
* resource/aws_cloudwatch_log_subscription_filter: Add support for ResourceNotFound ([#1414](https://github.com/terraform-providers/terraform-provider-aws/issues/1414))
* resource/aws_sns_topic_subscription: Prevent duplicate (un)subscribe during initial creation ([#1480](https://github.com/terraform-providers/terraform-provider-aws/issues/1480))
* resource/aws_alb: Cleanup ENIs after deleting ALB ([#1427](https://github.com/terraform-providers/terraform-provider-aws/issues/1427))
* resource/aws_s3_bucket: Wrap s3 calls in retry to avoid race during creation ([#891](https://github.com/terraform-providers/terraform-provider-aws/issues/891))
* resource/aws_eip: Remove from state on deletion ([#1551](https://github.com/terraform-providers/terraform-provider-aws/issues/1551))
* resource/aws_security_group: Adding second scenario where IPv6 is not supported ([#880](https://github.com/terraform-providers/terraform-provider-aws/issues/880))

## 0.1.4 (August 08, 2017)

FEATURES:

* **New Resource:** `aws_cloudwatch_dashboard` ([#1172](https://github.com/terraform-providers/terraform-provider-aws/issues/1172))
* **New Data Source:** `aws_internet_gateway` ([#1196](https://github.com/terraform-providers/terraform-provider-aws/issues/1196))
* **New Data Source:** `aws_efs_mount_target` ([#1255](https://github.com/terraform-providers/terraform-provider-aws/issues/1255))

IMPROVEMENTS:

* AWS SDK to log extra debug details on request errors ([#1210](https://github.com/terraform-providers/terraform-provider-aws/issues/1210))
* resource/aws_spot_fleet_request: Add support for  `wait_for_fulfillment` ([#1241](https://github.com/terraform-providers/terraform-provider-aws/issues/1241))
* resource/aws_autoscaling_schedule: Allow empty value ([#1268](https://github.com/terraform-providers/terraform-provider-aws/issues/1268))
* resource/aws_ssm_association: Add support for OutputLocation and Schedule Expression ([#1253](https://github.com/terraform-providers/terraform-provider-aws/issues/1253))
* resource/aws_ssm_patch_baseline: Update support for Operating System ([#1260](https://github.com/terraform-providers/terraform-provider-aws/issues/1260))
* resource/aws_db_instance: Expose db_instance ca_cert_identifier ([#1256](https://github.com/terraform-providers/terraform-provider-aws/issues/1256))
* resource/aws_rds_cluster: Add support for iam_roles to rds_cluster ([#1258](https://github.com/terraform-providers/terraform-provider-aws/issues/1258))
* resource/aws_rds_cluster_parameter_group: Support > 20 parameters ([#1298](https://github.com/terraform-providers/terraform-provider-aws/issues/1298))
* data-source/aws_iam_role: Normalize the IAM role data source ([#1330](https://github.com/terraform-providers/terraform-provider-aws/issues/1330))
* resource/aws_kinesis_stream: Increase Timeouts, add Timeout Support ([#1345](https://github.com/terraform-providers/terraform-provider-aws/issues/1345))

BUG FIXES:

* resource/aws_instance: Guard check for aws_instance UserData to prevent panic ([#1288](https://github.com/terraform-providers/terraform-provider-aws/issues/1288))
* resource/aws_config: Set AWS Config Configuration recorder & Delivery channel names as ForceNew ([#1247](https://github.com/terraform-providers/terraform-provider-aws/issues/1247))
* resource/aws_cloudtrail: Retry if IAM role isn't propagated yet ([#1312](https://github.com/terraform-providers/terraform-provider-aws/issues/1312))
* resource/aws_cloudtrail: Fix CloudWatch role ARN/group updates ([#1357](https://github.com/terraform-providers/terraform-provider-aws/issues/1357))
* resource/aws_eip_association: Avoid crash in EC2 Classic ([#1344](https://github.com/terraform-providers/terraform-provider-aws/issues/1344))
* resource/aws_elasticache_parameter_group: Allow removing parameters ([#1309](https://github.com/terraform-providers/terraform-provider-aws/issues/1309))
* resource/aws_kinesis: add retries for Kinesis throttling exceptions ([#1085](https://github.com/terraform-providers/terraform-provider-aws/issues/1085))
* resource/aws_kinesis_firehose: adding support for `ExtendedS3DestinationConfiguration` ([#1015](https://github.com/terraform-providers/terraform-provider-aws/issues/1015))
* resource/aws_spot_fleet_request: Ignore empty `key_name` ([#1203](https://github.com/terraform-providers/terraform-provider-aws/issues/1203))
* resource/aws_emr_instance_group: fix crash when changing `instance_group.count` ([#1287](https://github.com/terraform-providers/terraform-provider-aws/issues/1287))
* resource/aws_elasticsearch_domain: Fix updating config when update doesn't involve EBS ([#1131](https://github.com/terraform-providers/terraform-provider-aws/issues/1131))
* resource/aws_s3_bucket: Avoid crashing when no lifecycle rule is defined ([#1316](https://github.com/terraform-providers/terraform-provider-aws/issues/1316))
* resource/elastic_transcoder_preset: Fix provider validation ([#1338](https://github.com/terraform-providers/terraform-provider-aws/issues/1338))
* resource/aws_s3_bucket: Avoid crashing when `filter` is not set ([#1350](https://github.com/terraform-providers/terraform-provider-aws/issues/1350))

## 0.1.3 (July 25, 2017)

FEATURES:

* **New Data Source:** `aws_iam_instance_profile` ([#1024](https://github.com/terraform-providers/terraform-provider-aws/issues/1024))
* **New Data Source:** `aws_alb_target_group` ([#1037](https://github.com/terraform-providers/terraform-provider-aws/issues/1037))
* **New Data Source:** `aws_iam_group` ([#1140](https://github.com/terraform-providers/terraform-provider-aws/issues/1140))
* **New Resource:** `aws_api_gateway_request_validator` ([#1064](https://github.com/terraform-providers/terraform-provider-aws/issues/1064))
* **New Resource:** `aws_api_gateway_gateway_response` ([#1168](https://github.com/terraform-providers/terraform-provider-aws/issues/1168))
* **New Resource:** `aws_iot_policy` ([#986](https://github.com/terraform-providers/terraform-provider-aws/issues/986))
* **New Resource:** `aws_iot_certificate` ([#1225](https://github.com/terraform-providers/terraform-provider-aws/issues/1225))

IMPROVEMENTS:

* resource/aws_sqs_queue: Add support for Server-Side Encryption ([#962](https://github.com/terraform-providers/terraform-provider-aws/issues/962))
* resource/aws_vpc: Add support for classiclink_dns_support ([#1079](https://github.com/terraform-providers/terraform-provider-aws/issues/1079))
* resource/aws_lambda_function: Add support for lambda_function vpc_config update ([#1080](https://github.com/terraform-providers/terraform-provider-aws/issues/1080))
* resource/aws_lambda_function: Add support for lambda_function dead_letter_config update ([#1080](https://github.com/terraform-providers/terraform-provider-aws/issues/1080))
* resource/aws_route53_health_check: add support for health_check regions ([#1116](https://github.com/terraform-providers/terraform-provider-aws/issues/1116))
* resource/aws_spot_instance_request: add support for request launch group ([#1097](https://github.com/terraform-providers/terraform-provider-aws/issues/1097))
* resource/aws_rds_cluster_instance: Export the RDI Resource ID for the instance ([#1142](https://github.com/terraform-providers/terraform-provider-aws/issues/1142))
* resource/aws_sns_topic_subscription: Support password-protected HTTPS endpoints ([#861](https://github.com/terraform-providers/terraform-provider-aws/issues/861))

BUG FIXES:

* provider: Remove assumeRoleHash ([#1227](https://github.com/terraform-providers/terraform-provider-aws/issues/1227))
* resource/aws_ami: Retry on `InvalidAMIID.NotFound` ([#1035](https://github.com/terraform-providers/terraform-provider-aws/issues/1035))
* resource/aws_iam_server_certificate: Fix restriction on length of `name_prefix` ([#1217](https://github.com/terraform-providers/terraform-provider-aws/issues/1217))
* resource/aws_autoscaling_group: Fix handling of empty `vpc_zone_identifier` (EC2 classic & default VPC) ([#1191](https://github.com/terraform-providers/terraform-provider-aws/issues/1191))
* resource/aws_ecr_repository_policy: Add retry logic to work around IAM eventual consistency ([#1165](https://github.com/terraform-providers/terraform-provider-aws/issues/1165))
* resource/aws_ecs_service: Fixes normalization issues in placement_strategy ([#1025](https://github.com/terraform-providers/terraform-provider-aws/issues/1025))
* resource/aws_eip: Retry reading EIPs on creation ([#1053](https://github.com/terraform-providers/terraform-provider-aws/issues/1053))
* resource/aws_elastic_beanstalk_environment: Avoid spurious diffs of JSON-based `setting`s ([#901](https://github.com/terraform-providers/terraform-provider-aws/issues/901))
* resource/aws_opsworks_permission: Fix 'set permissions' failing to set ssh access ([#1038](https://github.com/terraform-providers/terraform-provider-aws/issues/1038))
* resource/aws_s3_bucket_notification: Fix missing `bucket` field after import ([#978](https://github.com/terraform-providers/terraform-provider-aws/issues/978))
* resource/aws_sfn_state_machine: Handle another NotFound exception type ([#1062](https://github.com/terraform-providers/terraform-provider-aws/issues/1062))
* resource/aws_ssm_parameter: ForceNew on ssm_parameter rename ([#1022](https://github.com/terraform-providers/terraform-provider-aws/issues/1022))
* resource/aws_instance: Update SourceDestCheck modification on new resources ([#1065](https://github.com/terraform-providers/terraform-provider-aws/issues/1065))
* resource/aws_spot_instance_request: fixed and issue with network interfaces configuration ([#1070](https://github.com/terraform-providers/terraform-provider-aws/issues/1070))
* resource/aws_rds_cluster: Modify RDS Cluster after restoring from snapshot, if required ([#926](https://github.com/terraform-providers/terraform-provider-aws/issues/926))
* resource/aws_kms_alias: Retry lookups after creation ([#1040](https://github.com/terraform-providers/terraform-provider-aws/issues/1040))
* resource/aws_internet_gateway: Retry deletion properly on `DependencyViolation` ([#1021](https://github.com/terraform-providers/terraform-provider-aws/issues/1021))
* resource/aws_elb: Cleanup ENIs after deleting ELB ([#1036](https://github.com/terraform-providers/terraform-provider-aws/issues/1036))
* resource/aws_kms_key: Retry lookups after creation ([#1039](https://github.com/terraform-providers/terraform-provider-aws/issues/1039))
* resource/aws_dms_replication_instance: Add modifying as a pending creation state ([#1114](https://github.com/terraform-providers/terraform-provider-aws/issues/1114))
* resource/aws_redshift_cluster: Trigger ForceNew aws_redshift_cluster on encrypted change ([#1120](https://github.com/terraform-providers/terraform-provider-aws/issues/1120))
* resource/aws_default_network_acl: Add support for ipv6_cidr_block ([#1113](https://github.com/terraform-providers/terraform-provider-aws/issues/1113))
* resource/aws_autoscaling_group: Suppress diffs when an empty set is specified for `availability_zones` ([#1190](https://github.com/terraform-providers/terraform-provider-aws/issues/1190))
* resource/aws_vpc: Ignore ClassicLink DNS support in unsupported regions ([#1176](https://github.com/terraform-providers/terraform-provider-aws/issues/1176))
* resource/elastic_beanstalk_configuration_template: Handle missing platform ([#1222](https://github.com/terraform-providers/terraform-provider-aws/issues/1222))
* r/elasticache_parameter_group: support more than 20 parameters ([#1221](https://github.com/terraform-providers/terraform-provider-aws/issues/1221))
* data-source/aws_db_instance: Fix the output of subnet_group_name ([#1141](https://github.com/terraform-providers/terraform-provider-aws/issues/1141))
* data-source/aws_iam_server_certificate: Fix restriction on length of `name_prefix` ([#1217](https://github.com/terraform-providers/terraform-provider-aws/issues/1217))

## 0.1.2 (June 30, 2017)

FEATURES:

* **New Resource**: `aws_network_interface_sg_attachment` ([#860](https://github.com/terraform-providers/terraform-provider-aws/issues/860))
* **New Data Source**: `aws_ecr_repository` ([#944](https://github.com/terraform-providers/terraform-provider-aws/issues/944))

IMPROVEMENTS:

* Added ability to change the deadline for the EC2 metadata API endpoint ([#950](https://github.com/terraform-providers/terraform-provider-aws/issues/950))
* resource/aws_api_gateway_integration: Add support for specifying cache key parameters ([#893](https://github.com/terraform-providers/terraform-provider-aws/issues/893))
* resource/aws_cloudwatch_event_target: Add ecs_target ([#977](https://github.com/terraform-providers/terraform-provider-aws/issues/977))
* resource/aws_vpn_connection: Add BGP related information on aws_vpn_connection ([#973](https://github.com/terraform-providers/terraform-provider-aws/issues/973))
* resource/aws_cloudformation_stack: Add timeout support ([#994](https://github.com/terraform-providers/terraform-provider-aws/issues/994))
* resource/aws_ssm_parameter: Add support for ssm parameter overwrite ([#1006](https://github.com/terraform-providers/terraform-provider-aws/issues/1006))
* resource/aws_codebuild_project: Add support for environment privileged_mode [GH1009]
* resource/aws_dms_endpoint: Add support for dynamodb as an endpoint target ([#1002](https://github.com/terraform-providers/terraform-provider-aws/issues/1002))
* resource/aws_s3_bucket: Support lifecycle tags filter ([#899](https://github.com/terraform-providers/terraform-provider-aws/issues/899))
* resource/aws_s3_bucket_object: Allow to set WebsiteRedirect on S3 object ([#1020](https://github.com/terraform-providers/terraform-provider-aws/issues/1020))

BUG FIXES:

* resource/aws_waf: Only set FieldToMatch.Data if not empty ([#953](https://github.com/terraform-providers/terraform-provider-aws/issues/953))
* resource/aws_elastic_beanstalk_application_version: Scope labels to application ([#956](https://github.com/terraform-providers/terraform-provider-aws/issues/956))
* resource/aws_s3_bucket: Allow use of `days = 0` with lifecycle transition ([#957](https://github.com/terraform-providers/terraform-provider-aws/issues/957))
* resource/aws_ssm_maintenance_window_task: Make task_parameters updateable on aws_ssm_maintenance_window_task resource ([#965](https://github.com/terraform-providers/terraform-provider-aws/issues/965))
* resource/aws_kinesis_stream: don't force stream destroy on shard_count update ([#894](https://github.com/terraform-providers/terraform-provider-aws/issues/894))
* resource/aws_cloudfront_distribution: Remove validation from custom_origin params ([#987](https://github.com/terraform-providers/terraform-provider-aws/issues/987))
* resource_aws_route53_record: Allow import of Route 53 records with underscores in the name ([#14717](https://github.com/hashicorp/terraform/pull/14717))
* d/aws_db_snapshot: Id was being set incorrectly ([#992](https://github.com/terraform-providers/terraform-provider-aws/issues/992))
* resource/aws_spot_fleet_request: Raise the create timeout to be 10m ([#993](https://github.com/terraform-providers/terraform-provider-aws/issues/993))
* d/aws_ecs_cluster: Add ARN as an exported param for aws_ecs_cluster ([#991](https://github.com/terraform-providers/terraform-provider-aws/issues/991))
* resource/aws_ebs_volume: Not setting the state for ebs_volume correctly ([#999](https://github.com/terraform-providers/terraform-provider-aws/issues/999))
* resource/aws_network_acl: Make action in ingress / egress case insensitive ([#1000](https://github.com/terraform-providers/terraform-provider-aws/issues/1000))

## 0.1.1 (June 21, 2017)

BUG FIXES:

* Fixing malformed ARN attribute for aws_security_group data source ([#910](https://github.com/terraform-providers/terraform-provider-aws/issues/910))

## 0.1.0 (June 20, 2017)

BACKWARDS INCOMPATIBILITIES / NOTES:

FEATURES:

* **New Resource:** `aws_vpn_gateway_route_propagation` [[#15137](https://github.com/terraform-providers/terraform-provider-aws/issues/15137)](https://github.com/hashicorp/terraform/pull/15137)

IMPROVEMENTS:

* resource/ebs_snapshot: Add support for tags ([#3](https://github.com/terraform-providers/terraform-provider-aws/issues/3))
* resource/aws_elasticsearch_domain: now retries on IAM role association failure ([#12](https://github.com/terraform-providers/terraform-provider-aws/issues/12))
* resource/codebuild_project: Increase timeout for creation retry (IAM) ([#904](https://github.com/terraform-providers/terraform-provider-aws/issues/904))
* resource/dynamodb_table: Expose stream_label attribute ([#20](https://github.com/terraform-providers/terraform-provider-aws/issues/20))
* resource/opsworks: Add support for configurable timeouts in AWS OpsWorks Instances. ([#857](https://github.com/terraform-providers/terraform-provider-aws/issues/857))
* Fix handling of AdRoll's hologram clients ([#17](https://github.com/terraform-providers/terraform-provider-aws/issues/17))
* resource/sqs_queue: Add support for name_prefix to aws_sqs_queue ([#855](https://github.com/terraform-providers/terraform-provider-aws/issues/855))
* resource/iam_role: Add support for iam_role tp force_detach_policies ([#890](https://github.com/terraform-providers/terraform-provider-aws/issues/890))

BUG FIXES:

* fix aws cidr validation error [[#15158](https://github.com/terraform-providers/terraform-provider-aws/issues/15158)](https://github.com/hashicorp/terraform/pull/15158)
* resource/elasticache_parameter_group: Retry deletion on InvalidCacheParameterGroupState ([#8](https://github.com/terraform-providers/terraform-provider-aws/issues/8))
* resource/security_group: Raise creation timeout ([#9](https://github.com/terraform-providers/terraform-provider-aws/issues/9))
* resource/rds_cluster: Retry modification on InvalidDBClusterStateFault ([#18](https://github.com/terraform-providers/terraform-provider-aws/issues/18))
* resource/lambda: Fix incorrect GovCloud regexes ([#16](https://github.com/terraform-providers/terraform-provider-aws/issues/16))
* Allow ipv6_cidr_block to be assigned to peering_connection ([#879](https://github.com/terraform-providers/terraform-provider-aws/issues/879))
* resource/rds_db_instance: Correctly create cross-region encrypted replica ([#865](https://github.com/terraform-providers/terraform-provider-aws/issues/865))
* resource/eip: dissociate EIP on update ([#878](https://github.com/terraform-providers/terraform-provider-aws/issues/878))
* resource/iam_server_certificate: Increase deletion timeout ([#907](https://github.com/terraform-providers/terraform-provider-aws/issues/907))
