package aws

import (
	"fmt"
	"log"
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
	resource.AddTestSweepers("aws_imagebuilder_infrastructure_configuration", &resource.Sweeper{
		Name: "aws_imagebuilder_infrastructure_configuration",
		F:    testSweepImageBuilderInfrastructureConfigurations,
	})
}

func testSweepImageBuilderInfrastructureConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).imagebuilderconn

	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListInfrastructureConfigurationsInput{}

	err = conn.ListInfrastructureConfigurationsPages(input, func(page *imagebuilder.ListInfrastructureConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, infrastructureConfigurationSummary := range page.InfrastructureConfigurationSummaryList {
			if infrastructureConfigurationSummary == nil {
				continue
			}

			arn := aws.StringValue(infrastructureConfigurationSummary.Arn)

			r := resourceAwsImageBuilderInfrastructureConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Image Builder Infrastructure Configuration (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Infrastructure Configuration sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Infrastructure Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsImageBuilderInfrastructureConfiguration_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamInstanceProfileResourceName := "aws_iam_instance_profile.test"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("infrastructure-configuration/%s", rName)),
					testAccCheckResourceAttrRfc3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_updated", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "false"),
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

func TestAccAwsImageBuilderInfrastructureConfiguration_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsImageBuilderInfrastructureConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_Description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_InstanceProfileName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamInstanceProfileResourceName := "aws_iam_instance_profile.test"
	iamInstanceProfileResourceName2 := "aws_iam_instance_profile.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceProfileName1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceProfileName2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName2, "name"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_InstanceTypes(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceTypes1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceTypes2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_KeyPair(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	keyPairResourceName := "aws_key_pair.test"
	keyPairResourceName2 := "aws_key_pair.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigKeyPair1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_pair", keyPairResourceName, "key_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigKeyPair2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "key_pair", keyPairResourceName2, "key_name"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_Logging_S3Logs_S3BucketName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketResourceName2 := "aws_s3_bucket.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3BucketName1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.s3_logs.0.s3_bucket_name", s3BucketResourceName, "bucket"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3BucketName2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.s3_logs.0.s3_bucket_name", s3BucketResourceName2, "bucket"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_Logging_S3Logs_S3KeyPrefix(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3KeyPrefix(rName, "/prefix1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "/prefix1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3KeyPrefix(rName, "/prefix2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "/prefix2/"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_ResourceTags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigResourceTags(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigResourceTags(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_SecurityGroupIds(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	securityGroupResourceName := "aws_security_group.test"
	securityGroupResourceName2 := "aws_security_group.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigSecurityGroupIds1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigSecurityGroupIds2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_SnsTopicArn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	snsTopicResourceName := "aws_sns_topic.test"
	snsTopicResourceName2 := "aws_sns_topic.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigSnsTopicArn1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", snsTopicResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigSnsTopicArn2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", snsTopicResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_SubnetId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	subnetResourceName := "aws_subnet.test"
	subnetResourceName2 := "aws_subnet.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigSubnetId1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigSubnetId2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
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
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderInfrastructureConfiguration_TerminateInstanceOnFailure(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigTerminateInstanceOnFailure(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderInfrastructureConfigurationConfigTerminateInstanceOnFailure(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsImageBuilderInfrastructureConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_infrastructure_configuration" {
			continue
		}

		input := &imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetInfrastructureConfiguration(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Infrastructure Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsImageBuilderInfrastructureConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

		input := &imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetInfrastructureConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.role.name
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
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
`, rName)
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigDescription(rName string, description string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  description           = %[2]q
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName, description))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceProfileName1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceProfileName2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_instance_profile" "test2" {
  name = aws_iam_role.role2.name
  role = aws_iam_role.role2.name
}

resource "aws_iam_role" "role2" {
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
  name = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test2.name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceTypes1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		testAccAvailableEc2InstanceTypeForRegion("t3.medium", "t2.medium"),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_types        = [data.aws_ec2_instance_type_offering.available.instance_type]
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigInstanceTypes2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		testAccAvailableEc2InstanceTypeForRegion("t3.large", "t2.large"),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_types        = [data.aws_ec2_instance_type_offering.available.instance_type]
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigKeyPair1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 example@example.com"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  key_pair              = aws_key_pair.test.key_name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigKeyPair2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_key_pair" "test2" {
  key_name   = "%[1]s-2"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 example@example.com"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  key_pair              = aws_key_pair.test2.key_name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3BucketName1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test.bucket
    }
  }
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3BucketName2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test2.bucket
    }
  }
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigLoggingS3LogsS3KeyPrefix(rName string, s3KeyPrefix string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test.bucket
      s3_key_prefix  = %[2]q
    }
  }
}
`, rName, s3KeyPrefix))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigName(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigResourceTags(rName string, resourceTagKey string, resourceTagValue string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  resource_tags = {
    %[2]q = %[3]q
  }
}
`, rName, resourceTagKey, resourceTagValue))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigSecurityGroupIds1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id]
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigSecurityGroupIds2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test2.id]
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigSubnetId1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id     = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id] # Required with subnet_id
  subnet_id             = aws_subnet.test.id
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigSubnetId2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 2, 2)
  vpc_id     = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id] # Required with subnet_id
  subnet_id             = aws_subnet.test2.id
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigSnsTopicArn1(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  sns_topic_arn         = aws_sns_topic.test.arn
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigSnsTopicArn2(rName string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test2" {
  name = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  sns_topic_arn         = aws_sns_topic.test2.arn
}
`, rName))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAwsImageBuilderInfrastructureConfigurationConfigTerminateInstanceOnFailure(rName string, terminateInstanceOnFailure bool) string {
	return composeConfig(
		testAccAwsImageBuilderInfrastructureConfigurationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name         = aws_iam_instance_profile.test.name
  name                          = %[1]q
  terminate_instance_on_failure = %[2]t
}
`, rName, terminateInstanceOnFailure))
}
