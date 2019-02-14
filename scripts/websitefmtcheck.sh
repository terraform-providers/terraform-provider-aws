#!/bin/bash

set -eou pipefail

npm list codedown > /dev/null 2>&1 || npm install --no-save codedown > /dev/null 2>&1

problems=false
for f in $(find website -name '*.markdown'); do
    if [ "${1-}" = "diff" ]; then
        echo "$f"
        cat "$f" | node_modules/.bin/codedown hcl | terraform fmt -diff=true -
    else
        cat "$f" | node_modules/.bin/codedown hcl | terraform fmt -check=true - || problems=true && echo "Formatting errors in $f"
    fi
done

if [ "$problems" = true ] ; then
    exit 1
fi
