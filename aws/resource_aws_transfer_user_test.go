package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSTransferUser_basic(t *testing.T) {
	var conf transfer.DescribedUser
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_user.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists("aws_transfer_user.foo", &conf),
					testAccMatchResourceAttrRegionalARN("aws_transfer_user.foo", "arn", "transfer", regexp.MustCompile(`user/.+`)),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_user.foo", "server_id", "aws_transfer_server.foo", "id"),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_user.foo", "role", "aws_iam_role.foo", "arn"),
				),
			},
			{
				ResourceName:      "aws_transfer_user.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSTransferUser_modifyWithOptions(t *testing.T) {
	var conf transfer.DescribedUser
	rName := acctest.RandString(10)
	rName2 := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_user.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_options(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists("aws_transfer_user.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "home_directory", "/home/tftestuser"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.%", "3"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.NAME", "tftestuser"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.ENV", "test"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.ADMIN", "test"),
				),
			},
			{
				Config: testAccAWSTransferUserConfig_modify(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists("aws_transfer_user.foo", &conf),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_user.foo", "role", "aws_iam_role.foo", "arn"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "home_directory", "/test"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.NAME", "tf-test-user"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "tags.TEST", "test2"),
				),
			},
			{
				Config: testAccAWSTransferUserConfig_forceNew(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists("aws_transfer_user.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "user_name", "tftestuser2"),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_user.foo", "role", "aws_iam_role.foo", "arn"),
					resource.TestCheckResourceAttr(
						"aws_transfer_user.foo", "home_directory", "/home/tftestuser2"),
				),
			},
		},
	})
}

func TestAccAWSTransferUser_disappears(t *testing.T) {
	var serverConf transfer.DescribedServer
	var userConf transfer.DescribedUser
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &serverConf),
					testAccCheckAWSTransferUserExists("aws_transfer_user.foo", &userConf),
					testAccCheckAWSTransferUserDisappears(&serverConf, &userConf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSTransferUserExists(n string, res *transfer.DescribedUser) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer User ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn
		userName := rs.Primary.Attributes["user_name"]
		serverID := rs.Primary.Attributes["server_id"]

		describe, err := conn.DescribeUser(&transfer.DescribeUserInput{
			ServerId: aws.String(serverID),
			UserName: aws.String(userName),
		})

		if err != nil {
			return err
		}

		*res = *describe.User

		return nil
	}
}

func testAccCheckAWSTransferUserDisappears(serverConf *transfer.DescribedServer, userConf *transfer.DescribedUser) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		params := &transfer.DeleteUserInput{
			ServerId: serverConf.ServerId,
			UserName: userConf.UserName,
		}

		_, err := conn.DeleteUser(params)
		if err != nil {
			return err
		}

		return waitForTransferUserDeletion(conn, *serverConf.ServerId, *userConf.UserName)
	}
}

func testAccCheckAWSTransferUserDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_user" {
			continue
		}

		userName := rs.Primary.Attributes["user_name"]
		serverID := rs.Primary.Attributes["server_id"]

		_, err := conn.DescribeUser(&transfer.DescribeUserInput{
			UserName: aws.String(userName),
			ServerId: aws.String(serverID),
		})

		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSTransferUserConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"

	tags = {
		NAME     = "tf-acc-test-transfer-server"
	}
}

resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-user-iam-role-%s"

	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": "transfer.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-user-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"s3:*"
			],
			"Resource": "*"
		}
	]
}
POLICY
}

resource "aws_transfer_user" "foo" {
	server_id      = "${aws_transfer_server.foo.id}"
	user_name      = "tftestuser"
	role           = "${aws_iam_role.foo.arn}"
}
	
