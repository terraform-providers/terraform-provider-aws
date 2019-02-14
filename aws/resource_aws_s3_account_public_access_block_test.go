package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// S3 account-level settings must run serialized
// for TeamCity environment
func TestAccAWSS3Account(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"PublicAccessBlock": {
			"basic":                 testAccAWSS3AccountPublicAccessBlock_basic,
			"disappears":            testAccAWSS3AccountPublicAccessBlock_disappears,
			"AccountId":             testAccAWSS3AccountPublicAccessBlock_AccountId,
			"BlockPublicAcls":       testAccAWSS3AccountPublicAccessBlock_BlockPublicAcls,
			"BlockPublicPolicy":     testAccAWSS3AccountPublicAccessBlock_BlockPublicPolicy,
			"IgnorePublicAcls":      testAccAWSS3AccountPublicAccessBlock_IgnorePublicAcls,
			"RestrictPublicBuckets": testAccAWSS3AccountPublicAccessBlock_RestrictPublicBuckets,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					// Explicitly sleep between tests for eventual consistency
					time.Sleep(5 * time.Second)
				})
			}
		})
	}
}

func testAccAWSS3AccountPublicAccessBlock_basic(t *testing.T) {
	var configuration1 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
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

func testAccAWSS3AccountPublicAccessBlock_disappears(t *testing.T) {
	var configuration1 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					testAccCheckAWSS3AccountPublicAccessBlockDisappears(),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSS3AccountPublicAccessBlock_AccountId(t *testing.T) {
	var configuration1 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigAccountId(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
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

func testAccAWSS3AccountPublicAccessBlock_BlockPublicAcls(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigBlockPublicAcls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigBlockPublicAcls(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
				),
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigBlockPublicAcls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
		},
	})
}

func testAccAWSS3AccountPublicAccessBlock_BlockPublicPolicy(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigBlockPublicPolicy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigBlockPublicPolicy(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
				),
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigBlockPublicPolicy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
		},
	})
}

func testAccAWSS3AccountPublicAccessBlock_IgnorePublicAcls(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigIgnorePublicAcls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigIgnorePublicAcls(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
				),
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigIgnorePublicAcls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
		},
	})
}

func testAccAWSS3AccountPublicAccessBlock_RestrictPublicBuckets(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3AccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigRestrictPublicBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigRestrictPublicBuckets(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
				),
			},
			{
				Config: testAccAWSS3AccountPublicAccessBlockConfigRestrictPublicBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
		},
	})
}

func testAccCheckAWSS3AccountPublicAccessBlockExists(resourceName string, configuration *s3control.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Account Public Access Block ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).s3controlconn

		input := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(rs.Primary.ID),
		}

		// Retry for eventual consistency
		var output *s3control.GetPublicAccessBlockOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.GetPublicAccessBlock(input)

			if isAWSErr(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration, "") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		if output == nil || output.PublicAccessBlockConfiguration == nil {
			return fmt.Errorf("S3 Account Public Access Block not found")
		}

		*configuration = *output.PublicAccessBlockConfiguration

		return nil
	}
}

func testAccCheckAWSS3AccountPublicAccessBlockDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3controlconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_account_public_access_block" {
			continue
		}

		input := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(rs.Primary.ID),
		}

		// Retry for eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.GetPublicAccessBlock(input)

			if isAWSErr(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration, "") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("S3 Account Public Access Block (%s) still exists", rs.Primary.ID))
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSS3AccountPublicAccessBlockDisappears() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).s3controlconn
		accountID := testAccProvider.Meta().(*AWSClient).accountid

		deleteInput := &s3control.DeletePublicAccessBlockInput{
			AccountId: aws.String(accountID),
		}

		_, err := conn.DeletePublicAccessBlock(deleteInput)

		if err != nil {
			return err
		}

		getInput := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(accountID),
		}

		// Retry for eventual consistency
		return resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.GetPublicAccessBlock(getInput)

			if isAWSErr(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration, "") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("S3 Account Public Access Block (%s) still exists", accountID))
		})
	}
}

func testAccAWSS3AccountPublicAccessBlockConfig() string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {}
`)
}

func testAccAWSS3AccountPublicAccessBlockConfigAccountId() string {
	return fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_s3_account_public_access_block" "test" {
  account_id = "${data.aws_caller_identity.test.account_id}"
}
`)
}

func testAccAWSS3AccountPublicAccessBlockConfigBlockPublicAcls(blockPublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_acls = %t
}
`, blockPublicAcls)
}

func testAccAWSS3AccountPublicAccessBlockConfigBlockPublicPolicy(blockPublicPolicy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_policy = %t
}
`, blockPublicPolicy)
}

func testAccAWSS3AccountPublicAccessBlockConfigIgnorePublicAcls(ignorePublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  ignore_public_acls = %t
}
`, ignorePublicAcls)
}

func testAccAWSS3AccountPublicAccessBlockConfigRestrictPublicBuckets(restrictPublicBuckets bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  restrict_public_buckets = %t
}
`, restrictPublicBuckets)
}
