package resource

import (
	tfjson "github.com/hashicorp/terraform-json"
	tftest "github.com/hashicorp/terraform-plugin-test/v2"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepNewConfig(t testing.T, c TestCase, wd *tftest.WorkingDir, step TestStep) error {
	t.Helper()

	var idRefreshCheck *terraform.ResourceState
	idRefresh := c.IDRefreshName != ""

	if !step.Destroy {
		var state *terraform.State
		err := runProviderCommand(t, func() error {
			state = getState(t, wd)
			return nil
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}
		if err := testStepTaint(state, step); err != nil {
			t.Fatalf("Error when tainting resources: %s", err)
		}
	}

	wd.RequireSetConfig(t, step.Config)

	// require a refresh before applying
	// failing to do this will result in data sources not being updated
	err := runProviderCommand(t, func() error {
		return wd.Refresh()
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	if !step.PlanOnly {
		err := runProviderCommand(t, func() error {
			return wd.Apply()
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}

		var state *terraform.State
		err = runProviderCommand(t, func() error {
			state = getState(t, wd)
			return nil
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}
		if step.Check != nil {
			state.IsBinaryDrivenTest = true
			if err := step.Check(state); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Test for perpetual diffs by performing a plan, a refresh, and another plan

	// do a plan
	err = runProviderCommand(t, func() error {
		return wd.CreatePlan()
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	var plan *tfjson.Plan
	err = runProviderCommand(t, func() error {
		plan = wd.RequireSavedPlan(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {
			var stdout string
			err = runProviderCommand(t, func() error {
				stdout = wd.RequireSavedPlanStdout(t)
				return nil
			}, wd, c.ProviderFactories)
			if err != nil {
				return err
			}
			t.Fatalf("After applying this test step, the plan was not empty.\nstdout:\n\n%s", stdout)
		}
	}

	// do a refresh
	if !c.PreventPostDestroyRefresh {
		err := runProviderCommand(t, func() error {
			return wd.Refresh()
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}
	}

	// do another plan
	err = runProviderCommand(t, func() error {
		return wd.CreatePlan()
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	err = runProviderCommand(t, func() error {
		plan = wd.RequireSavedPlan(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	// check if plan is empty
	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {
			var stdout string
			err = runProviderCommand(t, func() error {
				stdout = wd.RequireSavedPlanStdout(t)
				return nil
			}, wd, c.ProviderFactories)
			if err != nil {
				return err
			}
			t.Fatalf("After applying this test step and performing a `terraform refresh`, the plan was not empty.\nstdout\n\n%s", stdout)
		}
	}

	// ID-ONLY REFRESH
	// If we've never checked an id-only refresh and our state isn't
	// empty, find the first resource and test it.
	var state *terraform.State
	err = runProviderCommand(t, func() error {
		state = getState(t, wd)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}
	if idRefresh && idRefreshCheck == nil && !state.Empty() {
		// Find the first non-nil resource in the state
		for _, m := range state.Modules {
			if len(m.Resources) > 0 {
				if v, ok := m.Resources[c.IDRefreshName]; ok {
					idRefreshCheck = v
				}

				break
			}
		}

		// If we have an instance to check for refreshes, do it
		// immediately. We do it in the middle of another test
		// because it shouldn't affect the overall state (refresh
		// is read-only semantically) and we want to fail early if
		// this fails. If refresh isn't read-only, then this will have
		// caught a different bug.
		if idRefreshCheck != nil {
			if err := testIDRefresh(c, t, wd, step, idRefreshCheck); err != nil {
				t.Fatalf(
					"[ERROR] Test: ID-only test failed: %s", err)
			}
		}
	}

	return nil
}
