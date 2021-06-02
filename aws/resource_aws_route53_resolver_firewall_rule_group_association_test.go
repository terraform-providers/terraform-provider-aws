package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_firewall_rule_group_association", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule_group_association",
		F:    testSweepRoute53ResolverFirewallRuleGroupAssociations,
	})
}

func testSweepRoute53ResolverFirewallRuleGroupAssociations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).route53resolverconn
	var sweeperErrs *multierror.Error

	err = conn.ListFirewallRuleGroupAssociationsPages(&route53resolver.ListFirewallRuleGroupAssociationsInput{}, func(page *route53resolver.ListFirewallRuleGroupAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, firewallRuleGroupAssociation := range page.FirewallRuleGroupAssociations {
			id := aws.StringValue(firewallRuleGroupAssociation.Id)

			log.Printf("[INFO] Deleting Route53 Resolver DNS Firewall rule group association: %s", id)
			r := resourceAwsRoute53ResolverFirewallRuleGroupAssociation()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall rule group associations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall rule group associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRoute53ResolverFirewallRuleGroupAssociation_basic(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheck(t, route53resolver.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "priority", "101"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
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

func TestAccAWSRoute53ResolverFirewallRuleGroupAssociation_name(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNewName := acctest.RandomWithPrefix("tf-acc-test2")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheck(t, route53resolver.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rNewName),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverFirewallRuleGroupAssociation_mutationProtection(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheck(t, route53resolver.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig_mutationProtection(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig_mutationProtection(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "DISABLED"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverFirewallRuleGroupAssociation_priority(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheck(t, route53resolver.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig_priority(rName, 101),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "priority", "101"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig_priority(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "priority", "200"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverFirewallRuleGroupAssociation_disappears(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheck(t, route53resolver.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53ResolverFirewallRuleGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ResolverFirewallRuleGroupAssociation_tags(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheck(t, route53resolver.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoute53ResolverFirewallRuleGroupAssociationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverFirewallRuleGroupAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_firewall_rule_group_association" {
			continue
		}

		// Try to find the resource
		_, err := finder.FirewallRuleGroupAssociationByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver DNS Firewall rule group association still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverFirewallRuleGroupAssociationExists(n string, v *route53resolver.FirewallRuleGroupAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver DNS Firewall rule group association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		out, err := finder.FirewallRuleGroupAssociationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccRoute53ResolverFirewallRuleGroupAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRoute53ResolverFirewallRuleGroupAssociationConfig(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  mutation_protection    = "DISABLED"
  priority               = 101
  vpc_id                 = aws_vpc.test.id
}
`, testAccRoute53ResolverFirewallRuleGroupAssociationConfig_base(rName), rName)
}

func testAccRoute53ResolverFirewallRuleGroupAssociationConfig_mutationProtection(rName, mutationProtection string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  mutation_protection    = %[3]q
  priority               = 101
  vpc_id                 = aws_vpc.test.id
}
`, testAccRoute53ResolverFirewallRuleGroupAssociationConfig_base(rName), rName, mutationProtection)
}

func testAccRoute53ResolverFirewallRuleGroupAssociationConfig_priority(rName string, priority int) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = %[3]d
  vpc_id                 = aws_vpc.test.id
}
`, testAccRoute53ResolverFirewallRuleGroupAssociationConfig_base(rName), rName, priority)
}

func testAccRoute53ResolverFirewallRuleGroupAssociationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 101
  vpc_id                 = aws_vpc.test.id

  tags = {
    %[3]q = %[4]q
  }
}
`, testAccRoute53ResolverFirewallRuleGroupAssociationConfig_base(rName), rName, tagKey1, tagValue1)
}

func testAccRoute53ResolverFirewallRuleGroupAssociationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 101
  vpc_id                 = aws_vpc.test.id

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, testAccRoute53ResolverFirewallRuleGroupAssociationConfig_base(rName), rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
