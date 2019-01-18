package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iam_role", &resource.Sweeper{
		Name: "aws_iam_role",
		Dependencies: []string{
			"aws_batch_compute_environment",
			"aws_cognito_user_pool",
			"aws_config_configuration_aggregator",
			"aws_config_configuration_recorder",
			"aws_datasync_location_s3",
			"aws_dax_cluster",
			"aws_db_instance",
			"aws_db_option_group",
			"aws_eks_cluster",
			"aws_elastic_beanstalk_application",
			"aws_elastic_beanstalk_environment",
			"aws_elasticsearch_domain",
			"aws_glue_crawler",
			"aws_glue_job",
			"aws_instance",
			"aws_lambda_function",
			"aws_launch_configuration",
			"aws_redshift_cluster",
			"aws_spot_fleet_request",
		},
		F: testSweepIamRoles,
	})
}

func testSweepIamRoles(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).iamconn
	prefixes := []string{
		"ecs_instance_role",
		"ecs_tf",
		"EMR_AutoScaling_DefaultRole",
		"iam_emr",
		"terraform-",
		"test_role",
		"tf",
	}
	roles := make([]*iam.Role, 0)

	err = conn.ListRolesPages(&iam.ListRolesInput{}, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, role := range page.Roles {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(role.RoleName), prefix) {
					roles = append(roles, role)
					break
				}
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Role sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving IAM Roles: %s", err)
	}

	if len(roles) == 0 {
		log.Print("[DEBUG] No IAM Roles to sweep")
		return nil
	}

	for _, role := range roles {
		rolename := aws.StringValue(role.RoleName)
		log.Printf("[DEBUG] Deleting IAM Role: %s", rolename)

		if err := deleteAwsIamRoleInstanceProfiles(conn, rolename); err != nil {
			return fmt.Errorf("error deleting IAM Role (%s) instance profiles: %s", rolename, err)
		}

		if err := deleteAwsIamRolePolicyAttachments(conn, rolename); err != nil {
			return fmt.Errorf("error deleting IAM Role (%s) policy attachments: %s", rolename, err)
		}

		if err := deleteAwsIamRolePolicies(conn, rolename); err != nil {
			return fmt.Errorf("error deleting IAM Role (%s) policies: %s", rolename, err)
		}

		input := &iam.DeleteRoleInput{
			RoleName: aws.String(rolename),
		}

		_, err := conn.DeleteRole(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("Error deleting IAM Role (%s): %s", rolename, err)
		}
	}

	return nil
}

func TestAccAWSIAMRole_importBasic(t *testing.T) {
	resourceName := "aws_iam_role.role"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIAMRole_basic(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role", &conf),
					resource.TestCheckResourceAttr("aws_iam_role.role", "path", "/"),
					resource.TestCheckResourceAttrSet("aws_iam_role.role", "create_date"),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_basicWithDescription(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfigWithDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role", &conf),
					resource.TestCheckResourceAttr("aws_iam_role.role", "path", "/"),
					resource.TestCheckResourceAttr("aws_iam_role.role", "description", "This 1s a D3scr!pti0n with weird content: &@90ë“‘{«¡Çø}"),
				),
			},
			{
				Config: testAccAWSIAMRoleConfigWithUpdatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role", &conf),
					resource.TestCheckResourceAttr("aws_iam_role.role", "path", "/"),
					resource.TestCheckResourceAttr("aws_iam_role.role", "description", "This 1s an Upd@ted D3scr!pti0n with weird content: &90ë“‘{«¡Çø}"),
				),
			},
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role", &conf),
					resource.TestCheckResourceAttrSet("aws_iam_role.role", "create_date"),
					resource.TestCheckResourceAttr("aws_iam_role.role", "description", ""),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_namePrefix(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_iam_role.role",
		IDRefreshIgnore: []string{"name_prefix"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolePrefixNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role", &conf),
					testAccCheckAWSRoleGeneratedNamePrefix(
						"aws_iam_role.role", "test-role-"),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_testNameChange(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolePre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role_update_test", &conf),
				),
			},

			{
				Config: testAccAWSIAMRolePost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.role_update_test", &conf),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_badJSON(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIAMRoleConfig_badJson(rName),
				ExpectError: regexp.MustCompile(`.*contains an invalid JSON:.*`),
			},
		},
	})
}

func TestAccAWSIAMRole_disappears(t *testing.T) {
	var role iam.GetRoleOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.role"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					testAccCheckAWSRoleDisappears(&role),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMRole_force_detach_policies(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig_force_detach_policies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists("aws_iam_role.test", &conf),
					testAccAddAwsIAMRolePolicy("aws_iam_role.test"),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_MaxSessionDuration(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 3599),
				ExpectError: regexp.MustCompile(`expected max_session_duration to be in the range`),
			},
			{
				Config:      testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 43201),
				ExpectError: regexp.MustCompile(`expected max_session_duration to be in the range`),
			},
			{
				Config: testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 3700),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "max_session_duration", "3700"),
				),
			},
			{
				Config: testAccCheckIAMRoleConfig_MaxSessionDuration(rName, 3701),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "max_session_duration", "3701"),
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

