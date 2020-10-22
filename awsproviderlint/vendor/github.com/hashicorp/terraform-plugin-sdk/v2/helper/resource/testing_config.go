package resource

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepTaint(state *terraform.State, step TestStep) error {
	for _, p := range step.Taint {
		m := state.RootModule()
		if m == nil {
			return errors.New("no state")
		}
		rs, ok := m.Resources[p]
		if !ok {
			return fmt.Errorf("resource %q not found in state", p)
		}
		log.Printf("[WARN] Test: Explicitly tainting resource %q", p)
		rs.Taint()
	}
	return nil
}
