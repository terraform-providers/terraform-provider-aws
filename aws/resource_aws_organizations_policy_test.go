package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsOrganizationsPolicy_basic(t *testing.T) {
	var policy organizations.Policy
	content1 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	content2 := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "s3:*", "Resource": "*"}}`
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Required(rName, content1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile("policy/o-.+/service_control_policy/p-.+$")),
					resource.TestCheckResourceAttr(resourceName, "content", content1),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Required(rName, content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "content", content2),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/5073
func testAccAwsOrganizationsPolicy_concurrent(t *testing.T) {
	var policy1, policy2, policy3, policy4, policy5 organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName1 := "aws_organizations_policy.test1"
	resourceName2 := "aws_organizations_policy.test2"
	resourceName3 := "aws_organizations_policy.test3"
	resourceName4 := "aws_organizations_policy.test4"
	resourceName5 := "aws_organizations_policy.test5"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfigConcurrent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName1, &policy1),
					testAccCheckAwsOrganizationsPolicyExists(resourceName2, &policy2),
					testAccCheckAwsOrganizationsPolicyExists(resourceName3, &policy3),
					testAccCheckAwsOrganizationsPolicyExists(resourceName4, &policy4),
					testAccCheckAwsOrganizationsPolicyExists(resourceName5, &policy5),
				),
			},
		},
	})
}

func testAccAwsOrganizationsPolicy_description(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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

func testAccAwsOrganizationsPolicy_tags(t *testing.T) {
	var p1, p2, p3, p4 organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_TagA(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TerraformProviderAwsTest", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.Alpha", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_TagB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TerraformProviderAwsTest", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.Beta", "1"),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_TagC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &p3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TerraformProviderAwsTest", "true"),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_NoTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &p4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccAwsOrganizationsPolicy_disappears(t *testing.T) {
	var p organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &p),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsOrganizationsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsOrganizationsPolicy_type_AI_OPT_OUT(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_ai-opt-out_syntax.html
	AiOptOutPolicyContent := `{ "services": { "rekognition": { "opt_out_policy": { "@@assign": "optOut" } }, "lex": { "opt_out_policy": { "@@assign": "optIn" } } } }`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Type(rName, AiOptOutPolicyContent, organizations.PolicyTypeAiservicesOptOutPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeAiservicesOptOutPolicy),
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

func testAccAwsOrganizationsPolicy_type_Backup(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"
	// Reference: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_backup_syntax.html
	backupPolicyContent := fmt.Sprintf(`{
   "plans":{
      "PII_Backup_Plan":{
         "regions":{
            "@@assign":[
               "%[1]s",
               "%[2]s"
            ]
         },
         "rules":{
            "Hourly":{
               "schedule_expression":{
                  "@@assign":"cron(0 5/1 ? * * *)"
               },
               "start_backup_window_minutes":{
                  "@@assign":"480"
               },
               "complete_backup_window_minutes":{
                  "@@assign":"10080"
               },
               "lifecycle":{
                  "move_to_cold_storage_after_days":{
                     "@@assign":"180"
                  },
                  "delete_after_days":{
                     "@@assign":"270"
                  }
               },
               "target_backup_vault_name":{
                  "@@assign":"FortKnox"
               },
               "copy_actions":{
                  "arn:%[3]s:backup:%[1]s:$account:backup-vault:secondary_vault":{
                     "target_backup_vault_arn":{
                        "@@assign":"arn:%[3]s:backup:%[1]s:$account:backup-vault:secondary_vault"
                     },
                     "lifecycle":{
                        "delete_after_days":{
                           "@@assign":"100"
                        },
                        "move_to_cold_storage_after_days":{
                           "@@assign":"10"
                        }
                     }
                  }
               }
            }
         },
         "selections":{
            "tags":{
               "datatype":{
                  "iam_role_arn":{
                     "@@assign":"arn:%[3]s:iam::$account:role/MyIamRole"
                  },
                  "tag_key":{
                     "@@assign":"dataType"
                  },
                  "tag_value":{
                     "@@assign":[
                        "PII",
                        "RED"
                     ]
                  }
               }
            }
         }
      }
   }
}`, testAccGetAlternateRegion(), testAccGetRegion(), testAccGetPartition())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Type(rName, backupPolicyContent, organizations.PolicyTypeBackupPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeBackupPolicy),
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

func testAccAwsOrganizationsPolicy_type_SCP(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"
	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Type(rName, serviceControlPolicyContent, organizations.PolicyTypeServiceControlPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeServiceControlPolicy),
				),
			},
			{
				Config: testAccAwsOrganizationsPolicyConfig_Required(rName, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeServiceControlPolicy),
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

func testAccAwsOrganizationsPolicy_type_Tag(t *testing.T) {
	var policy organizations.Policy
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_organizations_policy.test"
	tagPolicyContent := `{ "tags": { "Product": { "tag_key": { "@@assign": "Product" }, "enforced_for": { "@@assign": [ "ec2:instance" ] } } } }`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_Type(rName, tagPolicyContent, organizations.PolicyTypeTagPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "type", organizations.PolicyTypeTagPolicy),
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

func testAccAwsOrganizationsPolicy_ImportAwsManagedPolicy(t *testing.T) {
	resourceName := "aws_organizations_policy.test"

	resourceID := "p-FullAWSAccess"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsPolicyConfig_AwsManagedPolicySetup,
			},
			{
				Config:        testAccAwsOrganizationsPolicyConfig_AwsManagedPolicy,
				ResourceName:  resourceName,
				ImportStateId: resourceID,
				ImportState:   true,
				ExpectError:   regexp.MustCompile(regexp.QuoteMeta(fmt.Sprintf("AWS-managed Organizations policy (%s) cannot be imported.", resourceID))),
			},
		},
	})
}

func testAccCheckAwsOrganizationsPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_policy" {
			continue
		}

		input := &organizations.DescribePolicyInput{
			PolicyId: &rs.Primary.ID,
		}

		resp, err := conn.DescribePolicy(input)

		if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
			continue
		}

		if isAWSErr(err, organizations.ErrCodePolicyNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Policy != nil {
			return fmt.Errorf("Policy %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsOrganizationsPolicyExists(resourceName string, policy *organizations.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		input := &organizations.DescribePolicyInput{
			PolicyId: &rs.Primary.ID,
		}

		resp, err := conn.DescribePolicy(input)

		if err != nil {
			return err
		}

		if resp == nil || resp.Policy == nil {
			return fmt.Errorf("Policy %q does not exist", rs.Primary.ID)
		}

		*policy = *resp.Policy

		return nil
	}
}

func testAccAwsOrganizationsPolicyConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  description = "%s"
  name        = "%s"

  depends_on = [aws_organizations_organization.test]
}
`, description, rName)
}

