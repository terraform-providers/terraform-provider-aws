package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceCloudwatch_Event_Connection_basic(t *testing.T) {
	dataSourceName := "data.aws_cloudwatch_event_connection.test"
	resourceName := "aws_cloudwatch_event_connection.api_key"

	name := acctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix("tf-acc-test")
	key := acctest.RandomWithPrefix("tf-acc-test")
	value := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatch_Event_ConnectionDataConfig(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "secret_arn", resourceName, "secret_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorization_type", resourceName, "authorization_type"),
				),
			},
		},
	})
}

func testAccAWSCloudwatch_Event_ConnectionDataConfig(name, description, authorizationType, key, value string) string {
	return composeConfig(
		testAccAWSCloudWatchEventConnectionConfig_apiKey(name, description, authorizationType, key, value),
		`
data "aws_cloudwatch_event_connection" "test" {
  name = aws_cloudwatch_event_connection.api_key.name
}
`)
}
