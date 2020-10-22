package glue

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ReadAwsGluePartitionID(id string) (catalogID string, dbName string, tableName string, values []string, error error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 4 {
		return "", "", "", []string{}, fmt.Errorf("expected ID in format catalog-id:database-name:table-name:values, received: %s", id)
	}
	vals := strings.Split(idParts[3], "#")
	return idParts[0], idParts[1], idParts[2], vals, nil
}

func CreateAwsGluePartitionID(catalogID, dbName, tableName string, values *schema.Set) string {
	return fmt.Sprintf("%s:%s:%s:%s", catalogID, dbName, tableName, stringifyAwsGluePartition(values))
}

func stringifyAwsGluePartition(partValues *schema.Set) string {
	var b bytes.Buffer
	for _, val := range partValues.List() {
		b.WriteString(fmt.Sprintf("%s#", val.(string)))
	}
	vals := strings.Trim(b.String(), "#")

	return vals
}
