package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestLBListenerARNFromRuleARN(t *testing.T) {
	cases := []struct {
		name     string
		arn      string
		expected string
	}{
		{
			name:     "valid listener rule arn",
			arn:      "arn:aws:elasticloadbalancing:us-east-1:012345678912:listener-rule/app/name/0123456789abcdef/abcdef0123456789/456789abcedf1234",
			expected: "arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789",
		},
		{
			name:     "listener arn",
			arn:      "arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789",
			expected: "",
		},
		{
			name:     "some other arn",
			arn:      "arn:aws:elasticloadbalancing:us-east-1:123456:targetgroup/my-targets/73e2d6bc24d8a067",
			expected: "",
		},
		{
			name:     "not an arn",
			arn:      "blah blah blah",
			expected: "",
		},
		{
			name:     "empty arn",
			arn:      "",
			expected: "",
		},
	}

	for _, tc := range cases {
		actual := lbListenerARNFromRuleARN(tc.arn)
		if actual != tc.expected {
			t.Fatalf("incorrect arn returned: %q\nExpected: %s\n     Got: %s", tc.name, tc.expected, actual)
		}
	}
}

func TestAccAWSLBListenerRule_basic(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandString(13))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"
	targetGroupResourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":           "0",
						"http_header.#":           "0",
						"http_request_method.#":   "0",
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
						"query_string.#":          "0",
						"source_ip.#":             "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/static/*"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_forwardWeighted(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-weighted-%s", acctest.RandString(13))
	targetGroupName1 := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))
	targetGroupName2 := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))

	resourceName := "aws_lb_listener_rule.weighted"
	frontEndListenerResourceName := "aws_lb_listener.front_end"
	targetGroup1ResourceName := "aws_lb_target_group.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_forwardWeighted(lbName, targetGroupName1, targetGroupName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_group.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.stickiness.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_changeForwardWeightedStickiness(lbName, targetGroupName1, targetGroupName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_group.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.stickiness.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_changeForwardWeightedToBasic(lbName, targetGroupName1, targetGroupName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.target_group_arn", targetGroup1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_BackwardsCompatibility(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandString(13))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))

	resourceName := "aws_alb_listener_rule.static"
	frontEndListenerResourceName := "aws_alb_listener.front_end"
	targetGroupResourceName := "aws_alb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfigBackwardsCompatibility(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":           "0",
						"http_header.#":           "0",
						"http_request_method.#":   "0",
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
						"query_string.#":          "0",
						"source_ip.#":             "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/static/*"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_redirect(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-redirect-%s", acctest.RandString(14))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_redirect(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "redirect"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.0.host", "#{host}"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.0.path", "/#{path}"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.0.query", "#{query}"),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.0.status_code", "HTTP_301"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_fixedResponse(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-fixedresponse-%s", acctest.RandString(9))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_fixedResponse(lbName, "Fixed response content"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "fixed-response"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.message_body", "Fixed response content"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
		},
	})
}

// Updating Action breaks Condition change logic GH-11323 and GH-11362
func TestAccAWSLBListenerRule_updateFixedResponse(t *testing.T) {
	var rule elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandString(13))

	resourceName := "aws_lb_listener_rule.static"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_fixedResponse(lbName, "Fixed Response 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.message_body", "Fixed Response 1"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_fixedResponse(lbName, "Fixed Response 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.message_body", "Fixed Response 2"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_updateRulePriority(t *testing.T) {
	var rule elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandString(13))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))

	resourceName := "aws_lb_listener_rule.static"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_updateRulePriority(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "priority", "101"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_changeListenerRuleArnForcesNew(t *testing.T) {
	var before, after elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandString(13))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))

	resourceName := "aws_lb_listener_rule.static"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &before),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_changeRuleArn(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &after),
					testAccCheckAWSLbListenerRuleRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_priority(t *testing.T) {
	var rule elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandString(13))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.first", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.first", "priority", "1"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_priorityLast(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.last", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.last", "priority", "4"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_priorityStatic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.last", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.last", "priority", "7"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_priorityLast(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.last", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.last", "priority", "7"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_priorityParallelism(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.last", &rule),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.0", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.1", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.2", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.3", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.4", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.5", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.6", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.7", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.8", "priority"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.parallelism.9", "priority"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.priority50000", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.priority50000", "priority", "50000"),
				),
			},
			{
				Config:      testAccAWSLBListenerRuleConfig_priority50001(lbName, targetGroupName),
				ExpectError: regexp.MustCompile(`Error creating LB Listener Rule: ValidationError`),
			},
			{
				Config:      testAccAWSLBListenerRuleConfig_priorityInUse(lbName, targetGroupName),
				ExpectError: regexp.MustCompile(`Error creating LB Listener Rule: PriorityInUse`),
			},
		},
	})
}

