## 3.7.0 (Unreleased)
## 3.6.0 (September 11, 2020)

FEATURES

* **New Resource:** `aws_db_proxy_default_target_group` ([#12743](https://github.com/terraform-providers/terraform-provider-aws/issues/12743))

BUG FIXES

* resource/aws_ec2_client_vpn_authorization_rule: Increase active and revoked timeouts from 1 to 5 minutes ([#15037](https://github.com/terraform-providers/terraform-provider-aws/issues/15037))

## 3.5.0 (September 03, 2020)

FEATURES

* **New Data Source:** `aws_docdb_orderable_db_instance` ([#14931](https://github.com/terraform-providers/terraform-provider-aws/issues/14931))
* **New Data Source:** `aws_lex_slot_type` ([#8916](https://github.com/terraform-providers/terraform-provider-aws/issues/8916))
* **New Data Source:** `aws_neptune_orderable_db_instance` ([#14953](https://github.com/terraform-providers/terraform-provider-aws/issues/14953))
* **New Data Source:** `aws_rds_orderable_db_instance` ([#14834](https://github.com/terraform-providers/terraform-provider-aws/issues/14834))
* **New Data Source:** `aws_vpc_peering_connections` ([#9491](https://github.com/terraform-providers/terraform-provider-aws/issues/9491))
* **New Resource:** `aws_codebuild_report_group` ([#12573](https://github.com/terraform-providers/terraform-provider-aws/issues/12573))
* **New Resource:** `aws_db_proxy` ([#12704](https://github.com/terraform-providers/terraform-provider-aws/issues/12704))
* **New Resource:** `aws_emr_instance_fleet` ([#14813](https://github.com/terraform-providers/terraform-provider-aws/issues/14813))
* **New Resource:** `aws_glue_user_defined_function` ([#12537](https://github.com/terraform-providers/terraform-provider-aws/issues/12537))
* **New Resource:** `aws_guardduty_filter` ([#14876](https://github.com/terraform-providers/terraform-provider-aws/issues/14876))
* **New Resource:** `aws_lex_slot_type` ([#8916](https://github.com/terraform-providers/terraform-provider-aws/issues/8916))

ENHANCEMENTS

* data-source/aws_cur_report_definition: Add `refresh_closed_reports` and `report_versioning` attributes ([#12428](https://github.com/terraform-providers/terraform-provider-aws/issues/12428))
* data-source/aws_outposts_outpost: Add `arn` argument ([#14967](https://github.com/terraform-providers/terraform-provider-aws/issues/14967))
* data-source/aws_route: Add `local_gateway_id` attribute ([#14864](https://github.com/terraform-providers/terraform-provider-aws/issues/14864))
* data-source/aws_route_table: Add `route` `local_gateway_id` attribute ([#14864](https://github.com/terraform-providers/terraform-provider-aws/issues/14864))
* resource/aws_acm_certificate: Provide additional plan-time validation for `subject_alternative_names` argument values ([#14782](https://github.com/terraform-providers/terraform-provider-aws/issues/14782))
* resource/aws_ami: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/terraform-providers/terraform-provider-aws/issues/14906))
* resource/aws_autoscaling_group: Support provider-level `ignore_tags` configuration ([#13868](https://github.com/terraform-providers/terraform-provider-aws/issues/13868))
* resource/aws_cloudtrail: Add `insight_selector` configuration block ([#12390](https://github.com/terraform-providers/terraform-provider-aws/issues/12390))
* resource/aws_cur_report_definition: Add `refresh_closed_reports` and `report_versioning` arguments ([#12428](https://github.com/terraform-providers/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `ATHENA` value in `additional_artifacts` argument plan-time validation ([#12428](https://github.com/terraform-providers/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `Parquet` value in `compression` and `format` argument plan-time validations ([#12428](https://github.com/terraform-providers/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `MONTHLY` value in `time_unit` argument plan-time validation ([#12428](https://github.com/terraform-providers/terraform-provider-aws/issues/12428))
* resource/aws_ebs_volume: Support io2 type ([#14894](https://github.com/terraform-providers/terraform-provider-aws/issues/14894))
* resource/aws_ec2_client_vpn_endpoint: Support `authentication_options` `type` argument `federated-authentication` value and new `saml_provider_arn` argument ([#14171](https://github.com/terraform-providers/terraform-provider-aws/issues/14171))
* resource/aws_emr_cluster: Add `core_instance_fleet` and `master_instance_fleet` configuration blocks ([#14788](https://github.com/terraform-providers/terraform-provider-aws/issues/14788))
* resource/aws_instance: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/terraform-providers/terraform-provider-aws/issues/14906))
* resource/aws_kinesis_firehose_delivery_stream: Add `elasticsearch_configuration` `vpc_config` configuration block ([#13269](https://github.com/terraform-providers/terraform-provider-aws/issues/13269))
* resource/aws_kinesis_firehose_delivery_stream: Add `elasticsearch_configuration` `cluster_endpoint` argument ([#12484](https://github.com/terraform-providers/terraform-provider-aws/issues/12484))
* resource/aws_kinesis_firehose_delivery_stream: Add various plan-time validations for arguments ([#12484](https://github.com/terraform-providers/terraform-provider-aws/issues/12484))
* resource/aws_launch_template: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/terraform-providers/terraform-provider-aws/issues/14906))
* resource/aws_msk_configuration: Support resource in-place updates and deletion ([#14826](https://github.com/terraform-providers/terraform-provider-aws/issues/14826))
* resource/aws_route: Add `local_gateway_id` argument ([#14864](https://github.com/terraform-providers/terraform-provider-aws/issues/14864))
* resource/aws_route_table: Add `route` `local_gateway_id` argument ([#14864](https://github.com/terraform-providers/terraform-provider-aws/issues/14864))
* resource/aws_spot_fleet_request: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/terraform-providers/terraform-provider-aws/issues/14906))
* resource/aws_wafv2_rule_group: Add `ip_set_forwarded_ip_config` configuration block to `ip_set_reference_statement` ([#14902](https://github.com/terraform-providers/terraform-provider-aws/issues/14902))
* resource/aws_wafv2_web_acl: Add `ip_set_forwarded_ip_config` configuration block to `ip_set_reference_statement` ([#14902](https://github.com/terraform-providers/terraform-provider-aws/issues/14902))

BUG FIXES

* resource/aws_autoscaling_group: Prevent unnecessary tag removal and recreation within tag updates ([#13868](https://github.com/terraform-providers/terraform-provider-aws/issues/13868))
* resource/aws_cloudfront_distribution: Prevent panic with missing `ForwardedValues` ([#14993](https://github.com/terraform-providers/terraform-provider-aws/issues/14993))
* resource/aws_dynamodb_table: Properly update `global_secondary_index` `non_key_attributes` values ([#9988](https://github.com/terraform-providers/terraform-provider-aws/issues/9988))
* resource/aws_emr_cluster: Prevent recreation when `ebs_config.volumes_per_instance` is greater than 1 ([#14858](https://github.com/terraform-providers/terraform-provider-aws/issues/14858))
* resource/aws_lambda_function_event_invoke_config: Prevent unexpected format of function resource error ([#14851](https://github.com/terraform-providers/terraform-provider-aws/issues/14851))
* resource/aws_lightsail_instance: Prevent panic with key-only tags ([#13868](https://github.com/terraform-providers/terraform-provider-aws/issues/13868))
* resource/aws_mq_configuration: Prevent additional revision creation with `tags` only updates ([#14850](https://github.com/terraform-providers/terraform-provider-aws/issues/14850))
* resource/aws_opsworks_stack: Suppress equivalent `custom_json` differences ([#14886](https://github.com/terraform-providers/terraform-provider-aws/issues/14886))
* resource/aws_rds_cluster_endpoint: Increase creation timeout to 30 minutes ([#14862](https://github.com/terraform-providers/terraform-provider-aws/issues/14862))
* resource/aws_route53_resolver_rule: Correct handling for single period (`.`) value in `domain_name` argument ([#15015](https://github.com/terraform-providers/terraform-provider-aws/issues/15015))
* resource/aws_route53_zone_association: Correctly handle zones with over 100 VPC associations ([#14885](https://github.com/terraform-providers/terraform-provider-aws/issues/14885))
* resource/aws_waf_rate_based_rule: Properly update `rate_limit` value ([#14964](https://github.com/terraform-providers/terraform-provider-aws/issues/14964))
* resource/aws_workspaces_workspace: Prevent error when `workspace_properties` `running_mode` is set to `ALWAYS_ON` ([#13976](https://github.com/terraform-providers/terraform-provider-aws/issues/13976))

## 3.4.0 (August 27, 2020)

FEATURES

* **New Data Source:** `aws_db_subnet_group` ([#9525](https://github.com/terraform-providers/terraform-provider-aws/issues/9525))
* **New Resource:** `aws_emr_managed_scaling_policy` ([#13965](https://github.com/terraform-providers/terraform-provider-aws/issues/13965))
* **New Resource:** `aws_guardduty_publishing_destination` ([#13894](https://github.com/terraform-providers/terraform-provider-aws/issues/13894))
* **New Resource:** `aws_securityhub_action_target` ([#10493](https://github.com/terraform-providers/terraform-provider-aws/issues/10493))
* **New Resource:** `aws_xray_encryption_config` ([#13600](https://github.com/terraform-providers/terraform-provider-aws/issues/13600))
* **New Resource:** `aws_xray_group` ([#13597](https://github.com/terraform-providers/terraform-provider-aws/issues/13597))

ENHANCEMENTS

* resource/aws_apigatewayv2_integration: Add `integration_subtype` argument (Support AWS service integrations for HTTP APIs) ([#14860](https://github.com/terraform-providers/terraform-provider-aws/issues/14860))
* resource/aws_elasticache_replication_group: Add plan-time validation for `notification_topic_arn` and `snapshot_arns` arguments ([#12974](https://github.com/terraform-providers/terraform-provider-aws/issues/12974))
* resource/aws_globalaccelerator_endpoint_group: Add `client_ip_preservation_enabled` argument to the `endpoint_configuration` configuration block ([#14486](https://github.com/terraform-providers/terraform-provider-aws/issues/14486))
* resource/aws_storagegateway_cached_iscsi_volume: Add `kms_encrypted` and `kms_key` arguments ([#12066](https://github.com/terraform-providers/terraform-provider-aws/issues/12066))
* resource/aws_storagegateway_gateway: Add `smb_security_strategy` argument ([#13563](https://github.com/terraform-providers/terraform-provider-aws/issues/13563))
* resource/aws_storagegateway_gateway: Add plan-time validation for `gateway_ip_address` argument ([#13563](https://github.com/terraform-providers/terraform-provider-aws/issues/13563))
* resource/aws_storagegateway_gateway: Add `average_download_rate_limit_in_bits_per_sec` and `average_upload_rate_limit_in_bits_per_sec` arguments ([#13568](https://github.com/terraform-providers/terraform-provider-aws/issues/13568))
* resource/aws_storagegateway_nfs_file_share: Add `cache_attributes` configuration block ([#14759](https://github.com/terraform-providers/terraform-provider-aws/issues/14759))
* resource/aws_storagegateway_nfs_file_share: Support `S3_INTELLIGENT_TIERING` value in `default_storage_class` argument plan-time validation ([#14759](https://github.com/terraform-providers/terraform-provider-aws/issues/14759))
* resource/aws_storagegateway_smb_file_share: Add `cache_attributes` configuration block and `case_sensitivity` argument ([#14790](https://github.com/terraform-providers/terraform-provider-aws/issues/14790))
* resource/aws_storagegateway_smb_file_share: Support `S3_INTELLIGENT_TIERING` value in `default_storage_class` argument plan-time validation ([#14790](https://github.com/terraform-providers/terraform-provider-aws/issues/14790))
* resource/aws_xray_sampling_rule: Add `tags` argument ([#14831](https://github.com/terraform-providers/terraform-provider-aws/issues/14831))

BUG FIXES

* resource/aws_acmpca_certificate_authority: Ensure `DELETED` status triggers state removal ([#13684](https://github.com/terraform-providers/terraform-provider-aws/issues/13684))
* resource/aws_appmesh_virtual_node: Prevent panics with empty `backend` configuration blocks ([#14074](https://github.com/terraform-providers/terraform-provider-aws/issues/14074))
* resource/aws_cloudfront_distribution: Preview panics during resource import with empty `forwarded_values.query_string` ([#14844](https://github.com/terraform-providers/terraform-provider-aws/issues/14844))
* resource/aws_elasticache_replication_group: Ensure `tags` are stored in Terraform state and properly updated ([#12974](https://github.com/terraform-providers/terraform-provider-aws/issues/12974))
* resource/aws_emr_instance_group: Increase creation and update timeout to 30 minutes ([#13077](https://github.com/terraform-providers/terraform-provider-aws/issues/13077)] / [[#14106](https://github.com/terraform-providers/terraform-provider-aws/issues/14106))
* resource/aws_globalaccelerator_accelerator: Increase creation timeout to 10 minutes ([#14486](https://github.com/terraform-providers/terraform-provider-aws/issues/14486))
* resource/aws_globalaccelerator_endpoint_group: Prevent differences with `health_check_path` defaults ([#14486](https://github.com/terraform-providers/terraform-provider-aws/issues/14486))
* resource/aws_glue_crawler: Properly update `schedule` value ([#14792](https://github.com/terraform-providers/terraform-provider-aws/issues/14792))

## 3.3.0 (August 20, 2020)

ENHANCEMENTS

* data-source/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/terraform-providers/terraform-provider-aws/issues/14663))
* provider: Support for appending information to User-Agent request headers with the `TF_APPEND_USER_AGENT` environment variable ([#14555](https://github.com/terraform-providers/terraform-provider-aws/issues/14555))
* resource/aws_apigatewayv2_api: Add `body` argument ([#12567](https://github.com/terraform-providers/terraform-provider-aws/issues/12567))
* resource/aws_customer_gateway: Support tag on create ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_dms_replication_instance: Add `allow_major_version_upgrade` argument ([#14550](https://github.com/terraform-providers/terraform-provider-aws/issues/14550))
* resource/aws_ec2_client_vpn_network_association: Allow specifying custom security groups ([#14146](https://github.com/terraform-providers/terraform-provider-aws/issues/14146))
* resource/aws_ec2_client_vpn_network_association: Support resource import ([#14146](https://github.com/terraform-providers/terraform-provider-aws/issues/14146))
* resource/aws_egress_only_intrenet_gateway:-Ssupport tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_eks_node_group: Support `AL2_ARM_64` value for `ami_type` argument plan-time validation ([#14729](https://github.com/terraform-providers/terraform-provider-aws/issues/14729))
* resource/aws_eks_node_group: Add `launch_template` configuration block ([#14639](https://github.com/terraform-providers/terraform-provider-aws/issues/14639))
* resource/aws_internet_gateway: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_lambda_function: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/terraform-providers/terraform-provider-aws/issues/14663))
* resource/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `compatible_runtimes` argument plan-time validation ([#14663](https://github.com/terraform-providers/terraform-provider-aws/issues/14663))
* resource/aws_launch_template: Support `elastic-gpu` and `spot-instances-request` in `tag_specifications` `resource_type` argument plan-time validation ([#14662](https://github.com/terraform-providers/terraform-provider-aws/issues/14662))
* resource/aws_network_acl: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_network_interface: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_route_table: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_security_group: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_spot_instance_request: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_storagegatway_smb_file_share: Add `audit_destination_arn` and `smb_acl_enabled` arguments ([#13572](https://github.com/terraform-providers/terraform-provider-aws/issues/13572))
* resource/aws_subnet: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_subnet: Add plan-time validation to `ipv6_cidr_block` argument ([#12303](https://github.com/terraform-providers/terraform-provider-aws/issues/12303))
* resource/aws_vpc_dhcp_options: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_vpc_peering_connection: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_vpn_connection: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_vpn_gateway: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_wafv2_rule_group: Add `forwarded_ip_config` configuration block to `geo_match_statement` ([#14685](https://github.com/terraform-providers/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Add `forwarded_ip_config` configuration block to `rate_based_statement` and `geo_match_statement` ([#14685](https://github.com/terraform-providers/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Support `FORWARDED_IP` value for `rate_based_statement` `aggregate_key_type` argument plan-time validation ([#14685](https://github.com/terraform-providers/terraform-provider-aws/issues/14685))

BUG FIXES

* resource/aws_api_gateway_vpc_link: Increase create, update, and delete timeouts to 20 minutes ([#10407](https://github.com/terraform-providers/terraform-provider-aws/issues/10407))
* resource/aws_apigatewayv2_stage: Set `execution_arn` attribute for HTTP APIs ([#14638](https://github.com/terraform-providers/terraform-provider-aws/issues/14638))
* resource/aws_db_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/terraform-providers/terraform-provider-aws/issues/12112))
* resource/aws_user_pool_domain: Ensure state removal when deleted outside Terraform ([#14732](https://github.com/terraform-providers/terraform-provider-aws/issues/14732))
* resource/aws_rds_cluster_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/terraform-providers/terraform-provider-aws/issues/12112))
* resource/aws_ssm_parameter: Handle retries after creation for asynchronous `data_type` validation process ([#14514](https://github.com/terraform-providers/terraform-provider-aws/issues/14514))
* resource/aws_storagegateway_nfs_file_share: Skip `UpdateSMBFileShare` API call when only `tags` change and remove extraneous `ListTagsForResource` API call during read ([#13590](https://github.com/terraform-providers/terraform-provider-aws/issues/13590))
* resource/aws_subnet: Ensure `ipv6_cidr_block` argument performs removal when removed from configuration ([#12303](https://github.com/terraform-providers/terraform-provider-aws/issues/12303))

## 3.2.0 (August 14, 2020)

ENHANCEMENTS

* data-source/aws_launch_configuration: Add `ebs_block_device` `no_device` attribute ([#14583](https://github.com/terraform-providers/terraform-provider-aws/issues/14583))
* data-source/aws_lb: Add `subnet_mapping` `private_ipv4_address` attribute ([#14545](https://github.com/terraform-providers/terraform-provider-aws/issues/14545))
* provider: Upgrade to Terraform Plugin SDK V2. There should be no breaking changes from a practitioner's perspective. Some validation errors should now feature enhanced messaging. ([#14432](https://github.com/terraform-providers/terraform-provider-aws/issues/14432))
* resource/aws_accessanalyzer_analyzer: Support `ORGANIZATION` value in `type` argument ([#14493](https://github.com/terraform-providers/terraform-provider-aws/issues/14493))
* resource/aws_codebuild_project: Support `WINDOWS_SERVER_2019_CONTAINER` value in `environment` `type` argument plan-time validation ([#14532](https://github.com/terraform-providers/terraform-provider-aws/issues/14532))
* resource/aws_organizations_organization: Support `AISERVICES_OPT_OUT_POLICY` value in `enabled_policy_types` argument plan-time validation (Support AI Opt Out policies) ([#14650](https://github.com/terraform-providers/terraform-provider-aws/issues/14650))
* resource/aws_organizations_policy: Support `AISERVICES_OPT_OUT_POLICY` value in `type` argument plan-time validation (Support AI Opt Out policies) ([#14528](https://github.com/terraform-providers/terraform-provider-aws/issues/14528))
* resource/aws_route53_health_check: Add `disabled` argument ([#14614](https://github.com/terraform-providers/terraform-provider-aws/issues/14614))

BUG FIXES

* data-source/aws_launch_template: Prevent type error with `network_interfaces` `delete_on_termination` attribute ([#14599](https://github.com/terraform-providers/terraform-provider-aws/issues/14599))
* resource/aws_acm_certificate_validation: Prevent panic with missing `DomainValidationOptions` `ResourceRecord` attribute in API response [[#14590](https://github.com/terraform-providers/terraform-provider-aws/issues/14590)] 
* resource/aws_ecr_repository: Prevent panic with missing `EncryptionConfiguration` attribute in API response ([#14584](https://github.com/terraform-providers/terraform-provider-aws/issues/14584))
* resource/aws_wafv2_rule_group: Prevent unnecessary resource recreation with `rule` updates ([#14617](https://github.com/terraform-providers/terraform-provider-aws/issues/14617))
* resource/aws_wafv2_web_acl: Prevent unnecessary resource recreation with `rule` updates ([#14616](https://github.com/terraform-providers/terraform-provider-aws/issues/14616))

## 3.1.0 (August 07, 2020)

NOTES:

* resource/aws_route53_zone_association: The addition of cross-account zone association support required the use of new `ListHostedZonesByVPC` API call and adding the VPC Region to the resource ID for new resources. Restrictive IAM permissions for Terraform and cross-region imports may require updates. ([#14215](https://github.com/terraform-providers/terraform-provider-aws/issues/14215))

FEATURES

* **New Data Source:** `aws_ec2_spot_price` ([#12504](https://github.com/terraform-providers/terraform-provider-aws/issues/12504))
* **New Resource**: `aws_route53_vpc_association_authorization` ([#14215](https://github.com/terraform-providers/terraform-provider-aws/issues/14215))

ENHANCEMENTS

* data-source/aws_ecr_repository: Allow `registry_id` as an argument ([#14368](https://github.com/terraform-providers/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `image_scanning_configuration` and `image_tag_mutability` attributes ([#14368](https://github.com/terraform-providers/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/terraform-providers/terraform-provider-aws/issues/14520))
* resource/aws_api_gateway_method_settings: Plan-time validation added to `settings` `unauthorized_cache_control_header_strategy` and `logging_level` arguments ([#12651](https://github.com/terraform-providers/terraform-provider-aws/issues/12651))
* resource/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/terraform-providers/terraform-provider-aws/issues/14520))
* resource/aws_lb: Add `subnet_mapping` configuration block `private_ipv4_address` argument ([#11404](https://github.com/terraform-providers/terraform-provider-aws/issues/11404))
* resource/aws_rds_global_cluster: Add `force_destroy` and `source_db_cluster_identifier` arguments ([#14487](https://github.com/terraform-providers/terraform-provider-aws/issues/14487))
* resource/aws_rds_global_cluster: Add `global_cluster_members` attribute ([#14487](https://github.com/terraform-providers/terraform-provider-aws/issues/14487))
* resource/aws_route53_zone_association: Cross-account zone associations can now be created in conjunction with the new `aws_route53_vpc_association_authorization` resource ([#14215](https://github.com/terraform-providers/terraform-provider-aws/issues/14215))
* resource/aws_ssm_parameter: Add `data_type` argument (support `aws:ec2:image` parameters) ([#13326](https://github.com/terraform-providers/terraform-provider-aws/issues/13326))

BUG FIXES

* data-source/aws_availability_zones: Prevent unexpected plan output every apply with `group_names` attribute ([#14412](https://github.com/terraform-providers/terraform-provider-aws/issues/14412))
* data-source/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/terraform-providers/terraform-provider-aws/issues/14481))
* resource/aws_api_gateway_method_settings: Allow `settings` `cache_ttl_in_seconds` argument to be set to 0 ([#12651](https://github.com/terraform-providers/terraform-provider-aws/issues/12651))
* resource/aws_elastictranscoder_preset: Prevent empty configuration block panics ([#14092](https://github.com/terraform-providers/terraform-provider-aws/issues/14092))
* resource/aws_lambda_event_source_mapping: Allow `maximum_retry_attempts` argument to be set to 0 ([#12479](https://github.com/terraform-providers/terraform-provider-aws/issues/12479))
* resource/aws_rds_cluster: Add an `InvalidDBClusterStateFault` retryable error condition for clusters part of a global cluster ([#14420](https://github.com/terraform-providers/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Increase retry timeout for deletion to 2 minutes ([#14420](https://github.com/terraform-providers/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Prevent error when both `global_cluster_identifier` and `replication_source_identifier` are configured on creation ([#14490](https://github.com/terraform-providers/terraform-provider-aws/issues/14490))
* resource/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/terraform-providers/terraform-provider-aws/issues/14481))
* resource/aws_secretsmanager_secret: Allow retries for IAM eventual consistency errors ([#14459](https://github.com/terraform-providers/terraform-provider-aws/issues/14459))
* resource/aws_security_group: Ensure `name_prefix` argument with hex digits `a` through `f` is properly imported ([#14475](https://github.com/terraform-providers/terraform-provider-aws/issues/14475))
* resource/aws_spot_fleet_request: Allow `target_capacity` argument to be updated to 0 ([#12759](https://github.com/terraform-providers/terraform-provider-aws/issues/12759))
* resource/aws_spot_fleet_request: Wait for modify operation completion (default timeout of 10 minutes) ([#12759](https://github.com/terraform-providers/terraform-provider-aws/issues/12759))
* resource/aws_vpc_dhcp_options_association: Properly trigger resource recreation when VPC is deleted outside Terraform ([#14367](https://github.com/terraform-providers/terraform-provider-aws/issues/14367))

## 3.0.0 (July 31, 2020)

NOTES:
* provider: This version is built using Go 1.14.5, including security fixes to the crypto/x509 and net/http packages.

BREAKING CHANGES

* provider: New versions of the provider can only be automatically installed on Terraform 0.12 and later ([#14143](https://github.com/terraform-providers/terraform-provider-aws/issues/14143))
* provider: All "removed" attributes are cut, using them would result in a Terraform Core level error ([#14001](https://github.com/terraform-providers/terraform-provider-aws/issues/14001))
* provider: Credential ordering has changed from static, environment, shared credentials, EC2 metadata, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) to static, environment, shared credentials, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: The `AWS_METADATA_TIMEOUT` environment variable no longer has any effect as we now depend on the default AWS Go SDK EC2 Metadata client timeout of one second with two retries ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Remove deprecated `kinesis_analytics` and `r53` custom service endpoint arguments ([#14238](https://github.com/terraform-providers/terraform-provider-aws/issues/14238))
* data-source/aws_availability_zones: Remove deprecated `blacklisted_names` and `blacklisted_zone_ids` arguments ([#14134](https://github.com/terraform-providers/terraform-provider-aws/issues/14134))
* data-source/aws_directory_service_directory: Return an error when a single result is not found ([#14006](https://github.com/terraform-providers/terraform-provider-aws/issues/14006))
* data-source/aws_ecr_repository: Return an error when a single result is not found ([#10520](https://github.com/terraform-providers/terraform-provider-aws/issues/10520))
* data-source/aws_efs_file_system: Return an error when a single result is not found ([#14005](https://github.com/terraform-providers/terraform-provider-aws/issues/14005))
* data-source/aws_launch_template: Return an error when a single result is not found ([#10521](https://github.com/terraform-providers/terraform-provider-aws/issues/10521))
* data-source/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in data-source ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* data-source/aws_route53_zone: Trailing period removed from `name` argument set in data-source ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_acm_certificate: `certificate_body`, `certificate_chain`, and `private_key` attributes are no longer stored in the Terraform state with hash values ([#9685](https://github.com/terraform-providers/terraform-provider-aws/issues/9685))
* resource/aws_acm_certificate: `domain_validation_options` attribute changed from list to set ([#14199](https://github.com/terraform-providers/terraform-provider-aws/issues/14199))
* resource/aws_acm_certificate: Plan-time validation added to `domain_name` and `subject_alternative_names` arguments to prevent usage of strings with trailing periods ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_api_gateway_method_settings: Remove `Computed` property from `throttling_burst_limit` and `throttling_rate_limit` arguments, enabling drift detection ([#14266](https://github.com/terraform-providers/terraform-provider-aws/issues/14266))
* resource/aws_api_gateway_method_settings: Update `throttling_burst_limit` and `throttling_rate_limit` argument defaults to match API default of `-1` to keep throttling disabled ([#14266](https://github.com/terraform-providers/terraform-provider-aws/issues/14266))
* resource/aws_autoscaling_group: `availability_zones` and `vpc_zone_identifier` argument conflict now reported at plan-time ([#12927](https://github.com/terraform-providers/terraform-provider-aws/issues/12927))
* resource/aws_autoscaling_group: Remove `Computed` property from `load_balancers` and `target_group_arns` arguments, enabling drift detection ([#14064](https://github.com/terraform-providers/terraform-provider-aws/issues/14064))
* resource/aws_cloudfront_distribution: `active_trusted_signers` argument renamed to `trusted_signers` to support accessing `items` in Terraform 0.12 ([#14339](https://github.com/terraform-providers/terraform-provider-aws/issues/14339))
* resource/aws_cloudwatch_log_group: Automatically trim `:*` suffix from `arn` attribute ([#14214](https://github.com/terraform-providers/terraform-provider-aws/issues/14214))
* resource/aws_codepipeline: Removes `GITHUB_TOKEN` environment variable ([#14175](https://github.com/terraform-providers/terraform-provider-aws/issues/14175))
* resource/aws_cognito_user_pool: Remove deprecated `admin_create_user_config` configuration block `unused_account_validity_days` argument ([#14294](https://github.com/terraform-providers/terraform-provider-aws/issues/14294))
* resource/aws_dx_gateway: Remove automatic `aws_dx_gateway_association` resource import ([#14124](https://github.com/terraform-providers/terraform-provider-aws/issues/14124))
* resource/aws_dx_gateway_association: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/terraform-providers/terraform-provider-aws/issues/14144))
* resource/aws_dx_gateway_association_proposal: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/terraform-providers/terraform-provider-aws/issues/14144))
* resource/aws_ebs_volume: Return an error when `iops` argument set to a value greater than 0 for volume types other than `io1` ([#14310](https://github.com/terraform-providers/terraform-provider-aws/issues/14310))
* resource/aws_elastic_transcoder_preset: Remove `video` configuration block `max_frame_rate` argument default value ([#7141](https://github.com/terraform-providers/terraform-provider-aws/issues/7141))
* resource/aws_emr_cluster: Remove deprecated `instance_group` configuration block, `core_instance_count`, `core_instance_type`, and `master_instance_type` arguments ([#14137](https://github.com/terraform-providers/terraform-provider-aws/issues/14137))
* resource/aws_glue_job: Remove deprecated `allocated_capacity` argument ([#14296](https://github.com/terraform-providers/terraform-provider-aws/issues/14296))
* resource/aws_iam_access_key: Remove deprecated `ses_smtp_password` attribute ([#14299](https://github.com/terraform-providers/terraform-provider-aws/issues/14299))
* resource/aws_iam_instance_profile: Remove deprecated `roles` argument ([#14303](https://github.com/terraform-providers/terraform-provider-aws/issues/14303))
* resource/aws_iam_server_certificate: Remove state hashing from `certificate_body`, `certificate_chain`, and `private_key` arguments for new or recreated resources ([#14187](https://github.com/terraform-providers/terraform-provider-aws/issues/14187))
* resource/aws_instance: Return an error when `ebs_block_device` `iops` or `root_block_device` `iops` argument set to a value greater than `0` for volume types other than `io1` ([#14310](https://github.com/terraform-providers/terraform-provider-aws/issues/14310))
* resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN ([#12876](https://github.com/terraform-providers/terraform-provider-aws/issues/12876))
* resource/aws_launch_template: `network_interfaces` `delete_on_termination` argument changed from `bool` to `string` type ([#8612](https://github.com/terraform-providers/terraform-provider-aws/issues/8612))
* resource/aws_lb_listener_rule: Remove deprecated `condition` configuration block `field` and `values` arguments ([#14309](https://github.com/terraform-providers/terraform-provider-aws/issues/14309))
* resource/aws_msk_cluster: Update `encryption_info` `encryption_in_transit` `client_broker` argument default to match API default of `TLS` ([#14132](https://github.com/terraform-providers/terraform-provider-aws/issues/14132))
* resource/aws_rds_cluster: Update `scaling_configuration` `min_capacity` argument default to match API default of `1` ([#14268](https://github.com/terraform-providers/terraform-provider-aws/issues/14268))
* resource/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in resource ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_route53_zone: Trailing period removed from `name` argument set in resource ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_s3_bucket: Remove automatic `aws_s3_bucket_policy` resource import ([#14121](https://github.com/terraform-providers/terraform-provider-aws/issues/14121))
* resource/aws_s3_bucket: Convert `region` to read-only attribute ([#14127](https://github.com/terraform-providers/terraform-provider-aws/issues/14127))
* resource/aws_s3_bucket_metric: Update `filter` argument to require at least one of the `prefix` or `tags` nested arguments ([#14230](https://github.com/terraform-providers/terraform-provider-aws/issues/14230))
* resource/aws_security_group: Remove automatic `aws_security_group_rule` resource import ([#12616](https://github.com/terraform-providers/terraform-provider-aws/issues/12616))
* resource/aws_ses_domain_identity: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_ses_domain_identity_verification: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_sns_platform_application: `platform_credential` and `platform_principal` attributes are no longer stored in the Terraform state with hash values ([#3894](https://github.com/terraform-providers/terraform-provider-aws/issues/3894))
* resource/aws_spot_fleet_request: Remove 24 hour default for `valid_until` argument ([#9718](https://github.com/terraform-providers/terraform-provider-aws/issues/9718))
* resource/aws_ssm_maintenance_window_task: Remove deprecated `logging_info` and `task_parameters` configuration blocks ([#14311](https://github.com/terraform-providers/terraform-provider-aws/issues/14311))

FEATURES

* **New Data Source:** aws_workspaces_directory ([#13529](https://github.com/terraform-providers/terraform-provider-aws/issues/13529))

ENHANCEMENTS

* provider: Always enable shared configuration file support (no longer require `AWS_SDK_LOAD_CONFIG` environment variable) ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Add `assume_role` configuration block `duration_seconds`, `policy_arns`, `tags`, and `transitive_tag_keys` arguments ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* data-source/aws_instance: Add `secondary_private_ips` attribute ([#14079](https://github.com/terraform-providers/terraform-provider-aws/issues/14079))
* data-source/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/terraform-providers/terraform-provider-aws/issues/14221))
* resource/aws_acm_certificate: Enable `domain_validation_options` usage in downstream resource `count` and `for_each` references ([#14199](https://github.com/terraform-providers/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Add plan-time validation to `authorizer_credentials` argument ([#12643](https://github.com/terraform-providers/terraform-provider-aws/issues/12643))
* resource/aws_api_gateway_method_settings: Add import support ([#14266](https://github.com/terraform-providers/terraform-provider-aws/issues/14266))
* resource/aws_apigatewayv2_integration: Add `request_parameters` attribute ([#14080](https://github.com/terraform-providers/terraform-provider-aws/issues/14080))
* resource/aws_apigatewayv2_integration: Add `tls_config` attribute ([#13013](https://github.com/terraform-providers/terraform-provider-aws/issues/13013))
* resource/aws_apigatewayv2_route: Support for updating route key ([#13833](https://github.com/terraform-providers/terraform-provider-aws/issues/13833))
* resource/aws_apigatewayv2_stage: Make `deployment_id` a `Computed` attribute ([#13644](https://github.com/terraform-providers/terraform-provider-aws/issues/13644))
* resource/aws_fsx_lustre_file_system: Add `deployment_type` and `per_unit_storage_throughput` attributes ([#13639](https://github.com/terraform-providers/terraform-provider-aws/issues/13639))
* resource_aws_fsx_windows_file_system - add `storage_type` argument. ([#14316](https://github.com/terraform-providers/terraform-provider-aws/issues/14316))
* resource_aws_fsx_windows_file_system: add support for multi-az ([#12676](https://github.com/terraform-providers/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: add `SINGLE_AZ_2` deployment type ([#12676](https://github.com/terraform-providers/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: adds `preferred_file_server_ip`, `remote_administration_endpoint` attributes ([#12676](https://github.com/terraform-providers/terraform-provider-aws/issues/12676))
* resource/aws_instance: Add `secondary_private_ips` argument (conflicts with `network_interface` configuration block) ([#14079](https://github.com/terraform-providers/terraform-provider-aws/issues/14079))

BUG FIXES

* provider: Ensure nil is not passed to RetryError helpers, may result in some bug fixes ([#14104](https://github.com/terraform-providers/terraform-provider-aws/issues/14104))
* provider: Ensure configured STS endpoint is used during `AssumeRole` API calls ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Prefer AWS shared configuration over EC2 metadata credentials by default ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Prefer CodeBuild, ECS, EKS credentials over EC2 metadata credentials by default ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* data-source/aws_lb: `enable_http2` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_acm_certificate: Prevent unexpected ordering differences with `domain_validation_options` attribute ([#14199](https://github.com/terraform-providers/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Allow `authorizer_result_ttl_in_seconds` to be set to 0 ([#12643](https://github.com/terraform-providers/terraform-provider-aws/issues/12643))
* resource/aws_apigatewayv2_integration: Correctly handle the `integration_method` attribute for AWS Lambda integrations([#13266](https://github.com/terraform-providers/terraform-provider-aws/issues/13266))
* resource/aws_apigatewayv2_integration: Correctly handle the `passthrough_behavior` attribute for HTTP APIs ([#13062](https://github.com/terraform-providers/terraform-provider-aws/issues/13062))
* resource/aws_apigatewayv2_stage: Correctly handle `default_route_setting` and `route_setting` `data_trace_enabled` and `logging_level` for HTTP APIs. `logging_level` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. ([#13809](https://github.com/terraform-providers/terraform-provider-aws/issues/13809))
* resource/aws_appautoscaling_target: Only retry `DeregisterScalableTarget` retries on all errors on deletion ([#14259](https://github.com/terraform-providers/terraform-provider-aws/issues/14259))
* resource/aws_dx_gateway_association: Increase default create/update/delete timeouts to 30 minutes ([#14144](https://github.com/terraform-providers/terraform-provider-aws/issues/14144))
* resource/aws_codepipeline: Only retry `CreatePipeline` errors for IAM eventual consistency errors ([#14264](https://github.com/terraform-providers/terraform-provider-aws/issues/14264))
* resource/aws_elasticsearch_domain: Update method to properly set `advanced_security_options` ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_lambda_function: Increase IAM retry timeout for creation to standard 2 minute timeout ([#14291](https://github.com/terraform-providers/terraform-provider-aws/issues/14291))
* resource/aws_lb_cookie_stickiness_policy: `lb_port` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_network_acl_rule: Immediately return `DescribeNetworkAcls` errors on creation ([#14261](https://github.com/terraform-providers/terraform-provider-aws/issues/14261))
* resource/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/terraform-providers/terraform-provider-aws/issues/14221))
* resource/aws_sns_topic_subscription: Immediately return `ListSubscriptionsByTopic` errors ([#14262](https://github.com/terraform-providers/terraform-provider-aws/issues/14262))
* resource/aws_spot_fleet_request: Only retry `RequestSpotFleet` on IAM eventual consistency errors and use standard 2 minute timeout ([#14265](https://github.com/terraform-providers/terraform-provider-aws/issues/14265))
* resource/aws_spot_instance_request: `primary_network_interface_id` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_ssm_activation: Only retry `CreateActivation` on IAM eventual consistency errors and use standard 2 minute timeout ([#14263](https://github.com/terraform-providers/terraform-provider-aws/issues/14263))
* resource/aws_ssm_association: `parameters` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))

## Previous Releases

For information on prior major releases, see their changelogs:

* [2.x and earlier](https://github.com/terraform-providers/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
