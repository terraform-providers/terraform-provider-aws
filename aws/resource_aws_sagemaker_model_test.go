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

const (
	image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListModelsInput{}
	resp, err := conn.ListModels(req)
	if err != nil {
		return fmt.Errorf("error listing models: %s", err)
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
				"error deleting sagemaker model (%s): %s",
				*model.ModelName, err)
		}
	}

	return nil
}

func TestAccAWSSagemakerModel_basic(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfig(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "name",
						fmt.Sprintf("terraform-testacc-sagemaker-model-%s", rName)),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "primary_container.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "primary_container.0.image", image),
					resource.TestCheckResourceAttrSet("aws_sagemaker_model.foo", "execution_role_arn"),
					resource.TestCheckResourceAttrSet("aws_sagemaker_model.foo", "arn"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_tags(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelConfigTags(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.foo", "bar"),
				),
			},
			{
				Config: testAccSagemakerModelConfigTagsUpdate(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "tags.bar", "baz"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_primaryContainerModelDataUrl(t *testing.T) {
	rName := acctest.RandString(10)
	modelDataUrl := fmt.Sprintf("https://s3-us-west-2.amazonaws.com/terraform-testacc-sagemaker-model-data-bucket-%s/model.tar.gz", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerModelDataUrlConfig(rName, image, modelDataUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.model_data_url", modelDataUrl),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_primaryContainerHostname(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerHostnameConfig(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.container_hostname", "foo"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_primaryContainerEnvironment(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerPrimaryContainerEnvironmentConfig(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.environment.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo",
						"primary_container.0.environment.foo", "bar"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_containers(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelContainers(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "container.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "container.0.image", image),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "container.1.image", image),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_vpcConfig(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelVpcConfig(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "vpc_config.#", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "vpc_config.0.subnets.#", "2"),
					resource.TestCheckResourceAttr("aws_sagemaker_model.foo", "vpc_config.0.security_group_ids.#", "2"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModel_networkIsolation(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerModelNetworkIsolation(rName, image),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerModelExists("aws_sagemaker_model.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_model.foo", "enable_network_isolation", "true"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_model.foo",
				ImportState:       true,
				ImportStateVerify: true,
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

func testAccSagemakerModelConfig(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image = "%s"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, rName)
}

func testAccSagemakerModelConfigTags(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image = "%s"
  }

  tags = {
    foo = "bar"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, rName)
}

func testAccSagemakerModelConfigTagsUpdate(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image = "%s"
  }

  tags = {
    bar = "baz"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, rName)
}

func testAccSagemakerPrimaryContainerModelDataUrlConfig(rName string, image string, modelDataUrl string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image          = "%s"
    model_data_url = "%s"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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

resource "aws_iam_policy" "foo" {
  name        = "terraform-testacc-sagemaker-model-%s"
  description = "Allow Sagemaker to create model"
  policy      = data.aws_iam_policy_document.foo.json
}

data "aws_iam_policy_document" "foo" {
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
      "arn:aws:s3:::${aws_s3_bucket.foo.bucket}",
      "arn:aws:s3:::${aws_s3_bucket.foo.bucket}/*",
    ]
  }
}

resource "aws_iam_role_policy_attachment" "foo" {
  role       = aws_iam_role.foo.name
  policy_arn = aws_iam_policy.foo.arn
}

resource "aws_s3_bucket" "foo" {
  bucket        = "terraform-testacc-sagemaker-model-data-bucket-%s"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.foo.bucket
  key     = "model.tar.gz"
  content = "some-data"
}
`, rName, image, modelDataUrl, rName, rName, rName)
}

func testAccSagemakerPrimaryContainerHostnameConfig(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image              = "%s"
    container_hostname = "foo"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, rName)
}

func testAccSagemakerPrimaryContainerEnvironmentConfig(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image = "%s"

    environment = {
      foo = "bar"
    }
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, rName)
}

func testAccSagemakerModelContainers(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn = aws_iam_role.foo.arn

  container {
    image = "%s"
  }

  container {
    image = "%s"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, image, rName)
}

func testAccSagemakerModelNetworkIsolation(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name                     = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn       = aws_iam_role.foo.arn
  enable_network_isolation = true

  primary_container {
    image = "%s"
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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
`, rName, image, rName)
}

func testAccSagemakerModelVpcConfig(rName string, image string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model" "foo" {
  name                     = "terraform-testacc-sagemaker-model-%s"
  execution_role_arn       = aws_iam_role.foo.arn
  enable_network_isolation = true

  primary_container {
    image = "%s"
  }

  vpc_config {
    subnets            = [aws_subnet.foo.id, aws_subnet.bar.id]
    security_group_ids = [aws_security_group.foo.id, aws_security_group.bar.id]
  }
}

resource "aws_iam_role" "foo" {
  name               = "terraform-testacc-sagemaker-model-%s"
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

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-sagemaker-model-%s"
  }
}

resource "aws_subnet" "foo" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-sagemaker-model-foo-%s"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-sagemaker-model-bar-%s"
  }
}

resource "aws_security_group" "foo" {
  name   = "terraform-testacc-sagemaker-model-foo-%s"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group" "bar" {
  name   = "terraform-testacc-sagemaker-model-bar-%s"
  vpc_id = aws_vpc.foo.id
}
`, rName, image, rName, rName, rName, rName, rName, rName)
}
