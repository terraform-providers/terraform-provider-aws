package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	LexBotDeleteTimeout      = 5 * time.Minute
	LexIntentDeleteTimeout   = 5 * time.Minute
	LexSlotTypeDeleteTimeout = 5 * time.Minute
)

func LexBotDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botId string) (*lexmodelbuildingservice.GetBotVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexBotStatus(conn, botId),
		Timeout: LexBotDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

func LexBotAliasDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexBotAliasStatus(conn, botAliasName, botName),
		Timeout: LexBotDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotAliasOutput); ok {
		return v, err
	}

	return nil, err
}

func LexIntentDeleted(conn *lexmodelbuildingservice.LexModelBuildingService, intentId string) (*lexmodelbuildingservice.GetIntentVersionsOutput, error) {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{LexModelBuildingServiceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: LexIntentStatus(conn, intentId),
		Timeout: LexIntentDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForState()

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetIntentVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

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
