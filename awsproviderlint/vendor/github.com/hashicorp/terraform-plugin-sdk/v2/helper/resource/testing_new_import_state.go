package resource

import (
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepNewImportState(t testing.T, c TestCase, helper *plugintest.Helper, wd *plugintest.WorkingDir, step TestStep, cfg string) error {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	if step.ResourceName == "" {
		t.Fatal("ResourceName is required for an import state test")
	}

	// get state from check sequence
	var state *terraform.State
	var err error
	err = runProviderCommand(t, func() error {
		state, err = getState(t, wd)
		if err != nil {
			return err
		}
		return nil
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
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
	err = importWd.SetConfig(step.Config)
	if err != nil {
		t.Fatalf("Error setting test config: %s", err)
	}

	err = runProviderCommand(t, func() error {
		return importWd.Init()
	}, importWd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		t.Fatalf("Error running init: %s", err)
	}

	err = runProviderCommand(t, func() error {
		return importWd.Import(step.ResourceName, importId)
	}, importWd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		return err
	}

	var importState *terraform.State
	err = runProviderCommand(t, func() error {
		importState, err = getState(t, importWd)
		if err != nil {
			return err
		}
		return nil
	}, importWd, c.ProviderFactories, c.ProtoV5ProviderFactories)
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
			for r2Key, r2 := range old {
				// Ensure that we do not match against data sources as they
				// cannot be imported and are not what we want to verify.
				// Mode is not present in ResourceState so we use the
				// stringified ResourceStateKey for comparison.
				if strings.HasPrefix(r2Key, "data.") {
					continue
				}

				if r2.Primary != nil && r2.Primary.ID == r.Primary.ID && r2.Type == r.Type && r2.Provider == r.Provider {
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

			// timeouts are only _sometimes_ added to state. To
			// account for this, just don't compare timeouts at
			// all.
			for k := range actual {
				if strings.HasPrefix(k, "timeouts.") {
					delete(actual, k)
				}
				if k == "timeouts" {
					delete(actual, k)
				}
			}
			for k := range expected {
				if strings.HasPrefix(k, "timeouts.") {
					delete(expected, k)
				}
				if k == "timeouts" {
					delete(expected, k)
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
