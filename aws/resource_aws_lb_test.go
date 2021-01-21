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
	resource.AddTestSweepers("aws_lb", &resource.Sweeper{
		Name: "aws_lb",
		F:    testSweepLBs,
		Dependencies: []string{
			"aws_api_gateway_vpc_link",
			"aws_vpc_endpoint_service",
		},
	})
}

func testSweepLBs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).elbv2conn

	err = conn.DescribeLoadBalancersPages(&elbv2.DescribeLoadBalancersInput{}, func(page *elbv2.DescribeLoadBalancersOutput, isLast bool) bool {
		if page == nil || len(page.LoadBalancers) == 0 {
			log.Print("[DEBUG] No LBs to sweep")
			return false
		}

		for _, loadBalancer := range page.LoadBalancers {
			name := aws.StringValue(loadBalancer.LoadBalancerName)

			log.Printf("[INFO] Deleting LB: %s", name)
			_, err := conn.DeleteLoadBalancer(&elbv2.DeleteLoadBalancerInput{
				LoadBalancerArn: loadBalancer.LoadBalancerArn,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete LB (%s): %s", name, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping LB sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving LBs: %s", err)
	}
	return nil
}

func TestLBCloudwatchSuffixFromARN(t *testing.T) {
	cases := []struct {
		name   string
		arn    *string
		suffix string
	}{
		{
			name:   "valid suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:loadbalancer/app/my-alb/abc123`), //lintignore:AWSAT003,AWSAT005
			suffix: `app/my-alb/abc123`,
		},
		{
			name:   "no suffix",
			arn:    aws.String(`arn:aws:elasticloadbalancing:us-east-1:123456:loadbalancer`), //lintignore:AWSAT003,AWSAT005
			suffix: ``,
		},
		{
			name:   "nil ARN",
			arn:    nil,
			suffix: ``,
		},
	}

	for _, tc := range cases {
		actual := lbSuffixFromARN(tc.arn)
		if actual != tc.suffix {
			t.Fatalf("bad suffix: %q\nExpected: %s\n     Got: %s", tc.name, tc.suffix, actual)
		}
	}
}

func TestAccAWSLB_ALB_basic(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/app/%s/.+", lbName))),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccAWSLB_NLB_basic(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_networkLoadbalancer(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/net/%s/.+", lbName))),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccAWSLB_LoadBalancerType_Gateway(t *testing.T) {
	var conf elbv2.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_LoadBalancerType_Gateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", elbv2.LoadBalancerTypeEnumGateway),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"idle_timeout",
				},
			},
		},
	})
}

func TestAccAWSLB_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(t *testing.T) {
	var conf elbv2.LoadBalancer
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", elbv2.LoadBalancerTypeEnumGateway),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"drop_invalid_header_fields",
					"enable_http2",
					"idle_timeout",
				},
			},
			{
				Config: testAccAWSLBConfig_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/gwy/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", elbv2.LoadBalancerTypeEnumGateway),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "false"),
				),
			},
		},
	})
}

func TestAccAWSLB_ALB_outpost(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-outpost-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_outpost(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf("loadbalancer/app/%s/.+", lbName))),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_outpost"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_mapping.0.outpost_id"),
					resource.TestCheckResourceAttrSet(resourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
				),
			},
		},
	})
}

func TestAccAWSLB_networkLoadbalancerEIP(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_networkLoadBalancerEIP(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSLB_NLB_privateipv4address(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-pipv4a-%s", acctest.RandString(10))
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_networkLoadBalancerPrivateIPV4Address(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLB_BackwardsCompatibility(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_alb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfigBackwardsCompatibility(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "application"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLB_generatedName(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_generatedName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSLB_generatesNameForZeroValue(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_zeroValueName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSLB_namePrefix(t *testing.T) {
	var conf elbv2.LoadBalancer
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_namePrefix(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestMatchResourceAttr(resourceName, "name",
						regexp.MustCompile("^tf-lb-")),
				),
			},
		},
	})
}

func TestAccAWSLB_tags(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
				),
			},
			{
				Config: testAccAWSLBConfig_updatedTags(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Type", "Sample Type Tag"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
				),
			},
			{
				Config: testAccAWSLBConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
				),
			},
		},
	})
}

