package aws

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// add sweeper to delete known test sgs
func init() {
	resource.AddTestSweepers("aws_security_group", &resource.Sweeper{
		Name: "aws_security_group",
		Dependencies: []string{
			"aws_subnet",
		},
		F: testSweepSecurityGroups,
	})
}

func testSweepSecurityGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	req := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("tf-acc-revoke*"),
					aws.String("tf-acc-test-*"),
				},
			},
		},
	}
	resp, err := conn.DescribeSecurityGroups(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Security Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving EC2 Security Groups: %s", err)
	}

	if len(resp.SecurityGroups) == 0 {
		log.Print("[DEBUG] No aws security groups to sweep")
		return nil
	}

	for _, sg := range resp.SecurityGroups {
		// revoke the rules
		if sg.IpPermissions != nil {
			req := &ec2.RevokeSecurityGroupIngressInput{
				GroupId:       sg.GroupId,
				IpPermissions: sg.IpPermissions,
			}

			if _, err = conn.RevokeSecurityGroupIngress(req); err != nil {
				return fmt.Errorf(
					"Error revoking default egress rule for Security Group (%s): %s",
					*sg.GroupId, err)
			}
		}

		if sg.IpPermissionsEgress != nil {
			req := &ec2.RevokeSecurityGroupEgressInput{
				GroupId:       sg.GroupId,
				IpPermissions: sg.IpPermissionsEgress,
			}

			if _, err = conn.RevokeSecurityGroupEgress(req); err != nil {
				return fmt.Errorf(
					"Error revoking default egress rule for Security Group (%s): %s",
					*sg.GroupId, err)
			}
		}
	}

	for _, sg := range resp.SecurityGroups {
		// delete the group
		_, err := conn.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: sg.GroupId,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting Security Group (%s): %s",
				*sg.GroupId, err)
		}
	}

	return nil
}

func TestProtocolStateFunc(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected string
	}{
		{
			input:    "tcp",
			expected: "tcp",
		},
		{
			input:    6,
			expected: "",
		},
		{
			input:    "17",
			expected: "udp",
		},
		{
			input:    "all",
			expected: "-1",
		},
		{
			input:    "-1",
			expected: "-1",
		},
		{
			input:    -1,
			expected: "",
		},
		{
			input:    "1",
			expected: "icmp",
		},
		{
			input:    "icmp",
			expected: "icmp",
		},
		{
			input:    1,
			expected: "",
		},
	}
	for _, c := range cases {
		result := protocolStateFunc(c.input)
		if result != c.expected {
			t.Errorf("Error matching protocol, expected (%s), got (%s)", c.expected, result)
		}
	}
}

func TestProtocolForValue(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "tcp",
			expected: "tcp",
		},
		{
			input:    "6",
			expected: "tcp",
		},
		{
			input:    "udp",
			expected: "udp",
		},
		{
			input:    "17",
			expected: "udp",
		},
		{
			input:    "all",
			expected: "-1",
		},
		{
			input:    "-1",
			expected: "-1",
		},
		{
			input:    "tCp",
			expected: "tcp",
		},
		{
			input:    "6",
			expected: "tcp",
		},
		{
			input:    "UDp",
			expected: "udp",
		},
		{
			input:    "17",
			expected: "udp",
		},
		{
			input:    "ALL",
			expected: "-1",
		},
		{
			input:    "icMp",
			expected: "icmp",
		},
		{
			input:    "1",
			expected: "icmp",
		},
	}

	for _, c := range cases {
		result := protocolForValue(c.input)
		if result != c.expected {
			t.Errorf("Error matching protocol, expected (%s), got (%s)", c.expected, result)
		}
	}
}

func calcSecurityGroupChecksum(rules []interface{}) int {
	var sum int = 0
	for _, rule := range rules {
		sum += resourceAwsSecurityGroupRuleHash(rule)
	}
	return sum
}

func TestResourceAwsSecurityGroupExpandCollapseRules(t *testing.T) {
	expected_compact_list := []interface{}{
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with description",
			"self":        true,
			"cidr_blocks": []interface{}{
				"10.0.0.1/32",
				"10.0.0.2/32",
				"10.0.0.3/32",
			},
		},
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with another description",
			"self":        false,
			"cidr_blocks": []interface{}{
				"192.168.0.1/32",
				"192.168.0.2/32",
			},
		},
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   int(8000),
			"to_port":     int(8080),
			"description": "",
			"self":        false,
			"ipv6_cidr_blocks": []interface{}{
				"fd00::1/128",
				"fd00::2/128",
			},
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"sg-11111",
				"sg-22222",
				"sg-33333",
			}),
		},
		map[string]interface{}{
			"protocol":    "udp",
			"from_port":   int(10000),
			"to_port":     int(10000),
			"description": "",
			"self":        false,
			"prefix_list_ids": []interface{}{
				"pl-111111",
				"pl-222222",
			},
		},
	}

	expected_expanded_list := []interface{}{
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with description",
			"self":        true,
		},
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with description",
			"self":        false,
			"cidr_blocks": []interface{}{
				"10.0.0.1/32",
			},
		},
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with description",
			"self":        false,
			"cidr_blocks": []interface{}{
				"10.0.0.2/32",
			},
		},
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with description",
			"self":        false,
			"cidr_blocks": []interface{}{
				"10.0.0.3/32",
			},
		},
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with another description",
			"self":        false,
			"cidr_blocks": []interface{}{
				"192.168.0.1/32",
			},
		},
		map[string]interface{}{
			"protocol":    "tcp",
			"from_port":   int(443),
			"to_port":     int(443),
			"description": "block with another description",
			"self":        false,
			"cidr_blocks": []interface{}{
				"192.168.0.2/32",
			},
		},
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   int(8000),
			"to_port":     int(8080),
			"description": "",
			"self":        false,
			"ipv6_cidr_blocks": []interface{}{
				"fd00::1/128",
			},
		},
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   int(8000),
			"to_port":     int(8080),
			"description": "",
			"self":        false,
			"ipv6_cidr_blocks": []interface{}{
				"fd00::2/128",
			},
		},
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   int(8000),
			"to_port":     int(8080),
			"description": "",
			"self":        false,
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"sg-11111",
			}),
		},
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   int(8000),
			"to_port":     int(8080),
			"description": "",
			"self":        false,
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"sg-22222",
			}),
		},
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   int(8000),
			"to_port":     int(8080),
			"description": "",
			"self":        false,
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"sg-33333",
			}),
		},
		map[string]interface{}{
			"protocol":    "udp",
			"from_port":   int(10000),
			"to_port":     int(10000),
			"description": "",
			"self":        false,
			"prefix_list_ids": []interface{}{
				"pl-111111",
			},
		},
		map[string]interface{}{
			"protocol":    "udp",
			"from_port":   int(10000),
			"to_port":     int(10000),
			"description": "",
			"self":        false,
			"prefix_list_ids": []interface{}{
				"pl-222222",
			},
		},
	}

	expected_compact_set := schema.NewSet(resourceAwsSecurityGroupRuleHash, expected_compact_list)
	actual_expanded_list := resourceAwsSecurityGroupExpandRules(expected_compact_set).List()

	if calcSecurityGroupChecksum(expected_expanded_list) != calcSecurityGroupChecksum(actual_expanded_list) {
		t.Fatalf("error matching expanded set for resourceAwsSecurityGroupExpandRules()")
	}

	actual_collapsed_list := resourceAwsSecurityGroupCollapseRules("ingress", expected_expanded_list)

	if calcSecurityGroupChecksum(expected_compact_list) != calcSecurityGroupChecksum(actual_collapsed_list) {
		t.Fatalf("error matching collapsed set for resourceAwsSecurityGroupCollapseRules()")
	}
}

