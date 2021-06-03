package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppFlowConnectorProfile_Amplitude(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.amplitude"
	connectorType := "Amplitude"

	apiKey := os.Getenv("AMPLITUDE_API_KEY")
	secretKey := os.Getenv("AMPLITUDE_SECRET_KEY")

	if apiKey == "" || secretKey == "" {
		t.Skip("All environment variables: AMPLITUDE_API_KEY, AMPLITUDE_SECRET_KEY must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Amplitude(connectorProfileName, apiKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Amplitude"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.amplitude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.amplitude.0.api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.amplitude.0.secret_key", secretKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.amplitude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAWSAppFlowConnectorProfile_Datadog(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.datadog"
	connectorType := "Datadog"

	apiKey := os.Getenv("DATADOG_API_KEY")
	applicationKey := os.Getenv("DATADOG_APPLICATION_KEY")
	instanceUrl := os.Getenv("DATADOG_INSTANCE_URL")

	if apiKey == "" || applicationKey == "" || instanceUrl == "" {
		t.Skip("All environment variables: DATADOG_API_KEY, DATADOG_APPLICATION_KEY, DATADOG_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Datadog(connectorProfileName, apiKey, applicationKey, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Datadog"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.datadog.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.datadog.0.api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.datadog.0.application_key", applicationKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.datadog.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.datadog.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAWSAppFlowConnectorProfile_Dynatrace(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.dynatrace"
	connectorType := "Dynatrace"

	apiToken := os.Getenv("DYNATRACE_API_TOKEN")
	instanceUrl := os.Getenv("DYNATRACE_INSTANCE_URL")

	if apiToken == "" || instanceUrl == "" {
		t.Skip("All environment variables: DYNATRACE_API_TOKEN, DYNATRACE_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Dynatrace(connectorProfileName, apiToken, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Dynatrace"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.dynatrace.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.dynatrace.0.api_token", apiToken),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.dynatrace.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.dynatrace.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAWSAppFlowConnectorProfile_Slack(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.slack"
	connectorType := "Slack"

	clientId := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	accessToken := os.Getenv("SLACK_ACCESS_TOKEN")
	instanceUrl := os.Getenv("SLACK_INSTANCE_URL")

	if clientId == "" || clientSecret == "" || accessToken == "" || instanceUrl == "" {
		t.Skip("All environment variables: SLACK_CLIENT_ID, SLACK_CLIENT_SECRET, SLACK_ACCESS_TOKEN, SLACK_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Slack(connectorProfileName, clientId, clientSecret, accessToken, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Slack"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.access_token", accessToken),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.client_id", clientId),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.client_secret", clientSecret),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.oauth_request.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_credentials.0.slack.0.oauth_request.0.redirect_uri"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.slack.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.slack.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAWSAppFlowConnectorProfile_Snowflake(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.snowflake"
	connectorType := "Snowflake"

	password := os.Getenv("SNOWFLAKE_PASSWORD")
	username := os.Getenv("SNOWFLAKE_USERNAME")
	accountName := os.Getenv("SNOWFLAKE_ACCOUNT_NAME")
	bucketName := connectorProfileName
	region := os.Getenv("SNOWFLAKE_REGION")
	stage := os.Getenv("SNOWFLAKE_STAGE")
	warehouse := os.Getenv("SNOWFLAKE_WAREHOUSE")

	if password == "" || username == "" || accountName == "" || region == "" || stage == "" || warehouse == "" {
		t.Skip("All environment variables: SNOWFLAKE_PASSWORD, SNOWFLAKE_USERNAME, SNOWFLAKE_ACCOUNT_NAME, SNOWFLAKE_REGION, SNOWFLAKE_STAGE, SNOWFLAKE_WAREHOUSE must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Snowflake(connectorProfileName, password, username, accountName, bucketName, region, stage, warehouse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Snowflake"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.snowflake.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.snowflake.0.password", password),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.snowflake.0.username", username),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.account_name", accountName),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.region", region),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.stage"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.snowflake.0.warehouse", warehouse),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAWSAppFlowConnectorProfile_Trendmicro(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.trendmicro"
	connectorType := "Trendmicro"

	apiSecretKey := os.Getenv("TRENDMICRO_API_SECRET_KEY")

	if apiSecretKey == "" {
		t.Skip("All environment variables: TRENDMICRO_API_SECRET_KEY must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Trendmicro(connectorProfileName, apiSecretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Trendmicro"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.trendmicro.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.trendmicro.0.api_secret_key", apiSecretKey),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func TestAccAWSAppFlowConnectorProfile_Zendesk(t *testing.T) {
	var connectorProfiles appflow.DescribeConnectorProfilesOutput

	connectorProfileName := acctest.RandomWithPrefix("tf-acc-test-appflow-connector-profile")
	resourceName := "aws_appflow_connector_profile.zendesk"
	connectorType := "Zendesk"

	clientId := os.Getenv("ZENDESK_CLIENT_ID")
	clientSecret := os.Getenv("ZENDESK_CLIENT_SECRET")
	accessToken := os.Getenv("ZENDESK_ACCESS_TOKEN")
	instanceUrl := os.Getenv("ZENDESK_INSTANCE_URL")

	if clientId == "" || clientSecret == "" || accessToken == "" || instanceUrl == "" {
		t.Skip("All environment variables: ZENDESK_CLIENT_ID, ZENDESK_CLIENT_SECRET, ZENDESK_ACCESS_TOKEN, ZENDESK_INSTANCE_URL must be set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowConnectorProfile_Zendesk(connectorProfileName, clientId, clientSecret, accessToken, instanceUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowConnectorProfileExists(resourceName, connectorType, &connectorProfiles),
					resource.TestCheckResourceAttr(resourceName, "connection_mode", "Public"),
					testAccCheckResourceAttrRegionalARN(resourceName, "connector_profile_arn", "appflow", fmt.Sprintf("connectorprofile/%s", connectorProfileName)),
					resource.TestCheckResourceAttr(resourceName, "connector_type", "Zendesk"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.access_token", accessToken),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.client_id", clientId),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.client_secret", clientSecret),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.oauth_request.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "connector_profile_config.0.connector_profile_credentials.0.zendesk.0.oauth_request.0.redirect_uri"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.zendesk.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_config.0.connector_profile_properties.0.zendesk.0.instance_url", instanceUrl),
					resource.TestCheckResourceAttr(resourceName, "connector_profile_name", connectorProfileName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connector_profile_config.0.connector_profile_credentials"},
			},
		},
	})
}

func testAccCheckAWSAppFlowConnectorProfileExists(n string, connectorType string, res *appflow.DescribeConnectorProfilesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appflowconn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		req := &appflow.DescribeConnectorProfilesInput{
			ConnectorProfileNames: []*string{aws.String(rs.Primary.Attributes["connector_profile_name"])},
			ConnectorType:         aws.String(connectorType),
		}
		describe, err := conn.DescribeConnectorProfiles(req)

		if len(describe.ConnectorProfileDetails) == 0 {
			return fmt.Errorf("AppFlow Connector profile %s does not exist.", n)
		}

		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccAWSAppFlowConnectorProfile_Amplitude(connectorProfileName string, apiKey string, secretKey string) string {
	return fmt.Sprintf(`
resource "aws_appflow_connector_profile" "amplitude" {
  connector_profile_name = %[1]q
  connector_type         = "Amplitude"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      amplitude {
        api_key    = %[2]q
        secret_key = %[3]q
      }
    }
    connector_profile_properties {
      amplitude {
      }
    }
  }

}
`, connectorProfileName, apiKey, secretKey)
}

func testAccAWSAppFlowConnectorProfile_Datadog(connectorProfileName string, apiKey string, applicationKey string, instanceUrl string) string {
	return fmt.Sprintf(`
resource "aws_appflow_connector_profile" "datadog" {
  connector_profile_name = %[1]q
  connector_type         = "Datadog"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      datadog {
        api_key         = %[2]q
        application_key = %[3]q
      }
    }
    connector_profile_properties {
      datadog {
        instance_url = %[4]q
      }
    }
  }

}
`, connectorProfileName, apiKey, applicationKey, instanceUrl)
}

func testAccAWSAppFlowConnectorProfile_Dynatrace(connectorProfileName string, apiToken string, instanceUrl string) string {
	return fmt.Sprintf(`
resource "aws_appflow_connector_profile" "dynatrace" {
  connector_profile_name = %[1]q
  connector_type         = "Dynatrace"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      dynatrace {
        api_token = %[2]q
      }
    }
    connector_profile_properties {
      dynatrace {
        instance_url = %[3]q
      }
    }
  }

}
`, connectorProfileName, apiToken, instanceUrl)
}

func testAccAWSAppFlowConnectorProfile_Slack(connectorProfileName string, clientId string, clientSecret string, accessToken string, instanceUrl string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "slack" {
  connection_mode        = "Public"
  connector_profile_name = %[1]q
  connector_type         = "Slack"
  connector_profile_config {
    connector_profile_credentials {
      slack {
        client_id     = %[2]q
        client_secret = %[3]q
        access_token  = %[4]q
        oauth_request {
          redirect_uri = "https://${data.aws_region.current.name}.console.aws.amazon.com/appflow/oauth"
        }
      }
    }
    connector_profile_properties {
      slack {
        instance_url = %[5]q
      }
    }
  }
}
`, connectorProfileName, clientId, clientSecret, accessToken, instanceUrl)
}

func testAccAWSAppFlowConnectorProfile_Snowflake(connectorProfileName string, password string, username string, accountName string, bucketName string, region string, stage string, warehouse string) string {

	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_s3_bucket" "snowflake" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_appflow_connector_profile" "snowflake" {
  connector_profile_name = %[2]q
  connector_type         = "Snowflake"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      snowflake {
        password = %[3]q
        username = %[4]q
      }
    }
    connector_profile_properties {
      snowflake {
        account_name = %[5]q
        bucket_name  = aws_s3_bucket.snowflake.id
        region       = %[6]q
        stage        = %[7]q
        warehouse    = %[8]q
      }
    }
  }

}
`, bucketName, connectorProfileName, password, username, accountName, region, stage, warehouse)
}

func testAccAWSAppFlowConnectorProfile_Trendmicro(connectorProfileName string, apiSecretKey string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "trendmicro" {
  connector_profile_name = %[1]q
  connector_type         = "Trendmicro"
  connection_mode        = "Public"
  connector_profile_config {
    connector_profile_credentials {
      trendmicro {
        api_secret_key = %[2]q
      }
    }
    connector_profile_properties {
      trendmicro {
      }
    }
  }

}
`, connectorProfileName, apiSecretKey)
}

func testAccAWSAppFlowConnectorProfile_Zendesk(connectorProfileName string, clientId string, clientSecret string, accessToken string, instanceUrl string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "zendesk" {
  connection_mode        = "Public"
  connector_profile_name = %[1]q
  connector_type         = "Zendesk"
  connector_profile_config {
    connector_profile_credentials {
      zendesk {
        client_id     = %[2]q
        client_secret = %[3]q
        access_token  = %[4]q
        oauth_request {
          redirect_uri = "https://${data.aws_region.current.name}.console.aws.amazon.com/appflow/oauth"
        }
      }
    }
    connector_profile_properties {
      zendesk {
        instance_url = %[5]q
      }
    }
  }
}
`, connectorProfileName, clientId, clientSecret, accessToken, instanceUrl)
}
