package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lb_target_group", &resource.Sweeper{
		Name: "aws_lb_target_group",
		F:    testSweepLBTargetGroups,
		Dependencies: []string{
			"aws_lb",
		},
	})
}

func testSweepLBTargetGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).elbv2conn

	err = conn.DescribeTargetGroupsPages(&elbv2.DescribeTargetGroupsInput{}, func(page *elbv2.DescribeTargetGroupsOutput, isLast bool) bool {
		if page == nil || len(page.TargetGroups) == 0 {
			log.Print("[DEBUG] No LB Target Groups to sweep")
			return false
		}

		for _, targetGroup := range page.TargetGroups {
			name := aws.StringValue(targetGroup.TargetGroupName)

			log.Printf("[INFO] Deleting LB Target Group: %s", name)
			_, err := conn.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
				TargetGroupArn: targetGroup.TargetGroupArn,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete LB Target Group (%s): %s", name, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping LB Target Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving LB Target Groups: %w", err)
	}
	return nil
}

func TestLBTargetGroupCloudwatchSuffixFromARN(t *testing.T) {
	cases := []struct {
		name   string
		arn    *string
		suffix string
	}{
		{
			name:   "valid suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup/my-targets/73e2d6bc24d8a067`), //lintignore:AWSAT003,AWSAT005
			suffix: `targetgroup/my-targets/73e2d6bc24d8a067`,
		},
		{
			name:   "no suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup`), //lintignore:AWSAT003,AWSAT005
			suffix: ``,
		},
		{
			name:   "nil ARN",
			arn:    nil,
			suffix: ``,
		},
	}

	for _, tc := range cases {
		actual := lbTargetGroupSuffixFromARN(tc.arn)
		if actual != tc.suffix {
			t.Fatalf("bad suffix: %q\nExpected: %s\n     Got: %s", tc.name, tc.suffix, actual)
		}
	}
}

func TestAccAWSLBTargetGroup_basic(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_basicUdp(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basicUdp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "514"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "UDP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "514"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_ProtocolVersion(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_ProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP2"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_withoutHealthcheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_withoutHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_networkLB_TargetGroup(t *testing.T) {
	var targetGroup1, targetGroup2, targetGroup3 elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "false"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&targetGroup1, 10),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&targetGroup1, 10),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&targetGroup1, 3),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&targetGroup1, 3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config:      testAccAWSLBTargetGroupConfig_typeTCPInvalidThreshold(rName),
				ExpectError: regexp.MustCompile(`health_check\.healthy_threshold [0-9]+ and health_check\.unhealthy_threshold [0-9]+ must be the same for target_groups with TCP protocol`),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCPThresholdUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup2),
					testAccCheckAWSLBTargetGroupNotRecreated(&targetGroup1, &targetGroup2),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&targetGroup2, 10),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "traffic-port"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&targetGroup2, 10),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "5"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&targetGroup2, 5),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "5"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&targetGroup2, 5),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCPIntervalUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup3),
					testAccCheckAWSLBTargetGroupRecreated(&targetGroup2, &targetGroup3),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_Protocol_Geneve(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:        testAccErrorCheck(t, elbv2.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfigProtocolGeneve(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "port", "6081"),
					resource.TestCheckResourceAttr(resourceName, "protocol", elbv2.ProtocolEnumGeneve),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"lambda_multi_value_headers_enabled",
					"proxy_protocol_v2",
					"slow_start",
				},
			},
		},
	})
}

