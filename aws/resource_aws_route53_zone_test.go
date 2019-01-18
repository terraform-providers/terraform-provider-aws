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
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	resourceName := "aws_route53_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfig(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s.", zoneName)),
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_multiple(t *testing.T) {
	var zone0, zone1, zone2, zone3, zone4 route53.GetHostedZoneOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigMultiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.test.0", &zone0),
					testAccCheckDomainName(&zone0, "subdomain0.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.test.1", &zone1),
					testAccCheckDomainName(&zone1, "subdomain1.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.test.2", &zone2),
					testAccCheckDomainName(&zone2, "subdomain2.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.test.3", &zone3),
					testAccCheckDomainName(&zone3, "subdomain3.terraformtest.com."),
					testAccCheckRoute53ZoneExists("aws_route53_zone.test.4", &zone4),
					testAccCheckDomainName(&zone4, "subdomain4.terraformtest.com."),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_Comment(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	resourceName := "aws_route53_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigComment(zoneName, "comment1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "comment", "comment1"),
				),
			},
			{
				Config: testAccRoute53ZoneConfigComment(zoneName, "comment2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "comment", "comment2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_DelegationSetID(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	delegationSetResourceName := "aws_route53_delegation_set.test"
	resourceName := "aws_route53_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigDelegationSetID(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttrPair(resourceName, "delegation_set_id", delegationSetResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_ForceDestroy(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	resourceName := "aws_route53_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigForceDestroy(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRoute53RecordsInZoneId(&zone, 100),
					testAccCreateRandomRoute53RecordsInZoneId(&zone, 5),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_ForceDestroy_TrailingPeriod(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	resourceName := "aws_route53_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigForceDestroyTrailingPeriod(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					// Add >100 records to verify pagination works ok
					testAccCreateRandomRoute53RecordsInZoneId(&zone, 100),
					testAccCreateRandomRoute53RecordsInZoneId(&zone, 5),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_Tags(t *testing.T) {
	var zone route53.GetHostedZoneOutput
	var td route53.ResourceTagSet

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_zone.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigTagsSingle(zoneName, "tag1key", "tag1value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1key", "tag1value"),
					testAccLoadTagsR53(&zone, &td),
					testAccCheckTagsR53(&td.Tags, "tag1key", "tag1value"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccRoute53ZoneConfigTagsMultiple(zoneName, "tag1key", "tag1valueupdated", "tag2key", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1key", "tag1valueupdated"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2key", "tag2value"),
					testAccLoadTagsR53(&zone, &td),
					testAccCheckTagsR53(&td.Tags, "tag1key", "tag1valueupdated"),
					testAccCheckTagsR53(&td.Tags, "tag2key", "tag2value"),
				),
			},
			{
				Config: testAccRoute53ZoneConfigTagsSingle(zoneName, "tag2key", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2key", "tag2value"),
					testAccLoadTagsR53(&zone, &td),
					testAccCheckTagsR53(&td.Tags, "tag2key", "tag2value"),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_VPC_Single(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_zone.test"
	vpcResourceName := "aws_vpc.test1"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigVPCSingle(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_VPC_Multiple(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_zone.test"
	vpcResourceName1 := "aws_vpc.test1"
	vpcResourceName2 := "aws_vpc.test2"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigVPCMultiple(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "2"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName1, &zone),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName2, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_VPC_Updates(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_zone.test"
	vpcResourceName1 := "aws_vpc.test1"
	vpcResourceName2 := "aws_vpc.test2"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigVPCSingle(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName1, &zone),
				),
			},
			{
				Config: testAccRoute53ZoneConfigVPCMultiple(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "2"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName1, &zone),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName2, &zone),
				),
			},
			{
				Config: testAccRoute53ZoneConfigVPCSingle(rName, zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName1, &zone),
				),
			},
		},
	})
}

func TestAccAWSRoute53Zone_VPCID(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	resourceName := "aws_route53_zone.test"
	vpcResourceName := "aws_vpc.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigVPCID(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(resourceName, &zone),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName, &zone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_VPCRegion(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	resourceName := "aws_route53_zone.test"
	vpcResourceName := "aws_vpc.test"
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckRoute53ZoneDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneConfigVPCRegion(zoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExistsWithProvider(resourceName, &zone,
						testAccAwsRegionProviderFunc("us-west-2", &providers)),
					resource.TestCheckResourceAttr(resourceName, "vpc.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					testAccCheckRoute53ZoneAssociatesWithVpc(vpcResourceName, &zone),
				),
			},
			{
				// Config must be provided for aliased provider
				Config:                  testAccRoute53ZoneConfigVPCRegion(zoneName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
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

func testAccCreateRandomRoute53RecordsInZoneId(zone *route53.GetHostedZoneOutput, recordsCount int) resource.TestCheckFunc {
	return testAccCreateRandomRoute53RecordsInZoneIdWithProvider(func() *schema.Provider { return testAccProvider }, zone, recordsCount)
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
						{Value: aws.String(fmt.Sprintf("random.%s", *zone.HostedZone.Name))},
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

		for _, vpc := range zone.VPCs {
			if aws.StringValue(vpc.VPCId) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("VPC: %s is not associated to Zone: %v", n, cleanZoneID(aws.StringValue(zone.HostedZone.Id)))
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
resource "aws_route53_zone" "test" {
  name = "%s."
}
`, zoneName)
}

func testAccRoute53ZoneConfigMultiple() string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  count = 5

  name = "subdomain${count.index}.terraformtest.com"
}
`)
}

func testAccRoute53ZoneConfigComment(zoneName, comment string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  comment = %q
  name    = "%s."
}
`, comment, zoneName)
}

func testAccRoute53ZoneConfigDelegationSetID(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {}

resource "aws_route53_zone" "test" {
  delegation_set_id = "${aws_route53_delegation_set.test.id}"
  name              = "%s."
}
`, zoneName)
}

func testAccRoute53ZoneConfigForceDestroy(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "%s"
}
`, zoneName)
}

func testAccRoute53ZoneConfigForceDestroyTrailingPeriod(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  force_destroy = true
  name          = "%s."
}
`, zoneName)
}

func testAccRoute53ZoneConfigTagsSingle(zoneName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%s."

  tags = {
    %q = %q
  }
}
`, zoneName, tag1Key, tag1Value)
}

func testAccRoute53ZoneConfigTagsMultiple(zoneName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "%s."

  tags = {
    %q = %q
    %q = %q
  }
}
`, zoneName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccRoute53ZoneConfigVPCID(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.29.0.0/24"

  tags = {
    Name = "terraform-testacc-route53-zone-private"
  }
}

resource "aws_route53_zone" "test" {
  name   = "%s."
  vpc_id = "${aws_vpc.test.id}"
}
`, zoneName)
}

func testAccRoute53ZoneConfigVPCRegion(zoneName string) string {
	return fmt.Sprintf(`
provider "aws" {
  alias  = "west"
  region = "us-west-2"
}

provider "aws" {
  alias  = "east"
  region = "us-east-1"
}

resource "aws_vpc" "test" {
  provider = "aws.east"

  cidr_block = "172.29.0.0/24"

  tags = {
    Name = "terraform-testacc-route53-zone-private-region"
  }
}

resource "aws_route53_zone" "test" {
  provider = "aws.west"

  name       = "%s."
  vpc_id     = "${aws_vpc.test.id}"
  vpc_region = "us-east-1"
}
`, zoneName)
}

func testAccRoute53ZoneConfigVPCSingle(rName, zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_route53_zone" "test" {
  name = "%s."

  vpc {
    vpc_id = "${aws_vpc.test1.id}"
  }
}
`, rName, zoneName)
}

func testAccRoute53ZoneConfigVPCMultiple(rName, zoneName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_route53_zone" "test" {
  name = "%s."

  vpc {
    vpc_id = "${aws_vpc.test1.id}"
  }

  vpc {
    vpc_id = "${aws_vpc.test2.id}"
  }
}
`, rName, rName, zoneName)
}