func TestResourceAwsSecurityGroupIPPermGather(t *testing.T) {
	raw := []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(1)),
			ToPort:     aws.Int64(int64(-1)),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupId:     aws.String("sg-11111"),
					Description: aws.String("desc"),
				},
			},
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(80)),
			ToPort:     aws.Int64(int64(80)),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				// VPC
				{
					GroupId: aws.String("sg-22222"),
				},
			},
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(443)),
			ToPort:     aws.Int64(int64(443)),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				// Classic
				{
					UserId:    aws.String("12345"),
					GroupId:   aws.String("sg-33333"),
					GroupName: aws.String("ec2_classic"),
				},
				{
					UserId:    aws.String("amazon-elb"),
					GroupId:   aws.String("sg-d2c979d3"),
					GroupName: aws.String("amazon-elb-sg"),
				},
			},
		},
		{
			IpProtocol: aws.String("-1"),
			FromPort:   aws.Int64(int64(0)),
			ToPort:     aws.Int64(int64(0)),
			PrefixListIds: []*ec2.PrefixListId{
				{
					PrefixListId: aws.String("pl-12345678"),
					Description:  aws.String("desc"),
				},
			},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				// VPC
				{
					GroupId: aws.String("sg-22222"),
				},
			},
		},
	}

	local := []map[string]interface{}{
		{
			"protocol":    "tcp",
			"from_port":   int64(1),
			"to_port":     int64(-1),
			"cidr_blocks": []string{"0.0.0.0/0"},
			"self":        true,
			"description": "desc",
		},
		{
			"protocol":  "tcp",
			"from_port": int64(80),
			"to_port":   int64(80),
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"sg-22222",
			}),
		},
		{
			"protocol":  "tcp",
			"from_port": int64(443),
			"to_port":   int64(443),
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"ec2_classic",
				"amazon-elb/amazon-elb-sg",
			}),
		},
		{
			"protocol":        "-1",
			"from_port":       int64(0),
			"to_port":         int64(0),
			"prefix_list_ids": []string{"pl-12345678"},
			"security_groups": schema.NewSet(schema.HashString, []interface{}{
				"sg-22222",
			}),
			"description": "desc",
		},
	}

	out := resourceAwsSecurityGroupIPPermGather("sg-11111", raw, aws.String("12345"))
	for _, i := range out {
		// loop and match rules, because the ordering is not guarneteed
		for _, l := range local {
			if i["from_port"] == l["from_port"] {

				if i["to_port"] != l["to_port"] {
					t.Fatalf("to_port does not match")
				}

				if _, ok := i["cidr_blocks"]; ok {
					if !reflect.DeepEqual(i["cidr_blocks"], l["cidr_blocks"]) {
						t.Fatalf("error matching cidr_blocks")
					}
				}

				if _, ok := i["security_groups"]; ok {
					outSet := i["security_groups"].(*schema.Set)
					localSet := l["security_groups"].(*schema.Set)

					if !outSet.Equal(localSet) {
						t.Fatalf("Security Group sets are not equal")
					}
				}
			}
		}
	}
}