func TestAccAWSLBTargetGroup_Protocol_Tcp_HealthCheck_Protocol(t *testing.T) {
	var targetGroup1, targetGroup2 elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCPIntervalUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, "/", 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup2),
					testAccCheckAWSLBTargetGroupRecreated(&targetGroup1, &targetGroup2),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_Protocol_Tls(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_Protocol_Tls(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_ProtocolVersion_GRPC_HealthCheck(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_GRPC_ProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/Test.Check/healthcheck"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "0-99"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_ProtocolVersion_HTTP_GRPC_Update(t *testing.T) {
	var targetGroup1 elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_GRPC_ProtocolVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "GRPC"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_networkLB_TargetGroupWithProxy(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &confBefore),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "false"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_withProxyProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &confAfter),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol_v2", "true"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_TCP_HTTPHealthCheck(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, "/healthz", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &confBefore),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&confBefore, 30),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/healthz"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&confBefore, 10),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "2"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&confBefore, 2),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "2"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&confBefore, 2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, "/healthz2", 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &confAfter),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&confAfter, 30),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/healthz2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&confAfter, 10),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&confAfter, 4),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&confAfter, 4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_BackwardsCompatibility(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfigBackwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_namePrefix(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_generatedName(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_generatedName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changeNameForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rNameBefore := acctest.RandomWithPrefix("tf-acc-test")
	rNameAfter := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rNameBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rNameBefore),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rNameAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameAfter),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changeProtocolForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updatedProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changePortForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updatedPort(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "port", "442"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changeVpcForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &before),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updatedVpc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &after),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_tags(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_enableHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_withoutHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "false"),
					testAccCheckAWSLBTargetGroupHealthCheckEnabled(&conf, false),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_enableHealthcheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					testAccCheckAWSLBTargetGroupHealthCheckEnabled(&conf, true),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_updateHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "60"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&conf, 60),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&conf, 3),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&conf, 3),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updateHealthCheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.enabled", "true"),
					testAccCheckAWSLBTargetGroupHealthCheckEnabled(&conf, true),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&conf, 30),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&conf, 4),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&conf, 4),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_updateSticknessEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickiness(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickiness(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickiness(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_defaults_application(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALB_defaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_defaults_network(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"
	healthCheckInvalid1 := `
path     = "/health"
interval = 10
port     = 8081
protocol = "TCP"
    `
	healthCheckInvalid2 := `
interval = 10
port     = 8081
protocol = "TCP"
matcher  = "200"
    `
	healthCheckInvalid3 := `
interval = 10
port     = 8081
protocol = "TCP"
timeout  = 4
    `
	healthCheckValid := `
interval = 10
port     = 8081
protocol = "TCP"
    `

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNLB_defaults(rName, healthCheckInvalid1),
				ExpectError: regexp.MustCompile("health_check.path is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccNLB_defaults(rName, healthCheckInvalid2),
				ExpectError: regexp.MustCompile("health_check.matcher is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccNLB_defaults(rName, healthCheckInvalid3),
				ExpectError: regexp.MustCompile("health_check.timeout is not supported for target_groups with TCP protocol"),
			},
			{
				Config: testAccNLB_defaults(rName, healthCheckValid),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "port", "443"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "200"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessDefaultNLB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessDefault(rName, "TCP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessDefault(rName, "UDP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessDefault(rName, "TCP_UDP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessDefaultALB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessDefault(rName, "HTTP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessValidNLB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP", "source_ip", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				// this test should be invalid but allowed to avoid breaking changes
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "false"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "UDP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "source_ip", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "source_ip"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessValidALB(t *testing.T) {
	var conf elbv2.TargetGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheck(t, elbv2.EndpointsID),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "HTTP", "lb_cookie", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "86400"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "HTTPS", "lb_cookie", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.0.cookie_duration", "86400"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessInvalidNLB(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "UDP", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessInvalidALB(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "HTTP", "source_ip", true),
				ExpectError: regexp.MustCompile("Stickiness type 'source_ip' is not supported for target groups with"),
			},
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "HTTPS", "source_ip", true),
				ExpectError: regexp.MustCompile("Stickiness type 'source_ip' is not supported for target groups with"),
			},
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TLS", "lb_cookie", true),
				ExpectError: regexp.MustCompile("Stickiness type 'lb_cookie' is not supported for target groups with"),
			},
			{
				Config:             testAccAWSLBTargetGroupConfig_stickinessValidity(rName, "TCP_UDP", "lb_cookie", false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLBTargetGroup_preserveClientIPValid(t *testing.T) {
	var conf elbv2.TargetGroup
	resourceName := "aws_lb_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_preserveClientIP(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "true"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_preserveClientIP(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSLBTargetGroupExists(n string, res *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Target Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		describe, err := conn.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.TargetGroups) != 1 ||
			*describe.TargetGroups[0].TargetGroupArn != rs.Primary.ID {
			return errors.New("Target Group not found")
		}

		*res = *describe.TargetGroups[0]
		return nil
	}
}

func testAccCheckAWSLBTargetGroupNotRecreated(i, j *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TargetGroupArn) != aws.StringValue(j.TargetGroupArn) {
			return fmt.Errorf("ELBv2 Target Group (%s) unexpectedly recreated (%s)", aws.StringValue(i.TargetGroupArn), aws.StringValue(j.TargetGroupArn))
		}

		return nil
	}
}

func testAccCheckAWSLBTargetGroupRecreated(i, j *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TargetGroupArn) == aws.StringValue(j.TargetGroupArn) {
			return fmt.Errorf("ELBv2 Target Group (%s) not recreated", aws.StringValue(i.TargetGroupArn))
		}

		return nil
	}
}

func testAccCheckAWSLBTargetGroupHealthCheckEnabled(res *elbv2.TargetGroup, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if res.HealthCheckEnabled == nil {
			return fmt.Errorf("Expected HealthCheckEnabled to be %t, given %#v",
				expected, res.HealthCheckEnabled)
		}
		if *res.HealthCheckEnabled != expected {
			return fmt.Errorf("Expected HealthCheckEnabled to be %t, given %t",
				expected, *res.HealthCheckEnabled)
		}
		return nil
	}
}

func testAccCheckAWSLBTargetGroupHealthCheckInterval(res *elbv2.TargetGroup, expected int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if res.HealthCheckIntervalSeconds == nil {
			return fmt.Errorf("Expected HealthCheckIntervalSeconds to be %d, given: %#v",
				expected, res.HealthCheckIntervalSeconds)
		}
		if *res.HealthCheckIntervalSeconds != expected {
			return fmt.Errorf("Expected HealthCheckIntervalSeconds to be %d, given: %d",
				expected, *res.HealthCheckIntervalSeconds)
		}
		return nil
	}
}

