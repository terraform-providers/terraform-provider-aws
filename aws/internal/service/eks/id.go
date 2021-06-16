package eks

import (
	"fmt"
	"strings"
)

const nodeGroupResourceIDSeparator = ":"

func NodeGroupCreateResourceID(clusterName, nodeGroupName string) string {
	parts := []string{clusterName, nodeGroupName}
	id := strings.Join(parts, nodeGroupResourceIDSeparator)

	return id
}

func NodeGroupParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, nodeGroupResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]snode-group-name", id, nodeGroupResourceIDSeparator)
}
