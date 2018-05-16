package aws

import (
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestCleanPrefix(t *testing.T) {
	cases := []struct {
		Input, Prefix, Output string
	}{
		{"/hostedzone/foo", "/hostedzone/", "foo"},
		{"/change/foo", "/change/", "foo"},
		{"/bar", "/test", "/bar"},
	}

	for _, tc := range cases {
		actual := cleanPrefix(tc.Input, tc.Prefix)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestCleanZoneID(t *testing.T) {
	cases := []struct {
		Input, Output string
	}{
		{"/hostedzone/foo", "foo"},
		{"/change/foo", "/change/foo"},
		{"/bar", "/bar"},
	}

	for _, tc := range cases {
		actual := cleanZoneID(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestCleanChangeID(t *testing.T) {
	cases := []struct {
		Input, Output string
	}{
		{"/hostedzone/foo", "/hostedzone/foo"},
		{"/change/foo", "foo"},
		{"/bar", "/bar"},
	}

	for _, tc := range cases {
		actual := cleanChangeID(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestAccAWSRoute53Zone_basic(t *testing.T) {
	var zone, zone0, zone1, zone2, zone3, zone4 route53.GetHostedZoneOutput
	var td route53.ResourceTagSet

	rString := acctest.RandString(8)
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route53_zone.main",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53ZoneConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.main", &zone),
					testAccLoadTagsR53(&zone, &td),
					testAccCheckTagsR53(&td.Tags, "foo", "bar"),
				),
			},
			resource.TestStep{
				Config: testAccRoute53ZoneCountConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.main.0", &zone0),
					testAccCheckDomainName(&zone0, "subdomain0.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.main.1", &zone1),
					testAccCheckDomainName(&zone1, "subdomain1.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.main.2", &zone2),
					testAccCheckDomainName(&zone2, "subdomain2.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.main.3", &zone3),
					testAccCheckDomainName(&zone3, "subdomain3.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.main.4", &zone4),
					testAccCheckDomainName(&zone4, "subdomain4.terraformtest.com."),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_forceDestroy(t *testing.T) {
	var zone, zoneWithDot route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	zoneName1 := fmt.Sprintf("%s-one.terraformtest.com", rString)
	zoneName2 := fmt.Sprintf("%s-two.terraformtest.com", rString)

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "aws_route53_zone.destroyable",
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckRoute53ZoneDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53ZoneConfig_forceDestroy(zoneName1, zoneName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExistsWithProvider("aws_route53_zone.destroyable", &zone,
						testAccAwsRegionProviderFunc("us-west-2", &providers)),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRoute53RecordsInZoneIdWithProvider(
						testAccAwsRegionProviderFunc("us-west-2", &providers), &zone, 100),
					testAccCreateRandomRoute53RecordsInZoneIdWithProvider(
						testAccAwsRegionProviderFunc("us-west-2", &providers), &zone, 5),

					testAccCheckRoute53ZoneExistsWithProvider("aws_route53_zone.with_trailing_dot", &zoneWithDot,
						testAccAwsRegionProviderFunc("us-west-2", &providers)),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRoute53RecordsInZoneIdWithProvider(
						testAccAwsRegionProviderFunc("us-west-2", &providers), &zoneWithDot, 100),
					testAccCreateRandomRoute53RecordsInZoneIdWithProvider(
						testAccAwsRegionProviderFunc("us-west-2", &providers), &zoneWithDot, 5),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_updateComment(t *testing.T) {
	var zone route53.GetHostedZoneOutput
	var td route53.ResourceTagSet

	rString := acctest.RandString(8)
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route53_zone.main",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53ZoneConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.main", &zone),
					testAccLoadTagsR53(&zone, &td),
					testAccCheckTagsR53(&td.Tags, "foo", "bar"),
					resource.TestCheckResourceAttr(
						"aws_route53_zone.main", "comment", "Custom comment"),
				),
			},

			resource.TestStep{
				Config: testAccRoute53ZoneConfigUpdateComment(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.main", &zone),
					testAccLoadTagsR53(&zone, &td),
					resource.TestCheckResourceAttr(
						"aws_route53_zone.main", "comment", "Change Custom Comment"),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_private_basic(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route53_zone.main",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53PrivateZoneConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.main", &zone),
					testAccCheckRoute53ZoneAssociatesWithVpc("aws_vpc.main", &zone),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_private_region(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IDRefreshName:     "aws_route53_zone.main",
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckRoute53ZoneDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53PrivateZoneRegionConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExistsWithProvider("aws_route53_zone.main", &zone,
						testAccAwsRegionProviderFunc("us-west-2", &providers)),
					testAccCheckRoute53ZoneAssociatesWithVpc("aws_vpc.main", &zone),
				),
			},
		},
	})
}

func testAccCheckRoute53ZoneDestroy(s *terraform.State) error {
	return testAccCheckRoute53ZoneDestroyWithProvider(s, testAccProvider)
}

func testAccCheckRoute53ZoneDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_zone" {
			continue
		}

		_, err := conn.GetHostedZone(&route53.GetHostedZoneInput{Id: aws.String(rs.Primary.ID)})
		if err == nil {
			return fmt.Errorf("Hosted zone still exists")
		}
	}
	return nil
}

func testAccCreateRandomRoute53RecordsInZoneIdWithProvider(providerF func() *schema.Provider, zone *route53.GetHostedZoneOutput, recordsCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider := providerF()
		conn := provider.Meta().(*AWSClient).r53conn

		var changes []*route53.Change
		if recordsCount > 100 {
			return fmt.Errorf("Route53 API only allows 100 record sets in a single batch")
		}
		for i := 0; i < recordsCount; i++ {
			changes = append(changes, &route53.Change{
				Action: aws.String("UPSERT"),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(fmt.Sprintf("%d-tf-acc-random.%s", acctest.RandInt(), *zone.HostedZone.Name)),
					Type: aws.String("CNAME"),
					ResourceRecords: []*route53.ResourceRecord{
						&route53.ResourceRecord{Value: aws.String(fmt.Sprintf("random.%s", *zone.HostedZone.Name))},
					},
					TTL: aws.Int64(int64(30)),
				},
			})
		}

		req := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: zone.HostedZone.Id,
			ChangeBatch: &route53.ChangeBatch{
				Comment: aws.String("Generated by Terraform"),
				Changes: changes,
			},
		}
		log.Printf("[DEBUG] Change set: %s\n", *req)
		resp, err := changeRoute53RecordSet(conn, req)
		if err != nil {
			return err
		}
		changeInfo := resp.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo
		err = waitForRoute53RecordSetToSync(conn, cleanChangeID(*changeInfo.Id))
		return err
	}
}

