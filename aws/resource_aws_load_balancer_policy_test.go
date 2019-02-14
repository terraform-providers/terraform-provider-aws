package aws

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLoadBalancerPolicy_basic(t *testing.T) {
	var policy elb.PolicyDescription
	loadBalancerResourceName := "aws_elb.test-lb"
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLoadBalancerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLoadBalancerPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLoadBalancerPolicyExists(resourceName, &policy),
					testAccCheckAWSLoadBalancerPolicyState(loadBalancerResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccAWSLoadBalancerPolicy_disappears(t *testing.T) {
	var loadBalancer elb.LoadBalancerDescription
	var policy elb.PolicyDescription
	loadBalancerResourceName := "aws_elb.test-lb"
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLoadBalancerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLoadBalancerPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(loadBalancerResourceName, &loadBalancer),
					testAccCheckAWSLoadBalancerPolicyExists(resourceName, &policy),
					testAccCheckAWSLoadBalancerPolicyDisappears(&loadBalancer, &policy),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSLoadBalancerPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(loadBalancerResourceName, &loadBalancer),
					testAccCheckAWSLoadBalancerPolicyExists(resourceName, &policy),
					testAccCheckAWSELBDisappears(&loadBalancer),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLoadBalancerPolicy_updateWhileAssigned(t *testing.T) {
	var policy elb.PolicyDescription
	loadBalancerResourceName := "aws_elb.test-lb"
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLoadBalancerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLoadBalancerPolicyConfig_updateWhileAssigned0(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLoadBalancerPolicyExists(resourceName, &policy),
					testAccCheckAWSLoadBalancerPolicyState(loadBalancerResourceName, resourceName),
				),
			},
			{
				Config: testAccAWSLoadBalancerPolicyConfig_updateWhileAssigned1(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLoadBalancerPolicyExists(resourceName, &policy),
					testAccCheckAWSLoadBalancerPolicyState(loadBalancerResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckAWSLoadBalancerPolicyExists(resourceName string, policyDescription *elb.PolicyDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer Policy ID is set for %s", resourceName)
		}

		loadBalancerName, policyName := resourceAwsLoadBalancerPolicyParseId(rs.Primary.ID)

		conn := testAccProvider.Meta().(*AWSClient).elbconn

		input := &elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(loadBalancerName),
			PolicyNames:      []*string{aws.String(policyName)},
		}

		output, err := conn.DescribeLoadBalancerPolicies(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.PolicyDescriptions) == 0 {
			return fmt.Errorf("Load Balancer Policy (%s) not found", rs.Primary.ID)
		}

		*policyDescription = *output.PolicyDescriptions[0]

		return nil
	}
}

func testAccCheckAWSLoadBalancerPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_load_balancer_policy" {
			continue
		}

		loadBalancerName, policyName := resourceAwsLoadBalancerPolicyParseId(rs.Primary.ID)
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
	}
	return nil
}

func testAccCheckAWSLoadBalancerPolicyDisappears(loadBalancer *elb.LoadBalancerDescription, policy *elb.PolicyDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elbconn

		input := elb.DeleteLoadBalancerPolicyInput{
			LoadBalancerName: loadBalancer.LoadBalancerName,
			PolicyName:       policy.PolicyName,
		}
		_, err := conn.DeleteLoadBalancerPolicy(&input)

		return err
	}
}

