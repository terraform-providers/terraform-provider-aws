package aws

import (
	"testing"
)

func TestAccAWSSecurityHub(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic": testAccAWSSecurityHubAccount_basic,
		},
		"ProductSubscription": {
			"basic": testAccAWSSecurityHubProductSubscription_basic,
		},
		"StandardsSubscription": {
			"basic": testAccAWSSecurityHubStandardsSubscription_basic,
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
