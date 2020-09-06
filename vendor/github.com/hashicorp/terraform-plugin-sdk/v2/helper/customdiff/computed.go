package customdiff

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ComputedIf returns a CustomizeDiffFunc that sets the given key's new value
// as computed if the given condition function returns true.
func ComputedIf(key string, f ResourceConditionFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		if f(ctx, d, meta) {
			d.SetNewComputed(key)
		}
		return nil
	}
}
