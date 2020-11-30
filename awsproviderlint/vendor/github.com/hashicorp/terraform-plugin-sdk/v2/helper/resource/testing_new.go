package resource

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func runPostTestDestroy(t testing.T, c TestCase, wd *plugintest.WorkingDir, factories map[string]func() (*schema.Provider, error), v5factories map[string]func() (tfprotov5.ProviderServer, error), statePreDestroy *terraform.State) error {
	t.Helper()

	err := runProviderCommand(t, func() error {
		return wd.Destroy()
	}, wd, factories, v5factories)
	if err != nil {
		return err
	}

	if c.CheckDestroy != nil {
		if err := c.CheckDestroy(statePreDestroy); err != nil {
			return err
		}
	}

	return nil
}

func runNewTest(t testing.T, c TestCase, helper *plugintest.Helper) {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true
	wd := helper.RequireNewWorkingDir(t)

	defer func() {
		var statePreDestroy *terraform.State
		var err error
		err = runProviderCommand(t, func() error {
			statePreDestroy, err = getState(t, wd)
			if err != nil {
				return err
			}
			return nil
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			t.Fatalf("Error retrieving state, there may be dangling resources: %s", err.Error())
			return
		}

		if !stateIsEmpty(statePreDestroy) {
			err := runPostTestDestroy(t, c, wd, c.ProviderFactories, c.ProtoV5ProviderFactories, statePreDestroy)
			if err != nil {
				t.Fatalf("Error running post-test destroy, there may be dangling resources: %s", err.Error())
			}
		}

		wd.Close()
	}()

	providerCfg, err := testProviderConfig(c)
	if err != nil {
		t.Fatal(err)
	}

	err = wd.SetConfig(providerCfg)
	if err != nil {
		t.Fatalf("Error setting test config: %s", err)
	}
	err = runProviderCommand(t, func() error {
		return wd.Init()
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
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
			if step.ExpectError != nil {
				if err == nil {
					t.Fatalf("Step %d/%d error running import: expected an error but got none", i+1, len(c.Steps))
				}
				if !step.ExpectError.MatchString(err.Error()) {
					t.Fatalf("Step %d/%d error running import, expected an error with pattern (%s), no match on: %s", i+1, len(c.Steps), step.ExpectError.String(), err)
				}
			} else {
				if c.ErrorCheck != nil {
					err = c.ErrorCheck(err)
				}
				if err != nil {
					t.Fatalf("Step %d/%d error running import: %s", i+1, len(c.Steps), err)
				}
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
				if c.ErrorCheck != nil {
					err = c.ErrorCheck(err)
				}
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

func getState(t testing.T, wd *plugintest.WorkingDir) (*terraform.State, error) {
	t.Helper()

	jsonState, err := wd.State()
	if err != nil {
		return nil, err
	}
	state, err := shimStateFromJson(jsonState)
	if err != nil {
		t.Fatal(err)
	}
	return state, nil
}

func stateIsEmpty(state *terraform.State) bool {
	return state.Empty() || !state.HasResources()
}

func planIsEmpty(plan *tfjson.Plan) bool {
	for _, rc := range plan.ResourceChanges {
		for _, a := range rc.Change.Actions {
			if a != tfjson.ActionNoop {
				return false
			}
		}
	}
	return true
}

func testIDRefresh(c TestCase, t testing.T, wd *plugintest.WorkingDir, step TestStep, r *terraform.ResourceState) error {
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
	err = wd.SetConfig(cfg)
	if err != nil {
		t.Fatalf("Error setting import test config: %s", err)
	}
	defer func() {
		err = wd.SetConfig(step.Config)
		if err != nil {
			t.Fatalf("Error resetting test config: %s", err)
		}
	}()

	// Refresh!
	err = runProviderCommand(t, func() error {
		err = wd.Refresh()
		if err != nil {
			t.Fatalf("Error running terraform refresh: %s", err)
		}
		state, err = getState(t, wd)
		if err != nil {
			return err
		}
		return nil
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
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
