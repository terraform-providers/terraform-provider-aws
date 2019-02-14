package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLBSSLNegotiationPolicy_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckLBSSLNegotiationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSslNegotiationPolicyConfig(
					fmt.Sprintf("tf-acctest-%s", acctest.RandString(10)), fmt.Sprintf("tf-test-lb-%s", acctest.RandString(5))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLNegotiationPolicy(
						"aws_elb.lb",
						"aws_lb_ssl_negotiation_policy.foo",
					),
					resource.TestCheckResourceAttr(
						"aws_lb_ssl_negotiation_policy.foo", "attribute.#", "7"),
				),
			},
		},
	})
}

func TestAccAWSLBSSLNegotiationPolicy_missingLB(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", acctest.RandString(5))

	// check that we can destroy the policy if the LB is missing
	removeLB := func() {
		conn := testAccProvider.Meta().(*AWSClient).elbconn
		deleteElbOpts := elb.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(lbName),
		}
		if _, err := conn.DeleteLoadBalancer(&deleteElbOpts); err != nil {
			t.Fatalf("Error deleting ELB: %s", err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckLBSSLNegotiationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSslNegotiationPolicyConfig(fmt.Sprintf("tf-acctest-%s", acctest.RandString(10)), lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLNegotiationPolicy(
						"aws_elb.lb",
						"aws_lb_ssl_negotiation_policy.foo",
					),
					resource.TestCheckResourceAttr(
						"aws_lb_ssl_negotiation_policy.foo", "attribute.#", "7"),
				),
			},
			{
				PreConfig: removeLB,
				Config:    testAccSslNegotiationPolicyConfig(fmt.Sprintf("tf-acctest-%s", acctest.RandString(10)), lbName),
			},
		},
	})
}

func testAccCheckLBSSLNegotiationPolicyDestroy(s *terraform.State) error {
	elbconn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elb" && rs.Type != "aws_lb_ssl_negotiation_policy" {
			continue
		}

		// Check that the ELB is destroyed
		if rs.Type == "aws_elb" {
			describe, err := elbconn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
				LoadBalancerNames: []*string{aws.String(rs.Primary.ID)},
			})

			if err == nil {
				if len(describe.LoadBalancerDescriptions) != 0 &&
					*describe.LoadBalancerDescriptions[0].LoadBalancerName == rs.Primary.ID {
					return fmt.Errorf("ELB still exists")
				}
			}

			// Verify the error
			providerErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}

			if providerErr.Code() != "LoadBalancerNotFound" {
				return fmt.Errorf("Unexpected error: %s", err)
			}
		} else {
			// Check that the SSL Negotiation Policy is destroyed
			elbName, _, policyName := resourceAwsLBSSLNegotiationPolicyParseId(rs.Primary.ID)
			_, err := elbconn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
				LoadBalancerName: aws.String(elbName),
				PolicyNames:      []*string{aws.String(policyName)},
			})

			if err == nil {
				return fmt.Errorf("ELB SSL Negotiation Policy still exists")
			}
		}
	}

	return nil
}

func testAccCheckLBSSLNegotiationPolicy(elbResource string, policyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[elbResource]
		if !ok {
			return fmt.Errorf("Not found: %s", elbResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[policyResource]
		if !ok {
			return fmt.Errorf("Not found: %s", policyResource)
		}

		elbconn := testAccProvider.Meta().(*AWSClient).elbconn

		elbName, _, policyName := resourceAwsLBSSLNegotiationPolicyParseId(policy.Primary.ID)
		resp, err := elbconn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(elbName),
			PolicyNames:      []*string{aws.String(policyName)},
		})

		if err != nil {
			fmt.Printf("[ERROR] Problem describing load balancer policy '%s': %s", policyName, err)
			return err
		}

		if len(resp.PolicyDescriptions) != 1 {
			return fmt.Errorf("Unable to find policy %#v", resp.PolicyDescriptions)
		}

		attrmap := policyAttributesToMap(&resp.PolicyDescriptions[0].PolicyAttributeDescriptions)
		if attrmap["Protocol-TLSv1"] != "false" {
			return fmt.Errorf("Policy attribute 'Protocol-TLSv1' was of value %s instead of false!", attrmap["Protocol-TLSv1"])
		}
		if attrmap["Protocol-TLSv1.1"] != "false" {
			return fmt.Errorf("Policy attribute 'Protocol-TLSv1.1' was of value %s instead of false!", attrmap["Protocol-TLSv1.1"])
		}
		if attrmap["Protocol-TLSv1.2"] != "true" {
			return fmt.Errorf("Policy attribute 'Protocol-TLSv1.2' was of value %s instead of true!", attrmap["Protocol-TLSv1.2"])
		}
		if attrmap["Server-Defined-Cipher-Order"] != "true" {
			return fmt.Errorf("Policy attribute 'Server-Defined-Cipher-Order' was of value %s instead of true!", attrmap["Server-Defined-Cipher-Order"])
		}
		if attrmap["ECDHE-RSA-AES128-GCM-SHA256"] != "true" {
			return fmt.Errorf("Policy attribute 'ECDHE-RSA-AES128-GCM-SHA256' was of value %s instead of true!", attrmap["ECDHE-RSA-AES128-GCM-SHA256"])
		}
		if attrmap["AES128-GCM-SHA256"] != "true" {
			return fmt.Errorf("Policy attribute 'AES128-GCM-SHA256' was of value %s instead of true!", attrmap["AES128-GCM-SHA256"])
		}
		if attrmap["EDH-RSA-DES-CBC3-SHA"] != "false" {
			return fmt.Errorf("Policy attribute 'EDH-RSA-DES-CBC3-SHA' was of value %s instead of false!", attrmap["EDH-RSA-DES-CBC3-SHA"])
		}

		return nil
	}
}

func policyAttributesToMap(attributes *[]*elb.PolicyAttributeDescription) map[string]string {
	attrmap := make(map[string]string)

	for _, attrdef := range *attributes {
		attrmap[*attrdef.AttributeName] = *attrdef.AttributeValue
	}

	return attrmap
}

// Sets the SSL Negotiation policy with attributes.
func testAccSslNegotiationPolicyConfig(certName string, lbName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_iam_server_certificate" "test_cert" {
  name             = "%s"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
  private_key      = "${tls_private_key.example.private_key_pem}"
}

resource "aws_elb" "lb" {
	name = "%s"
	availability_zones = ["us-west-2a"]
	listener {
		instance_port = 8000
		instance_protocol = "https"
		lb_port = 443
		lb_protocol = "https"
		ssl_certificate_id = "${aws_iam_server_certificate.test_cert.arn}"
	}
}
resource "aws_lb_ssl_negotiation_policy" "foo" {
	name = "foo-policy"
	load_balancer = "${aws_elb.lb.id}"
	lb_port = 443
	attribute {
    	name = "Protocol-TLSv1"
        value = "false"
    }
    attribute {
        name = "Protocol-TLSv1.1"
        value = "false"
    }
    attribute {
        name = "Protocol-TLSv1.2"
        value = "true"
    }
    attribute {
        name = "Server-Defined-Cipher-Order"
        value = "true"
    }
    attribute {
        name = "ECDHE-RSA-AES128-GCM-SHA256"
        value = "true"
    }
    attribute {
        name = "AES128-GCM-SHA256"
        value = "true"
    }
    attribute {
        name = "EDH-RSA-DES-CBC3-SHA"
        value = "false"
    }
}
`, certName, lbName)
}
