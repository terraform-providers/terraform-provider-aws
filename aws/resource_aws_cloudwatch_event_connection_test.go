package aws

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_connection", &resource.Sweeper{
		Name: "aws_cloudwatch_event_connection",
		F:    testSweepCloudWatchEventConnection,
	})
}

func testSweepCloudWatchEventConnection(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	var sweeperErrs *multierror.Error

	input := &events.ListConnectionsInput{
		Limit: aws.Int64(100),
	}
	var connections []*events.Connection
	for {
		output, err := conn.ListConnections(input)
		if err != nil {
			return err
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, connection := range connections {
		input := &events.DeleteConnectionInput{
			Name: connection.Name,
		}
		_, err := conn.DeleteConnection(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting CloudWatch Event Connection (%s): %w", *connection.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d CloudWatch Event Connections", len(connections))

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudWatchEventConnection_apiKey(t *testing.T) {
	var v1, v2, v3 events.DescribeConnectionOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix("tf-acc-test")
	key := acctest.RandomWithPrefix("tf-acc-test")
	value := acctest.RandomWithPrefix("tf-acc-test")

	nameModified := acctest.RandomWithPrefix("tf-acc-test")
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")
	keyModified := acctest.RandomWithPrefix("tf-acc-test")
	valueModified := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_connection.api_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_apiKey(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_parameters.0.api_key.0.value"},
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_apiKey(
					nameModified,
					descriptionModified,
					authorizationType,
					keyModified,
					valueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckCloudWatchEventConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", keyModified),
				),
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_apiKey(
					nameModified,
					descriptionModified,
					authorizationType,
					keyModified,
					valueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v3),
					testAccCheckCloudWatchEventConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", keyModified),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventConnection_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeConnectionOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "BASIC"
	description := acctest.RandomWithPrefix("tf-acc-test")
	username := acctest.RandomWithPrefix("tf-acc-test")
	password := acctest.RandomWithPrefix("tf-acc-test")

	nameModified := acctest.RandomWithPrefix("tf-acc-test")
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")
	usernameModified := acctest.RandomWithPrefix("tf-acc-test")
	passwordModified := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_connection.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_basic(
					name,
					description,
					authorizationType,
					username,
					password,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", username),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_parameters.0.basic.0.password"},
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_basic(
					nameModified,
					descriptionModified,
					authorizationType,
					usernameModified,
					passwordModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckCloudWatchEventConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", usernameModified),
				),
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_basic(
					nameModified,
					descriptionModified,
					authorizationType,
					usernameModified,
					passwordModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v3),
					testAccCheckCloudWatchEventConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", usernameModified),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventConnection_oAuth(t *testing.T) {
	var v1, v2, v3 events.DescribeConnectionOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "OAUTH_CLIENT_CREDENTIALS"
	description := acctest.RandomWithPrefix("tf-acc-test")
	// oauth
	authorizationEndpoint := "https://www.hashicorp.com/products/terraform"
	httpMethod := "POST"

	// client_parameters
	clientID := acctest.RandomWithPrefix("tf-acc-test")
	clientSecret := acctest.RandomWithPrefix("tf-acc-test")

	// oauth_http_parameters
	bodyKey := acctest.RandomWithPrefix("tf-acc-test")
	bodyValue := acctest.RandomWithPrefix("tf-acc-test")
	bodyIsSecretValue := true

	headerKey := acctest.RandomWithPrefix("tf-acc-test")
	headerValue := acctest.RandomWithPrefix("tf-acc-test")
	headerIsSecretValue := true

	queryStringKey := acctest.RandomWithPrefix("tf-acc-test")
	queryStringValue := acctest.RandomWithPrefix("tf-acc-test")
	queryStringIsSecretValue := true

	// modified
	nameModified := acctest.RandomWithPrefix("tf-acc-test")
	descriptionModified := acctest.RandomWithPrefix("tf-acc-test")
	// oauth
	authorizationEndpointModified := "https://www.hashicorp.com/"
	httpMethodModified := "GET"

	// client_parameters
	clientIDModified := acctest.RandomWithPrefix("tf-acc-test")
	clientSecretModified := acctest.RandomWithPrefix("tf-acc-test")

	// oauth_http_parameters modified
	bodyKeyModified := acctest.RandomWithPrefix("tf-acc-test")
	bodyValueModified := acctest.RandomWithPrefix("tf-acc-test")
	bodyIsSecretValueModified := false

	headerKeyModified := acctest.RandomWithPrefix("tf-acc-test")
	headerValueModified := acctest.RandomWithPrefix("tf-acc-test")
	headerIsSecretValueModified := false

	queryStringKeyModified := acctest.RandomWithPrefix("tf-acc-test")
	queryStringValueModified := acctest.RandomWithPrefix("tf-acc-test")
	queryStringIsSecretValueModified := false

	resourceName := "aws_cloudwatch_event_connection.oauth"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_oauth(
					name,
					description,
					authorizationType,
					authorizationEndpoint,
					httpMethod,
					clientID,
					clientSecret,
					bodyKey,
					bodyValue,
					bodyIsSecretValue,
					headerKey,
					headerValue,
					headerIsSecretValue,
					queryStringKey,
					queryStringValue,
					queryStringIsSecretValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientID),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auth_parameters.0.oauth.0.client_parameters.0.client_secret",
					"auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.value",
					"auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.value",
					"auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.value",
				},
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_oauth(
					nameModified,
					descriptionModified,
					authorizationType,
					authorizationEndpointModified,
					httpMethodModified,
					clientIDModified,
					clientSecretModified,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckCloudWatchEventConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientIDModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_oauth(
					nameModified,
					descriptionModified,
					authorizationType,
					authorizationEndpointModified,
					httpMethodModified,
					clientIDModified,
					clientSecretModified,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v3),
					testAccCheckCloudWatchEventConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientIDModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventConnection_invocationHttpParameters(t *testing.T) {
	var v1, v2, v3 events.DescribeConnectionOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix("tf-acc-test")
	key := acctest.RandomWithPrefix("tf-acc-test")
	value := acctest.RandomWithPrefix("tf-acc-test")

	// invocation_http_parameters
	bodyKey := acctest.RandomWithPrefix("tf-acc-test")
	bodyValue := acctest.RandomWithPrefix("tf-acc-test")
	bodyIsSecretValue := true

	headerKey := acctest.RandomWithPrefix("tf-acc-test")
	headerValue := acctest.RandomWithPrefix("tf-acc-test")
	headerIsSecretValue := true

	queryStringKey := acctest.RandomWithPrefix("tf-acc-test")
	queryStringValue := acctest.RandomWithPrefix("tf-acc-test")
	queryStringIsSecretValue := true

	// invocation_http_parameters modified
	bodyKeyModified := acctest.RandomWithPrefix("tf-acc-test")
	bodyValueModified := acctest.RandomWithPrefix("tf-acc-test")
	bodyIsSecretValueModified := false

	headerKeyModified := acctest.RandomWithPrefix("tf-acc-test")
	headerValueModified := acctest.RandomWithPrefix("tf-acc-test")
	headerIsSecretValueModified := false

	queryStringKeyModified := acctest.RandomWithPrefix("tf-acc-test")
	queryStringValueModified := acctest.RandomWithPrefix("tf-acc-test")
	queryStringIsSecretValueModified := false

	resourceName := "aws_cloudwatch_event_connection.invocation_http_parameters"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_invocationHttpParameters(
					name,
					description,
					authorizationType,
					key,
					value,
					bodyKey,
					bodyValue,
					bodyIsSecretValue,
					headerKey,
					headerValue,
					headerIsSecretValue,
					queryStringKey,
					queryStringValue,
					queryStringIsSecretValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.key", bodyKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.key", fmt.Sprintf("second-%s", bodyKey)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.key", fmt.Sprintf("second-%s", headerKey)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.key", fmt.Sprintf("second-%s", queryStringKey)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auth_parameters.0.api_key.0.value",
					"auth_parameters.0.invocation_http_parameters.0.body.0.value",
					"auth_parameters.0.invocation_http_parameters.0.body.1.value",
					"auth_parameters.0.invocation_http_parameters.0.header.0.value",
					"auth_parameters.0.invocation_http_parameters.0.header.1.value",
					"auth_parameters.0.invocation_http_parameters.0.query_string.0.value",
					"auth_parameters.0.invocation_http_parameters.0.query_string.1.value",
				},
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_invocationHttpParameters(
					name,
					description,
					authorizationType,
					key,
					value,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v2),
					testAccCheckCloudWatchEventConnectionNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.key", fmt.Sprintf("second-%s", bodyKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.key", fmt.Sprintf("second-%s", headerKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.key", fmt.Sprintf("second-%s", queryStringKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_invocationHttpParameters(
					name,
					description,
					authorizationType,
					key,
					value,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v3),
					testAccCheckCloudWatchEventConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.key", fmt.Sprintf("second-%s", bodyKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.key", fmt.Sprintf("second-%s", headerKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.key", fmt.Sprintf("second-%s", queryStringKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventConnection_disappears(t *testing.T) {
	var v events.DescribeConnectionOutput
	name := acctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix("tf-acc-test")
	key := acctest.RandomWithPrefix("tf-acc-test")
	value := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_connection.api_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventConnectionConfig_apiKey(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventConnectionExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchEventConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloudWatchEventConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_connection" {
			continue
		}

		_, err := finder.ConnectionByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudWatch Events connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudWatchEventConnectionExists(n string, v *events.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

		output, err := finder.ConnectionByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCloudWatchEventConnectionRecreated(i, j *events.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ConnectionArn) == aws.StringValue(j.ConnectionArn) {
			return fmt.Errorf("CloudWatch Events Connection not recreated")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventConnectionNotRecreated(i, j *events.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ConnectionArn) != aws.StringValue(j.ConnectionArn) {
			return fmt.Errorf("CloudWatch Events Connection was recreated")
		}
		return nil
	}
}

func testAccAWSCloudWatchEventConnectionConfig_apiKey(name, description, authorizationType, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "api_key" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    api_key {
      key   = %[4]q
      value = %[5]q
    }
  }
}
`, name,
		description,
		authorizationType,
		key,
		value)
}

func testAccAWSCloudWatchEventConnectionConfig_basic(name, description, authorizationType, username, password string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "basic" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    basic {
      username = %[4]q
      password = %[5]q
    }
  }
}
`, name,
		description,
		authorizationType,
		username,
		password)
}

func testAccAWSCloudWatchEventConnectionConfig_oauth(
	name,
	description,
	authorizationType,
	authorizationEndpoint,
	httpMethod,
	clientID,
	clientSecret string,
	bodyKey string,
	bodyValue string,
	bodyIsSecretValue bool,
	headerKey string,
	headerValue string,
	headerIsSecretValue bool,
	queryStringKey string,
	queryStringValue string,
	queryStringIsSecretValue bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "oauth" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    oauth {
      authorization_endpoint = %[4]q
      http_method            = %[5]q
      client_parameters {
        client_id     = %[6]q
        client_secret = %[7]q
      }

      oauth_http_parameters {
        body {
          key             = %[8]q
          value           = %[9]q
          is_value_secret = %[10]t
        }

        header {
          key             = %[11]q
          value           = %[12]q
          is_value_secret = %[13]t
        }

        query_string {
          key             = %[14]q
          value           = %[15]q
          is_value_secret = %[16]t
        }
      }
    }
  }
}
`, name,
		description,
		authorizationType,
		authorizationEndpoint,
		httpMethod,
		clientID,
		clientSecret,
		bodyKey,
		bodyValue,
		bodyIsSecretValue,
		headerKey,
		headerValue,
		headerIsSecretValue,
		queryStringKey,
		queryStringValue,
		queryStringIsSecretValue)
}

func testAccAWSCloudWatchEventConnectionConfig_invocationHttpParameters(
	name,
	description,
	authorizationType,
	key,
	value string,
	bodyKey string,
	bodyValue string,
	bodyIsSecretValue bool,
	headerKey string,
	headerValue string,
	headerIsSecretValue bool,
	queryStringKey string,
	queryStringValue string,
	queryStringIsSecretValue bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "invocation_http_parameters" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    api_key {
      key   = %[4]q
      value = %[5]q
    }

    invocation_http_parameters {
      body {
        key             = %[6]q
        value           = %[7]q
        is_value_secret = %[8]t
      }

      body {
        key             = "second-%[6]s"
        value           = "second-%[7]s"
        is_value_secret = %[8]t
      }

      header {
        key             = %[9]q
        value           = %[10]q
        is_value_secret = %[11]t
      }

      header {
        key             = "second-%[9]s"
        value           = "second-%[10]s"
        is_value_secret = %[11]t
      }

      query_string {
        key             = %[12]q
        value           = %[13]q
        is_value_secret = %[14]t
      }

      query_string {
        key             = "second-%[12]s"
        value           = "second-%[13]s"
        is_value_secret = %[14]t
      }
    }
  }
}
`, name,
		description,
		authorizationType,
		key,
		value,
		bodyKey,
		bodyValue,
		bodyIsSecretValue,
		headerKey,
		headerValue,
		headerIsSecretValue,
		queryStringKey,
		queryStringValue,
		queryStringIsSecretValue)
}
