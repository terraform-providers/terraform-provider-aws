package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

// sweeperAwsClients is a shared cache of regional AWSClient
// This prevents client re-initialization for every resource with no benefit.
var sweeperAwsClients map[string]interface{}

func TestMain(m *testing.M) {
	sweeperAwsClients = make(map[string]interface{})
	resource.TestMain(m)
}

// sharedClientForRegion returns a common AWSClient setup needed for the sweeper
// functions for a given region
func sharedClientForRegion(region string) (interface{}, error) {
	if client, ok := sweeperAwsClients[region]; ok {
		return client, nil
	}

	_, _, err := envvar.RequireOneOf([]string{envvar.AwsProfile, envvar.AwsAccessKeyId, envvar.AwsContainerCredentialsFullUri}, "credentials for running sweepers")
	if err != nil {
		return nil, err
	}

	if os.Getenv(envvar.AwsAccessKeyId) != "" {
		_, err := envvar.Require(envvar.AwsSecretAccessKey, "static credentials value when using "+envvar.AwsAccessKeyId)
		if err != nil {
			return nil, err
		}
	}

	conf := &Config{
		MaxRetries: 5,
		Region:     region,
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.Client()
	if err != nil {
		return nil, fmt.Errorf("error getting AWS client")
	}

	sweeperAwsClients[region] = client

	return client, nil
}

type testSweepResource struct {
	d        *schema.ResourceData
	meta     interface{}
	resource *schema.Resource
}

func NewTestSweepResource(resource *schema.Resource, d *schema.ResourceData, meta interface{}) *testSweepResource {
	return &testSweepResource{
		d:        d,
		meta:     meta,
		resource: resource,
	}
}

func testSweepResourceOrchestrator(sweepResources []*testSweepResource) error {
	var g multierror.Group

	for _, sweepResource := range sweepResources {
		sweepResource := sweepResource

		g.Go(func() error {
			return testAccDeleteResource(sweepResource.resource, sweepResource.d, sweepResource.meta)
		})
	}

	return g.Wait().ErrorOrNil()
}

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func testSweepSkipSweepError(err error) bool {
	// Ignore missing API endpoints
	if isAWSErr(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if isAWSErr(err, "UnsupportedOperation", "") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if isAWSErr(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if isAWSErr(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// GovCloud has endpoints that respond with (no message provided):
	// AccessDeniedException:
	// Since acceptance test sweepers are best effort and this response is very common,
	// we allow bypassing this error globally instead of individual test sweeper fixes.
	if isAWSErr(err, "AccessDeniedException", "") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if isAWSErr(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: The action DescribeTransitGatewayAttachments is not valid for this web service
	if isAWSErr(err, "InvalidAction", "is not valid") {
		return true
	}
	// For example from GovCloud SES.SetActiveReceiptRuleSet.
	if isAWSErr(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	return false
}

// Check sweeper API call error for reasons to skip a specific resource
// These include AccessDenied or AccessDeniedException for individual resources, e.g. managed by central IT
func testSweepSkipResourceError(err error) bool {
	// Since acceptance test sweepers are best effort, we allow bypassing this error globally
	// instead of individual test sweeper fixes.
	return tfawserr.ErrCodeContains(err, "AccessDenied")
}
