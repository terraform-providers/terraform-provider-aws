package aws

import (
	"fmt"
	"strings"
	"testing"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMRole_basic(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

func TestAccAWSIAMRole_force_detach_policies(t *testing.T) {
	var conf iam.GetRoleOutput
	rName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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
      "Resource": "arn:aws:s3:::*"
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
      "Resource": "arn:aws:s3:::*"
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