func testAccCheckAWSLBTargetGroupHealthCheckTimeout(res *elbv2.TargetGroup, expected int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if res.HealthCheckTimeoutSeconds == nil {
			return fmt.Errorf("Expected HealthCheckTimeoutSeconds to be %d, given: %#v",
				expected, res.HealthCheckTimeoutSeconds)
		}
		if *res.HealthCheckTimeoutSeconds != expected {
			return fmt.Errorf("Expected HealthCheckTimeoutSeconds to be %d, given: %d",
				expected, *res.HealthCheckTimeoutSeconds)
		}
		return nil
	}
}

func testAccCheckAWSLBTargetGroupHealthyThreshold(res *elbv2.TargetGroup, expected int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if res.HealthyThresholdCount == nil {
			return fmt.Errorf("Expected HealthyThresholdCount to be %d, given: %#v",
				expected, res.HealthyThresholdCount)
		}
		if *res.HealthyThresholdCount != expected {
			return fmt.Errorf("Expected HealthyThresholdCount to be %d, given: %d",
				expected, *res.HealthyThresholdCount)
		}
		return nil
	}
}

func testAccCheckAWSLBTargetGroupUnhealthyThreshold(res *elbv2.TargetGroup, expected int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if res.UnhealthyThresholdCount == nil {
			return fmt.Errorf("Expected.UnhealthyThresholdCount to be %d, given: %#v",
				expected, res.UnhealthyThresholdCount)
		}
		if *res.UnhealthyThresholdCount != expected {
			return fmt.Errorf("Expected.UnhealthyThresholdCount to be %d, given: %d",
				expected, *res.UnhealthyThresholdCount)
		}
		return nil
	}
}

func testAccCheckAWSLBTargetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_target_group" && rs.Type != "aws_alb_target_group" {
			continue
		}

		describe, err := conn.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.TargetGroups) != 0 &&
				*describe.TargetGroups[0].TargetGroupArn == rs.Primary.ID {
				return fmt.Errorf("Target Group %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
			return nil
		} else {
			return fmt.Errorf("unexpected error checking ALB destroyed: %w", err)
		}
	}

	return nil
}

func testAccALB_defaults(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  # HTTP Only

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "HTTP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccNLB_defaults(rName, healthCheckBlock string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  health_check {
    %[2]s
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, healthCheckBlock)
}

func testAccAWSLBTargetGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_ProtocolVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name             = %[1]q
  port             = 443
  protocol         = "HTTPS"
  protocol_version = "HTTP2"
  vpc_id           = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfigProtocolGeneve(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 80
    protocol = "HTTP"
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    %[2]q = %[3]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSLBTargetGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSLBTargetGroupConfig_basicUdp(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 514
  protocol = "UDP"
  vpc_id   = aws_vpc.test.id

  health_check {
    protocol = "TCP"
    port     = 514
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_withoutHealthcheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}
`, rName)
}

func testAccAWSLBTargetGroupConfigBackwardsCompatibility(rName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_enableHealthcheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"

  health_check {
    path     = "/health"
    interval = 60
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_updatedPort(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 442
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_updatedProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTP"
  vpc_id   = aws_vpc.test2.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_GRPC_ProtocolVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name             = %[1]q
  port             = 80
  protocol         = "HTTP"
  protocol_version = "GRPC"
  vpc_id           = aws_vpc.test2.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/Test.Check/healthcheck"
    interval            = 60
    port                = 8080
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "0-99"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_updatedVpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_updateHealthCheck(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_Protocol_Tls(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TLS"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_typeTCP(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_typeTCP_withProxyProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  proxy_protocol_v2    = "true"
  deregistration_delay = 200

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_typeTCPInvalidThreshold(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 4
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_typeTCPThresholdUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 5
    unhealthy_threshold = 5
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_typeTCPIntervalUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  health_check {
    interval            = 30
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 5
    unhealthy_threshold = 5
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(rName, path string, threshold int) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = "%[1]s"
  port     = 8082
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  health_check {
    healthy_threshold   = %[2]d
    unhealthy_threshold = %[2]d
    timeout             = "10"
    port                = "443"
    path                = "%[3]s"
    protocol            = "HTTPS"
    interval            = 30
    matcher             = "200-399"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, threshold, path)
}

func testAccAWSLBTargetGroupConfig_stickiness(rName string, addStickinessBlock bool, enabled bool) string {
	var stickinessBlock string

	if addStickinessBlock {
		stickinessBlock = fmt.Sprintf(`
stickiness {
  enabled         = "%[1]t"
  type            = "lb_cookie"
  cookie_duration = 10000
}
`, enabled)
	}

	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  %[2]s

  health_check {
    path                = "/health2"
    interval            = 30
    port                = 8082
    protocol            = "HTTPS"
    timeout             = 4
    healthy_threshold   = 4
    unhealthy_threshold = 4
    matcher             = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, stickinessBlock)
}

func testAccAWSLBTargetGroupConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_generatedName(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSLBTargetGroupConfig_stickinessDefault(rName, protocol string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 25
  protocol    = %[2]q
  vpc_id      = aws_vpc.test.id
}
`, rName, protocol)
}

func testAccAWSLBTargetGroupConfig_stickinessValidity(rName, protocol, stickyType string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 25
  protocol    = %[1]q
  vpc_id      = aws_vpc.test.id

  stickiness {
    type    = %[2]q
    enabled = %[3]t
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[4]q
  }
}
`, protocol, stickyType, enabled, rName)
}

func testAccAWSLBTargetGroupConfig_preserveClientIP(rName string, preserveClientIP bool) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  preserve_client_ip = %[2]t

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}`, rName, preserveClientIP)
}
