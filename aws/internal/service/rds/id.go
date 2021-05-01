package rds

import (
	"fmt"
	"strings"
)

func ResourceAwsDbProxyEndpointParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/db_proxy_endpoint_name", id)
	}
	return idParts[0], idParts[1], nil
}
