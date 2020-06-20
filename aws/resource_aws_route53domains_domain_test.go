package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSRoute53DomainsDomain_Basic(t *testing.T) {
	domainName := os.Getenv("ROUTE53DOMAINS_DOMAIN")
	if domainName == "" {
		t.Skip("Environment variable ROUTE53DOMAINS_DOMAIN is not set")
	}

	resourceName := "aws_route53domains_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DomainsDomainConfig_Basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestMatchResourceAttr(resourceName, "name_servers.#", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "tags.%", regexp.MustCompile(`^\d+$`)),
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

func TestAccAWSRoute53DomainsDomain_Tags(t *testing.T) {
	domainName := os.Getenv("ROUTE53DOMAINS_DOMAIN")
	if domainName == "" {
		t.Skip("Environment variable ROUTE53DOMAINS_DOMAIN is not set")
	}

	resourceName := "aws_route53domains_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DomainsDomainConfig_TagsSingle(domainName, "tag1key", "tag1value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1key", "tag1value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53DomainsDomainConfig_TagsMultiple(domainName, "tag1key", "tag1valueupdated", "tag2key", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1key", "tag1valueupdated"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2key", "tag2value"),
				),
			},
			{
				Config: testAccRoute53DomainsDomainConfig_TagsSingle(domainName, "tag2key", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2key", "tag2value"),
				),
			},
		},
	})
}

func TestAccAWSRoute53DomainsDomain_AutoRenew(t *testing.T) {
	domainName := os.Getenv("ROUTE53DOMAINS_DOMAIN")
	if domainName == "" {
		t.Skip("Environment variable ROUTE53DOMAINS_DOMAIN is not set")
	}

	resourceName := "aws_route53domains_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DomainsDomainConfig_AutoRenew(domainName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_renew", "false"),
				),
			},
			{
				Config: testAccRoute53DomainsDomainConfig_AutoRenew(domainName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_renew", "true"),
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

func TestAccAWSRoute53DomainsDomain_TransferLock(t *testing.T) {
	domainName := os.Getenv("ROUTE53DOMAINS_DOMAIN")
	if domainName == "" {
		t.Skip("Environment variable ROUTE53DOMAINS_DOMAIN is not set")
	}

	resourceName := "aws_route53domains_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DomainsDomainConfig_TransferLock(domainName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "transfer_lock", "false"),
				),
			},
			{
				Config: testAccRoute53DomainsDomainConfig_TransferLock(domainName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "transfer_lock", "true"),
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

func TestAccAWSRoute53DomainsDomain_Privacy(t *testing.T) {
	domainName := os.Getenv("ROUTE53DOMAINS_DOMAIN")
	if domainName == "" {
		t.Skip("Environment variable ROUTE53DOMAINS_DOMAIN is not set")
	}

	resourceName := "aws_route53domains_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DomainsDomainConfig_Privacy(domainName, true, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", "true"),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", "false"),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", "false"),
				),
			},
			{
				Config: testAccRoute53DomainsDomainConfig_Privacy(domainName, false, true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", "false"),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", "true"),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", "false"),
				),
			},
			{
				Config: testAccRoute53DomainsDomainConfig_Privacy(domainName, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", "true"),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", "true"),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", "true"),
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

func TestAccAWSRoute53DomainsDomain_NameServers(t *testing.T) {
	domainName := os.Getenv("ROUTE53DOMAINS_DOMAIN")
	if domainName == "" {
		t.Skip("Environment variable ROUTE53DOMAINS_DOMAIN is not set")
	}

	resourceName := "aws_route53domains_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DomainsDomainConfig_NameServers(domainName, "b.iana-servers.net", "c.iana-servers.net"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name_servers.0.name", "b.iana-servers.net"),
					resource.TestCheckResourceAttr(resourceName, "name_servers.1.name", "c.iana-servers.net"),
				),
			},
			{
				Config: testAccRoute53DomainsDomainConfig_NameServers(domainName, "a.iana-servers.net", "b.iana-servers.net"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name_servers.0.name", "a.iana-servers.net"),
					resource.TestCheckResourceAttr(resourceName, "name_servers.1.name", "b.iana-servers.net"),
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

func testAccRoute53DomainsDomainConfig_Basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name = "%s"
}
`, domainName)
}

func testAccRoute53DomainsDomainConfig_TagsSingle(domainName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name = "%s"

  tags = {
    %q = %q
  }
}
`, domainName, tag1Key, tag1Value)
}

func testAccRoute53DomainsDomainConfig_TagsMultiple(domainName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, domainName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccRoute53DomainsDomainConfig_AutoRenew(domainName, autoRenew string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name = "%s"
  auto_renew  = %s
}
`, domainName, autoRenew)
}

func testAccRoute53DomainsDomainConfig_TransferLock(domainName, transferLock string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name   = "%s"
  transfer_lock = %s
}
`, domainName, transferLock)
}

func testAccRoute53DomainsDomainConfig_Privacy(domainName string, adminPrivacy, registrantPrivacy, techPrivacy bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name        = "%s"
  admin_privacy      = %t
  registrant_privacy = %t
  tech_privacy       = %t
}
`, domainName, adminPrivacy, registrantPrivacy, techPrivacy)
}

func testAccRoute53DomainsDomainConfig_NameServers(domainName, nameServer1, nameServer2 string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name = "%s"

  name_servers {
	name = "%s"
  }
  name_servers {
	name = "%s"
  }
}
`, domainName, nameServer1, nameServer2)
}
