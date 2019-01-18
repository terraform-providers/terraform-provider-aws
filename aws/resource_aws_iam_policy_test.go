package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMPolicy_basic(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "policy", `{"Version":"2012-10-17","Statement":[{"Action":["ec2:Describe*"],"Effect":"Allow","Resource":"*"}]}`),
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

func TestAccAWSIAMPolicy_description(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func TestAccAWSIAMPolicy_disappears(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					testAccCheckAWSIAMPolicyDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMPolicy_namePrefix(t *testing.T) {
	var out iam.GetPolicyOutput
	namePrefix := "tf-acc-test-"
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyConfigNamePrefix(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(fmt.Sprintf("^%s", namePrefix))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSIAMPolicy_path(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyConfigPath(rName, "/path1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "path", "/path1/"),
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

func TestAccAWSIAMPolicy_policy(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_policy.test"
	policy1 := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:Describe*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
	policy2 := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIAMPolicyConfigPolicy(rName, "not-json"),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
			{
				Config: testAccAWSIAMPolicyConfigPolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "policy", policy1),
				),
			},
			{
				Config: testAccAWSIAMPolicyConfigPolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "policy", policy2),
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

func testAccCheckAWSIAMPolicyExists(resource string, res *iam.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Policy name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAWSIAMPolicyDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_policy" {
			continue
		}

		_, err := iamconn.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IAM Policy (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSIAMPolicyDisappears(out *iam.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		params := &iam.DeletePolicyInput{
			PolicyArn: out.Policy.Arn,
		}

		_, err := iamconn.DeletePolicy(params)
		return err
	}
}

func testAccAWSIAMPolicyConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  description = %q
  name        = %q
  policy      = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:Describe*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
}
`, description, rName)
}

func testAccAWSIAMPolicyConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name   = %q
  policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:Describe*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
}
`, rName)
}

func testAccAWSIAMPolicyConfigNamePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name_prefix = %q
  policy      = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:Describe*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
}
`, namePrefix)
}

func testAccAWSIAMPolicyConfigPath(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name   = %q
  path   = %q
  policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:Describe*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
}
`, rName, path)
}

func testAccAWSIAMPolicyConfigPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name   = %q
  policy = %q
}
`, rName, policy)
}
