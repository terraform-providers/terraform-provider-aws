package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/transfer"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSTransferServer_basic(t *testing.T) {
	var conf transfer.DescribedServer
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					testAccMatchResourceAttrRegionalARN("aws_transfer_server.foo", "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestMatchResourceAttr(
						"aws_transfer_server.foo", "endpoint", regexp.MustCompile(fmt.Sprintf("^s-[a-z0-9]+.server.transfer.%s.amazonaws.com$", testAccGetRegion()))),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr("aws_transfer_server.foo", "tags.%", "0"),
				),
			},
			{
				ResourceName:            "aws_transfer_server.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerConfig_basicUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.NAME", "tf-acc-test-transfer-server"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.ENV", "test"),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_server.foo", "logging_role", "aws_iam_role.foo", "arn"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_apigateway(t *testing.T) {
	var conf transfer.DescribedServer
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_apigateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrSet(
						"aws_transfer_server.foo", "invocation_role"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.NAME", "tf-acc-test-transfer-server"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "tags.TYPE", "apigateway"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_disappears(t *testing.T) {
	var conf transfer.DescribedServer

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					testAccCheckAWSTransferServerDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSTransferServer_forcedestroy(t *testing.T) {
	var conf transfer.DescribedServer
	var roleConf iam.GetRoleOutput
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_transfer_server.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_forcedestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					testAccCheckAWSRoleExists("aws_iam_role.foo", &roleConf),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(
						"aws_transfer_server.foo", "force_destroy", "true"),
					testAccCheckAWSTransferCreateUser(&conf, &roleConf, rName),
					testAccCheckAWSTransferCreateSshKey(&conf, rName),
				),
			},
			{
				ResourceName:            "aws_transfer_server.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccCheckAWSTransferServerExists(n string, res *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Server ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn

		describe, err := conn.DescribeServer(&transfer.DescribeServerInput{
			ServerId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *describe.Server

		return nil
	}
}

func testAccCheckAWSTransferServerDisappears(conf *transfer.DescribedServer) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		params := &transfer.DeleteServerInput{
			ServerId: conf.ServerId,
		}

		_, err := conn.DeleteServer(params)
		if err != nil {
			return err
		}

		return waitForTransferServerDeletion(conn, *conf.ServerId)
	}
}

func testAccCheckAWSTransferServerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_server" {
			continue
		}

		_, err := conn.DescribeServer(&transfer.DescribeServerInput{
			ServerId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err == nil {
			return fmt.Errorf("Transfer Server (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSTransferCreateUser(describedServer *transfer.DescribedServer, getRoleOutput *iam.GetRoleOutput, userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		input := &transfer.CreateUserInput{
			ServerId: describedServer.ServerId,
			UserName: aws.String(userName),
			Role:     getRoleOutput.Role.Arn,
		}

		if _, err := conn.CreateUser(input); err != nil {
			return fmt.Errorf("error creating Transfer User (%s) on Server (%s): %s", userName, aws.StringValue(describedServer.ServerId), err)
		}

		return nil
	}
}

func testAccCheckAWSTransferCreateSshKey(describedServer *transfer.DescribedServer, userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		input := &transfer.ImportSshPublicKeyInput{
			ServerId:         describedServer.ServerId,
			UserName:         aws.String(userName),
			SshPublicKeyBody: aws.String("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"),
		}

		if _, err := conn.ImportSshPublicKey(input); err != nil {
			return fmt.Errorf("error creating Transfer SSH Public Key for  (%s/%s): %s", userName, aws.StringValue(describedServer.ServerId), err)
		}

		return nil
	}
}

const testAccAWSTransferServerConfig_basic = `
resource "aws_transfer_server" "foo" {}
`

func testAccAWSTransferServerConfig_basicUpdate(rName string) string {

	return fmt.Sprintf(`

resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-server-iam-role-%s"
  
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
	name = "tf-test-transfer-server-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoCloudWatchLogs",
		"Effect": "Allow",
		"Action": [
			"logs:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}

resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role = "${aws_iam_role.foo.arn}"

  tags = {
	NAME   = "tf-acc-test-transfer-server"
	ENV    = "test"
  }
}
`, rName, rName)
}

func testAccAWSTransferServerConfig_apigateway(rName string) string {

	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
	name = "test"
}

resource "aws_api_gateway_resource" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	parent_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
	path_part = "test"
}

resource "aws_api_gateway_method" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "GET"
	authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "${aws_api_gateway_method.test.http_method}"
	status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "${aws_api_gateway_method.test.http_method}"

	type = "HTTP"
	uri = "https://www.google.de"
	integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	resource_id = "${aws_api_gateway_resource.test.id}"
	http_method = "${aws_api_gateway_integration.test.http_method}"
	status_code = "${aws_api_gateway_method_response.error.status_code}"
}

resource "aws_api_gateway_deployment" "test" {
	depends_on = ["aws_api_gateway_integration.test"]

	rest_api_id = "${aws_api_gateway_rest_api.test.id}"
	stage_name = "test"
	description = "%s"
	stage_description = "%s"


  variables = {
    "a" = "2"
  }
}


resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-server-iam-role-for-apigateway-%s"

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
	name = "tf-test-transfer-server-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoCloudWatchLogs",
		"Effect": "Allow",
		"Action": [
			"logs:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}

resource "aws_transfer_server" "foo" {
	identity_provider_type	= "API_GATEWAY"
	url 				   	= "https://${aws_api_gateway_rest_api.test.id}.execute-api.us-west-2.amazonaws.com${aws_api_gateway_resource.test.path}"
	invocation_role 	   	= "${aws_iam_role.foo.arn}"
	logging_role 		   	= "${aws_iam_role.foo.arn}"

	tags = {
	  NAME     = "tf-acc-test-transfer-server"
	  TYPE	   = "apigateway"
	}
}
`, rName, rName, rName, rName)

}

func testAccAWSTransferServerConfig_forcedestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
	force_destroy = true
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
`, rName, rName)
}
