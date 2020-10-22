package waiter

import (
	"github.com/aws/aws-sdk-go/service/glue"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	MLTransformDeleteTimeout = 2 * time.Minute
)

// MLTransformDeleted waits for an Operation to return Success
func MLTransformDeleted(conn *glue.Glue, transformId string) (*glue.GetMLTransformOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.TransformStatusTypeNotReady, glue.TransformStatusTypeReady, glue.TransformStatusTypeDeleting},
		Target:  []string{},
		Refresh: MLTransformStatus(conn, transformId),
		Timeout: MLTransformDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetMLTransformOutput); ok {
		return output, err
	}

	return nil, err
}