func TestAccAWSIAMRole_PermissionsBoundary(t *testing.T) {
	var role iam.GetRoleOutput

	rName := acctest.RandString(10)
	resourceName := "aws_iam_role.role"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", testAccGetPartition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", testAccGetPartition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckAWSRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test update
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
					testAccCheckAWSRolePermissionsBoundary(&role, permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// Test removal
			{
				Config: testAccAWSIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckAWSRolePermissionsBoundary(&role, ""),
				),
			},
			// Test addition
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckAWSRolePermissionsBoundary(&role, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccCheckIAMRoleConfig_PermissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckAWSRolePermissionsBoundary(&role, ""),
				),
			},
		},
	})
}

func TestAccAWSIAMRole_tags(t *testing.T) {
	var role iam.GetRoleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRoleConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-value2"),
				),
			},
			{
				Config: testAccAWSIAMRoleConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRoleExists(resourceName, &role),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-value"),
				),
			},
		},
	})
}

func testAccCheckAWSRoleDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_role" {
			continue
		}

		// Try to get role
		_, err := iamconn.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSRoleExists(n string, res *iam.GetRoleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAWSRoleDisappears(getRoleOutput *iam.GetRoleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		roleName := aws.StringValue(getRoleOutput.Role.RoleName)

		_, err := iamconn.DeleteRole(&iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			return fmt.Errorf("error deleting role %q: %s", roleName, err)
		}

		return nil
	}
}

func testAccCheckAWSRoleGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

// Attach inline policy outside of terraform CRUD.
func testAccAddAwsIAMRolePolicy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		input := &iam.PutRolePolicyInput{
			RoleName: aws.String(rs.Primary.ID),
			PolicyDocument: aws.String(`{
			  "Version": "2012-10-17",
			  "Statement": {
			    "Effect": "Allow",
			    "Action": "*",
			    "Resource": "*"
			  }
			}`),
			PolicyName: aws.String(resource.UniqueId()),
		}

		_, err := iamconn.PutRolePolicy(input)
		return err
	}
}

func testAccCheckAWSRolePermissionsBoundary(getRoleOutput *iam.GetRoleOutput, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if getRoleOutput.Role.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *getRoleOutput.Role.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckIAMRoleConfig_MaxSessionDuration(rName string, maxSessionDuration int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name   = "test-role-%s"
	path = "/"
	max_session_duration = %d
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}
`, rName, maxSessionDuration)
}

func testAccCheckIAMRoleConfig_PermissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  assume_role_policy   = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
  name                 = "test-role-%s"
  path                 = "/"
  permissions_boundary = %q
}
`, rName, permissionsBoundary)
}

func testAccAWSIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name   = "test-role-%s"
  path = "/"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}
`, rName)
}

func testAccAWSIAMRoleConfigWithDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name   = "test-role-%s"
  description = "This 1s a D3scr!pti0n with weird content: &@90ë“‘{«¡Çø}"
  path = "/"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}
`, rName)
}

func testAccAWSIAMRoleConfigWithUpdatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name   = "test-role-%s"
  description = "This 1s an Upd@ted D3scr!pti0n with weird content: &90ë“‘{«¡Çø}"
  path = "/"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}
`, rName)
}

func testAccAWSIAMRolePrefixNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name_prefix = "test-role-%s"
  path = "/"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}
`, rName)
}

func testAccAWSIAMRolePre(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role_update_test" {
  name = "tf_old_name_%s"
  path = "/test/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "role_update_test" {
  name = "role_update_test_%s"
  role = "${aws_iam_role.role_update_test.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "role_update_test" {
  name = "role_update_test_%s"
  path = "/test/"
  roles = ["${aws_iam_role.role_update_test.name}"]
}
`, rName, rName, rName)
}

func testAccAWSIAMRolePost(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role_update_test" {
  name = "tf_new_name_%s"
  path = "/test/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "role_update_test" {
  name = "role_update_test_%s"
  role = "${aws_iam_role.role_update_test.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::*"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "role_update_test" {
  name = "role_update_test_%s"
  path = "/test/"
  roles = ["${aws_iam_role.role_update_test.name}"]
}
`, rName, rName, rName)
}

func testAccAWSIAMRoleConfig_badJson(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "my_instance_role" {
  name = "test-role-%s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
  {
    "Action": "sts:AssumeRole",
    "Principal": {
    "Service": "ec2.amazonaws.com",
    },
    "Effect": "Allow",
    "Sid": ""
  }
  ]
}
POLICY
}
`, rName)
}

func testAccAWSIAMRoleConfig_force_detach_policies(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "test" {
  name = "tf-iam-role-policy-%s"
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = "tf-iam-policy-%s"
  description = "A test policy"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
      "iam:ChangePassword"
    ],
    "Resource": "*",
    "Effect": "Allow"
  }
]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role = "${aws_iam_role.test.name}"
  policy_arn = "${aws_iam_policy.test.arn}"
}

resource "aws_iam_role" "test" {
  name = "tf-iam-role-%s"
  force_detach_policies = true
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName, rName, rName)
}

func testAccAWSIAMRoleConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}

`, rName)
}

func testAccAWSIAMRoleConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    tag2 = "test-value"
  }
}
`, rName)
}
