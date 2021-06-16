package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func AppByID(conn *amplify.Amplify, id string) (*amplify.App, error) {
	input := &amplify.GetAppInput{
		AppId: aws.String(id),
	}

	output, err := conn.GetApp(input)

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.App == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.App, nil
}

func BackendEnvironmentByAppIDAndEnvironmentName(conn *amplify.Amplify, appID, environmentName string) (*amplify.BackendEnvironment, error) {
	input := &amplify.GetBackendEnvironmentInput{
		AppId:           aws.String(appID),
		EnvironmentName: aws.String(environmentName),
	}

	output, err := conn.GetBackendEnvironment(input)

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BackendEnvironment == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.BackendEnvironment, nil
}

func BranchByAppIDAndBranchName(conn *amplify.Amplify, appID, branchName string) (*amplify.Branch, error) {
	input := &amplify.GetBranchInput{
		AppId:      aws.String(appID),
		BranchName: aws.String(branchName),
	}

	output, err := conn.GetBranch(input)

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Branch == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Branch, nil
}

func DomainAssociationByAppIDAndDomainName(conn *amplify.Amplify, appID, domainName string) (*amplify.DomainAssociation, error) {
	input := &amplify.GetDomainAssociationInput{
		AppId:      aws.String(appID),
		DomainName: aws.String(domainName),
	}

	output, err := conn.GetDomainAssociation(input)

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainAssociation == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.DomainAssociation, nil
}

func WebhookByID(conn *amplify.Amplify, id string) (*amplify.Webhook, error) {
	input := &amplify.GetWebhookInput{
		WebhookId: aws.String(id),
	}

	output, err := conn.GetWebhook(input)

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Webhook == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Webhook, nil
}
