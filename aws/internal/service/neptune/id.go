package glue

import (
	"fmt"
	"strings"
)

func ReadAwsNeptuneClusterEndpointId(id string) (clusterIdentifier string, endpointIndetifer string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format clusterIdentifier:endpointIndetifer, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
