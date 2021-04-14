package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

// Tests are serialized as SagmMaker Domain resources are limited to 1 per account by default.
// SageMaker UserProfile and App depend on the Domain resources and as such are also part of the serialized test suite.
func TestAccAWSSagemakerDomain_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Domain": {
			"basic":                                testAccAWSSagemakerDomain_basic,
			"disappears":                           testAccAWSSagemakerDomain_tags,
			"tags":                                 testAccAWSSagemakerDomain_disappears,
			"tensorboardAppSettings":               testAccAWSSagemakerDomain_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage":      testAccAWSSagemakerDomain_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":             testAccAWSSagemakerDomain_kernelGatewayAppSettings,
			"kernelGatewayAppSettings_customImage": testAccAWSSagemakerDomain_kernelGatewayAppSettings_customImage,
			"jupyterServerAppSettings":             testAccAWSSagemakerDomain_jupyterServerAppSettings,
			"kms":                                  testAccAWSSagemakerDomain_kms,
			"securityGroup":                        testAccAWSSagemakerDomain_securityGroup,
			"sharingSettings":                      testAccAWSSagemakerDomain_sharingSettings,
		},
		"UserProfile": {
			"basic":                           testAccAWSSagemakerUserProfile_basic,
			"disappears":                      testAccAWSSagemakerUserProfile_tags,
			"tags":                            testAccAWSSagemakerUserProfile_disappears,
			"tensorboardAppSettings":          testAccAWSSagemakerUserProfile_tensorboardAppSettings,
			"tensorboardAppSettingsWithImage": testAccAWSSagemakerUserProfile_tensorboardAppSettingsWithImage,
			"kernelGatewayAppSettings":        testAccAWSSagemakerUserProfile_kernelGatewayAppSettings,
			"jupyterServerAppSettings":        testAccAWSSagemakerUserProfile_jupyterServerAppSettings,
		},
		"App": {
			"basic":        testAccAWSSagemakerApp_basic,
			"disappears":   testAccAWSSagemakerApp_tags,
			"tags":         testAccAWSSagemakerApp_disappears,
			"resourceSpec": testAccAWSSagemakerApp_resourceSpec,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func init() {
	resource.AddTestSweepers("aws_sagemaker_domain", &resource.Sweeper{
		Name: "aws_sagemaker_domain",
		F:    testSweepSagemakerDomains,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_file_system",
			"aws_sagemaker_user_profile",
		},
	})
}

func testSweepSagemakerDomains(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn
	var sweeperErrs *multierror.Error

	err = conn.ListDomainsPages(&sagemaker.ListDomainsInput{}, func(page *sagemaker.ListDomainsOutput, lastPage bool) bool {
		for _, domain := range page.Domains {

			r := resourceAwsSagemakerDomain()
			d := r.Data(nil)
			d.SetId(aws.StringValue(domain.DomainId))
			err = r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Sagemaker Domains: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAWSSagemakerDomain_basic(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", "arn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_id"),
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

func testAccAWSSagemakerDomain_kms(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainKMSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "arn"),
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

func testAccAWSSagemakerDomain_tags(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
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
				Config: testAccAWSSagemakerDomainConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerDomainConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSSagemakerDomain_securityGroup(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigSecurityGroup1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerDomainConfigSecurityGroup2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "2"),
				),
			},
		},
	})
}

func testAccAWSSagemakerDomain_sharingSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigSharingSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.sharing_settings.0.notebook_output_option", "Allowed"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.sharing_settings.0.s3_kms_key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_user_settings.0.sharing_settings.0.s3_output_path"),
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

func testAccAWSSagemakerDomain_tensorboardAppSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigTensorBoardAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccAWSSagemakerDomain_tensorboardAppSettingsWithImage(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigTensorBoardAppSettingsWithImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.tensor_board_app_settings.0.default_resource_spec.0.sagemaker_image_arn", "aws_sagemaker_image.test", "arn"),
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

func testAccAWSSagemakerDomain_kernelGatewayAppSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigKernelGatewayAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccAWSSagemakerDomain_kernelGatewayAppSettings_customImage(t *testing.T) {

	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigKernelGatewayAppSettingsCustomImage(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.default_resource_spec.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.0.app_image_config_name", "aws_sagemaker_app_image_config.test", "app_image_config_name"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.kernel_gateway_app_settings.0.custom_image.0.image_name", "aws_sagemaker_image.test", "image_name"),
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

func testAccAWSSagemakerDomain_jupyterServerAppSettings(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigJupyterServerAppSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.default_resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.jupyter_server_app_settings.0.default_resource_spec.0.instance_type", "ml.t3.micro"),
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

func testAccAWSSagemakerDomain_disappears(t *testing.T) {
	var domain sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &domain),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_domain" {
			continue
		}

		domain, err := finder.DomainByName(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker Domain (%s): %w", rs.Primary.ID, err)
		}

		domainArn := aws.StringValue(domain.DomainArn)
		domainID, err := decodeSagemakerDomainID(domainArn)
		if err != nil {
			return err
		}

		if domainID == rs.Primary.ID {
			return fmt.Errorf("sagemaker domain %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerDomainExists(n string, codeRepo *sagemaker.DescribeDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.DomainByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccAWSSagemakerDomainConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}
`, rName)
}

func testAccAWSSagemakerDomainBasicConfig(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
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

func testAccAWSSagemakerDomainKMSConfig(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test"
  deletion_window_in_days = 7
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]
  kms_key_id  = aws_kms_key.test.arn

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigSecurityGroup1(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%[1]s"
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = [aws_security_group.test.id]
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigSecurityGroup2(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = [aws_security_group.test.id, aws_security_group.test2.id]
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerDomainConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSagemakerDomainConfigSharingSettings(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}


resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    sharing_settings {
      notebook_output_option = "Allowed"
      s3_kms_key_id          = aws_kms_key.test.arn
      s3_output_path         = "s3://${aws_s3_bucket.test.bucket}/sharing"
    }
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigTensorBoardAppSettings(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
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

func testAccAWSSagemakerDomainConfigTensorBoardAppSettingsWithImage(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
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

func testAccAWSSagemakerDomainConfigJupyterServerAppSettings(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
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

func testAccAWSSagemakerDomainConfigKernelGatewayAppSettings(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
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

func testAccAWSSagemakerDomainConfigKernelGatewayAppSettingsCustomImage(rName, baseImage string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }
  }
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn

    kernel_gateway_app_settings {
      custom_image {
        app_image_config_name = aws_sagemaker_app_image_config.test.app_image_config_name
        image_name            = aws_sagemaker_image_version.test.image_name
      }
    }
  }
}
`, rName, baseImage)
}
