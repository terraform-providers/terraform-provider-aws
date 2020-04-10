package resource

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

func testStepNewConfig(t *testing.T, c TestCase, wd *tftest.WorkingDir, step TestStep) error {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	var idRefreshCheck *terraform.ResourceState
	idRefresh := c.IDRefreshName != ""

	if !step.Destroy {
		state := getState(t, wd)
		if err := testStepTaint(state, step); err != nil {
			t.Fatalf("Error when tainting resources: %s", err)
		}
	}

	wd.RequireSetConfig(t, step.Config)

	if !step.PlanOnly {
		err := wd.Apply()
		if err != nil {
			return err
		}

		state := getState(t, wd)
		if step.Check != nil {
			if err := step.Check(state); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Test for perpetual diffs by performing a plan, a refresh, and another plan

	// do a plan
	wd.RequireCreatePlan(t)
	plan := wd.RequireSavedPlan(t)

	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {

			t.Fatalf("After applying this test step, the plan was not empty. %s", spewConf.Sdump(plan))
		}
	}

	// do a refresh
	if !c.PreventPostDestroyRefresh {
		wd.RequireRefresh(t)
	}

	// do another plan
	wd.RequireCreatePlan(t)
	plan = wd.RequireSavedPlan(t)

	// check if plan is empty
	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {

			t.Fatalf("After applying this test step and performing a `terraform refresh`, the plan was not empty. %s", spewConf.Sdump(plan))
		}
	}

	// ID-ONLY REFRESH
	// If we've never checked an id-only refresh and our state isn't
	// empty, find the first resource and test it.
	state := getState(t, wd)
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
