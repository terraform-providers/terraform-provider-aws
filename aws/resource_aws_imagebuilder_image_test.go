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
	resource.AddTestSweepers("aws_imagebuilder_image", &resource.Sweeper{
		Name: "aws_imagebuilder_image",
		F:    testSweepImageBuilderImages,
	})
}

func testSweepImageBuilderImages(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).imagebuilderconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &imagebuilder.ListImagesInput{
		Owner: aws.String(imagebuilder.OwnershipSelf),
	}

	err = conn.ListImagesPages(input, func(page *imagebuilder.ListImagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageVersion := range page.ImageVersionList {
			if imageVersion == nil {
				continue
			}

			// Retrieve the Image's Build Version ARNs required as input
			// to the resourceAwsImageBuilderImage()'s Delete operation
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19851
			imageVersionArn := aws.StringValue(imageVersion.Arn)

			input := &imagebuilder.ListImageBuildVersionsInput{
				ImageVersionArn: imageVersion.Arn,
			}

			err := conn.ListImageBuildVersionsPages(input, func(page *imagebuilder.ListImageBuildVersionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, imageSummary := range page.ImageSummaryList {
					if imageSummary == nil {
						continue
					}

					imageBuildVersionArn := aws.StringValue(imageSummary.Arn)

					r := resourceAwsImageBuilderImage()
					d := r.Data(nil)
					d.SetId(imageBuildVersionArn)

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Image Builder Image Build Versions for image (%s): %w", imageVersionArn, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Image Builder Images for %s: %w", region, err))
	}

	if err := testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Image Builder Images for %s: %w", region, err))
	}

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Image sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAwsImageBuilderImage_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigRequired(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", regexp.MustCompile(fmt.Sprintf("image/%s/1.0.0/[1-9][0-9]*", rName))),
					testAccCheckResourceAttrRfc3339(resourceName, "date_created"),
					resource.TestCheckNoResourceAttr(resourceName, "distribution_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "720"),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "os_version", "Amazon Linux 2"),
					resource.TestCheckResourceAttr(resourceName, "output_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile(`1.0.0/[1-9][0-9]*`)),
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

func TestAccAwsImageBuilderImage_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigRequired(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsImageBuilderImage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsImageBuilderImage_DistributionConfigurationArn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	distributionConfigurationResourceName := "aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigDistributionConfigurationArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, "arn"),
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

func TestAccAwsImageBuilderImage_EnhancedImageMetadataEnabled(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigEnhancedImageMetadataEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "false"),
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

func TestAccAwsImageBuilderImage_ImageTestsConfiguration_ImageTestsEnabled(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigImageTestsConfigurationImageTestsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "false"),
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

func TestAccAwsImageBuilderImage_ImageTestsConfiguration_TimeoutMinutes(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigImageTestsConfigurationTimeoutMinutes(rName, 721),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "721"),
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

func TestAccAwsImageBuilderImage_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsImageBuilderImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImageConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
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
				Config: testAccAwsImageBuilderImageConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsImageBuilderImageConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsImageBuilderImageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_image_pipeline" {
			continue
		}

		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetImage(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Image (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsImageBuilderImageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetImage(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAwsImageBuilderImageConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_imagebuilder_component" "update-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/1.0.0"
}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilder,
  ]
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilder" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilder"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}
`, rName)
}

func testAccAwsImageBuilderImageConfigDistributionConfigurationArn(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}

resource "aws_imagebuilder_image" "test" {
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.test.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName))
}

func testAccAwsImageBuilderImageConfigEnhancedImageMetadataEnabled(rName string, enhancedImageMetadataEnabled bool) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  enhanced_image_metadata_enabled  = %[2]t
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName, enhancedImageMetadataEnabled))
}

func testAccAwsImageBuilderImageConfigImageTestsConfigurationImageTestsEnabled(rName string, imageTestsEnabled bool) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_tests_configuration {
    image_tests_enabled = %[2]t
  }
}
`, rName, imageTestsEnabled))
}

func testAccAwsImageBuilderImageConfigImageTestsConfigurationTimeoutMinutes(rName string, timeoutMinutes int) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_tests_configuration {
    timeout_minutes = %[2]d
  }
}
`, rName, timeoutMinutes))
}

func testAccAwsImageBuilderImageConfigRequired(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`)
}

func testAccAwsImageBuilderImageConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAwsImageBuilderImageConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return composeConfig(
		testAccAwsImageBuilderImageConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
