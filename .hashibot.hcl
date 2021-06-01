poll "closed_issue_locker" "locker" {
  schedule             = "0 10 17 * * *"
  closed_for           = "720h" # 30 days
  max_issues           = 500
  sleep_between_issues = "5s"

  message = <<-EOF
    I'm going to lock this issue because it has been closed for _30 days_ ⏳. This helps our maintainers find and focus on the active issues.

    If you feel this issue should be reopened, we encourage creating a new issue linking back to this one for added context. Thanks!
  EOF
}

behavior "deprecated_import_commenter" "hashicorp_terraform" {
  import_regexp = "github.com/hashicorp/terraform/"
  marker_label  = "terraform-plugin-sdk-migration"

  message = <<-EOF
    Hello, and thank you for your contribution!

    This project recently migrated to the [standalone Terraform Plugin SDK](https://www.terraform.io/docs/extend/plugin-sdk.html). While the migration helps speed up future feature requests and bug fixes to the Terraform AWS Provider's interface with Terraform, it has the unfortunate consequence of requiring minor changes to pull requests created using the old SDK.

    This pull request appears to include the Go import path `${var.import_path}`, which was from the older SDK. The newer SDK uses import paths beginning with `github.com/hashicorp/terraform-plugin-sdk/`.

    To resolve this situation without losing any existing work, you may be able to Git rebase your branch against the current master branch (example below); replacing any remaining old import paths with the newer ones.

    ```console
    $ git fetch --all
    $ git rebase origin/master
    ```

    Another option is to create a new branch from the current master with the same code changes (replacing the import paths), submit a new pull request, and close this existing pull request.

    We apologize for this inconvenience and appreciate your effort. Thank you for contributing and helping make the Terraform AWS Provider better for everyone.
  EOF
}

behavior "deprecated_import_commenter" "sdkv1" {
  import_regexp = "github.com/hashicorp/terraform-plugin-sdk/(helper/(acctest|customdiff|logging|resource|schema|structure|validation)|terraform)"
  marker_label  = "terraform-plugin-sdk-v1"

  message = <<-EOF
    Hello, and thank you for your contribution!

    This project recently upgraded to [V2 of the Terraform Plugin SDK](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html)

    This pull request appears to include at least one V1 import path of the SDK (`${var.import_path}`). Please import the V2 path `github.com/hashicorp/terraform-plugin-sdk/v2/helper/PACKAGE`

    To resolve this situation without losing any existing work, you may be able to Git rebase your branch against the current master branch (example below); replacing any remaining old import paths with the newer ones.

    ```console
    $ git fetch --all
    $ git rebase origin/master
    ```

    Another option is to create a new branch from the current master with the same code changes (replacing the import paths), submit a new pull request, and close this existing pull request.

    We apologize for this inconvenience and appreciate your effort. Thank you for contributing and helping make the Terraform AWS Provider better for everyone.
  EOF
}

behavior "deprecated_import_commenter" "sdkv1_deprecated" {
  import_regexp = "github.com/hashicorp/terraform-plugin-sdk/helper/(hashcode|mutexkv|encryption)"
  marker_label  = "terraform-plugin-sdk-v1"

  message = <<-EOF
    Hello, and thank you for your contribution!
    This pull request appears to include the Go import path `${var.import_path}`, which was deprecated after upgrading to [V2 of the Terraform Plugin SDK](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html).
    You may use a now internalized version of the package found in `github.com/terraform-providers/terraform-provider-aws/aws/internal/PACKAGE`.
  EOF
}

queued_behavior "release_commenter" "releases" {
  repo_prefix = "terraform-provider-"

  message = <<-EOF
    This has been released in [version ${var.release_version} of the Terraform AWS provider](${var.changelog_link}). Please see the [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#provider-versions) or reach out if you need any assistance upgrading.

    For further feature requests or bug reports with this functionality, please create a [new GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/new/choose) following the template for triage. Thanks!
  EOF
}

# Catch the following in issues:
# *aws_XXX
# * aws_XXX
# * `aws_XXX`
# -aws_XXX
# - aws_XXX
# - `aws_XXX`
# data "aws_XXX"
# resource "aws_XXX"
# NOTE: Go regexp does not support negative lookaheads
behavior "regexp_issue_labeler_v2" "service_labels" {
  regexp = "(\\* ?`?|- ?`?|data \"|resource \")aws_(\\w+)"

  label_map = {
    "service/accessanalyzer" = [
      "aws_accessanalyzer_",
    ],
    "service/acm" = [
      "aws_acm_",
    ],
    "service/acmpca" = [
      "aws_acmpca_",
    ],
    "service/alexaforbusiness" = [
      "aws_alexaforbusiness_",
    ],
    "service/amplify" = [
      "aws_amplify_",
    ],
    "service/apigateway" = [
      # Catch aws_api_gateway_XXX but not aws_api_gateway_v2_
      "aws_api_gateway_([^v]|v[^2]|v2[^_])",
    ],
    "service/apigatewayv2" = [
      "aws_api_gateway_v2_",
      "aws_apigatewayv2_",
    ],
    "service/applicationautoscaling" = [
      "aws_appautoscaling_",
    ],
    "service/applicationdiscoveryservice" = [
      "aws_applicationdiscoveryservice_",
    ],
    "service/applicationinsights" = [
      "aws_applicationinsights_",
    ],
    "service/appmesh" = [
      "aws_appmesh_",
    ],
    "service/appstream" = [
      "aws_appstream_",
    ],
    "service/appsync" = [
      "aws_appsync_",
    ],
    "service/athena" = [
      "aws_athena_",
    ],
    "service/autoscaling" = [
      "aws_autoscaling_",
      "aws_launch_configuration",
    ],
    "service/autoscalingplans" = [
      "aws_autoscalingplans_",
    ],
    "service/backup" = [
      "aws_backup_",
    ],
    "service/batch" = [
      "aws_batch_",
    ],
    "service/budgets" = [
      "aws_budgets_",
    ],
    "service/cloud9" = [
      "aws_cloud9_",
    ],
    "service/clouddirectory" = [
      "aws_clouddirectory_",
    ],
    "service/cloudformation" = [
      "aws_cloudformation_",
    ],
    "service/cloudfront" = [
      "aws_cloudfront_",
    ],
    "service/cloudhsmv2" = [
      "aws_cloudhsm_v2_",
    ],
    "service/cloudsearch" = [
      "aws_cloudsearch_",
    ],
    "service/cloudtrail" = [
      "aws_cloudtrail",
    ],
    "service/cloudwatch" = [
      "aws_cloudwatch_([^e]|e[^v]|ev[^e]|eve[^n]|even[^t]|event[^_]|[^l]|l[^o]|lo[^g]|log[^_])",
    ],
    "service/cloudwatchevents" = [
      "aws_cloudwatch_event_",
    ],
    "service/cloudwatchlogs" = [
      "aws_cloudwatch_log_",
    ],
    "service/codeartifact" = [
      "aws_codeartifact_",
    ],
    "service/codebuild" = [
      "aws_codebuild_",
    ],
    "service/codecommit" = [
      "aws_codecommit_",
    ],
    "service/codedeploy" = [
      "aws_codedeploy_",
    ],
    "service/codepipeline" = [
      "aws_codepipeline",
    ],
    "service/codestar" = [
      "aws_codestar_",
    ],
    "service/codestarconnections" = [
      "aws_codestarconnections_",
    ],
    "service/codestarnotifications" = [
      "aws_codestarnotifications_",
    ],
    "service/cognito" = [
      "aws_cognito_",
    ],
    "service/configservice" = [
      "aws_config_",
    ],
    "service/connect" = [
      "aws_connect_",
    ],
    "service/databasemigrationservice" = [
      "aws_dms_",
    ],
    "service/dataexchange" = [
      "aws_dataexchange_",
    ],
    "service/datapipeline" = [
      "aws_datapipeline_",
    ],
    "service/datasync" = [
      "aws_datasync_",
    ],
    "service/dax" = [
      "aws_dax_",
    ],
    "service/devicefarm" = [
      "aws_devicefarm_",
    ],
    "service/directconnect" = [
      "aws_dx_",
    ],
    "service/directoryservice" = [
      "aws_directory_service_",
    ],
    "service/dlm" = [
      "aws_dlm_",
    ],
    "service/docdb" = [
      "aws_docdb_",
    ],
    "service/dynamodb" = [
      "aws_dynamodb_",
    ],
    "service/ec2" = [
      "aws_ami",
      "aws_availability_zone",
      "aws_customer_gateway",
      "aws_(default_)?(network_acl|route_table|security_group|subnet|vpc)",
      "aws_ebs_",
      "aws_ec2_",
      "aws_egress_only_internet_gateway",
      "aws_eip",
      "aws_flow_log",
      "aws_instance",
      "aws_internet_gateway",
      "aws_key_pair",
      "aws_launch_template",
      "aws_main_route_table_association",
      "aws_network_interface",
      "aws_placement_group",
      "aws_prefix_list",
      "aws_spot",
      "aws_route(\"|`|$)",
      "aws_vpn_",
      "aws_volume_attachment",
    ],
    "service/ecr" = [
      "aws_ecr_",
    ],
    "service/ecrpublic" = [
      "aws_ecrpublic_",
    ],
    "service/ecs" = [
      "aws_ecs_",
    ],
    "service/efs" = [
      "aws_efs_",
    ],
    "service/eks" = [
      "aws_eks_",
    ],
    "service/elastic-transcoder" = [
      "aws_elastictranscoder_",
    ],
    "service/elasticache" = [
      "aws_elasticache_",
    ],
    "service/elasticbeanstalk" = [
      "aws_elastic_beanstalk_",
    ],
    "service/elasticsearch" = [
      "aws_elasticsearch_",
    ],
    "service/elb" = [
      "aws_app_cookie_stickiness_policy",
      "aws_elb",
      "aws_lb_cookie_stickiness_policy",
      "aws_lb_ssl_negotiation_policy",
      "aws_load_balancer_",
      "aws_proxy_protocol_policy",
    ],
    "service/elbv2" = [
      "aws_(a)?lb(\"|`|$)",
      # Catch aws_lb_XXX but not aws_lb_cookie_ or aws_lb_ssl_ (Classic ELB)
      "aws_(a)?lb_([^c]|c[^o]|co[^o]|coo[^k]|cook[^i]|cooki[^e]|cookie[^_]|[^s]|s[^s]|ss[^l]|ssl[^_])",
    ],
    "service/emr" = [
      "aws_emr_",
    ],
    "service/eventbridge" = [
      # EventBridge is rebranded CloudWatch Events
      "aws_cloudwatch_event_",
    ],
    "service/firehose" = [
      "aws_kinesis_firehose_",
    ],
    "service/fms" = [
      "aws_fms_",
    ],
    "service/forecast" = [
      "aws_forecast_",
    ],
    "service/fsx" = [
      "aws_fsx_",
    ],
    "service/gamelift" = [
      "aws_gamelift_",
    ],
    "service/glacier" = [
      "aws_glacier_",
    ],
    "service/globalaccelerator" = [
      "aws_globalaccelerator_",
    ],
    "service/glue" = [
      "aws_glue_",
    ],
    "service/greengrass" = [
      "aws_greengrass_",
    ],
    "service/guardduty" = [
      "aws_guardduty_",
    ],
    "service/iam" = [
      "aws_iam_",
    ],
    "service/identitystore" = [
      "aws_identitystore_",
    ],
    "service/imagebuilder" = [
      "aws_imagebuilder_",
    ],
    "service/inspector" = [
      "aws_inspector_",
    ],
    "service/iot" = [
      "aws_iot_",
    ],
    "service/iotanalytics" = [
      "aws_iotanalytics_",
    ],
    "service/iotevents" = [
      "aws_iotevents_",
    ],
    "service/kafka" = [
      "aws_msk_",
    ],
    "service/kinesis" = [
      # Catch aws_kinesis_XXX but not aws_kinesis_firehose_
      "aws_kinesis_([^f]|f[^i]|fi[^r]|fir[^e]|fire[^h]|fireh[^o]|fireho[^s]|firehos[^e]|firehose[^_])",
    ],
    "service/kinesisanalytics" = [
      "aws_kinesis_analytics_",
    ],
    "service/kinesisanalyticsv2" = [
      "aws_kinesisanalyticsv2_",
    ],
    "service/kms" = [
      "aws_kms_",
    ],
    "service/lakeformation" = [
      "aws_lakeformation_",
    ],
    "service/lambda" = [
      "aws_lambda_",
    ],
    "service/lexmodelbuildingservice" = [
      "aws_lex_",
    ],
    "service/licensemanager" = [
      "aws_licensemanager_",
    ],
    "service/lightsail" = [
      "aws_lightsail_",
    ],
    "service/machinelearning" = [
      "aws_machinelearning_",
    ],
    "service/macie" = [
      "aws_macie_",
    ],
    "service/macie2" = [
      "aws_macie2_",
    ],
    "service/marketplacecatalog" = [
      "aws_marketplace_catalog_",
    ],
    "service/mediaconnect" = [
      "aws_media_connect_",
    ],
    "service/mediaconvert" = [
      "aws_media_convert_",
    ],
    "service/medialive" = [
      "aws_media_live_",
    ],
    "service/mediapackage" = [
      "aws_media_package_",
    ],
    "service/mediastore" = [
      "aws_media_store_",
    ],
    "service/mediatailor" = [
      "aws_media_tailor_",
    ],
    "service/mobile" = [
      "aws_mobile_",
    ],
    "service/mq" = [
      "aws_mq_",
    ],
    "service/mwaa" = [
      "aws_mwaa_",
    ],
    "service/neptune" = [
      "aws_neptune_",
    ],
    "service/networkfirewall" = [
      "aws_networkfirewall_",
    ],
    "service/networkmanager" = [
      "aws_networkmanager_",
    ],
    "service/opsworks" = [
      "aws_opsworks_",
    ],
    "service/organizations" = [
      "aws_organizations_",
    ],
    "service/outposts" = [
      "aws_outposts_",
    ],
    "service/personalize" = [
      "aws_personalize_",
    ],
    "service/pinpoint" = [
      "aws_pinpoint_",
    ],
    "service/polly" = [
      "aws_polly_",
    ],
    "service/pricing" = [
      "aws_pricing_",
    ],
    "service/qldb" = [
      "aws_qldb_",
    ],
    "service/quicksight" = [
      "aws_quicksight_",
    ],
    "service/ram" = [
      "aws_ram_",
    ],
    "service/rds" = [
      "aws_db_",
      "aws_rds_",
    ],
    "service/redshift" = [
      "aws_redshift_",
    ],
    "service/resourcegroups" = [
      "aws_resourcegroups_",
    ],
    "service/resourcegroupstaggingapi" = [
      "aws_resourcegroupstaggingapi_",
    ],
    "service/robomaker" = [
      "aws_robomaker_",
    ],
    "service/route53" = [
      # Catch aws_route53_XXX but not aws_route53_domains_ or aws_route53_resolver_
      "aws_route53_([^d]|d[^o]|do[^m]|dom[^a]|doma[^i]|domai[^n]|domain[^s]|domains[^_]|[^r]|r[^e]|re[^s]|res[^o]|reso[^l]|resol[^v]|resolv[^e]|resolve[^r]|resolver[^_])",
    ],
    "service/route53domains" = [
      "aws_route53domains_",
    ],
    "service/route53resolver" = [
      "aws_route53_resolver_",
    ],
    "service/s3" = [
      "aws_canonical_user_id",
      "aws_s3_bucket",
    ],
    "service/s3control" = [
      "aws_s3_account_",
      "aws_s3control_",
    ],
    "service/s3outposts" = [
      "aws_s3outposts_",
    ],
    "service/sagemaker" = [
      "aws_sagemaker_",
    ],
    "service/secretsmanager" = [
      "aws_secretsmanager_",
    ],
    "service/securityhub" = [
      "aws_securityhub_",
    ],
    "service/serverlessapplicationrepository" = [
      "aws_serverlessapplicationrepository_",
    ],
    "service/servicecatalog" = [
      "aws_servicecatalog_",
    ],
    "service/servicediscovery" = [
      "aws_service_discovery_",
    ],
    "service/servicequotas" = [
      "aws_servicequotas_",
    ],
    "service/ses" = [
      "aws_ses_",
    ],
    "service/sfn" = [
      "aws_sfn_",
    ],
    "service/shield" = [
      "aws_shield_",
    ],
    "service/signer" = [
      "aws_signer_",
    ],
    "service/simpledb" = [
      "aws_simpledb_",
    ],
    "service/snowball" = [
      "aws_snowball_",
    ],
    "service/sns" = [
      "aws_sns_",
    ],
    "service/sqs" = [
      "aws_sqs_",
    ],
    "service/ssm" = [
      "aws_ssm_",
    ],
    "service/ssoadmin" = [
      "aws_ssoadmin_",
    ],
    "service/storagegateway" = [
      "aws_storagegateway_",
    ],
    "service/sts" = [
      "aws_caller_identity",
    ],
    "service/swf" = [
      "aws_swf_",
    ],
    "service/synthetics" = [
      "aws_synthetics_",
    ],
    "service/timestreamwrite" = [
      "aws_timestreamwrite_",
    ],
    "service/transfer" = [
      "aws_transfer_",
    ],
    "service/waf" = [
      "aws_waf_",
      "aws_wafregional_",
    ],
    "service/wafv2" = [
      "aws_wafv2_",
    ],
    "service/workdocs" = [
      "aws_workdocs_",
    ],
    "service/worklink" = [
      "aws_worklink_",
    ],
    "service/workmail" = [
      "aws_workmail_",
    ],
    "service/workspaces" = [
      "aws_workspaces_",
    ],
    "service/xray" = [
      "aws_xray_",
    ],
  }
}

behavior "pull_request_path_labeler" "service_labels" {
  label_map = {
    # label provider related changes
    "provider" = [
      "*.md",
      ".github/**/*",
      ".gitignore",
      ".go-version",
      ".hashibot.hcl",
      "aws/auth_helpers.go",
      "aws/awserr.go",
      "aws/config.go",
      "aws/*_aws_arn*",
      "aws/*_aws_ip_ranges*",
      "aws/*_aws_partition*",
      "aws/*_aws_region*",
      "aws/internal/flatmap/*",
      "aws/internal/keyvaluetags/*",
      "aws/internal/naming/*",
      "aws/provider.go",
      "aws/utils.go",
      "docs/*.md",
      "docs/contributing/**/*",
      "GNUmakefile",
      "infrastructure/**/*",
      "main.go",
      "website/docs/index.html.markdown",
      "website/**/arn*",
      "website/**/ip_ranges*",
      "website/**/partition*",
      "website/**/region*"
    ]
    "dependencies" = [
      ".github/dependabot.yml",
    ]
    "documentation" = [
      "docs/**/*",
      "website/**/*",
      "*.md",
    ]
    "examples" = [
      "examples/**/*",
    ]
    "tests" = [
      "**/*_test.go",
      "**/testdata/**/*",
      "**/test-fixtures/**/*",
      ".github/workflows/*",
      ".gometalinter.json",
      ".markdownlinkcheck.json",
      ".markdownlint.yml",
      "staticcheck.conf"
    ]
    # label services
    "service/accessanalyzer" = [
      "aws/internal/service/accessanalyzer/**/*",
      "**/*_accessanalyzer_*",
      "**/accessanalyzer_*"
    ]
    "service/acm" = [
      "aws/internal/service/acm/**/*",
      "**/*_acm_*",
      "**/acm_*"
    ]
    "service/acmpca" = [
      "aws/internal/service/acmpca/**/*",
      "**/*_acmpca_*",
      "**/acmpca_*"
    ]
    "service/alexaforbusiness" = [
      "aws/internal/service/alexaforbusiness/**/*",
      "**/*_alexaforbusiness_*",
      "**/alexaforbusiness_*"
    ]
    "service/amplify" = [
      "aws/internal/service/amplify/**/*",
      "**/*_amplify_*",
      "**/amplify_*"
    ]
    "service/apigateway" = [
      "aws/internal/service/apigateway/**/*",
      "**/*_api_gateway_[^v][^2][^_]*",
      "**/*_api_gateway_vpc_link*",
      "**/api_gateway_[^v][^2][^_]*",
      "**/api_gateway_vpc_link*"
    ]
    "service/apigatewayv2" = [
      "aws/internal/service/apigatewayv2/**/*",
      "**/*_api_gateway_v2_*",
      "**/*_apigatewayv2_*",
      "**/api_gateway_v2_*",
      "**/apigatewayv2_*"
    ]
    "service/applicationautoscaling" = [
      "aws/internal/service/applicationautoscaling/**/*",
      "**/*_appautoscaling_*",
      "**/appautoscaling_*"
    ]
    "service/applicationinsights" = [
      "aws/internal/service/applicationinsights/**/*",
      "**/*_applicationinsights_*",
      "**/applicationinsights_*"
    ]
    "service/appmesh" = [
      "aws/internal/service/appmesh/**/*",
      "**/*_appmesh_*",
      "**/appmesh_*"
    ]
    "service/appstream" = [
      "aws/internal/service/appstream/**/*",
      "**/*_appstream_*",
      "**/appstream_*"
    ]
    "service/appsync" = [
      "aws/internal/service/appsync/**/*",
      "**/*_appsync_*",
      "**/appsync_*"
    ]
    "service/athena" = [
      "aws/internal/service/athena/**/*",
      "**/*_athena_*",
      "**/athena_*"
    ]
    "service/autoscaling" = [
      "aws/internal/service/autoscaling/**/*",
      "**/*_autoscaling_*",
      "**/autoscaling_*",
      "aws/*_aws_launch_configuration*",
      "website/**/launch_configuration*"
    ]
    "service/autoscalingplans" = [
      "aws/internal/service/autoscalingplans/**/*",
      "**/*_autoscalingplans_*",
      "**/autoscalingplans_*"
    ]
    "service/backup" = [
      "aws/internal/service/backup/**/*",
      "**/*backup_*",
      "**/backup_*"
    ]
    "service/batch" = [
      "aws/internal/service/batch/**/*",
      "**/*_batch_*",
      "**/batch_*"
    ]
    "service/budgets" = [
      "aws/internal/service/budgets/**/*",
      "**/*_budgets_*",
      "**/budgets_*"
    ]
    "service/cloud9" = [
      "aws/internal/service/cloud9/**/*",
      "**/*_cloud9_*",
      "**/cloud9_*"
    ]
    "service/clouddirectory" = [
      "aws/internal/service/clouddirectory/**/*",
      "**/*_clouddirectory_*",
      "**/clouddirectory_*"
    ]
    "service/cloudformation" = [
      "aws/internal/service/cloudformation/**/*",
      "**/*_cloudformation_*",
      "**/cloudformation_*"
    ]
    "service/cloudfront" = [
      "aws/internal/service/cloudfront/**/*",
      "**/*_cloudfront_*",
      "**/cloudfront_*"
    ]
    "service/cloudhsmv2" = [
      "aws/internal/service/cloudhsmv2/**/*",
      "**/*_cloudhsm_v2_*",
      "**/cloudhsm_v2_*"
    ]
    "service/cloudsearch" = [
      "aws/internal/service/cloudsearch/**/*",
      "**/*_cloudsearch_*",
      "**/cloudsearch_*"
    ]
    "service/cloudtrail" = [
      "aws/internal/service/cloudtrail/**/*",
      "**/*_cloudtrail*",
      "**/cloudtrail*"
    ]
    "service/cloudwatch" = [
      "aws/internal/service/cloudwatch/**/*",
      "**/*_cloudwatch_dashboard*",
      "**/*_cloudwatch_metric_alarm*",
      "**/cloudwatch_dashboard*",
      "**/cloudwatch_metric_alarm*"
    ]
    "service/cloudwatchevents" = [
      "aws/internal/service/cloudwatchevents/**/*",
      "**/*_cloudwatch_event_*",
      "**/cloudwatch_event_*"
    ]
    "service/cloudwatchlogs" = [
      "aws/internal/service/cloudwatchlogs/**/*",
      "**/*_cloudwatch_log_*",
      "**/cloudwatch_log_*"
    ]
    "service/codeartifact" = [
      "aws/internal/service/codeartifact/**/*",
      "**/*_codeartifact_*",
      "**/codeartifact_*"
    ]
    "service/codebuild" = [
      "aws/internal/service/codebuild/**/*",
      "**/*_codebuild_*",
      "**/codebuild_*"
    ]
    "service/codecommit" = [
      "aws/internal/service/codecommit/**/*",
      "**/*_codecommit_*",
      "**/codecommit_*"
    ]
    "service/codedeploy" = [
      "aws/internal/service/codedeploy/**/*",
      "**/*_codedeploy_*",
      "**/codedeploy_*"
    ]
    "service/codepipeline" = [
      "aws/internal/service/codepipeline/**/*",
      "**/*_codepipeline_*",
      "**/codepipeline_*"
    ]
    "service/codestar" = [
      "aws/internal/service/codestar/**/*",
      "**/*_codestar_*",
      "**/codestar_*"
    ]
    "service/codestarconnections" = [
      "aws/internal/service/codestarconnections/**/*",
      "**/*_codestarconnections_*",
      "**/codestarconnections_*"
    ]
    "service/codestarnotifications" = [
      "aws/internal/service/codestarnotifications/**/*",
      "**/*_codestarnotifications_*",
      "**/codestarnotifications_*"
    ]
    "service/cognito" = [
      "aws/internal/service/cognitoidentity/**/*",
      "aws/internal/service/cognitoidentityprovider/**/*",
      "**/*_cognito_*",
      "**/cognito_*"
    ]
    "service/comprehend" = [
      "aws/internal/service/comprehend/**/*",
      "**/*_comprehend_*",
      "**/comprehend_*"
    ]
    "service/configservice" = [
      "aws/internal/service/configservice/**/*",
      "aws/*_aws_config_*",
      "website/**/config_*"
    ]
    "service/connect" = [
      "aws/internal/service/connect/**/*",
      "aws/*_aws_connect_*",
      "website/**/connect_*"
    ]
    "service/costandusagereportservice" = [
      "aws/internal/service/costandusagereportservice/**/*",
      "aws/*_aws_cur_*",
      "website/**/cur_*"
    ]
    "service/databasemigrationservice" = [
      "aws/internal/service/databasemigrationservice/**/*",
      "**/*_dms_*",
      "**/dms_*"
    ]
    "service/dataexchange" = [
      "aws/internal/service/dataexchange/**/*",
      "**/*_dataexchange_*",
      "**/dataexchange_*",
    ]
    "service/datapipeline" = [
      "aws/internal/service/datapipeline/**/*",
      "**/*_datapipeline_*",
      "**/datapipeline_*",
    ]
    "service/datasync" = [
      "aws/internal/service/datasync/**/*",
      "**/*_datasync_*",
      "**/datasync_*",
    ]
    "service/dax" = [
      "aws/internal/service/dax/**/*",
      "**/*_dax_*",
      "**/dax_*"
    ]
    "service/devicefarm" = [
      "aws/internal/service/devicefarm/**/*",
      "**/*_devicefarm_*",
      "**/devicefarm_*"
    ]
    "service/directconnect" = [
      "aws/internal/service/directconnect/**/*",
      "**/*_dx_*",
      "**/dx_*"
    ]
    "service/directoryservice" = [
      "aws/internal/service/directoryservice/**/*",
      "**/*_directory_service_*",
      "**/directory_service_*"
    ]
    "service/dlm" = [
      "aws/internal/service/dlm/**/*",
      "**/*_dlm_*",
      "**/dlm_*"
    ]
    "service/docdb" = [
      "aws/internal/service/docdb/**/*",
      "**/*_docdb_*",
      "**/docdb_*"
    ]
    "service/dynamodb" = [
      "aws/internal/service/dynamodb/**/*",
      "**/*_dynamodb_*",
      "**/dynamodb_*"
    ]
    # Special casing this one because the files aren't _ec2_
    "service/ec2" = [
      "aws/internal/service/ec2/**/*",
      "**/*_ec2_*",
      "**/ec2_*",
      "aws/*_aws_ami*",
      "aws/*_aws_availability_zone*",
      "aws/*_aws_customer_gateway*",
      "aws/*_aws_default_network_acl*",
      "aws/*_aws_default_route_table*",
      "aws/*_aws_default_security_group*",
      "aws/*_aws_default_subnet*",
      "aws/*_aws_default_vpc*",
      "aws/*_aws_ebs_*",
      "aws/*_aws_egress_only_internet_gateway*",
      "aws/*_aws_eip*",
      "aws/*_aws_flow_log*",
      "aws/*_aws_instance*",
      "aws/*_aws_internet_gateway*",
      "aws/*_aws_key_pair*",
      "aws/*_aws_launch_template*",
      "aws/*_aws_main_route_table_association*",
      "aws/*_aws_nat_gateway*",
      "aws/*_aws_network_acl*",
      "aws/*_aws_network_interface*",
      "aws/*_aws_placement_group*",
      "aws/*_aws_prefix_list*",
      "aws/*_aws_route_table*",
      "aws/*_aws_route.*",
      "aws/*_aws_security_group*",
      "aws/*_aws_snapshot_create_volume_permission*",
      "aws/*_aws_spot*",
      "aws/*_aws_subnet*",
      "aws/*_aws_vpc*",
      "aws/*_aws_vpn*",
      "aws/*_aws_volume_attachment*",
      "website/**/availability_zone*",
      "website/**/customer_gateway*",
      "website/**/default_network_acl*",
      "website/**/default_route_table*",
      "website/**/default_security_group*",
      "website/**/default_subnet*",
      "website/**/default_vpc*",
      "website/**/ebs_*",
      "website/**/egress_only_internet_gateway*",
      "website/**/eip*",
      "website/**/flow_log*",
      "website/**/instance*",
      "website/**/internet_gateway*",
      "website/**/key_pair*",
      "website/**/launch_template*",
      "website/**/main_route_table_association*",
      "website/**/nat_gateway*",
      "website/**/network_acl*",
      "website/**/network_interface*",
      "website/**/placement_group*",
      "website/**/prefix_list*",
      "website/**/route_table*",
      "website/**/route.*",
      "website/**/security_group*",
      "website/**/snapshot_create_volume_permission*",
      "website/**/spot_*",
      "website/**/subnet*",
      "website/**/vpc*",
      "website/**/vpn*",
      "website/**/volume_attachment*"
    ]
    "service/ecr" = [
      "aws/internal/service/ecr/**/*",
      "**/*_ecr_*",
      "**/ecr_*"
    ]
    "service/ecrpublic" = [
      "aws/internal/service/ecrpublic/**/*",
      "**/*_ecrpublic_*",
      "**/ecrpublic_*"
    ]
    "service/ecs" = [
      "aws/internal/service/ecs/**/*",
      "**/*_ecs_*",
      "**/ecs_*"
    ]
    "service/efs" = [
      "aws/internal/service/efs/**/*",
      "**/*_efs_*",
      "**/efs_*"
    ]
    "service/eks" = [
      "aws/internal/service/eks/**/*",
      "**/*_eks_*",
      "**/eks_*"
    ]
    "service/elastic-transcoder" = [
      "aws/internal/service/elastictranscoder/**/*",
      "**/*_elastictranscoder_*",
      "**/elastictranscoder_*",
      "**/*_elastic_transcoder_*",
      "**/elastic_transcoder_*"
    ]
    "service/elasticache" = [
      "aws/internal/service/elasticache/**/*",
      "**/*_elasticache_*",
      "**/elasticache_*"
    ]
    "service/elasticbeanstalk" = [
      "aws/internal/service/elasticbeanstalk/**/*",
      "**/*_elastic_beanstalk_*",
      "**/elastic_beanstalk_*"
    ]
    "service/elasticsearch" = [
      "aws/internal/service/elasticsearchservice/**/*",
      "**/*_elasticsearch_*",
      "**/elasticsearch_*",
      "**/*_elasticsearchservice*"
    ]
    "service/elb" = [
      "aws/internal/service/elb/**/*",
      "aws/*_aws_app_cookie_stickiness_policy*",
      "aws/*_aws_elb*",
      "aws/*_aws_lb_cookie_stickiness_policy*",
      "aws/*_aws_lb_ssl_negotiation_policy*",
      "aws/*_aws_load_balancer*",
      "aws/*_aws_proxy_protocol_policy*",
      "website/**/app_cookie_stickiness_policy*",
      "website/**/elb*",
      "website/**/lb_cookie_stickiness_policy*",
      "website/**/lb_ssl_negotiation_policy*",
      "website/**/load_balancer*",
      "website/**/proxy_protocol_policy*"
    ]
    "service/elbv2" = [
      "aws/internal/service/elbv2/**/*",
      "aws/*_lb.*",
      "aws/*_lb_listener*",
      "aws/*_lb_target_group*",
      "website/**/lb.*",
      "website/**/lb_listener*",
      "website/**/lb_target_group*"
    ]
    "service/emr" = [
      "aws/internal/service/emr/**/*",
      "**/*_emr_*",
      "**/emr_*"
    ]
    "service/eventbridge" = [
      # EventBridge is rebranded CloudWatch Events
      "aws/internal/service/cloudwatchevents/**/*",
      "**/*_cloudwatch_event_*",
      "**/cloudwatch_event_*"
    ]
    "service/firehose" = [
      "aws/internal/service/firehose/**/*",
      "**/*_firehose_*",
      "**/firehose_*"
    ]
    "service/fms" = [
      "aws/internal/service/fms/**/*",
      "**/*_fms_*",
      "**/fms_*"
    ]
    "service/fsx" = [
      "aws/internal/service/fsx/**/*",
      "**/*_fsx_*",
      "**/fsx_*"
    ]
    "service/gamelift" = [
      "aws/internal/service/gamelift/**/*",
      "**/*_gamelift_*",
      "**/gamelift_*"
    ]
    "service/glacier" = [
      "aws/internal/service/glacier/**/*",
      "**/*_glacier_*",
      "**/glacier_*"
    ]
    "service/globalaccelerator" = [
      "aws/internal/service/globalaccelerator/**/*",
      "**/*_globalaccelerator_*",
      "**/globalaccelerator_*"
    ]
    "service/glue" = [
      "aws/internal/service/glue/**/*",
      "**/*_glue_*",
      "**/glue_*"
    ]
    "service/greengrass" = [
      "aws/internal/service/greengrass/**/*",
      "**/*_greengrass_*",
      "**/greengrass_*"
    ]
    "service/guardduty" = [
      "aws/internal/service/guardduty/**/*",
      "**/*_guardduty_*",
      "**/guardduty_*"
    ]
    "service/iam" = [
      "aws/internal/service/iam/**/*",
      "**/*_iam_*",
      "**/iam_*"
    ]
    "service/identitystore" = [
      "aws/internal/service/identitystore/**/*",
      "**/*_identitystore_*",
      "**/identitystore_*"
    ]
    "service/imagebuilder" = [
      "aws/internal/service/imagebuilder/**/*",
      "**/*_imagebuilder_*",
      "**/imagebuilder_*"
    ]
    "service/inspector" = [
      "aws/internal/service/inspector/**/*",
      "**/*_inspector_*",
      "**/inspector_*"
    ]
    "service/iot" = [
      "aws/internal/service/iot/**/*",
      "**/*_iot_*",
      "**/iot_*"
    ]
    "service/iotanalytics" = [
      "aws/internal/service/iotanalytics/**/*",
      "**/*_iotanalytics_*",
      "**/iotanalytics_*"
    ]
    "service/iotevents" = [
      "aws/internal/service/iotevents/**/*",
      "**/*_iotevents_*",
      "**/iotevents_*"
    ]
    "service/kafka" = [
      "aws/internal/service/kafka/**/*",
      "**/*_msk_*",
      "**/msk_*",
    ]
    "service/kinesis" = [
      "aws/internal/service/kinesis/**/*",
      "aws/*_aws_kinesis_stream*",
      "website/kinesis_stream*"
    ]
    "service/kinesisanalytics" = [
      "aws/internal/service/kinesisanalytics/**/*",
      "**/*_kinesis_analytics_*",
      "**/kinesis_analytics_*"
    ]
    "service/kinesisanalyticsv2" = [
      "aws/internal/service/kinesisanalyticsv2/**/*",
      "**/*_kinesisanalyticsv2_*",
      "**/kinesisanalyticsv2_*"
    ]
    "service/kms" = [
      "aws/internal/service/kms/**/*",
      "**/*_kms_*",
      "**/kms_*"
    ]
    "service/lakeformation" = [
      "aws/internal/service/lakeformation/**/*",
      "**/*_lakeformation_*",
      "**/lakeformation_*"
    ]
    "service/lambda" = [
      "aws/internal/service/lambda/**/*",
      "**/*_lambda_*",
      "**/lambda_*"
    ]
    "service/lexmodelbuildingservice" = [
      "aws/internal/service/lexmodelbuildingservice/**/*",
      "**/*_lex_*",
      "**/lex_*"
    ]
    "service/licensemanager" = [
      "aws/internal/service/licensemanager/**/*",
      "**/*_licensemanager_*",
      "**/licensemanager_*"
    ]
    "service/lightsail" = [
      "aws/internal/service/lightsail/**/*",
      "**/*_lightsail_*",
      "**/lightsail_*"
    ]
    "service/machinelearning" = [
      "aws/internal/service/machinelearning/**/*",
      "**/*_machinelearning_*",
      "**/machinelearning_*"
    ]
    "service/macie" = [
      "aws/internal/service/macie/**/*",
      "**/*_macie_*",
      "**/macie_*"
    ]
    "service/macie2" = [
      "aws/internal/service/macie2/**/*",
      "**/*_macie2_*",
      "**/macie2_*"
    ]
    "service/marketplacecatalog" = [
      "aws/internal/service/marketplacecatalog/**/*",
      "**/*_marketplace_catalog_*",
      "**/marketplace_catalog_*"
    ]
    "service/mediaconnect" = [
      "aws/internal/service/mediaconnect/**/*",
      "**/*_media_connect_*",
      "**/media_connect_*"
    ]
    "service/mediaconvert" = [
      "aws/internal/service/mediaconvert/**/*",
      "**/*_media_convert_*",
      "**/media_convert_*"
    ]
    "service/medialive" = [
      "aws/internal/service/medialive/**/*",
      "**/*_media_live_*",
      "**/media_live_*"
    ]
    "service/mediapackage" = [
      "aws/internal/service/mediapackage/**/*",
      "**/*_media_package_*",
      "**/media_package_*"
    ]
    "service/mediastore" = [
      "aws/internal/service/mediastore/**/*",
      "**/*_media_store_*",
      "**/media_store_*"
    ]
    "service/mediatailor" = [
      "aws/internal/service/mediatailor/**/*",
      "**/*_media_tailor_*",
      "**/media_tailor_*",
    ]
    "service/mobile" = [
      "aws/internal/service/mobile/**/*",
      "**/*_mobile_*",
      "**/mobile_*"
    ],
    "service/mq" = [
      "aws/internal/service/mq/**/*",
      "**/*_mq_*",
      "**/mq_*"
    ]
    "service/mwaa" = [
      "aws/internal/service/mwaa/**/*",
      "**/*_mwaa_*",
      "**/mwaa_*"
    ]
    "service/neptune" = [
      "aws/internal/service/neptune/**/*",
      "**/*_neptune_*",
      "**/neptune_*"
    ]
    "service/networkfirewall" = [
      "aws/internal/service/networkfirewall/**/*",
      "**/*_networkfirewall_*",
      "**/networkfirewall_*",
    ]
    "service/networkmanager" = [
      "aws/internal/service/networkmanager/**/*",
      "**/*_networkmanager_*",
      "**/networkmanager_*"
    ]
    "service/opsworks" = [
      "aws/internal/service/opsworks/**/*",
      "**/*_opsworks_*",
      "**/opsworks_*"
    ]
    "service/organizations" = [
      "aws/internal/service/organizations/**/*",
      "**/*_organizations_*",
      "**/organizations_*"
    ]
    "service/outposts" = [
      "aws/internal/service/outposts/**/*",
      "**/*_outposts_*",
      "**/outposts_*"
    ]
    "service/pinpoint" = [
      "aws/internal/service/pinpoint/**/*",
      "**/*_pinpoint_*",
      "**/pinpoint_*"
    ]
    "service/polly" = [
      "aws/internal/service/polly/**/*",
      "**/*_polly_*",
      "**/polly_*"
    ]
    "service/pricing" = [
      "aws/internal/service/pricing/**/*",
      "**/*_pricing_*",
      "**/pricing_*"
    ]
    "service/qldb" = [
      "aws/internal/service/qldb/**/*",
      "**/*_qldb_*",
      "**/qldb_*"
    ]
    "service/quicksight" = [
      "aws/internal/service/quicksight/**/*",
      "**/*_quicksight_*",
      "**/quicksight_*"
    ]
    "service/ram" = [
      "aws/internal/service/ram/**/*",
      "**/*_ram_*",
      "**/ram_*"
    ]
    "service/rds" = [
      "aws/internal/service/rds/**/*",
      "aws/*_aws_db_*",
      "aws/*_aws_rds_*",
      "website/**/db_*",
      "website/**/rds_*"
    ]
    "service/redshift" = [
      "aws/internal/service/redshift/**/*",
      "**/*_redshift_*",
      "**/redshift_*"
    ]
    "service/resourcegroups" = [
      "aws/internal/service/resourcegroups/**/*",
      "**/*_resourcegroups_*",
      "**/resourcegroups_*"
    ]
    "service/resourcegroupstaggingapi" = [
      "aws/internal/service/resourcegroupstaggingapi/**/*",
      "**/*_resourcegroupstaggingapi_*",
      "**/resourcegroupstaggingapi_*"
    ]
    "service/robomaker" = [
      "aws/internal/service/robomaker/**/*",
      "**/*_robomaker_*",
      "**/robomaker_*",
    ]
    "service/route53" = [
      "aws/internal/service/route53/**/*",
      "**/*_route53_delegation_set*",
      "**/*_route53_health_check*",
      "**/*_route53_query_log*",
      "**/*_route53_record*",
      "**/*_route53_vpc_association_authorization*",
      "**/*_route53_zone*",
      "**/route53_delegation_set*",
      "**/route53_health_check*",
      "**/route53_query_log*",
      "**/route53_record*",
      "**/route53_vpc_association_authorization*",
      "**/route53_zone*"
    ]
    "service/route53domains" = [
      "aws/internal/service/route53domains/**/*",
      "**/*_route53domains_*",
      "**/route53domains_*"
    ]
    "service/route53resolver" = [
      "aws/internal/service/route53resolver/**/*",
      "**/*_route53_resolver_*",
      "**/route53_resolver_*"
    ]
    "service/s3" = [
      "aws/internal/service/s3/**/*",
      "**/*_s3_bucket*",
      "**/s3_bucket*",
      "aws/*_aws_canonical_user_id*",
      "website/**/canonical_user_id*"
    ]
    "service/s3control" = [
      "aws/internal/service/s3control/**/*",
      "**/*_s3_account_*",
      "**/s3_account_*",
      "**/*_s3control_*",
      "**/s3control_*"
    ]
    "service/s3outposts" = [
      "aws/internal/service/s3outposts/**/*",
      "**/*_s3outposts_*",
      "**/s3outposts_*"
    ]
    "service/sagemaker" = [
      "aws/internal/service/sagemaker/**/*",
      "**/*_sagemaker_*",
      "**/sagemaker_*"
    ]
    "service/secretsmanager" = [
      "aws/internal/service/secretsmanager/**/*",
      "**/*_secretsmanager_*",
      "**/secretsmanager_*"
    ]
    "service/securityhub" = [
      "aws/internal/service/securityhub/**/*",
      "**/*_securityhub_*",
      "**/securityhub_*"
    ]
    "service/serverlessapplicationrepository" = [
      "aws/internal/service/serverlessapplicationrepository/**/*",
      "**/*_serverlessapplicationrepository_*",
      "**/serverlessapplicationrepository_*"
    ]
    "service/servicecatalog" = [
      "aws/internal/service/servicecatalog/**/*",
      "**/*_servicecatalog_*",
      "**/servicecatalog_*"
    ]
    "service/servicediscovery" = [
      "aws/internal/service/servicediscovery/**/*",
      "**/*_service_discovery_*",
      "**/service_discovery_*"
    ]
    "service/servicequotas" = [
      "aws/internal/service/servicequotas/**/*",
      "**/*_servicequotas_*",
      "**/servicequotas_*"
    ]
    "service/ses" = [
      "aws/internal/service/ses/**/*",
      "**/*_ses_*",
      "**/ses_*"
    ]
    "service/sfn" = [
      "aws/internal/service/sfn/**/*",
      "**/*_sfn_*",
      "**/sfn_*"
    ]
    "service/shield" = [
      "aws/internal/service/shield/**/*",
      "**/*_shield_*",
      "**/shield_*",
    ],
    "service/signer" = [
      "**/*_signer_*",
      "**/signer_*"
    ]
    "service/simpledb" = [
      "aws/internal/service/simpledb/**/*",
      "**/*_simpledb_*",
      "**/simpledb_*"
    ]
    "service/snowball" = [
      "aws/internal/service/snowball/**/*",
      "**/*_snowball_*",
      "**/snowball_*"
    ]
    "service/sns" = [
      "aws/internal/service/sns/**/*",
      "**/*_sns_*",
      "**/sns_*"
    ]
    "service/sqs" = [
      "aws/internal/service/sqs/**/*",
      "**/*_sqs_*",
      "**/sqs_*"
    ]
    "service/ssm" = [
      "aws/internal/service/ssm/**/*",
      "**/*_ssm_*",
      "**/ssm_*"
    ]
    "service/ssoadmin" = [
      "aws/internal/service/ssoadmin/**/*",
      "**/*_ssoadmin_*",
      "**/ssoadmin_*"
    ]
    "service/storagegateway" = [
      "aws/internal/service/storagegateway/**/*",
      "**/*_storagegateway_*",
      "**/storagegateway_*"
    ]
    "service/sts" = [
      "aws/internal/service/sts/**/*",
      "aws/*_aws_caller_identity*",
      "website/**/caller_identity*"
    ]
    "service/swf" = [
      "aws/internal/service/swf/**/*",
      "**/*_swf_*",
      "**/swf_*"
    ]
    "service/synthetics" = [
      "aws/internal/service/synthetics/**/*",
      "**/*_synthetics_*",
      "**/synthetics_*"
    ]
    "service/timestreamwrite" = [
      "aws/internal/service/timestreamwrite/**/*",
      "**/*_timestreamwrite_*",
      "**/timestreamwrite_*"
    ]
    "service/transfer" = [
      "aws/internal/service/transfer/**/*",
      "**/*_transfer_*",
      "**/transfer_*"
    ]
    "service/waf" = [
      "aws/internal/service/waf/**/*",
      "aws/internal/service/wafregional/**/*",
      "**/*_waf_*",
      "**/waf_*",
      "**/*_wafregional_*",
      "**/wafregional_*"
    ]
    "service/wafv2" = [
      "aws/internal/service/wafv2/**/*",
      "**/*_wafv2_*",
      "**/wafv2_*",
    ]
    "service/workdocs" = [
      "aws/internal/service/workdocs/**/*",
      "**/*_workdocs_*",
      "**/workdocs_*"
    ]
    "service/worklink" = [
      "aws/internal/service/worklink/**/*",
      "**/*_worklink_*",
      "**/worklink_*"
    ]
    "service/workmail" = [
      "aws/internal/service/workmail/**/*",
      "**/*_workmail_*",
      "**/workmail_*"
    ]
    "service/workspaces" = [
      "aws/internal/service/workspaces/**/*",
      "**/*_workspaces_*",
      "**/workspaces_*"
    ]
    "service/xray" = [
      "aws/internal/service/xray/**/*",
      "**/*_xray_*",
      "**/xray_*"
    ]
  }
}

behavior "regexp_issue_labeler" "panic_label" {
    regexp = "panic:"
    labels = ["crash", "bug"]
}

behavior "remove_labels_on_reply" "remove_stale" {
    labels = ["waiting-response", "stale"]
    only_non_maintainers = true
}

behavior "pull_request_size_labeler" "size" {
    label_prefix = "size/"
    label_map = {
        "size/XS" = {
            from = 0
            to = 30
        }
        "size/S" = {
            from = 31
            to = 60
        }
        "size/M" = {
            from = 61
            to = 150
        }
        "size/L" = {
            from = 151
            to = 300
        }
        "size/XL" = {
            from = 301
            to = 1000
        }
        "size/XXL" = {
            from = 1001
            to = 0
        }
    }
}
