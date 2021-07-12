package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsEc2ManagedPrefixListEntry_ipv4(t *testing.T) {
	var managedPrefixList ec2.ManagedPrefixList
	var managedPrefixListEntries []*ec2.PrefixListEntry
	entry := &ec2.PrefixListEntry{
		Cidr: aws.String("10.0.0.0/8"),
	}
	rName := acctest.RandomWithPrefix("tf-acc-test")

	testEntryCount := func(*terraform.State) error {
		if len(managedPrefixListEntries) != 1 {
			return fmt.Errorf("Wrong EC2 Managed Prefix List Entry count, expected %d, got %d",
				1, len(managedPrefixListEntries))
		}

		entry := managedPrefixListEntries[0]
		if *entry.Cidr != "10.0.0.0/8" {
			return fmt.Errorf("Wrong EC2 Managed Prefix List Entry, expected %v, got %v",
				"10.0.0.0/8", *entry.Cidr)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryIpv4Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListExists("aws_ec2_managed_prefix_list.web", &managedPrefixList, &managedPrefixListEntries),
					testAccCheckAWSEc2ManagedPrefixListEntryAttributes("aws_ec2_managed_prefix_list_entry.entry_1", &managedPrefixListEntries, entry),
					resource.TestCheckResourceAttrPair(
						"aws_ec2_managed_prefix_list_entry.entry_1", "prefix_list_id", "aws_ec2_managed_prefix_list.web", "id"),
					resource.TestCheckResourceAttr(
						"aws_ec2_managed_prefix_list_entry.entry_1", "cidr", "10.0.0.0/8"),
					testEntryCount,
				),
			},
			{
				ResourceName:      "aws_ec2_managed_prefix_list_entry.entry_1",
				ImportState:       true,
				ImportStateIdFunc: testAccAwsEc2ManagedPrefixListEntryImportStateIdFunc("aws_ec2_managed_prefix_list_entry.entry_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_ipv6(t *testing.T) {
	var managedPrefixList ec2.ManagedPrefixList
	var managedPrefixListEntries []*ec2.PrefixListEntry
	entry := &ec2.PrefixListEntry{
		Cidr: aws.String("::/0"),
	}

	rName := acctest.RandomWithPrefix("tf-acc-test")
	entryName := "aws_ec2_managed_prefix_list_entry.entry_1"

	testEntryCount := func(*terraform.State) error {
		if len(managedPrefixListEntries) != 1 {
			return fmt.Errorf("Wrong EC2 Managed Prefix List Entry count, expected %d, got %d",
				1, len(managedPrefixListEntries))
		}

		entry := managedPrefixListEntries[0]
		if *entry.Cidr != "::/0" {
			return fmt.Errorf("Wrong EC2 Managed Prefix List Entry, expected %v, got %v",
				"::/0", *entry.Cidr)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryIpv6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListExists("aws_ec2_managed_prefix_list.web", &managedPrefixList, &managedPrefixListEntries),
					testAccCheckAWSEc2ManagedPrefixListEntryAttributes("aws_ec2_managed_prefix_list_entry.entry_1", &managedPrefixListEntries, entry),
					resource.TestCheckResourceAttrPair(
						entryName, "prefix_list_id", "aws_ec2_managed_prefix_list.web", "id"),
					resource.TestCheckResourceAttr("aws_ec2_managed_prefix_list_entry.entry_1", "cidr", "::/0"),
					testEntryCount,
				),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_expectInvalidTypeError(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEc2ManagedPrefixListEntryExpectInvalidType(rName),
				ExpectError: regexp.MustCompile(`invalid CIDR address: ::/244`),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_expectInvalidCIDR(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEc2ManagedPrefixListEntryInvalidIPv4CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccAwsEc2ManagedPrefixListEntryInvalidIPv6CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_description(t *testing.T) {
	var managedPrefixList ec2.ManagedPrefixList
	var managedPrefixListEntries []*ec2.PrefixListEntry
	var entry ec2.PrefixListEntry

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListExists("aws_ec2_managed_prefix_list.web", &managedPrefixList, &managedPrefixListEntries),
					testAccCheckAWSEc2ManagedPrefixListEntryAttributes("aws_ec2_managed_prefix_list_entry.entry_1", &managedPrefixListEntries, &entry),
					resource.TestCheckResourceAttr("aws_ec2_managed_prefix_list_entry.entry_1", "description", "TF acceptance test ec2 managed prefix list entry"),
				),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_disappears(t *testing.T) {
	var managedPrefixList ec2.ManagedPrefixList
	var managedPrefixListEntries []*ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.entry_1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryIpv4Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListExists("aws_ec2_managed_prefix_list.web", &managedPrefixList, &managedPrefixListEntries),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ManagedPrefixListEntry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEc2ManagedPrefixListEntryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list" {
			continue
		}

		// Retrieve our list
		req := &ec2.DescribePrefixListsInput{
			PrefixListIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribePrefixLists(req)
		if err == nil {
			if len(resp.PrefixLists) > 0 && *resp.PrefixLists[0].PrefixListId == rs.Primary.ID {
				return fmt.Errorf("EC2 Managed Prefix List (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		// Confirm error code is what we want
		if ec2err.Code() != "InvalidPrefixListId.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSEc2ManagedPrefixListExists(n string, prefixList *ec2.ManagedPrefixList, entries *[]*ec2.PrefixListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		req := &ec2.DescribeManagedPrefixListsInput{
			PrefixListIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeManagedPrefixLists(req)
		if err != nil {
			return err
		}

		if len(resp.PrefixLists) > 0 && *resp.PrefixLists[0].PrefixListId == rs.Primary.ID {
			*prefixList = *resp.PrefixLists[0]
			input := &ec2.GetManagedPrefixListEntriesInput{
				PrefixListId: prefixList.PrefixListId,
			}
			remoteEntries, err := getEc2ManagedPrefixListEntries(conn, input)
			if err != nil {
				return err
			}
			*entries = remoteEntries
			log.Printf("[DEBUG] [within-tests] Entries are : %v", entries)
			return nil
		}

		return fmt.Errorf("EC2 Managed Prefix List Entry not found")
	}
}

func testAccCheckAWSEc2ManagedPrefixListEntryAttributes(n string, entries *[]*ec2.PrefixListEntry, entry *ec2.PrefixListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("EC2 Managed Prefix List Entry Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List Entry is set")
		}

		if entry == nil {
			entry = &ec2.PrefixListEntry{
				Cidr: aws.String("10.0.0.0/8"),
			}
		}

		var matchingEntry *ec2.PrefixListEntry
		log.Printf("[DEBUG] [within-tests] Entries are : %v", entries)
		if len(*entries) == 0 {
			return fmt.Errorf("No Entries")
		}

		for _, r := range *entries {
			if entry.Cidr != nil && r.Cidr != nil && *entry.Cidr != *r.Cidr {
				continue
			}

			matchingEntry = r
		}

		if matchingEntry != nil {
			log.Printf("[DEBUG] Matching entry found : %s", matchingEntry)
			return nil
		}

		return fmt.Errorf("Error here\n\tlooking for %s, wasn't found in %s", entry, entries)
	}
}

func testAccAwsEc2ManagedPrefixListEntryImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		plID := rs.Primary.Attributes["prefix_list_id"]
		cidrBlock := rs.Primary.Attributes["cidr"]

		var parts []string
		parts = append(parts, plID)
		parts = append(parts, cidrBlock)

		return strings.Join(parts, ","), nil
	}
}

func testAccAwsEc2ManagedPrefixListEntryIpv4Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "web" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list_entry" "entry_1" {
  cidr           = "10.0.0.0/8"
  prefix_list_id = aws_ec2_managed_prefix_list.web.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryIpv6Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "web" {
  name           = %[1]q
  address_family = "IPv6"
  max_entries    = 5

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list_entry" "entry_1" {
  cidr           = "::/0"
  prefix_list_id = aws_ec2_managed_prefix_list.web.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryDescriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "web" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list_entry" "entry_1" {
  cidr           = "10.0.0.0/8"
  description    = "TF acceptance test ec2 managed prefix list entry"
  prefix_list_id = aws_ec2_managed_prefix_list.web.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryExpectInvalidType(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "web" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "ipv6" {
  cidr           = "::/244"
  prefix_list_id = aws_ec2_managed_prefix_list.web.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryInvalidIPv4CIDR(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "foo" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "ipv4" {
  cidr           = "1.2.3.4/33"
  prefix_list_id = aws_ec2_managed_prefix_list.foo.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryInvalidIPv6CIDR(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "foo" {
  name           = %[1]q
  address_family = "IPv6"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "ipv6" {
  cidr           = "::/244"
  prefix_list_id = aws_ec2_managed_prefix_list.foo.id
}
`, rName)
}
