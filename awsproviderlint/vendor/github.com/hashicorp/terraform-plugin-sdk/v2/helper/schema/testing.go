package schema

import (
	"context"

	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestResourceDataRaw creates a ResourceData from a raw configuration map.
func TestResourceDataRaw(t testing.T, schema map[string]*Schema, raw map[string]interface{}) *ResourceData {
	t.Helper()

	c := terraform.NewResourceConfigRaw(raw)

	sm := schemaMap(schema)
	diff, err := sm.Diff(context.Background(), nil, c, nil, nil, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := sm.Data(nil, diff)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return result
}
