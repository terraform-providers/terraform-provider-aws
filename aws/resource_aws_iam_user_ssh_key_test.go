package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSUserSSHKey_basic(t *testing.T) {
	var conf iam.GetSSHPublicKeyOutput

	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccAWSSSHKeyConfig_sshEncoding, ri)
	resourceName := "aws_iam_user_ssh_key.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserSSHKeyExists(resourceName, "Inactive", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSUserSSHKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSUserSSHKey_pemEncoding(t *testing.T) {
	var conf iam.GetSSHPublicKeyOutput

	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccAWSSSHKeyConfig_pemEncoding, ri)
	resourceName := "aws_iam_user_ssh_key.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserSSHKeyExists(resourceName, "Active", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSUserSSHKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSUserSSHKeyDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user_ssh_key" {
			continue
		}

		username := rs.Primary.Attributes["username"]
		encoding := rs.Primary.Attributes["encoding"]
		_, err := iamconn.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(rs.Primary.ID),
			UserName:       aws.String(username),
			Encoding:       aws.String(encoding),
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

func testAccCheckAWSUserSSHKeyExists(n, status string, res *iam.GetSSHPublicKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSHPublicKeyID is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		username := rs.Primary.Attributes["username"]
		encoding := rs.Primary.Attributes["encoding"]
		resp, err := iamconn.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(rs.Primary.ID),
			UserName:       aws.String(username),
			Encoding:       aws.String(encoding),
		})
		if err != nil {
			return err
		}

		*res = *resp

		keyStruct := resp.SSHPublicKey

		if *keyStruct.Status != status {
			return fmt.Errorf("Key status has wrong status should be %s is %s", status, *keyStruct.Status)
		}

		return nil
	}
}

func testAccAWSUserSSHKeyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		username := rs.Primary.Attributes["username"]
		sshPublicKeyId := rs.Primary.Attributes["ssh_public_key_id"]
		encoding := rs.Primary.Attributes["encoding"]

		return fmt.Sprintf("%s:%s:%s", username, sshPublicKeyId, encoding), nil
	}
}

const testAccAWSSSHKeyConfig_sshEncoding = `
resource "aws_iam_user" "user" {
	name = "test-user-%d"
	path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
	username = "${aws_iam_user.user.name}"
	encoding = "SSH"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	status = "Inactive"
}
`

const testAccAWSSSHKeyConfig_pemEncoding = `
resource "aws_iam_user" "user" {
	name = "test-user-%d"
	path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
	username = "${aws_iam_user.user.name}"
	encoding = "PEM"
	public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9xercjxBRM1dC191/AbF
3TLEM9cdnBIpCgxGNGiI+NaoMTAj/4rXp3ql0iBWQaeb4sz72qCEd1JvcSuzxqFv
IIrqRp/hD7sSAOHAzOL8zqjpIjD4c+VytMIRI5Fc06OPktKbrw2bsCLHYlvZsYSX
O7YATS9HGJVkmFZM+Bv37JTX0T1uZmADOPX+H4bcT2+aJOENi4PXTylRzvwYHruc
KDHO0WNKdXo+g+AihROpcpkgyaVtGB1/8KhPfnHxGroe8WXBtKvbdrWuhen5l9Go
L6RcmaPGhW13lAa+6LEgiTYL2r1mzP9Op4lqzr2F9scFnYV5l0q21/GW2m1aIQSu
NQIDAQAB
-----END PUBLIC KEY-----
EOF
}
`
