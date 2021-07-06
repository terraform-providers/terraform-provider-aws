package eks

import (
	"fmt"
	"strings"
)

const addonResourceIDSeparator = ":"

func AddonCreateResourceID(clusterName, addonName string) string {
	parts := []string{clusterName, addonName}
	id := strings.Join(parts, addonResourceIDSeparator)

	return id
}

func AddonParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, addonResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]saddon-name", id, addonResourceIDSeparator)
}

const fargateProfileResourceIDSeparator = ":"

func FargateProfileCreateResourceID(clusterName, fargateProfileName string) string {
	parts := []string{clusterName, fargateProfileName}
	id := strings.Join(parts, fargateProfileResourceIDSeparator)

	return id
}

func FargateProfileParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, fargateProfileResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cluster-name%[2]sfargate-profile-name", id, fargateProfileResourceIDSeparator)
}

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
