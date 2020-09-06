package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
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
	resourceName := "aws_sagemaker_endpoint_configuration.foo"

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

func TestAccAWSSagemakerEndpointConfiguration_ProductionVariants_InitialVariantWeight(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.foo"

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

func TestAccAWSSagemakerEndpointConfiguration_ProductionVariants_AcceleratorType(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.foo"

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

func TestAccAWSSagemakerEndpointConfiguration_KmsKeyId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfiguration_Config_KmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_arn"),
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

func TestAccAWSSagemakerEndpointConfiguration_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				Config: testAccSagemakerEndpointConfigurationConfig_Tags_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.bar", "baz"),
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
resource "aws_sagemaker_model" "foo" {
  name               = %q
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
  }
}

resource "aws_iam_role" "foo" {
  name               = %q
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
`, rName, rName)
}

func testAccSagemakerEndpointConfigurationConfig_Basic(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
  }

  production_variants {
    variant_name           = "variant-2"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 0.5
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
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
resource "aws_sagemaker_endpoint_configuration" "foo" {
  name        = %q
  kms_key_arn = aws_kms_key.foo.arn

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }
}

resource "aws_kms_key" "foo" {
  description             = %q
  deletion_window_in_days = 10
}
`, rName, rName)
}

func testAccSagemakerEndpointConfigurationConfig_Tags(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    foo = "bar"
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_Tags_Update(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    bar = "baz"
  }
}
`, rName)
}
