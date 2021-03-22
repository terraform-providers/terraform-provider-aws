package resource

import (
	"errors"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepNewConfig(t testing.T, c TestCase, wd *plugintest.WorkingDir, step TestStep) error {
	t.Helper()

	var idRefreshCheck *terraform.ResourceState
	idRefresh := c.IDRefreshName != ""

	if !step.Destroy {
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
			return err
		}
		if err := testStepTaint(state, step); err != nil {
			return fmt.Errorf("Error when tainting resources: %s", err)
		}
	}

	err := wd.SetConfig(step.Config)
	if err != nil {
		return fmt.Errorf("Error setting config: %w", err)
	}

	// require a refresh before applying
	// failing to do this will result in data sources not being updated
	err = runProviderCommand(t, func() error {
		return wd.Refresh()
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		return fmt.Errorf("Error running pre-apply refresh: %w", err)
	}

	// If this step is a PlanOnly step, skip over this first Plan and
	// subsequent Apply, and use the follow-up Plan that checks for
	// permadiffs
	if !step.PlanOnly {
		// Plan!
		err := runProviderCommand(t, func() error {
			if step.Destroy {
				return wd.CreateDestroyPlan()
			}
			return wd.CreatePlan()
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			return fmt.Errorf("Error running pre-apply plan: %w", err)
		}

		// We need to keep a copy of the state prior to destroying such
		// that the destroy steps can verify their behavior in the
		// check function
		var stateBeforeApplication *terraform.State
		err = runProviderCommand(t, func() error {
			stateBeforeApplication, err = getState(t, wd)
			if err != nil {
				return err
			}
			return nil
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			return fmt.Errorf("Error retrieving pre-apply state: %w", err)
		}

		// Apply the diff, creating real resources
		err = runProviderCommand(t, func() error {
			return wd.Apply()
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			if step.Destroy {
				return fmt.Errorf("Error running destroy: %w", err)
			}
			return fmt.Errorf("Error running apply: %w", err)
		}

		// Get the new state
		var state *terraform.State
		err = runProviderCommand(t, func() error {
			state, err = getState(t, wd)
			if err != nil {
				return err
			}
			return nil
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			return fmt.Errorf("Error retrieving state after apply: %w", err)
		}

		// Run any configured checks
		if step.Check != nil {
			state.IsBinaryDrivenTest = true
			if step.Destroy {
				if err := step.Check(stateBeforeApplication); err != nil {
					return fmt.Errorf("Check failed: %w", err)
				}
			} else {
				if err := step.Check(state); err != nil {
					return fmt.Errorf("Check failed: %w", err)
				}
			}
		}
	}

	// Test for perpetual diffs by performing a plan, a refresh, and another plan

	// do a plan
	err = runProviderCommand(t, func() error {
		if step.Destroy {
			return wd.CreateDestroyPlan()
		}
		return wd.CreatePlan()
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		return fmt.Errorf("Error running post-apply plan: %w", err)
	}

	var plan *tfjson.Plan
	err = runProviderCommand(t, func() error {
		var err error
		plan, err = wd.SavedPlan()
		return err
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		return fmt.Errorf("Error retrieving post-apply plan: %w", err)
	}

	if !planIsEmpty(plan) && !step.ExpectNonEmptyPlan {
		var stdout string
		err = runProviderCommand(t, func() error {
			var err error
			stdout, err = wd.SavedPlanRawStdout()
			return err
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			return fmt.Errorf("Error retrieving formatted plan output: %w", err)
		}
		return fmt.Errorf("After applying this test step, the plan was not empty.\nstdout:\n\n%s", stdout)
	}

	// do a refresh
	if !step.Destroy || (step.Destroy && !step.PreventPostDestroyRefresh) {
		err := runProviderCommand(t, func() error {
			return wd.Refresh()
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			return fmt.Errorf("Error running post-apply refresh: %w", err)
		}
	}

	// do another plan
	err = runProviderCommand(t, func() error {
		if step.Destroy {
			return wd.CreateDestroyPlan()
		}
		return wd.CreatePlan()
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		return fmt.Errorf("Error running second post-apply plan: %w", err)
	}

	err = runProviderCommand(t, func() error {
		var err error
		plan, err = wd.SavedPlan()
		return err
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
	if err != nil {
		return fmt.Errorf("Error retrieving second post-apply plan: %w", err)
	}

	// check if plan is empty
	if !planIsEmpty(plan) && !step.ExpectNonEmptyPlan {
		var stdout string
		err = runProviderCommand(t, func() error {
			var err error
			stdout, err = wd.SavedPlanRawStdout()
			return err
		}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
		if err != nil {
			return fmt.Errorf("Error retrieving formatted second plan output: %w", err)
		}
		return fmt.Errorf("After applying this test step and performing a `terraform refresh`, the plan was not empty.\nstdout\n\n%s", stdout)
	} else if step.ExpectNonEmptyPlan && planIsEmpty(plan) {
		return errors.New("Expected a non-empty plan, but got an empty plan")
	}

	// ID-ONLY REFRESH
	// If we've never checked an id-only refresh and our state isn't
	// empty, find the first resource and test it.
	var state *terraform.State
	err = runProviderCommand(t, func() error {
		state, err = getState(t, wd)
		if err != nil {
			return err
		}
		return nil
	}, wd, c.ProviderFactories, c.ProtoV5ProviderFactories)
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
				return fmt.Errorf(
					"[ERROR] Test: ID-only test failed: %s", err)
			}
		}
	}

	return nil
}
