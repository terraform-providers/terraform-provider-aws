package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAWSAppmeshVirtualNodeMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1-noBackendsOrDns": {
			StateVersion: 0,
			Attributes: map[string]string{
				"spec.0.backends.#":          "0",
				"spec.0.service_discovery.#": "0",
			},
			Expected: map[string]string{
				"spec.0.backend.#":           "0",
				"spec.0.service_discovery.#": "0",
			},
		},
		"v0_1-withBackendAndDns": {
			StateVersion: 0,
			Attributes: map[string]string{
				"spec.0.backends.#":                             "1",
				"spec.0.backends.1255689679":                    "servicea.simpleapp.local",
				"spec.0.service_discovery.#":                    "1",
				"spec.0.service_discovery.0.dns.#":              "1",
				"spec.0.service_discovery.0.dns.0.service_name": "serviceb.simpleapp.local",
			},
			Expected: map[string]string{
				"spec.0.backend.#":                                        "1",
				"spec.0.backend.0.virtual_service.#":                      "1",
				"spec.0.backend.0.virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
				"spec.0.service_discovery.#":                              "1",
				"spec.0.service_discovery.0.dns.#":                        "1",
				"spec.0.service_discovery.0.dns.0.hostname":               "serviceb.simpleapp.local",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: tc.Attributes,
		}

		is, err := resourceAwsAppmeshVirtualNode().MigrateState(tc.StateVersion, is, tc.Meta)
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