func TestAccAWSLB_networkLoadbalancer_updateCrossZone(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-nlbcz-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_networkLoadbalancer(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &pre),
					testAccCheckAWSLBAttribute(resourceName, "load_balancing.cross_zone.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "true"),
				),
			},
			{
				Config: testAccAWSLBConfig_networkLoadbalancer(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &mid),
					testAccCheckAWSLBAttribute(resourceName, "load_balancing.cross_zone.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "false"),
					testAccCheckAWSlbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccAWSLBConfig_networkLoadbalancer(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &post),
					testAccCheckAWSLBAttribute(resourceName, "load_balancing.cross_zone.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_cross_zone_load_balancing", "true"),
					testAccCheckAWSlbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccAWSLB_applicationLoadBalancer_updateHttp2(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawsalb-http2-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_enableHttp2(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &pre),
					testAccCheckAWSLBAttribute(resourceName, "routing.http2.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", "false"),
				),
			},
			{
				Config: testAccAWSLBConfig_enableHttp2(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &mid),
					testAccCheckAWSLBAttribute(resourceName, "routing.http2.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", "true"),
					testAccCheckAWSlbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccAWSLBConfig_enableHttp2(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &post),
					testAccCheckAWSLBAttribute(resourceName, "routing.http2.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_http2", "false"),
					testAccCheckAWSlbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccAWSLB_applicationLoadBalancer_updateDropInvalidHeaderFields(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawsalb-headers-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb.lb_test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_enableDropInvalidHeaderFields(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists("aws_lb.lb_test", &pre),
					testAccCheckAWSLBAttribute("aws_lb.lb_test", "routing.http.drop_invalid_header_fields.enabled", "false"),
					resource.TestCheckResourceAttr("aws_lb.lb_test", "drop_invalid_header_fields", "false"),
				),
			},
			{
				Config: testAccAWSLBConfig_enableDropInvalidHeaderFields(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists("aws_lb.lb_test", &mid),
					testAccCheckAWSLBAttribute("aws_lb.lb_test", "routing.http.drop_invalid_header_fields.enabled", "true"),
					resource.TestCheckResourceAttr("aws_lb.lb_test", "drop_invalid_header_fields", "true"),
					testAccCheckAWSlbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccAWSLBConfig_enableDropInvalidHeaderFields(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists("aws_lb.lb_test", &post),
					testAccCheckAWSLBAttribute("aws_lb.lb_test", "routing.http.drop_invalid_header_fields.enabled", "false"),
					resource.TestCheckResourceAttr("aws_lb.lb_test", "drop_invalid_header_fields", "false"),
					testAccCheckAWSlbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccAWSLB_applicationLoadBalancer_updateDeletionProtection(t *testing.T) {
	var pre, mid, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawsalb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_enableDeletionProtection(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &pre),
					testAccCheckAWSLBAttribute(resourceName, "deletion_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
				),
			},
			{
				Config: testAccAWSLBConfig_enableDeletionProtection(lbName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &mid),
					testAccCheckAWSLBAttribute(resourceName, "deletion_protection.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "true"),
					testAccCheckAWSlbARNs(&pre, &mid),
				),
			},
			{
				Config: testAccAWSLBConfig_enableDeletionProtection(lbName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &post),
					testAccCheckAWSLBAttribute(resourceName, "deletion_protection.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					testAccCheckAWSlbARNs(&mid, &post),
				),
			},
		},
	})
}

func TestAccAWSLB_updatedSecurityGroups(t *testing.T) {
	var pre, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
				),
			},
			{
				Config: testAccAWSLBConfig_updateSecurityGroups(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					testAccCheckAWSlbARNs(&pre, &post),
				),
			},
		},
	})
}

func TestAccAWSLB_updatedSubnets(t *testing.T) {
	var pre, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_basic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
				),
			},
			{
				Config: testAccAWSLBConfig_updateSubnets(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "3"),
					testAccCheckAWSlbARNs(&pre, &post),
				),
			},
		},
	})
}

func TestAccAWSLB_updatedIpAddressType(t *testing.T) {
	var pre, post elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfigWithIpAddressType(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "ipv4"),
				),
			},
			{
				Config: testAccAWSLBConfigWithIpAddressTypeUpdated(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &post),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "dualstack"),
				),
			},
		},
	})
}

// TestAccAWSALB_noSecurityGroup regression tests the issue in #8264,
// where if an ALB is created without a security group, a default one
// is assigned.
func TestAccAWSLB_noSecurityGroup(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-nosg-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_nosg(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(resourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
				),
			},
		},
	})
}

