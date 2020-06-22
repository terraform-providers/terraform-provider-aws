package resource

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

func testStepNewImportState(t *testing.T, c TestCase, wd *tftest.WorkingDir, step TestStep, cfg string) error {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	if step.ResourceName == "" {
		t.Fatal("ResourceName is required for an import state test")
	}

	// get state from check sequence
	state := getState(t, wd)

	// Determine the ID to import
	var importId string
	switch {
	case step.ImportStateIdFunc != nil:
		var err error
		importId, err = step.ImportStateIdFunc(state)
		if err != nil {
			t.Fatal(err)
		}
	case step.ImportStateId != "":
		importId = step.ImportStateId
	default:
		resource, err := testResource(step, state)
		if err != nil {
			t.Fatal(err)
		}
		importId = resource.Primary.ID
	}
	importId = step.ImportStateIdPrefix + importId

	// Create working directory for import tests
	if step.Config == "" {
		step.Config = cfg
		if step.Config == "" {
			t.Fatal("Cannot import state with no specified config")
		}
	}
	importWd := acctest.TestHelper.RequireNewWorkingDir(t)
	defer importWd.Close()
	importWd.RequireSetConfig(t, step.Config)
	importWd.RequireInit(t)
	importWd.RequireImport(t, step.ResourceName, importId)
	importState := getState(t, wd)

	// Go through the imported state and verify
	if step.ImportStateCheck != nil {
		var states []*terraform.InstanceState
		for _, r := range importState.RootModule().Resources {
			if r.Primary != nil {
				is := r.Primary.DeepCopy()
				is.Ephemeral.Type = r.Type // otherwise the check function cannot see the type
				states = append(states, is)
			}
		}
		if err := step.ImportStateCheck(states); err != nil {
			t.Fatal(err)
		}
	}

	// Verify that all the states match
	if step.ImportStateVerify {
		new := importState.RootModule().Resources
		old := state.RootModule().Resources

		for _, r := range new {
			// Find the existing resource
			var oldR *terraform.ResourceState
			for _, r2 := range old {
				if r2.Primary != nil && r2.Primary.ID == r.Primary.ID && r2.Type == r.Type {
					oldR = r2
					break
				}
			}
			if oldR == nil {
				t.Fatalf(
					"Failed state verification, resource with ID %s not found",
					r.Primary.ID)
			}

			// We'll try our best to find the schema for this resource type
			// so we can ignore Removed fields during validation. If we fail
			// to find the schema then we won't ignore them and so the test
			// will need to rely on explicit ImportStateVerifyIgnore, though
			// this shouldn't happen in any reasonable case.
			// KEM CHANGE FROM OLD FRAMEWORK: Fail test if this happens.
			var rsrcSchema *schema.Resource
			providerAddr, diags := addrs.ParseAbsProviderConfigStr("provider." + r.Provider + "." + r.Type)
			if diags.HasErrors() {
				t.Fatalf("Failed to find schema for resource with ID %s", r.Primary)
			}

			providerType := providerAddr.ProviderConfig.Type
			if provider, ok := step.providers[providerType]; ok {
				if provider, ok := provider.(*schema.Provider); ok {
					rsrcSchema = provider.ResourcesMap[r.Type]
				}
			}

			// don't add empty flatmapped containers, so we can more easily
			// compare the attributes
			skipEmpty := func(k, v string) bool {
				if strings.HasSuffix(k, ".#") || strings.HasSuffix(k, ".%") {
					if v == "0" {
						return true
					}
				}
				return false
			}

			// Compare their attributes
			actual := make(map[string]string)
			for k, v := range r.Primary.Attributes {
				if skipEmpty(k, v) {
					continue
				}
				actual[k] = v
			}

			expected := make(map[string]string)
			for k, v := range oldR.Primary.Attributes {
				if skipEmpty(k, v) {
					continue
				}
				expected[k] = v
			}

			// Remove fields we're ignoring
			for _, v := range step.ImportStateVerifyIgnore {
				for k := range actual {
					if strings.HasPrefix(k, v) {
						delete(actual, k)
					}
				}
				for k := range expected {
					if strings.HasPrefix(k, v) {
						delete(expected, k)
					}
				}
			}

			// Also remove any attributes that are marked as "Removed" in the
			// schema, if we have a schema to check that against.
			if rsrcSchema != nil {
				for k := range actual {
					for _, schema := range rsrcSchema.SchemasForFlatmapPath(k) {
						if schema.Removed != "" {
							delete(actual, k)
							break
						}
					}
				}
				for k := range expected {
					for _, schema := range rsrcSchema.SchemasForFlatmapPath(k) {
						if schema.Removed != "" {
							delete(expected, k)
							break
						}
					}
				}
			}

			if !reflect.DeepEqual(actual, expected) {
				// Determine only the different attributes
				for k, v := range expected {
					if av, ok := actual[k]; ok && v == av {
						delete(expected, k)
						delete(actual, k)
					}
				}

				t.Fatalf(
					"ImportStateVerify attributes not equivalent. Difference is shown below. Top is actual, bottom is expected."+
						"\n\n%s\n\n%s",
					spewConf.Sdump(actual), spewConf.Sdump(expected))
			}
		}
	}

	return nil
}
