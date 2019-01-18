## 2.0.0 (Unreleased)
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
* resource/aws_iot_topic_rule: Add `firehose` `seperator` argument ([#5734](https://github.com/terraform-providers/terraform-provider-aws/issues/5734))
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

This release is happening outside the normal release schedule to accomodate a crash fix for the `aws_lb_target_group` resource. It appears an ELBv2 service update rolling out currently is the root cause. The potential for this crash has been present since the initial resource in Terraform 0.7.7 and all versions of the AWS provider up to v1.13.0.

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
* resource/aws_eip_assocation: Retry association for pending instances ([#3072](https://github.com/terraform-providers/terraform-provider-aws/issues/3072))
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
