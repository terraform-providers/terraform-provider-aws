package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAWSFmsPolicy_basic(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFmsAdmin(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck:   testAccErrorCheck(t, fms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount("aws_fms_policy.test", "arn", "fms", "policy/.+"),
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

func testAccAWSFmsPolicy_cloudfrontDistribution(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFmsAdmin(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck:   testAccErrorCheck(t, fms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_cloudfrontDistribution(fmsPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount("aws_fms_policy.test", "arn", "fms", "policy/.+"),
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

func testAccAWSFmsPolicy_includeMap(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFmsAdmin(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck:   testAccErrorCheck(t, fms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_include(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount("aws_fms_policy.test", "arn", "fms", "policy/.+"),
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

func testAccAWSFmsPolicy_update(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	fmsPolicyName2 := fmt.Sprintf("tf-fms-%s2", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFmsAdmin(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck:   testAccErrorCheck(t, fms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount("aws_fms_policy.test", "arn", "fms", "policy/.+"),
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

func testAccAWSFmsPolicy_tags(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFmsAdmin(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck:   testAccErrorCheck(t, fms.EndpointsID),
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

func testAccFmsPolicyConfigBase() string {
	return composeConfig(
		testAccFmsAdminRegionProviderConfig(),
		`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["fms.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_fms_admin_account" "test" {
  account_id = aws_organizations_organization.test.master_account_id
}
`)
}

func testAccFmsPolicyConfig(name string, group string) string {
	return composeConfig(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  exclude_map {
    account = [data.aws_caller_identity.current.account_id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}

func testAccFmsPolicyConfig_cloudfrontDistribution(name string) string {
	return composeConfig(
		testAccFmsPolicyConfigBase(),
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

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "sts:ExternalId" = "${data.aws_caller_identity.current.account_id}"
        }
      }
      Effect = "Allow"
      Principal = {
        Service = "firehose.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q

  inline_policy {
    name = "test"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [{
        Action = [
          "s3:AbortMultipartUpload",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:PutObject",
        ]
        Effect = "Allow"
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*"
        ]
        Sid = ""
      }]
    })
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = "aws-waf-logs-%[1]s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, name))
}

func testAccFmsPolicyConfig_updated(name string, group string) string {
	return composeConfig(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = true
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  exclude_map {
    account = [data.aws_caller_identity.current.account_id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"ALLOW\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  lifecycle {
    create_before_destroy = false
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}

func testAccFmsPolicyConfig_include(name string, group string) string {
	return composeConfig(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = "%[1]s"
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  include_map {
    account = [data.aws_caller_identity.current.account_id]
  }

  security_service_policy_data {
    type                 = "WAF"
    managed_service_data = "{\"type\": \"WAF\", \"ruleGroups\": [{\"id\":\"${aws_wafregional_rule_group.test.id}\", \"overrideAction\" : {\"type\": \"COUNT\"}}],\"defaultAction\": {\"type\": \"BLOCK\"}, \"overrideCustomerWebACLAssociation\": false}"
  }

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}

func testAccFmsPolicyConfig_tags(name string, group string) string {
	return composeConfig(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
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

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}

func testAccFmsPolicyConfig_tagsChanged(name string, group string) string {
	return composeConfig(
		testAccFmsPolicyConfigBase(),
		fmt.Sprintf(`
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

  depends_on = [aws_fms_admin_account.test]
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = "%[2]s"
}
`, name, group))
}
