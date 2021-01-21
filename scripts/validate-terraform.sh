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
rules=(
    # Syntax checks
    "--only=terraform_deprecated_interpolation"
    "--only=terraform_deprecated_index"
    "--only=terraform_comment_syntax"
    # Ensure valid instance types
    "--only=aws_db_instance_invalid_type"
    "--only=aws_elasticache_cluster_invalid_type"
    # Ensure modern instance types
    "--only=aws_instance_previous_type"
    "--only=aws_db_instance_previous_type"
    "--only=aws_elasticache_cluster_previous_type"
    # Prevent some configuration errors
    "--only=aws_route_specified_multiple_targets"
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
        fi
    done < <( $TERRAFMT_CMD blocks --fmtcompat --json "$filename" | jq --compact-output '.blocks[]?' )
done

exit $exit_code
