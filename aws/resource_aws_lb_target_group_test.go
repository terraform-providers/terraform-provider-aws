package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).elbv2conn

	prefixes := []string{
		"tf-",
		"tf-test-",
		"tf-acc-test-",
		"test",
	}

	err = conn.DescribeTargetGroupsPages(&elbv2.DescribeTargetGroupsInput{}, func(page *elbv2.DescribeTargetGroupsOutput, isLast bool) bool {
		if page == nil || len(page.TargetGroups) == 0 {
			log.Print("[DEBUG] No LB Target Groups to sweep")
			return false
		}

		for _, targetGroup := range page.TargetGroups {
			name := aws.StringValue(targetGroup.TargetGroupName)
			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(name, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping LB Target Group: %s", name)
				continue
			}
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
		return fmt.Errorf("Error retrieving LB Target Groups: %s", err)
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
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup/my-targets/73e2d6bc24d8a067`),
			suffix: `targetgroup/my-targets/73e2d6bc24d8a067`,
		},
		{
			name:   "no suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup`),
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
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "slow_start", "0"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.TestName", "TestAccAWSLBTargetGroup_basic"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_networkLB_TargetGroup(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &confBefore),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "TCP"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "proxy_protocol_v2", "false"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&confBefore, 10),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "traffic-port"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&confBefore, 10),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&confBefore, 3),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&confBefore, 3),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Name", "TestAcc_networkLB_TargetGroup"),
				),
			},
			{
				Config:      testAccAWSLBTargetGroupConfig_typeTCPInvalidThreshold(targetGroupName),
				ExpectError: regexp.MustCompile(`health_check\.healthy_threshold [0-9]+ and health_check\.unhealthy_threshold [0-9]+ must be the same for target_groups with TCP protocol`),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCPThresholdUpdated(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &confAfter),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "TCP"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&confAfter, 10),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "traffic-port"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&confAfter, 10),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "5"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&confAfter, 5),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "5"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&confAfter, 5),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Name", "TestAcc_networkLB_TargetGroup"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCPIntervalUpdated(targetGroupName),

				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("Health check interval cannot be updated"),
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

