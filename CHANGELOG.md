## 3.27.0 (Unreleased)

ENHANCEMENTS:

* data-source/aws_subnet: Add `customer_owned_ipv4_pool` and `map_customer_owned_ip_on_launch` attributes ([#16676](https://github.com/hashicorp/terraform-provider-aws/issues/16676))
* resource/aws_glacier_vault: Add plan-time validation for `notification` configuration block `events` and `sns_topic_arn` arguments ([#12645](https://github.com/hashicorp/terraform-provider-aws/issues/12645))
* resource/aws_iam_access_key: Add `create_date` attribute ([#17318](https://github.com/hashicorp/terraform-provider-aws/issues/17318))
* resource/aws_iam_access_key: Support resource import ([#17321](https://github.com/hashicorp/terraform-provider-aws/issues/17321))
* resource/aws_subnet: Add `customer_owned_ipv4_pool` and `map_customer_owned_ip_on_launch` attributes ([#16676](https://github.com/hashicorp/terraform-provider-aws/issues/16676))

BUG FIXES:

* resource/aws_glacier_vault: Prevent crash with `GetVaultAccessPolicy` API errors ([#12645](https://github.com/hashicorp/terraform-provider-aws/issues/12645))
* resource/aws_glacier_vault: Properly remove from state when resource does not exist ([#12645](https://github.com/hashicorp/terraform-provider-aws/issues/12645))
* resource/aws_iam_access_key: Ensure `Inactive` `status` is properly configured during resource creation ([#17322](https://github.com/hashicorp/terraform-provider-aws/issues/17322))
* resource/aws_kinesis_firehose_delivery_stream: Use standard retry timeout for IAM eventual consistency and retry on LakeFormation access errors ([#17254](https://github.com/hashicorp/terraform-provider-aws/issues/17254))

## 3.26.0 (January 28, 2021)

NOTES:

* data-source/aws_route53_zone: The Route 53 `ListResourceRecordSets` API call has been implemented to support the `name_servers` attribute for private Hosted Zones similar to the resource implementation. Environments using restrictive IAM permissions may require updates. ([#17002](https://github.com/hashicorp/terraform-provider-aws/issues/17002))

FEATURES:

* **New Data Source:** `aws_imagebuilder_image` ([#16710](https://github.com/hashicorp/terraform-provider-aws/issues/16710))
* **New Resource:** `aws_imagebuilder_image` ([#16710](https://github.com/hashicorp/terraform-provider-aws/issues/16710))
* **New Resource:** `aws_prometheus_workspace` ([#16882](https://github.com/hashicorp/terraform-provider-aws/issues/16882))
* **New Resource:** `aws_sagemaker_app_image_config` ([#17221](https://github.com/hashicorp/terraform-provider-aws/issues/17221))

ENHANCEMENTS:

* data-source/aws_elasticache_replication_group: Add `multi_az_enabled` argument ([#17320](https://github.com/hashicorp/terraform-provider-aws/issues/17320))
* data-source/aws_vpc_peering_connection: Add `cidr_block_set` and `peer_cidr_block_set` attributes ([#13420](https://github.com/hashicorp/terraform-provider-aws/issues/13420))
* provider: Support AWS Single-Sign On (SSO) cached credentials ([#17340](https://github.com/hashicorp/terraform-provider-aws/issues/17340))
* resource/aws_codeartifact_domain: Make `encryption_key` optional ([#17262](https://github.com/hashicorp/terraform-provider-aws/issues/17262))
* resource/aws_elasticache_replication_group: Add `multi_az_enabled` argument ([#17320](https://github.com/hashicorp/terraform-provider-aws/issues/17320))
* resource/aws_elasticache_replication_group: Allow changing `cluster_mode.replica_count` without re-creation ([#17301](https://github.com/hashicorp/terraform-provider-aws/issues/17301))

BUG FIXES:

* data-source/aws_elb_hosted_zone_id: Correct values for `cn-north-1` and `cn-northwest-1` regions ([#17226](https://github.com/hashicorp/terraform-provider-aws/issues/17226))
* data-source/aws_lb_listener: Prevent error when retrieving a listener whose default action contains weighted target groups ([#17238](https://github.com/hashicorp/terraform-provider-aws/issues/17238))
* data-source/aws_route53_zone: Ensure `name_servers` is populated for private Hosted Zones ([#17002](https://github.com/hashicorp/terraform-provider-aws/issues/17002))
* resource/aws_ebs_volume: Allow both `size` and `snapshot_id` attributes to be specified ([#17243](https://github.com/hashicorp/terraform-provider-aws/issues/17243))
* resource/aws_elasticache_replication_group: Correctly update computed `member_clusters` values ([#17201](https://github.com/hashicorp/terraform-provider-aws/issues/17201))
* resource/aws_sagemaker_code_repository: fix doc name ([#17221](https://github.com/hashicorp/terraform-provider-aws/issues/17221))

## 3.25.0 (January 22, 2021)

NOTES

* resource/aws_lightsail_instance: The `ipv6_address` attribute has been deprecated. Use the `ipv6_addresses` attribute instead. This is due to a backwards incompatible change in the Lightsail API. ([#17155](https://github.com/hashicorp/terraform-provider-aws/issues/17155))

FEATURES

* **New Resource:** `aws_backup_global_settings` ([#16475](https://github.com/hashicorp/terraform-provider-aws/issues/16475))
* **New Resource:** `aws_sagemaker_feature_group` ([#16728](https://github.com/hashicorp/terraform-provider-aws/issues/16728))
* **New Resource:** `aws_sagemaker_image_version` ([#17141](https://github.com/hashicorp/terraform-provider-aws/issues/17141))
* **New Resource:** `aws_sagemaker_user_profile` ([#17123](https://github.com/hashicorp/terraform-provider-aws/issues/17123))

ENHANCEMENTS

* data-source/aws_ami: Add `throughput` attribute to `block_device_mappings` `ebs` attribute ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* data-source/aws_ebs_volume: Add `throughput` attribute ([#16517](https://github.com/hashicorp/terraform-provider-aws/issues/16517))
* data-source/aws_elasticache_replication_group: Adds `arn` attribute ([#15348](https://github.com/hashicorp/terraform-provider-aws/issues/15348))
* data-source/aws_iam_user: Add `tags` attribute ([#13287](https://github.com/hashicorp/terraform-provider-aws/issues/13287))
* resource/aws_ami: Support `volume_type` value of `gp3` and add `throughput` argument to `ebs_block_device` configuration block ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* resource/aws_ami_copy: Add `throughput` argument to `ebs_block_device` configuration block ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* resource/aws_ami_from_instance: Add `throughput` argument to `ebs_block_device` configuration block ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* resource/aws_ebs_volume: Add `throughput` argument ([#16517](https://github.com/hashicorp/terraform-provider-aws/issues/16517))
* resource/aws_elasticache_replication_group: Adds `arn` attribute ([#15348](https://github.com/hashicorp/terraform-provider-aws/issues/15348))
* resource/aws_lightsail_instance: Add `ipv6_addresses` attribute ([#17155](https://github.com/hashicorp/terraform-provider-aws/issues/17155))
* resource/aws_sagemaker_domain: Delete implicit EFS file system ([#17123](https://github.com/hashicorp/terraform-provider-aws/issues/17123))

BUG FIXES 

* data-source/aws_lambda_function: Prevent error when getting Code Signing Config for container image based lambdas during read ([#17180](https://github.com/hashicorp/terraform-provider-aws/issues/17180))
* provider: Fix error messages for missing required blocks not including the block name ([#17211](https://github.com/hashicorp/terraform-provider-aws/issues/17211))
* provider: Prevent panic when sending Ctrl-C (SIGINT) to Terraform ([#17211](https://github.com/hashicorp/terraform-provider-aws/issues/17211))
* resource/aws_api_gateway_authorizer: Ensure `authorizer_credentials` are configured when `type` is `COGNITO_USER_POOLS` ([#16614](https://github.com/hashicorp/terraform-provider-aws/issues/16614))
* resource/aws_api_gateway_rest_api: Allow `api_key_source`, `binary_media_types`, and `description` arguments to be omitted from configuration with OpenAPI specification import (`body` argument) ([#17099](https://github.com/hashicorp/terraform-provider-aws/issues/17099))
* resource/aws_api_gateway_rest_api: Ensure `api_key_source`, `binary_media_types`, `description`, `minimum_compression_size`, `name`, and `policy` configuration values are correctly applied as an override after OpenAPI specification import (`body` argument) ([#17099](https://github.com/hashicorp/terraform-provider-aws/issues/17099))
* resource/aws_api_gateway_rest_api: Fix `disable_execute_api_endpoint` and `endpoint_configuration` `vpc_endpoint_ids` handling with OpenAPI specification import (`body` argument) ([#17209](https://github.com/hashicorp/terraform-provider-aws/issues/17209))
* resource/aws_lakeformation_data_lake_settings: Avoid unnecessary resource cycling ([#17189](https://github.com/hashicorp/terraform-provider-aws/issues/17189))
* resource/aws_lakeformation_permissions: Handle resources with multiple permissions ([#17189](https://github.com/hashicorp/terraform-provider-aws/issues/17189))
* resource/aws_lambda_function: Prevent panic with missing `FunctionConfiguration` `PackageType` attribute in API response ([#16544](https://github.com/hashicorp/terraform-provider-aws/issues/16544))
* resource/aws_lambda_function: Prevent panic with missing environment variable value ([#17056](https://github.com/hashicorp/terraform-provider-aws/issues/17056))
* resource/aws_sagemaker_image: Fix catching image not found on read error ([#17141](https://github.com/hashicorp/terraform-provider-aws/issues/17141))


## 3.24.1 (January 15, 2021)

BUG FIXES

* data-source/instance: Fix EBS and root block device tags issue with "Invalid address to set" ([#17136](https://github.com/hashicorp/terraform-provider-aws/issues/17136))

## 3.24.0 (January 14, 2021)

FEATURES

* **New Data Source:** `aws_api_gateway_domain_name` ([#12489](https://github.com/hashicorp/terraform-provider-aws/issues/12489))
* **New Data Source:** `aws_identitystore_group` ([#15322](https://github.com/hashicorp/terraform-provider-aws/issues/15322))
* **New Data Source:** `aws_identitystore_user` ([#15322](https://github.com/hashicorp/terraform-provider-aws/issues/15322))
* **New Resource:** `aws_cloudwatch_composite_alarm` ([#15023](https://github.com/hashicorp/terraform-provider-aws/issues/15023))
* **New Resource:** `aws_fms_policy` ([#9594](https://github.com/hashicorp/terraform-provider-aws/issues/9594))
* **New Resource:** `aws_route53_resolver_dnssec_config` ([#17012](https://github.com/hashicorp/terraform-provider-aws/issues/17012))
* **New Resource:** `aws_sagemaker_domain` ([#16077](https://github.com/hashicorp/terraform-provider-aws/issues/16077))
* **New Resource:** `aws_ssoadmin_account_assignment` ([#15322](https://github.com/hashicorp/terraform-provider-aws/issues/15322))

ENHANCEMENTS

* data-source/aws_workspaces_directory: Add access properties ([#16688](https://github.com/hashicorp/terraform-provider-aws/issues/16688))
* resource/aws_api_gateway_base_path_mapping: Support in-place updates for `api_id`, `base_path`, and `stage_name` ([#16147](https://github.com/hashicorp/terraform-provider-aws/issues/16147))
* resource/aws_api_gateway_domain_name: Add `mutual_tls_authentication` configuration block ([#15258](https://github.com/hashicorp/terraform-provider-aws/issues/15258))
* resource/aws_api_gateway_integration: Add `tls_config` configuration block ([#15499](https://github.com/hashicorp/terraform-provider-aws/issues/15499))
* resource/aws_api_gateway_method: Add `operation_name` argument ([#13282](https://github.com/hashicorp/terraform-provider-aws/issues/13282))
* resource/aws_api_gateway_rest_api: Add `disable_execute_api_endpoint` argument ([#16198](https://github.com/hashicorp/terraform-provider-aws/issues/16198))
* resource/aws_api_gateway_rest_api: Add `parameters` argument ([#7374](https://github.com/hashicorp/terraform-provider-aws/issues/7374))
* resource/aws_apigatewayv2_integration: Add `response_parameters` attribute ([#17043](https://github.com/hashicorp/terraform-provider-aws/issues/17043))
* resource/aws_codepipeline: Deprecates GitHub v1 (OAuth token) authentication and removes hashing of GitHub token ([#16959](https://github.com/hashicorp/terraform-provider-aws/issues/16959))
* resource/aws_codepipeline: Adds GitHub v2 (CodeStar Connetion) authentication ([#16959](https://github.com/hashicorp/terraform-provider-aws/issues/16959))
* resource/aws_dms_endpoint: Add `s3_settings` `date_partition_enabled` argument ([#16827](https://github.com/hashicorp/terraform-provider-aws/issues/16827))
* resource/aws_elasticache_cluster: Add support for final snapshot with Redis engine ([#15592](https://github.com/hashicorp/terraform-provider-aws/issues/15592))
* resource/aws_elasticache_replication_group: Add support for final snapshot ([#15592](https://github.com/hashicorp/terraform-provider-aws/issues/15592))
* resource/aws_globalaccelerator_accelerator: Add custom timeouts ([#17112](https://github.com/hashicorp/terraform-provider-aws/issues/17112))
* resource/aws_globalaccelerator_endpoint_group: Add custom timeouts ([#17112](https://github.com/hashicorp/terraform-provider-aws/issues/17112))
* resource/aws_globalaccelerator_endpoint_listener: Add custom timeouts ([#17112](https://github.com/hashicorp/terraform-provider-aws/issues/17112))
* resource/aws_instance: Add `tags` parameter to `root_block_device`, `ebs_block_device` blocks.([#15474](https://github.com/hashicorp/terraform-provider-aws/issues/15474))
* resource/aws_workspaces_directory: Add access properties ([#16688](https://github.com/hashicorp/terraform-provider-aws/issues/16688))

BUG FIXES

* resource/aws_appmesh_route: Allow an empty `match` attribute to specified for a `grpc_route`, indicating that any service should be matched ([#16867](https://github.com/hashicorp/terraform-provider-aws/issues/16867))
* resource/aws_db_instance: Correctly validate `final_snapshot_identifier` argument at plan-time ([#16885](https://github.com/hashicorp/terraform-provider-aws/issues/16885))
* resource/aws_dms_endpoint: Support `extra_connection_attributes` for all engine names during create and read ([#16827](https://github.com/hashicorp/terraform-provider-aws/issues/16827))
* resource/aws_instance: Prevent `volume_tags` from improperly interfering with `tags` in `aws_ebs_volume` ([#15474](https://github.com/hashicorp/terraform-provider-aws/issues/15474))
* resource/aws_networkfirewall_rule_group: Prevent resource recreation due to `stateful_rule` changes after creation ([#16884](https://github.com/hashicorp/terraform-provider-aws/issues/16884))
* resource/aws_route53_zone_association: Prevent deletion errors for missing Hosted Zone or VPC association ([#17023](https://github.com/hashicorp/terraform-provider-aws/issues/17023))
* resource/aws_sagemaker_image - fix error on wait for delete when image does not exist ([#16077](https://github.com/hashicorp/terraform-provider-aws/issues/16077))
* resource/aws_s3_bucket_inventory: Prevent crashes with empty `destination`, `filter`, and `schedule` configuration blocks ([#17055](https://github.com/hashicorp/terraform-provider-aws/issues/17055))
* service/apigateway: All operations will now automatically retry on `ConflictException: Unable to complete operation due to concurrent modification. Please try again later.` errors.

## 3.23.0 (January 08, 2021)

FEATURES

* **New Data Source:** `aws_ssoadmin_instances` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Data Source:** `aws_ssoadmin_permission_set` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Resource:** `aws_sagemaker_image` ([#16082](https://github.com/hashicorp/terraform-provider-aws/issues/16082))
* **New Resource:** `aws_ssoadmin_managed_policy_attachment` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Resource:** `aws_ssoadmin_permission_set` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Resource:** `aws_ssoadmin_permission_set_inline_policy` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))

ENHANCEMENTS

* data-source/aws_imagebuilder_image_recipe: Add `working_directory` attribute ([#16947](https://github.com/hashicorp/terraform-provider-aws/issues/16947))
* data-source/aws_elasticache_replication_group: Add reader_endpoint_address attribute ([#9979](https://github.com/hashicorp/terraform-provider-aws/issues/9979))
* resource/aws_elasticache_replication_group: Add reader_endpoint_address attribute ([#9979](https://github.com/hashicorp/terraform-provider-aws/issues/9979))
* resource/aws_elasticache_replication_group: Allows configuring `replicas_per_node_group` for "Redis (cluster mode disabled)" ([#16829](https://github.com/hashicorp/terraform-provider-aws/issues/16829))
* resource/aws_imagebuilder_image_recipe: Add `working_directory` argument ([#16947](https://github.com/hashicorp/terraform-provider-aws/issues/16947))
* resource/aws_glue_crawler: add support for `lineage_configuration` and `recrawl_policy` ([#16714](https://github.com/hashicorp/terraform-provider-aws/issues/16714))
* resource/aws_glue_crawler: add plan time validations to `name`, `description` and `table_prefix` ([#16714](https://github.com/hashicorp/terraform-provider-aws/issues/16714))
* resource/aws_kinesis_stream: Update `retention_period` argument plan-time validation to include up to 8760 hours ([#16608](https://github.com/hashicorp/terraform-provider-aws/issues/16608))
* resource/aws_msk_cluster: Support `PER_TOPIC_PER_PARTITION` value for `enhanced_monitoring` argument plan-time validation ([#16914](https://github.com/hashicorp/terraform-provider-aws/issues/16914))
* resource/aws_route53_zone: Add length validations for `delegation_set_id` and `name` arguments ([#12340](https://github.com/hashicorp/terraform-provider-aws/issues/12340))
* resource/aws_vpc_endpoint_service: Make `private_dns_name` configurable and add `private_dns_name_configuration` attribute ([#16495](https://github.com/hashicorp/terraform-provider-aws/issues/16495))

BUG FIXES

* resource/aws_emr_cluster: Remove from state instead of returning an error on long terminated cluster ([#16924](https://github.com/hashicorp/terraform-provider-aws/issues/16924))
* resource/aws_glue_catalog_table: Glue table partition keys should be set to empty list instead of being unset ([#16727](https://github.com/hashicorp/terraform-provider-aws/issues/16727))
* resource/aws_imagebuilder_distribution_configuration: Remove `user_ids` argument maximum limit ([#16905](https://github.com/hashicorp/terraform-provider-aws/issues/16905))
* resource/aws_transfer_user: Update `user_name` argument validation to support 100 characters ([#16938](https://github.com/hashicorp/terraform-provider-aws/issues/16938))

## 3.22.0 (December 18, 2020)

FEATURES

* **New Data Source:** `aws_ec2_managed_prefix_list` ([#16738](https://github.com/hashicorp/terraform-provider-aws/issues/16738))
* **New Data Source:** `aws_lakeformation_data_lake_settings` ([#13250](https://github.com/hashicorp/terraform-provider-aws/issues/13250))
* **New Data Source:** `aws_lakeformation_permissions` ([#13396](https://github.com/hashicorp/terraform-provider-aws/issues/13396))
* **New Data Source:** `aws_lakeformation_resource` ([#13396](https://github.com/hashicorp/terraform-provider-aws/issues/13396))
* **New Resource:** `aws_codestarconnections_connection` ([#15990](https://github.com/hashicorp/terraform-provider-aws/issues/15990))
* **New Resource:** `aws_ec2_managed_prefix_list` ([#14068](https://github.com/hashicorp/terraform-provider-aws/issues/14068))
* **New Resource:** `aws_lakeformation_data_lake_settings` ([#13250](https://github.com/hashicorp/terraform-provider-aws/issues/13250))
* **New Resource:** `aws_lakeformation_permissions` ([#13396](https://github.com/hashicorp/terraform-provider-aws/issues/13396))
* **New Resource:** `aws_lakeformation_resource` ([#13267](https://github.com/hashicorp/terraform-provider-aws/issues/13267))

ENHANCEMENTS

* data-source/aws_autoscaling_group: Adds `launch_template` attribute ([#16297](https://github.com/hashicorp/terraform-provider-aws/issues/16297))
* data-source/aws_availability_zone: Add `parent_zone_id`, `parent_zone_name`, and `zone_type` attributes (additional support for Local and Wavelength Zones) ([#16770](https://github.com/hashicorp/terraform-provider-aws/issues/16770))
* data-source/aws_eip: Add `carrier_ip` attribute ([#16724](https://github.com/hashicorp/terraform-provider-aws/issues/16724))
* data-source/aws_instance: Add `enclave_options` attribute (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* data-source/aws_instance: Add `ebs_block_device` and `root_block_device` configuration block `throughput` attribute ([#16620](https://github.com/hashicorp/terraform-provider-aws/issues/16620))
* data-source/aws_launch_configuration: Add `metadata_options` attribute ([#14637](https://github.com/hashicorp/terraform-provider-aws/issues/14637))
* data-source/aws_launch_template: Add `enclave_options` attribute (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* data-source/aws_network_interface: Add `association` `carrier_ip` and `customer_owned_ip` attributes ([#16723](https://github.com/hashicorp/terraform-provider-aws/issues/16723))
* resource/aws_autoscaling_group: Adds support for Instance Refresh ([#16678](https://github.com/hashicorp/terraform-provider-aws/issues/16678))
* resource/aws_eip: Add `carrier_ip` attribute ([#16724](https://github.com/hashicorp/terraform-provider-aws/issues/16724))
* resource/aws_instance: Add `enclave_options` configuration block (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* resource/aws_instance: Add `ebs_block_device` and `root_block_device` configuration block `throughput` attribute ([#16620](https://github.com/hashicorp/terraform-provider-aws/issues/16620))
* resource/aws_kinesis_firehose_delivery_stream: Mark `http_endpoint_configuration` `access_key` as sensitive ([#16684](https://github.com/hashicorp/terraform-provider-aws/issues/16684))
* resource/aws_launch_configuration: Add `metadata_options` configuration block ([#14637](https://github.com/hashicorp/terraform-provider-aws/issues/14637))
* resource/aws_launch_template: Add `enclave_options` configuration block (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* resource/aws_vpn_connection: Add support for VPN tunnel options and enable acceleration, DPDTimeoutAction, StartupAction, local/remote IPv4/IPv6 network CIDR and tunnel inside IP version. ([#14740](https://github.com/hashicorp/terraform-provider-aws/issues/14740))

BUG FIXES

* data-source/aws_ec2_coip_pools: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_ec2_local_gateways: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_ec2_local_gateway_route_tables: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_prefix_list: Using `name` argument no longer overrides other arguments ([#16739](https://github.com/hashicorp/terraform-provider-aws/issues/16739))
* resource/aws_db_instance: Fix missing `db_subnet_group_name` in API request when using `restore_to_point_in_time` ([#16830](https://github.com/hashicorp/terraform-provider-aws/issues/16830))
* resource/aws_eip_association: Handle eventual consistency when creating resource ([#16808](https://github.com/hashicorp/terraform-provider-aws/issues/16808))
* resource/aws_main_route_table_association: Prevent crash on creation when VPC main route table association is not found ([#16680](https://github.com/hashicorp/terraform-provider-aws/issues/16680))
* resource/aws_workspaces_workspace: Prevent panic from terminated WorkSpace ([#16692](https://github.com/hashicorp/terraform-provider-aws/issues/16692))

## 3.21.0 (December 11, 2020)

NOTES

* resource/aws_imagebuilder_image_recipe: Previously the ordering of `component` configuration blocks was not properly handled by the resource, which could cause unexpected behavior with multiple Components. These configurations may see the ordering difference being fixed after upgrade. ([#16566](https://github.com/hashicorp/terraform-provider-aws/issues/16566))

FEATURES

* **New Resource:** `aws_ec2_carrier_gateway` ([#16252](https://github.com/hashicorp/terraform-provider-aws/issues/16252))
* **New Resource:** `aws_glue_schema` ([#16612](https://github.com/hashicorp/terraform-provider-aws/issues/16612))

ENHANCEMENTS

* data-source/aws_launch_template: Add `associate_carrier_ip_address` attribute to `network_interfaces` configuration block ([#16707](https://github.com/hashicorp/terraform-provider-aws/issues/16707))
* data-source/aws_launch_template: Add `throughput` attribute to `block_device_mappings.ebs` configuration block ([#16649](https://github.com/hashicorp/terraform-provider-aws/issues/16649))
* data-source/aws_launch_template: Support `id` as argument ([#16457](https://github.com/hashicorp/terraform-provider-aws/issues/16457))
* resource/aws_appmesh_virtual_node: Add `listener.connection_pool` attribute ([#16167](https://github.com/hashicorp/terraform-provider-aws/issues/16167))
* resource/aws_appmesh_virtual_node: Add `listener.outlier_detection` attribute ([#16167](https://github.com/hashicorp/terraform-provider-aws/issues/16167))
* resource/aws_launch_template: Add `associate_carrier_ip_address` attribute to `network_interfaces` configuration block ([#16707](https://github.com/hashicorp/terraform-provider-aws/issues/16707))
* resource/aws_launch_template: Add `throughput` attribute to `block_device_mappings.ebs` configuration block ([#16649](https://github.com/hashicorp/terraform-provider-aws/issues/16649))
* resource/aws_spot_fleet_request: Add `throughput` attribute to `launch_specification.ebs_block_device` and `launch_specification.root_block_device` configuration blocks ([#16652](https://github.com/hashicorp/terraform-provider-aws/issues/16652))
* resource/aws_ssm_maintenance_window: Add `schedule_offset` argument ([#16569](https://github.com/hashicorp/terraform-provider-aws/issues/16569))
* resource/aws_workspaces_workspace: Add failed request error code along with message ([#16459](https://github.com/hashicorp/terraform-provider-aws/issues/16459))

BUG FIXES

* data-source/aws_customer_gateway: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway_peering_attachment: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway_route_table: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway_vpc_attachment: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_guardduty_detector: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_imagebuilder_image_recipe: Ensure proper ordering of `component` attribute ([#16566](https://github.com/hashicorp/terraform-provider-aws/issues/16566))
* resource/aws_backup_plan: Prevent plan-time validation error for pre-existing resources with `lifecycle` `delete_after` and/or `copy_action` `lifecycle` `delete_after` arguments configured ([#16605](https://github.com/hashicorp/terraform-provider-aws/issues/16605))
* resource/aws_imagebuilder_image_recipe: Ensure proper ordering of `component` configuration blocks ([#16566](https://github.com/hashicorp/terraform-provider-aws/issues/16566))
* resource/aws_workspaces_directory: Fix empty custom_security_group_id & default_ou ([#16589](https://github.com/hashicorp/terraform-provider-aws/issues/16589))

## 3.20.0 (December 03, 2020)

ENHANCEMENTS

* resource/aws_backup_plan: Add plan-time validation for various arguments ([#16476](https://github.com/hashicorp/terraform-provider-aws/issues/16476))
* resource/aws_eks_node_group: Make `capacity_type` a `Computed` attribute ([#16552](https://github.com/hashicorp/terraform-provider-aws/issues/16552))
* resource/aws_lambda_event_source_mapping: Add support for updating `maximum_batching_window_in_seconds` for SQS queue event sources ([#16518](https://github.com/hashicorp/terraform-provider-aws/issues/16518))
* resource/aws_ssm_maintenance_window_target: Add plan-time validation for `owner_information` and `targets` arguments ([#16478](https://github.com/hashicorp/terraform-provider-aws/issues/16478))
* resource/aws_storagegateway_gateway - add `timeout_in_seconds`, `organizational_unit`, `domain_controllers` arguments for `smb_active_directory_settings` block. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))
* resource/aws_storagegateway_gateway - add `smb_active_directory_settings. active_directory_status`, `ec2_instance_id`, `endpoint_type`, `host_environment`, and `gateway_network_interface` attributes. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))
* resource/aws_storagegateway_gateway - add plan time validations for `smb_guest_password`, `smb_active_directory_settings. username`, `smb_active_directory_settings. password`, `smb_active_directory_settings. domain_name`, `gateway_timezone`, and `gateway_name`. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))
* resource/aws_storagegateway_gateway - add support for `medium_changer_type`  value `medium_changer_type`. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))

BUG FIXES

* resource/aws_backup_plan: Retry on eventual consistency error during deletion ([#16476](https://github.com/hashicorp/terraform-provider-aws/issues/16476))
* resource/aws_cloudwatch_event_target: Prevent potential panic and prevent recreation after state upgrade with custom `event_bus_name` value ([#16484](https://github.com/hashicorp/terraform-provider-aws/issues/16484))
* resource/aws_ec2_client_vpn_network_association: Increase associate and disassociate timeouts from 10min to 30min ([#16522](https://github.com/hashicorp/terraform-provider-aws/issues/16522))
* resource/aws_instance: Automatically retry instance restart on eventual consistency error during `instance_type` in-place update ([#16443](https://github.com/hashicorp/terraform-provider-aws/issues/16443))
* resource/aws_lambda_function: Prevent error during deletion when resource not found ([#16183](https://github.com/hashicorp/terraform-provider-aws/issues/16183))
* resource/aws_ssm_maintenance_window_target: Remove from state if not found ([#16478](https://github.com/hashicorp/terraform-provider-aws/issues/16478))

## 3.19.0 (December 01, 2020)

FEATURES

* **New Resource:** `aws_glue_registry` ([#16418](https://github.com/hashicorp/terraform-provider-aws/issues/16418))

ENHANCEMENTS

* resource/aws_apigatewayv2_domain_name: Add `mutual_tls_authentication` attribute to support mutual TLS authentication ([#15249](https://github.com/hashicorp/terraform-provider-aws/issues/15249))
* resource/aws_appmesh_virtual_gateway: Add `listener.connection_pool` attribute ([#16168](https://github.com/hashicorp/terraform-provider-aws/issues/16168))
* data-source/aws_eks_cluster: add `kubernetes_network_config` attribute ([#15518](https://github.com/hashicorp/terraform-provider-aws/issues/15518))
* resource/aws_storagegateway_smb_file_share - add support for `notification_policy` and `access_based_enumeration`. ([#16414](https://github.com/hashicorp/terraform-provider-aws/issues/16414))
* resource/aws_storagegateway_smb_file_share - add plan time validation to `invalid_user_list` and `valid_user_list`. ([#16414](https://github.com/hashicorp/terraform-provider-aws/issues/16414))
* resource/aws_cognito_user_pool: add support for account recovery setting. ([#12444](https://github.com/hashicorp/terraform-provider-aws/issues/12444))
* resource/aws_eks_cluster: add `kubernetes_network_config` argument ([#15518](https://github.com/hashicorp/terraform-provider-aws/issues/15518))
* resource/aws_eks_node_group: Add `capacity_type` argument and support multiple `instance_types` (Support Spot Node Groups) ([#16510](https://github.com/hashicorp/terraform-provider-aws/issues/16510))
* resource/aws_lambda_function: Add support for Container Images ([#16512](https://github.com/hashicorp/terraform-provider-aws/issues/16512))

BUG FIXES

* resource/aws_fsx_windows_file_system: Prevent potential panics, unexpected errors, and use correct operation timeout on update ([#16488](https://github.com/hashicorp/terraform-provider-aws/issues/16488))

## 3.18.0 (November 25, 2020)

FEATURES

* **New Data Source:** `aws_imagebuilder_image_pipeline` ([#16299](https://github.com/hashicorp/terraform-provider-aws/issues/16299))
* **New Data Source:** `aws_imagebuilder_image_recipe` ([#16218](https://github.com/hashicorp/terraform-provider-aws/issues/16218))
* **New Data Source:** `aws_serverlessrepository_application` ([#15874](https://github.com/hashicorp/terraform-provider-aws/issues/15874))
* **New Resource:** `aws_backup_region_settings` ([#16114](https://github.com/hashicorp/terraform-provider-aws/issues/16114))
* **New Resource:** `aws_imagebuilder_image_pipeline` ([#16299](https://github.com/hashicorp/terraform-provider-aws/issues/16299))
* **New Resource:** `aws_imagebuilder_image_recipe` ([#16218](https://github.com/hashicorp/terraform-provider-aws/issues/16218))
* **New Resource:** `aws_msk_scram_secret_association` ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* **New Resource:** `aws_networkfirewall_resource_policy` ([#16279](https://github.com/hashicorp/terraform-provider-aws/issues/16279))
* **New Resource:** `aws_serverlessrepository_stack` ([#15874](https://github.com/hashicorp/terraform-provider-aws/issues/15874))

ENHANCEMENTS

* data-source/aws_codeartifact_repository_endpoint: Support `nuget` value in `format` argument plan-time validation ([#16422](https://github.com/hashicorp/terraform-provider-aws/issues/16422))
* data-source/aws_msk_cluster: Add `bootstrap_brokers_sasl_scram` attribute ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* resource/aws_db_proxy_default_target_group: Make `connection_pool_config` optional ([#16303](https://github.com/hashicorp/terraform-provider-aws/issues/16303))
* resource/aws_kinesisanalyticsv2_application: `runtime_environment` now supports `FLINK-1_11` ([#16389](https://github.com/hashicorp/terraform-provider-aws/issues/16389))
* resource/aws_msk_cluster: Add `bootstrap_brokers_sasl_scram` attribute ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* resource/aws_msk_cluster: Add `client_authentication` `sasl` `scram` argument ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* resource/aws_networkfirewall_firewall: Add `firewall_status` attribute to expose VPC endpoints ([#16399](https://github.com/hashicorp/terraform-provider-aws/issues/16399))

BUG FIXES

* data-source/aws_lambda_function: Prevent Lambda `GetFunctionCodeSigningConfig` API call error outside AWS Commercial regions ([#16412](https://github.com/hashicorp/terraform-provider-aws/issues/16412))
* resource/aws_cloudwatch_event_permission: Prevent `arn: invalid prefix` error during read in some environments ([#16319](https://github.com/hashicorp/terraform-provider-aws/issues/16319))
* resource/aws_kinesis_analytics_application: Respect the order of 'record_column' attributes ([#16260](https://github.com/hashicorp/terraform-provider-aws/issues/16260))
* resource/aws_kinesisanalyticsv2_application: Respect the order of 'record_column' attributes ([#16260](https://github.com/hashicorp/terraform-provider-aws/issues/16260))
* resource/aws_lambda_function: Prevent Lambda `GetFunctionCodeSigningConfig` API call error outside AWS Commercial regions ([#16412](https://github.com/hashicorp/terraform-provider-aws/issues/16412))
* resource/aws_lb_listener: Mark `port` argument as optional and only default `protocol` argument to `HTTP` for Application Load Balancers (Support Gateway Load Balancer) ([#16306](https://github.com/hashicorp/terraform-provider-aws/issues/16306))
* resource/aws_securityhub_member: Prevent `invited` attribute updates due to recent API changes ([#16404](https://github.com/hashicorp/terraform-provider-aws/issues/16404))

## 3.17.0 (November 24, 2020)

FEATURES

* **New Data Source:** `aws_lambda_code_signing_config` ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* **New Data Source:** `aws_signer_signing_job` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Data Source:** `aws_signer_signing_profile` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Resource:** `aws_lambda_code_signing_config` ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* **New Resource:** `aws_signer_signing_job` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Resource:** `aws_signer_signing_profile` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Resource:** `aws_signer_signing_profile_permission` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))

ENHANCEMENTS

* data-source/aws_lambda_function: Add `code_signing_config_arn`, `signing_profile_version_arn`, and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* data-source/aws_lambda_layer_version: Add `signing_profile_version_arn` and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* resource/aws_accessanalyzer_analyzer: Adds plan time validation to `analyzer_name` ([#16265](https://github.com/hashicorp/terraform-provider-aws/issues/16265))
* resource/aws_accessanalyzer_analyzer: Adds plan time validation to `analyzer_name` ([#16265](https://github.com/hashicorp/terraform-provider-aws/issues/16265))
* resource/aws_fsx_windows_file_system: Support updating `throughput_capacity` and `storage_capacity` ([#15582](https://github.com/hashicorp/terraform-provider-aws/issues/15582))
* resource/aws_glue_catalog_table: Add partition index support ([#16194](https://github.com/hashicorp/terraform-provider-aws/issues/16194))
* resource/aws_lambda_function: Add `code_signing_config_arn` argument and `signing_profile_version_arn` and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* resource/aws_lambda_layer_version: Add `signing_profile_version_arn` and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* resource/aws_storagegateway_nfs_file_share: Add support for `notification_policy`. ([#16340](https://github.com/hashicorp/terraform-provider-aws/issues/16340))
* resource/aws_storagegateway_nfs_file_share: Add plan time validation for `client_list`, `nfs_file_share_defaults. directory_mode`, `nfs_file_share_defaults. file_mode`, `nfs_file_share_defaults. group_id`, `nfs_file_share_defaults. owner_id` ([#16340](https://github.com/hashicorp/terraform-provider-aws/issues/16340))
* resource/aws_workspaces_directory: Allows assigning IP group ([#14451](https://github.com/hashicorp/terraform-provider-aws/issues/14451))

BUG FIXES

* resource/aws_fsx_windows_file_system: Update the default creation timeout from 30 to 45 minutes ([#16363](https://github.com/hashicorp/terraform-provider-aws/issues/16363))
* resource/aws_lb: Fix `enable_cross_zone_load_balancing` argument handling with Gateway Load Balancers ([#16314](https://github.com/hashicorp/terraform-provider-aws/issues/16314))

## 3.16.0 (November 18, 2020)

* **New Data Source:** `aws_imagebuilder_component` ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* **New Data Source:** `aws_imagebuilder_distribution_configuration` ([#16180](https://github.com/hashicorp/terraform-provider-aws/issues/16180))
* **New Data Source:** `aws_imagebuilder_infrastructure_configuration` ([#16186](https://github.com/hashicorp/terraform-provider-aws/issues/16186))
* **New Resource:** `aws_api_gateway_rest_api_policy` ([#13619](https://github.com/hashicorp/terraform-provider-aws/issues/13619))
* **New Resource:** `aws_backup_vault_policy` ([#16112](https://github.com/hashicorp/terraform-provider-aws/issues/16112))
* **New Resource:** `aws_glue_dev_endpoint` ([#7895](https://github.com/hashicorp/terraform-provider-aws/pull/7895))
* **New Resource:** `aws_imagebuilder_component` ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* **New Resource:** `aws_imagebuilder_distribution_configuration` ([#16180](https://github.com/hashicorp/terraform-provider-aws/issues/16180))
* **New Resource:** `aws_imagebuilder_infrastructure_configuration` ([#16186](https://github.com/hashicorp/terraform-provider-aws/issues/16186))
* **New Resource:** `aws_networkfirewall_firewall` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))
* **New Resource:** `aws_networkfirewall_firewall_policy` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))
* **New Resource:** `aws_networkfirewall_logging_configuration` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))
* **New Resource:** `aws_networkfirewall_rule_group` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))

ENHANCEMENTS

* resource/aws_globalaccelerator_endpoint_group: Add `arn` and `port_override` attributes ([#16121](https://github.com/hashicorp/terraform-provider-aws/issues/16121))
* resource/aws_glue_catalog_table: Add support for `parameters` argument to `storage_descriptor.columns` block ([#16052](https://github.com/hashicorp/terraform-provider-aws/issues/16052))
* resource/aws_glue_catalog_table: Add plan time validation for `description`, `name`, `partition_keys.name`, `partition_keys.comment`, `partition_keys.type`, `retention`, `view_original_text`, `view_expanded_text`, `storage_descriptor.name`, `storage_descriptor.comment`, `storage_descriptor.type`, `storage_descriptor.bucket_columns`, `storage_descriptor.ser_de_info.name`, `storage_descriptor.skewed_info.skewed_column_names`, `storage_descriptor.sort_columns.column`, `storage_descriptor.sort_columns.sort_order` ([#16052](https://github.com/hashicorp/terraform-provider-aws/issues/16052))
* resource/aws_msk_cluster: Support in-place `kafka_version` upgrade ([#13654](https://github.com/hashicorp/terraform-provider-aws/issues/13654))
* resource/aws_storagegateway_smb_file_share: Add `file_share_name` argument ([#16008](https://github.com/hashicorp/terraform-provider-aws/issues/16008))
* resource_aws_storagegateway_nfs_file_share: Add `file_share_name` argument ([#16072](https://github.com/hashicorp/terraform-provider-aws/issues/16072))

BUG FIXES

* data-source/aws_s3_bucket: Use provider credentials when getting the bucket region (fix AWS China non-ICP S3 Buckets and other restrictive environments) ([#15481](https://github.com/hashicorp/terraform-provider-aws/issues/15481))
* resource/aws_apigatewayv2_stage: Correctly handle deletion of route_settings ([#16133](https://github.com/hashicorp/terraform-provider-aws/issues/16133))
* resource/aws_backup_plan - `lifecycle` block in `copy_action` is optional ([#16116](https://github.com/hashicorp/terraform-provider-aws/issues/16116))
* resource/aws_eks_fargate_profile: Serialize multiple profile creation and deletion to prevent `ResourceInUseException` errors ([#14020](https://github.com/hashicorp/terraform-provider-aws/issues/14020))
* resource/aws_organizations_organization: Prevent recreation when `feature_set` is updated to `ALL` ([#15473](https://github.com/hashicorp/terraform-provider-aws/issues/15473))
* resource/aws_s3_bucket: Use provider credentials when getting the bucket region (fix AWS China non-ICP S3 Buckets and other restrictive environments) ([#15481](https://github.com/hashicorp/terraform-provider-aws/issues/15481))
* resource/aws_s3_bucket_object: Correctly updates `version_id` when certain configuration keys are changed ([#14900](https://github.com/hashicorp/terraform-provider-aws/issues/14900))

## 3.15.0 (November 12, 2020)

ENHANCEMENTS

* data-source/aws_ec2_transit_gateway_route_table: Add `arn` attribute ([#13921](https://github.com/hashicorp/terraform-provider-aws/issues/13921))
* data-source/aws_ec2_transit_gateway_vpc_attachment: Add `appliance_mode_support` attribute ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* data-source/aws_route_table: Add `route` `vpc_endpoint_id` attribute ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_db_instance: Add `restore_to_point_in_time` argument and `latest_restorable_time` attribute ([#15969](https://github.com/hashicorp/terraform-provider-aws/issues/15969))
* resource/aws_default_route_table: Add `route` configuration block `vpc_endpoint_id` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_ec2_transit_gateway: Support in-place updates for most arguments ([#15556](https://github.com/hashicorp/terraform-provider-aws/issues/15556))
* resource/aws_ec2_transit_gateway_route_table: Add `arn` attribute ([#13921](https://github.com/hashicorp/terraform-provider-aws/issues/13921))
* resource/aws_ec2_transit_gateway_vpc_attachment: Add `appliance_mode_support` argument ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* resource/aws_ec2_transit_gateway_vpc_attachment_accepter: Add `appliance_mode_support` attribute ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* resource/aws_kinesis_firehose_delivery_stream: Add `http_endpoint_configuration` configuration block ([#15356](https://github.com/hashicorp/terraform-provider-aws/issues/15356))
* resource/aws_lb: Support `load_balancer_type` argument value of `gateway` ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_lb_target_group: Support `protocol` argument value of `GENEVE` ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_rds_cluster: Add `restore_to_point_in_time` argument ([#7031](https://github.com/hashicorp/terraform-provider-aws/issues/7031))
* resource/aws_route: Add `vpc_endpoint_id` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_route_table: Add `route` configuration block `vpc_endpoint_id` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_vpc_endpoint: Support `vpc_endpoint_type` argument value `GatewayLoadBalancer` ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_vpc_endpoint_service: Add `gateway_load_balancer_arns` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_workspaces_workspace: Add configurable timeouts ([#15479](https://github.com/hashicorp/terraform-provider-aws/issues/15479))

BUG FIXES

* data-source/aws_network_interface: Prevent crash with ENI attachments missing DeviceIndex or AttachmentID ([#15567](https://github.com/hashicorp/terraform-provider-aws/issues/15567))
* resource/aws_cognito_identity_pool: Update `identity_pool_name` argument validation to include additional characters supported by the API ([#15773](https://github.com/hashicorp/terraform-provider-aws/issues/15773))
* resource/aws_db_instance: Ignore `DBInstanceNotFound` error during deletion ([#15942](https://github.com/hashicorp/terraform-provider-aws/issues/15942))
* resource/aws_ecs_service: Properly remove resource from Terraform state with `ClusterNotFoundException` error ([#15927](https://github.com/hashicorp/terraform-provider-aws/issues/15927))
* resource/aws_eip: In EC2-Classic, wait until Instance returns as associated during create or update ([#16032](https://github.com/hashicorp/terraform-provider-aws/issues/16032))
* resource/aws_eip_association: Retry on additional EC2 Address eventual consistency errors on creation ([#16032](https://github.com/hashicorp/terraform-provider-aws/issues/16032))
* resource/aws_eip_association: In EC2-Classic, wait until Instance returns as associated during creation ([#16032](https://github.com/hashicorp/terraform-provider-aws/issues/16032))
* resource/aws_kinesis_analytics_application: Handle IAM role eventual consistency issues ([#16125](https://github.com/hashicorp/terraform-provider-aws/issues/16125))
* resource/aws_kinesisanalyticsv2_application: Handle IAM role eventual consistency issues ([#16125](https://github.com/hashicorp/terraform-provider-aws/issues/16125))
* resource/aws_lb_target_group: Allow invalid configurations that were allowed prior to 3.10. ([#15613](https://github.com/hashicorp/terraform-provider-aws/issues/15613))
* resource/aws_network_interface: Prevent crash with ENI attachments missing DeviceIndex or AttachmentID ([#15567](https://github.com/hashicorp/terraform-provider-aws/issues/15567))
* resource/aws_s3_bucket: Add plan-time validation to `acl` ([#15327](https://github.com/hashicorp/terraform-provider-aws/issues/15327))
* resource/aws_workspaces_bundle: Fix empty (private) owner ([#14535](https://github.com/hashicorp/terraform-provider-aws/issues/14535))

## 3.14.1 (November 06, 2020)

BUG FIXES

* resource/aws_cloudwatch_event_target: Prevent regression from version 3.14.0 with `ListTargetsByRuleInput.EventBusName` error ([#16075](https://github.com/hashicorp/terraform-provider-aws/issues/16075))

## 3.14.0 (November 06, 2020)

FEATURES

* **New Data Source:** `aws_route53_resolver_endpoint` ([#8628](https://github.com/hashicorp/terraform-provider-aws/issues/8628))
* **New Data Source:** `aws_sagemaker_prebuilt_ecr_image` ([#15924](https://github.com/hashicorp/terraform-provider-aws/pull/15924))
* **New Data Source:** `aws_workspaces_workspace` ([#14135](https://github.com/hashicorp/terraform-provider-aws/issues/14135))
* **New Resource:** `aws_secretsmanager_secret_policy` ([#14468](https://github.com/hashicorp/terraform-provider-aws/pull/14468))

ENHANCEMENTS

* resource/aws_apigatewayv2_integration: `timeout_milliseconds` has different valid ranges and default values between HTTP and WebSocket APIs. `timeout_milliseconds` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. ([#16017](https://github.com/hashicorp/terraform-provider-aws/issues/16017))
* resource/aws_cloudwatch_event_permission: Add `event_bus_name` ([#15922](https://github.com/hashicorp/terraform-provider-aws/issues/15922))
* resource/aws_cloudwatch_event_target: Add plan time validation to `arn`, `role_arn`, `launch_type`, `task_definition_arn` ([#11685](https://github.com/hashicorp/terraform-provider-aws/issues/11685))
* resource/aws_cloudwatch_event_target: Add `event_bus_name` ([#15799](https://github.com/hashicorp/terraform-provider-aws/issues/15799))
* resource/aws_codeartifact_domain: add `tags` argument. ([#16006](https://github.com/hashicorp/terraform-provider-aws/issues/16006))
* resource/aws_codeartifact_repository: add `tags` argument. ([#16006](https://github.com/hashicorp/terraform-provider-aws/issues/16006))
* resource/aws_eip: Add `network_border_group` argument ([#14028](https://github.com/hashicorp/terraform-provider-aws/issues/14028))
* resource/aws_glue_catalog_database: add plan time validations for `description` and `name`. ([#15956](https://github.com/hashicorp/terraform-provider-aws/issues/15956))
* resource/aws_glue_crawler: Support MongoDB target ([#15934](https://github.com/hashicorp/terraform-provider-aws/issues/15934))
* resource/aws_glue_trigger: Add plan time validation to `name` ([#15793](https://github.com/hashicorp/terraform-provider-aws/issues/15793))
* resource/aws_glue_trigger: Add `security_configuration` and `notification_property` arguments to `actions` block ([#15793](https://github.com/hashicorp/terraform-provider-aws/issues/15793))
* resource/aws_kinesis_analytics_application: Wait for resource deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: `inputs.parallelism` is a computed attribute. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Handle `inputs.processing_configuration` addition and deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Handle `reference_data_sources` deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Handle `cloudwatch_logging_options` deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Set the `description` attribute on creation. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_sagemaker_endpoint_configuration: Add support for `data_capture_config`. ([#15887](https://github.com/hashicorp/terraform-provider-aws/issues/15887))
* resource/aws_sagemaker_endpoint_configuration: Add plan time validation for `production_variants.accelerator_type`, `production_variants.instance_type`. ([#15887](https://github.com/hashicorp/terraform-provider-aws/issues/15887))
* resource/aws_sagemaker_model: Add support for `primary_container. image_config` and `containers.image_config` ([#15957](https://github.com/hashicorp/terraform-provider-aws/issues/15957))
* resource/aws_sagemaker_model: Add plan time validation for `execution_role_arn`  ([#15957](https://github.com/hashicorp/terraform-provider-aws/issues/15957))

BUG FIXES

* resource/aws_datasync_task: Allow `UNAVAILABLE` as pending status during creation ([#15949](https://github.com/hashicorp/terraform-provider-aws/issues/15949))
* resource/aws_glue_classifier: Fix `quote_symbol` being optional ([#15948](https://github.com/hashicorp/terraform-provider-aws/issues/15948))
* resource/aws_lambda_function: Publish version if value of `publish` is only change ([#15020](https://github.com/hashicorp/terraform-provider-aws/issues/15020))
* resource/aws_rds_cluster: Prevent error removing cluster from global cluster when not found ([#15938](https://github.com/hashicorp/terraform-provider-aws/issues/15938))
* resource/aws_rds_cluster: Prevent recreation when using `snapshot_identifier` and `kms_key_id` without `storage_encrypted = true` ([#15915](https://github.com/hashicorp/terraform-provider-aws/issues/15915))
* resource/aws_rds_cluster_instance: Add Cluster Identifier to creation error message ([#15939](https://github.com/hashicorp/terraform-provider-aws/issues/15939))
* resource/aws_rds_global_cluster: Prevent error removing cluster from global cluster when not found ([#15938](https://github.com/hashicorp/terraform-provider-aws/issues/15938))

## 3.13.0 (October 29, 2020)

NOTES

* data-source/aws_autoscaling_groups: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_caller_identity: The `id` attribute has changed to the ID of the AWS Account. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_snapshot_ids: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_volumes: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_coip_pools: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_instance_type_offerings: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_route_tables: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateways: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_spot_price: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_efs_access_points: The `id` attribute has changed to the EFS File System identifier. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_glue_script: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_inspector_rules_packages: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_instances: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_ciphertext: The `id` attribute has changed to the KMS Key. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_secrets: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15725](https://github.com/hashicorp/terraform-provider-aws/issues/15725))
* data-source/aws_network_acls: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_network_interfaces: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_organizations_organizational_units: The `id` attribute has changed to the parent identifier. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_outposts: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_sites: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route_tables: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route53_resolver_rules: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_s3_bucket_objects: The `id` attribute has changed to the name of the S3 Bucket. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_security_groups: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpc_peering_connections: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpcs: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))

FEATURES

* **New Resource:** `aws_glue_resource_policy` ([#10361](https://github.com/hashicorp/terraform-provider-aws/issues/10361))
* **New Resource:** `aws_s3control_bucket` ([#15510](https://github.com/hashicorp/terraform-provider-aws/issues/15510))
* **New Resource:** `aws_s3control_bucket_lifecycle_configuration` ([#15604](https://github.com/hashicorp/terraform-provider-aws/issues/15604))
* **New Resource:** `aws_s3control_bucket_policy` ([#15575](https://github.com/hashicorp/terraform-provider-aws/issues/15575))
* **New Resource:** `aws_s3outposts_endpoint` ([#15585](https://github.com/hashicorp/terraform-provider-aws/issues/15585))
* **New Resource:** `aws_sagemaker_code_repository` ([#15809](https://github.com/hashicorp/terraform-provider-aws/issues/15809))
* **New Resource:** `aws_storagegateway_tape_pool` ([#15370](https://github.com/hashicorp/terraform-provider-aws/issues/15370))

ENHANCEMENTS

* resource/aws_cloudwatch_event_rule: Add `event_bus_name` ([#15727](https://github.com/hashicorp/terraform-provider-aws/issues/15727))
* resource/aws_ecs_service: Add `wait_for_steady_state` argument ([#3485](https://github.com/hashicorp/terraform-provider-aws/issues/3485))
* resource/aws_s3_access_point: Support S3 on Outposts ([#15621](https://github.com/hashicorp/terraform-provider-aws/issues/15621))
* resource/aws_sagemaker_model: Add `container` configuration block `mode` argument ([#15371](https://github.com/hashicorp/terraform-provider-aws/issues/15371))
* resource/aws_sagemaker_notebook_instance: Add support for `additional_code_repositories` ([#15830](https://github.com/hashicorp/terraform-provider-aws/issues/15830))
* resource/aws_sagemaker_notebook_instance: Add `url` and `network_interface_id` attributes ([#15802](https://github.com/hashicorp/terraform-provider-aws/issues/15802))

BUG FIXES

* data-source/aws_autoscaling_groups: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_caller_identity: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_snapshot_ids: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_volumes: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_coip_pools: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_instance_type_offerings: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_route_tables: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateways: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_spot_price: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_efs_access_points: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_glue_script: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_inspector_rules_packages: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_instances: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_ciphertext: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_secrets: Prevent plan differences with the `id` attribute ([#15725](https://github.com/hashicorp/terraform-provider-aws/issues/15725))
* data-source/aws_network_acls: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_network_interfaces: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_organizations_organizational_units: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_outposts: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_sites: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route_tables: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route53_resolver_rules: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_s3_bucket_objects: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_security_groups: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpc_peering_connections: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpcs: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* resource/aws_apigatewayv2_integration: Correctly handle update of AWS service integrations ([#15894](https://github.com/hashicorp/terraform-provider-aws/issues/15894))
* resource/aws_api_gateway_usage_plan: Change `api_stages` to from List to Set ([#14345](https://github.com/hashicorp/terraform-provider-aws/issues/14345))
* resource/aws_lambda_function: Update published `version` and `qualified_arn` on config changes ([#15121](https://github.com/hashicorp/terraform-provider-aws/issues/15121))
* resource/aws_rds_global_cluster: Prevent recreation when using encrypted `source_db_cluster_identifier` without `storage_encrypted` ([#15916](https://github.com/hashicorp/terraform-provider-aws/issues/15916))
* resource/aws_vpc_peering_connection_options: Only modify options that have changed ([#12126](https://github.com/hashicorp/terraform-provider-aws/issues/12126))

## 3.12.0 (October 22, 2020)

FEATURES

* **New Data Source:** `aws_rds_certificate` ([#15789](https://github.com/hashicorp/terraform-provider-aws/issues/15789))
* **New Resource:** `aws_autoscalingplans_scaling_plan` ([#8965](https://github.com/hashicorp/terraform-provider-aws/issues/8965))
* **New Resource:** `aws_cloudwatch_event_bus` ([#10256](https://github.com/hashicorp/terraform-provider-aws/issues/10256))
* **New Resource:** `aws_kinesisanalyticsv2_application` ([#11652](https://github.com/hashicorp/terraform-provider-aws/issues/11652))
* **New Resource:** `aws_storagegateway_stored_iscsi_volume` ([#12027](https://github.com/hashicorp/terraform-provider-aws/issues/12027))

ENHANCEMENTS

* resource/aws_cloudwatch_event_target: Add validation to `input_transformer.input_paths` map ([#15669](https://github.com/hashicorp/terraform-provider-aws/issues/15669))
* resource/aws_codeartifact_repository - support external connections ([#15569](https://github.com/hashicorp/terraform-provider-aws/issues/15569))
* resource/aws_fsx_lustre_file_system: Add `copy_tags_to_backups` support ([#15687](https://github.com/hashicorp/terraform-provider-aws/issues/15687))
* resource/aws_fsx_lustre_file_system: Increased maximum `automatic_backup_retention_days` from 35 to 90 ([#15641](https://github.com/hashicorp/terraform-provider-aws/issues/15641))
* resource/aws_fsx_windows_file_system: Increased maximum `automatic_backup_retention_days` from 35 to 90 ([#15641](https://github.com/hashicorp/terraform-provider-aws/issues/15641))
* resource/aws_glue_catalog_table: add validation checks for resource properties ([#12523](https://github.com/hashicorp/terraform-provider-aws/issues/12523))
* resource/aws_network_interface: Add `ipv6_addresses` and `ipv6_address_count` arguments ([#12281](https://github.com/hashicorp/terraform-provider-aws/issues/12281))
* resource/aws_sagemaker_notebook_instance: `lifecycle_config_name` and `root_access`  are updateable. ([#15385](https://github.com/hashicorp/terraform-provider-aws/issues/15385))
* resource/aws_sagemaker_notebook_instance: plan time validation for `role_arn`, `instance_type`. ([#15385](https://github.com/hashicorp/terraform-provider-aws/issues/15385))

BUGFIXES

* resource/aws_workspaces_workspace: Fix terminated state resolution ([#15705](https://github.com/hashicorp/terraform-provider-aws/issues/15705))
* resource/aws_glue_table_catalog_table: Prevent errors on `unset` of `ser_de_info.name` ([#15127](https://github.com/hashicorp/terraform-provider-aws/issues/15127))
* resource/aws_glue_security_configuration: Don't send empty `kms_arn` if mode is `DISABLED` ([#13618](https://github.com/hashicorp/terraform-provider-aws/issues/13618))

## 3.11.0 (October 15, 2020)

FEATURES

* **New Data Source:** `aws_codeartifact_repository_endpoint` ([#15566](https://github.com/hashicorp/terraform-provider-aws/issues/15566))
* **New Resource:** `aws_appmesh_gateway_route` ([#15638](https://github.com/hashicorp/terraform-provider-aws/issues/15638))
* **New Resource:** `aws_appmesh_virtual_gateway` ([#15611](https://github.com/hashicorp/terraform-provider-aws/issues/15611))

BUG FIXES

* resource/aws_ec2_transit_gateway_route: Prevent plan errors with compressed IPv6 addresses ([#14846](https://github.com/hashicorp/terraform-provider-aws/issues/14846))

ENHANCEMENTS

* data-source/aws_workspaces_directory: Add workspaces creation properties ([#14577](https://github.com/hashicorp/terraform-provider-aws/issues/14577))
* resource/aws_backup_plan: Add support for AdvancedBackupSettings ([#15341](https://github.com/hashicorp/terraform-provider-aws/issues/15341))
* resource/aws_sagemaker_notebook_instance: Add `default_code_repository` attribute ([#13772](https://github.com/hashicorp/terraform-provider-aws/issues/13772))
* resource/aws_sagemaker_notebook_instance: Add `volume_size` attribute ([#15521](https://github.com/hashicorp/terraform-provider-aws/issues/15521))
* resource/aws_workspaces_directory: Add workspaces creation properties ([#14577](https://github.com/hashicorp/terraform-provider-aws/issues/14577))

## 3.10.0 (October 09, 2020)

FEATURES

* **New Data Source:** `aws_codeartifact_authorization_token` ([#15425](https://github.com/hashicorp/terraform-provider-aws/issues/15425))
* **New Data Source:** `aws_ec2_instance_type` ([#13124](https://github.com/hashicorp/terraform-provider-aws/issues/13124))
* **New Data Source:** `aws_lex_bot_alias` ([#8919](https://github.com/hashicorp/terraform-provider-aws/issues/8919))
* **New Data Source:** `aws_redshift_orderable_cluster` ([#15438](https://github.com/hashicorp/terraform-provider-aws/issues/15438))
* **New Resource:** `aws_codeartifact_repository_permissions_policy` ([#15562](https://github.com/hashicorp/terraform-provider-aws/issues/15562))
* **New Resource:** `aws_lex_bot_alias` ([#8919](https://github.com/hashicorp/terraform-provider-aws/issues/8919))
* **New Resource:** `aws_s3_bucket_ownership_controls` ([#15482](https://github.com/hashicorp/terraform-provider-aws/issues/15482))

NOTES

* data-source/aws_acm_certificate: The `id` attribute has changed to the ARN of the ACM Certificate. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_autoscaling_group: The `id` attribute has changed to the name of the Auto Scaling Group. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_availability_zones: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_db_event_categories: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_default_kms_key: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_encryption_by_default: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ec2_instance_type_offering: The `id` attribute has changed to the EC2 Instance Type. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_authorization_token: The `id` attribute has changed to the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_image: The `id` attribute has changed to the SHA256 digest of the ECR Image. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_eks_cluster_auth: The `id` attribute has changed to the name of the EKS Cluster. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_iam_account_alias: The `id` attribute has changed to the AWS Account Alias. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_kms_alias: The `id` attribute has changed to the ARN of the KMS Alias. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_partition: The `id` attribute has changed to the identifier of the AWS Partition. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_regions: The `id` attribute has changed to the identifier of the AWS Partition. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_sns_topic: The `id` attribute has changed to the ARN of the SNS Topic. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))

ENHANCEMENTS

* data-source/aws_batch_compute_environment: Add `tags` attribute ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* data-source/aws_batch_job_queue: Add `tags` attribute ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* data-source/aws_vpc_endpoint_service: Accept `service_type` as argument ([#15467](https://github.com/hashicorp/terraform-provider-aws/issues/15467))
* resource/aws_appmesh_route: Add `timeout` configuration block to `grpc_route`, `http_route`, `http2_route` and `tcp_route` attributes. ([#14361](https://github.com/hashicorp/terraform-provider-aws/issues/14361))
* resource/aws_appmesh_virtual_node: Add `timeout` configuration block to `listener` attribute. ([#14361](https://github.com/hashicorp/terraform-provider-aws/issues/14361))
* resource/aws_batch_compute_environment: Add `tags` argument ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_batch_job_definition: Add `tags` argument ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_batch_job_queue: Add `tags` argument ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_lb_target_group: Add `source_ip` as an option for the `stickiness.type` argument. ([#15295](https://github.com/hashicorp/terraform-provider-aws/issues/15295))
* resource/aws_sns_topic_subscription: Create subscriptions with attributes (delivery policy, filter policy, etc.) instead of separate API calls ([#10496](https://github.com/hashicorp/terraform-provider-aws/issues/10496))

BUG FIXES

* data-source/aws_acm_certificate: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_autoscaling_group: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_availability_zones: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_db_event_categories: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_default_kms_key: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_encryption_by_default: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ec2_instance_type_offering: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_authorization_token: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_image: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_eks_cluster_auth: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_iam_account_alias: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_kms_alias: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_partition: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_regions: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_sns_topic: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* resource/aws_acm_certificate: Prevent unexpected timeout error on deletion due to API retries ([#15522](https://github.com/hashicorp/terraform-provider-aws/issues/15522))
* resource/aws_batch_job_definition: Prevent unexpected plan difference for `container_properties` argument value with new secrets support ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_codestarnotifications_notification_rule: Prevent unexpected timeout error during target deletion due to API retries ([#15523](https://github.com/hashicorp/terraform-provider-aws/issues/15523))
* resource/aws_config_remediation_configuration: Prevent unexpected timeout error on deletion due to API retries ([#15524](https://github.com/hashicorp/terraform-provider-aws/issues/15524))
* resource/aws_db_proxy: Increase default deletion timeout to 60 minutes ([#15537](https://github.com/hashicorp/terraform-provider-aws/issues/15537))
* resource/aws_db_proxy_target: Ensure `db_proxy_name` and `target_group_name` attributes are properly imported ([#15537](https://github.com/hashicorp/terraform-provider-aws/issues/15537))
* resource/aws_ecs_cluster: Prevent IAM Service Linked Role error on first ECS provision ([#15457](https://github.com/hashicorp/terraform-provider-aws/issues/15457))
* resource/aws_emr_instance_fleet: Prevent error on deletion when EMR Cluster is no longer running ([#15548](https://github.com/hashicorp/terraform-provider-aws/issues/15548))
* resource/aws_emr_managed_scaling_policy: Ensure `cluster_id` attribute is properly saved during import ([#15541](https://github.com/hashicorp/terraform-provider-aws/issues/15541))
* resource/aws_emr_managed_scaling_policy: Handle additional cases where resource should be removed from Terraform state ([#15541](https://github.com/hashicorp/terraform-provider-aws/issues/15541))
* resource/aws_gamelift_fleet: Prevent unexpected timeout error on creation due to API retries ([#15526](https://github.com/hashicorp/terraform-provider-aws/issues/15526))
* resource/aws_glue_workflow: Ensure `max_concurrent_runs` attribute is properly saved during import ([#15538](https://github.com/hashicorp/terraform-provider-aws/issues/15538))
* resource/aws_lex_bot: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_lex_bot_alias: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_lex_intent: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_lex_slot_type: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_organizations_policy: Prevent errors with imported AWS-managed Organizations policies ([#15446](https://github.com/hashicorp/terraform-provider-aws/issues/15446))
* resource/aws_s3_bucket: Correctly handle provider-level ignored tag configuration ([#12013](https://github.com/hashicorp/terraform-provider-aws/issues/12013))
* resource/aws_s3_bucket: Correctly set expiration for lifecycle_rule with abort_incomplete_multipart_upload_days set ([#15263](https://github.com/hashicorp/terraform-provider-aws/issues/15263))
* resource/aws_s3_bucket_analytics_configuration: Prevent unexpected timeout error on deletion due to API retries ([#15529](https://github.com/hashicorp/terraform-provider-aws/issues/15529))
* resource/aws_s3_bucket_object: Correctly handle provider-level ignored tag configuration ([#12013](https://github.com/hashicorp/terraform-provider-aws/issues/12013))

## 3.9.0 (October 02, 2020)

FEATURES

* **New Resource:** `aws_backup_vault_notifications` ([#12501](https://github.com/hashicorp/terraform-provider-aws/issues/12501))
* **New Resource:** `aws_codeartifact_domain` ([#13743](https://github.com/hashicorp/terraform-provider-aws/issues/13743))
* **New Resource:** `aws_codeartifact_domain_permissions` ([#13753](https://github.com/hashicorp/terraform-provider-aws/issues/13753))
* **New Resource:** `aws_codeartifact_repository` ([#14429](https://github.com/hashicorp/terraform-provider-aws/issues/14429))
* **New Resource:** `aws_db_proxy_target` ([#12784](https://github.com/hashicorp/terraform-provider-aws/issues/12784))
* **New Resource:** `aws_glue_data_catalog_encryption_settings` ([#14916](https://github.com/hashicorp/terraform-provider-aws/issues/14916))
* **New Resource:** `aws_glue_ml_transform` ([#14909](https://github.com/hashicorp/terraform-provider-aws/issues/14909))
* **New Resource:** `aws_glue_partition` ([#12547](https://github.com/hashicorp/terraform-provider-aws/issues/12547))
* **New Resource:** `aws_lex_bot` ([#8918](https://github.com/hashicorp/terraform-provider-aws/issues/8918))
* **New Resource:** `aws_lex_intent` ([#8917](https://github.com/hashicorp/terraform-provider-aws/issues/8917))
* **New Data Source:** `aws_lex_bot` ([#8918](https://github.com/hashicorp/terraform-provider-aws/issues/8918))
* **New Data Source:** `aws_lex_intent` ([#8917](https://github.com/hashicorp/terraform-provider-aws/issues/8917))

ENHANCEMENTS

* resource/aws_appmesh_route: Add `grpc_route` and `http2_route` attributes to support gRPC and HTTP/2 services ([#11669](https://github.com/hashicorp/terraform-provider-aws/issues/11669))
* resource/aws_appmesh_route: Add `retry_policy` attribute to support App Mesh retry policies ([#11660](https://github.com/hashicorp/terraform-provider-aws/issues/11660))
* resource/aws_appmesh_virtual_node: Add `grpc` and `http2` as valid values for the `protocol` attribute ([#11669](https://github.com/hashicorp/terraform-provider-aws/issues/11669))
* resource/aws_appmesh_virtual_node: Add `spec.backend_defaults`, `spec.backend.virtual_service.client_policy` and `spec.listener.tls` attributes to support TLS in transit encryption ([#12541](https://github.com/hashicorp/terraform-provider-aws/issues/12541))
* resource/aws_appmesh_virtual_router: Add `grpc` and `http2` as valid values for the `protocol` attribute ([#11669](https://github.com/hashicorp/terraform-provider-aws/issues/11669))
* resource/aws_fsx_lustre_file_system: Add `auto_import_policy`  argument ([#15231](https://github.com/hashicorp/terraform-provider-aws/issues/15231))
* resource/aws_fsx_lustre_file_system: Support `daily_automatic_backup_start_time` ([#15299](https://github.com/hashicorp/terraform-provider-aws/issues/15299))
* resource/aws_fsx_lustre_file_system: Add `storage_type` and `drive_cache_type` ([#14727](https://github.com/hashicorp/terraform-provider-aws/issues/14727))
* resource/aws_glue_crawler: Add `connection_name` field to `s3_target` block ([#15350](https://github.com/hashicorp/terraform-provider-aws/issues/15350))
* resource/aws_sagemaker_notebook_instance: Ability to configure root access for Sagemaker notebook instances ([#14184](https://github.com/hashicorp/terraform-provider-aws/issues/14184))

BUG FIXES

* data-source/aws_s3_bucket_object: Prevent crash when S3 HeadObject returns empty response ([#14154](https://github.com/hashicorp/terraform-provider-aws/issues/14154))
* resource/aws_db_instance: Prevent ordering differences with `enabled_cloudwatch_logs_exports` argument ([#15404](https://github.com/hashicorp/terraform-provider-aws/issues/15404))
* resource/aws_ec2_client_vpn_authorization_rule: Increased active and revoked timeouts from 5 to 10 minutes ([#15367](https://github.com/hashicorp/terraform-provider-aws/issues/15367))
* resource/aws_rds_cluster: Prevent ordering differences with `enabled_cloudwatch_logs_exports` argument ([#15404](https://github.com/hashicorp/terraform-provider-aws/issues/15404))
* resource/aws_redshift_cluster: Increase default update timeout to 75 minutes ([#15339](https://github.com/hashicorp/terraform-provider-aws/issues/15339))

## 3.8.0 (September 24, 2020)

FEATURES

* **New Resource:** `aws_datasync_location_fsx_windows` ([#12686](https://github.com/hashicorp/terraform-provider-aws/pull/12686))
* **New Resource:** `aws_route53_resolver_query_log_config`. ([#14897](https://github.com/hashicorp/terraform-provider-aws/pull/14897))
* **New Resource:** `aws_route53_resolver_query_log_config_association`. ([#14901](https://github.com/hashicorp/terraform-provider-aws/pull/14901))
* **New Data Source:** `aws_rds_engine_version` ([#15228](https://github.com/hashicorp/terraform-provider-aws/pull/15228))
* **New Data Source:** `aws_docdb_engine_version` ([#15253](https://github.com/hashicorp/terraform-provider-aws/pull/15253))
* **New Data Source:** `aws_neptune_engine_version` ([#15259](https://github.com/hashicorp/terraform-provider-aws/pull/15259))
* **New Data Source:** `aws_workspaces_image` ([#11428](https://github.com/hashicorp/terraform-provider-aws/issues/11428))

ENHANCEMENTS

* data-source/aws_lb: Add `customer_owned_ipv4_pool` and `subnet_mapping` `outpost_id` attributes ([#15170](https://github.com/hashicorp/terraform-provider-aws/issues/15170))
* resource/aws_apigatewayv2_api: Add `disable_execute_api_endpoint` attribute ([#15250](https://github.com/hashicorp/terraform-provider-aws/issues/15250))
* resource/aws_apigatewayv2_authorizer: Add `authorizer_payload_format_version`, `authorizer_result_ttl_in_seconds` and `enable_simple_responses` attribute to support Lambda authorizers for HTTP APIs ([#15232](https://github.com/hashicorp/terraform-provider-aws/issues/15232))
* resource/aws_apigatewayv2_authorizer: Change `identity_sources` to an optional attribute ([#15232](https://github.com/hashicorp/terraform-provider-aws/issues/15232))
* resource/aws_appmesh_mesh: Add `mesh_owner` and `resource_owner` attributes ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_route: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_virtual_node: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_virtual_router: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_virtual_service: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_elasticsearch_domain: Support `AUDIT_LOGS` log type ([#15218](https://github.com/hashicorp/terraform-provider-aws/issues/15218))
* resource/aws_glue_connection: Support `NETWORK` connection type ([#14818](https://github.com/hashicorp/terraform-provider-aws/issues/14818))
* resource/aws_glue_crawler: Add support for `scan_all` and `scan_rate` arguments for ddb targets ([#14819](https://github.com/hashicorp/terraform-provider-aws/issues/14819))
* resource/aws_glue_crawler: Allow removing `table_prefix` ([#15268](https://github.com/hashicorp/terraform-provider-aws/issues/15268))
* resource/aws_glue_job: Add `non_overridable_arguments` argument ([#14793](https://github.com/hashicorp/terraform-provider-aws/issues/14793))
* resource/aws_glue_workflow: Add `tags` argument ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_glue_workflow: Add `arn` attribute ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_glue_workflow: Add `max_concurrent_runs` argument ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_glue_workflow: Plan time validation for `name` ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_fsx_lustre_file_system: Add support for backup retention ([#14446](https://github.com/hashicorp/terraform-provider-aws/issues/14446))
* resource/aws_fsx_lustre_file_system: Add `kms_key_id` argument ([#15057](https://github.com/hashicorp/terraform-provider-aws/issues/15057))
* resource/aws_fsx_lustre_file_system: Add `mount_name` argument ([#14313](https://github.com/hashicorp/terraform-provider-aws/issues/14313))
* resource/aws_lb: Add `customer_owned_ipv4_pool` argument and `subnet_mapping` `outpost_id` attribute ([#15170](https://github.com/hashicorp/terraform-provider-aws/issues/15170))
* resource/aws_organizations_policy: Add `tags` argument ([#15316](https://github.com/hashicorp/terraform-provider-aws/issues/15316))
* resource/aws_rds_cluster: Add `allow_major_version_upgrade` argument ([#14709](https://github.com/hashicorp/terraform-provider-aws/issues/14709))
* resource/aws_storagegateway_smb_file_share: Add `admin_user_list` argument ([#12196](https://github.com/hashicorp/terraform-provider-aws/issues/12196))
* resource/aws_transfer_server: Support `VPC` value for `endpoint_type` argument and add `endpoint_details` configuration block `address_allocation_ids`, `subnet_ids`, and `vpc_id` arguments ([#12599](https://github.com/hashicorp/terraform-provider-aws/issues/12599))
* resource/aws_transfer_user: Add `home_directory_mappings` configuration blocks and `home_directory_type` argument ([#13591](https://github.com/hashicorp/terraform-provider-aws/issues/13591))

BUG FIXES

* resource/aws_dynamodb_table: Ensure changes in `name`, `range_key`, `projection_type`, or `non_key_attributes` of a `local_secondary_index` configuration block force resource recreation ([#12335](https://github.com/hashicorp/terraform-provider-aws/issues/12335))
* resource/aws_dynamodb_table: Ensure `local_secondary_index` `non_key_attributes` are sent through API requests on resource creation ([#15115](https://github.com/hashicorp/terraform-provider-aws/issues/15115))
* resource/aws_efs_mount_target: Increase create timeout to 30 minutes ([#15293](https://github.com/hashicorp/terraform-provider-aws/issues/15293))
* resource/aws_fsx_lustre_file_system: Change `aws_fsx_lustre_file_system's`'s `network_interface_ids` to `TypeList` to preserve ordering. ([#14314](https://github.com/hashicorp/terraform-provider-aws/issues/14314))
* resource/aws_neptune_cluster_instance: Add `configuring-enhanced-monitoring` to expected states when creating and updating ([#15284](https://github.com/hashicorp/terraform-provider-aws/issues/15284))
* resource/aws_vpn_gateway: Increase VPC detachment timeout to 30 minutes ([#15201](https://github.com/hashicorp/terraform-provider-aws/issues/15201))
* resource/aws_vpn_gateway_attachment: Increase VPC detachment timeout to 30 minutes ([#15201](https://github.com/hashicorp/terraform-provider-aws/issues/15201))

## 3.7.0 (September 17, 2020)

FEATURES

* **New Resource:** `aws_config_remediation_configuration` ([#13884](https://github.com/hashicorp/terraform-provider-aws/issues/13884))

ENHANCEMENTS

* resource/aws_db_cluster_snapshot: Add plan-time validation for `db_cluster_snapshot_identifier` argument ([#15132](https://github.com/hashicorp/terraform-provider-aws/issues/15132))
* resource/aws_kinesis_firehose_delivery_stream: Add `server_side_encryption` `key_arn` and `key_type` arguments (support KMS Customer Managed Key encryption) ([#11954](https://github.com/hashicorp/terraform-provider-aws/issues/11954))

BUG FIXES

* data-source/aws_kms_secrets: Prevent `plaintext` values to appear in CLI output with Terraform 0.13 ([#15169](https://github.com/hashicorp/terraform-provider-aws/issues/15169))
* resource/aws_acm_certificate: Prevent tagging is not permitted on re-import error ([#15060](https://github.com/hashicorp/terraform-provider-aws/issues/15060))
* resource/aws_cognito_identity_pool: Prevent ordering differences for `openid_connect_provider_arns` argument ([#15178](https://github.com/hashicorp/terraform-provider-aws/issues/15178))

## 3.6.0 (September 11, 2020)

FEATURES

* **New Resource:** `aws_db_proxy_default_target_group` ([#12743](https://github.com/hashicorp/terraform-provider-aws/issues/12743))

BUG FIXES

* resource/aws_ec2_client_vpn_authorization_rule: Increase active and revoked timeouts from 1 to 5 minutes ([#15037](https://github.com/hashicorp/terraform-provider-aws/issues/15037))

## 3.5.0 (September 03, 2020)

FEATURES

* **New Data Source:** `aws_docdb_orderable_db_instance` ([#14931](https://github.com/hashicorp/terraform-provider-aws/issues/14931))
* **New Data Source:** `aws_lex_slot_type` ([#8916](https://github.com/hashicorp/terraform-provider-aws/issues/8916))
* **New Data Source:** `aws_neptune_orderable_db_instance` ([#14953](https://github.com/hashicorp/terraform-provider-aws/issues/14953))
* **New Data Source:** `aws_rds_orderable_db_instance` ([#14834](https://github.com/hashicorp/terraform-provider-aws/issues/14834))
* **New Data Source:** `aws_vpc_peering_connections` ([#9491](https://github.com/hashicorp/terraform-provider-aws/issues/9491))
* **New Resource:** `aws_codebuild_report_group` ([#12573](https://github.com/hashicorp/terraform-provider-aws/issues/12573))
* **New Resource:** `aws_db_proxy` ([#12704](https://github.com/hashicorp/terraform-provider-aws/issues/12704))
* **New Resource:** `aws_emr_instance_fleet` ([#14813](https://github.com/hashicorp/terraform-provider-aws/issues/14813))
* **New Resource:** `aws_glue_user_defined_function` ([#12537](https://github.com/hashicorp/terraform-provider-aws/issues/12537))
* **New Resource:** `aws_guardduty_filter` ([#14876](https://github.com/hashicorp/terraform-provider-aws/issues/14876))
* **New Resource:** `aws_lex_slot_type` ([#8916](https://github.com/hashicorp/terraform-provider-aws/issues/8916))

ENHANCEMENTS

* data-source/aws_cur_report_definition: Add `refresh_closed_reports` and `report_versioning` attributes ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* data-source/aws_outposts_outpost: Add `arn` argument ([#14967](https://github.com/hashicorp/terraform-provider-aws/issues/14967))
* data-source/aws_route: Add `local_gateway_id` attribute ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* data-source/aws_route_table: Add `route` `local_gateway_id` attribute ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* resource/aws_acm_certificate: Provide additional plan-time validation for `subject_alternative_names` argument values ([#14782](https://github.com/hashicorp/terraform-provider-aws/issues/14782))
* resource/aws_ami: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_autoscaling_group: Support provider-level `ignore_tags` configuration ([#13868](https://github.com/hashicorp/terraform-provider-aws/issues/13868))
* resource/aws_cloudtrail: Add `insight_selector` configuration block ([#12390](https://github.com/hashicorp/terraform-provider-aws/issues/12390))
* resource/aws_cur_report_definition: Add `refresh_closed_reports` and `report_versioning` arguments ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `ATHENA` value in `additional_artifacts` argument plan-time validation ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `Parquet` value in `compression` and `format` argument plan-time validations ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `MONTHLY` value in `time_unit` argument plan-time validation ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_ebs_volume: Support io2 type ([#14894](https://github.com/hashicorp/terraform-provider-aws/issues/14894))
* resource/aws_ec2_client_vpn_endpoint: Support `authentication_options` `type` argument `federated-authentication` value and new `saml_provider_arn` argument ([#14171](https://github.com/hashicorp/terraform-provider-aws/issues/14171))
* resource/aws_emr_cluster: Add `core_instance_fleet` and `master_instance_fleet` configuration blocks ([#14788](https://github.com/hashicorp/terraform-provider-aws/issues/14788))
* resource/aws_instance: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_kinesis_firehose_delivery_stream: Add `elasticsearch_configuration` `vpc_config` configuration block ([#13269](https://github.com/hashicorp/terraform-provider-aws/issues/13269))
* resource/aws_kinesis_firehose_delivery_stream: Add `elasticsearch_configuration` `cluster_endpoint` argument ([#12484](https://github.com/hashicorp/terraform-provider-aws/issues/12484))
* resource/aws_kinesis_firehose_delivery_stream: Add various plan-time validations for arguments ([#12484](https://github.com/hashicorp/terraform-provider-aws/issues/12484))
* resource/aws_launch_template: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_msk_configuration: Support resource in-place updates and deletion ([#14826](https://github.com/hashicorp/terraform-provider-aws/issues/14826))
* resource/aws_route: Add `local_gateway_id` argument ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* resource/aws_route_table: Add `route` `local_gateway_id` argument ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* resource/aws_spot_fleet_request: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_wafv2_rule_group: Add `ip_set_forwarded_ip_config` configuration block to `ip_set_reference_statement` ([#14902](https://github.com/hashicorp/terraform-provider-aws/issues/14902))
* resource/aws_wafv2_web_acl: Add `ip_set_forwarded_ip_config` configuration block to `ip_set_reference_statement` ([#14902](https://github.com/hashicorp/terraform-provider-aws/issues/14902))

BUG FIXES

* resource/aws_autoscaling_group: Prevent unnecessary tag removal and recreation within tag updates ([#13868](https://github.com/hashicorp/terraform-provider-aws/issues/13868))
* resource/aws_cloudfront_distribution: Prevent panic with missing `ForwardedValues` ([#14993](https://github.com/hashicorp/terraform-provider-aws/issues/14993))
* resource/aws_dynamodb_table: Properly update `global_secondary_index` `non_key_attributes` values ([#9988](https://github.com/hashicorp/terraform-provider-aws/issues/9988))
* resource/aws_emr_cluster: Prevent recreation when `ebs_config.volumes_per_instance` is greater than 1 ([#14858](https://github.com/hashicorp/terraform-provider-aws/issues/14858))
* resource/aws_lambda_function_event_invoke_config: Prevent unexpected format of function resource error ([#14851](https://github.com/hashicorp/terraform-provider-aws/issues/14851))
* resource/aws_lightsail_instance: Prevent panic with key-only tags ([#13868](https://github.com/hashicorp/terraform-provider-aws/issues/13868))
* resource/aws_mq_configuration: Prevent additional revision creation with `tags` only updates ([#14850](https://github.com/hashicorp/terraform-provider-aws/issues/14850))
* resource/aws_opsworks_stack: Suppress equivalent `custom_json` differences ([#14886](https://github.com/hashicorp/terraform-provider-aws/issues/14886))
* resource/aws_rds_cluster_endpoint: Increase creation timeout to 30 minutes ([#14862](https://github.com/hashicorp/terraform-provider-aws/issues/14862))
* resource/aws_route53_resolver_rule: Correct handling for single period (`.`) value in `domain_name` argument ([#15015](https://github.com/hashicorp/terraform-provider-aws/issues/15015))
* resource/aws_route53_zone_association: Correctly handle zones with over 100 VPC associations ([#14885](https://github.com/hashicorp/terraform-provider-aws/issues/14885))
* resource/aws_waf_rate_based_rule: Properly update `rate_limit` value ([#14964](https://github.com/hashicorp/terraform-provider-aws/issues/14964))
* resource/aws_workspaces_workspace: Prevent error when `workspace_properties` `running_mode` is set to `ALWAYS_ON` ([#13976](https://github.com/hashicorp/terraform-provider-aws/issues/13976))

## 3.4.0 (August 27, 2020)

FEATURES

* **New Data Source:** `aws_db_subnet_group` ([#9525](https://github.com/hashicorp/terraform-provider-aws/issues/9525))
* **New Resource:** `aws_emr_managed_scaling_policy` ([#13965](https://github.com/hashicorp/terraform-provider-aws/issues/13965))
* **New Resource:** `aws_guardduty_publishing_destination` ([#13894](https://github.com/hashicorp/terraform-provider-aws/issues/13894))
* **New Resource:** `aws_securityhub_action_target` ([#10493](https://github.com/hashicorp/terraform-provider-aws/issues/10493))
* **New Resource:** `aws_xray_encryption_config` ([#13600](https://github.com/hashicorp/terraform-provider-aws/issues/13600))
* **New Resource:** `aws_xray_group` ([#13597](https://github.com/hashicorp/terraform-provider-aws/issues/13597))

ENHANCEMENTS

* resource/aws_apigatewayv2_integration: Add `integration_subtype` argument (Support AWS service integrations for HTTP APIs) ([#14860](https://github.com/hashicorp/terraform-provider-aws/issues/14860))
* resource/aws_elasticache_replication_group: Add plan-time validation for `notification_topic_arn` and `snapshot_arns` arguments ([#12974](https://github.com/hashicorp/terraform-provider-aws/issues/12974))
* resource/aws_globalaccelerator_endpoint_group: Add `client_ip_preservation_enabled` argument to the `endpoint_configuration` configuration block ([#14486](https://github.com/hashicorp/terraform-provider-aws/issues/14486))
* resource/aws_storagegateway_cached_iscsi_volume: Add `kms_encrypted` and `kms_key` arguments ([#12066](https://github.com/hashicorp/terraform-provider-aws/issues/12066))
* resource/aws_storagegateway_gateway: Add `smb_security_strategy` argument ([#13563](https://github.com/hashicorp/terraform-provider-aws/issues/13563))
* resource/aws_storagegateway_gateway: Add plan-time validation for `gateway_ip_address` argument ([#13563](https://github.com/hashicorp/terraform-provider-aws/issues/13563))
* resource/aws_storagegateway_gateway: Add `average_download_rate_limit_in_bits_per_sec` and `average_upload_rate_limit_in_bits_per_sec` arguments ([#13568](https://github.com/hashicorp/terraform-provider-aws/issues/13568))
* resource/aws_storagegateway_nfs_file_share: Add `cache_attributes` configuration block ([#14759](https://github.com/hashicorp/terraform-provider-aws/issues/14759))
* resource/aws_storagegateway_nfs_file_share: Support `S3_INTELLIGENT_TIERING` value in `default_storage_class` argument plan-time validation ([#14759](https://github.com/hashicorp/terraform-provider-aws/issues/14759))
* resource/aws_storagegateway_smb_file_share: Add `cache_attributes` configuration block and `case_sensitivity` argument ([#14790](https://github.com/hashicorp/terraform-provider-aws/issues/14790))
* resource/aws_storagegateway_smb_file_share: Support `S3_INTELLIGENT_TIERING` value in `default_storage_class` argument plan-time validation ([#14790](https://github.com/hashicorp/terraform-provider-aws/issues/14790))
* resource/aws_xray_sampling_rule: Add `tags` argument ([#14831](https://github.com/hashicorp/terraform-provider-aws/issues/14831))

BUG FIXES

* resource/aws_acmpca_certificate_authority: Ensure `DELETED` status triggers state removal ([#13684](https://github.com/hashicorp/terraform-provider-aws/issues/13684))
* resource/aws_appmesh_virtual_node: Prevent panics with empty `backend` configuration blocks ([#14074](https://github.com/hashicorp/terraform-provider-aws/issues/14074))
* resource/aws_cloudfront_distribution: Preview panics during resource import with empty `forwarded_values.query_string` ([#14844](https://github.com/hashicorp/terraform-provider-aws/issues/14844))
* resource/aws_elasticache_replication_group: Ensure `tags` are stored in Terraform state and properly updated ([#12974](https://github.com/hashicorp/terraform-provider-aws/issues/12974))
* resource/aws_emr_instance_group: Increase creation and update timeout to 30 minutes ([#13077](https://github.com/hashicorp/terraform-provider-aws/issues/13077)] / [[#14106](https://github.com/hashicorp/terraform-provider-aws/issues/14106))
* resource/aws_globalaccelerator_accelerator: Increase creation timeout to 10 minutes ([#14486](https://github.com/hashicorp/terraform-provider-aws/issues/14486))
* resource/aws_globalaccelerator_endpoint_group: Prevent differences with `health_check_path` defaults ([#14486](https://github.com/hashicorp/terraform-provider-aws/issues/14486))
* resource/aws_glue_crawler: Properly update `schedule` value ([#14792](https://github.com/hashicorp/terraform-provider-aws/issues/14792))

## 3.3.0 (August 20, 2020)

ENHANCEMENTS

* data-source/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/hashicorp/terraform-provider-aws/issues/14663))
* provider: Support for appending information to User-Agent request headers with the `TF_APPEND_USER_AGENT` environment variable ([#14555](https://github.com/hashicorp/terraform-provider-aws/issues/14555))
* resource/aws_apigatewayv2_api: Add `body` argument ([#12567](https://github.com/hashicorp/terraform-provider-aws/issues/12567))
* resource/aws_customer_gateway: Support tag on create ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_dms_replication_instance: Add `allow_major_version_upgrade` argument ([#14550](https://github.com/hashicorp/terraform-provider-aws/issues/14550))
* resource/aws_ec2_client_vpn_network_association: Allow specifying custom security groups ([#14146](https://github.com/hashicorp/terraform-provider-aws/issues/14146))
* resource/aws_ec2_client_vpn_network_association: Support resource import ([#14146](https://github.com/hashicorp/terraform-provider-aws/issues/14146))
* resource/aws_egress_only_intrenet_gateway:-Ssupport tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_eks_node_group: Support `AL2_ARM_64` value for `ami_type` argument plan-time validation ([#14729](https://github.com/hashicorp/terraform-provider-aws/issues/14729))
* resource/aws_eks_node_group: Add `launch_template` configuration block ([#14639](https://github.com/hashicorp/terraform-provider-aws/issues/14639))
* resource/aws_internet_gateway: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_lambda_function: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/hashicorp/terraform-provider-aws/issues/14663))
* resource/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `compatible_runtimes` argument plan-time validation ([#14663](https://github.com/hashicorp/terraform-provider-aws/issues/14663))
* resource/aws_launch_template: Support `elastic-gpu` and `spot-instances-request` in `tag_specifications` `resource_type` argument plan-time validation ([#14662](https://github.com/hashicorp/terraform-provider-aws/issues/14662))
* resource/aws_network_acl: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_network_interface: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_route_table: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_security_group: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_spot_instance_request: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_storagegatway_smb_file_share: Add `audit_destination_arn` and `smb_acl_enabled` arguments ([#13572](https://github.com/hashicorp/terraform-provider-aws/issues/13572))
* resource/aws_subnet: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_subnet: Add plan-time validation to `ipv6_cidr_block` argument ([#12303](https://github.com/hashicorp/terraform-provider-aws/issues/12303))
* resource/aws_vpc_dhcp_options: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_vpc_peering_connection: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_vpn_connection: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_vpn_gateway: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_wafv2_rule_group: Add `forwarded_ip_config` configuration block to `geo_match_statement` ([#14685](https://github.com/hashicorp/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Add `forwarded_ip_config` configuration block to `rate_based_statement` and `geo_match_statement` ([#14685](https://github.com/hashicorp/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Support `FORWARDED_IP` value for `rate_based_statement` `aggregate_key_type` argument plan-time validation ([#14685](https://github.com/hashicorp/terraform-provider-aws/issues/14685))

BUG FIXES

* resource/aws_api_gateway_vpc_link: Increase create, update, and delete timeouts to 20 minutes ([#10407](https://github.com/hashicorp/terraform-provider-aws/issues/10407))
* resource/aws_apigatewayv2_stage: Set `execution_arn` attribute for HTTP APIs ([#14638](https://github.com/hashicorp/terraform-provider-aws/issues/14638))
* resource/aws_db_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/hashicorp/terraform-provider-aws/issues/12112))
* resource/aws_user_pool_domain: Ensure state removal when deleted outside Terraform ([#14732](https://github.com/hashicorp/terraform-provider-aws/issues/14732))
* resource/aws_rds_cluster_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/hashicorp/terraform-provider-aws/issues/12112))
* resource/aws_ssm_parameter: Handle retries after creation for asynchronous `data_type` validation process ([#14514](https://github.com/hashicorp/terraform-provider-aws/issues/14514))
* resource/aws_storagegateway_nfs_file_share: Skip `UpdateSMBFileShare` API call when only `tags` change and remove extraneous `ListTagsForResource` API call during read ([#13590](https://github.com/hashicorp/terraform-provider-aws/issues/13590))
* resource/aws_subnet: Ensure `ipv6_cidr_block` argument performs removal when removed from configuration ([#12303](https://github.com/hashicorp/terraform-provider-aws/issues/12303))

## 3.2.0 (August 14, 2020)

ENHANCEMENTS

* data-source/aws_launch_configuration: Add `ebs_block_device` `no_device` attribute ([#14583](https://github.com/hashicorp/terraform-provider-aws/issues/14583))
* data-source/aws_lb: Add `subnet_mapping` `private_ipv4_address` attribute ([#14545](https://github.com/hashicorp/terraform-provider-aws/issues/14545))
* provider: Upgrade to Terraform Plugin SDK V2. There should be no breaking changes from a practitioner's perspective. Some validation errors should now feature enhanced messaging. ([#14432](https://github.com/hashicorp/terraform-provider-aws/issues/14432))
* resource/aws_accessanalyzer_analyzer: Support `ORGANIZATION` value in `type` argument ([#14493](https://github.com/hashicorp/terraform-provider-aws/issues/14493))
* resource/aws_codebuild_project: Support `WINDOWS_SERVER_2019_CONTAINER` value in `environment` `type` argument plan-time validation ([#14532](https://github.com/hashicorp/terraform-provider-aws/issues/14532))
* resource/aws_organizations_organization: Support `AISERVICES_OPT_OUT_POLICY` value in `enabled_policy_types` argument plan-time validation (Support AI Opt Out policies) ([#14650](https://github.com/hashicorp/terraform-provider-aws/issues/14650))
* resource/aws_organizations_policy: Support `AISERVICES_OPT_OUT_POLICY` value in `type` argument plan-time validation (Support AI Opt Out policies) ([#14528](https://github.com/hashicorp/terraform-provider-aws/issues/14528))
* resource/aws_route53_health_check: Add `disabled` argument ([#14614](https://github.com/hashicorp/terraform-provider-aws/issues/14614))

BUG FIXES

* data-source/aws_launch_template: Prevent type error with `network_interfaces` `delete_on_termination` attribute ([#14599](https://github.com/hashicorp/terraform-provider-aws/issues/14599))
* resource/aws_acm_certificate_validation: Prevent panic with missing `DomainValidationOptions` `ResourceRecord` attribute in API response [[#14590](https://github.com/hashicorp/terraform-provider-aws/issues/14590)] 
* resource/aws_ecr_repository: Prevent panic with missing `EncryptionConfiguration` attribute in API response ([#14584](https://github.com/hashicorp/terraform-provider-aws/issues/14584))
* resource/aws_wafv2_rule_group: Prevent unnecessary resource recreation with `rule` updates ([#14617](https://github.com/hashicorp/terraform-provider-aws/issues/14617))
* resource/aws_wafv2_web_acl: Prevent unnecessary resource recreation with `rule` updates ([#14616](https://github.com/hashicorp/terraform-provider-aws/issues/14616))

## 3.1.0 (August 07, 2020)

NOTES:

* resource/aws_route53_zone_association: The addition of cross-account zone association support required the use of new `ListHostedZonesByVPC` API call and adding the VPC Region to the resource ID for new resources. Restrictive IAM permissions for Terraform and cross-region imports may require updates. ([#14215](https://github.com/hashicorp/terraform-provider-aws/issues/14215))

FEATURES

* **New Data Source:** `aws_ec2_spot_price` ([#12504](https://github.com/hashicorp/terraform-provider-aws/issues/12504))
* **New Resource**: `aws_route53_vpc_association_authorization` ([#14215](https://github.com/hashicorp/terraform-provider-aws/issues/14215))

ENHANCEMENTS

* data-source/aws_ecr_repository: Allow `registry_id` as an argument ([#14368](https://github.com/hashicorp/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `image_scanning_configuration` and `image_tag_mutability` attributes ([#14368](https://github.com/hashicorp/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/hashicorp/terraform-provider-aws/issues/14520))
* resource/aws_api_gateway_method_settings: Plan-time validation added to `settings` `unauthorized_cache_control_header_strategy` and `logging_level` arguments ([#12651](https://github.com/hashicorp/terraform-provider-aws/issues/12651))
* resource/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/hashicorp/terraform-provider-aws/issues/14520))
* resource/aws_lb: Add `subnet_mapping` configuration block `private_ipv4_address` argument ([#11404](https://github.com/hashicorp/terraform-provider-aws/issues/11404))
* resource/aws_rds_global_cluster: Add `force_destroy` and `source_db_cluster_identifier` arguments ([#14487](https://github.com/hashicorp/terraform-provider-aws/issues/14487))
* resource/aws_rds_global_cluster: Add `global_cluster_members` attribute ([#14487](https://github.com/hashicorp/terraform-provider-aws/issues/14487))
* resource/aws_route53_zone_association: Cross-account zone associations can now be created in conjunction with the new `aws_route53_vpc_association_authorization` resource ([#14215](https://github.com/hashicorp/terraform-provider-aws/issues/14215))
* resource/aws_ssm_parameter: Add `data_type` argument (support `aws:ec2:image` parameters) ([#13326](https://github.com/hashicorp/terraform-provider-aws/issues/13326))

BUG FIXES

* data-source/aws_availability_zones: Prevent unexpected plan output every apply with `group_names` attribute ([#14412](https://github.com/hashicorp/terraform-provider-aws/issues/14412))
* data-source/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/hashicorp/terraform-provider-aws/issues/14481))
* resource/aws_api_gateway_method_settings: Allow `settings` `cache_ttl_in_seconds` argument to be set to 0 ([#12651](https://github.com/hashicorp/terraform-provider-aws/issues/12651))
* resource/aws_elastictranscoder_preset: Prevent empty configuration block panics ([#14092](https://github.com/hashicorp/terraform-provider-aws/issues/14092))
* resource/aws_lambda_event_source_mapping: Allow `maximum_retry_attempts` argument to be set to 0 ([#12479](https://github.com/hashicorp/terraform-provider-aws/issues/12479))
* resource/aws_rds_cluster: Add an `InvalidDBClusterStateFault` retryable error condition for clusters part of a global cluster ([#14420](https://github.com/hashicorp/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Increase retry timeout for deletion to 2 minutes ([#14420](https://github.com/hashicorp/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Prevent error when both `global_cluster_identifier` and `replication_source_identifier` are configured on creation ([#14490](https://github.com/hashicorp/terraform-provider-aws/issues/14490))
* resource/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/hashicorp/terraform-provider-aws/issues/14481))
* resource/aws_secretsmanager_secret: Allow retries for IAM eventual consistency errors ([#14459](https://github.com/hashicorp/terraform-provider-aws/issues/14459))
* resource/aws_security_group: Ensure `name_prefix` argument with hex digits `a` through `f` is properly imported ([#14475](https://github.com/hashicorp/terraform-provider-aws/issues/14475))
* resource/aws_spot_fleet_request: Allow `target_capacity` argument to be updated to 0 ([#12759](https://github.com/hashicorp/terraform-provider-aws/issues/12759))
* resource/aws_spot_fleet_request: Wait for modify operation completion (default timeout of 10 minutes) ([#12759](https://github.com/hashicorp/terraform-provider-aws/issues/12759))
* resource/aws_vpc_dhcp_options_association: Properly trigger resource recreation when VPC is deleted outside Terraform ([#14367](https://github.com/hashicorp/terraform-provider-aws/issues/14367))

## 3.0.0 (July 31, 2020)

NOTES:
* provider: This version is built using Go 1.14.5, including security fixes to the crypto/x509 and net/http packages.

BREAKING CHANGES

* provider: New versions of the provider can only be automatically installed on Terraform 0.12 and later ([#14143](https://github.com/hashicorp/terraform-provider-aws/issues/14143))
* provider: All "removed" attributes are cut, using them would result in a Terraform Core level error ([#14001](https://github.com/hashicorp/terraform-provider-aws/issues/14001))
* provider: Credential ordering has changed from static, environment, shared credentials, EC2 metadata, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) to static, environment, shared credentials, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: The `AWS_METADATA_TIMEOUT` environment variable no longer has any effect as we now depend on the default AWS Go SDK EC2 Metadata client timeout of one second with two retries ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Remove deprecated `kinesis_analytics` and `r53` custom service endpoint arguments ([#14238](https://github.com/hashicorp/terraform-provider-aws/issues/14238))
* data-source/aws_availability_zones: Remove deprecated `blacklisted_names` and `blacklisted_zone_ids` arguments ([#14134](https://github.com/hashicorp/terraform-provider-aws/issues/14134))
* data-source/aws_directory_service_directory: Return an error when a single result is not found ([#14006](https://github.com/hashicorp/terraform-provider-aws/issues/14006))
* data-source/aws_ecr_repository: Return an error when a single result is not found ([#10520](https://github.com/hashicorp/terraform-provider-aws/issues/10520))
* data-source/aws_efs_file_system: Return an error when a single result is not found ([#14005](https://github.com/hashicorp/terraform-provider-aws/issues/14005))
* data-source/aws_launch_template: Return an error when a single result is not found ([#10521](https://github.com/hashicorp/terraform-provider-aws/issues/10521))
* data-source/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in data-source ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* data-source/aws_route53_zone: Trailing period removed from `name` argument set in data-source ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_acm_certificate: `certificate_body`, `certificate_chain`, and `private_key` attributes are no longer stored in the Terraform state with hash values ([#9685](https://github.com/hashicorp/terraform-provider-aws/issues/9685))
* resource/aws_acm_certificate: `domain_validation_options` attribute changed from list to set ([#14199](https://github.com/hashicorp/terraform-provider-aws/issues/14199))
* resource/aws_acm_certificate: Plan-time validation added to `domain_name` and `subject_alternative_names` arguments to prevent usage of strings with trailing periods ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_api_gateway_method_settings: Remove `Computed` property from `throttling_burst_limit` and `throttling_rate_limit` arguments, enabling drift detection ([#14266](https://github.com/hashicorp/terraform-provider-aws/issues/14266))
* resource/aws_api_gateway_method_settings: Update `throttling_burst_limit` and `throttling_rate_limit` argument defaults to match API default of `-1` to keep throttling disabled ([#14266](https://github.com/hashicorp/terraform-provider-aws/issues/14266))
* resource/aws_autoscaling_group: `availability_zones` and `vpc_zone_identifier` argument conflict now reported at plan-time ([#12927](https://github.com/hashicorp/terraform-provider-aws/issues/12927))
* resource/aws_autoscaling_group: Remove `Computed` property from `load_balancers` and `target_group_arns` arguments, enabling drift detection ([#14064](https://github.com/hashicorp/terraform-provider-aws/issues/14064))
* resource/aws_cloudfront_distribution: `active_trusted_signers` argument renamed to `trusted_signers` to support accessing `items` in Terraform 0.12 ([#14339](https://github.com/hashicorp/terraform-provider-aws/issues/14339))
* resource/aws_cloudwatch_log_group: Automatically trim `:*` suffix from `arn` attribute ([#14214](https://github.com/hashicorp/terraform-provider-aws/issues/14214))
* resource/aws_codepipeline: Removes `GITHUB_TOKEN` environment variable ([#14175](https://github.com/hashicorp/terraform-provider-aws/issues/14175))
* resource/aws_cognito_user_pool: Remove deprecated `admin_create_user_config` configuration block `unused_account_validity_days` argument ([#14294](https://github.com/hashicorp/terraform-provider-aws/issues/14294))
* resource/aws_dx_gateway: Remove automatic `aws_dx_gateway_association` resource import ([#14124](https://github.com/hashicorp/terraform-provider-aws/issues/14124))
* resource/aws_dx_gateway_association: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/hashicorp/terraform-provider-aws/issues/14144))
* resource/aws_dx_gateway_association_proposal: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/hashicorp/terraform-provider-aws/issues/14144))
* resource/aws_ebs_volume: Return an error when `iops` argument set to a value greater than 0 for volume types other than `io1` ([#14310](https://github.com/hashicorp/terraform-provider-aws/issues/14310))
* resource/aws_elastic_transcoder_preset: Remove `video` configuration block `max_frame_rate` argument default value ([#7141](https://github.com/hashicorp/terraform-provider-aws/issues/7141))
* resource/aws_emr_cluster: Remove deprecated `instance_group` configuration block, `core_instance_count`, `core_instance_type`, and `master_instance_type` arguments ([#14137](https://github.com/hashicorp/terraform-provider-aws/issues/14137))
* resource/aws_glue_job: Remove deprecated `allocated_capacity` argument ([#14296](https://github.com/hashicorp/terraform-provider-aws/issues/14296))
* resource/aws_iam_access_key: Remove deprecated `ses_smtp_password` attribute ([#14299](https://github.com/hashicorp/terraform-provider-aws/issues/14299))
* resource/aws_iam_instance_profile: Remove deprecated `roles` argument ([#14303](https://github.com/hashicorp/terraform-provider-aws/issues/14303))
* resource/aws_iam_server_certificate: Remove state hashing from `certificate_body`, `certificate_chain`, and `private_key` arguments for new or recreated resources ([#14187](https://github.com/hashicorp/terraform-provider-aws/issues/14187))
* resource/aws_instance: Return an error when `ebs_block_device` `iops` or `root_block_device` `iops` argument set to a value greater than `0` for volume types other than `io1` ([#14310](https://github.com/hashicorp/terraform-provider-aws/issues/14310))
* resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN ([#12876](https://github.com/hashicorp/terraform-provider-aws/issues/12876))
* resource/aws_launch_template: `network_interfaces` `delete_on_termination` argument changed from `bool` to `string` type ([#8612](https://github.com/hashicorp/terraform-provider-aws/issues/8612))
* resource/aws_lb_listener_rule: Remove deprecated `condition` configuration block `field` and `values` arguments ([#14309](https://github.com/hashicorp/terraform-provider-aws/issues/14309))
* resource/aws_msk_cluster: Update `encryption_info` `encryption_in_transit` `client_broker` argument default to match API default of `TLS` ([#14132](https://github.com/hashicorp/terraform-provider-aws/issues/14132))
* resource/aws_rds_cluster: Update `scaling_configuration` `min_capacity` argument default to match API default of `1` ([#14268](https://github.com/hashicorp/terraform-provider-aws/issues/14268))
* resource/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in resource ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_route53_zone: Trailing period removed from `name` argument set in resource ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_s3_bucket: Remove automatic `aws_s3_bucket_policy` resource import ([#14121](https://github.com/hashicorp/terraform-provider-aws/issues/14121))
* resource/aws_s3_bucket: Convert `region` to read-only attribute ([#14127](https://github.com/hashicorp/terraform-provider-aws/issues/14127))
* resource/aws_s3_bucket_metric: Update `filter` argument to require at least one of the `prefix` or `tags` nested arguments ([#14230](https://github.com/hashicorp/terraform-provider-aws/issues/14230))
* resource/aws_security_group: Remove automatic `aws_security_group_rule` resource import ([#12616](https://github.com/hashicorp/terraform-provider-aws/issues/12616))
* resource/aws_ses_domain_identity: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_ses_domain_identity_verification: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_sns_platform_application: `platform_credential` and `platform_principal` attributes are no longer stored in the Terraform state with hash values ([#3894](https://github.com/hashicorp/terraform-provider-aws/issues/3894))
* resource/aws_spot_fleet_request: Remove 24 hour default for `valid_until` argument ([#9718](https://github.com/hashicorp/terraform-provider-aws/issues/9718))
* resource/aws_ssm_maintenance_window_task: Remove deprecated `logging_info` and `task_parameters` configuration blocks ([#14311](https://github.com/hashicorp/terraform-provider-aws/issues/14311))

FEATURES

* **New Data Source:** aws_workspaces_directory ([#13529](https://github.com/hashicorp/terraform-provider-aws/issues/13529))

ENHANCEMENTS

* provider: Always enable shared configuration file support (no longer require `AWS_SDK_LOAD_CONFIG` environment variable) ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Add `assume_role` configuration block `duration_seconds`, `policy_arns`, `tags`, and `transitive_tag_keys` arguments ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* data-source/aws_instance: Add `secondary_private_ips` attribute ([#14079](https://github.com/hashicorp/terraform-provider-aws/issues/14079))
* data-source/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/hashicorp/terraform-provider-aws/issues/14221))
* resource/aws_acm_certificate: Enable `domain_validation_options` usage in downstream resource `count` and `for_each` references ([#14199](https://github.com/hashicorp/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Add plan-time validation to `authorizer_credentials` argument ([#12643](https://github.com/hashicorp/terraform-provider-aws/issues/12643))
* resource/aws_api_gateway_method_settings: Add import support ([#14266](https://github.com/hashicorp/terraform-provider-aws/issues/14266))
* resource/aws_apigatewayv2_integration: Add `request_parameters` attribute ([#14080](https://github.com/hashicorp/terraform-provider-aws/issues/14080))
* resource/aws_apigatewayv2_integration: Add `tls_config` attribute ([#13013](https://github.com/hashicorp/terraform-provider-aws/issues/13013))
* resource/aws_apigatewayv2_route: Support for updating route key ([#13833](https://github.com/hashicorp/terraform-provider-aws/issues/13833))
* resource/aws_apigatewayv2_stage: Make `deployment_id` a `Computed` attribute ([#13644](https://github.com/hashicorp/terraform-provider-aws/issues/13644))
* resource/aws_fsx_lustre_file_system: Add `deployment_type` and `per_unit_storage_throughput` attributes ([#13639](https://github.com/hashicorp/terraform-provider-aws/issues/13639))
* resource_aws_fsx_windows_file_system - add `storage_type` argument. ([#14316](https://github.com/hashicorp/terraform-provider-aws/issues/14316))
* resource_aws_fsx_windows_file_system: add support for multi-az ([#12676](https://github.com/hashicorp/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: add `SINGLE_AZ_2` deployment type ([#12676](https://github.com/hashicorp/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: adds `preferred_file_server_ip`, `remote_administration_endpoint` attributes ([#12676](https://github.com/hashicorp/terraform-provider-aws/issues/12676))
* resource/aws_instance: Add `secondary_private_ips` argument (conflicts with `network_interface` configuration block) ([#14079](https://github.com/hashicorp/terraform-provider-aws/issues/14079))

BUG FIXES

* provider: Ensure nil is not passed to RetryError helpers, may result in some bug fixes ([#14104](https://github.com/hashicorp/terraform-provider-aws/issues/14104))
* provider: Ensure configured STS endpoint is used during `AssumeRole` API calls ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Prefer AWS shared configuration over EC2 metadata credentials by default ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Prefer CodeBuild, ECS, EKS credentials over EC2 metadata credentials by default ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* data-source/aws_lb: `enable_http2` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_acm_certificate: Prevent unexpected ordering differences with `domain_validation_options` attribute ([#14199](https://github.com/hashicorp/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Allow `authorizer_result_ttl_in_seconds` to be set to 0 ([#12643](https://github.com/hashicorp/terraform-provider-aws/issues/12643))
* resource/aws_apigatewayv2_integration: Correctly handle the `integration_method` attribute for AWS Lambda integrations([#13266](https://github.com/hashicorp/terraform-provider-aws/issues/13266))
* resource/aws_apigatewayv2_integration: Correctly handle the `passthrough_behavior` attribute for HTTP APIs ([#13062](https://github.com/hashicorp/terraform-provider-aws/issues/13062))
* resource/aws_apigatewayv2_stage: Correctly handle `default_route_setting` and `route_setting` `data_trace_enabled` and `logging_level` for HTTP APIs. `logging_level` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. ([#13809](https://github.com/hashicorp/terraform-provider-aws/issues/13809))
* resource/aws_appautoscaling_target: Only retry `DeregisterScalableTarget` retries on all errors on deletion ([#14259](https://github.com/hashicorp/terraform-provider-aws/issues/14259))
* resource/aws_dx_gateway_association: Increase default create/update/delete timeouts to 30 minutes ([#14144](https://github.com/hashicorp/terraform-provider-aws/issues/14144))
* resource/aws_codepipeline: Only retry `CreatePipeline` errors for IAM eventual consistency errors ([#14264](https://github.com/hashicorp/terraform-provider-aws/issues/14264))
* resource/aws_elasticsearch_domain: Update method to properly set `advanced_security_options` ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_lambda_function: Increase IAM retry timeout for creation to standard 2 minute timeout ([#14291](https://github.com/hashicorp/terraform-provider-aws/issues/14291))
* resource/aws_lb_cookie_stickiness_policy: `lb_port` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_network_acl_rule: Immediately return `DescribeNetworkAcls` errors on creation ([#14261](https://github.com/hashicorp/terraform-provider-aws/issues/14261))
* resource/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/hashicorp/terraform-provider-aws/issues/14221))
* resource/aws_sns_topic_subscription: Immediately return `ListSubscriptionsByTopic` errors ([#14262](https://github.com/hashicorp/terraform-provider-aws/issues/14262))
* resource/aws_spot_fleet_request: Only retry `RequestSpotFleet` on IAM eventual consistency errors and use standard 2 minute timeout ([#14265](https://github.com/hashicorp/terraform-provider-aws/issues/14265))
* resource/aws_spot_instance_request: `primary_network_interface_id` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_ssm_activation: Only retry `CreateActivation` on IAM eventual consistency errors and use standard 2 minute timeout ([#14263](https://github.com/hashicorp/terraform-provider-aws/issues/14263))
* resource/aws_ssm_association: `parameters` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))

## Previous Releases

For information on prior major releases, see their changelogs:

* [2.x and earlier](https://github.com/hashicorp/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
