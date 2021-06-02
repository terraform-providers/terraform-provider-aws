package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AutoScalingConfigurationCreateTimeout = 2 * time.Minute
	AutoScalingConfigurationDeleteTimeout = 2 * time.Minute

	ConnectionDeleteTimeout = 5 * time.Minute

	CustomDomainAssociationCreateTimeout = 5 * time.Minute
	CustomDomainAssociationDeleteTimeout = 5 * time.Minute

	ServiceCreateTimeout = 20 * time.Minute
	ServiceDeleteTimeout = 20 * time.Minute
	ServiceUpdateTimeout = 20 * time.Minute
)

func AutoScalingConfigurationActive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{AutoScalingConfigurationStatusActive},
		Refresh: AutoScalingConfigurationStatus(ctx, conn, arn),
		Timeout: AutoScalingConfigurationCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func AutoScalingConfigurationInactive(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{AutoScalingConfigurationStatusActive},
		Target:  []string{AutoScalingConfigurationStatusInactive},
		Refresh: AutoScalingConfigurationStatus(ctx, conn, arn),
		Timeout: AutoScalingConfigurationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ConnectionDeleted(ctx context.Context, conn *apprunner.AppRunner, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ConnectionStatusPendingHandshake, apprunner.ConnectionStatusAvailable, apprunner.ConnectionStatusDeleted},
		Target:  []string{},
		Refresh: ConnectionStatus(ctx, conn, name),
		Timeout: ConnectionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func CustomDomainAssociationCreated(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusCreating},
		Target:  []string{CustomDomainAssociationStatusPendingCertificateDnsValidation},
		Refresh: CustomDomainStatus(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func CustomDomainAssociationDeleted(ctx context.Context, conn *apprunner.AppRunner, domainName, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{CustomDomainAssociationStatusActive, CustomDomainAssociationStatusDeleting},
		Target:  []string{},
		Refresh: CustomDomainStatus(ctx, conn, domainName, serviceArn),
		Timeout: CustomDomainAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ServiceCreated(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusRunning},
		Refresh: ServiceStatus(ctx, conn, serviceArn),
		Timeout: ServiceCreateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ServiceUpdated(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusRunning},
		Refresh: ServiceStatus(ctx, conn, serviceArn),
		Timeout: ServiceUpdateTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ServiceDeleted(ctx context.Context, conn *apprunner.AppRunner, serviceArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.ServiceStatusRunning, apprunner.ServiceStatusOperationInProgress},
		Target:  []string{apprunner.ServiceStatusDeleted},
		Refresh: ServiceStatus(ctx, conn, serviceArn),
		Timeout: ServiceDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
