package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAWSAppmeshVirtualRouterMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1-emptySpec": {
			StateVersion: 0,
			Attributes: map[string]string{
				"name":   "svcb",
				"spec.#": "1",
			},
			Expected: map[string]string{
				"name":   "svcb",
				"spec.#": "1",
			},
		},
		"v0_1-nonEmptySpec": {
			StateVersion: 0,
			Attributes: map[string]string{
				"name":                           "svcb",
				"spec.#":                         "1",
				"spec.0.service_names.#":         "1",
				"spec.0.service_names.423761483": "serviceb.simpleapp.local",
			},
			Expected: map[string]string{
				"name":   "svcb",
				"spec.#": "1",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: tc.Attributes,
		}

		is, err := resourceAwsAppmeshVirtualRouter().MigrateState(tc.StateVersion, is, tc.Meta)
		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}
