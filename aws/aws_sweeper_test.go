package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	if os.Getenv("AWS_PROFILE") == "" && (os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "") {
		return nil, fmt.Errorf("must provide environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or environment variable AWS_PROFILE")
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