func TestAccAWSSecurityGroup_importBasic(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 2: group, 2 rules
		if len(s) != 2 {
			return fmt.Errorf("expected 2 states: %#v", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig,
			},

			{
				ResourceName:            "aws_security_group.web",
				ImportState:             true,
				ImportStateCheck:        checkFn,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_importIpv6(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 3: group, 2 rules
		if len(s) != 3 {
			return fmt.Errorf("expected 3 states: %#v", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigIpv6,
			},

			{
				ResourceName:     "aws_security_group.web",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSSecurityGroup_importSelf(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_importSelf,
			},

			{
				ResourceName:            "aws_security_group.allow_all",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_importSourceSecurityGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_importSourceSecurityGroup,
			},

			{
				ResourceName:            "aws_security_group.test_group_1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_importIPRangeAndSecurityGroupWithSameRules(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 4: group, 3 rules
		if len(s) != 4 {
			return fmt.Errorf("expected 4 states: %#v", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_importIPRangeAndSecurityGroupWithSameRules,
			},

			{
				ResourceName:     "aws_security_group.test_group_1",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSSecurityGroup_importIPRangesWithSameRules(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 4: group, 2 rules
		if len(s) != 3 {
			return fmt.Errorf("expected 3 states: %#v", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_importIPRangesWithSameRules,
			},

			{
				ResourceName:     "aws_security_group.test_group_1",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSSecurityGroup_importPrefixList(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 2: group, 1 rule
		if len(s) != 2 {
			return fmt.Errorf("expected 2 states: %#v", s)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigPrefixListEgress,
			},

			{
				ResourceName:     "aws_security_group.egress",
				ImportState:      true,
				ImportStateCheck: checkFn,
			},
		},
	})
}

func TestAccAWSSecurityGroup_basic(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributes(&group),
					resource.TestMatchResourceAttr("aws_security_group.web", "arn", regexp.MustCompile(`^arn:[^:]+:ec2:[^:]+:[^:]+:security-group/.+$`)),
					resource.TestCheckResourceAttr("aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr("aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.to_port", "8000"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleGathering(t *testing.T) {
	var group ec2.SecurityGroup
	sgName := fmt.Sprintf("tf-acc-security-group-%s", acctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_ruleGathering(sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					resource.TestCheckResourceAttr("aws_security_group.test", "name", sgName),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.#", "3"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.description", "egress for all ipv6"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.from_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.protocol", "-1"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.2760422146.to_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.description", "egress for all ipv4"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.from_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.protocol", "-1"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.test", "egress.3161496341.to_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.#", "5"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.cidr_blocks.0", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.description", "ingress from 192.168.0.0/16"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1274017860.to_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.description", "ingress from all ipv6"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1396402051.to_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.cidr_blocks.#", "2"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.cidr_blocks.0", "10.0.2.0/24"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.cidr_blocks.1", "10.0.3.0/24"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.description", "ingress from 10.0.0.0/16"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.1889111182.to_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.cidr_blocks.#", "2"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.cidr_blocks.0", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.cidr_blocks.1", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.self", "true"),
					resource.TestCheckResourceAttr("aws_security_group.test", "ingress.2038285407.to_port", "80"),
				),
			},
		},
	})
}

// cycleIpPermForGroup returns an IpPermission struct with a configured
// UserIdGroupPair for the groupid given. Used in
// TestAccAWSSecurityGroup_forceRevokeRules_should_fail to create a cyclic rule
// between 2 security groups
func cycleIpPermForGroup(groupId string) *ec2.IpPermission {
	var perm ec2.IpPermission
	perm.FromPort = aws.Int64(0)
	perm.ToPort = aws.Int64(0)
	perm.IpProtocol = aws.String("icmp")
	perm.UserIdGroupPairs = make([]*ec2.UserIdGroupPair, 1)
	perm.UserIdGroupPairs[0] = &ec2.UserIdGroupPair{
		GroupId: aws.String(groupId),
	}
	return &perm
}

// testAddRuleCycle returns a TestCheckFunc to use at the end of a test, such
// that a Security Group Rule cyclic dependency will be created between the two
// Security Groups. A companion function, testRemoveRuleCycle, will undo this.
func testAddRuleCycle(primary, secondary *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if primary.GroupId == nil {
			return fmt.Errorf("Primary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}
		if secondary.GroupId == nil {
			return fmt.Errorf("Secondary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		// cycle from primary to secondary
		perm1 := cycleIpPermForGroup(*secondary.GroupId)
		// cycle from secondary to primary
		perm2 := cycleIpPermForGroup(*primary.GroupId)

		req1 := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       primary.GroupId,
			IpPermissions: []*ec2.IpPermission{perm1},
		}
		req2 := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       secondary.GroupId,
			IpPermissions: []*ec2.IpPermission{perm2},
		}

		var err error
		_, err = conn.AuthorizeSecurityGroupEgress(req1)
		if err != nil {
			return fmt.Errorf(
				"Error authorizing primary security group %s rules: %s", *primary.GroupId,
				err)
		}
		_, err = conn.AuthorizeSecurityGroupEgress(req2)
		if err != nil {
			return fmt.Errorf(
				"Error authorizing secondary security group %s rules: %s", *secondary.GroupId,
				err)
		}
		return nil
	}
}

// testRemoveRuleCycle removes the cyclic dependency between two security groups
// that was added in testAddRuleCycle
func testRemoveRuleCycle(primary, secondary *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if primary.GroupId == nil {
			return fmt.Errorf("Primary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}
		if secondary.GroupId == nil {
			return fmt.Errorf("Secondary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		for _, sg := range []*ec2.SecurityGroup{primary, secondary} {
			var err error
			if sg.IpPermissions != nil {
				req := &ec2.RevokeSecurityGroupIngressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissions,
				}

				if _, err = conn.RevokeSecurityGroupIngress(req); err != nil {
					return fmt.Errorf(
						"Error revoking default ingress rule for Security Group in testRemoveCycle (%s): %s",
						*primary.GroupId, err)
				}
			}

			if sg.IpPermissionsEgress != nil {
				req := &ec2.RevokeSecurityGroupEgressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissionsEgress,
				}

				if _, err = conn.RevokeSecurityGroupEgress(req); err != nil {
					return fmt.Errorf(
						"Error revoking default egress rule for Security Group in testRemoveCycle (%s): %s",
						*sg.GroupId, err)
				}
			}
		}
		return nil
	}
}

// This test should fail to destroy the Security Groups and VPC, due to a
// dependency cycle added outside of terraform's management. There is a sweeper
// 'aws_vpc' and 'aws_security_group' that cleans these up, however, the test is
// written to allow Terraform to clean it up because we do go and revoke the
// cyclic rules that were added.
func TestAccAWSSecurityGroup_forceRevokeRules_true(t *testing.T) {
	var primary ec2.SecurityGroup
	var secondary ec2.SecurityGroup

	// Add rules to create a cycle between primary and secondary. This prevents
	// Terraform/AWS from being able to destroy the groups
	testAddCycle := testAddRuleCycle(&primary, &secondary)
	// Remove the rules that created the cycle; Terraform/AWS can now destroy them
	testRemoveCycle := testRemoveRuleCycle(&primary, &secondary)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create the configuration with 2 security groups, then create a
			// dependency cycle such that they cannot be deleted
			{
				Config: testAccAWSSecurityGroupConfig_revoke_base,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.primary", &primary),
					testAccCheckAWSSecurityGroupExists("aws_security_group.secondary", &secondary),
					testAddCycle,
				),
			},
			// Verify the DependencyViolation error by using a configuration with the
			// groups removed. Terraform tries to destroy them but cannot. Expect a
			// DependencyViolation error
			{
				Config:      testAccAWSSecurityGroupConfig_revoke_base_removed,
				ExpectError: regexp.MustCompile("DependencyViolation"),
			},
			// Restore the config (a no-op plan) but also remove the dependencies
			// between the groups with testRemoveCycle
			{
				Config: testAccAWSSecurityGroupConfig_revoke_base,
				// ExpectError: regexp.MustCompile("DependencyViolation"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.primary", &primary),
					testAccCheckAWSSecurityGroupExists("aws_security_group.secondary", &secondary),
					testRemoveCycle,
				),
			},
			// Again try to apply the config with the sgs removed; it should work
			{
				Config: testAccAWSSecurityGroupConfig_revoke_base_removed,
			},
			////
			// now test with revoke_rules_on_delete
			////
			// create the configuration with 2 security groups, then create a
			// dependency cycle such that they cannot be deleted. In this
			// configuration, each Security Group has `revoke_rules_on_delete`
			// specified, and should delete with no issue
			{
				Config: testAccAWSSecurityGroupConfig_revoke_true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.primary", &primary),
					testAccCheckAWSSecurityGroupExists("aws_security_group.secondary", &secondary),
					testAddCycle,
				),
			},
			// Again try to apply the config with the sgs removed; it should work,
			// because we've told the SGs to forcefully revoke their rules first
			{
				Config: testAccAWSSecurityGroupConfig_revoke_base_removed,
			},
		},
	})
}

func TestAccAWSSecurityGroup_forceRevokeRules_false(t *testing.T) {
	var primary ec2.SecurityGroup
	var secondary ec2.SecurityGroup

	// Add rules to create a cycle between primary and secondary. This prevents
	// Terraform/AWS from being able to destroy the groups
	testAddCycle := testAddRuleCycle(&primary, &secondary)
	// Remove the rules that created the cycle; Terraform/AWS can now destroy them
	testRemoveCycle := testRemoveRuleCycle(&primary, &secondary)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create the configuration with 2 security groups, then create a
			// dependency cycle such that they cannot be deleted. These Security
			// Groups are configured to explicitly not revoke rules on delete,
			// `revoke_rules_on_delete = false`
			{
				Config: testAccAWSSecurityGroupConfig_revoke_false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.primary", &primary),
					testAccCheckAWSSecurityGroupExists("aws_security_group.secondary", &secondary),
					testAddCycle,
				),
			},
			// Verify the DependencyViolation error by using a configuration with the
			// groups removed, and the Groups not configured to revoke their ruls.
			// Terraform tries to destroy them but cannot. Expect a
			// DependencyViolation error
			{
				Config:      testAccAWSSecurityGroupConfig_revoke_base_removed,
				ExpectError: regexp.MustCompile("DependencyViolation"),
			},
			// Restore the config (a no-op plan) but also remove the dependencies
			// between the groups with testRemoveCycle
			{
				Config: testAccAWSSecurityGroupConfig_revoke_false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.primary", &primary),
					testAccCheckAWSSecurityGroupExists("aws_security_group.secondary", &secondary),
					testRemoveCycle,
				),
			},
			// Again try to apply the config with the sgs removed; it should work
			{
				Config: testAccAWSSecurityGroupConfig_revoke_base_removed,
			},
		},
	})
}

