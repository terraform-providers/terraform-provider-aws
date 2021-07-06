package waiter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	AddonCreatedTimeout = 20 * time.Minute
	AddonUpdatedTimeout = 20 * time.Minute
	AddonDeletedTimeout = 40 * time.Minute
)

func AddonCreated(ctx context.Context, conn *eks.EKS, clusterName, addonName string) (*eks.Addon, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.AddonStatusCreating},
		Target:  []string{eks.AddonStatusActive},
		Refresh: AddonStatus(ctx, conn, clusterName, addonName),
		Timeout: AddonCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Addon); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.AddonStatusCreateFailed && health != nil {
			var errs *multierror.Error

			for _, issue := range health.Issues {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(issue.Code), aws.StringValue(issue.Message)))
			}
			tfresource.SetLastError(err, errs.ErrorOrNil())
		}

		return output, err
	}

	return nil, err
}

func AddonDeleted(ctx context.Context, conn *eks.EKS, clusterName, addonName string) (*eks.Addon, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.AddonStatusActive, eks.AddonStatusDeleting},
		Target:  []string{},
		Refresh: AddonStatus(ctx, conn, clusterName, addonName),
		Timeout: AddonDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Addon); ok {
		return output, err
	}

	return nil, err
}

func AddonUpdateSuccessful(ctx context.Context, conn *eks.EKS, clusterName, addonName, id string) (*eks.Update, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: AddonUpdateStatus(ctx, conn, clusterName, addonName, id),
		Timeout: AddonUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			var errs *multierror.Error

			for _, e := range output.Errors {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(e.ErrorCode), aws.StringValue(e.ErrorMessage)))
			}
			tfresource.SetLastError(err, errs.ErrorOrNil())
		}

		return output, err
	}

	return nil, err
}

func ClusterCreated(conn *eks.EKS, name string, timeout time.Duration) (*eks.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.ClusterStatusCreating},
		Target:  []string{eks.ClusterStatusActive},
		Refresh: ClusterStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Cluster); ok {
		return output, err
	}

	return nil, err
}

func ClusterDeleted(conn *eks.EKS, name string, timeout time.Duration) (*eks.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.ClusterStatusActive, eks.ClusterStatusDeleting},
		Target:  []string{},
		Refresh: ClusterStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Cluster); ok {
		return output, err
	}

	return nil, err
}

func ClusterUpdateSuccessful(conn *eks.EKS, name, id string, timeout time.Duration) (*eks.Update, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: ClusterUpdateStatus(conn, name, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			var errs *multierror.Error

			for _, e := range output.Errors {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(e.ErrorCode), aws.StringValue(e.ErrorMessage)))
			}
			tfresource.SetLastError(err, errs.ErrorOrNil())
		}

		return output, err
	}

	return nil, err
}

func FargateProfileCreated(conn *eks.EKS, clusterName, fargateProfileName string, timeout time.Duration) (*eks.FargateProfile, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusCreating},
		Target:  []string{eks.FargateProfileStatusActive},
		Refresh: FargateProfileStatus(conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func FargateProfileDeleted(conn *eks.EKS, clusterName, fargateProfileName string, timeout time.Duration) (*eks.FargateProfile, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusActive, eks.FargateProfileStatusDeleting},
		Target:  []string{},
		Refresh: FargateProfileStatus(conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func NodegroupCreated(conn *eks.EKS, clusterName, nodeGroupName string, timeout time.Duration) (*eks.Nodegroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.NodegroupStatusCreating},
		Target:  []string{eks.NodegroupStatusActive},
		Refresh: NodegroupStatus(conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Nodegroup); ok {
		return output, err
	}

	return nil, err
}

func NodegroupDeleted(conn *eks.EKS, clusterName, nodeGroupName string, timeout time.Duration) (*eks.Nodegroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.NodegroupStatusActive, eks.NodegroupStatusDeleting},
		Target:  []string{},
		Refresh: NodegroupStatus(conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Nodegroup); ok {
		return output, err
	}

	return nil, err
}

func NodegroupUpdateSuccessful(conn *eks.EKS, clusterName, nodeGroupName, id string, timeout time.Duration) (*eks.Update, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: NodegroupUpdateStatus(conn, clusterName, nodeGroupName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			var errs *multierror.Error

			for _, e := range output.Errors {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(e.ErrorCode), aws.StringValue(e.ErrorMessage)))
			}
			tfresource.SetLastError(err, errs.ErrorOrNil())
		}

		return output, err
	}

	return nil, err
}