func TestAccAWSLBTargetGroup_networkLB_TargetGroupWithProxy(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &confBefore),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "proxy_protocol_v2", "false"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_withProxyProtocol(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &confAfter),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "proxy_protocol_v2", "true"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_TCP_HTTPHealthCheck(t *testing.T) {
	var confBefore, confAfter elbv2.TargetGroup
	rString := acctest.RandString(8)
	targetGroupName := fmt.Sprintf("test-tg-tcp-http-hc-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(targetGroupName, "/healthz", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &confBefore),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "TCP"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "300"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "30"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&confBefore, 30),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/healthz"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&confBefore, 10),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "2"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&confBefore, 2),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "2"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&confBefore, 2),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Name", "TestAcc_networkLB_HTTPHealthCheck"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(targetGroupName, "/healthz2", 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &confAfter),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "TCP"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "300"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "30"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&confAfter, 30),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/healthz2"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "10"),
					testAccCheckAWSLBTargetGroupHealthCheckTimeout(&confAfter, 10),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&confAfter, 4),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&confAfter, 4),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Name", "TestAcc_networkLB_HTTPHealthCheck"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroupBackwardsCompatibility(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_alb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfigBackwardsCompatibility(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_alb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_alb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_alb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "slow_start", "0"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.interval", "60"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_alb_target_group.test", "tags.TestName", "TestAccAWSLBTargetGroup_basic"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_namePrefix(t *testing.T) {
	var conf elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestMatchResourceAttr("aws_lb_target_group.test", "name", regexp.MustCompile("^tf-")),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_generatedName(t *testing.T) {
	var conf elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changeNameForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	targetGroupNameBefore := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	targetGroupNameAfter := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(4, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupNameBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &before),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupNameBefore),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupNameAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &after),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupNameAfter),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changeProtocolForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &before),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updatedProtocol(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &after),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changePortForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &before),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updatedPort(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &after),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "442"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_changeVpcForceNew(t *testing.T) {
	var before, after elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &before),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updatedVpc(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &after),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_tags(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.TestName", "TestAccAWSLBTargetGroup_basic"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updateTags(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Environment", "Production"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Type", "ALB Target Group"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_updateHealthCheck(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "60"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&conf, 60),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&conf, 3),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "3"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&conf, 3),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.matcher", "200-299"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_updateHealthCheck(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "30"),
					testAccCheckAWSLBTargetGroupHealthCheckInterval(&conf, 30),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupHealthyThreshold(&conf, 4),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "4"),
					testAccCheckAWSLBTargetGroupUnhealthyThreshold(&conf, 4),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_updateSticknessEnabled(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickiness(targetGroupName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickiness(targetGroupName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.matcher", "200"),
				),
			},
			{
				Config: testAccAWSLBTargetGroupConfig_stickiness(targetGroupName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.cookie_duration", "10000"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.path", "/health2"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8082"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "4"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.matcher", "200"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_defaults_application(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALB_defaults(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "slow_start", "0"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "5"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Name", "TestAccAWSLBTargetGroup_application_LB_defaults"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_defaults_network(t *testing.T) {
	var conf elbv2.TargetGroup
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	healthCheckInvalid1 := `
    path = "/health"
    interval = 10
    port     = 8081
    protocol = "TCP"
    `
	healthCheckInvalid2 := `
    interval = 10
    port     = 8081
    protocol = "TCP"
    matcher = "200"
    `
	healthCheckInvalid3 := `
    interval = 10
    port     = 8081
    protocol = "TCP"
    timeout = 4
    `
	healthCheckValid := `
    interval = 10
    port     = 8081
    protocol = "TCP"
    `

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNLB_defaults(targetGroupName, healthCheckInvalid1),
				ExpectError: regexp.MustCompile("health_check.path is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccNLB_defaults(targetGroupName, healthCheckInvalid2),
				ExpectError: regexp.MustCompile("health_check.matcher is not supported for target_groups with TCP protocol"),
			},
			{
				Config:      testAccNLB_defaults(targetGroupName, healthCheckInvalid3),
				ExpectError: regexp.MustCompile("health_check.timeout is not supported for target_groups with TCP protocol"),
			},
			{
				Config: testAccNLB_defaults(targetGroupName, healthCheckValid),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "name", targetGroupName),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "protocol", "TCP"),
					resource.TestCheckResourceAttrSet("aws_lb_target_group.test", "vpc_id"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "deregistration_delay", "200"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "slow_start", "0"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.interval", "10"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.timeout", "10"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "tags.Name", "TestAccAWSLBTargetGroup_application_LB_defaults"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessWithTCPDisabled(t *testing.T) {
	var conf elbv2.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupConfig_stickinessWithTCP(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &conf),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.type", "lb_cookie"),
					resource.TestCheckResourceAttr("aws_lb_target_group.test", "stickiness.0.cookie_duration", "86400"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroup_stickinessWithTCPEnabledShouldError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBTargetGroupConfig_stickinessWithTCP(true),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Network Load Balancers do not support Stickiness"),
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
			return fmt.Errorf("Unexpected error checking ALB destroyed: %s", err)
		}
	}

	return nil
}

func testAccALB_defaults(name string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 443
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.test.id}"

  deregistration_delay = 200
  slow_start = 0

  # HTTP Only
  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    interval = 10
    port     = 8081
    protocol = "HTTP"
    healthy_threshold = 3
    unhealthy_threshold = 3
  }
  tags = {
    Name = "TestAccAWSLBTargetGroup_application_LB_defaults"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-alb-defaults"
  }
}`, name)
}

func testAccNLB_defaults(name, healthCheckBlock string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 443
  protocol = "TCP"
  vpc_id   = "${aws_vpc.test.id}"

  deregistration_delay = 200
  slow_start = 0

  health_check {
                %s
  }

  tags = {
    Name = "TestAccAWSLBTargetGroup_application_LB_defaults"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-nlb-defaults"
  }
}`, name, healthCheckBlock)
}