func TestAccAWSSecurityGroup_ipv6(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr("aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2293451516.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.2293451516.to_port", "8000"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_namePrefix(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_security_group.baz",
		IDRefreshIgnore: []string{"name_prefix"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupPrefixNameConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.baz", &group),
					testAccCheckAWSSecurityGroupGeneratedNamePrefix(
						"aws_security_group.baz", "baz-"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_self(t *testing.T) {
	var group ec2.SecurityGroup

	checkSelf := func(s *terraform.State) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("bad: %#v", group)
			}
		}()

		if *group.IpPermissions[0].UserIdGroupPairs[0].GroupId != *group.GroupId {
			return fmt.Errorf("bad: %#v", group)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigSelf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr("aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3971148406.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3971148406.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3971148406.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3971148406.self", "true"),
					checkSelf,
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_vpc(t *testing.T) {
	var group ec2.SecurityGroup

	testCheck := func(*terraform.State) error {
		if *group.VpcId == "" {
			return fmt.Errorf("should have vpc ID")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigVpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributes(&group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_vpcNegOneIngress(t *testing.T) {
	var group ec2.SecurityGroup

	testCheck := func(*terraform.State) error {
		if *group.VpcId == "" {
			return fmt.Errorf("should have vpc ID")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigVpcNegOneIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributesNegOneProtocol(&group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.956249133.protocol", "-1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.956249133.from_port", "0"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.956249133.to_port", "0"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.956249133.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.956249133.cidr_blocks.0", "10.0.0.0/8"),
					testCheck,
				),
			},
		},
	})
}
func TestAccAWSSecurityGroup_vpcProtoNumIngress(t *testing.T) {
	var group ec2.SecurityGroup

	testCheck := func(*terraform.State) error {
		if *group.VpcId == "" {
			return fmt.Errorf("should have vpc ID")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigVpcProtoNumIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.2449525218.protocol", "50"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.2449525218.from_port", "0"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.2449525218.to_port", "0"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.2449525218.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.2449525218.cidr_blocks.0", "10.0.0.0/8"),
					testCheck,
				),
			},
		},
	})
}
func TestAccAWSSecurityGroup_MultiIngress(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigMultiIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_Change(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
				),
			},
			{
				Config: testAccAWSSecurityGroupConfigChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributesChanged(&group),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_RuleDescription(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigRuleDescription("Egress description", "Ingress description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.description", "Egress description"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.2129912301.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.description", "Ingress description"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1147649399.to_port", "8000"),
				),
			},
			// Change just the rule descriptions.
			{
				Config: testAccAWSSecurityGroupConfigRuleDescription("New egress description", "New ingress description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.description", "New egress description"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.746197026.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.description", "New ingress description"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.1341057959.to_port", "8000"),
				),
			},
			// Remove just the rule descriptions.
			{
				Config: testAccAWSSecurityGroupConfigEmptyRuleDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.to_port", "8000"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_generatedName(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Managed by Terraform"),
					func(s *terraform.State) error {
						if group.GroupName == nil {
							return fmt.Errorf("bad: No SG name")
						}
						if !strings.HasPrefix(*group.GroupName, "terraform-") {
							return fmt.Errorf("No terraform- prefix: %s", *group.GroupName)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_DefaultEgress_VPC(t *testing.T) {

	// VPC
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_security_group.worker",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigDefaultEgress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExistsWithoutDefault("aws_security_group.worker"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_DefaultEgress_Classic(t *testing.T) {
	var group ec2.SecurityGroup

	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		IDRefreshName: "aws_security_group.web",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigClassic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
				),
			},
		},
	})
}

// Testing drift detection with groups containing the same port and types
func TestAccAWSSecurityGroup_drift(t *testing.T) {
	var group ec2.SecurityGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_drift(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "2"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.cidr_blocks.0", "206.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.to_port", "8000"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_drift_complex(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_drift_complex(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr("aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "3"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.cidr_blocks.0", "206.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.657243763.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "3"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.cidr_blocks.0", "206.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.657243763.to_port", "8000"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_invalidCIDRBlock(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSecurityGroupInvalidIngressCidr,
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccAWSSecurityGroupInvalidEgressCidr,
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccAWSSecurityGroupInvalidIPv6IngressCidr,
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
			{
				Config:      testAccAWSSecurityGroupInvalidIPv6EgressCidr,
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

func testAccCheckAWSSecurityGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_security_group" {
			continue
		}

		// Retrieve our group
		req := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeSecurityGroups(req)
		if err == nil {
			if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupId == rs.Primary.ID {
				return fmt.Errorf("Security Group (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		// Confirm error code is what we want
		if ec2err.Code() != "InvalidGroup.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSSecurityGroupGeneratedNamePrefix(
	resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

func testAccCheckAWSSecurityGroupExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		req := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeSecurityGroups(req)
		if err != nil {
			return err
		}

		if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupId == rs.Primary.ID {
			*group = *resp.SecurityGroups[0]
			return nil
		}

		return fmt.Errorf("Security Group not found")
	}
}

func testAccCheckAWSSecurityGroupAttributes(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := &ec2.IpPermission{
			FromPort:   aws.Int64(80),
			ToPort:     aws.Int64(8000),
			IpProtocol: aws.String("tcp"),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("10.0.0.0/8")}},
		}

		if *group.GroupName != "terraform_acceptance_test_example" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}

		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}

		if len(group.IpPermissions) == 0 {
			return fmt.Errorf("No IPPerms")
		}

		// Compare our ingress
		if !reflect.DeepEqual(group.IpPermissions[0], p) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IpPermissions[0],
				p)
		}

		return nil
	}
}

func testAccCheckAWSSecurityGroupAttributesNegOneProtocol(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := &ec2.IpPermission{
			IpProtocol: aws.String("-1"),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("10.0.0.0/8")}},
		}

		if *group.GroupName != "terraform_acceptance_test_example" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}

		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}

		if len(group.IpPermissions) == 0 {
			return fmt.Errorf("No IPPerms")
		}

		// Compare our ingress
		if !reflect.DeepEqual(group.IpPermissions[0], p) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IpPermissions[0],
				p)
		}

		return nil
	}
}

func TestAccAWSSecurityGroup_tags(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.foo", &group),
					testAccCheckTags(&group.Tags, "foo", "bar"),
				),
			},

			{
				Config: testAccAWSSecurityGroupConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.foo", &group),
					testAccCheckTags(&group.Tags, "foo", ""),
					testAccCheckTags(&group.Tags, "bar", "baz"),
					testAccCheckTags(&group.Tags, "env", "Production"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_CIDRandGroups(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupCombindCIDRandGroups,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.mixed", &group),
					// testAccCheckAWSSecurityGroupAttributes(&group),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressWithCidrAndSGs(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupSGandCidrAttributes(&group),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "2"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.cidr_blocks.0", "192.168.0.1/32"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.from_port", "22"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.to_port", "22"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressWithCidrAndSGs_classic(t *testing.T) {
	var group ec2.SecurityGroup

	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs_classic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupSGandCidrAttributes(&group),
					resource.TestCheckResourceAttr("aws_security_group.web", "egress.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.#", "2"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.cidr_blocks.0", "192.168.0.1/32"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.from_port", "22"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.protocol", "tcp"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.web", "ingress.3893008652.to_port", "22"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_egressWithPrefixList(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigPrefixListEgress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.egress", &group),
					testAccCheckAWSSecurityGroupEgressPrefixListAttributes(&group),
					resource.TestCheckResourceAttr(
						"aws_security_group.egress", "egress.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressWithPrefixList(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigPrefixListIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.ingress", &group),
					testAccCheckAWSSecurityGroupIngressPrefixListAttributes(&group),
					resource.TestCheckResourceAttr(
						"aws_security_group.ingress", "ingress.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ipv4andipv6Egress(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigIpv4andIpv6Egress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.egress", &group),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.#", "2"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.from_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.protocol", "-1"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.482069346.to_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.description", ""),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.from_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.protocol", "-1"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.security_groups.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.self", "false"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "egress.706749478.to_port", "0"),
					resource.TestCheckResourceAttr("aws_security_group.egress", "ingress.#", "0"),
				),
			},
		},
	})
}

