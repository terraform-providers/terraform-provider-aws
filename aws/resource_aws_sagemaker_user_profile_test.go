package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_user_profile", &resource.Sweeper{
		Name: "aws_sagemaker_user_profile",
		F:    testSweepSagemakerUserProfiles,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_file_system",
		},
	})
}

func testSweepSagemakerUserProfiles(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListUserProfilesPages(&sagemaker.ListUserProfilesInput{}, func(page *sagemaker.ListUserProfilesOutput, lastPage bool) bool {
		for _, instance := range page.UserProfiles {
			input := &sagemaker.DeleteUserProfileInput{
				UserProfileName: instance.UserProfileName,
				DomainId:        instance.DomainId,
			}

			userProfile := aws.StringValue(instance.UserProfileName)
			log.Printf("[INFO] Deleting SageMaker User Profile: %s", userProfile)
			if _, err := conn.DeleteUserProfile(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker User Profile (%s): %s", userProfile, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker domains: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerUserProfile_basic(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_profile_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", "aws_sagemaker_domain.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`user-profile/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_uid"),
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

func TestAccAWSSagemakerUserProfile_tags(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
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
				Config: testAccAWSSagemakerUserProfileConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerUserProfileConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerUserProfile_tensorboardAppSettings(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileConfigTensorBoardAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func TestAccAWSSagemakerUserProfile_tensorboardAppSettingsWithImage(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileConfigTensorBoardAppSettingsWithImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.sagemaker_image_arn", "aws_sagemaker_image.test", "arn"),
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

func TestAccAWSSagemakerUserProfile_kernelGatewayAppSettings(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileConfigKernelGatewayAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func TestAccAWSSagemakerUserProfile_jupyterServerAppSettings(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileConfigJupyterServerAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.jupyter_server_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.jupyter_server_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_settings.0.jupyter_server_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func TestAccAWSSagemakerUserProfile_disappears(t *testing.T) {
	var domain sagemaker.DescribeUserProfileOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerUserProfileBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerUserProfileExists(resourceName, &domain),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerUserProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerUserProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_user_settings" {
			continue
		}

		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]

		userProfile, err := finder.UserProfileByName(conn, domainID, userProfileName)
		if err != nil {
			return nil
		}

		userProfileArn := aws.StringValue(userProfile.UserProfileArn)
		if userProfileArn == rs.Primary.ID {
			return fmt.Errorf("SageMaker User Profile %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerUserProfileExists(n string, userProfile *sagemaker.DescribeUserProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]

		resp, err := finder.UserProfileByName(conn, domainID, userProfileName)
		if err != nil {
			return err
		}

		*userProfile = *resp

		return nil
	}
}

func testAccAWSSagemakerUserProfileConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

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

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSSagemakerUserProfileBasicConfig(rName string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerUserProfileConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerUserProfileConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSagemakerUserProfileConfigTensorBoardAppSettings(rName string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    tensor_board_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName)
}

func testAccAWSSagemakerUserProfileConfigTensorBoardAppSettingsWithImage(rName string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    tensor_board_app_settings {
      default_resource_spec {
        instance_type       = "ml.t3.micro"
        sagemaker_image_arn = aws_sagemaker_image.test.arn
      }
    }
  }
}
`, rName)
}

func testAccAWSSagemakerUserProfileConfigJupyterServerAppSettings(rName string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    jupyter_server_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName)
}

func testAccAWSSagemakerUserProfileConfigKernelGatewayAppSettings(rName string) string {
	return testAccAWSSagemakerUserProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q

  user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      default_resource_spec {
        instance_type = "ml.t3.micro"
      }
    }
  }
}
`, rName)
}
