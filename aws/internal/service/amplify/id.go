package amplify

import (
	"fmt"
	"strings"
)

const backendEnvironmentResourceIDSeparator = "/"

func BackendEnvironmentCreateResourceID(appID, environmentName string) string {
	parts := []string{appID, environmentName}
	id := strings.Join(parts, backendEnvironmentResourceIDSeparator)

	return id
}

func BackendEnvironmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, backendEnvironmentResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPID%[2]sENVIRONMENTNAME", id, backendEnvironmentResourceIDSeparator)
}

const branchResourceIDSeparator = "/"

func BranchCreateResourceID(appID, branchName string) string {
	parts := []string{appID, branchName}
	id := strings.Join(parts, branchResourceIDSeparator)

	return id
}

func BranchParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, branchResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPID%[2]sBRANCHNAME", id, branchResourceIDSeparator)
}

const domainAssociationResourceIDSeparator = "/"

func DomainAssociationCreateResourceID(appID, domainName string) string {
	parts := []string{appID, domainName}
	id := strings.Join(parts, domainAssociationResourceIDSeparator)

	return id
}

func DomainAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, domainAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPID%[2]sDOMAINNAME", id, domainAssociationResourceIDSeparator)
}
