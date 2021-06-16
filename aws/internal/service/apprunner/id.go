package apprunner

import (
	"fmt"
	"strings"
)

func CustomDomainAssociationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected domain_name,service_arn", id)
	}
	return idParts[0], idParts[1], nil
}
