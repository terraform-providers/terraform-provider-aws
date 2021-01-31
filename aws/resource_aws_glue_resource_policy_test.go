package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func CreateTablePolicy(action string) string {
	return fmt.Sprintf(`{
  "Version" : "2012-10-17",
  "Statement" : [
    {
      "Effect" : "Allow",
      "Action" : [
        "%s"
      ],
      "Principal" : {
         "AWS": "*"
       },
      "Resource" : "arn:%s:glue:%s:%s:*"
    }
  ]
}`, action, testAccGetPartition(), testAccGetRegion(), testAccGetAccountID())
}

func testAccAWSGlueResourcePolicy_basic(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueResourcePolicy_Required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSGlueResourcePolicy(resourceName, "glue:CreateTable"),
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

func testAccAWSGlueResourcePolicy_disappears(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueResourcePolicy_Required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSGlueResourcePolicy(resourceName, "glue:CreateTable"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSGlueResourcePolicy_Required(action string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = ["%s"]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_glue_resource_policy" "test" {
  policy = data.aws_iam_policy_document.glue-example-policy.json
}
`, action)
}

func testAccAWSGlueResourcePolicy_update(t *testing.T) {
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueResourcePolicy_Required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSGlueResourcePolicy(resourceName, "glue:CreateTable"),
				),
			},
			{
				Config: testAccAWSGlueResourcePolicy_Required("glue:DeleteTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSGlueResourcePolicy(resourceName, "glue:DeleteTable"),
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

func testAccAWSGlueResourcePolicy(n string, action string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy id set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		policy, err := conn.GetResourcePolicy(&glue.GetResourcePolicyInput{})
		if err != nil {
			return fmt.Errorf("Get resource policy error: %v", err)
		}

		actualPolicyText := *policy.PolicyInJson

		expectedPolicy := CreateTablePolicy(action)
		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicy)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicy, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSGlueResourcePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	policy, err := conn.GetResourcePolicy(&glue.GetResourcePolicyInput{})

	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "Policy not found") {
			return nil
		}
		return err
	}

	if *policy.PolicyInJson != "" {
		return fmt.Errorf("Aws glue resource policy still exists: %s", *policy.PolicyInJson)
	}
	return nil
}
