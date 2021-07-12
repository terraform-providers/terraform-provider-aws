package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
)

func EbsSnapshotImportCompleted(conn *ec2.EC2, importTaskID string) (*ec2.SnapshotTaskDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{tfec2.EbsSnapshotImportActive,
			tfec2.EbsSnapshotImportUpdating,
			tfec2.EbsSnapshotImportValidating,
			tfec2.EbsSnapshotImportValidated,
			tfec2.EbsSnapshotImportConverting,
		},
		Target:  []string{tfec2.EbsSnapshotImportCompleted},
		Refresh: EbsSnapshotImportStatus(conn, importTaskID),
		Timeout: 60 * time.Minute,
		Delay:   10 * time.Second,
	}

	detail, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	} else {
		return detail.(*ec2.SnapshotTaskDetail), nil
	}
}
