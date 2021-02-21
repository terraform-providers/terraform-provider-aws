package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSFmsPolicy_basic(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccMatchResourceAttrRegionalARN("aws_fms_policy.test", "arn", "fms", regexp.MustCompile(`policy/`)),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				ResourceName:            "aws_fms_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_update_token", "delete_all_policy_resources"},
			},
		},
	})
}

func TestAccAWSFmsPolicy_cloudfrontDistribution(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_cloudfrontDistribution(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccMatchResourceAttrRegionalARN("aws_fms_policy.test", "arn", "fms", regexp.MustCompile(`policy/`)),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				ResourceName:            "aws_fms_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_update_token", "delete_all_policy_resources"},
			},
		},
	})
}

func TestAccAWSFmsPolicy_includeMap(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_include(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccMatchResourceAttrRegionalARN("aws_fms_policy.test", "arn", "fms", regexp.MustCompile(`policy/`)),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				ResourceName:            "aws_fms_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_update_token", "delete_all_policy_resources"},
			},
		},
	})
}

func TestAccAWSFmsPolicy_update(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	fmsPolicyName2 := fmt.Sprintf("tf-fms-%s2", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccMatchResourceAttrRegionalARN("aws_fms_policy.test", "arn", "fms", regexp.MustCompile(`policy/`)),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
			{
				Config: testAccFmsPolicyConfig_updated(fmsPolicyName2, wafRuleGroupName),
			},
		},
	})
}

func TestAccAWSFmsPolicy_tags(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_tags(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.%", "2"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.Usage", "original"),
				),
			},
			{
				Config: testAccFmsPolicyConfig_tagsChanged(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.%", "1"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckAwsFmsPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_policy" {
			continue
		}

		policyID := rs.Primary.Attributes["id"]

		input := &fms.GetPolicyInput{
			PolicyId: aws.String(policyID),
		}

		resp, err := conn.GetPolicy(input)

		if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp.Policy.PolicyId != nil {
			return fmt.Errorf("[DESTROY Error] Fms Policy (%s) not deleted", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAwsFmsPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccFmsPolicyConfig(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  exclude_map {
    account = [data.aws_organizations_organization.example.accounts[0].id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }
}

data "aws_organizations_organization" "example" {}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group)
}

func testAccFmsPolicyConfig_cloudfrontDistribution(name string, group string) string {
	return composeConfig(
		testAccWebACLLoggingConfigurationDependenciesConfig(name),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(name),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type         = "AWS::CloudFront::Distribution"

  security_service_policy_data {
    type                 = "WAFV2"
    managed_service_data = "{\"type\":\"WAFV2\",\"preProcessRuleGroups\":[{\"ruleGroupArn\":null,\"overrideAction\":{\"type\":\"NONE\"},\"managedRuleGroupIdentifier\":{\"version\":null,\"vendorName\":\"AWS\",\"managedRuleGroupName\":\"AWSManagedRulesAmazonIpReputationList\"},\"ruleGroupType\":\"ManagedRuleGroup\",\"excludeRules\":[]}],\"postProcessRuleGroups\":[],\"defaultAction\":{\"type\":\"ALLOW\"},\"overrideCustomerWebACLAssociation\":false,\"loggingConfiguration\":{\"logDestinationConfigs\":[\"${aws_kinesis_firehose_delivery_stream.test.arn}\"],\"redactedFields\":[{\"redactedFieldType\":\"SingleHeader\",\"redactedFieldValue\":\"Cookies\"}]}}"
  }
}


resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group),
	)
}

func testAccFmsPolicyConfig_updated(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = true
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  exclude_map {
    account = [data.aws_organizations_organization.example.accounts[0].id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"ALLOW\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  lifecycle {
    create_before_destroy = false
  }
}

data "aws_organizations_organization" "example" {}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}

resource "aws_wafregional_rule_group" "test2" {
  metric_name = "MyTest2"
  name        = "%[2]s"
}
`, name, group)
}

func testAccFmsPolicyConfig_include(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  include_map {
    account = [data.aws_organizations_organization.example.accounts[0].id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }
}

data "aws_organizations_organization" "example" {}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group)
}

func testAccFmsPolicyConfig_tags(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  resource_tags = {
    Environment = "Testing"
    Usage       = "original"
  }

}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group)
}

func testAccFmsPolicyConfig_tagsChanged(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  resource_tags = {
    Usage = "changed"
  }

}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group)
}
