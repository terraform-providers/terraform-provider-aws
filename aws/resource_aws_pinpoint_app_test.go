package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_pinpoint_app", &resource.Sweeper{
		Name: "aws_pinpoint_app",
		F:    testSweepPinpointApps,
	})
}

func testSweepPinpointApps(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).pinpointconn

	input := &pinpoint.GetAppsInput{}

	for {
		output, err := conn.GetApps(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping Pinpoint app sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Pinpoint apps: %s", err)
		}

		if len(output.ApplicationsResponse.Item) == 0 {
			log.Print("[DEBUG] No Pinpoint apps to sweep")
			return nil
		}

		for _, item := range output.ApplicationsResponse.Item {
			name := aws.StringValue(item.Name)

			log.Printf("[INFO] Deleting Pinpoint app %s", name)
			_, err := conn.DeleteApp(&pinpoint.DeleteAppInput{
				ApplicationId: item.Id,
			})
			if err != nil {
				return fmt.Errorf("Error deleting Pinpoint app %s: %s", name, err)
			}
		}

		if output.ApplicationsResponse.NextToken == nil {
			break
		}
		input.Token = output.ApplicationsResponse.NextToken
	}

	return nil
}

func TestAccAWSPinpointApp_basic(t *testing.T) {
	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_withGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
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

func TestAccAWSPinpointApp_CampaignHookLambda(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_CampaignHookLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "campaign_hook.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "campaign_hook.0.mode", "DELIVERY"),
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

func TestAccAWSPinpointApp_Limits(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_Limits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "limits.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "limits.0.total", "100"),
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

func TestAccAWSPinpointApp_QuietTime(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_QuietTime(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "quiet_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quiet_time.0.start", "00:00"),
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

func TestAccAWSPinpointApp_Tags(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_Tag1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSPinpointAppConfig_Tag2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSPinpointAppConfig_Tag1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccPreCheckAWSPinpointApp(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	input := &pinpoint.GetAppsInput{}

	_, err := conn.GetApps(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSPinpointAppExists(n string, application *pinpoint.ApplicationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint app with that ID exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).pinpointconn

		// Check if the app exists
		params := &pinpoint.GetAppInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApp(params)

		if err != nil {
			return err
		}

		*application = *output.ApplicationResponse

		return nil
	}
}

const testAccAWSPinpointAppConfig_withGeneratedName = `
resource "aws_pinpoint_app" "test" {}
`

func testAccAWSPinpointAppConfig_CampaignHookLambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  campaign_hook {
    lambda_function_name = aws_lambda_function.test.arn
    mode                 = "DELIVERY"
  }

  depends_on = [aws_lambda_permission.test]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdapinpoint.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambdapinpoint.handler"
  runtime       = "nodejs12.x"
  publish       = true
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "aws" {}

data "aws_region" "current" {}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromPinpoint"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "pinpoint.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  source_arn    = "arn:${data.aws_partition.current.partition}:mobiletargeting:${data.aws_region.current.name}:${data.aws_caller_identity.aws.account_id}:/apps/*"
}
`, rName)
}

func testAccAWSPinpointAppConfig_Limits(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  limits {
    daily               = 3
    maximum_duration    = 600
    messages_per_second = 50
    total               = 100
  }
}
`, rName)
}

func testAccAWSPinpointAppConfig_QuietTime(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  quiet_time {
    start = "00:00"
    end   = "03:00"
  }
}
`, rName)
}

func testAccAWSPinpointAppConfig_Tag1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSPinpointAppConfig_Tag2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCheckAWSPinpointAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_app" {
			continue
		}

		// Check if the topic exists by fetching its attributes
		params := &pinpoint.GetAppInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetApp(params)
		if err != nil {
			if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("App exists when it should be destroyed!")
	}

	return nil
}
