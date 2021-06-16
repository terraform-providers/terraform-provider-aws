package glue

import (
	"fmt"
	"strings"
)

func DecodeBudgetsBudgetActionID(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("Unexpected format of ID (%q), expected AccountID:ActionID:BudgetName", id)
	}
	return parts[0], parts[1], parts[2], nil
}