// testAccAWSSecurityGroupRulesPerGroupLimitFromEnv returns security group rules per group limit
// Currently this information is not available from any EC2 or Trusted Advisor API
// Prefers the EC2_SECURITY_GROUP_RULES_PER_GROUP_LIMIT environment variable or defaults to 50
func testAccAWSSecurityGroupRulesPerGroupLimitFromEnv() int {
	const defaultLimit = 50
	const envVar = "EC2_SECURITY_GROUP_RULES_PER_GROUP_LIMIT"

	envLimitStr := os.Getenv(envVar)
	if envLimitStr == "" {
		return defaultLimit
	}
	envLimitInt, err := strconv.Atoi(envLimitStr)
	if err != nil {
		log.Printf("[WARN] Error converting %q environment variable value %q to integer: %s", envVar, envLimitStr, err)
		return defaultLimit
	}
	if envLimitInt <= 50 {
		return defaultLimit
	}
	return envLimitInt
}

func testAccCheckAWSSecurityGroupSGandCidrAttributes(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *group.GroupName != "terraform_acceptance_test_example" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}

		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}

		if len(group.IpPermissions) == 0 {
			return fmt.Errorf("No IPPerms")
		}

		if len(group.IpPermissions) != 2 {
			return fmt.Errorf("Expected 2 ingress rules, got %d", len(group.IpPermissions))
		}

		for _, p := range group.IpPermissions {
			if *p.FromPort == int64(22) {
				if len(p.IpRanges) != 1 || p.UserIdGroupPairs != nil {
					return fmt.Errorf("Found ip perm of 22, but not the right ipranges / pairs: %s", p)
				}
				continue
			} else if *p.FromPort == int64(80) {
				if len(p.IpRanges) != 1 || len(p.UserIdGroupPairs) != 1 {
					return fmt.Errorf("Found ip perm of 80, but not the right ipranges / pairs: %s", p)
				}
				continue
			}
			return fmt.Errorf("Found a rouge rule")
		}

		return nil
	}
}

func testAccCheckAWSSecurityGroupEgressPrefixListAttributes(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *group.GroupName != "terraform_acceptance_test_prefix_list_egress" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}
		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}
		if len(group.IpPermissionsEgress) == 0 {
			return fmt.Errorf("No egress IPPerms")
		}
		if len(group.IpPermissionsEgress) != 1 {
			return fmt.Errorf("Expected 1 egress rule, got %d", len(group.IpPermissions))
		}

		p := group.IpPermissionsEgress[0]

		if len(p.PrefixListIds) != 1 {
			return fmt.Errorf("Expected 1 prefix list, got %d", len(p.PrefixListIds))
		}

		return nil
	}
}

func testAccCheckAWSSecurityGroupIngressPrefixListAttributes(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *group.GroupName != "terraform_acceptance_test_prefix_list_ingress" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}
		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}
		if len(group.IpPermissions) == 0 {
			return fmt.Errorf("No IPPerms")
		}
		if len(group.IpPermissions) != 1 {
			return fmt.Errorf("Expected 1 rule, got %d", len(group.IpPermissions))
		}

		p := group.IpPermissions[0]

		if len(p.PrefixListIds) != 1 {
			return fmt.Errorf("Expected 1 prefix list, got %d", len(p.PrefixListIds))
		}

		return nil
	}
}

func testAccCheckAWSSecurityGroupAttributesChanged(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(80),
				ToPort:     aws.Int64(9000),
				IpProtocol: aws.String("tcp"),
				IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("10.0.0.0/8")}},
			},
			{
				FromPort:   aws.Int64(80),
				ToPort:     aws.Int64(8000),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
					{
						CidrIp: aws.String("10.0.0.0/8"),
					},
				},
			},
		}

		if *group.GroupName != "terraform_acceptance_test_example" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}

		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}

		// Compare our ingress
		if len(group.IpPermissions) != 2 {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IpPermissions,
				p)
		}

		if *group.IpPermissions[0].ToPort == 8000 {
			group.IpPermissions[1], group.IpPermissions[0] =
				group.IpPermissions[0], group.IpPermissions[1]
		}

		if len(group.IpPermissions[1].IpRanges) > 1 {
			if *group.IpPermissions[1].IpRanges[0].CidrIp != "0.0.0.0/0" {
				group.IpPermissions[1].IpRanges[0], group.IpPermissions[1].IpRanges[1] =
					group.IpPermissions[1].IpRanges[1], group.IpPermissions[1].IpRanges[0]
			}
		}

		if !reflect.DeepEqual(group.IpPermissions, p) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IpPermissions,
				p)
		}

		return nil
	}
}

func testAccCheckAWSSecurityGroupExistsWithoutDefault(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		req := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeSecurityGroups(req)
		if err != nil {
			return err
		}

		if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupId == rs.Primary.ID {
			group := *resp.SecurityGroups[0]

			if len(group.IpPermissionsEgress) != 1 {
				return fmt.Errorf("Security Group should have only 1 egress rule, got %d", len(group.IpPermissionsEgress))
			}
		}

		return nil
	}
}

