package kinesisanalyticsv2

import (
	"fmt"
	"strings"
)

const applicationSnapshotIDSeparator = "/"

func ApplicationSnapshotCreateID(applicationName, snapshotName string) string {
	parts := []string{applicationName, snapshotName}
	id := strings.Join(parts, applicationSnapshotIDSeparator)

	return id
}

func ApplicationSnapshotParseID(id string) (string, string, error) {
	parts := strings.Split(id, applicationSnapshotIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%q), expected application-name%ssnapshot-name", id, applicationSnapshotIDSeparator)
}
