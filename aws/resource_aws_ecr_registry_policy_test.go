package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrRegistryPolicy_serial(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"basic":      testAccAWSEcrRegistryPolicy_basic,
		"disappears": testAccAWSEcrRegistryPolicy_disappears,
	}

	for name, testFunc := range testFuncs {
		testFunc := testFunc

		t.Run(name, func(t *testing.T) {
			testFunc(t)
		})
	}
}

func testAccAWSEcrRegistryPolicy_basic(t *testing.T) {
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRegistryPolicy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRegistryPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:ReplicateImage".+`)),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrRegistryPolicyUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRegistryPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:ReplicateImage".+`)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:CreateRepository".+`)),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func testAccAWSEcrRegistryPolicy_disappears(t *testing.T) {
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRegistryPolicy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRegistryPolicyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcrRegistryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEcrRegistryPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_registry_policy" {
			continue
		}

		_, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSEcrRegistryPolicyExists(name string, res *ecr.GetRegistryPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR registry policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ecrconn

		output, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
				return fmt.Errorf("ECR repository %s not found", rs.Primary.ID)
			}
			return err
		}

		*res = *output

		return nil
	}
}

func testAccAWSEcrRegistryPolicy() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "testpolicy",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : [
          "ecr:ReplicateImage"
        ],
        "Resource" : [
          "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
        ]
      }
    ]
  })
}
`
}

func testAccAWSEcrRegistryPolicyUpdated() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "testpolicy",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : [
          "ecr:ReplicateImage",
          "ecr:CreateRepository"
        ],
        "Resource" : [
          "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
        ]
      }
    ]
  })
}
`
}