func TestAccAWSSecurityGroup_failWithDiffMismatch(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_failWithDiffMismatch,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.nat", &group),
					resource.TestCheckResourceAttr("aws_security_group.nat", "egress.#", "0"),
					resource.TestCheckResourceAttr("aws_security_group.nat", "ingress.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitExceededAppend(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
			// append a rule to step over the limit
			{
				Config:      testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit+1),
				ExpectError: regexp.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				PreConfig: func() {
					// should have the original rules still
					err := testSecurityGroupRuleCount(*group.GroupId, 0, ruleLimit)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitCidrBlockExceededAppend(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigCidrBlockRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, 1),
				),
			},
			// append a rule to step over the limit
			{
				Config:      testAccAWSSecurityGroupConfigCidrBlockRuleLimit(0, ruleLimit+1),
				ExpectError: regexp.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				PreConfig: func() {
					// should have the original cidr blocks still in 1 rule
					err := testSecurityGroupRuleCount(*group.GroupId, 0, 1)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}

					id := *group.GroupId

					conn := testAccProvider.Meta().(*AWSClient).ec2conn
					req := &ec2.DescribeSecurityGroupsInput{
						GroupIds: []*string{aws.String(id)},
					}
					resp, err := conn.DescribeSecurityGroups(req)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}

					var match *ec2.SecurityGroup
					if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupId == id {
						match = resp.SecurityGroups[0]
					}

					if match == nil {
						t.Fatalf("PreConfig check failed: security group %s not found", id)
					}

					if cidrCount := len(match.IpPermissionsEgress[0].IpRanges); cidrCount != ruleLimit {
						t.Fatalf("PreConfig check failed: rule does not have previous IP ranges, has %d", cidrCount)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigCidrBlockRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, 1),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitExceededPrepend(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
			// prepend a rule to step over the limit
			{
				Config:      testAccAWSSecurityGroupConfigRuleLimit(1, ruleLimit+1),
				ExpectError: regexp.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				PreConfig: func() {
					// should have the original rules still (limit - 1 because of the shift)
					err := testSecurityGroupRuleCount(*group.GroupId, 0, ruleLimit-1)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitExceededAllNew(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
			// add a rule to step over the limit with entirely new rules
			{
				Config:      testAccAWSSecurityGroupConfigRuleLimit(100, ruleLimit+1),
				ExpectError: regexp.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				// all the rules should have been revoked and the add failed
				PreConfig: func() {
					err := testSecurityGroupRuleCount(*group.GroupId, 0, 0)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_rulesDropOnError(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// Create a valid security group with some rules and make sure it exists
			{
				Config: testAccAWSSecurityGroupConfig_rulesDropOnError_Init,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.test", &group),
				),
			},
			// Add a bad rule to trigger API error
			{
				Config:      testAccAWSSecurityGroupConfig_rulesDropOnError_AddBadRule,
				ExpectError: regexp.MustCompile("InvalidGroupId.Malformed"),
			},
			// All originally added rules must survive. This will return non-empty plan if anything changed.
			{
				Config:   testAccAWSSecurityGroupConfig_rulesDropOnError_Init,
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckAWSSecurityGroupRuleCount(group *ec2.SecurityGroup, expectedIngressCount, expectedEgressCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := *group.GroupId
		return testSecurityGroupRuleCount(id, expectedIngressCount, expectedEgressCount)
	}
}

func testSecurityGroupRuleCount(id string, expectedIngressCount, expectedEgressCount int) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn
	req := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(id)},
	}
	resp, err := conn.DescribeSecurityGroups(req)
	if err != nil {
		return err
	}

	var group *ec2.SecurityGroup
	if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupId == id {
		group = resp.SecurityGroups[0]
	}

	if group == nil {
		return fmt.Errorf("Security group %s not found", id)
	}

	if actual := len(group.IpPermissions); actual != expectedIngressCount {
		return fmt.Errorf("Security group ingress rule count %d does not match %d", actual, expectedIngressCount)
	}

	if actual := len(group.IpPermissionsEgress); actual != expectedEgressCount {
		return fmt.Errorf("Security group egress rule count %d does not match %d", actual, expectedEgressCount)
	}

	return nil
}

func testAccAWSSecurityGroupConfigRuleLimit(egressStartIndex, egressRulesCount int) string {
	c := `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-rule-limit"
  }
}

resource "aws_security_group" "test" {
  name = "terraform_acceptance_test_rule_limit"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.test.id}"

	tags = {
    Name = "tf-acc-test"
  }

	// egress rules to exhaust the limit
`

	for i := egressStartIndex; i < egressRulesCount+egressStartIndex; i++ {
		c += fmt.Sprintf(`
  egress {
		protocol = "tcp"
		from_port = "${80 + %[1]d}"
		to_port = "${80 + %[1]d}"
		cidr_blocks = ["${cidrhost("10.1.0.0/16", %[1]d)}/32"]
	}
`, i)
	}

	c += "\n}"

	return c
}

func testAccAWSSecurityGroupConfigCidrBlockRuleLimit(egressStartIndex, egressRulesCount int) string {
	c := `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-rule-limit"
  }
}

resource "aws_security_group" "test" {
  name = "terraform_acceptance_test_rule_limit"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.test.id}"

	tags = {
    Name = "tf-acc-test"
  }

	// egress rules to exhaust the limit
	egress {
		protocol = "tcp"
		from_port = "80"
		to_port = "80"
		cidr_blocks = [
`

	for i := egressStartIndex; i < egressRulesCount+egressStartIndex; i++ {
		c += fmt.Sprintf(`
		"${cidrhost("10.1.0.0/16", %[1]d)}/32",
`, i)
	}

	c += "\n\t\t]\n\t}\n}"

	return c
}

const testAccAWSSecurityGroupConfigEmptyRuleDescription = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-empty-rule-description"
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_desc_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  tags = {
    Name = "tf-acc-test"
  }
}`

const testAccAWSSecurityGroupConfigIpv6 = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-ipv6"
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    ipv6_cidr_blocks = ["::/0"]
  }

	tags = {
		Name = "tf-acc-test"
	}
}
`

const testAccAWSSecurityGroupConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group"
	}
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

	tags = {
		Name = "tf-acc-revoke-test"
	}
}
`

const testAccAWSSecurityGroupConfig_revoke_base_removed = `
resource "aws_vpc" "sg-race-revoke" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-revoke"
	}
}
`
const testAccAWSSecurityGroupConfig_revoke_base = `
resource "aws_vpc" "sg-race-revoke" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-revoke"
	}
}

resource "aws_security_group" "primary" {
  name = "tf-acc-sg-race-revoke-primary"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.sg-race-revoke.id}"

	tags = {
		Name = "tf-acc-revoke-test-primary"
	}
}

resource "aws_security_group" "secondary" {
  name = "tf-acc-sg-race-revoke-secondary"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.sg-race-revoke.id}"

	tags = {
		Name = "tf-acc-revoke-test-secondary"
	}
}
`

const testAccAWSSecurityGroupConfig_revoke_false = `
resource "aws_vpc" "sg-race-revoke" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-revoke"
	}
}

resource "aws_security_group" "primary" {
  name = "tf-acc-sg-race-revoke-primary"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.sg-race-revoke.id}"

	tags = {
		Name = "tf-acc-revoke-test-primary"
	}

  revoke_rules_on_delete = false
}

resource "aws_security_group" "secondary" {
  name = "tf-acc-sg-race-revoke-secondary"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.sg-race-revoke.id}"

	tags = {
		Name = "tf-acc-revoke-test-secondary"
	}

  revoke_rules_on_delete = false
}
`

const testAccAWSSecurityGroupConfig_revoke_true = `
resource "aws_vpc" "sg-race-revoke" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-revoke"
	}
}

resource "aws_security_group" "primary" {
  name = "tf-acc-sg-race-revoke-primary"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.sg-race-revoke.id}"

	tags = {
		Name = "tf-acc-revoke-test-primary"
	}

  revoke_rules_on_delete = true
}

resource "aws_security_group" "secondary" {
  name = "tf-acc-sg-race-revoke-secondary"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.sg-race-revoke.id}"

	tags = {
		Name = "tf-acc-revoke-test-secondary"
	}

  revoke_rules_on_delete = true
}
`

const testAccAWSSecurityGroupConfigChange = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-change"
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 9000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["0.0.0.0/0", "10.0.0.0/8"]
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

func testAccAWSSecurityGroupConfigRuleDescription(egressDescription, ingressDescription string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-description"
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = "%s"
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = "%s"
  }

	tags = {
		Name = "tf-acc-test"
	}
}
`, ingressDescription, egressDescription)
}

const testAccAWSSecurityGroupConfigSelf = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-self"
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    self = true
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

const testAccAWSSecurityGroupConfigVpc = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-vpc"
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

	egress {
		protocol = "tcp"
		from_port = 80
		to_port = 8000
		cidr_blocks = ["10.0.0.0/8"]
	}
}
`

const testAccAWSSecurityGroupConfigVpcNegOneIngress = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-vpc-neg-one-ingress"
	}
}

resource "aws_security_group" "web" {
	name = "terraform_acceptance_test_example"
	description = "Used in the terraform acceptance tests"
	vpc_id = "${aws_vpc.foo.id}"

	ingress {
		protocol = "-1"
		from_port = 0
		to_port = 0
		cidr_blocks = ["10.0.0.0/8"]
	}
}
`