func TestAccAWSLB_ALB_AccessLogs(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", acctest.RandString(6))
	lbName := fmt.Sprintf("testaccawslbaccesslog-%s", acctest.RandString(4))
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfigALBAccessLogs(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigALBAccessLogs(false, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigALBAccessLogs(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigALBAccessLogsNoBlocks(lbName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLB_ALB_AccessLogs_Prefix(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", acctest.RandString(6))
	lbName := fmt.Sprintf("testaccawslbaccesslog-%s", acctest.RandString(4))
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfigALBAccessLogs(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigALBAccessLogs(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigALBAccessLogs(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLB_NLB_AccessLogs(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", acctest.RandString(6))
	lbName := fmt.Sprintf("testaccawslbaccesslog-%s", acctest.RandString(4))
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfigNLBAccessLogs(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigNLBAccessLogs(false, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigNLBAccessLogs(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigNLBAccessLogsNoBlocks(lbName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "false"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLB_NLB_AccessLogs_Prefix(t *testing.T) {
	var conf elbv2.LoadBalancer
	bucketName := fmt.Sprintf("tf-test-access-logs-%s", acctest.RandString(6))
	lbName := fmt.Sprintf("testaccawslbaccesslog-%s", acctest.RandString(4))
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfigNLBAccessLogs(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigNLBAccessLogs(true, lbName, bucketName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLBConfigNLBAccessLogs(true, lbName, bucketName, "prefix1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.bucket", bucketName),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.enabled", "true"),
					testAccCheckAWSLBAttribute(resourceName, "access_logs.s3.prefix", "prefix1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLB_networkLoadbalancer_subnet_change(t *testing.T) {
	var conf elbv2.LoadBalancer
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandString(10))
	resourceName := "aws_lb.lb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBConfig_networkLoadbalancer_subnets(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", lbName),
					resource.TestCheckResourceAttr(resourceName, "internal", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "testAccAWSLBConfig_networkLoadbalancer_subnets"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_type", "network"),
				),
			},
		},
	})
}

func testAccCheckAWSlbARNs(pre, post *elbv2.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(pre.LoadBalancerArn) != aws.StringValue(post.LoadBalancerArn) {
			return errors.New("LB has been recreated. ARNs are different")
		}

		return nil
	}
}

func testAccCheckAWSLBExists(n string, res *elbv2.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LB ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		describe, err := conn.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{
			LoadBalancerArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.LoadBalancers) != 1 ||
			aws.StringValue(describe.LoadBalancers[0].LoadBalancerArn) != rs.Primary.ID {
			return errors.New("LB not found")
		}

		*res = *describe.LoadBalancers[0]
		return nil
	}
}

func testAccCheckAWSLBAttribute(n, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LB ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn
		attributesResp, err := conn.DescribeLoadBalancerAttributes(&elbv2.DescribeLoadBalancerAttributesInput{
			LoadBalancerArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Error retrieving LB Attributes: %s", err)
		}

		for _, attr := range attributesResp.Attributes {
			if aws.StringValue(attr.Key) == key {
				if aws.StringValue(attr.Value) == value {
					return nil
				}
				return fmt.Errorf("LB attribute %s expected: %q actual: %q", key, value, aws.StringValue(attr.Value))
			}
		}
		return fmt.Errorf("LB attribute %s does not exist on LB: %s", key, rs.Primary.ID)
	}
}

func testAccCheckAWSLBDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb" && rs.Type != "aws_alb" {
			continue
		}

		describe, err := conn.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{
			LoadBalancerArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.LoadBalancers) != 0 &&
				aws.StringValue(describe.LoadBalancers[0].LoadBalancerArn) == rs.Primary.ID {
				return fmt.Errorf("LB %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if isLoadBalancerNotFound(err) {
			return nil
		} else {
			return fmt.Errorf("Unexpected error checking LB destroyed: %s", err)
		}
	}

	return nil
}

func testAccPreCheckElbv2GatewayLoadBalancer(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	input := &elbv2.DescribeAccountLimitsInput{}

	output, err := conn.DescribeAccountLimits(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected ELBv2 Gateway Load Balancer PreCheck error: %s", err)
	}

	if output == nil {
		t.Fatal("unexpected ELBv2 Gateway Load Balancer PreCheck error: empty response")
	}

	for _, limit := range output.Limits {
		if limit == nil {
			continue
		}

		if aws.StringValue(limit.Name) == "gateway-load-balancers" {
			return
		}
	}

	t.Skip("skipping acceptance testing: region does not support ELBv2 Gateway Load Balancers")
}

func testAccAWSLBConfigWithIpAddressTypeUpdated(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test_1.id, aws_subnet.alb_test_2.id]

  ip_address_type = "dualstack"

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.lb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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

resource "aws_egress_only_internet_gateway" "igw" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_vpc" "alb_test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-lb-with-ip-address-type-updated"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_subnet" "alb_test_1" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-updated-1"
  }
}

resource "aws_subnet" "alb_test_2" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 2)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-updated-2"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, lbName))
}

func testAccAWSLBConfigWithIpAddressType(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test_1.id, aws_subnet.alb_test_2.id]

  ip_address_type = "ipv4"

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.lb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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

resource "aws_egress_only_internet_gateway" "igw" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_vpc" "alb_test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-lb-with-ip-address-type"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.alb_test.id
}

resource "aws_subnet" "alb_test_1" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-1"
  }
}

