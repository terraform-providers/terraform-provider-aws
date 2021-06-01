package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_imagebuilder_component", &resource.Sweeper{
		Name: "aws_imagebuilder_component",
		F:    testSweepImageBuilderComponents,
	})
}

func testSweepImageBuilderComponents(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).imagebuilderconn

	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListComponentsInput{
		Owner: aws.String(imagebuilder.OwnershipSelf),
	}

	err = conn.ListComponentsPages(input, func(page *imagebuilder.ListComponentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, componentVersion := range page.ComponentVersionList {
			if componentVersion == nil {
				continue
			}

			arn := aws.StringValue(componentVersion.Arn)
			input := &imagebuilder.ListComponentBuildVersionsInput{
				ComponentVersionArn: componentVersion.Arn,
			}

			err := conn.ListComponentBuildVersionsPages(input, func(page *imagebuilder.ListComponentBuildVersionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, componentSummary := range page.ComponentSummaryList {
					if componentSummary == nil {
						continue
					}

					arn := aws.StringValue(componentSummary.Arn)

					r := resourceAwsImageBuilderComponent()
					d := r.Data(nil)
					d.SetId(arn)

					err := r.Delete(d, client)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting Image Builder Component (%s): %w", arn, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
						continue
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErr := fmt.Errorf("error listing Image Builder Component (%s) versions: %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Component sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Components: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsImageBuilderComponent_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", regexp.MustCompile(fmt.Sprintf("component/%s/1.0.0/[1-9][0-9]*", rName))),
					resource.TestCheckResourceAttr(resourceName, "change_description", ""),
					resource.TestMatchResourceAttr(resourceName, "data", regexp.MustCompile(`schemaVersion`)),
					testAccCheckResourceAttrRfc3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "supported_os_versions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", imagebuilder.ComponentTypeBuild),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0.0"),
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

func TestAccAwsImageBuilderComponent_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsImageBuilderComponent(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsImageBuilderComponent_ChangeDescription(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigChangeDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "change_description", "description1"),
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

func TestAccAwsImageBuilderComponent_Description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func TestAccAwsImageBuilderComponent_KmsKeyId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigKmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccAwsImageBuilderComponent_Platform_Windows(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigPlatformWindows(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformWindows),
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

func TestAccAwsImageBuilderComponent_SupportedOsVersions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigSupportedOsVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "supported_os_versions.#", "1"),
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

func TestAccAwsImageBuilderComponent_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
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
				Config: testAccAwsImageBuilderComponentConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsImageBuilderComponentConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderComponent_Uri(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentConfigUri(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderComponentExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"uri"},
			},
		},
	})
}

func testAccCheckAwsImageBuilderComponentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_component" {
			continue
		}

		input := &imagebuilder.GetComponentInput{
			ComponentBuildVersionArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetComponent(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Component (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Component (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsImageBuilderComponentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

		input := &imagebuilder.GetComponentInput{
			ComponentBuildVersionArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetComponent(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Component (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAwsImageBuilderComponentConfigChangeDescription(rName string, changeDescription string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  change_description = %[2]q
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}
`, rName, changeDescription)
}

func testAccAwsImageBuilderComponentConfigDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  description = %[2]q
  name        = %[1]q
  platform    = "Linux"
  version     = "1.0.0"
}
`, rName, description)
}

func testAccAwsImageBuilderComponentConfigKmsKeyId(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  kms_key_id = aws_kms_key.test.arn
  name       = %[1]q
  platform   = "Linux"
  version    = "1.0.0"
}
`, rName)
}

func testAccAwsImageBuilderComponentConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}
`, rName)
}

func testAccAwsImageBuilderComponentConfigPlatformWindows(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecutePowerShell"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Windows"
  version  = "1.0.0"
}
`, rName)
}

func testAccAwsImageBuilderComponentConfigSupportedOsVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name                  = %[1]q
  platform              = "Linux"
  supported_os_versions = ["Amazon Linux 2"]
  version               = "1.0.0"
}
`, rName)
}

func testAccAwsImageBuilderComponentConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsImageBuilderComponentConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsImageBuilderComponentConfigUri(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  content = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  key = "test.yml"
}

resource "aws_imagebuilder_component" "test" {
  name     = %[1]q
  platform = "Linux"
  uri      = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_bucket_object.test.key}"
  version  = "1.0.0"
}
`, rName)
}
