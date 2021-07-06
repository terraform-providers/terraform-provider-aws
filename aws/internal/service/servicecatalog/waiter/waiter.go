package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ProductReadyTimeout  = 3 * time.Minute
	ProductDeleteTimeout = 3 * time.Minute

	TagOptionReadyTimeout  = 3 * time.Minute
	TagOptionDeleteTimeout = 3 * time.Minute

	PortfolioShareCreateTimeout = 3 * time.Minute

	OrganizationsAccessStableTimeout = 3 * time.Minute
	ConstraintReadyTimeout           = 3 * time.Minute
	ConstraintDeleteTimeout          = 3 * time.Minute

	ProductPortfolioAssociationReadyTimeout  = 3 * time.Minute
	ProductPortfolioAssociationDeleteTimeout = 3 * time.Minute

	ServiceActionReadyTimeout  = 3 * time.Minute
	ServiceActionDeleteTimeout = 3 * time.Minute

	BudgetResourceAssociationReadyTimeout  = 3 * time.Minute
	BudgetResourceAssociationDeleteTimeout = 3 * time.Minute

	TagOptionResourceAssociationReadyTimeout  = 3 * time.Minute
	TagOptionResourceAssociationDeleteTimeout = 3 * time.Minute

	ProvisioningArtifactReadyTimeout   = 3 * time.Minute
	ProvisioningArtifactDeletedTimeout = 3 * time.Minute

	PrincipalPortfolioAssociationReadyTimeout  = 3 * time.Minute
	PrincipalPortfolioAssociationDeleteTimeout = 3 * time.Minute

	LaunchPathsReadyTimeout = 3 * time.Minute

	ProvisionedProductReadyTimeout  = 3 * time.Minute
	ProvisionedProductDeleteTimeout = 3 * time.Minute

	RecordReadyTimeout = 3 * time.Minute

	PortfolioConstraintsReadyTimeout = 3 * time.Minute

	MinTimeout                 = 2 * time.Second
	NotFoundChecks             = 5
	ContinuousTargetOccurrence = 2

	StatusNotFound    = "NOT_FOUND"
	StatusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	StatusCreated = "CREATED"

	OrganizationAccessStatusError = "ERROR"
)

func ProductReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable, StatusCreated},
		Refresh:                   ProductStatus(conn, acceptLanguage, productID),
		Timeout:                   ProductReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	return nil, err
}

func ProductDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: ProductStatus(conn, acceptLanguage, productID),
		Timeout: ProductDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, nil
	}

	return nil, err
}

func TagOptionReady(conn *servicecatalog.ServiceCatalog, id string) (*servicecatalog.TagOptionDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: TagOptionStatus(conn, id),
		Timeout: TagOptionReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.TagOptionDetail); ok {
		return output, err
	}

	return nil, err
}

func TagOptionDeleted(conn *servicecatalog.ServiceCatalog, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: TagOptionStatus(conn, id),
		Timeout: TagOptionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func PortfolioShareReady(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string, acceptRequired bool) (*servicecatalog.PortfolioShareDetail, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: PortfolioShareStatus(conn, portfolioID, shareType, principalID),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func PortfolioShareCreatedWithToken(conn *servicecatalog.ServiceCatalog, token string, acceptRequired bool) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	targets := []string{servicecatalog.ShareStatusCompleted}

	if !acceptRequired {
		targets = append(targets, servicecatalog.ShareStatusInProgress)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  targets,
		Refresh: PortfolioShareStatusWithToken(conn, token),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func PortfolioShareDeleted(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) (*servicecatalog.PortfolioShareDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, servicecatalog.ShareStatusCompleted, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: PortfolioShareStatus(conn, portfolioID, shareType, principalID),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil, nil
	}

	if output, ok := outputRaw.(*servicecatalog.PortfolioShareDetail); ok {
		return output, err
	}

	return nil, err
}

func PortfolioShareDeletedWithToken(conn *servicecatalog.ServiceCatalog, token string) (*servicecatalog.DescribePortfolioShareStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.ShareStatusNotStarted, servicecatalog.ShareStatusInProgress, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.ShareStatusCompleted},
		Refresh: PortfolioShareStatusWithToken(conn, token),
		Timeout: PortfolioShareCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribePortfolioShareStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func OrganizationsAccessStable(conn *servicecatalog.ServiceCatalog) (string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.AccessStatusUnderChange, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.AccessStatusEnabled, servicecatalog.AccessStatusDisabled},
		Refresh: OrganizationsAccessStatus(conn),
		Timeout: OrganizationsAccessStableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.GetAWSOrganizationsAccessStatusOutput); ok {
		return aws.StringValue(output.AccessStatus), err
	}

	return "", err
}

func ConstraintReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) (*servicecatalog.DescribeConstraintOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, servicecatalog.StatusCreating, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   ConstraintStatus(conn, acceptLanguage, id),
		Timeout:                   ConstraintReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeConstraintOutput); ok {
		return output, err
	}

	return nil, err
}

func ConstraintDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable, servicecatalog.StatusCreating},
		Target:  []string{StatusNotFound},
		Refresh: ConstraintStatus(conn, acceptLanguage, id),
		Timeout: ConstraintDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ProductPortfolioAssociationReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) (*servicecatalog.PortfolioDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   ProductPortfolioAssociationStatus(conn, acceptLanguage, portfolioID, productID),
		Timeout:                   ProductPortfolioAssociationReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.PortfolioDetail); ok {
		return output, err
	}

	return nil, err
}

func ProductPortfolioAssociationDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: ProductPortfolioAssociationStatus(conn, acceptLanguage, portfolioID, productID),
		Timeout: ProductPortfolioAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ServiceActionReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) (*servicecatalog.ServiceActionDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: ServiceActionStatus(conn, acceptLanguage, id),
		Timeout: ServiceActionReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.ServiceActionDetail); ok {
		return output, err
	}

	return nil, err
}

func ServiceActionDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: ServiceActionStatus(conn, acceptLanguage, id),
		Timeout: ServiceActionDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	return err
}

func BudgetResourceAssociationReady(conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) (*servicecatalog.BudgetDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: BudgetResourceAssociationStatus(conn, budgetName, resourceID),
		Timeout: BudgetResourceAssociationReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.BudgetDetail); ok {
		return output, err
	}

	return nil, err
}

func BudgetResourceAssociationDeleted(conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: BudgetResourceAssociationStatus(conn, budgetName, resourceID),
		Timeout: BudgetResourceAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func TagOptionResourceAssociationReady(conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) (*servicecatalog.ResourceDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: TagOptionResourceAssociationStatus(conn, tagOptionID, resourceID),
		Timeout: TagOptionResourceAssociationReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.ResourceDetail); ok {
		return output, err
	}

	return nil, err
}

func TagOptionResourceAssociationDeleted(conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: TagOptionResourceAssociationStatus(conn, tagOptionID, resourceID),
		Timeout: TagOptionResourceAssociationDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ProvisioningArtifactReady(conn *servicecatalog.ServiceCatalog, id, productID string) (*servicecatalog.DescribeProvisioningArtifactOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable, StatusCreated},
		Refresh:                   ProvisioningArtifactStatus(conn, id, productID),
		Timeout:                   ProvisioningArtifactReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisioningArtifactOutput); ok {
		return output, err
	}

	return nil, err
}

func ProvisioningArtifactDeleted(conn *servicecatalog.ServiceCatalog, id, productID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: ProvisioningArtifactStatus(conn, id, productID),
		Timeout: ProvisioningArtifactDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func PrincipalPortfolioAssociationReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) (*servicecatalog.Principal, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   PrincipalPortfolioAssociationStatus(conn, acceptLanguage, principalARN, portfolioID),
		Timeout:                   PrincipalPortfolioAssociationReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.Principal); ok {
		return output, err
	}

	return nil, err
}

func PrincipalPortfolioAssociationDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{servicecatalog.StatusAvailable},
		Target:         []string{StatusNotFound, StatusUnavailable},
		Refresh:        PrincipalPortfolioAssociationStatus(conn, acceptLanguage, principalARN, portfolioID),
		Timeout:        PrincipalPortfolioAssociationDeleteTimeout,
		NotFoundChecks: 1,
	}

	_, err := stateConf.WaitForState()

	return err
}

func LaunchPathsReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) ([]*servicecatalog.LaunchPathSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   LaunchPathsStatus(conn, acceptLanguage, productID),
		Timeout:                   LaunchPathsReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*servicecatalog.LaunchPathSummary); ok {
		return output, err
	}

	return nil, err
}

func ProvisionedProductReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string) (*servicecatalog.DescribeProvisionedProductOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable, servicecatalog.ProvisionedProductStatusUnderChange, servicecatalog.ProvisionedProductStatusPlanInProgress},
		Target:                    []string{servicecatalog.StatusAvailable},
		Refresh:                   ProvisionedProductStatus(conn, acceptLanguage, id, name),
		Timeout:                   ProvisionedProductReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisionedProductOutput); ok {
		return output, err
	}

	return nil, err
}

func ProvisionedProductTerminated(conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusAvailable, servicecatalog.ProvisionedProductStatusUnderChange},
		Target:  []string{StatusNotFound, StatusUnavailable},
		Refresh: ProvisionedProductStatus(conn, acceptLanguage, id, name),
		Timeout: ProvisionedProductDeleteTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func RecordReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) (*servicecatalog.DescribeRecordOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{StatusNotFound, StatusUnavailable, servicecatalog.ProvisionedProductStatusUnderChange, servicecatalog.ProvisionedProductStatusPlanInProgress},
		Target:                    []string{servicecatalog.RecordStatusSucceeded, servicecatalog.StatusAvailable},
		Refresh:                   RecordStatus(conn, acceptLanguage, id),
		Timeout:                   RecordReadyTimeout,
		ContinuousTargetOccurence: ContinuousTargetOccurrence,
		NotFoundChecks:            NotFoundChecks,
		MinTimeout:                MinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeRecordOutput); ok {
		return output, err
	}

	return nil, err
}

func PortfolioConstraintsReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) ([]*servicecatalog.ConstraintDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound},
		Target:  []string{servicecatalog.StatusAvailable},
		Refresh: PortfolioConstraintsStatus(conn, acceptLanguage, portfolioID, productID),
		Timeout: PortfolioConstraintsReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*servicecatalog.ConstraintDetail); ok {
		return output, err
	}

	return nil, err
}
