package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSRoute53ResolverRuleDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_rule.example"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_rule_id"
	ds2ResourceName := "data.aws_route53_resolver_rule.by_domain_name"
	ds3ResourceName := "data.aws_route53_resolver_rule.by_name_and_rule_type"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck: testAccErrorCheckSkipRoute53(t),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ResolverRule_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverRuleDataSource_ResolverEndpointIdWithTags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_rule.example"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck: testAccErrorCheckSkipRoute53(t),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ResolverRule_resolverEndpointIdWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverRuleDataSource_SharedByMe(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_rule.example"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSRoute53Resolver(t)
		},
		ErrorCheck:        testAccErrorCheckSkipRoute53(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ResolverRule_sharedByMe(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttr(ds1ResourceName, "share_status", "SHARED_BY_ME"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ResolverRuleDataSource_SharedWithMe(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_rule.example"
	ds1ResourceName := "data.aws_route53_resolver_rule.by_resolver_endpoint_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSRoute53Resolver(t)
		},
		ErrorCheck:        testAccErrorCheckSkipRoute53(t),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ResolverRule_sharedWithMe(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "domain_name", resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_endpoint_id", resourceName, "resolver_endpoint_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "resolver_rule_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "rule_type", resourceName, "rule_type"),
					resource.TestCheckResourceAttr(ds1ResourceName, "share_status", "SHARED_WITH_ME"),
					// Tags cannot be retrieved for rules shared with us.
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.%", "0"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceAwsRoute53ResolverRule_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "%[1]s.example.com"
  rule_type   = "SYSTEM"
  name        = %[1]q
}

data "aws_route53_resolver_rule" "by_resolver_rule_id" {
  resolver_rule_id = aws_route53_resolver_rule.example.id
}

data "aws_route53_resolver_rule" "by_domain_name" {
  domain_name = aws_route53_resolver_rule.example.domain_name
}

data "aws_route53_resolver_rule" "by_name_and_rule_type" {
  name      = aws_route53_resolver_rule.example.name
  rule_type = aws_route53_resolver_rule.example.rule_type
}
`, rName)
}

func testAccDataSourceAwsRoute53ResolverRule_resolverEndpointIdWithTags(rName string) string {
	return testAccRoute53ResolverRuleConfig_resolverEndpoint(rName) + fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "%[1]s.example.com"
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.bar.id

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  resolver_endpoint_id = aws_route53_resolver_rule.example.resolver_endpoint_id
}
`, rName)
}

func testAccDataSourceAwsRoute53ResolverRule_sharedByMe(rName string) string {
	return testAccAlternateAccountProviderConfig() + testAccRoute53ResolverRuleConfig_resolverEndpoint(rName) + fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "%[1]s.example.com"
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.bar.id

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_route53_resolver_rule.example.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_organizations_organization" "test" {}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  resolver_endpoint_id = aws_route53_resolver_rule.example.resolver_endpoint_id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName)
}

func testAccDataSourceAwsRoute53ResolverRule_sharedWithMe(rName string) string {
	return testAccAlternateAccountProviderConfig() + testAccRoute53ResolverRuleConfig_resolverEndpoint(rName) + fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "%[1]s.example.com"
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.bar.id

  target_ip {
    ip = "192.0.2.7"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2"
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_route53_resolver_rule.example.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_organizations_organization" "test" {}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_route53_resolver_rule" "by_resolver_endpoint_id" {
  provider = "awsalternate"

  resolver_endpoint_id = aws_route53_resolver_rule.example.resolver_endpoint_id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName)
}
