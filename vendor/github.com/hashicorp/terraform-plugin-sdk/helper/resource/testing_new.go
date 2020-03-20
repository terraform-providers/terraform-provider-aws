package resource

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

func getState(t *testing.T, wd *tftest.WorkingDir) *terraform.State {
	jsonState := wd.RequireState(t)
	state, err := shimStateFromJson(jsonState)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func RunNewTest(t *testing.T, c TestCase, providers map[string]terraform.ResourceProvider) {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true
	wd := acctest.TestHelper.RequireNewWorkingDir(t)

	defer func() {
		wd.RequireDestroy(t)

		if c.CheckDestroy != nil {
			statePostDestroy := getState(t, wd)

			if err := c.CheckDestroy(statePostDestroy); err != nil {
				t.Fatal(err)
			}
		}
		wd.Close()
	}()

	providerCfg := testProviderConfig(c)

	wd.RequireSetConfig(t, providerCfg)
	wd.RequireInit(t)

	// use this to track last step succesfully applied
	// acts as default for import tests
	var appliedCfg string

	for i, step := range c.Steps {

		if step.PreConfig != nil {
			step.PreConfig()
		}

		if step.SkipFunc != nil {
			skip, err := step.SkipFunc()
			if err != nil {
				t.Fatal(err)
			}
			if skip {
				log.Printf("[WARN] Skipping step %d", i)
				continue
			}
		}

		if step.ImportState {
			step.providers = providers
			err := testStepNewImportState(t, c, wd, step, appliedCfg)
			if err != nil {
				t.Fatal(err)
			}
			continue
		}

		if step.Config != "" {
			err := testStepNewConfig(t, c, wd, step)
			if step.ExpectError != nil {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				if !step.ExpectError.MatchString(err.Error()) {
					t.Fatalf("Expected an error with pattern, no match on: %s", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}
			appliedCfg = step.Config
			continue
		}

		t.Fatal("Unsupported test mode")
	}
}

func planIsEmpty(plan *tfjson.Plan) bool {
	for _, rc := range plan.ResourceChanges {
		if rc.Mode == tfjson.DataResourceMode {
			// Skip data sources as the current implementation ignores
			// existing state and they are all re-read every time
			continue
		}

		for _, a := range rc.Change.Actions {
			if a != tfjson.ActionNoop {
				return false
			}
		}
	}
	return true
}
func testIDRefresh(c TestCase, t *testing.T, wd *tftest.WorkingDir, step TestStep, r *terraform.ResourceState) error {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	// Build the state. The state is just the resource with an ID. There
	// are no attributes. We only set what is needed to perform a refresh.
	state := terraform.NewState()
	state.RootModule().Resources = make(map[string]*terraform.ResourceState)
	state.RootModule().Resources[c.IDRefreshName] = &terraform.ResourceState{}

	// Temporarily set the config to a minimal provider config for the refresh
	// test. After the refresh we can reset it.
	cfg := testProviderConfig(c)
	wd.RequireSetConfig(t, cfg)
	defer wd.RequireSetConfig(t, step.Config)

	// Refresh!
	wd.RequireRefresh(t)
	state = getState(t, wd)

	// Verify attribute equivalence.
	actualR := state.RootModule().Resources[c.IDRefreshName]
	if actualR == nil {
		return fmt.Errorf("Resource gone!")
	}
	if actualR.Primary == nil {
		return fmt.Errorf("Resource has no primary instance")
	}
	actual := actualR.Primary.Attributes
	expected := r.Primary.Attributes
	// Remove fields we're ignoring
	for _, v := range c.IDRefreshIgnore {
		for k, _ := range actual {
			if strings.HasPrefix(k, v) {
				delete(actual, k)
			}
		}
		for k, _ := range expected {
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

		spewConf := spew.NewDefaultConfig()
		spewConf.SortKeys = true
		return fmt.Errorf(
			"Attributes not equivalent. Difference is shown below. Top is actual, bottom is expected."+
				"\n\n%s\n\n%s",
			spewConf.Sdump(actual), spewConf.Sdump(expected))
	}

	return nil
}
