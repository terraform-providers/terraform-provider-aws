package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	LexSlotTypeDeleteTimeout = 5 * time.Minute
)

func LexSlotTypeDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, slotTypeId string) (*lexmodelbuildingservice.GetSlotTypeVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexSlotTypeStatus(conn, slotTypeId),
		Timeout: LexSlotTypeDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeVersionsOutput); ok {
		return v, err
	}

	return nil, err
}
