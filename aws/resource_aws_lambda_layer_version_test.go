package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lambda_layer", &resource.Sweeper{
		Name: "aws_lambda_layer",
		F:    testSweepLambdaLayerVersions,
	})
}

func testSweepLambdaLayerVersions(region string) error {
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
		if !strings.HasPrefix(*l.LayerName, "tf_acc_") {
			continue
		}

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

func TestAccAWSLambdaLayerVersion_basic(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	layerName := fmt.Sprintf("tf_acc_lambda_layer_basic_%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionBasic(layerName),
				Check:  testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccAWSLambdaLayerVersion_update(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	layerName := fmt.Sprintf("tf_acc_lambda_layer_basic_%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionCreateBeforeDestroy(layerName, "test-fixtures/lambdatest.zip"),
				Check:  testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "source_code_hash"},
			},

			{
				Config: testAccAWSLambdaLayerVersionCreateBeforeDestroy(layerName, "test-fixtures/lambdatest_modified.zip"),
				Check:  testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
			},
		},
	})
}

func TestAccAWSLambdaLayerVersion_s3(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := acctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_s3_%s", rString)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-layer-s3-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionS3(bucketName, layerName),
				Check:  testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key"},
			},
		},
	})
}

func TestAccAWSLambdaLayerVersion_compatibleRuntimes(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := acctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_runtimes_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionCompatibleRuntimes(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "2"),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccAWSLambdaLayerVersion_description(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := acctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_description_%s", rString)
	testDescription := "test description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionDescription(layerName, testDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "description", testDescription),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccAWSLambdaLayerVersion_licenseInfo(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := acctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_license_info_%s", rString)
	testLicenseInfo := "MIT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionLicenseInfo(layerName, testLicenseInfo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "license_info", testLicenseInfo),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func testAccCheckLambdaLayerVersionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_layer_version" {
			continue
		}

		layerName, version, err := resourceAwsLambdaLayerVersionParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetLayerVersion(&lambda.GetLayerVersionInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})
		if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Lambda Layer Version (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsLambdaLayerVersionExists(res, layerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda Layer version not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda Layer ID not set")
		}

		if rs.Primary.Attributes["version"] == "" {
			return fmt.Errorf("Lambda Layer Version not set")
		}

		version, err := strconv.Atoi(rs.Primary.Attributes["version"])
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn
		_, err = conn.GetLayerVersion(&lambda.GetLayerVersionInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(int64(version)),
		})
		return err
	}
}

func testAccAWSLambdaLayerVersionBasic(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
	filename = "test-fixtures/lambdatest.zip"
	layer_name = "%s"
}
`, layerName)
}

func testAccAWSLambdaLayerVersionS3(bucketName, layerName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
	bucket = "%s"
}

resource "aws_s3_bucket_object" "lambda_code" {
	bucket = "${aws_s3_bucket.lambda_bucket.id}"
	key = "lambdatest.zip"
	source = "test-fixtures/lambdatest.zip"
}

resource "aws_lambda_layer_version" "lambda_layer_test" {
	s3_bucket = "${aws_s3_bucket.lambda_bucket.id}"
	s3_key = "${aws_s3_bucket_object.lambda_code.id}"
	layer_name = "%s"
}
`, bucketName, layerName)
}

func testAccAWSLambdaLayerVersionCreateBeforeDestroy(layerName string, filename string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename         = "%s"
  layer_name       = "%s"
  source_code_hash = "${base64sha256(file("%s"))}"

  lifecycle {
    create_before_destroy = true
  }
}
`, filename, layerName, filename)
}

func testAccAWSLambdaLayerVersionCompatibleRuntimes(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
	filename = "test-fixtures/lambdatest.zip"
	layer_name = "%s"

	compatible_runtimes = ["nodejs8.10", "nodejs6.10"]
}
`, layerName)
}

func testAccAWSLambdaLayerVersionDescription(layerName string, description string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
	filename = "test-fixtures/lambdatest.zip"
	layer_name = "%s"

	description = "%s"
}
`, layerName, description)
}

func testAccAWSLambdaLayerVersionLicenseInfo(layerName string, licenseInfo string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
	filename = "test-fixtures/lambdatest.zip"
	layer_name = "%s"

	license_info = "%s"
}
`, layerName, licenseInfo)
}
