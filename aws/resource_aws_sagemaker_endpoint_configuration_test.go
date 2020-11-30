package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_endpoint_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint_configuration",
		Dependencies: []string{
			"aws_sagemaker_model",
		},
		F: testSweepSagemakerEndpointConfigurations,
	})
}

func testSweepSagemakerEndpointConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListEndpointConfigsInput{
		NameContains: aws.String("tf-acc-test"),
	}
	resp, err := conn.ListEndpointConfigs(req)
	if err != nil {
		return fmt.Errorf("error listing endpoint configs: %s", err)
	}

	if len(resp.EndpointConfigs) == 0 {
		log.Print("[DEBUG] No SageMaker endpoint config to sweep")
		return nil
	}

	for _, endpointConfig := range resp.EndpointConfigs {
		_, err := conn.DeleteEndpointConfig(&sagemaker.DeleteEndpointConfigInput{
			EndpointConfigName: endpointConfig.EndpointConfigName,
		})
		if err != nil {
			return fmt.Errorf(
				"failed to delete SageMaker endpoint config (%s): %s",
				*endpointConfig.EndpointConfigName, err)
		}
	}

	return nil
}

func TestAccAWSSagemakerEndpointConfiguration_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "production_variants.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.variant_name", "variant-1"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.model_name", rName),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.initial_instance_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.instance_type", "ml.t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.initial_variant_weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.#", "0"),
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

func TestAccAWSSagemakerEndpointConfiguration_productionVariants_InitialVariantWeight(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "production_variants.1.initial_variant_weight", "0.5"),
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

func TestAccAWSSagemakerEndpointConfiguration_productionVariants_AcceleratorType(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.accelerator_type", "ml.eia1.medium"),
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

func TestAccAWSSagemakerEndpointConfiguration_kmsKeyId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfiguration_Config_KmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccAWSSagemakerEndpointConfiguration_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
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
				Config: testAccSagemakerEndpointConfigurationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSagemakerEndpointConfigurationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfiguration_dataCaptureConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationDataCaptureConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.enable_capture", "true"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.initial_sampling_percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.destination_s3_uri", fmt.Sprintf("s3://%s/", rName)),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.capture_options.0.capture_mode", "Input"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.capture_options.1.capture_mode", "Output"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.capture_content_type_header.0.json_content_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_capture_config.0.capture_content_type_header.0.json_content_types.*", "application/json"),
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

func TestAccAWSSagemakerEndpointConfiguration_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerEndpointConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSagemakerEndpointConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint_configuration" {
			continue
		}

		input := &sagemaker.DescribeEndpointConfigInput{
			EndpointConfigName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeEndpointConfig(input)

		if isAWSErr(err, "ValidationException", "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Endpoint Configuration (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSagemakerEndpointConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SageMaker endpoint config not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker endpoint config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeEndpointConfigInput{
			EndpointConfigName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeEndpointConfig(opts)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSagemakerEndpointConfigurationConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}

resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_Basic(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
  }

  production_variants {
    variant_name           = "variant-2"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 0.5
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    accelerator_type       = "ml.eia1.medium"
    initial_variant_weight = 1
  }
}
`, rName)
}

func testAccSagemakerEndpointConfiguration_Config_KmsKeyId(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name        = %q
  kms_key_arn = aws_kms_key.test.arn

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }
}

resource "aws_kms_key" "test" {
  description             = %q
  deletion_window_in_days = 10
}
`, rName, rName)
}

func testAccSagemakerEndpointConfigurationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSagemakerEndpointConfigurationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSagemakerEndpointConfigurationDataCaptureConfig(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  data_capture_config {
    enable_capture              = true
    initial_sampling_percentage = 50
    destination_s3_uri          = "s3://${aws_s3_bucket.test.bucket}/"

    capture_options {
      capture_mode = "Input"
    }

    capture_options {
      capture_mode = "Output"
    }

    capture_content_type_header {
      json_content_types = ["application/json"]
    }
  }
}
`, rName)
}
