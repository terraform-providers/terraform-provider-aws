package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSTransferSshKey_basic(t *testing.T) {
	var conf transfer.SshPublicKey
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferSshKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferSshKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferSshKeyExists("aws_transfer_ssh_key.foo", &conf),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_ssh_key.foo", "server_id", "aws_transfer_server.foo", "id"),
					resource.TestCheckResourceAttrPair(
						"aws_transfer_ssh_key.foo", "user_name", "aws_transfer_user.foo", "user_name"),
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

func testAccCheckAWSTransferSshKeyExists(n string, res *transfer.SshPublicKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Ssh Public Key ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn
		serverID, userName, sshKeyID, err := decodeTransferSshKeyId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Transfer SSH Public Key ID: %s", err)
		}

		describe, err := conn.DescribeUser(&transfer.DescribeUserInput{
			ServerId: aws.String(serverID),
			UserName: aws.String(userName),
		})

		if err != nil {
			return err
		}

		for _, sshPublicKey := range describe.User.SshPublicKeys {
			if sshKeyID == *sshPublicKey.SshPublicKeyId {
				*res = *sshPublicKey
				return nil
			}
		}

		return fmt.Errorf("Transfer Ssh Public Key doesn't exists.")
	}
}

func testAccCheckAWSTransferSshKeyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_ssh_key" {
			continue
		}
		serverID, userName, sshKeyID, err := decodeTransferSshKeyId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Transfer SSH Public Key ID: %s", err)
		}

		describe, err := conn.DescribeUser(&transfer.DescribeUserInput{
			UserName: aws.String(userName),
			ServerId: aws.String(serverID),
		})

		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		for _, sshPublicKey := range describe.User.SshPublicKeys {
			if sshKeyID == *sshPublicKey.SshPublicKeyId {
				return fmt.Errorf("Transfer SSH Public Key still exists")
			}
		}
	}

	return nil
}

func testAccAWSTransferSshKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"
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


resource "aws_transfer_ssh_key" "foo" {
	server_id = "${aws_transfer_server.foo.id}"
	user_name = "${aws_transfer_user.foo.user_name}"
	body 	  = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
	`, rName, rName)
}
