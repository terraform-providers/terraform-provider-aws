#!/usr/bin/env bash

set -eo pipefail

# This script accepts the filename and an array of options for tflint.
# To call it, e.g.
# rules=(
#   "--enable-rule=terraform_deprecated_interpolation"
#   "--enable-rule=terraform_deprecated_index"
# )
# ./scripts/validate-terraform-file.sh "$filename"  "${rules[@]}"

filename=$1
shift
rules=( "$@" )

exit_code=0

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
done < <( terrafmt blocks --json "$filename" | jq --compact-output '.blocks[]?' )

exit $exit_code
