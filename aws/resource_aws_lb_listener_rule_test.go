package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "listener_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "100"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRuleBackwardsCompatibility(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_alb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfigBackwardsCompatibility(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_alb_listener_rule.static", &conf),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "arn"),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "listener_arn"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "priority", "100"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_alb_listener_rule.static", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_alb_listener_rule.static", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_redirect(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-redirect-%s", acctest.RandStringFromCharSet(14, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_redirect(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "listener_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "100"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.type", "redirect"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.0.host", "#{host}"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.0.path", "/#{path}"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.0.port", "443"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.0.query", "#{query}"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.0.status_code", "HTTP_301"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_fixedResponse(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-fixedresponse-%s", acctest.RandStringFromCharSet(9, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_fixedResponse(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "listener_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "100"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.type", "fixed-response"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.fixed_response.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.fixed_response.0.content_type", "text/plain"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.fixed_response.0.message_body", "Fixed response content"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "action.0.fixed_response.0.status_code", "200"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.static", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_updateRulePriority(t *testing.T) {
	var rule elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "100"),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_updateRulePriority(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &rule),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.static", "priority", "101"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_changeListenerRuleArnForcesNew(t *testing.T) {
	var before, after elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.static",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &before),
				),
			},
			{
				Config: testAccAWSLBListenerRuleConfig_changeRuleArn(lbName, targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.static", &after),
					testAccCheckAWSLbListenerRuleRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_multipleConditionThrowsError(t *testing.T) {
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLBListenerRuleConfig_multipleConditions(lbName, targetGroupName),
				ExpectError: regexp.MustCompile(`attribute supports 1 item maximum`),
			},
		},
	})
}

