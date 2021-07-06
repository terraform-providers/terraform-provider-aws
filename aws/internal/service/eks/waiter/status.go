package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func AddonStatus(ctx context.Context, conn *eks.EKS, clusterName, addonName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.AddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func AddonUpdateStatus(ctx context.Context, conn *eks.EKS, clusterName, addonName, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.AddonUpdateByClusterNameAddonNameAndID(ctx, conn, clusterName, addonName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func ClusterStatus(conn *eks.EKS, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ClusterByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func ClusterUpdateStatus(conn *eks.EKS, name, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ClusterUpdateByNameAndID(conn, name, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func FargateProfileStatus(conn *eks.EKS, clusterName, fargateProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FargateProfileByClusterNameAndFargateProfileName(conn, clusterName, fargateProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func NodegroupStatus(conn *eks.EKS, clusterName, nodeGroupName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.NodegroupByClusterNameAndNodegroupName(conn, clusterName, nodeGroupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func NodegroupUpdateStatus(conn *eks.EKS, clusterName, nodeGroupName, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.NodegroupUpdateByClusterNameNodegroupNameAndID(conn, clusterName, nodeGroupName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
