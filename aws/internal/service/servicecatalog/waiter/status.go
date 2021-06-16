package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/finder"
)

func ProductStatus(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(productID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeProductAsAdmin(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceInUseException) {
			return nil, StatusUnavailable, err
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeLimitExceededException) {
			return nil, StatusUnavailable, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing product status: %w", err)
		}

		if output == nil || output.ProductViewDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing product status: empty product view detail")
		}

		return output, aws.StringValue(output.ProductViewDetail.Status), err
	}
}

func TagOptionStatus(conn *servicecatalog.ServiceCatalog, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(id),
		}

		output, err := conn.DescribeTagOption(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing tag option: %w", err)
		}

		if output == nil || output.TagOptionDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing tag option: empty tag option detail")
		}

		return output.TagOptionDetail, servicecatalog.StatusAvailable, err
	}
}

func PortfolioShareStatusWithToken(conn *servicecatalog.ServiceCatalog, token string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribePortfolioShareStatusInput{
			PortfolioShareToken: aws.String(token),
		}
		output, err := conn.DescribePortfolioShareStatus(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.ShareStatusError, fmt.Errorf("error describing portfolio share status: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing portfolio share status: empty response")
		}

		return output, aws.StringValue(output.Status), err
	}
}

func PortfolioShareStatus(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.PortfolioShare(conn, portfolioID, shareType, principalID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.ShareStatusError, fmt.Errorf("error finding portfolio share: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("error finding portfolio share (%s:%s:%s): empty response", portfolioID, shareType, principalID),
			}
		}

		if !aws.BoolValue(output.Accepted) {
			return output, servicecatalog.ShareStatusInProgress, err
		}

		return output, servicecatalog.ShareStatusCompleted, err
	}
}

func OrganizationsAccessStatus(conn *servicecatalog.ServiceCatalog) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.GetAWSOrganizationsAccessStatusInput{}

		output, err := conn.GetAWSOrganizationsAccessStatus(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {

			return nil, OrganizationAccessStatusError, fmt.Errorf("error getting Organizations Access: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("error getting Organizations Access: empty response")
		}

		return output, aws.StringValue(output.AccessStatus), err
	}
}

func ConstraintStatus(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeConstraintInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeConstraint(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("constraint not found (accept language %s, ID: %s): %s", acceptLanguage, id, err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing constraint: %w", err)
		}

		if output == nil || output.ConstraintDetail == nil {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("describing constraint (accept language %s, ID: %s): empty response", acceptLanguage, id),
			}
		}

		return output, aws.StringValue(output.Status), err
	}
}

func ProductPortfolioAssociationStatus(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ProductPortfolioAssociation(conn, acceptLanguage, portfolioID, productID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("product portfolio association not found (%s): %s", tfservicecatalog.ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID), err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing product portfolio association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("finding product portfolio association (%s): empty response", tfservicecatalog.ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID)),
			}
		}

		return output, servicecatalog.StatusAvailable, err
	}
}

func ServiceActionStatus(conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeServiceAction(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing Service Action: %w", err)
		}

		if output == nil || output.ServiceActionDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing Service Action: empty Service Action Detail")
		}

		return output.ServiceActionDetail, servicecatalog.StatusAvailable, nil
	}
}

func BudgetResourceAssociationStatus(conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.BudgetResourceAssociation(conn, budgetName, resourceID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", tfservicecatalog.BudgetResourceAssociationID(budgetName, resourceID), err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", tfservicecatalog.BudgetResourceAssociationID(budgetName, resourceID)),
			}
		}

		return output, servicecatalog.StatusAvailable, err
	}
}

func TagOptionResourceAssociationStatus(conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.TagOptionResourceAssociation(conn, tagOptionID, resourceID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", tfservicecatalog.TagOptionResourceAssociationID(tagOptionID, resourceID), err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", tfservicecatalog.TagOptionResourceAssociationID(tagOptionID, resourceID)),
			}
		}

		return output, servicecatalog.StatusAvailable, err
	}
}

func ProvisioningArtifactStatus(conn *servicecatalog.ServiceCatalog, id, productID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProvisioningArtifactId: aws.String(id),
			ProductId:              aws.String(productID),
		}

		output, err := conn.DescribeProvisioningArtifact(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, err
		}

		if output == nil || output.ProvisioningArtifactDetail == nil {
			return nil, StatusUnavailable, err
		}

		return output, aws.StringValue(output.Status), err
	}
}

func PrincipalPortfolioAssociationStatus(conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.PrincipalPortfolioAssociation(conn, acceptLanguage, principalARN, portfolioID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing principal portfolio association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, err
		}

		return output, servicecatalog.StatusAvailable, err
	}
}
