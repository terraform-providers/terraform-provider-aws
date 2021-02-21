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
	resource.AddTestSweepers("aws_imagebuilder_distribution_configuration", &resource.Sweeper{
		Name: "aws_imagebuilder_distribution_configuration",
		F:    testSweepImageBuilderDistributionConfigurations,
	})
}

func testSweepImageBuilderDistributionConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).imagebuilderconn

	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListDistributionConfigurationsInput{}

	err = conn.ListDistributionConfigurationsPages(input, func(page *imagebuilder.ListDistributionConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, distributionConfigurationSummary := range page.DistributionConfigurationSummaryList {
			if distributionConfigurationSummary == nil {
				continue
			}

			arn := aws.StringValue(distributionConfigurationSummary.Arn)

			r := resourceAwsImageBuilderDistributionConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Image Builder Distribution Configuration (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Distribution Configuration sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Distribution Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsImageBuilderDistributionConfiguration_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("distribution-configuration/%s", rName)),
					testAccCheckResourceAttrRfc3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_updated", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAwsImageBuilderDistributionConfiguration_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsImageBuilderDistributionConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesMultipleRegion(nil, 2),
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistribution2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "2"),
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

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_AmiTags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationAmiTags(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":               "1",
						"ami_distribution_configuration.0.ami_tags.%":    "1",
						"ami_distribution_configuration.0.ami_tags.key1": "value1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationAmiTags(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":               "1",
						"ami_distribution_configuration.0.ami_tags.%":    "1",
						"ami_distribution_configuration.0.ami_tags.key2": "value2",
					}),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_Description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":             "1",
						"ami_distribution_configuration.0.description": "description1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":             "1",
						"ami_distribution_configuration.0.description": "description2",
					}),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_KmsKeyId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationKmsKeyId1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationKmsKeyId2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_LaunchPermission_UserGroups(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationLaunchPermissionUserGroups(rName, "all"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_groups.*", "all"),
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

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_LaunchPermission_UserIds(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationLaunchPermissionUserIds(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_ids.*", "111111111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationLaunchPermissionUserIds(rName, "222222222222"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_ids.*", "222222222222"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationName(rName, "name1-{{ imagebuilder:buildDate }}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":      "1",
						"ami_distribution_configuration.0.name": "name1-{{ imagebuilder:buildDate }}",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationName(rName, "name2-{{ imagebuilder:buildDate }}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":      "1",
						"ami_distribution_configuration.0.name": "name2-{{ imagebuilder:buildDate }}",
					}),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_AmiDistributionConfiguration_TargetAccountIds(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationTargetAccountIds(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.target_account_ids.*", "111111111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationTargetAccountIds(rName, "222222222222"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.target_account_ids.*", "222222222222"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Distribution_LicenseConfigurationArns(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	licenseConfigurationResourceName := "aws_licensemanager_license_configuration.test"
	licenseConfigurationResourceName2 := "aws_licensemanager_license_configuration.test2"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionLicenseConfigurationArns1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.license_configuration_arns.*", licenseConfigurationResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigDistributionLicenseConfigurationArns2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					testAccCheckResourceAttrRfc3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.license_configuration_arns.*", licenseConfigurationResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccAwsImageBuilderDistributionConfiguration_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
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
				Config: testAccAwsImageBuilderDistributionConfigurationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsImageBuilderDistributionConfigurationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsImageBuilderDistributionConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_distribution_configuration" {
			continue
		}

		input := &imagebuilder.GetDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistributionConfiguration(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Distribution Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsImageBuilderDistributionConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

		input := &imagebuilder.GetDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDistributionConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAwsImageBuilderDistributionConfigurationConfigDescription(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  description = %[2]q
  name        = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistribution2(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationAmiTags(rName string, amiTagKey string, amiTagValue string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      ami_tags = {
        %[2]q = %[3]q
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, amiTagKey, amiTagValue)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationDescription(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      description = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationKmsKeyId1(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      kms_key_id = aws_kms_key.test.arn
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationKmsKeyId2(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      kms_key_id = aws_kms_key.test2.arn
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationLaunchPermissionUserGroups(rName string, userGroup string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      launch_permission {
        user_groups = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, userGroup)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationLaunchPermissionUserIds(rName string, userId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      launch_permission {
        user_ids = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, userId)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationName(rName string, name string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, name)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionAmiDistributionConfigurationTargetAccountIds(rName string, targetAccountId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      target_account_ids = [%[2]q]
    }

    region = data.aws_region.current.name
  }
}
`, rName, targetAccountId)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionLicenseConfigurationArns1(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "Socket"
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    license_configuration_arns = [aws_licensemanager_license_configuration.test.id]
    region                     = data.aws_region.current.name
  }
}
`, rName)
}

func testAccAwsImageBuilderDistributionConfigurationConfigDistributionLicenseConfigurationArns2(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_licensemanager_license_configuration" "test2" {
  name                  = %[1]q
  license_counting_type = "Socket"
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    license_configuration_arns = [aws_licensemanager_license_configuration.test2.id]
    region                     = data.aws_region.current.name
  }
}
`, rName)
}

func testAccAwsImageBuilderDistributionConfigurationConfigName(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccAwsImageBuilderDistributionConfigurationConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsImageBuilderDistributionConfigurationConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
