package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceAwsAppmeshVirtualRouterMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found App Mesh virtual router state v0; migrating to v1")
		return migrateAppmeshVirtualRouterStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateAppmeshVirtualRouterStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty App Mesh virtual router state; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)
	// Remove 'spec' attribute.
	for k := range is.Attributes {
		if strings.HasPrefix(k, "spec.") {
			delete(is.Attributes, k)
		}
	}
	// Add back and empty 'spec'.
	is.Attributes["spec.#"] = "1"

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
