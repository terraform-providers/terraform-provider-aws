package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Configuration to return Deleted
	ConfigurationDeletedTimeout = 5 * time.Minute
)

// ConfigurationDeleted waits for an Configuration to return Deleted
func ConfigurationDeleted(conn *kafka.Kafka, arn string) (*kafka.DescribeConfigurationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ConfigurationStateDeleting},
		Target:  []string{ConfigurationStateDeleted},
		Refresh: ConfigurationState(conn, arn),
		Timeout: ConfigurationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafka.DescribeConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}
