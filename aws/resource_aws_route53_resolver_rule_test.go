package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_rule", &resource.Sweeper{
		Name: "aws_route53_resolver_rule",
		F:    testSweepRoute53ResolverRules,
		Dependencies: []string{
			"aws_route53_resolver_rule_association",
		},
	})
}

func testSweepRoute53ResolverRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).route53resolverconn

	var errors error
	err = conn.ListResolverRulesPages(&route53resolver.ListResolverRulesInput{}, func(page *route53resolver.ListResolverRulesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, resolverRule := range page.ResolverRules {
			id := aws.StringValue(resolverRule.Id)

			ownerID := aws.StringValue(resolverRule.OwnerId)
			if ownerID != client.(*AWSClient).accountid {
				log.Printf("[INFO] Skipping Route53 Resolver rule %q, owned by %q", id, ownerID)
				continue
			}

			log.Printf("[INFO] Deleting Route53 Resolver rule %q", id)
			_, err := conn.DeleteResolverRule(&route53resolver.DeleteResolverRuleInput{
				ResolverRuleId: aws.String(id),
			})
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				continue
			}
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting Route53 Resolver rule (%s): %w", id, err))
				continue
			}

			err = route53ResolverRuleWaitUntilTargetState(conn, id, 10*time.Minute,
				[]string{route53resolver.ResolverRuleStatusDeleting},
				[]string{route53ResolverRuleStatusDeleted})
			if err != nil {
				errors = multierror.Append(errors, err)
				continue
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver rule sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("error retrievingRoute53 Resolver rules: %w", err))
	}

	return errors
}

func TestAccAWSRoute53ResolverRule_basic(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basicNoTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSRoute53ResolverRule_justDotDomainName(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig("."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSRoute53ResolverRule_trailingDotDomainName(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig("example.com."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSRoute53ResolverRule_tags(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basicTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverRuleConfig_basicTagsChanged,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccRoute53ResolverRuleConfig_basicNoTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverRule_updateName(t *testing.T) {
	var rule1, rule2 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	name1 := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())
	name2 := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basicName(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverRuleConfig_basicName(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule2),
					testAccCheckRoute53ResolverRulesSame(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverRule_forward(t *testing.T) {
	var rule1, rule2, rule3 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	resourceNameEp1 := "aws_route53_resolver_endpoint.foo"
	resourceNameEp2 := "aws_route53_resolver_endpoint.bar"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_forward(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp1, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.6",
						"port": "53",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverRuleConfig_forwardTargetIpChanged(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule2),
					testAccCheckRoute53ResolverRulesSame(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp1, "id"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.7",
						"port": "53",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.17",
						"port": "54",
					}),
				),
			},
			{
				Config: testAccRoute53ResolverRuleConfig_forwardEndpointChanged(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule3),
					testAccCheckRoute53ResolverRulesSame(&rule3, &rule2),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp2, "id"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.7",
						"port": "53",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.17",
						"port": "54",
					}),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverRule_forwardEndpointRecreate(t *testing.T) {
	var rule1, rule2 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	resourceNameEp := "aws_route53_resolver_endpoint.foo"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_forward(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.6",
						"port": "53",
					}),
				),
			},
			{
				Config: testAccRoute53ResolverRuleConfig_forwardEndpointRecreate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule2),
					testAccCheckRoute53ResolverRulesDifferent(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.6",
						"port": "53",
					}),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverRulesSame(before, after *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Arn != *after.Arn {
			return fmt.Errorf("Expected Route 53 Resolver rule ARNs to be the same. But they were: %v, %v", *before.Arn, *after.Arn)
		}
		return nil
	}
}

func testAccCheckRoute53ResolverRulesDifferent(before, after *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Arn == *after.Arn {
			return fmt.Errorf("Expected Route 53 Resolver rule ARNs to be different. But they were both: %v", *before.Arn)
		}
		return nil
	}
}

func testAccCheckRoute53ResolverRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverRule(&route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver rule still exists: %s", rs.Primary.ID)
	}
	return nil
}

func testAccCheckRoute53ResolverRuleExists(n string, rule *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		res, err := conn.GetResolverRule(&route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*rule = *res.ResolverRule

		return nil
	}
}

func testAccRoute53ResolverRuleConfig(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = %[1]q
  rule_type   = "SYSTEM"
}
`, domainName)
}

const testAccRoute53ResolverRuleConfig_basicNoTags = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"
}
`

const testAccRoute53ResolverRuleConfig_basicTags = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`

const testAccRoute53ResolverRuleConfig_basicTagsChanged = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"

  tags = {
    Usage = "changed"
  }
}
`

func testAccRoute53ResolverRuleConfig_basicName(name string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"
  name        = %q
}
`, name)
}

func testAccRoute53ResolverRuleConfig_forward(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "192.0.2.6"
  }
}
`, testAccRoute53ResolverRuleConfig_resolverEndpoint(name), name)
}

func testAccRoute53ResolverRuleConfig_forwardTargetIpChanged(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "192.0.2.7"
  }

  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, testAccRoute53ResolverRuleConfig_resolverEndpoint(name), name)
}

func testAccRoute53ResolverRuleConfig_forwardEndpointChanged(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.bar.id

  target_ip {
    ip = "192.0.2.7"
  }

  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, testAccRoute53ResolverRuleConfig_resolverEndpoint(name), name)
}

func testAccRoute53ResolverRuleConfig_forwardEndpointRecreate(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "192.0.2.6"
  }
}
`, testAccRoute53ResolverRuleConfig_resolverEndpointRecreate(name), name)
}

func testAccRoute53ResolverRuleConfig_resolverVpc(name string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "sn1" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s_1"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s_2"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "%[1]s_3"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = aws_vpc.foo.id
  name   = "%[1]s_1"

  tags = {
    Name = "%[1]s_1"
  }
}

resource "aws_security_group" "sg2" {
  vpc_id = aws_vpc.foo.id
  name   = "%[1]s_2"

  tags = {
    Name = "%[1]s_2"
  }
}
`, name)
}

func testAccRoute53ResolverRuleConfig_resolverEndpoint(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "OUTBOUND"
  name      = "%[2]s_1"

  security_group_ids = [
    aws_security_group.sg1.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
  }
}

resource "aws_route53_resolver_endpoint" "bar" {
  direction = "OUTBOUND"
  name      = "%[2]s_2"

  security_group_ids = [
    aws_security_group.sg1.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn3.id
  }
}
`, testAccRoute53ResolverRuleConfig_resolverVpc(name), name)
}

func testAccRoute53ResolverRuleConfig_resolverEndpointRecreate(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "OUTBOUND"
  name      = "%[2]s_1"

  security_group_ids = [
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
  }
}

resource "aws_route53_resolver_endpoint" "bar" {
  direction = "OUTBOUND"
  name      = "%[2]s_2"

  security_group_ids = [
    aws_security_group.sg1.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn3.id
  }
}
`, testAccRoute53ResolverRuleConfig_resolverVpc(name), name)
}