`, rName, rName)
}

func testAccAWSTransferUserConfig_options(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"
	
	tags = {
		NAME     = "tf-acc-test-transfer-server"
	}
}

resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-user-iam-role-%s"

	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "transfer.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-user-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoS3",
		"Effect": "Allow",
		"Action": [
			"s3:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}

data "aws_iam_policy_document" "foo" {
	statement {
	  sid = "ListHomeDir"
  
	  actions = [
		"s3:ListBucket",
	  ]
  
	  resources = [
		"arn:aws:s3:::&{transfer:HomeBucket}",
	  ]
	}

	statement {
		sid = "AWSTransferRequirements"
	
		actions = [
		  "s3:ListAllMyBuckets",
		  "s3:GetBucketLocation",
		]
	
		resources = [
		  "*",
		]
	}

	statement {
		sid = "HomeDirObjectAccess"
	
		actions = [
		  "s3:PutObject",
		  "s3:GetObject",
		  "s3:DeleteObjectVersion",
		  "s3:DeleteObject",
		  "s3:GetObjectVersion",
		]
	
		resources = [
		  "arn:aws:s3:::&{transfer:HomeDirectory}*",
		]
	}
}

resource "aws_transfer_user" "foo" {
	server_id      = "${aws_transfer_server.foo.id}"
	user_name      = "tftestuser"
	role           = "${aws_iam_role.foo.arn}"
	policy         = "${data.aws_iam_policy_document.foo.json}"
	home_directory = "/home/tftestuser"

	tags = {
		NAME  = "tftestuser"
		ENV   = "test"
		ADMIN = "test"
	}
}


	`, rName, rName)
}

func testAccAWSTransferUserConfig_modify(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"
	
	tags = {
		NAME     = "tf-acc-test-transfer-server"
	}
}

resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-user-iam-role-%s"
  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "transfer.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-user-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoS3",
		"Effect": "Allow",
		"Action": [
			"s3:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}

data "aws_iam_policy_document" "foo" {
	statement {
	  sid = "ListHomeDir"
  
	  actions = [
		"s3:ListBucket",
	  ]
  
	  resources = [
		"arn:aws:s3:::&{transfer:HomeBucket}",
	  ]
	}

	statement {
		sid = "AWSTransferRequirements"
	
		actions = [
		  "s3:ListAllMyBuckets",
		  "s3:GetBucketLocation",
		]
	
		resources = [
		  "*",
		]
	}

	statement {
		sid = "HomeDirObjectAccess"
	
		actions = [
		  "s3:PutObject",
		  "s3:GetObject",
		  "s3:GetObjectVersion",
		]
	
		resources = [
		  "arn:aws:s3:::&{transfer:HomeDirectory}*",
		]
	}
}

resource "aws_transfer_user" "foo" {
	server_id      = "${aws_transfer_server.foo.id}"
	user_name      = "tftestuser"
	role           = "${aws_iam_role.foo.arn}"
	policy         = "${data.aws_iam_policy_document.foo.json}"
	home_directory = "/test"

	tags = {
		NAME  = "tf-test-user"
		TEST   = "test2"
	}
}


	`, rName, rName)
}

func testAccAWSTransferUserConfig_forceNew(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"
	
	tags = {
		NAME     = "tf-acc-test-transfer-server"
	}
}


resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-user-iam-role-%s"
  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "transfer.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-user-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoS3",
		"Effect": "Allow",
		"Action": [
			"s3:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}

data "aws_iam_policy_document" "foo" {
	statement {
	  sid = "ListHomeDir"
  
	  actions = [
		"s3:ListBucket",
	  ]
  
	  resources = [
		"arn:aws:s3:::&{transfer:HomeBucket}",
	  ]
	}

	statement {
		sid = "AWSTransferRequirements"
	
		actions = [
		  "s3:ListAllMyBuckets",
		  "s3:GetBucketLocation",
		]
	
		resources = [
		  "*",
		]
	}

	statement {
		sid = "HomeDirObjectAccess"
	
		actions = [
		  "s3:PutObject",
		  "s3:GetObject",
		  "s3:DeleteObjectVersion",
		  "s3:DeleteObject",
		  "s3:GetObjectVersion",
		]
	
		resources = [
		  "arn:aws:s3:::&{transfer:HomeDirectory}*",
		]
	}
}

resource "aws_transfer_user" "foo" {
	server_id      = "${aws_transfer_server.foo.id}"
	user_name      = "tftestuser2"
	role           = "${aws_iam_role.foo.arn}"
	policy         = "${data.aws_iam_policy_document.foo.json}"
	home_directory = "/home/tftestuser2"
	tags = {
		NAME  = "tf-test-user"
		TEST   = "test2"
	}
}

	`, rName, rName)
}