const testAccAWSSecurityGroupConfigVpcProtoNumIngress = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-vpc-proto-num-ingress"
	}
}

resource "aws_security_group" "web" {
	name = "terraform_acceptance_test_example"
	description = "Used in the terraform acceptance tests"
	vpc_id = "${aws_vpc.foo.id}"

	ingress {
		protocol = "50"
		from_port = 0
		to_port = 0
		cidr_blocks = ["10.0.0.0/8"]
	}
}
`

const testAccAWSSecurityGroupConfigMultiIngress = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-multi-ingress"
	}
}

resource "aws_security_group" "worker" {
  name = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "tcp"
    from_port = 22
    to_port = 22
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 800
    to_port = 800
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    security_groups = ["${aws_security_group.worker.id}"]
  }

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

const testAccAWSSecurityGroupConfigTags = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-tags"
	}
}

resource "aws_security_group" "foo" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    foo = "bar"
  }
}
`

const testAccAWSSecurityGroupConfigTagsUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-tags"
	}
}

resource "aws_security_group" "foo" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    bar = "baz"
    env = "Production"
  }
}
`

const testAccAWSSecurityGroupConfig_generatedName = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-generated-name"
	}
}

resource "aws_security_group" "web" {
  vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "tf-acc-test"
	}
}
`

const testAccAWSSecurityGroupConfigDefaultEgress = `
resource "aws_vpc" "tf_sg_egress_test" {
    cidr_block = "10.0.0.0/16"
  tags = {
        Name = "terraform-testacc-security-group-default-egress"
    }
}

resource "aws_security_group" "worker" {
  name = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
        vpc_id = "${aws_vpc.tf_sg_egress_test.id}"

  egress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

const testAccAWSSecurityGroupConfigClassic = `
resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
}
`

const testAccAWSSecurityGroupPrefixNameConfig = `
provider "aws" {
  region = "us-east-1"
}

resource "aws_security_group" "baz" {
   name_prefix = "baz-"
   description = "Used in the terraform acceptance tests"
}
`

func testAccAWSSecurityGroupConfig_drift() string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name = "tf_acc_%d"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["206.0.0.0/8"]
  }

  tags = {
                Name = "tf-acc-test"
        }
}
`, acctest.RandInt())
}

func testAccAWSSecurityGroupConfig_drift_complex() string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-drift-complex"
	}
}

resource "aws_security_group" "otherweb" {
  name        = "tf_acc_%d"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_security_group" "web" {
  name        = "tf_acc_%d"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["206.0.0.0/8"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    security_groups = ["${aws_security_group.otherweb.id}"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["206.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    security_groups = ["${aws_security_group.otherweb.id}"]
  }

  tags = {
    Name = "tf-acc-test"
  }
}`, acctest.RandInt(), acctest.RandInt())
}

const testAccAWSSecurityGroupInvalidIngressCidr = `
resource "aws_security_group" "foo" {
  name = "testing-foo"
  description = "foo-testing"
  ingress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["1.2.3.4/33"]
  }
}`

const testAccAWSSecurityGroupInvalidEgressCidr = `
resource "aws_security_group" "foo" {
  name = "testing-foo"
  description = "foo-testing"
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["1.2.3.4/33"]
  }
}`

const testAccAWSSecurityGroupInvalidIPv6IngressCidr = `
resource "aws_security_group" "foo" {
  name = "testing-foo"
  description = "foo-testing"
  ingress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    ipv6_cidr_blocks = ["::/244"]
  }
}`

const testAccAWSSecurityGroupInvalidIPv6EgressCidr = `
resource "aws_security_group" "foo" {
  name = "testing-foo"
  description = "foo-testing"
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    ipv6_cidr_blocks = ["::/244"]
  }
}`

const testAccAWSSecurityGroupCombindCIDRandGroups = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-combine-rand-groups"
	}
}

resource "aws_security_group" "two" {
	name = "tf-test-1"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-test-1"
	}
}

resource "aws_security_group" "one" {
	name = "tf-test-2"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-test-w"
	}
}

resource "aws_security_group" "three" {
	name = "tf-test-3"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-test-3"
	}
}

resource "aws_security_group" "mixed" {
  name = "tf-mix-test"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16", "10.1.0.0/16", "10.7.0.0/16"]

    security_groups = [
      "${aws_security_group.one.id}",
      "${aws_security_group.two.id}",
      "${aws_security_group.three.id}",
    ]
  }

  tags = {
    Name = "tf-mix-test"
  }
}
`

const testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-security-group-ingress-w-cidr-and-sg"
	}
}

resource "aws_security_group" "other_web" {
  name        = "tf_other_acc_tests"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group" "web" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol  = "tcp"
    from_port = "22"
    to_port   = "22"

    cidr_blocks = [
      "192.168.0.1/32",
    ]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 8000
    cidr_blocks     = ["10.0.0.0/8"]
    security_groups = ["${aws_security_group.other_web.id}"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`

const testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs_classic = `
resource "aws_security_group" "other_web" {
  name        = "tf_other_acc_tests"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group" "web" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol  = "tcp"
    from_port = "22"
    to_port   = "22"

    cidr_blocks = [
      "192.168.0.1/32",
    ]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 8000
    cidr_blocks     = ["10.0.0.0/8"]
    security_groups = ["${aws_security_group.other_web.name}"]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`

// fails to apply in one pass with the error "diffs didn't match during apply"
// GH-2027
const testAccAWSSecurityGroupConfig_failWithDiffMismatch = `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-fail-w-diff-mismatch"
  }
}

resource "aws_security_group" "ssh_base" {
  name   = "test-ssh-base"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_security_group" "jump" {
  name   = "test-jump"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_security_group" "provision" {
  name   = "test-provision"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_security_group" "nat" {
  vpc_id      = "${aws_vpc.main.id}"
  name        = "nat"
  description = "For nat servers "

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = ["${aws_security_group.jump.id}"]
  }

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = ["${aws_security_group.provision.id}"]
  }
}
`
const testAccAWSSecurityGroupConfig_importSelf = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-import-self"
  }
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group_rule" "allow_all" {
  type        = "ingress"
  from_port   = 0
  to_port     = 65535
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.allow_all.id}"
}

resource "aws_security_group_rule" "allow_all-1" {
  type      = "ingress"
  from_port = 65534
  to_port   = 65535
  protocol  = "tcp"

  self              = true
  security_group_id = "${aws_security_group.allow_all.id}"
}
`

const testAccAWSSecurityGroupConfig_importSourceSecurityGroup = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-import-source-sg"
  }
}

resource "aws_security_group" "test_group_1" {
  name        = "test group 1"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group" "test_group_2" {
  name        = "test group 2"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group" "test_group_3" {
  name        = "test group 3"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group_rule" "allow_test_group_2" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = "${aws_security_group.test_group_1.id}"
  security_group_id = "${aws_security_group.test_group_2.id}"
}