func testAccCheckAWSLoadBalancerPolicyState(elbResource string, policyResource string) resource.TestCheckFunc {
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
		loadBalancerName, policyName := resourceAwsLoadBalancerPolicyParseId(policy.Primary.ID)
		loadBalancerPolicies, err := elbconn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(loadBalancerName),
			PolicyNames:      []*string{aws.String(policyName)},
		})

		if err != nil {
			return err
		}

		for _, loadBalancerPolicy := range loadBalancerPolicies.PolicyDescriptions {
			if *loadBalancerPolicy.PolicyName == policyName {
				if *loadBalancerPolicy.PolicyTypeName != policy.Primary.Attributes["policy_type_name"] {
					return fmt.Errorf("PolicyTypeName does not match")
				}
				policyAttributeCount, err := strconv.Atoi(policy.Primary.Attributes["policy_attribute.#"])
				if err != nil {
					return err
				}
				if len(loadBalancerPolicy.PolicyAttributeDescriptions) != policyAttributeCount {
					return fmt.Errorf("PolicyAttributeDescriptions length mismatch")
				}
				policyAttributes := make(map[string]string)
				for k, v := range policy.Primary.Attributes {
					if strings.HasPrefix(k, "policy_attribute.") && strings.HasSuffix(k, ".name") {
						key := v
						value_key := fmt.Sprintf("%s.value", strings.TrimSuffix(k, ".name"))
						policyAttributes[key] = policy.Primary.Attributes[value_key]
					}
				}
				for _, policyAttribute := range loadBalancerPolicy.PolicyAttributeDescriptions {
					if *policyAttribute.AttributeValue != policyAttributes[*policyAttribute.AttributeName] {
						return fmt.Errorf("PollicyAttribute Value mismatch %s != %s: %s", *policyAttribute.AttributeValue, policyAttributes[*policyAttribute.AttributeName], policyAttributes)
					}
				}
			}
		}

		return nil
	}
}

func testAccAWSLoadBalancerPolicyConfig_basic(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_elb" "test-lb" {
		name = "test-lb-%d"
		availability_zones = ["us-west-2a"]

		listener {
			instance_port = 80
			instance_protocol = "http"
			lb_port = 80
			lb_protocol = "http"
		}

	tags = {
			Name = "tf-acc-test"
		}
	}

	resource "aws_load_balancer_policy" "test-policy" {
		load_balancer_name = "${aws_elb.test-lb.name}"
		policy_name = "test-policy-%d"
		policy_type_name = "AppCookieStickinessPolicyType"
		policy_attribute {
			name = "CookieName"
			value = "magic_cookie"
		}
	}`, rInt, rInt)
}

func testAccAWSLoadBalancerPolicyConfig_updateWhileAssigned0(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_elb" "test-lb" {
		name = "test-lb-%d"
		availability_zones = ["us-west-2a"]

		listener {
			instance_port = 80
			instance_protocol = "http"
			lb_port = 80
			lb_protocol = "http"
		}

	tags = {
			Name = "tf-acc-test"
		}
	}

	resource "aws_load_balancer_policy" "test-policy" {
		load_balancer_name = "${aws_elb.test-lb.name}"
		policy_name = "test-policy-%d"
		policy_type_name = "AppCookieStickinessPolicyType"
		policy_attribute {
			name = "CookieName"
			value = "magic_cookie"
		}
	}

	resource "aws_load_balancer_listener_policy" "test-lb-test-policy-80" {
		load_balancer_name = "${aws_elb.test-lb.name}"
		load_balancer_port = 80
		policy_names = [
			"${aws_load_balancer_policy.test-policy.policy_name}"
		]
	}`, rInt, rInt)
}

func testAccAWSLoadBalancerPolicyConfig_updateWhileAssigned1(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_elb" "test-lb" {
		name = "test-lb-%d"
		availability_zones = ["us-west-2a"]

		listener {
			instance_port = 80
			instance_protocol = "http"
			lb_port = 80
			lb_protocol = "http"
		}

	tags = {
			Name = "tf-acc-test"
		}
	}

	resource "aws_load_balancer_policy" "test-policy" {
		load_balancer_name = "${aws_elb.test-lb.name}"
		policy_name = "test-policy-%d"
		policy_type_name = "AppCookieStickinessPolicyType"
		policy_attribute {
			name = "CookieName"
			value = "unicorn_cookie"
		}
	}

	resource "aws_load_balancer_listener_policy" "test-lb-test-policy-80" {
		load_balancer_name = "${aws_elb.test-lb.name}"
		load_balancer_port = 80
		policy_names = [
			"${aws_load_balancer_policy.test-policy.policy_name}"
		]
	}`, rInt, rInt)
}
