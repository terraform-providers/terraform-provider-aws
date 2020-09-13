package resource

import (
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	tftest "github.com/hashicorp/terraform-plugin-test/v2"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepNewImportState(t testing.T, c TestCase, helper *tftest.Helper, wd *tftest.WorkingDir, step TestStep, cfg string) error {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	if step.ResourceName == "" {
		t.Fatal("ResourceName is required for an import state test")
	}

	// get state from check sequence
	var state *terraform.State
	err := runProviderCommand(t, func() error {
		state = getState(t, wd)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		t.Fatalf("Error getting state: %s", err)
	}

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
	importWd := helper.RequireNewWorkingDir(t)
	defer importWd.Close()
	importWd.RequireSetConfig(t, step.Config)

	err = runProviderCommand(t, func() error {
		importWd.RequireInit(t)
		return nil
	}, importWd, c.ProviderFactories)
	if err != nil {
		t.Fatalf("Error running init: %s", err)
	}

	err = runProviderCommand(t, func() error {
		importWd.RequireImport(t, step.ResourceName, importId)
		return nil
	}, importWd, c.ProviderFactories)
	if err != nil {
		t.Fatalf("Error running import: %s", err)
	}

	var importState *terraform.State
	err = runProviderCommand(t, func() error {
		importState = getState(t, wd)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		t.Fatalf("Error getting state: %s", err)
	}

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