func testAccAwsOrganizationsPolicyConfig_TagA(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = "%s"

  depends_on = [aws_organizations_organization.test]

  tags = {
    TerraformProviderAwsTest = true
    Alpha                    = 1
  }
}
`, rName)
}

func testAccAwsOrganizationsPolicyConfig_TagB(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = "%s"

  depends_on = [aws_organizations_organization.test]

  tags = {
    TerraformProviderAwsTest = true
    Beta                     = 1
  }
}
`, rName)
}

func testAccAwsOrganizationsPolicyConfig_TagC(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = "%s"

  depends_on = [aws_organizations_organization.test]

  tags = {
    TerraformProviderAwsTest = true
  }
}
`, rName)
}

func testAccAwsOrganizationsPolicyConfig_NoTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = "%s"

  depends_on = [aws_organizations_organization.test]
}
`, rName)
}

func testAccAwsOrganizationsPolicyConfig_Required(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = %s
  name    = "%s"

  depends_on = [aws_organizations_organization.test]
}
`, strconv.Quote(content), rName)
}

func testAccAwsOrganizationsPolicyConfigConcurrent(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test1" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "cloudtrail:StopLogging",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s1"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test2" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "ec2:DeleteFlowLogs",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s2"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test3" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "logs:DeleteLogGroup",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s3"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test4" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "config:DeleteConfigRule",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s4"

  depends_on = [aws_organizations_organization.test]
}

resource "aws_organizations_policy" "test5" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Deny",
    "Action": "iam:DeleteRolePermissionsBoundary",
    "Resource": "*"
  }
}
EOF

  name = "%[1]s5"

  depends_on = [aws_organizations_organization.test]
}
`, rName)
}

func testAccAwsOrganizationsPolicyConfig_Type(rName, content, policyType string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  content = %s
  name    = "%s"
  type    = "%s"

  depends_on = [aws_organizations_organization.test]
}
`, strconv.Quote(content), rName, policyType)
}

const testAccAwsOrganizationsPolicyConfig_AwsManagedPolicySetup = `
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}
`

const testAccAwsOrganizationsPolicyConfig_AwsManagedPolicy = `
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}

resource "aws_organizations_policy" "test" {
  name = "FullAWSAccess"
}
`