func TestAccAWSLBListenerRule_priority(t *testing.T) {
	var rule elbv2.Rule
	lbName := fmt.Sprintf("testrule-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.first",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
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
	lbName := fmt.Sprintf("testrule-cognito-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	certificateName := fmt.Sprintf("testcert-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	cognitoPrefix := fmt.Sprintf("testcog-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.cognito",
		Providers:     testAccProvidersWithTLS,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_cognito(lbName, targetGroupName, certificateName, cognitoPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.cognito", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "listener_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "priority", "100"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.#", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.0.type", "authenticate-cognito"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "action.0.authenticate_cognito.0.user_pool_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "action.0.authenticate_cognito.0.user_pool_client_id"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "action.0.authenticate_cognito.0.user_pool_domain"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.0.authenticate_cognito.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.0.authenticate_cognito.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.1.order", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "action.1.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "action.1.target_group_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.cognito", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.cognito", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_oidc(t *testing.T) {
	var conf elbv2.Rule
	lbName := fmt.Sprintf("testrule-oidc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	certificateName := fmt.Sprintf("testcert-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener_rule.oidc",
		Providers:     testAccProvidersWithTLS,
		CheckDestroy:  testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_oidc(lbName, targetGroupName, certificateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerRuleExists("aws_lb_listener_rule.oidc", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.oidc", "arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.oidc", "listener_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "priority", "100"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.#", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.authorization_endpoint", "https://example.com/authorization_endpoint"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.client_id", "s6BhdRkqt3"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.client_secret", "7Fjfp0ZBr1KtDRbnfVdmIw"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.token_endpoint", "https://example.com/token_endpoint"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.user_info_endpoint", "https://example.com/user_info_endpoint"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.0.authenticate_oidc.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.1.order", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "action.1.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.oidc", "action.1.target_group_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "condition.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "condition.1366281676.field", "path-pattern"),
					resource.TestCheckResourceAttr("aws_lb_listener_rule.oidc", "condition.1366281676.values.#", "1"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_rule.oidc", "condition.1366281676.values.0"),
				),
			},
		},
	})
}

func TestAccAWSLBListenerRule_Action_Order(t *testing.T) {
	var rule elbv2.Rule
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_Action_Order(rName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSLBListenerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRuleConfig_Action_Order(rName),
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

func testAccAWSLBListenerRuleConfig_multipleConditions(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*", "static"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-multiple-conditions"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-multiple-conditions-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_basic(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfigBackwardsCompatibility(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_alb_listener_rule" "static" {
  listener_arn = "${aws_alb_listener.front_end.arn}"
  priority = 100

  action {
    type = "forward"
    target_group_arn = "${aws_alb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_alb_listener" "front_end" {
   load_balancer_arn = "${aws_alb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_alb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_alb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-bc-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_redirect(lbName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "redirect"
    redirect {
      port = "443"
      protocol = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.alb_test.id}"
  protocol = "HTTP"
  port = "80"

  default_action {
    type = "redirect"
    redirect {
      port = "443"
      protocol = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_redirect"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-redirect"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-redirect-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName)
}

func testAccAWSLBListenerRuleConfig_fixedResponse(lbName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code = "200"
    }
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.alb_test.id}"
  protocol = "HTTP"
  port = "80"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code = "200"
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_fixedResponse"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-fixedresponse"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-fixedresponse-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName)
}

func testAccAWSLBListenerRuleConfig_updateRulePriority(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 101

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-update-rule-priority"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-update-rule-priority-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_changeRuleArn(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end_ruleupdate.arn}"
  priority = 101

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb_listener" "front_end_ruleupdate" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "8080"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-change-rule-arn"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-change-rule-arn-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_priorityBase(lbName, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-priority"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-priority-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityBase(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "first" {
  listener_arn = "${aws_lb_listener.front_end.arn}"

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/first/*"]
  }
}

resource "aws_lb_listener_rule" "third" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 3

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/third/*"]
  }

  depends_on = ["aws_lb_listener_rule.first"]
}
`)
}

func testAccAWSLBListenerRuleConfig_priorityLast(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "last" {
  listener_arn = "${aws_lb_listener.front_end.arn}"

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/last/*"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_priorityStatic(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityFirst(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "last" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 7

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/last/*"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_priorityParallelism(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityStatic(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "parallelism" {
  count = 10

  listener_arn = "${aws_lb_listener.front_end.arn}"

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/${count.index}/*"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priorityBase(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "priority50000" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority     = 50000

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/50000/*"]
  }
}
`)
}

// priority out of range (1, 50000)
func testAccAWSLBListenerRuleConfig_priority50001(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "priority50001" {
  listener_arn = "${aws_lb_listener.front_end.arn}"

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/50001/*"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_priorityInUse(lbName, targetGroupName string) string {
	return testAccAWSLBListenerRuleConfig_priority50000(lbName, targetGroupName) + fmt.Sprintf(`
resource "aws_lb_listener_rule" "priority50000_in_use" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority     = 50000

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/50000_in_use/*"]
  }
}
`)
}

func testAccAWSLBListenerRuleConfig_cognito(lbName string, targetGroupName string, certificateName string, cognitoPrefix string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "cognito" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "authenticate-cognito"
    authenticate_cognito {
      user_pool_arn = "${aws_cognito_user_pool.test.arn}"
      user_pool_client_id = "${aws_cognito_user_pool_client.test.id}"
      user_pool_domain = "${aws_cognito_user_pool_domain.test.domain}"

      authentication_request_extra_params = {
        param  = "test"
      }
    }
  }

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_iam_server_certificate" "test" {
  name = "terraform-test-cert-%s"
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.test.private_key_pem}"

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

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTPS"
   port = "443"
   ssl_policy = "ELBSecurityPolicy-2016-08"
   certificate_arn = "${aws_iam_server_certificate.test.arn}"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_cognito"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-cognito"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-cognito-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
  name = "%s-pool"
}

resource "aws_cognito_user_pool_client" "test" {
  name = "%s-pool-client"
  user_pool_id = "${aws_cognito_user_pool.test.id}"
  generate_secret = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows = ["code", "implicit"]
  allowed_oauth_scopes = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]
  callback_urls = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri = "https://www.example.com/redirect"
  logout_urls = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool_domain" "test" {
  domain = "%s-pool-domain"
  user_pool_id = "${aws_cognito_user_pool.test.id}"
}`, lbName, targetGroupName, certificateName, cognitoPrefix, cognitoPrefix, cognitoPrefix)
}

func testAccAWSLBListenerRuleConfig_oidc(lbName string, targetGroupName string, certificateName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener_rule" "oidc" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority = 100

  action {
    type = "authenticate-oidc"
    authenticate_oidc {
      authorization_endpoint =  "https://example.com/authorization_endpoint"
      client_id = "s6BhdRkqt3"
      client_secret = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer = "https://example.com"
      token_endpoint = "https://example.com/token_endpoint"
      user_info_endpoint = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param  = "test"
      }
    }
  }

  action {
    type = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_iam_server_certificate" "test" {
  name = "terraform-test-cert-%s"
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.test.private_key_pem}"

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

resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTPS"
   port = "443"
   ssl_policy = "ELBSecurityPolicy-2016-08"
   certificate_arn = "${aws_iam_server_certificate.test.arn}"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_cognito"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

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
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-rule-cognito"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-rule-cognito-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
}`, lbName, targetGroupName, certificateName)
}

func testAccAWSLBListenerRuleConfig_Action_Order(rName string) string {
	return fmt.Sprintf(`
variable "rName" {
  default = %q
}

data "aws_availability_zones" "available" {}

resource "aws_lb_listener_rule" "test" {
  listener_arn = "${aws_lb_listener.test.arn}"

  action {
    order = 1
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param  = "test"
      }
    }
  }

  action {
    order            = 2
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }

  condition {
    field = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_iam_server_certificate" "test" {
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  name             = "${var.rName}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  validity_period_hours = 12

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = "${aws_lb.test.id}"
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${aws_iam_server_certificate.test.arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.id}"
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  internal                   = true
  name                       = "${var.rName}"
  security_groups            = ["${aws_security_group.test.id}"]
  subnets                    = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
}

resource "aws_lb_target_group" "test" {
  name     = "${var.rName}"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.test.id}"

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
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "${var.rName}"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block              = "10.0.${count.index}.0/24"
  map_public_ip_on_launch = true
  vpc_id                  = "${aws_vpc.test.id}"

  tags = {
    Name = "${var.rName}"
  }
}

resource "aws_security_group" "test" {
  name        = "${var.rName}"
  vpc_id      = "${aws_vpc.test.id}"

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
    Name = "${var.rName}"
  }
}`, rName)
}