func testAccCheckRoute53ZoneExists(n string, zone *route53.GetHostedZoneOutput) resource.TestCheckFunc {
	return testAccCheckRoute53ZoneExistsWithProvider(n, zone, func() *schema.Provider { return testAccProvider })
}

func testAccCheckRoute53ZoneExistsWithProvider(n string, zone *route53.GetHostedZoneOutput, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No hosted zone ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).r53conn
		resp, err := conn.GetHostedZone(&route53.GetHostedZoneInput{Id: aws.String(rs.Primary.ID)})
		if err != nil {
			return fmt.Errorf("Hosted zone err: %v", err)
		}

		aws_comment := *resp.HostedZone.Config.Comment
		rs_comment := rs.Primary.Attributes["comment"]
		if rs_comment != "" && rs_comment != aws_comment {
			return fmt.Errorf("Hosted zone with comment '%s' found but does not match '%s'", aws_comment, rs_comment)
		}

		if !*resp.HostedZone.Config.PrivateZone {
			sorted_ns := make([]string, len(resp.DelegationSet.NameServers))
			for i, ns := range resp.DelegationSet.NameServers {
				sorted_ns[i] = *ns
			}
			sort.Strings(sorted_ns)
			for idx, ns := range sorted_ns {
				attribute := fmt.Sprintf("name_servers.%d", idx)
				dsns := rs.Primary.Attributes[attribute]
				if dsns != ns {
					return fmt.Errorf("Got: %v for %v, Expected: %v", dsns, attribute, ns)
				}
			}
		}

		*zone = *resp
		return nil
	}
}

