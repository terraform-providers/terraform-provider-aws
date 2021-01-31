package aws

import "testing"

func TestAccAWSGlue_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ResourcePolicy": {
			"basic":      testAccAWSGlueResourcePolicy_basic,
			"update":     testAccAWSGlueResourcePolicy_update,
			"disappears": testAccAWSGlueResourcePolicy_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
