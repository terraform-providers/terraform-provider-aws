package aws

import (
	"errors"
	"fmt"
	"log"
	"os"
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
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

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
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	// This test must always run in us-east-1
	// BadRequestException: Invalid certificate ARN: arn:aws:acm:us-west-2:123456789012:certificate/xxxxx. Certificate must be in 'us-east-1'.
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	customDomainName := os.Getenv("AWS_COGNITO_USER_POOL_DOMAIN_ROOT_DOMAIN")
	if customDomainName == "" {
		t.Skip(
			"Environment variable AWS_COGNITO_USER_POOL_DOMAIN_ROOT_DOMAIN is not set. " +
				"This environment variable must be set to the fqdn of " +
				"an ISSUED ACM certificate in us-east-1 to enable this test.")
	}

	customSubDomainName := fmt.Sprintf("%s.%s", fmt.Sprintf("tf-acc-test-domain-%d", acctest.RandInt()), customDomainName)
	// For now, use an environment variable to limit running this test
	certificateArn := os.Getenv("AWS_COGNITO_USER_POOL_DOMAIN_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skip(
			"Environment variable AWS_COGNITO_USER_POOL_DOMAIN_CERTIFICATE_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"an ISSUED ACM certificate in us-east-1 to enable this test.")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolDomainConfig_custom(customSubDomainName, poolName, certificateArn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolDomainExists("aws_cognito_user_pool_domain.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_domain.main", "domain", customSubDomainName),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_domain.main", "certificate_arn", certificateArn),
					resource.TestCheckResourceAttr("aws_cognito_user_pool.main", "name", poolName),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "aws_account_id"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "cloudfront_distribution_arn"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "s3_bucket"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_pool_domain.main", "version"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserPoolDomain_disappears(t *testing.T) {
	domainName := fmt.Sprintf("tf-acc-test-domain-%d", acctest.RandInt())
	poolName := fmt.Sprintf("tf-acc-test-pool-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
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

func testAccAWSCognitoUserPoolDomainConfig_custom(customSubDomainName, poolName, certificateArn string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool_domain" "main" {
  domain          = "%s"
  user_pool_id    = aws_cognito_user_pool.main.id
  certificate_arn = "%s"
}

resource "aws_cognito_user_pool" "main" {
  name = "%s"
}
`, customSubDomainName, certificateArn, poolName)
}
