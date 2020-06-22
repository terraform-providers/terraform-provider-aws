package tfawsresource

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestTestCheckTypeSetElemAttr(t *testing.T) {
	testCases := []struct {
		Description       string
		ResourceAddress   string
		ResourceAttribute string
		Value             string
		TerraformState    *terraform.State
		ExpectedError     func(err error) bool
	}{
		{
			Description:       "no resources",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:         []string{"root"},
						Outputs:      map[string]*terraform.OutputState{},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "Not found: example_thing.test")
			},
		},
		{
			Description:       "resource not found",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_other_thing.test": {
								Type:     "example_other_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "Not found: example_thing.test")
			},
		},
		{
			Description:       "no primary instance",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Deposed: []*terraform.InstanceState{
									{
										ID: "11111",
										Meta: map[string]interface{}{
											"schema_version": 0,
										},
										Attributes: map[string]string{
											"%":  "1",
											"id": "11111",
										},
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "No primary instance: example_thing.test")
			},
		},
		{
			Description:       "attribute path does not end with sentinel value",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test",
			Value:             "",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "does not end with the special value")
			},
		},
		{
			Description:       "attribute not found",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "single root TypeSet attribute match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "value1",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":          "3",
										"id":         "11111",
										"test.%":     "1",
										"test.12345": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single root TypeSet attribute mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "value2",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":          "3",
										"id":         "11111",
										"test.%":     "1",
										"test.12345": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "multiple root TypeSet attribute match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "value1",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":          "4",
										"id":         "11111",
										"test.%":     "2",
										"test.12345": "value2",
										"test.67890": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "multiple root TypeSet attribute mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Value:             "value3",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":          "4",
										"id":         "11111",
										"test.%":     "2",
										"test.12345": "value2",
										"test.67890": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "single nested TypeSet attribute match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.0.nested_test.*",
			Value:             "value1",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                        "4",
										"id":                       "11111",
										"test.%":                   "1",
										"test.0.nested_test.12345": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single nested TypeSet attribute mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.0.nested_test.*",
			Value:             "value2",
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                        "4",
										"id":                       "11111",
										"test.%":                   "1",
										"test.0.nested_test.12345": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.0.nested_test.*\"")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			err := TestCheckTypeSetElemAttr(testCase.ResourceAddress, testCase.ResourceAttribute, testCase.Value)(testCase.TerraformState)

			if err != nil {
				if testCase.ExpectedError == nil {
					t.Fatalf("expected no error, got error: %s", err)
				}

				if !testCase.ExpectedError(err) {
					t.Fatalf("unexpected error: %s", err)
				}

				t.Logf("received expected error: %s", err)
				return
			}

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error, got no error")
			}
		})
	}
}

