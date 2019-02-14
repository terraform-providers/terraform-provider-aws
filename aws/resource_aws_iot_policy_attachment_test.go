package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIotPolicyAttachment_basic(t *testing.T) {
	policyName := acctest.RandomWithPrefix("PolicyName-")
	policyName2 := acctest.RandomWithPrefix("PolicyName2-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotPolicyAttchmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotPolicyAttachmentConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotPolicyAttachmentExists("aws_iot_policy_attachment.att"),
					testAccCheckAWSIotPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName}),
				),
			},
			{
				Config: testAccAWSIotPolicyAttachmentConfigUpdate1(policyName, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotPolicyAttachmentExists("aws_iot_policy_attachment.att"),
					testAccCheckAWSIotPolicyAttachmentExists("aws_iot_policy_attachment.att2"),
					testAccCheckAWSIotPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName, policyName2}),
				),
			},
			{
				Config: testAccAWSIotPolicyAttachmentConfigUpdate2(policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotPolicyAttachmentExists("aws_iot_policy_attachment.att2"),
					testAccCheckAWSIotPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName2}),
				),
			},
			{
				Config: testAccAWSIotPolicyAttachmentConfigUpdate3(policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotPolicyAttachmentExists("aws_iot_policy_attachment.att2"),
					testAccCheckAWSIotPolicyAttachmentExists("aws_iot_policy_attachment.att3"),
					testAccCheckAWSIotPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName2}),
					testAccCheckAWSIotPolicyAttachmentCertStatus("aws_iot_certificate.cert2", []string{policyName2}),
				),
			},
		},
	})

}

func testAccCheckAWSIotPolicyAttchmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_policy_attachment" {
			continue
		}

		target := rs.Primary.Attributes["target"]
		policyName := rs.Primary.Attributes["policy"]

		input := &iot.ListAttachedPoliciesInput{
			PageSize:  aws.Int64(250),
			Recursive: aws.Bool(false),
			Target:    aws.String(target),
		}

		var policy *iot.Policy
		err := listIotPolicyAttachmentPages(conn, input, func(out *iot.ListAttachedPoliciesOutput, lastPage bool) bool {
			for _, att := range out.Policies {
				if policyName == aws.StringValue(att.PolicyName) {
					policy = att
					return false
				}
			}
			return true
		})

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "The certificate given in the principal does not exist.") {
			continue
		} else if err != nil {
			return err
		}

		if policy == nil {
			continue
		}

		return fmt.Errorf("IOT Policy Attachment (%s) still exists", rs.Primary.Attributes["id"])
	}
	return nil
}

func testAccCheckAWSIotPolicyAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		target := rs.Primary.Attributes["target"]
		policyName := rs.Primary.Attributes["policy"]

		policy, err := getIotPolicyAttachment(conn, target, policyName)

		if err != nil {
			return fmt.Errorf("Error: Failed to get attached policies for target %s (%s): %s", target, n, err)
		}

		if policy == nil {
			return fmt.Errorf("Error: Policy %s is not attached to target (%s)", policyName, target)
		}

		return nil
	}
}

func testAccCheckAWSIotPolicyAttachmentCertStatus(n string, policies []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotconn

		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		certARN := rs.Primary.Attributes["arn"]

		out, err := conn.ListAttachedPolicies(&iot.ListAttachedPoliciesInput{
			Target:   aws.String(certARN),
			PageSize: aws.Int64(250),
		})

		if err != nil {
			return fmt.Errorf("Error: Cannot list attached policies for target %s: %s", certARN, err)
		}

		if len(out.Policies) != len(policies) {
			return fmt.Errorf("Error: Invalid attached policies count for target %s, expected %d, got %d",
				certARN,
				len(policies),
				len(out.Policies))
		}

		for _, p1 := range policies {
			found := false
			for _, p2 := range out.Policies {
				if p1 == aws.StringValue(p2.PolicyName) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Error: Policy %s is not attached to target %s", p1, certARN)
			}
		}

		return nil
	}
}

func testAccAWSIotPolicyAttachmentConfig(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_policy" "policy" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iot_policy_attachment" "att" {
  policy = "${aws_iot_policy.policy.name}"
  target = "${aws_iot_certificate.cert.arn}"
}
`, policyName)
}

func testAccAWSIotPolicyAttachmentConfigUpdate1(policyName, policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_policy" "policy" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iot_policy" "policy2" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iot_policy_attachment" "att" {
  policy = "${aws_iot_policy.policy.name}"
  target = "${aws_iot_certificate.cert.arn}"
}

resource "aws_iot_policy_attachment" "att2" {
  policy = "${aws_iot_policy.policy2.name}"
  target = "${aws_iot_certificate.cert.arn}"
}
`, policyName, policyName2)
}

func testAccAWSIotPolicyAttachmentConfigUpdate2(policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_policy" "policy2" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iot_policy_attachment" "att2" {
  policy = "${aws_iot_policy.policy2.name}"
  target = "${aws_iot_certificate.cert.arn}"
}
`, policyName2)
}

func testAccAWSIotPolicyAttachmentConfigUpdate3(policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_certificate" "cert2" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_policy" "policy2" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iot_policy_attachment" "att2" {
  policy = "${aws_iot_policy.policy2.name}"
  target = "${aws_iot_certificate.cert.arn}"
}

resource "aws_iot_policy_attachment" "att3" {
  policy = "${aws_iot_policy.policy2.name}"
  target = "${aws_iot_certificate.cert2.arn}"
}
`, policyName2)
}
