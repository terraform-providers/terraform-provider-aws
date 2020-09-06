#!/usr/bin/env bash

set -eo pipefail

# This script works from stdin and expects one filename per line.
# To call it, e.g.
# find ./website/docs -type f \( -name '*.md' -o -name '*.markdown' \) \
#   | ./scripts/validate-terraform.sh

TERRAFMT_CMD="terrafmt"
if [ -f ~/developer/terrafmt/terrafmt ]; then TERRAFMT_CMD="$HOME/developer/terrafmt/terrafmt"; fi

exit_code=0

# Configure the rules for tflint.
# The *_invalid_* rules disabled here prevent evaluation of expressions.
rules=(
    # Syntax checks
    "--enable-rule=terraform_deprecated_interpolation"
    "--enable-rule=terraform_deprecated_index"
    "--enable-rule=terraform_comment_syntax"
    # Ensure modern instance types
    "--enable-rule=aws_instance_previous_type"
    "--enable-rule=aws_db_instance_previous_type"
    "--enable-rule=aws_elasticache_cluster_previous_type"
    # Prevent some configuration errors
    "--enable-rule=aws_route_specified_multiple_targets"
    # Prevent expression evaluation
    "--disable-rule=aws_acm_certificate_invalid_certificate_body"
    "--disable-rule=aws_acm_certificate_invalid_certificate_chain"
    "--disable-rule=aws_acm_certificate_invalid_private_key"
    "--disable-rule=aws_acmpca_certificate_authority_invalid_type"
    "--disable-rule=aws_appsync_datasource_invalid_name"
    "--disable-rule=aws_appsync_function_invalid_name"
    "--disable-rule=aws_appsync_graphql_api_invalid_authentication_type"
    "--disable-rule=aws_athena_workgroup_invalid_name"
    "--disable-rule=aws_athena_workgroup_invalid_state"
    "--disable-rule=aws_backup_selection_invalid_name"
    "--disable-rule=aws_backup_vault_invalid_name"
    "--disable-rule=aws_cloudwatch_log_group_invalid_name"
    "--disable-rule=aws_codecommit_repository_invalid_repository_name"
    "--disable-rule=aws_cognito_user_pool_invalid_name"
    "--disable-rule=aws_cur_report_definition_invalid_report_name"
    "--disable-rule=aws_db_instance_default_parameter_group"
    "--disable-rule=aws_dynamodb_table_invalid_name"
    "--disable-rule=aws_ecr_repository_invalid_name"
    "--disable-rule=aws_elasticsearch_domain_invalid_domain_name"
    "--disable-rule=aws_iam_group_invalid_name"
    "--disable-rule=aws_iam_instance_profile_invalid_name"
    "--disable-rule=aws_iam_policy_invalid_name"
    "--disable-rule=aws_iam_role_invalid_name"
    "--disable-rule=aws_iam_role_policy_invalid_name"
    "--disable-rule=aws_iam_server_certificate_invalid_name"
    "--disable-rule=aws_iam_server_certificate_invalid_path"
    "--disable-rule=aws_iam_user_invalid_name"
    "--disable-rule=aws_iam_user_invalid_path"
    "--disable-rule=aws_iam_user_invalid_permissions_boundary"
    "--disable-rule=aws_kinesis_firehose_delivery_stream_invalid_name"
    "--disable-rule=aws_kinesis_stream_invalid_name"
    "--disable-rule=aws_kms_alias_invalid_name"
    "--disable-rule=aws_lambda_function_invalid_function_name"
    "--disable-rule=aws_lambda_layer_version_invalid_layer_name"
    "--disable-rule=aws_launch_template_invalid_name"
    "--disable-rule=aws_lb_target_group_invalid_protocol"
    "--disable-rule=aws_lb_target_group_invalid_target_type"
    "--disable-rule=aws_lightsail_key_pair_invalid_name"
    "--disable-rule=aws_ssm_document_invalid_name"
    "--disable-rule=aws_ssm_patch_baseline_invalid_name"
)
while read -r filename ; do
    echo "$filename"
    block_number=0

    while IFS= read -r block ; do
        ((block_number+=1))
        start_line=$(echo "$block" | jq '.start_line')
        end_line=$(echo "$block" | jq '.end_line')
        text=$(echo "$block" | jq --raw-output '.text')

        td=$(mktemp -d)
        tf="$td/main.tf"

        echo "$text" > "$tf"

        # We need to capture the output and error code here. We don't want to exit on the first error
        set +e
        tflint_output=$(tflint "${rules[@]}" "$tf" 2>&1)
        tflint_exitcode=$?
        set -e

        if [ $tflint_exitcode -ne 0 ]; then
            echo "ERROR: File \"$filename\", block #$block_number (lines $start_line-$end_line):"
            echo "$tflint_output"
            echo
            exit_code=1
            exit $exit_code
        fi
    done < <( $TERRAFMT_CMD blocks --fmtcompat --json "$filename" | jq --compact-output '.blocks[]?' )
done

exit $exit_code
