package resource

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
	tftest "github.com/hashicorp/terraform-plugin-test/v2"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func runPostTestDestroy(t testing.T, c TestCase, wd *tftest.WorkingDir, factories map[string]func() (*schema.Provider, error)) error {
	t.Helper()

	err := runProviderCommand(t, func() error {
		wd.RequireDestroy(t)
		return nil
	}, wd, factories)
	if err != nil {
		return err
	}

	if c.CheckDestroy != nil {
		var statePostDestroy *terraform.State
		err := runProviderCommand(t, func() error {
			statePostDestroy = getState(t, wd)
			return nil
		}, wd, factories)
		if err != nil {
			return err
		}

		if err := c.CheckDestroy(statePostDestroy); err != nil {
			return err
		}
	}

	return nil
}

func runNewTest(t testing.T, c TestCase, helper *tftest.Helper) {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true
	wd := helper.RequireNewWorkingDir(t)

	defer func() {
		var statePreDestroy *terraform.State
		err := runProviderCommand(t, func() error {
			statePreDestroy = getState(t, wd)
			return nil
		}, wd, c.ProviderFactories)
		if err != nil {
			t.Fatalf("Error retrieving state, there may be dangling resources: %s", err.Error())
			return
		}

		if !stateIsEmpty(statePreDestroy) {
			runPostTestDestroy(t, c, wd, c.ProviderFactories)
		}

		wd.Close()
	}()

	providerCfg, err := testProviderConfig(c)
	if err != nil {
		t.Fatal(err)
	}

	wd.RequireSetConfig(t, providerCfg)
	err = runProviderCommand(t, func() error {
		wd.RequireInit(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		t.Fatalf("Error running init: %s", err.Error())
		return
	}

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
				log.Printf("[WARN] Skipping step %d/%d", i+1, len(c.Steps))
				continue
			}
		}

		if step.ImportState {
			err := testStepNewImportState(t, c, helper, wd, step, appliedCfg)
			if err != nil {
				t.Fatal(err)
			}
			continue
		}

		if step.Config != "" {
			err := testStepNewConfig(t, c, wd, step)
			if step.ExpectError != nil {
				if err == nil {
					t.Fatalf("Step %d/%d, expected an error but got none", i+1, len(c.Steps))
				}
				if !step.ExpectError.MatchString(err.Error()) {
					t.Fatalf("Step %d/%d, expected an error with pattern, no match on: %s", i+1, len(c.Steps), err)
				}
			} else {
				if err != nil {
					t.Fatalf("Step %d/%d error: %s", i+1, len(c.Steps), err)
				}
			}
			appliedCfg = step.Config
			continue
		}

		t.Fatal("Unsupported test mode")
	}
}

func getState(t testing.T, wd *tftest.WorkingDir) *terraform.State {
	t.Helper()

	jsonState := wd.RequireState(t)
	state, err := shimStateFromJson(jsonState)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func stateIsEmpty(state *terraform.State) bool {
	return state.Empty() || !state.HasResources()
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

func testIDRefresh(c TestCase, t testing.T, wd *tftest.WorkingDir, step TestStep, r *terraform.ResourceState) error {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	// Build the state. The state is just the resource with an ID. There
	// are no attributes. We only set what is needed to perform a refresh.
	state := terraform.NewState()
	state.RootModule().Resources = make(map[string]*terraform.ResourceState)
	state.RootModule().Resources[c.IDRefreshName] = &terraform.ResourceState{}

	// Temporarily set the config to a minimal provider config for the refresh
	// test. After the refresh we can reset it.
	cfg, err := testProviderConfig(c)
	if err != nil {
		return err
	}
	wd.RequireSetConfig(t, cfg)
	defer wd.RequireSetConfig(t, step.Config)

	// Refresh!
	err = runProviderCommand(t, func() error {
		wd.RequireRefresh(t)
		state = getState(t, wd)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

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

		spewConf := spew.NewDefaultConfig()
		spewConf.SortKeys = true
		return fmt.Errorf(
			"Attributes not equivalent. Difference is shown below. Top is actual, bottom is expected."+
				"\n\n%s\n\n%s",
			spewConf.Sdump(actual), spewConf.Sdump(expected))
	}

	return nil
}
