package waiter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func EksAddonStatus(ctx context.Context, conn *eks.EKS, addonName, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}
		if err != nil {
			return output, "", err
		}
		if output == nil || output.Addon == nil {
			return nil, "", fmt.Errorf("EKS Cluster (%s) add-on (%s) missing", clusterName, addonName)
		}
		return output.Addon, aws.StringValue(output.Addon.Status), nil
	}
}

func EksAddonUpdateStatus(ctx context.Context, conn *eks.EKS, clusterName, addonName, updateID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeUpdateWithContext(ctx, &eks.DescribeUpdateInput{
			Name:      aws.String(clusterName),
			AddonName: aws.String(addonName),
			UpdateId:  aws.String(updateID),
		})
		if err != nil {
			return output, "", err
		}
		if output == nil || output.Update == nil {
			return nil, "", fmt.Errorf("EKS Cluster (%s) add-on (%s) update (%s) missing", clusterName, addonName, updateID)
		}
		return output.Update, aws.StringValue(output.Update.Status), nil
	}
}
