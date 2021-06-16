package aws

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func testResourceAwsSsmPatchGroupStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":          "testgroup",
		"baseline_id": "pb-0c4e592064EXAMPLE",
		"patch_group": "testgroup",
	}
}

func testResourceAwsSsmPatchGroupStateDataV1() map[string]interface{} {
	v0 := testResourceAwsSsmPatchGroupStateDataV0()
	return map[string]interface{}{
		"id":          fmt.Sprintf("%s,%s", v0["patch_group"], v0["baseline_id"]),
		"baseline_id": v0["baseline_id"],
		"patch_group": v0["patch_group"],
	}
}

func TestResourceAWSSSMPatchGroupStateUpgradeV0(t *testing.T) {
	expected := testResourceAwsSsmPatchGroupStateDataV1()
	actual, err := resourceAwsSsmPatchGroupStateUpgradeV0(context.Background(), testResourceAwsSsmPatchGroupStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