resource "aws_security_group_rule" "allow_test_group_3" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = "${aws_security_group.test_group_1.id}"
  security_group_id = "${aws_security_group.test_group_3.id}"
}
`

const testAccAWSSecurityGroupConfig_importIPRangeAndSecurityGroupWithSameRules = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-import-ip-range-and-sg"
  }
}

resource "aws_security_group" "test_group_1" {
  name        = "test group 1"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group" "test_group_2" {
  name        = "test group 2"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group_rule" "allow_security_group" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = "${aws_security_group.test_group_2.id}"
  security_group_id = "${aws_security_group.test_group_1.id}"
}

resource "aws_security_group_rule" "allow_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  cidr_blocks = ["10.0.0.0/32"]
  security_group_id = "${aws_security_group.test_group_1.id}"
}

resource "aws_security_group_rule" "allow_ipv6_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  ipv6_cidr_blocks = ["::/0"]
  security_group_id = "${aws_security_group.test_group_1.id}"
}
`

const testAccAWSSecurityGroupConfig_importIPRangesWithSameRules = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-import-ip-ranges"
  }
}

resource "aws_security_group" "test_group_1" {
  name        = "test group 1"
  vpc_id      = "${aws_vpc.foo.id}"
}

resource "aws_security_group_rule" "allow_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  cidr_blocks = ["10.0.0.0/32"]
  security_group_id = "${aws_security_group.test_group_1.id}"
}

resource "aws_security_group_rule" "allow_ipv6_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  ipv6_cidr_blocks = ["::/0"]
  security_group_id = "${aws_security_group.test_group_1.id}"
}
`

const testAccAWSSecurityGroupConfigIpv4andIpv6Egress = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
  tags = {
      Name = "terraform-testacc-security-group-ipv4-and-ipv6-egress"
  }
}

resource "aws_security_group" "egress" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"
  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks  = ["0.0.0.0/0"]
  }
  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    ipv6_cidr_blocks  = ["::/0"]
  }
}
`

const testAccAWSSecurityGroupConfigPrefixListEgress = `
data "aws_region" "current" {}

resource "aws_vpc" "tf_sg_prefix_list_egress_test" {
    cidr_block = "10.0.0.0/16"
  tags = {
        Name = "terraform-testacc-security-group-prefix-list-egress"
    }
}

resource "aws_route_table" "default" {
    vpc_id = "${aws_vpc.tf_sg_prefix_list_egress_test.id}"
}

resource "aws_vpc_endpoint" "test" {
  	vpc_id = "${aws_vpc.tf_sg_prefix_list_egress_test.id}"
  	service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
  	route_table_ids = ["${aws_route_table.default.id}"]
  	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid":"AllowAll",
			"Effect":"Allow",
			"Principal":"*",
			"Action":"*",
			"Resource":"*"
		}
	]
}
POLICY
}

resource "aws_security_group" "egress" {
    name = "terraform_acceptance_test_prefix_list_egress"
    description = "Used in the terraform acceptance tests"
    vpc_id = "${aws_vpc.tf_sg_prefix_list_egress_test.id}"

    egress {
      protocol = "-1"
      from_port = 0
      to_port = 0
      prefix_list_ids = ["${aws_vpc_endpoint.test.prefix_list_id}"]
    }
}
`

const testAccAWSSecurityGroupConfigPrefixListIngress = `
data "aws_region" "current" {}

resource "aws_vpc" "tf_sg_prefix_list_ingress_test" {
    cidr_block = "10.0.0.0/16"
  tags = {
        Name = "terraform-testacc-security-group-prefix-list-ingress"
    }
}

resource "aws_route_table" "default" {
    vpc_id = "${aws_vpc.tf_sg_prefix_list_ingress_test.id}"
}

resource "aws_vpc_endpoint" "test" {
    vpc_id = "${aws_vpc.tf_sg_prefix_list_ingress_test.id}"
    service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
    route_table_ids = ["${aws_route_table.default.id}"]
    policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid":"AllowAll",
            "Effect":"Allow",
            "Principal":"*",
            "Action":"*",
            "Resource":"*"
        }
    ]
}
POLICY
}

resource "aws_security_group" "ingress" {
    name = "terraform_acceptance_test_prefix_list_ingress"
    description = "Used in the terraform acceptance tests"
    vpc_id = "${aws_vpc.tf_sg_prefix_list_ingress_test.id}"

    ingress {
      protocol = "-1"
      from_port = 0
      to_port = 0
      prefix_list_ids = ["${aws_vpc_endpoint.test.prefix_list_id}"]
    }
}
`

func testAccAWSSecurityGroupConfig_ruleGathering(sgName string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "${var.name}"
  }
}

resource "aws_route_table" "default" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = "${aws_vpc.test.id}"
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = ["${aws_route_table.default.id}"]

  policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid":"AllowAll",
			"Effect":"Allow",
			"Principal":"*",
			"Action":"*",
			"Resource":"*"
		}
	]
}
POLICY
}

resource "aws_security_group" "source1" {
  name        = "${var.name}-source1"
  description = "terraform acceptance test for security group as source1"
  vpc_id      = "${aws_vpc.test.id}"
}

resource "aws_security_group" "source2" {
  name        = "${var.name}-source2"
  description = "terraform acceptance test for security group as source2"
  vpc_id      = "${aws_vpc.test.id}"
}

resource "aws_security_group" "test" {
  name        = "${var.name}"
  description = "terraform acceptance test for security group"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["10.0.0.0/24", "10.0.1.0/24"]
    self        = true
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24"]
    description = "ingress from 10.0.0.0/16"
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["192.168.0.0/16"]
    description = "ingress from 192.168.0.0/16"
  }

  ingress {
    protocol         = "tcp"
    from_port        = 80
    to_port          = 80
    ipv6_cidr_blocks = ["::/0"]
    description      = "ingress from all ipv6"
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
    security_groups = ["${aws_security_group.source1.id}", "${aws_security_group.source2.id}"]
    description     = "ingress from other security groups"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "egress for all ipv4"
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/0"]
    description      = "egress for all ipv6"
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    prefix_list_ids = ["${aws_vpc_endpoint.test.prefix_list_id}"]
    description     = "egress for vpc endpoints"
  }
}
`, sgName)
}

const testAccAWSSecurityGroupConfig_rulesDropOnError_Init = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-drop-rules-test"
  }
}

resource "aws_security_group" "test_ref0" {
  name = "terraform_acceptance_test_drop_rules_ref0"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_security_group" "test_ref1" {
  name = "terraform_acceptance_test_drop_rules_ref1"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_security_group" "test" {
  name = "terraform_acceptance_test_drop_rules"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test"
  }

  ingress {
    protocol = "tcp"
    from_port = "80"
    to_port = "80"
    security_groups = [
      "${aws_security_group.test_ref0.id}",
      "${aws_security_group.test_ref1.id}",
    ]
  }
}
`

const testAccAWSSecurityGroupConfig_rulesDropOnError_AddBadRule = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-security-group-drop-rules-test"
  }
}

resource "aws_security_group" "test_ref0" {
  name = "terraform_acceptance_test_drop_rules_ref0"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_security_group" "test_ref1" {
  name = "terraform_acceptance_test_drop_rules_ref1"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_security_group" "test" {
  name = "terraform_acceptance_test_drop_rules"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test"
  }

  ingress {
    protocol = "tcp"
    from_port = "80"
    to_port = "80"
    security_groups = [
      "${aws_security_group.test_ref0.id}",
      "${aws_security_group.test_ref1.id}",
      "sg-malformed", # non-existent rule to trigger API error
    ]
  }
}
`