resource "aws_subnet" "alb_test_2" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block         = cidrsubnet(aws_vpc.alb_test.ipv6_cidr_block, 8, 2)

  tags = {
    Name = "tf-acc-lb-with-ip-address-type-2"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, lbName))
}

func testAccAWSLBConfig_basic(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccAWSLBConfig_outpost(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.alb_test.id
}

resource "aws_route_table" "test" {
  vpc_id     = aws_vpc.alb_test.id
  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.test]
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.alb_test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_lb" "lb_test" {
  name                       = "%s"
  security_groups            = [aws_security_group.alb_test.id]
  customer_owned_ipv4_pool   = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  idle_timeout               = 30
  enable_deletion_protection = false
  subnets                    = [aws_subnet.alb_test.id]

  tags = {
    Name = "TestAccAWSALB_outpost"
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-outpost"
  }
}

resource "aws_subnet" "alb_test" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = "tf-acc-lb-outpost"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_outpost"
  }
}
`, lbName))
}

func testAccAWSLBConfig_enableHttp2(lbName string, http2 bool) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  enable_http2 = %t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, http2))
}

func testAccAWSLBConfig_enableDropInvalidHeaderFields(lbName string, dropInvalid bool) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  drop_invalid_header_fields = %t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, dropInvalid))
}

func testAccAWSLBConfig_enableDeletionProtection(lbName string, deletion_protection bool) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = %t

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName, deletion_protection))
}

func testAccAWSLBConfig_networkLoadbalancer_subnets(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-network-load-balancer-subnets"
  }
}

resource "aws_lb" "lb_test" {
  name = "%s"

  subnets = [
    aws_subnet.alb_test_1.id,
    aws_subnet.alb_test_2.id,
    aws_subnet.alb_test_3.id,
  ]

  load_balancer_type               = "network"
  internal                         = true
  idle_timeout                     = 60
  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = false

  tags = {
    Name = "testAccAWSLBConfig_networkLoadbalancer_subnets"
  }
}

resource "aws_subnet" "alb_test_1" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-lb-network-load-balancer-subnets-1"
  }
}

resource "aws_subnet" "alb_test_2" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-lb-network-load-balancer-subnets-2"
  }
}

resource "aws_subnet" "alb_test_3" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "tf-acc-lb-network-load-balancer-subnets-3"
  }
}
`, lbName))
}

func testAccAWSLBConfig_networkLoadbalancer(lbName string, cz bool) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name               = "%s"
  internal           = true
  load_balancer_type = "network"

  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = %t

  subnet_mapping {
    subnet_id = aws_subnet.alb_test.id
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "terraform-testacc-network-load-balancer"
  }
}

resource "aws_subnet" "alb_test" {
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-load-balancer"
  }
}
`, lbName, cz))
}

func testAccAWSLBConfig_LoadBalancerType_Gateway(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName))
}

func testAccAWSLBConfig_LoadBalancerType_Gateway_EnableCrossZoneLoadBalancing(rName string, enableCrossZoneLoadBalancing bool) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  enable_cross_zone_load_balancing = %[2]t
  load_balancer_type               = "gateway"
  name                             = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName, enableCrossZoneLoadBalancing))
}

func testAccAWSLBConfig_networkLoadBalancerEIP(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-network-load-balancer-eip"
  }
}

resource "aws_subnet" "public" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.10.${count.index}.0/24"
  vpc_id            = aws_vpc.main.id

  tags = {
    Name = "tf-acc-lb-network-load-balancer-eip-${count.index}"
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.default.id
  }
}

resource "aws_route_table_association" "a" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_lb" "lb_test" {
  name               = "%s"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = aws_subnet.public[0].id
    allocation_id = aws_eip.lb[0].id
  }

  subnet_mapping {
    subnet_id     = aws_subnet.public[1].id
    allocation_id = aws_eip.lb[1].id
  }

  depends_on = [aws_internet_gateway.default]
}

resource "aws_eip" "lb" {
  count = "2"
}
`, lbName))
}

