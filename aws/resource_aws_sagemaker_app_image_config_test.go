package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_app_image_config", &resource.Sweeper{
		Name: "aws_sagemaker_app_image_config",
		F:    testSweepSagemakerAppImageConfigs,
	})
}

func testSweepSagemakerAppImageConfigs(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).sagemakerconn
	input := &sagemaker.ListAppImageConfigsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListAppImageConfigs(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker App Image Config for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Example Thing: %w", err))
			return sweeperErrs
		}

		for _, config := range output.AppImageConfigs {
			name := aws.StringValue(config.AppImageConfigName)
			input := &sagemaker.DeleteAppImageConfigInput{
				AppImageConfigName: aws.String(name),
			}

			log.Printf("[INFO] Deleting SageMaker App Image Config: %s", name)
			_, err := conn.DeleteAppImageConfig(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SageMaker App Image Config (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSagemakerAppImageConfig_basic(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("app-image-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSSagemakerAppImageConfig_kernelGatewayImageConfig_kernalSpecs(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.display_name", rName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerAppImageConfig_kernelGatewayImageConfig_fileSystemConfig(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", "100"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", "1000"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/home/sagemaker-user"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/test"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerAppImageConfig_disappears(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppImageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppImageConfigBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppImageConfigExists(resourceName, &config),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerAppImageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerAppImageConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_app_image_config" {
			continue
		}

		config, err := finder.AppImageConfigByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(config.AppImageConfigName) == rs.Primary.ID {
			return fmt.Errorf("Sagemaker App Image Config %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerAppImageConfigExists(n string, config *sagemaker.DescribeAppImageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker App Image Config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.AppImageConfigByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*config = *resp

		return nil
	}
}

func testAccAWSSagemakerAppImageConfigBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs1(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigKernalSpecs2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name         = "%[1]s-2"
      display_name = %[1]q
    }
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig1(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }

    file_system_config {}
  }
}
`, rName)
}

func testAccAWSSagemakerAppImageConfigKernelGatewayImageConfigFileSystemConfig2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }

    file_system_config {
      default_gid = 0
      default_uid = 0
      mount_path  = "/test"
    }
  }
}
`, rName)
}
