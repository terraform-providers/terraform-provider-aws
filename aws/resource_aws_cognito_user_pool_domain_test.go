package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cognito_user_pool_domain", &resource.Sweeper{
		Name: "aws_cognito_user_pool_domain",
		F:    testSweepCognitoUserPoolDomains,
	})
}

func testSweepCognitoUserPoolDomains(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).cognitoidpconn

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(int64(50)),
	}

	err = conn.ListUserPoolsPages(input, func(resp *cognitoidentityprovider.ListUserPoolsOutput, isLast bool) bool {
		if len(resp.UserPools) == 0 {
			log.Print("[DEBUG] No Cognito user pools (i.e. domains) to sweep")
			return false
		}

		for _, u := range resp.UserPools {
			output, err := conn.DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{
				UserPoolId: u.Id,
			})
			if err != nil {
				log.Printf("[ERROR] Failed describing Cognito user pool (%s): %s", aws.StringValue(u.Name), err)
				continue
			}
			if output.UserPool != nil && output.UserPool.Domain != nil {
				domain := aws.StringValue(output.UserPool.Domain)

				log.Printf("[INFO] Deleting Cognito user pool domain: %s", domain)
				_, err := conn.DeleteUserPoolDomain(&cognitoidentityprovider.DeleteUserPoolDomainInput{
					Domain:     output.UserPool.Domain,
					UserPoolId: u.Id,
				})
				if err != nil {
					log.Printf("[ERROR] Failed deleting Cognito user pool domain (%s): %s", domain, err)
				}
			}
		}
		return !isLast
	})

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Cognito User Pool Domain sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Cognito User Pools: %s", err)
	}

	return nil
}

func TestAccAWSCognitoUserPoolDomain_basic(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-test-domain-%d", acctest.RandInt())
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolDomainConfig_basic(domainName, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolDomainExists("aws_cognito_user_pool_domain.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_domain.main", "domain", domainName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "name", poolName),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "aws_account_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "cloudfront_distribution_arn"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "s3_bucket"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "version"),
				),
			},
			{
				ResourceName:      "aws_cognito_user_pool_domain.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPoolDomain_custom(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandString(10))

	acmCertificateResourceName := "aws_acm_certificate.test"
	cognitoUserPoolResourceName := "aws_cognito_user_pool.test"
	resourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckCognitoUserPoolCustomDomain(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSCognitoUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolDomainConfig_custom(rootDomain, domain, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolDomainExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "aws_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
					//lintignore:AWSAT001 // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11666
					resource.TestMatchResourceAttr(resourceName, "cloudfront_distribution_arn", regexp.MustCompile(`[a-z0-9]+.cloudfront.net$`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain", acmCertificateResourceName, "domain_name"),
					resource.TestMatchResourceAttr(resourceName, "s3_bucket", regexp.MustCompile(`^.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", cognitoUserPoolResourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile(`^.+$`)),
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

func TestAccAWSCognitoUserPoolDomain_disappears(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-test-domain-%d", acctest.RandInt())
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandString(10))
	resourceName := "aws_cognito_user_pool_domain.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolDomainConfig_basic(domainName, poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolDomainExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCognitoUserPoolDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCognitoUserPoolDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		_, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSCognitoUserPoolDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_domain" {
			continue
		}

		_, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccAWSCognitoUserPoolDomainConfig_basic(domainName, poolName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_domain" "main" {
  domain       = "%s"
  user_pool_id = aws_cognito_user_pool.main.id
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, domainName, poolName)
}

func testAccAWSCognitoUserPoolDomainConfig_custom(rootDomain string, domain string, poolName string) string {
	return composeConfig(
		testAccCognitoUserPoolCustomDomainRegionProviderConfig(),
		fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

#
# for_each acceptance testing requires:
# https://github.com/hashicorp/terraform-plugin-sdk/issues/536
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }

#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
# }

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

resource "aws_cognito_user_pool" "test" {
  name = %[3]q
}

resource "aws_cognito_user_pool_domain" "test" {
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  domain          = aws_acm_certificate.test.domain_name
  user_pool_id    = aws_cognito_user_pool.test.id
}
`, rootDomain, domain, poolName))
}
