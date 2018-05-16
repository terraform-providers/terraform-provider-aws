package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAwsElasticacheReplicationGroupMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0Tov1": {
			StateVersion: 0,
			Attributes: map[string]string{
				"cluster_mode.#":                                  "1",
				"cluster_mode.4170186206.num_node_groups":         "2",
				"cluster_mode.4170186206.replicas_per_node_group": "1",
			},
			Expected: map[string]string{
				"cluster_mode.#":                         "1",
				"cluster_mode.0.num_node_groups":         "2",
				"cluster_mode.0.replicas_per_node_group": "1",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "rg-migration",
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsElasticacheReplicationGroupMigrateState(tc.StateVersion, is, tc.Meta)

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

func TestAwsElasticacheReplicationGroupMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceAwsElasticacheReplicationGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	is, err = resourceAwsElasticacheReplicationGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