func testAccAWSLBTargetGroupConfig_basic(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200
  slow_start = 0

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags = {
    TestName = "TestAccAWSLBTargetGroup_basic"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-basic"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfigBackwardsCompatibility(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_alb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200
  slow_start = 0

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags = {
    TestName = "TestAccAWSLBTargetGroup_basic"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-bc"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_updatedPort(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 442
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags = {
    TestName = "TestAccAWSLBTargetGroup_basic"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-basic"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_updatedProtocol(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTP"
  vpc_id = "${aws_vpc.test2.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags = {
    TestName = "TestAccAWSLBTargetGroup_basic"
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.10.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-basic-2"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-basic"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_updatedVpc(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags = {
    TestName = "TestAccAWSLBTargetGroup_basic"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-updated-vpc"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_updateTags(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags = {
    Environment = "Production"
    Type = "ALB Target Group"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-update-tags"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_updateHealthCheck(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health2"
    interval = 30
    port = 8082
    protocol = "HTTPS"
    timeout = 4
    healthy_threshold = 4
    unhealthy_threshold = 4
    matcher = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-update-health-check"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_Protocol_Tls(targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-lb-target-group-protocol-tls"
  }
}

resource "aws_lb_target_group" "test" {
  name     = %q
  port     = 443
  protocol = "TLS"
  vpc_id   = "${aws_vpc.test.id}"

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = "tf-acc-test-lb-target-group-protocol-tls"
  }
}
`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_typeTCP(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8082
  protocol = "TCP"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  health_check {
    interval = 10
    port = "traffic-port"
    protocol = "TCP"
    healthy_threshold = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = "TestAcc_networkLB_TargetGroup"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-type-tcp"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_typeTCP_withProxyProtocol(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8082
  protocol = "TCP"
  vpc_id = "${aws_vpc.test.id}"
	
	proxy_protocol_v2 = "true"    
	deregistration_delay = 200

  health_check {
    interval = 10
    port = "traffic-port"
    protocol = "TCP"
    healthy_threshold = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = "TestAcc_networkLB_TargetGroup"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-type-tcp"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_typeTCPInvalidThreshold(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8082
  protocol = "TCP"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  health_check {
    interval = 10
    port = "traffic-port"
    protocol = "TCP"
    healthy_threshold = 3
    unhealthy_threshold = 4
  }

  tags = {
    Name = "TestAcc_networkLB_TargetGroup"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-type-tcp"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_typeTCPThresholdUpdated(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8082
  protocol = "TCP"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  health_check {
    interval = 10
    port = "traffic-port"
    protocol = "TCP"
    healthy_threshold = 5
    unhealthy_threshold = 5
  }

  tags = {
    Name = "TestAcc_networkLB_TargetGroup"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-type-tcp-threshold-updated"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_typeTCPIntervalUpdated(targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8082
  protocol = "TCP"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  health_check {
    interval = 30
    port = "traffic-port"
    protocol = "TCP"
    healthy_threshold = 5
    unhealthy_threshold = 5
  }

  tags = {
    Name = "TestAcc_networkLB_TargetGroup"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-target-group-type-tcp-interval-updated"
  }
}`, targetGroupName)
}

func testAccAWSLBTargetGroupConfig_typeTCP_HTTPHealthCheck(targetGroupName, path string, threshold int) string {
	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name     = "%[1]s"
  port     = 8082
  protocol = "TCP"
  vpc_id   = "${aws_vpc.test.id}"

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
    Name = "TestAcc_networkLB_HTTPHealthCheck"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-lb-target-group-type-tcp-http-health-check"
  }
}`, targetGroupName, threshold, path)
}

func testAccAWSLBTargetGroupConfig_stickiness(targetGroupName string, addStickinessBlock bool, enabled bool) string {
	var stickinessBlock string

	if addStickinessBlock {
		stickinessBlock = fmt.Sprintf(`stickiness {
	    enabled = "%t"
	    type = "lb_cookie"
	    cookie_duration = 10000
	  }`, enabled)
	}

	return fmt.Sprintf(`resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  %s

  health_check {
    path = "/health2"
    interval = 30
    port = 8082
    protocol = "HTTPS"
    timeout = 4
    healthy_threshold = 4
    unhealthy_threshold = 4
    matcher = "200"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    TestName = "terraform-testacc-lb-target-group-stickiness"
  }
}`, targetGroupName, stickinessBlock)
}

const testAccAWSLBTargetGroupConfig_namePrefix = `
resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port = 80
  protocol = "HTTP"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-lb-target-group-name-prefix"
	}
}
`

const testAccAWSLBTargetGroupConfig_generatedName = `
resource "aws_lb_target_group" "test" {
  port = 80
  protocol = "HTTP"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-lb-target-group-generated-name"
	}
}
`

func testAccAWSLBTargetGroupConfig_stickinessWithTCP(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name_prefix = "tf-"
  port        = 25
  protocol    = "TCP"
  vpc_id      = "${aws_vpc.test.id}"

  stickiness {
    type    = "lb_cookie"
    enabled = %t
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "testAccAWSLBTargetGroupConfig_namePrefix"
  }
}
`, enabled)
}
