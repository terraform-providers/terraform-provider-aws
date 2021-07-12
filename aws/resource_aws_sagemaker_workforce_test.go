package aws

import (
	"fmt"
	"log"
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

func init() {
	resource.AddTestSweepers("aws_sagemaker_workforce", &resource.Sweeper{
		Name: "aws_sagemaker_workforce",
		F:    testSweepSagemakerWorkforces,
	})
}

func testSweepSagemakerWorkforces(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sagemakerconn
	var sweeperErrs *multierror.Error

	err = conn.ListWorkforcesPages(&sagemaker.ListWorkforcesInput{}, func(page *sagemaker.ListWorkforcesOutput, lastPage bool) bool {
		for _, workforce := range page.Workforces {

			r := resourceAwsSagemakerWorkforce()
			d := r.Data(nil)
			d.SetId(aws.StringValue(workforce.WorkforceName))
			err := r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker workforce sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Sagemaker Workforces: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSagemakerWorkforce_cognitoConfig(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_workforce.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerWorkforceCognitoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "workforce_name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workforce/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cognito_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cognito_config.0.client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "cognito_config.0.user_pool", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
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

func TestAccAWSSagemakerWorkforce_oidcConfig(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_workforce.test"
	endpoint1 := "https://example.com"
	endpoint2 := "https://test.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerWorkforceOidcConfig(rName, endpoint1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "workforce_name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workforce/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cognito_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.authorization_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_id", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_secret", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.issuer", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.jwks_uri", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.logout_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.token_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.user_info_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oidc_config.0.client_secret"},
			},
			{
				Config: testAccAWSSagemakerWorkforceOidcConfig(rName, endpoint2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "workforce_name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workforce/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cognito_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.authorization_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_id", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_secret", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.issuer", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.jwks_uri", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.logout_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.token_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.user_info_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
		},
	})
}
func TestAccAWSSagemakerWorkforce_sourceIpConfig(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_workforce.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerWorkforceSourceIpConfig1(rName, "1.1.1.1/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "1.1.1.1/32"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerWorkforceSourceIpConfig2(rName, "2.2.2.2/32", "3.3.3.3/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "2.2.2.2/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "3.3.3.3/32"),
				),
			},
			{
				Config: testAccAWSSagemakerWorkforceSourceIpConfig1(rName, "2.2.2.2/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "2.2.2.2/32"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerWorkforce_disappears(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_workforce.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerWorkforceCognitoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerWorkforceExists(resourceName, &workforce),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerWorkforce(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerWorkforceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_workforce" {
			continue
		}

		workforce, err := finder.WorkforceByName(conn, rs.Primary.ID)
		if tfawserr.ErrMessageContains(err, "ValidationException", "No workforce") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker Workforce (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(workforce.WorkforceName) == rs.Primary.ID {
			return fmt.Errorf("SageMaker Workforce %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerWorkforceExists(n string, workforce *sagemaker.Workforce) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker workforce ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.WorkforceByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*workforce = *resp

		return nil
	}
}

func testAccAWSSagemakerWorkforceBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name            = %[1]q
  generate_secret = true
  user_pool_id    = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName)
}

func testAccAWSSagemakerWorkforceCognitoConfig(rName string) string {
	return testAccAWSSagemakerWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }
}
`, rName)
}

func testAccAWSSagemakerWorkforceSourceIpConfig1(rName, cidr1 string) string {
	return testAccAWSSagemakerWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }

  source_ip_config {
    cidrs = [%[2]q]
  }
}
`, rName, cidr1)
}

func testAccAWSSagemakerWorkforceSourceIpConfig2(rName, cidr1, cidr2 string) string {
	return testAccAWSSagemakerWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }

  source_ip_config {
    cidrs = [%[2]q, %[3]q]
  }
}
`, rName, cidr1, cidr2)
}

func testAccAWSSagemakerWorkforceOidcConfig(rName, endpoint string) string {
	return testAccAWSSagemakerWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  oidc_config {
    authorization_endpoint = %[2]q
    client_id              = %[1]q
    client_secret          = %[1]q
    issuer                 = %[2]q
    jwks_uri               = %[2]q
    logout_endpoint        = %[2]q
    token_endpoint         = %[2]q
    user_info_endpoint     = %[2]q
  }
}
`, rName, endpoint)
}