func testAccAWSLBConfig_networkLoadBalancerPrivateIPV4Address(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = "%s"
  internal                   = true
  load_balancer_type         = "network"
  enable_deletion_protection = false

  subnet_mapping {
    subnet_id            = aws_subnet.test.id
    private_ipv4_address = "10.10.0.15"
  }

  tags = {
    Name = "TestAccAWSALB_privateipv4address"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "TestAccAWSALB_privateipv4address"
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "TestAccAWSALB_privateipv4address"
  }
}
`, lbName))
}

func testAccAWSLBConfigBackwardsCompatibility(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_alb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-bc-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccAWSLBConfig_updateSubnets(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-update-subnets"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 3
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-update-subnets-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccAWSLBConfig_generatedName() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_lb" "lb_test" {
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-generated-name"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.alb_test.id

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-generated-name-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`)
}

func testAccAWSLBConfig_zeroValueName() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_lb" "lb_test" {
  name            = ""
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

# See https://github.com/hashicorp/terraform-provider-aws/issues/2498
output "lb_name" {
  value = aws_lb.lb_test.name
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-zero-value-name"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.alb_test.id

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-zero-value-name-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`)
}

func testAccAWSLBConfig_namePrefix() string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), `
resource "aws_lb" "lb_test" {
  name_prefix     = "tf-lb-"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-name-prefix"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-name-prefix-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`)
}

func testAccAWSLBConfig_updatedTags(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Environment = "Production"
    Type        = "Sample Type Tag"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-updated-tags"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-updated-tags-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}

func testAccAWSLBConfigALBAccessLogsBase(bucketName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_elb_service_account" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-access-logs"
  }
}

resource "aws_subnet" "alb_test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-lb-access-logs-${count.index}"
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["s3:PutObject"]
    effect    = "Allow"
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      type        = "AWS"
      identifiers = [data.aws_elb_service_account.current.arn]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, bucketName))
}

func testAccAWSLBConfigALBAccessLogs(enabled bool, lbName, bucketName, bucketPrefix string) string {
	return composeConfig(testAccAWSLBConfigALBAccessLogsBase(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.alb_test.*.id

  access_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, lbName, enabled, bucketPrefix))
}

func testAccAWSLBConfigALBAccessLogsNoBlocks(lbName, bucketName string) string {
	return composeConfig(testAccAWSLBConfigALBAccessLogsBase(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.alb_test.*.id
}
`, lbName))
}

func testAccAWSLBConfigNLBAccessLogsBase(bucketName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_elb_service_account" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-access-logs"
  }
}

resource "aws_subnet" "alb_test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-lb-access-logs-${count.index}"
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["s3:PutObject"]
    effect    = "Allow"
    resources = ["${aws_s3_bucket.test.arn}/*"]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }

  statement {
    actions   = ["s3:GetBucketAcl"]
    effect    = "Allow"
    resources = [aws_s3_bucket.test.arn]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, bucketName))
}

func testAccAWSLBConfigNLBAccessLogs(enabled bool, lbName, bucketName, bucketPrefix string) string {
	return composeConfig(testAccAWSLBConfigNLBAccessLogsBase(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.alb_test.*.id

  access_logs {
    bucket  = aws_s3_bucket_policy.test.bucket
    enabled = %[2]t
    prefix  = %[3]q
  }
}
`, lbName, enabled, bucketPrefix))
}

func testAccAWSLBConfigNLBAccessLogsNoBlocks(lbName, bucketName string) string {
	return composeConfig(testAccAWSLBConfigNLBAccessLogsBase(bucketName), fmt.Sprintf(`
resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.alb_test.*.id
}
`, lbName))
}

func testAccAWSLBConfig_nosg(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name     = "%s"
  internal = true
  subnets  = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-no-sg"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-no-sg-${count.index}"
  }
}
`, lbName))
}

func testAccAWSLBConfig_updateSecurityGroups(lbName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_lb" "lb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id, aws_security_group.alb_test_2.id]
  subnets         = aws_subnet.alb_test.*.id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-update-security-groups"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-update-security-groups-${count.index}"
  }
}

resource "aws_security_group" "alb_test_2" {
  name        = "allow_all_alb_test_2"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic_2"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}
`, lbName))
}
