package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAppsyncApiKey_basic(t *testing.T) {
	var apiKey appsync.ApiKey
	dateAfterSevenDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(7)).Truncate(time.Hour)
	resourceName := "aws_appsync_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncApiKeyConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					testAccCheckAwsAppsyncApiKeyExpiresDate(&apiKey, dateAfterSevenDays),
					resource.TestMatchResourceAttr(resourceName, "key", regexp.MustCompile(`.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAppsyncApiKey_Description(t *testing.T) {
	var apiKey appsync.ApiKey
	resourceName := "aws_appsync_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncApiKeyConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAppsyncApiKeyConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAppsyncApiKey_Expires(t *testing.T) {
	var apiKey appsync.ApiKey
	dateAfterTenDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(10)).Truncate(time.Hour)
	dateAfterTwentyDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(20)).Truncate(time.Hour)
	resourceName := "aws_appsync_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncApiKeyConfig_Expires(rName, dateAfterTenDays.Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExists(resourceName, &apiKey),
					testAccCheckAwsAppsyncApiKeyExpiresDate(&apiKey, dateAfterTenDays),
				),
			},
			{
				Config: testAccAppsyncApiKeyConfig_Expires(rName, dateAfterTwentyDays.Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExists(resourceName, &apiKey),
					testAccCheckAwsAppsyncApiKeyExpiresDate(&apiKey, dateAfterTwentyDays),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsAppsyncApiKeyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_api_key" {
			continue
		}

		apiID, keyID, err := decodeAppSyncApiKeyId(rs.Primary.ID)
		if err != nil {
			return err
		}

		apiKey, err := getAppsyncApiKey(apiID, keyID, conn)
		if err == nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		if apiKey != nil && aws.StringValue(apiKey.Id) == keyID {
			return fmt.Errorf("Appsync API Key ID %q still exists", rs.Primary.ID)
		}

		return nil

	}
	return nil
}

func testAccCheckAwsAppsyncApiKeyExists(resourceName string, apiKey *appsync.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync API Key Not found in state: %s", resourceName)
		}

		apiID, keyID, err := decodeAppSyncApiKeyId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn
		key, err := getAppsyncApiKey(apiID, keyID, conn)
		if err != nil {
			return err
		}

		if key == nil || key.Id == nil {
			return fmt.Errorf("Appsync API Key %q not found", rs.Primary.ID)
		}

		*apiKey = *key

		return nil
	}
}

func testAccCheckAwsAppsyncApiKeyExpiresDate(apiKey *appsync.ApiKey, expectedTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiKeyExpiresTime := time.Unix(aws.Int64Value(apiKey.Expires), 0)
		if !apiKeyExpiresTime.Equal(expectedTime) {
			return fmt.Errorf("Appsync API Key expires difference: got %s and expected %s", apiKeyExpiresTime.Format(time.RFC3339), expectedTime.Format(time.RFC3339))
		}

		return nil
	}
}

func testAccAppsyncApiKeyConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}
resource "aws_appsync_api_key" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
  description = %q
}

`, rName, description)
}

func testAccAppsyncApiKeyConfig_Expires(rName, expires string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}
resource "aws_appsync_api_key" "test" {
  api_id  = "${aws_appsync_graphql_api.test.id}"
  expires = %q
}

`, rName, expires)
}

func testAccAppsyncApiKeyConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}
resource "aws_appsync_api_key" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
}

`, rName)
}
