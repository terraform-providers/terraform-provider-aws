package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLoadBalancerBackendServerPolicy_basic(t *testing.T) {
	privateKey1 := tlsRsaPrivateKeyPem(2048)
	privateKey2 := tlsRsaPrivateKeyPem(2048)
	publicKey1 := tlsRsaPublicKeyPem(privateKey1)
	publicKey2 := tlsRsaPublicKeyPem(privateKey2)
	certificate1 := tlsRsaX509SelfSignedCertificatePem(privateKey1, "example.com")
	rString := acctest.RandString(8)
	lbName := fmt.Sprintf("tf-acc-lb-bsp-basic-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLoadBalancerBackendServerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLoadBalancerBackendServerPolicyConfig_basic0(lbName, privateKey1, publicKey1, certificate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLoadBalancerPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.test-pubkey-policy0"),
					testAccCheckAWSLoadBalancerPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.test-backend-auth-policy0"),
					testAccCheckAWSLoadBalancerBackendServerPolicyState(lbName, "test-backend-auth-policy0", true),
				),
			},
			{
				Config: testAccAWSLoadBalancerBackendServerPolicyConfig_basic1(lbName, privateKey1, publicKey1, certificate1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLoadBalancerPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.test-pubkey-policy0"),
					testAccCheckAWSLoadBalancerPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.test-pubkey-policy1"),
					testAccCheckAWSLoadBalancerPolicyState("aws_elb.test-lb", "aws_load_balancer_policy.test-backend-auth-policy0"),
					testAccCheckAWSLoadBalancerBackendServerPolicyState(lbName, "test-backend-auth-policy0", true),
				),
			},
			{
				Config: testAccAWSLoadBalancerBackendServerPolicyConfig_basic2(lbName, privateKey1, certificate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLoadBalancerBackendServerPolicyState(lbName, "test-backend-auth-policy0", false),
				),
			},
		},
	})
}

func policyInBackendServerPolicies(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func testAccCheckAWSLoadBalancerBackendServerPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		switch {
		case rs.Type == "aws_load_balancer_policy":
			loadBalancerName, policyName := resourceAwsLoadBalancerBackendServerPoliciesParseId(rs.Primary.ID)
			out, err := conn.DescribeLoadBalancerPolicies(
				&elb.DescribeLoadBalancerPoliciesInput{
					LoadBalancerName: aws.String(loadBalancerName),
					PolicyNames:      []*string{aws.String(policyName)},
				})
			if err != nil {
				if ec2err, ok := err.(awserr.Error); ok && (ec2err.Code() == "PolicyNotFound" || ec2err.Code() == "LoadBalancerNotFound") {
					continue
				}
				return err
			}
			if len(out.PolicyDescriptions) > 0 {
				return fmt.Errorf("Policy still exists")
			}
		case rs.Type == "aws_load_balancer_backend_policy":
			loadBalancerName, policyName := resourceAwsLoadBalancerBackendServerPoliciesParseId(rs.Primary.ID)
			out, err := conn.DescribeLoadBalancers(
				&elb.DescribeLoadBalancersInput{
					LoadBalancerNames: []*string{aws.String(loadBalancerName)},
				})
			if err != nil {
				if ec2err, ok := err.(awserr.Error); ok && (ec2err.Code() == "LoadBalancerNotFound") {
					continue
				}
				return err
			}
			for _, backendServer := range out.LoadBalancerDescriptions[0].BackendServerDescriptions {
				policyStrings := []string{}
				for _, pol := range backendServer.PolicyNames {
					policyStrings = append(policyStrings, *pol)
				}
				if policyInBackendServerPolicies(policyName, policyStrings) {
					return fmt.Errorf("Policy still exists and is assigned")
				}
			}
		default:
			continue
		}
	}
	return nil
}

