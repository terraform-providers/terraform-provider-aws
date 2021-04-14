#!/bin/bash

set -o errexit
set -o nounset

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__parent="$(dirname "$__dir")"

CHANGELOG_FILE_NAME="CHANGELOG.md"
CHANGELOG_TMP_FILE_NAME="CHANGELOG.tmp"
TARGET_SHA=$(git rev-parse HEAD)
PREVIOUS_RELEASE_TAG=$(git describe --abbrev=0 --match='v*.*.*' --tags)
PREVIOUS_RELEASE_SHA=$(git rev-list -n 1 $PREVIOUS_RELEASE_TAG)

if [ $TARGET_SHA == $PREVIOUS_RELEASE_SHA ]; then
  echo "Nothing to do"
  exit 0
fi

PREVIOUS_CHANGELOG=$(sed -n -e "/# ${PREVIOUS_RELEASE_TAG#v}/,\$p" $__parent/$CHANGELOG_FILE_NAME)

if [ -z "$PREVIOUS_CHANGELOG" ]
then
    echo "Unable to locate previous changelog contents."
    exit 1
fi 

CHANGELOG=$($(go env GOPATH)/bin/changelog-build -this-release $TARGET_SHA \
                      -last-release $PREVIOUS_RELEASE_SHA \
                      -git-dir $__parent \
                      -entries-dir .changelog \
                      -changelog-template $__dir/changelog.tmpl \
                      -note-template $__dir/release-note.tmpl \
                      -storage-mode filesystem)
if [ -z "$CHANGELOG" ]
then
    echo "No changelog generated."
    exit 0
fi

rm -f $CHANGELOG_TMP_FILE_NAME

sed -n -e "1{/# /p;}" $__parent/$CHANGELOG_FILE_NAME > $CHANGELOG_TMP_FILE_NAME
echo "$CHANGELOG" >> $CHANGELOG_TMP_FILE_NAME
echo >> $CHANGELOG_TMP_FILE_NAME
echo "$PREVIOUS_CHANGELOG" >> $CHANGELOG_TMP_FILE_NAME

cp $CHANGELOG_TMP_FILE_NAME $CHANGELOG_FILE_NAME

rm $CHANGELOG_TMP_FILE_NAME

echo "Successfully generated changelog."

exit 0