package aws

import (
	"testing"
)

func TestAccAWSFms_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			"basic": testAccAwsFmsAdminAccount_basic,
		},
		"Policy": {
			"basic":                  testAccAWSFmsPolicy_basic,
			"cloudfrontDistribution": testAccAWSFmsPolicy_cloudfrontDistribution,
			"includeMap":             testAccAWSFmsPolicy_includeMap,
			"update":                 testAccAWSFmsPolicy_update,
			"tags":                   testAccAWSFmsPolicy_tags,
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
