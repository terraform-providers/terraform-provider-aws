package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_model", &resource.Sweeper{
		Name: "aws_sagemaker_model",
		F:    testSweepSagemakerModels,
	})
}

func testSweepSagemakerModels(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListModelsInput{}
	resp, err := conn.ListModels(req)
	if err != nil {
		return fmt.Errorf("error listing models: %w", err)
	}

	if len(resp.Models) == 0 {
		log.Print("[DEBUG] No sagemaker models to sweep")
		return nil
	}

	for _, model := range resp.Models {
		_, err := conn.DeleteModel(&sagemaker.DeleteModelInput{
			ModelName: model.ModelName,
		})
		if err != nil {
			return fmt.Errorf(
				"error deleting sagemaker model (%s): %w", aws.StringValue(model.ModelName), err)
		}
	}

	return nil
}

func TestAccAWSSagemakerModel_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_container.0.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.mode", "SingleModel"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("model/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enable_network_isolation", "false"),
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

func TestAccAWSSagemakerModel_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccSagemakerModelConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSagemakerModelConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccAWSSagemakerModel_primaryContainerModelDataUrl(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerModelDataUrlConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "primary_container.0.model_data_url"),
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

func TestAccAWSSagemakerModel_primaryContainerHostname(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerHostnameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.container_hostname", "test"),
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

func TestAccAWSSagemakerModel_primaryContainerImageConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerImageConfigConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.0.repository_access_mode", "Platform"),
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

func TestAccAWSSagemakerModel_primaryContainerEnvironment(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerEnvironmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.test", "bar"),
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

func TestAccAWSSagemakerModel_primaryContainerModeSingle(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerModeSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.mode", "SingleModel"),
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

func TestAccAWSSagemakerModel_containers(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelContainers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "container.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "container.0.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttrPair(resourceName, "container.1.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
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

func TestAccAWSSagemakerModel_vpcConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelVpcConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
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

func TestAccAWSSagemakerModel_networkIsolation(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelNetworkIsolation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_network_isolation", "true"),
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

func TestAccAWSSagemakerModel_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSagemakerModelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_model" {
			continue
		}

		resp, err := conn.ListModels(&sagemaker.ListModelsInput{
			NameContains: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.Models) > 0 {
				return fmt.Errorf("Sagemaker models still exist")
			}

			return nil
		}

		sagemakerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if sagemakerErr.Code() != "ResourceNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSagemakerModelExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker model ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		DescribeModelOpts := &sagemaker.DescribeModelInput{
			ModelName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeModel(DescribeModelOpts)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSagemakerModelConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`, rName)
}

func testAccSagemakerModelConfig(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName)
}

func testAccSagemakerModelConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSagemakerModelConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSagemakerPrimaryContainerModelDataUrlConfig(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "https://s3.amazonaws.com/${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "Allow Sagemaker to create model"
  policy      = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "model.tar.gz"
  content = "some-data"
}
`, rName)
}

func testAccSagemakerPrimaryContainerHostnameConfig(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image              = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    container_hostname = "test"
  }
}
`, rName)
}

func testAccSagemakerPrimaryContainerImageConfigConfig(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path

    image_config {
      repository_access_mode = "Platform"
    }
  }
}
`, rName)
}

func testAccSagemakerPrimaryContainerEnvironmentConfig(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path

    environment = {
      test = "bar"
    }
  }
}
`, rName)
}

func testAccSagemakerPrimaryContainerModeSingle(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    mode  = "SingleModel"
  }
}
`, rName)
}

func testAccSagemakerModelContainers(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName)
}

func testAccSagemakerModelNetworkIsolation(rName string) string {
	return testAccSagemakerModelConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  execution_role_arn       = aws_iam_role.test.arn
  enable_network_isolation = true

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName)
}

func testAccSagemakerModelVpcConfig(rName string) string {
	return testAccSagemakerModelConfigBase(rName) +
		testAccAvailableAZsNoOptInConfig() +
		fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  execution_role_arn       = aws_iam_role.test.arn
  enable_network_isolation = true

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  vpc_config {
    subnets            = [aws_subnet.test.id, aws_subnet.bar.id]
    security_group_ids = [aws_security_group.test.id, aws_security_group.bar.id]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "bar" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
