package waiter

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Minimum amount of times to verify change propagation
	PropagationContinuousTargetOccurence = 2

	// Minimum amount of time to wait between S3control change polls
	PropagationMinTimeout = 5 * time.Second

	// Maximum amount of time to wait for S3control changes to propagate
	PropagationTimeout = 1 * time.Minute
)

func PublicAccessBlockConfigurationBlockPublicAclsUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   PublicAccessBlockConfigurationBlockPublicAcls(conn, accountID),
		Timeout:                   PropagationTimeout,
		MinTimeout:                PropagationMinTimeout,
		ContinuousTargetOccurence: PropagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}

func PublicAccessBlockConfigurationBlockPublicPolicyUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   PublicAccessBlockConfigurationBlockPublicPolicy(conn, accountID),
		Timeout:                   PropagationTimeout,
		MinTimeout:                PropagationMinTimeout,
		ContinuousTargetOccurence: PropagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}

func PublicAccessBlockConfigurationIgnorePublicAclsUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   PublicAccessBlockConfigurationIgnorePublicAcls(conn, accountID),
		Timeout:                   PropagationTimeout,
		MinTimeout:                PropagationMinTimeout,
		ContinuousTargetOccurence: PropagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}

func PublicAccessBlockConfigurationRestrictPublicBucketsUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   PublicAccessBlockConfigurationRestrictPublicBuckets(conn, accountID),
		Timeout:                   PropagationTimeout,
		MinTimeout:                PropagationMinTimeout,
		ContinuousTargetOccurence: PropagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}