func TestTestCheckTypeSetElemNestedAttrs(t *testing.T) {
	testCases := []struct {
		Description       string
		ResourceAddress   string
		ResourceAttribute string
		Values            map[string]string
		TerraformState    *terraform.State
		ExpectedError     func(err error) bool
	}{
		{
			Description:       "no resources",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values:            map[string]string{},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:         []string{"root"},
						Outputs:      map[string]*terraform.OutputState{},
						Resources:    map[string]*terraform.ResourceState{},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "Not found: example_thing.test")
			},
		},
		{
			Description:       "resource not found",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values:            map[string]string{},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_other_thing.test": {
								Type:     "example_other_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "Not found: example_thing.test")
			},
		},
		{
			Description:       "no primary instance",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values:            map[string]string{},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Deposed: []*terraform.InstanceState{
									{
										ID: "11111",
										Meta: map[string]interface{}{
											"schema_version": 0,
										},
										Attributes: map[string]string{
											"%":  "1",
											"id": "11111",
										},
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "No primary instance: example_thing.test")
			},
		},
		{
			Description:       "value map has no non-empty values",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values:            map[string]string{"key": ""},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "has no non-empty values")
			},
		},
		{
			Description:       "attribute path does not end with sentinel value",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test",
			Values:            map[string]string{},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "does not end with the special value")
			},
		},
		{
			Description:       "attribute not found",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values:            map[string]string{"key": "value"},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":  "1",
										"id": "11111",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "single root TypeSet attribute single value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value1",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "3",
										"id":              "11111",
										"test.%":          "1",
										"test.12345.key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single root TypeSet attribute single value mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value2",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "3",
										"id":              "11111",
										"test.%":          "1",
										"test.12345.key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "single root TypeSet attribute single nested value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1.0.nested_key1": "value1",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                             "3",
										"id":                            "11111",
										"test.%":                        "1",
										"test.12345.key1.0.nested_key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single root TypeSet attribute single nested value mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1.0.nested_key1": "value2",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                             "3",
										"id":                            "11111",
										"test.%":                        "1",
										"test.12345.key1.0.nested_key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "single root TypeSet attribute multiple value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "4",
										"id":              "11111",
										"test.%":          "1",
										"test.12345.key1": "value1",
										"test.12345.key2": "value2",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single root TypeSet attribute unset/empty value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value1",
				"key2": "",
				"key3": "",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "4",
										"id":              "11111",
										"test.%":          "1",
										"test.12345.key1": "value1",
										"test.12345.key2": "",
										// key3 is unset
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single root TypeSet attribute multiple value mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value1",
				"key2": "value3",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "4",
										"id":              "11111",
										"test.%":          "1",
										"test.12345.key1": "value1",
										"test.12345.key2": "value2",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.*\"")
			},
		},
		{
			Description:       "multiple root TypeSet attribute single value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value1",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "4",
										"id":              "11111",
										"test.%":          "2",
										"test.12345.key1": "value2",
										"test.67890.key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "multiple root TypeSet attribute multiple value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.*",
			Values: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":               "6",
										"id":              "11111",
										"test.%":          "2",
										"test.12345.key1": "value2",
										"test.12345.key2": "value3",
										"test.67890.key1": "value1",
										"test.67890.key2": "value2",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single nested TypeSet attribute single value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.0.nested_test.*",
			Values: map[string]string{
				"key1": "value1",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                             "4",
										"id":                            "11111",
										"test.%":                        "1",
										"test.0.nested_test.%":          "1",
										"test.0.nested_test.12345.key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single nested TypeSet attribute single value mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.0.nested_test.*",
			Values: map[string]string{
				"key1": "value2",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                             "4",
										"id":                            "11111",
										"test.%":                        "1",
										"test.0.nested_test.%":          "1",
										"test.0.nested_test.12345.key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.0.nested_test.*\"")
			},
		},
		{
			Description:       "single nested TypeSet attribute single nested value match",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.0.nested_test.*",
			Values: map[string]string{
				"key1.0.nested_key1": "value1",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                               "5",
										"id":                              "11111",
										"test.%":                          "1",
										"test.0.nested_test.%":            "1",
										"test.0.nested_test.12345.key1.%": "1",
										"test.0.nested_test.12345.key1.0.nested_key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
		},
		{
			Description:       "single nested TypeSet attribute single nested value mismatch",
			ResourceAddress:   "example_thing.test",
			ResourceAttribute: "test.0.nested_test.*",
			Values: map[string]string{
				"key1.0.nested_key1": "value2",
			},
			TerraformState: &terraform.State{
				Version: 3,
				Modules: []*terraform.ModuleState{
					{
						Path:    []string{"root"},
						Outputs: map[string]*terraform.OutputState{},
						Resources: map[string]*terraform.ResourceState{
							"example_thing.test": {
								Type:     "example_thing",
								Provider: "example",
								Primary: &terraform.InstanceState{
									ID: "11111",
									Meta: map[string]interface{}{
										"schema_version": 0,
									},
									Attributes: map[string]string{
										"%":                               "5",
										"id":                              "11111",
										"test.%":                          "1",
										"test.0.nested_test.%":            "1",
										"test.0.nested_test.12345.key1.%": "1",
										"test.0.nested_test.12345.key1.0.nested_key1": "value1",
									},
								},
							},
						},
						Dependencies: []string{},
					},
				},
			},
			ExpectedError: func(err error) bool {
				return strings.Contains(err.Error(), "\"example_thing.test\" no TypeSet element \"test.0.nested_test.*\"")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			err := TestCheckTypeSetElemNestedAttrs(testCase.ResourceAddress, testCase.ResourceAttribute, testCase.Values)(testCase.TerraformState)

			if err != nil {
				if testCase.ExpectedError == nil {
					t.Fatalf("expected no error, got error: %s", err)
				}

				if !testCase.ExpectedError(err) {
					t.Fatalf("unexpected error: %s", err)
				}

				t.Logf("received expected error: %s", err)
				return
			}

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error, got no error")
			}
		})
	}
}