func testAccCheckAWSLoadBalancerBackendServerPolicyState(loadBalancerName string, loadBalancerBackendAuthPolicyName string, assigned bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		elbconn := testAccProvider.Meta().(*AWSClient).elbconn

		loadBalancerDescription, err := elbconn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{aws.String(loadBalancerName)},
		})
		if err != nil {
			return err
		}

		for _, backendServer := range loadBalancerDescription.LoadBalancerDescriptions[0].BackendServerDescriptions {
			policyStrings := []string{}
			for _, pol := range backendServer.PolicyNames {
				policyStrings = append(policyStrings, *pol)
			}
			if policyInBackendServerPolicies(loadBalancerBackendAuthPolicyName, policyStrings) != assigned {
				if assigned {
					return fmt.Errorf("Policy no longer assigned %s not in %+v", loadBalancerBackendAuthPolicyName, policyStrings)
				} else {
					return fmt.Errorf("Policy exists and is assigned")
				}
			}
		}

		return nil
	}
}

func testAccAWSLoadBalancerBackendServerPolicyConfig_basic0(rName, privateKey, publicKey, certificate string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_iam_server_certificate" "test-iam-cert0" {
  name_prefix      = "test_cert_"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_elb" "test-lb" {
  name               = "%[1]s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port      = 443
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test-iam-cert0.arn
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test-pubkey-policy0" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-pubkey-policy0"
  policy_type_name   = "PublicKeyPolicyType"

  policy_attribute {
    name  = "PublicKey"
    value = replace(replace(replace("%[4]s", "\n", ""), "-----BEGIN PUBLIC KEY-----", ""), "-----END PUBLIC KEY-----", "")
  }
}

resource "aws_load_balancer_policy" "test-backend-auth-policy0" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-backend-auth-policy0"
  policy_type_name   = "BackendServerAuthenticationPolicyType"

  policy_attribute {
    name  = "PublicKeyPolicyName"
    value = aws_load_balancer_policy.test-pubkey-policy0.policy_name
  }
}

resource "aws_load_balancer_backend_server_policy" "test-backend-auth-policies-443" {
  load_balancer_name = aws_elb.test-lb.name
  instance_port      = 443

  policy_names = [
    aws_load_balancer_policy.test-backend-auth-policy0.policy_name,
  ]
}
`, rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(privateKey), tlsPemEscapeNewlines(publicKey))
}

func testAccAWSLoadBalancerBackendServerPolicyConfig_basic1(rName, privateKey1, publicKey1, certificate1, publicKey2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_iam_server_certificate" "test-iam-cert0" {
  name_prefix      = "test_cert_"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_elb" "test-lb" {
  name               = "%[1]s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port      = 443
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test-iam-cert0.arn
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test-pubkey-policy0" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-pubkey-policy0"
  policy_type_name   = "PublicKeyPolicyType"

  policy_attribute {
    name  = "PublicKey"
    value = replace(replace(replace("%[4]s", "\n", ""), "-----BEGIN PUBLIC KEY-----", ""), "-----END PUBLIC KEY-----", "")
  }
}

resource "aws_load_balancer_policy" "test-pubkey-policy1" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-pubkey-policy1"
  policy_type_name   = "PublicKeyPolicyType"

  policy_attribute {
    name  = "PublicKey"
    value = replace(replace(replace("%[5]s", "\n", ""), "-----BEGIN PUBLIC KEY-----", ""), "-----END PUBLIC KEY-----", "")
  }
}

resource "aws_load_balancer_policy" "test-backend-auth-policy0" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-backend-auth-policy0"
  policy_type_name   = "BackendServerAuthenticationPolicyType"

  policy_attribute {
    name  = "PublicKeyPolicyName"
    value = aws_load_balancer_policy.test-pubkey-policy1.policy_name
  }
}

resource "aws_load_balancer_backend_server_policy" "test-backend-auth-policies-443" {
  load_balancer_name = aws_elb.test-lb.name
  instance_port      = 443

  policy_names = [
    aws_load_balancer_policy.test-backend-auth-policy0.policy_name,
  ]
}
`, rName, tlsPemEscapeNewlines(certificate1), tlsPemEscapeNewlines(privateKey1), tlsPemEscapeNewlines(publicKey1), tlsPemEscapeNewlines(publicKey2))
}

func testAccAWSLoadBalancerBackendServerPolicyConfig_basic2(rName, privateKey, certificate string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_iam_server_certificate" "test-iam-cert0" {
  name_prefix      = "test_cert_"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_elb" "test-lb" {
  name               = "%[1]s"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port      = 443
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test-iam-cert0.arn
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`, rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(privateKey))
}
