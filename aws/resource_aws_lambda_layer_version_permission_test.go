package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lambda_layer_version_permission", &resource.Sweeper{
		Name: "aws_lambda_layer_version_permission",
		F:    testSweepLambdaLayerVersionPermission,
	})
}

func testSweepLambdaLayerVersionPermission(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	lambdaconn := client.(*AWSClient).lambdaconn
	resp, err := lambdaconn.ListLayers(&lambda.ListLayersInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Layer sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Lambda layers: %s", err)
	}

	if len(resp.Layers) == 0 {
		log.Print("[DEBUG] No aws lambda layers to sweep")
		return nil
	}

	for _, l := range resp.Layers {
		versionResp, err := lambdaconn.ListLayerVersions(&lambda.ListLayerVersionsInput{
			LayerName: l.LayerName,
		})
		if err != nil {
			return fmt.Errorf("Error retrieving versions for lambda layer: %s", err)
		}

		for _, v := range versionResp.LayerVersions {
			_, err := lambdaconn.DeleteLayerVersion(&lambda.DeleteLayerVersionInput{
				LayerName:     l.LayerName,
				VersionNumber: v.Version,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccAWSLambdaLayerVersionPermission_all(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.lambda_layer_permission"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionPermission_all(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists("aws_lambda_layer_version.lambda_layer", rName),
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_arn", "aws_lambda_layer_version.lambda_layer", "layer_arn"),
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

func TestAccAWSLambdaLayerVersionPermission_org(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.lambda_layer_permission"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionPermission_org(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists("aws_lambda_layer_version.lambda_layer", rName),
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttr(resourceName, "organization_id", "o-0123456789"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_arn", "aws_lambda_layer_version.lambda_layer", "layer_arn"),
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

func TestAccAWSLambdaLayerVersionPermission_account(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.lambda_layer_permission"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionPermission_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists("aws_lambda_layer_version.lambda_layer", rName),
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "456789820214"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_arn", "aws_lambda_layer_version.lambda_layer", "layer_arn"),
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

func TestAccAWSLambdaLayerVersionPermission_disappears(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.lambda_layer_permission"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionPermission_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLambdaLayerVersionPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Creating Lambda layer and Lambda layer permissions

func testAccAWSLambdaLayerVersionPermission_all(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "lambda_layer_permission" {
  layer_arn     = aws_lambda_layer_version.lambda_layer.layer_arn
  layer_version = aws_lambda_layer_version.lambda_layer.version
  action        = "lambda:GetLayerVersion"
  statement_id  = "xaccount"
  principal     = "*"
}
`, layerName)
}

func testAccAWSLambdaLayerVersionPermission_org(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "lambda_layer_permission" {
  layer_arn       = aws_lambda_layer_version.lambda_layer.layer_arn
  layer_version   = aws_lambda_layer_version.lambda_layer.version
  action          = "lambda:GetLayerVersion"
  statement_id    = "xaccount"
  principal       = "*"
  organization_id = "o-0123456789"
}
`, layerName)
}

func testAccAWSLambdaLayerVersionPermission_account(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "lambda_layer_permission" {
  layer_arn     = aws_lambda_layer_version.lambda_layer.layer_arn
  layer_version = aws_lambda_layer_version.lambda_layer.version
  action        = "lambda:GetLayerVersion"
  statement_id  = "xaccount"
  principal     = "456789820214"
}
`, layerName)
}

func testAccCheckAwsLambdaLayerVersionPermissionExists(res, layerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda Layer version permission not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda Layer version policy ID not set")
		}

		if rs.Primary.Attributes["revision_id"] == "" {
			return fmt.Errorf("Lambda Layer Version Permission not set")
		}

		_, _, version, err := resourceAwsLambdaLayerVersionPermissionParseId(rs.Primary.Attributes["id"])
		if err != nil {
			return fmt.Errorf("Error parsing lambda layer ID: %s", err)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		_, err = conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})

		if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
			return err
		}

		return err
	}
}

func testAccCheckLambdaLayerVersionPermissionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_layer_version_permission" {
			continue
		}

		layerName, _, version, err := resourceAwsLambdaLayerVersionPermissionParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})

		if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		// as I've created Lambda layer, not only layer permission, need to check if layer was destroyed.
		err = testAccCheckLambdaLayerVersionDestroy(s)
		if err != nil {
			return err
		}

		return fmt.Errorf("Lambda Layer Version Permission (%s) still exists", rs.Primary.ID)
	}
	return nil
}
