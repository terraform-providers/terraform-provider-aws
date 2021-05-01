package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	BrokerCreateTimeout = 30 * time.Minute
	BrokerDeleteTimeout = 30 * time.Minute
	BrokerRebootTimeout = 30 * time.Minute
)

func BrokerCreated(conn *mq.MQ, id string) (*mq.DescribeBrokerResponse, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			mq.BrokerStateCreationInProgress,
			mq.BrokerStateRebootInProgress,
		},
		Target:  []string{mq.BrokerStateRunning},
		Timeout: BrokerCreateTimeout,
		Refresh: BrokerStatus(conn, id),
	}
	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*mq.DescribeBrokerResponse); ok {
		return output, err
	}

	return nil, err
}

func BrokerDeleted(conn *mq.MQ, id string) (*mq.DescribeBrokerResponse, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			mq.BrokerStateCreationFailed,
			mq.BrokerStateDeletionInProgress,
			mq.BrokerStateRebootInProgress,
			mq.BrokerStateRunning,
		},
		Target:  []string{},
		Timeout: BrokerDeleteTimeout,
		Refresh: BrokerStatus(conn, id),
	}
	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*mq.DescribeBrokerResponse); ok {
		return output, err
	}

	return nil, err
}

func BrokerRebooted(conn *mq.MQ, id string) (*mq.DescribeBrokerResponse, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			mq.BrokerStateRebootInProgress,
		},
		Target:  []string{mq.BrokerStateRunning},
		Timeout: BrokerRebootTimeout,
		Refresh: BrokerStatus(conn, id),
	}
	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*mq.DescribeBrokerResponse); ok {
		return output, err
	}

	return nil, err
}