func TestAccAWSLBListenerRule_cognito(t *testing.T) {
	var conf elbv2.Rule
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")
	lbName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_lb_listener_rule.cognito"
	frontEndListenerResourceName := "aws_lb_listener.front_end"
	targetGroupResourceName := "aws_lb_target_group.test"
	cognitoPoolResourceName := "aws_cognito_user_pool.test"
	cognitoPoolClientResourceName := "aws_cognito_user_pool_client.test"
	cognitoPoolDomainResourceName := "aws_cognito_user_pool_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_cognito(lbName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "authenticate-cognito"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.authenticate_cognito.0.user_pool_arn", cognitoPoolResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.authenticate_cognito.0.user_pool_client_id", cognitoPoolClientResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.authenticate_cognito.0.user_pool_domain", cognitoPoolDomainResourceName, "domain"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_cognito.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr(resourceName, "action.1.order", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.1.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "action.1.target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_oidc(t *testing.T) {
	var conf elbv2.Rule
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")
	lbName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_lb_listener_rule.oidc"
	frontEndListenerResourceName := "aws_lb_listener.front_end"
	targetGroupResourceName := "aws_lb_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_oidc(lbName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.authorization_endpoint", "https://example.com/authorization_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.client_id", "s6BhdRkqt3"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.client_secret", "7Fjfp0ZBr1KtDRbnfVdmIw"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.token_endpoint", "https://example.com/token_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.user_info_endpoint", "https://example.com/user_info_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.authenticate_oidc.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr(resourceName, "action.1.order", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.1.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "action.1.target_group_arn", targetGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_Action_Order(t *testing.T) {
	var rule elbv2.Rule
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_Action_Order(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.1.order", "2"),
				),
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6171
func TestAccAWSLBListenerRule_Action_Order_Recreates(t *testing.T) {
	var rule elbv2.Rule
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_Action_Order(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.1.order", "2"),
					testAccCheckAWSLBListenerRuleActionOrderDisappears(&rule, 1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionAttributesCount(t *testing.T) {
	err_many := regexp.MustCompile("Only one of host_header, http_header, http_request_method, path_pattern, query_string or source_ip can be set in a condition block")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBListenerRuleConfig_conditionAttributesCount_http_header(),
				ExpectError: err_many,
			},
			{
				Config:      testAccAWSLBListenerRuleConfig_conditionAttributesCount_http_request_method(),
				ExpectError: err_many,
			},
			{
				Config:      testAccAWSLBListenerRuleConfig_conditionAttributesCount_path_pattern(),
				ExpectError: err_many,
			},
			{
				Config:      testAccAWSLBListenerRuleConfig_conditionAttributesCount_query_string(),
				ExpectError: err_many,
			},
			{
				Config:      testAccAWSLBListenerRuleConfig_conditionAttributesCount_source_ip(),
				ExpectError: err_many,
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionHostHeader(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-hostHeader-%s", acctest.RandString(12))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionHostHeader(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":          "1",
						"host_header.0.values.#": "2",
						"http_header.#":          "0",
						"http_request_method.#":  "0",
						"path_pattern.#":         "0",
						"query_string.#":         "0",
						"source_ip.#":            "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.host_header.0.values.*", "example.com"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.host_header.0.values.*", "www.example.com"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionHttpHeader(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-httpHeader-%s", acctest.RandString(12))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionHttpHeader(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":                  "0",
						"http_header.#":                  "1",
						"http_header.0.http_header_name": "X-Forwarded-For",
						"http_header.0.values.#":         "2",
						"http_request_method.#":          "0",
						"path_pattern.#":                 "0",
						"query_string.#":                 "0",
						"source_ip.#":                    "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_header.0.values.*", "10.0.0.*"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_header.0.values.*", "192.168.1.*"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":                  "0",
						"http_header.#":                  "1",
						"http_header.0.http_header_name": "Zz9~|_^.-+*'&%$#!0aA",
						"http_header.0.values.#":         "1",
						"http_request_method.#":          "0",
						"path_pattern.#":                 "0",
						"query_string.#":                 "0",
						"source_ip.#":                    "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_header.0.values.*", "RFC7230 Validity"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionHttpHeader_invalid(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBListenerRuleConfig_conditionHttpHeader_invalid(),
				ExpectError: regexp.MustCompile(`expected value of condition.0.http_header.0.http_header_name to match regular expression`),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionHttpRequestMethod(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-httpRequest-%s", acctest.RandString(11))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionHttpRequestMethod(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":                  "0",
						"http_header.#":                  "0",
						"http_request_method.#":          "1",
						"http_request_method.0.values.#": "2",
						"path_pattern.#":                 "0",
						"query_string.#":                 "0",
						"source_ip.#":                    "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_request_method.0.values.*", "GET"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_request_method.0.values.*", "POST"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionPathPattern(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-pathPattern-%s", acctest.RandString(11))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionPathPattern(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":           "0",
						"http_header.#":           "0",
						"http_request_method.#":   "0",
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "2",
						"query_string.#":          "0",
						"source_ip.#":             "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/cgi-bin/*"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/public/*"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionQueryString(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-queryString-%s", acctest.RandString(11))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionQueryString(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":         "0",
						"http_header.#":         "0",
						"http_request_method.#": "0",
						"path_pattern.#":        "0",
						"query_string.#":        "2",
						"source_ip.#":           "0",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*.query_string.*", map[string]string{
						"key":   "",
						"value": "surprise",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*.query_string.*", map[string]string{
						"key":   "",
						"value": "blank",
					}),
					// TODO: TypeSet check helpers cannot make distinction between the 2 set items
					// because we had to write a new check for the "downstream" nested set
					// a distinguishing attribute on the outer set would be solve this.
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":         "0",
						"http_header.#":         "0",
						"http_request_method.#": "0",
						"path_pattern.#":        "0",
						"query_string.#":        "2",
						"source_ip.#":           "0",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*.query_string.*", map[string]string{
						"key":   "foo",
						"value": "baz",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*.query_string.*", map[string]string{
						"key":   "foo",
						"value": "bar",
					}),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionSourceIp(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-sourceIp-%s", acctest.RandString(14))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionSourceIp(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":         "0",
						"http_header.#":         "0",
						"http_request_method.#": "0",
						"path_pattern.#":        "0",
						"query_string.#":        "0",
						"source_ip.#":           "1",
						"source_ip.0.values.#":  "2",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "dead:cafe::/64"),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "192.168.0.0/16"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionUpdateMixed(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-mixed-%s", acctest.RandString(17))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionMixed(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/public/*"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"source_ip.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "192.168.0.0/16"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_conditionMixed_updated(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/public/*"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"source_ip.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "dead:cafe::/64"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_conditionMixed_updated2(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/cgi-bin/*"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"source_ip.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "dead:cafe::/64"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionMultiple(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-condMulti-%s", acctest.RandString(13))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionMultiple(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "5"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":                  "0",
						"http_header.#":                  "1",
						"http_header.0.http_header_name": "X-Forwarded-For",
						"http_header.0.values.#":         "1",
						"http_request_method.#":          "0",
						"path_pattern.#":                 "0",
						"query_string.#":                 "0",
						"source_ip.#":                    "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_header.0.values.*", "192.168.1.*"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":         "0",
						"http_header.#":         "0",
						"http_request_method.#": "0",
						"path_pattern.#":        "0",
						"query_string.#":        "0",
						"source_ip.#":           "1",
						"source_ip.0.values.#":  "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "192.168.0.0/16"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":                  "0",
						"http_header.#":                  "0",
						"http_request_method.#":          "1",
						"http_request_method.0.values.#": "1",
						"path_pattern.#":                 "0",
						"query_string.#":                 "0",
						"source_ip.#":                    "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_request_method.0.values.*", "GET"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":           "0",
						"http_header.#":           "0",
						"http_request_method.#":   "0",
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
						"query_string.#":          "0",
						"source_ip.#":             "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/public/*"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":          "1",
						"host_header.0.values.#": "1",
						"http_header.#":          "0",
						"http_request_method.#":  "0",
						"path_pattern.#":         "0",
						"query_string.#":         "0",
						"source_ip.#":            "0",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.host_header.0.values.*", "example.com"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_conditionUpdateMultiple(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-condMulti-%s", acctest.RandString(13))

	resourceName := "aws_lb_listener_rule.static"
	frontEndListenerResourceName := "aws_lb_listener.front_end"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_conditionMultiple(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "5"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"http_header.#":                  "1",
						"http_header.0.http_header_name": "X-Forwarded-For",
						"http_header.0.values.#":         "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_header.0.values.*", "192.168.1.*"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"source_ip.#":          "1",
						"source_ip.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "192.168.0.0/16"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"http_request_method.#":          "1",
						"http_request_method.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_request_method.0.values.*", "GET"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/public/*"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":          "1",
						"host_header.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.host_header.0.values.*", "example.com"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_conditionMultiple_updated(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticloadbalancing", regexp.MustCompile(fmt.Sprintf(`listener-rule/app/%s/.+$`, lbName))),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", frontEndListenerResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "5"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"http_header.#":                  "1",
						"http_header.0.http_header_name": "X-Forwarded-For",
						"http_header.0.values.#":         "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_header.0.values.*", "192.168.2.*"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"source_ip.#":          "1",
						"source_ip.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.source_ip.0.values.*", "192.168.0.0/24"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"http_request_method.#":          "1",
						"http_request_method.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.http_request_method.0.values.*", "POST"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"path_pattern.#":          "1",
						"path_pattern.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.path_pattern.0.values.*", "/public/2/*"),

					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "condition.*", map[string]string{
						"host_header.#":          "1",
						"host_header.0.values.#": "1",
					}),
					tfawsresource.TestCheckTypeSetElemAttr(resourceName, "condition.*.host_header.0.values.*", "foobar.com"),
				),
			},
		},
	})
}

func testAccCheckAWSLBListenerRuleActionOrderDisappears(rule *elbv2.Rule, actionOrderToDelete int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var newActions []*elbv2.Action

		for i, action := range rule.Actions {
			if int(aws.Int64Value(action.Order)) == actionOrderToDelete {
				newActions = append(rule.Actions[:i], rule.Actions[i+1:]...)
				break
			}
		}

		if len(newActions) == 0 {
			return fmt.Errorf("Unable to find action order %d from actions: %#v", actionOrderToDelete, rule.Actions)
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		input := &elbv2.ModifyRuleInput{
			Actions: newActions,
			RuleArn: rule.RuleArn,
		}

		_, err := conn.ModifyRule(input)

		return err
	}
}

func testAccCheckAWSLbListenerRuleRecreated(t *testing.T,
	before, after *elbv2.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.RuleArn == *after.RuleArn {
			t.Fatalf("Expected change of Listener Rule ARNs, but both were %v", before.RuleArn)
		}
		return nil
	}
}

func testAccCheckAWSLBListenerRuleExists(n string, res *elbv2.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Listener Rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		describe, err := conn.DescribeRules(&elbv2.DescribeRulesInput{
			RuleArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.Rules) != 1 ||
			*describe.Rules[0].RuleArn != rs.Primary.ID {
			return errors.New("Listener Rule not found")
		}

		*res = *describe.Rules[0]
		return nil
	}
}

func testAccCheckAWSLBListenerRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_listener_rule" && rs.Type != "aws_alb_listener_rule" {
			continue
		}

		describe, err := conn.DescribeRules(&elbv2.DescribeRulesInput{
			RuleArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.Rules) != 0 &&
				*describe.Rules[0].RuleArn == rs.Primary.ID {
				return fmt.Errorf("Listener Rule %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if isAWSErr(err, elbv2.ErrCodeRuleNotFoundException, "") {
			return nil
		} else {
			return fmt.Errorf("Unexpected error checking LB Listener Rule destroyed: %s", err)
		}
	}

	return nil
}

func testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-basic-${count.index}"
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
`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_forwardWeighted(lbName, targetGroupName1 string, targetGroupName2 string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "weighted" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test1.arn
        weight = 1
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 1
      }
    }
  }

  condition {
    path_pattern {
      values = ["/weighted/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test1.arn
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test1" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

resource "aws_lb_target_group" "test2" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-basic-${count.index}"
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
`, lbName, targetGroupName1, targetGroupName2)
}

func testAccAWSLBListenerRuleConfig_changeForwardWeightedStickiness(lbName, targetGroupName1 string, targetGroupName2 string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "weighted" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test1.arn
        weight = 1
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 1
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }

  condition {
    path_pattern {
      values = ["/weighted/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test1.arn
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test1" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

resource "aws_lb_target_group" "test2" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-basic-${count.index}"
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
`, lbName, targetGroupName1, targetGroupName2)
}

func testAccAWSLBListenerRuleConfig_changeForwardWeightedToBasic(lbName, targetGroupName1 string, targetGroupName2 string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "weighted" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test1.arn
  }

  condition {
    path_pattern {
      values = ["/weighted/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test1.arn
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test1" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

resource "aws_lb_target_group" "test2" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-basic-${count.index}"
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
`, lbName, targetGroupName1, targetGroupName2)
}

func testAccAWSLBListenerRuleConfigBackwardsCompatibility(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_alb_listener_rule" "static" {
  listener_arn = aws_alb_listener.front_end.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_alb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_alb_listener" "front_end" {
  load_balancer_arn = aws_alb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_alb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_alb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-bc-${count.index}"
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
`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_redirect(lbName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_redirect"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-redirect"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-redirect-${count.index}"
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
    Name = "TestAccAWSALB_redirect"
  }
}
`, lbName)
}

func testAccAWSLBListenerRuleConfig_fixedResponse(lbName, response string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "%s"
      status_code  = "200"
    }
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code  = "200"
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_fixedResponse"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-fixedresponse"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-fixedresponse-${count.index}"
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
    Name = "TestAccAWSALB_fixedresponse"
  }
}
`, response, lbName)
}

func testAccAWSLBListenerRuleConfig_updateRulePriority(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 101

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-update-rule-priority"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-update-rule-priority-${count.index}"
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
`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_changeRuleArn(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end_ruleupdate.arn
  priority     = 101

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_listener" "front_end_ruleupdate" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "8080"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-change-rule-arn"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-change-rule-arn-${count.index}"
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
`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_priorityBase(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-priority"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-priority-${count.index}"
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
`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityBase(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "first" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/first/*"]
    }
  }
}

resource "aws_lb_listener_rule" "third" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 3

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/third/*"]
    }
  }

  depends_on = [aws_lb_listener_rule.first]
}
`
}

func testAccAWSLBListenerRuleConfig_priorityLast(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "last" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/last/*"]
    }
  }
}
`
}

func testAccAWSLBListenerRuleConfig_priorityStatic(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "last" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 7

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/last/*"]
    }
  }
}
`
}

func testAccAWSLBListenerRuleConfig_priorityParallelism(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityStatic(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "parallelism" {
  count = 10

  listener_arn = aws_lb_listener.front_end.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/${count.index}/*"]
    }
  }
}
`
}

func testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityBase(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "priority50000" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 50000

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/50000/*"]
    }
  }
}
`
}

// priority out of range (1, 50000)
func testAccAWSLBListenerRuleConfig_priority50001(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "priority50001" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/50001/*"]
    }
  }
}
`
}

func testAccAWSLBListenerRuleConfig_priorityInUse(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName) + `
resource "aws_lb_listener_rule" "priority50000_in_use" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 50000

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/50000_in_use/*"]
    }
  }
}
`
}

func testAccAWSLBListenerRuleConfig_cognito(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "cognito" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "authenticate-cognito"

    authenticate_cognito {
      user_pool_arn       = aws_cognito_user_pool.test.arn
      user_pool_client_id = aws_cognito_user_pool_client.test.id
      user_pool_domain    = aws_cognito_user_pool_domain.test.domain

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%[1]s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_cognito"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%[1]s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-cognito"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-cognito-${count.index}"
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
    Name = "TestAccAWSALB_cognito"
  }
}

resource "aws_cognito_user_pool" "test" {
  name = "%[1]s-pool"
}

resource "aws_cognito_user_pool_client" "test" {
  name                                 = "%[1]s-pool-client"
  user_pool_id                         = aws_cognito_user_pool.test.id
  generate_secret                      = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_scopes                 = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]
  callback_urls                        = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri                 = "https://www.example.com/redirect"
  logout_urls                          = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = "%[1]s-pool-domain"
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAWSLBListenerRuleConfig_oidc(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "oidc" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%[1]s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_cognito"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%[1]s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-cognito"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-listener-rule-cognito-${count.index}"
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
    Name = "TestAccAWSALB_cognito"
  }
}
`, rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAWSLBListenerRuleConfig_Action_Order(rName, key, certificate string) string {
	return fmt.Sprintf(`
variable "rName" {
  default = %[1]q
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn

  action {
    order = 1
    type  = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  action {
    order            = 2
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_iam_server_certificate" "test" {
  certificate_body = "%[2]s"
  name             = var.rName
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  internal        = true
  name            = var.rName
  security_groups = [aws_security_group.test.id]
  subnets         = [aws_subnet.test[0].id, aws_subnet.test[1].id]
}

resource "aws_lb_target_group" "test" {
  name     = var.rName
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

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
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = "10.0.${count.index}.0/24"
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

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
    Name = var.rName
  }
}
`, rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAWSLBListenerRuleConfig_condition_error(condition string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "error" {
  listener_arn = "arn:aws:elasticloadbalancing:us-west-2:111111111111:listener/app/example/1234567890abcdef/1234567890abcdef"
  priority     = 100

  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Static"
      status_code  = 200
    }
  }

  %s
}
`, condition)
}

func testAccAWSLBListenerRuleConfig_conditionAttributesCount_http_header() string {
	return testAccAWSLBListenerRuleConfig_condition_error(`
condition {
  host_header {
    values = ["example.com"]
  }

  http_header {
    http_header_name = "X-Clacks-Overhead"
    values           = ["GNU Terry Pratchett"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_conditionAttributesCount_http_request_method() string {
	return testAccAWSLBListenerRuleConfig_condition_error(`
condition {
  host_header {
    values = ["example.com"]
  }

  http_request_method {
    values = ["POST"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_conditionAttributesCount_path_pattern() string {
	return testAccAWSLBListenerRuleConfig_condition_error(`
condition {
  host_header {
    values = ["example.com"]
  }

  path_pattern {
    values = ["/"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_conditionAttributesCount_query_string() string {
	return testAccAWSLBListenerRuleConfig_condition_error(`
condition {
  host_header {
    values = ["example.com"]
  }

  query_string {
    key   = "foo"
    value = "bar"
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_conditionAttributesCount_source_ip() string {
	return testAccAWSLBListenerRuleConfig_condition_error(`
condition {
  host_header {
    values = ["example.com"]
  }

  source_ip {
    values = ["192.168.0.0/16"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_condition_base(condition, name, lbName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Static"
      status_code  = 200
    }
  }

  %s
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Not Found"
      status_code  = 404
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test[0].id, aws_subnet.alb_test[1].id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_condition%s"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "TestAccAWSALB_condition%s"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "TestAccAWSALB_condition%s-${count.index}"
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
    Name = "TestAccAWSALB_condition%s"
  }
}
`, condition, lbName, name, name, name, name)
}

func testAccAWSLBListenerRuleConfig_conditionHostHeader(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  host_header {
    values = ["example.com", "www.example.com"]
  }
}
`, "HostHeader", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionHttpHeader(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  http_header {
    http_header_name = "X-Forwarded-For"
    values           = ["192.168.1.*", "10.0.0.*"]
  }
}

condition {
  http_header {
    http_header_name = "Zz9~|_^.-+*'&%$#!0aA"
    values           = ["RFC7230 Validity"]
  }
}
`, "HttpHeader", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionHttpHeader_invalid() string {
	return `
resource "aws_lb_listener_rule" "static" {
  listener_arn = "arn:aws:elasticloadbalancing:us-west-2:111111111111:listener/app/test/xxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxx"
  priority     = 100

  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Static"
      status_code  = 200
    }
  }

  condition {
    http_header {
      http_header_name = "Invalid@"
      values           = ["RFC7230 Validity"]
    }
  }
}
`
}

func testAccAWSLBListenerRuleConfig_conditionHttpRequestMethod(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  http_request_method {
    values = ["GET", "POST"]
  }
}
`, "HttpRequestMethod", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionPathPattern(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  path_pattern {
    values = ["/public/*", "/cgi-bin/*"]
  }
}
`, "PathPattern", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionQueryString(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  query_string {
    value = "surprise"
  }

  query_string {
    key   = ""
    value = "blank"
  }
}

condition {
  query_string {
    key   = "foo"
    value = "bar"
  }

  query_string {
    key   = "foo"
    value = "baz"
  }
}
`, "QueryString", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionSourceIp(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  source_ip {
    values = [
      "192.168.0.0/16",
      "dead:cafe::/64",
    ]
  }
}
`, "SourceIp", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionMixed(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  path_pattern {
    values = ["/public/*"]
  }
}

condition {
  source_ip {
    values = [
      "192.168.0.0/16",
    ]
  }
}
`, "Mixed", lbName)
}

// Update new style condition without modifying deprecated. Issue GH-11323
func testAccAWSLBListenerRuleConfig_conditionMixed_updated(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  path_pattern {
    values = ["/public/*"]
  }
}

condition {
  source_ip {
    values = [
      "dead:cafe::/64",
    ]
  }
}
`, "Mixed", lbName)
}

// Then update deprecated syntax without touching new. Issue GH-11362
func testAccAWSLBListenerRuleConfig_conditionMixed_updated2(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  path_pattern {
    values = ["/cgi-bin/*"]
  }
}

condition {
  source_ip {
    values = [
      "dead:cafe::/64",
    ]
  }
}
`, "Mixed", lbName)
}

// Currently a maximum of 5 condition values per rule
func testAccAWSLBListenerRuleConfig_conditionMultiple(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  host_header {
    values = ["example.com"]
  }
}

condition {
  http_header {
    http_header_name = "X-Forwarded-For"
    values           = ["192.168.1.*"]
  }
}

condition {
  http_request_method {
    values = ["GET"]
  }
}

condition {
  path_pattern {
    values = ["/public/*"]
  }
}

condition {
  source_ip {
    values = ["192.168.0.0/16"]
  }
}
`, "Multiple", lbName)
}

func testAccAWSLBListenerRuleConfig_conditionMultiple_updated(lbName string) string {
	return testAccAWSLBListenerRuleConfig_condition_base(`
condition {
  host_header {
    values = ["foobar.com"]
  }
}

condition {
  http_header {
    http_header_name = "X-Forwarded-For"
    values           = ["192.168.2.*"]
  }
}

condition {
  http_request_method {
    values = ["POST"]
  }
}

condition {
  path_pattern {
    values = ["/public/2/*"]
  }
}

condition {
  source_ip {
    values = ["192.168.0.0/24"]
  }
}
`, "Multiple", lbName)
}
