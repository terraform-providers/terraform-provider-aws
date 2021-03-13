package waiter

import (
	"time"

	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	UpgradeSuccessMinTimeout = 10 * time.Second
	UpgradeSuccessDelay      = 30 * time.Second
)

// UpgradeSucceeded waits for an Upgrade to return Success
func UpgradeSucceeded(conn *elasticsearch.ElasticsearchService, name string, timeout time.Duration) (*elasticsearch.GetUpgradeStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{elasticsearch.UpgradeStatusInProgress},
		Target:     []string{elasticsearch.UpgradeStatusSucceeded},
		Refresh:    UpgradeStatus(conn, name),
		Timeout:    timeout,
		MinTimeout: UpgradeSuccessMinTimeout,
		Delay:      UpgradeSuccessDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*elasticsearch.GetUpgradeStatusOutput); ok {
		return output, err
	}

	return nil, err
}
