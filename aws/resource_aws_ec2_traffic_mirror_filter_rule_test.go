package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEc2TrafficMirrorFilterRule_basic(t *testing.T) {
	resourceName := "aws_ec2_traffic_mirror_filter_rule.rule"
	dstCidr := "10.0.0.0/8"
	srcCidr := "0.0.0.0/0"
	ruleNum := 1
	action := "accept"
	direction := "ingress"
	description := "test rule"
	protocol := 6
	srcPortFrom := 32000
	srcPortTo := 64000
	dstPortFrom := 10000
	dstPortTo := 10001

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorFilterRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorFilterRuleDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccEc2TrafficMirrorFilterRuleConfig(dstCidr, srcCidr, action, direction, ruleNum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterRuleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexp.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckNoResourceAttr(resourceName, "destination_port_range"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "source_port_range"),
				),
			},
			// Add all optionals
			{
				Config: testAccEc2TrafficMirrorFilterRuleConfigFull(dstCidr, srcCidr, action, direction, description, ruleNum, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterRuleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexp.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.from_port", strconv.Itoa(dstPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.0.to_port", strconv.Itoa(dstPortTo)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.from_port", strconv.Itoa(srcPortFrom)),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.0.to_port", strconv.Itoa(srcPortTo)),
					resource.TestCheckResourceAttr(resourceName, "protocol", strconv.Itoa(protocol)),
				),
			},
			// remove optionals
			{
				Config: testAccEc2TrafficMirrorFilterRuleConfig(dstCidr, srcCidr, action, direction, ruleNum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterRuleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "traffic_mirror_filter_id", regexp.MustCompile("tmf-.*")),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", dstCidr),
					resource.TestCheckResourceAttr(resourceName, "rule_action", action),
					resource.TestCheckResourceAttr(resourceName, "rule_number", strconv.Itoa(ruleNum)),
					resource.TestCheckResourceAttr(resourceName, "source_cidr_block", srcCidr),
					resource.TestCheckResourceAttr(resourceName, "traffic_direction", direction),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_port_range.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_port_range.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEc2TrafficMirrorFilterRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSEc2TrafficMirrorFilterRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		ruleId := rs.Primary.ID
		filterId := rs.Primary.Attributes["traffic_mirror_filter_id"]

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(filterId),
			},
		})

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return fmt.Errorf("Traffic mirror filter %s not found", rs.Primary.ID)
		}

		filter := out.TrafficMirrorFilters[0]
		var ruleList []*ec2.TrafficMirrorFilterRule
		ruleList = append(ruleList, filter.IngressFilterRules...)
		ruleList = append(ruleList, filter.EgressFilterRules...)

		var exists bool
		for _, rule := range ruleList {
			if *rule.TrafficMirrorFilterRuleId == ruleId {
				exists = true
				break
			}
		}

		if !exists {
			return fmt.Errorf("Rule %s not found inside filter %s", ruleId, filterId)
		}

		return nil
	}
}

func testAccEc2TrafficMirrorFilterRuleConfig(dstCidr, srcCidr, action, dir string, num int) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "filter" {
}

resource "aws_ec2_traffic_mirror_filter_rule" "rule" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  destination_cidr_block   = "%s"
  rule_action              = "%s"
  rule_number              = %d
  source_cidr_block        = "%s"
  traffic_direction        = "%s"
}
`, dstCidr, action, num, srcCidr, dir)
}

func testAccEc2TrafficMirrorFilterRuleConfigFull(dstCidr, srcCidr, action, dir, description string, ruleNum, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo, protocol int) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "filter" {
}

resource "aws_ec2_traffic_mirror_filter_rule" "rule" {
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  destination_cidr_block   = "%s"
  rule_action              = "%s"
  rule_number              = %d
  source_cidr_block        = "%s"
  traffic_direction        = "%s"
  description              = "%s"
  protocol                 = %d
  source_port_range {
    from_port = %d
    to_port   = %d
  }
  destination_port_range {
    from_port = %d
    to_port   = %d
  }
}
`, dstCidr, action, ruleNum, srcCidr, dir, description, protocol, srcPortFrom, srcPortTo, dstPortFrom, dstPortTo)
}

func testAccPreCheckAWSEc2TrafficMirrorFilterRule(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	_, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{})

	if testAccPreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter rule acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckAWSEc2TrafficMirrorFilterRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_filter_rule" {
			continue
		}

		ruleId := rs.Primary.ID
		filterId := rs.Primary.Attributes["traffic_mirror_filter_id"]

		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(filterId),
			},
		})

		if isAWSErr(err, "InvalidTrafficMirrorFilterId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return nil
		}

		filter := out.TrafficMirrorFilters[0]
		var ruleList []*ec2.TrafficMirrorFilterRule
		ruleList = append(ruleList, filter.IngressFilterRules...)
		ruleList = append(ruleList, filter.EgressFilterRules...)

		for _, rule := range ruleList {
			if *rule.TrafficMirrorFilterRuleId == ruleId {
				return fmt.Errorf("Rule %s still exists in filter %s", ruleId, filterId)
			}
		}
	}

	return nil
}

func testAccAWSEc2TrafficMirrorFilterRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["traffic_mirror_filter_id"], rs.Primary.ID), nil
	}
}
