package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestAccAWSRoute53DelegationSet_basic(t *testing.T) {
	rString := acctest.RandString(8)
	refName := fmt.Sprintf("tf_acc_%s", rString)
	resourceName := "aws_route53_delegation_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"reference_name"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckRoute53DelegationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DelegationSetConfig(refName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53DelegationSetExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}

func TestAccAWSRoute53DelegationSet_withZones(t *testing.T) {
	var zone route53.GetHostedZoneOutput

	rString := acctest.RandString(8)
	refName := fmt.Sprintf("tf_acc_%s", rString)
	resourceName := "aws_route53_delegation_set.test"
	zoneName1 := fmt.Sprintf("%s-primary.terraformtest.com", rString)
	zoneName2 := fmt.Sprintf("%s-secondary.terraformtest.com", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"reference_name"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckRoute53DelegationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53DelegationSetWithZonesConfig(refName, zoneName1, zoneName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53DelegationSetExists(resourceName),
					testAccCheckRoute53ZoneExists("aws_route53_zone.primary", &zone),
					testAccCheckRoute53ZoneExists("aws_route53_zone.secondary", &zone),
					testAccCheckRoute53NameServersMatch(resourceName, "aws_route53_zone.primary"),
					testAccCheckRoute53NameServersMatch(resourceName, "aws_route53_zone.secondary"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name"},
			},
		},
	})
}

func testAccCheckRoute53DelegationSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_delegation_set" {
			continue
		}

		_, err := conn.GetReusableDelegationSet(&route53.GetReusableDelegationSetInput{Id: aws.String(rs.Primary.ID)})
		if err == nil {
			return fmt.Errorf("Delegation set still exists")
		}
	}
	return nil
}

func testAccCheckRoute53DelegationSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No delegation set ID is set")
		}

		out, err := conn.GetReusableDelegationSet(&route53.GetReusableDelegationSetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Delegation set does not exist: %#v", rs.Primary.ID)
		}

		setID := cleanDelegationSetId(*out.DelegationSet.Id)
		if setID != rs.Primary.ID {
			return fmt.Errorf("Delegation set ID does not match:\nExpected: %#v\nReturned: %#v", rs.Primary.ID, setID)
		}

		return nil
	}
}

func testAccCheckRoute53NameServersMatch(delegationSetName, zoneName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn

		delegationSetLocal, ok := s.RootModule().Resources[delegationSetName]
		if !ok {
			return fmt.Errorf("Not found: %s", delegationSetName)
		}
		delegationSet, err := conn.GetReusableDelegationSet(&route53.GetReusableDelegationSetInput{
			Id: aws.String(delegationSetLocal.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Delegation set does not exist: %#v", delegationSetLocal.Primary.ID)
		}

		hostedZoneLocal, ok := s.RootModule().Resources[zoneName]
		if !ok {
			return fmt.Errorf("Not found: %s", zoneName)
		}
		hostedZone, err := conn.GetHostedZone(&route53.GetHostedZoneInput{
			Id: aws.String(hostedZoneLocal.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Delegation set does not exist: %#v", hostedZoneLocal.Primary.ID)
		}

		if !reflect.DeepEqual(delegationSet.DelegationSet.NameServers, hostedZone.DelegationSet.NameServers) {
			return fmt.Errorf("Name servers do not match:\nDelegation Set: %#v\nHosted Zone:%#v",
				delegationSet.DelegationSet.NameServers, hostedZone.DelegationSet.NameServers)
		}

		return nil
	}
}

func testAccRoute53DelegationSetConfig(refName string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {
	reference_name = "%s"
}
`, refName)
}

func testAccRoute53DelegationSetWithZonesConfig(refName, zoneName1, zoneName2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_delegation_set" "test" {
    reference_name = "%s"
}

resource "aws_route53_zone" "primary" {
    name = "%s"
    delegation_set_id = "${aws_route53_delegation_set.test.id}"
}

resource "aws_route53_zone" "secondary" {
    name = "%s"
    delegation_set_id = "${aws_route53_delegation_set.test.id}"
}
`, refName, zoneName1, zoneName2)
}
