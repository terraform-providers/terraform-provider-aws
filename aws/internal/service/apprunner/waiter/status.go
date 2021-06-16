package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/finder"
)

const (
	AutoScalingConfigurationStatusActive   = "active"
	AutoScalingConfigurationStatusInactive = "inactive"

	CustomDomainAssociationStatusActive                          = "active"
	CustomDomainAssociationStatusCreating                        = "creating"
	CustomDomainAssociationStatusDeleting                        = "deleting"
	CustomDomainAssociationStatusPendingCertificateDnsValidation = "pending_certificate_dns_validation"
)

func AutoScalingConfigurationStatus(ctx context.Context, conn *apprunner.AppRunner, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeAutoScalingConfigurationInput{
			AutoScalingConfigurationArn: aws.String(arn),
		}

		output, err := conn.DescribeAutoScalingConfigurationWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.AutoScalingConfiguration == nil {
			return nil, "", nil
		}

		return output.AutoScalingConfiguration, aws.StringValue(output.AutoScalingConfiguration.Status), nil
	}
}

func ConnectionStatus(ctx context.Context, conn *apprunner.AppRunner, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := finder.ConnectionSummaryByName(ctx, conn, name)

		if err != nil {
			return nil, "", err
		}

		if c == nil {
			return nil, "", nil
		}

		return c, aws.StringValue(c.Status), nil
	}
}

func CustomDomainStatus(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		customDomain, err := finder.CustomDomain(ctx, conn, domainName, serviceArn)

		if err != nil {
			return nil, "", err
		}

		if customDomain == nil {
			return nil, "", nil
		}

		return customDomain, aws.StringValue(customDomain.Status), nil
	}
}

func ServiceStatus(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(serviceArn),
		}

		output, err := conn.DescribeServiceWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Service == nil {
			return nil, "", nil
		}

		return output.Service, aws.StringValue(output.Service.Status), nil
	}
}
