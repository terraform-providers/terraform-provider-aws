package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	UpgradeStatusUnknown = "Unknown"
)

// UpgradeStatus fetches the Domain upgrade status
func UpgradeStatus(conn *elasticsearch.ElasticsearchService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.GetUpgradeStatus(&elasticsearch.GetUpgradeStatusInput{
			DomainName: aws.String(name),
		})
		if err != nil {
			return nil, UpgradeStatusUnknown, err
		}

		// Elasticsearch upgrades consist of multiple steps:
		// https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-version-migration.html
		// Prevent false positive completion where the UpgradeStep is not the final UPGRADE step.
		if aws.StringValue(out.StepStatus) == elasticsearch.UpgradeStatusSucceeded && aws.StringValue(out.UpgradeStep) != elasticsearch.UpgradeStepUpgrade {
			return out, elasticsearch.UpgradeStatusInProgress, nil
		}

		return out, aws.StringValue(out.StepStatus), nil
	}
}
