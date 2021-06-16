package aws

import (
	"testing"
	"time"
)

// Serialize to limit API rate-limit exceeded errors.
func TestAccAwsCur_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ReportDefinition": {
			"basic":      testAccAwsCurReportDefinition_basic,
			"disappears": testAccAwsCurReportDefinition_disappears,
			"textOrCsv":  testAccAwsCurReportDefinition_textOrCsv,
			"parquet":    testAccAwsCurReportDefinition_parquet,
			"athena":     testAccAwsCurReportDefinition_athena,
			"refresh":    testAccAwsCurReportDefinition_refresh,
			"overwrite":  testAccAwsCurReportDefinition_overwrite,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					// Explicitly sleep between tests.
					time.Sleep(5 * time.Second)
				})
			}
		})
	}
}