func testAccCheckRoute53ZoneAssociatesWithVpc(n string, zone *route53.GetHostedZoneOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		var associatedVPC *route53.VPC
		for _, vpc := range zone.VPCs {
			if *vpc.VPCId == rs.Primary.ID {
				associatedVPC = vpc
			}
		}
		if associatedVPC == nil {
			return fmt.Errorf("VPC: %v is not associated to Zone: %v", n, cleanZoneID(*zone.HostedZone.Id))
		}
		return nil
	}
}

func testAccLoadTagsR53(zone *route53.GetHostedZoneOutput, td *route53.ResourceTagSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn

		zone := cleanZoneID(*zone.HostedZone.Id)
		req := &route53.ListTagsForResourceInput{
			ResourceId:   aws.String(zone),
			ResourceType: aws.String("hostedzone"),
		}

		resp, err := conn.ListTagsForResource(req)
		if err != nil {
			return err
		}

		if resp.ResourceTagSet != nil {
			*td = *resp.ResourceTagSet
		}

		return nil
	}
}
func testAccCheckDomainName(zone *route53.GetHostedZoneOutput, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if zone.HostedZone.Name == nil {
			return fmt.Errorf("Empty name in HostedZone for domain %s", domain)
		}

		if *zone.HostedZone.Name == domain {
			return nil
		}

		return fmt.Errorf("Invalid domain name. Expected %s is %s", domain, *zone.HostedZone.Name)
	}
}
func testAccRoute53ZoneConfig(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
	name = "%s."
	comment = "Custom comment"

	tags {
		foo = "bar"
		Name = "tf-route53-tag-test"
	}
}
`, zoneName)
}

func testAccRoute53ZoneCountConfig() string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
	name = "subdomain${count.index}.terraformtest.com"

	count = 5
}
`)
}

func testAccRoute53ZoneConfig_forceDestroy(zoneName1, zoneName2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "destroyable" {
	name = "%s"
	force_destroy = true
}

resource "aws_route53_zone" "with_trailing_dot" {
	name = "%s."
	force_destroy = true
}
`, zoneName1, zoneName2)
}

func testAccRoute53ZoneConfigUpdateComment(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
	name = "%s."
	comment = "Change Custom Comment"

	tags {
		foo = "bar"
		Name = "tf-route53-tag-test"
	}
}
`, zoneName)
}

func testAccRoute53PrivateZoneConfig(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "main" {
	cidr_block = "172.29.0.0/24"
	instance_tenancy = "default"
	enable_dns_support = true
	enable_dns_hostnames = true
	tags {
		Name = "terraform-testacc-route53-zone-private"
	}
}

resource "aws_route53_zone" "main" {
	name = "%s."
	vpc_id = "${aws_vpc.main.id}"
}
`, zoneName)
}

func testAccRoute53PrivateZoneRegionConfig(zoneName string) string {
	return fmt.Sprintf(`
provider "aws" {
	alias = "west"
	region = "us-west-2"
}

provider "aws" {
	alias = "east"
	region = "us-east-1"
}

resource "aws_vpc" "main" {
	provider = "aws.east"
	cidr_block = "172.29.0.0/24"
	instance_tenancy = "default"
	enable_dns_support = true
	enable_dns_hostnames = true
	tags {
		Name = "terraform-testacc-route53-zone-private-region"
	}
}

resource "aws_route53_zone" "main" {
	provider = "aws.west"
	name = "%s."
	vpc_id = "${aws_vpc.main.id}"
	vpc_region = "us-east-1"
}
`, zoneName)
}
