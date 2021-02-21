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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
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

	input := &ec2.DescribeSecurityGroupsInput{}

	// Delete all non-default EC2 Security Group Rules to prevent DependencyViolation errors
	err = conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, sg := range page.SecurityGroups {
			if aws.StringValue(sg.GroupName) == "default" {
				log.Printf("[DEBUG] Skipping default EC2 Security Group: %s", aws.StringValue(sg.GroupId))
				continue
			}

			if sg.IpPermissions != nil {
				req := &ec2.RevokeSecurityGroupIngressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissions,
				}

				if _, err = conn.RevokeSecurityGroupIngress(req); err != nil {
					log.Printf("[ERROR] Error revoking ingress rule for Security Group (%s): %s", aws.StringValue(sg.GroupId), err)
				}
			}

			if sg.IpPermissionsEgress != nil {
				req := &ec2.RevokeSecurityGroupEgressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissionsEgress,
				}

				if _, err = conn.RevokeSecurityGroupEgress(req); err != nil {
					log.Printf("[ERROR] Error revoking egress rule for Security Group (%s): %s", aws.StringValue(sg.GroupId), err)
				}
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Security Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 Security Groups: %s", err)
	}

	err = conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, sg := range page.SecurityGroups {
			if aws.StringValue(sg.GroupName) == "default" {
				log.Printf("[DEBUG] Skipping default EC2 Security Group: %s", aws.StringValue(sg.GroupId))
				continue
			}

			input := &ec2.DeleteSecurityGroupInput{
				GroupId: sg.GroupId,
			}

			// Handle EC2 eventual consistency
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteSecurityGroup(input)

				if isAWSErr(err, "DependencyViolation", "") {
					return resource.RetryableError(err)
				}
				if err != nil {
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if err != nil {
				log.Printf("[ERROR] Error deleting Security Group (%s): %s", aws.StringValue(sg.GroupId), err)
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 Security Groups: %s", err)
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

func TestAccAWSSecurityGroup_allowAll(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_allowAll,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_sourceSecurityGroup(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_sourceSecurityGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_IPRangeAndSecurityGroupWithSameRules(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_IPRangeAndSecurityGroupWithSameRules,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_IPRangesWithSameRules(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_IPRangesWithSameRules,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_basic(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupAttributes(&group),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`security-group/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_egressConfigMode(t *testing.T) {
	var securityGroup1, securityGroup2, securityGroup3 ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigEgressConfigModeBlocks(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &securityGroup1),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccAWSSecurityGroupConfigEgressConfigModeNoBlocks(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &securityGroup2),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "2"),
				),
			},
			{
				Config: testAccAWSSecurityGroupConfigEgressConfigModeZeroed(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &securityGroup3),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressConfigMode(t *testing.T) {
	var securityGroup1, securityGroup2, securityGroup3 ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigIngressConfigModeBlocks(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &securityGroup1),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccAWSSecurityGroupConfigIngressConfigModeNoBlocks(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &securityGroup2),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
				),
			},
			{
				Config: testAccAWSSecurityGroupConfigIngressConfigModeZeroed(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &securityGroup3),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleGathering(t *testing.T) {
	var group ec2.SecurityGroup
	sgName := fmt.Sprintf("tf-acc-security-group-%s", acctest.RandString(7))
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_ruleGathering(sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", sgName),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "0",
						"description":        "egress for all ipv6",
						"from_port":          "0",
						"ipv6_cidr_blocks.#": "1",
						"ipv6_cidr_blocks.0": "::/0",
						"prefix_list_ids.#":  "0",
						"protocol":           "-1",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "0.0.0.0/0",
						"description":        "egress for all ipv4",
						"from_port":          "0",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "-1",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "192.168.0.0/16",
						"description":        "ingress from 192.168.0.0/16",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "0",
						"description":        "ingress from all ipv6",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "1",
						"ipv6_cidr_blocks.0": "::/0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "2",
						"cidr_blocks.0":      "10.0.2.0/24",
						"cidr_blocks.1":      "10.0.3.0/24",
						"description":        "ingress from 10.0.0.0/16",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "2",
						"cidr_blocks.0":      "10.0.0.0/24",
						"cidr_blocks.1":      "10.0.1.0/24",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "true",
						"to_port":            "80",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// This test should fail to destroy the Security Groups and VPC, due to a
// dependency cycle added outside of terraform's management. There is a sweeper
// 'aws_vpc' and 'aws_security_group' that cleans these up, however, the test is
// written to allow Terraform to clean it up because we do go and revoke the
// cyclic rules that were added.
func TestAccAWSSecurityGroup_forceRevokeRulesTrue(t *testing.T) {
	var primary ec2.SecurityGroup
	var secondary ec2.SecurityGroup
	resourceName := "aws_security_group.primary"
	resourceName2 := "aws_security_group.secondary"

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
					testAccCheckAWSSecurityGroupExists(resourceName, &primary),
					testAccCheckAWSSecurityGroupExists(resourceName2, &secondary),
					testAddCycle,
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
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
					testAccCheckAWSSecurityGroupExists(resourceName, &primary),
					testAccCheckAWSSecurityGroupExists(resourceName2, &secondary),
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
					testAccCheckAWSSecurityGroupExists(resourceName, &primary),
					testAccCheckAWSSecurityGroupExists(resourceName2, &secondary),
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

func TestAccAWSSecurityGroup_forceRevokeRulesFalse(t *testing.T) {
	var primary ec2.SecurityGroup
	var secondary ec2.SecurityGroup
	resourceName := "aws_security_group.primary"
	resourceName2 := "aws_security_group.secondary"

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
					testAccCheckAWSSecurityGroupExists(resourceName, &primary),
					testAccCheckAWSSecurityGroupExists(resourceName2, &secondary),
					testAddCycle,
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
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
					testAccCheckAWSSecurityGroupExists(resourceName, &primary),
					testAccCheckAWSSecurityGroupExists(resourceName2, &secondary),
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
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "0",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "1",
						"ipv6_cidr_blocks.0": "::/0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "0",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "1",
						"ipv6_cidr_blocks.0": "::/0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_Name_Generated(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
func TestAccAWSSecurityGroup_Name_TerraformPrefix(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigName("terraform-test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-test"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_NamePrefix(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigNamePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
func TestAccAWSSecurityGroup_NamePrefix_TerraformPrefix(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigNamePrefix("terraform-test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "terraform-test"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_self(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

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
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigSelf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":  "tcp",
						"from_port": "80",
						"to_port":   "8000",
						"self":      "true",
					}),
					checkSelf,
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_vpc(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigVpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupAttributes(&group),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					testAccAWSSecurityGroupCheckVPCIDExists(&group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_vpcNegOneIngress(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigVpcNegOneIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupAttributesNegOneProtocol(&group),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "-1",
						"from_port":     "0",
						"to_port":       "0",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					testAccAWSSecurityGroupCheckVPCIDExists(&group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}
func TestAccAWSSecurityGroup_vpcProtoNumIngress(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigVpcProtoNumIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "50",
						"from_port":     "0",
						"to_port":       "0",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					testAccAWSSecurityGroupCheckVPCIDExists(&group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}
func TestAccAWSSecurityGroup_multiIngress(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigMultiIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_change(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccAWSSecurityGroupConfigChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupAttributesChanged(&group),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleDescription(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigRuleDescription("Egress description", "Ingress description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "Egress description",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "Ingress description",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			// Change just the rule descriptions.
			{
				Config: testAccAWSSecurityGroupConfigRuleDescription("New egress description", "New ingress description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "New egress description",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "New ingress description",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
				),
			},
			// Remove just the rule descriptions.
			{
				Config: testAccAWSSecurityGroupConfigEmptyRuleDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":     "1",
						"cidr_blocks.0":     "10.0.0.0/8",
						"description":       "",
						"from_port":         "80",
						"protocol":          "tcp",
						"security_groups.#": "0",
						"self":              "false",
						"to_port":           "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":     "1",
						"cidr_blocks.0":     "10.0.0.0/8",
						"description":       "",
						"from_port":         "80",
						"protocol":          "tcp",
						"security_groups.#": "0",
						"self":              "false",
						"to_port":           "8000",
					}),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_defaultEgressVPC(t *testing.T) {
	resourceName := "aws_security_group.test"

	// VPC
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigDefaultEgress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExistsWithoutDefault(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_defaultEgressClassic(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSSecurityGroupEc2ClassicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigClassic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupEc2ClassicExists(resourceName, &group),
				),
			},
			{
				Config:                  testAccAWSSecurityGroupConfigClassic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// Testing drift detection with groups containing the same port and types
func TestAccAWSSecurityGroup_drift(t *testing.T) {
	resourceName := "aws_security_group.test"
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_drift(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "206.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// In rules with cidr_block drift, import only creates a single ingress
				// rule with the cidr_blocks de-normalized. During subsequent apply, its
				// normalized to create the 2 ingress rules seen in checks above.
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "ingress", "egress"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_driftComplex(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_drift_complex(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "206.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "206.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// In rules with cidr_block drift, import only creates a single ingress
				// rule with the cidr_blocks de-normalized. During subsequent apply, its
				// normalized to create the 2 ingress rules seen in checks above.
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "ingress", "egress"},
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

func TestAccAWSSecurityGroup_tags(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAWSSecurityGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccAWSAWSSecurityGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAWSSecurityGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_CIDRandGroups(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupCombindCIDRandGroups,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					// testAccCheckAWSSecurityGroupAttributes(&group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressWithCidrAndSGsVPC(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupSGandCidrAttributes(&group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "10.0.0.0/8",
						"description":        "",
						"from_port":          "80",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "192.168.0.1/32",
						"description":        "",
						"from_port":          "22",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "22",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressWithCidrAndSGsClassic(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSSecurityGroupEc2ClassicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs_classic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupEc2ClassicExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "192.168.0.1/32",
						"description":        "",
						"from_port":          "22",
						"ipv6_cidr_blocks.#": "0",
						"protocol":           "tcp",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "22",
					}),
				),
			},
			{
				Config:                  testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs_classic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_egressWithPrefixList(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigPrefixListEgress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupEgressPrefixListAttributes(&group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_ingressWithPrefixList(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigPrefixListIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupIngressPrefixListAttributes(&group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccAWSSecurityGroup_ipv4andipv6Egress(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfigIpv4andIpv6Egress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "1",
						"cidr_blocks.0":      "0.0.0.0/0",
						"description":        "",
						"from_port":          "0",
						"ipv6_cidr_blocks.#": "0",
						"prefix_list_ids.#":  "0",
						"protocol":           "-1",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":      "0",
						"description":        "",
						"from_port":          "0",
						"ipv6_cidr_blocks.#": "1",
						"ipv6_cidr_blocks.0": "::/0",
						"prefix_list_ids.#":  "0",
						"protocol":           "-1",
						"security_groups.#":  "0",
						"self":               "false",
						"to_port":            "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "egress"},
			},
		},
	})
}

func testAccAWSSecurityGroupCheckVPCIDExists(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if aws.StringValue(group.VpcId) == "" {
			return fmt.Errorf("should have vpc ID")
		}
		return nil
	}
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

func testAccCheckAWSSecurityGroupEc2ClassicDestroy(s *terraform.State) error {
	conn := testAccProviderEc2Classic.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_security_group" {
			continue
		}

		input := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSecurityGroups(input)

		if tfawserr.ErrCodeEquals(err, "InvalidGroup.NotFound") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing EC2 Security Group (%s): %w", rs.Primary.ID, err)
		}

		for _, sg := range output.SecurityGroups {
			if aws.StringValue(sg.GroupId) == rs.Primary.ID {
				return fmt.Errorf("EC2 Security Group (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
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

func testAccCheckAWSSecurityGroupEc2ClassicExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group is set")
		}

		conn := testAccProviderEc2Classic.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSecurityGroups(input)

		if err != nil {
			return fmt.Errorf("error describing EC2 Security Group (%s): %w", rs.Primary.ID, err)
		}

		for _, sg := range output.SecurityGroups {
			if aws.StringValue(sg.GroupId) == rs.Primary.ID {
				*group = *sg
				return nil
			}
		}

		return fmt.Errorf("EC2 Security Group (%s) not found", rs.Primary.ID)
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

	resourceName := "aws_security_group.nat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityGroupConfig_failWithDiffMismatch,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitExceededAppend(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup

	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
					resource.TestCheckResourceAttr(resourceName, "egress.#", strconv.Itoa(ruleLimit)),
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
					err := testSecurityGroupRuleCount(aws.StringValue(group.GroupId), 0, ruleLimit)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
					resource.TestCheckResourceAttr(resourceName, "egress.#", strconv.Itoa(ruleLimit)),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitCidrBlockExceededAppend(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup

	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigCidrBlockRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
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
					err := testSecurityGroupRuleCount(aws.StringValue(group.GroupId), 0, 1)
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
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, 1),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitExceededPrepend(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup

	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
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
					err := testSecurityGroupRuleCount(aws.StringValue(group.GroupId), 0, ruleLimit-1)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_ruleLimitExceededAllNew(t *testing.T) {
	ruleLimit := testAccAWSSecurityGroupRulesPerGroupLimitFromEnv()

	var group ec2.SecurityGroup

	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
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
					err := testSecurityGroupRuleCount(aws.StringValue(group.GroupId), 0, 0)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccAWSSecurityGroupConfigRuleLimit(0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
					testAccCheckAWSSecurityGroupRuleCount(&group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_rulesDropOnError(t *testing.T) {
	var group ec2.SecurityGroup

	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			// Create a valid security group with some rules and make sure it exists
			{
				Config: testAccAWSSecurityGroupConfig_rulesDropOnError_Init,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists(resourceName, &group),
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
		id := aws.StringValue(group.GroupId)
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
	var egressRules strings.Builder
	for i := egressStartIndex; i < egressRulesCount+egressStartIndex; i++ {
		fmt.Fprintf(&egressRules, `
  egress {
    protocol    = "tcp"
    from_port   = "${80 + %[1]d}"
    to_port     = "${80 + %[1]d}"
    cidr_blocks = ["${cidrhost("10.1.0.0/16", %[1]d)}/32"]
  }
`, i)
	}

	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-limit"
  }
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_rule_limit"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test"
  }

  # egress rules to exhaust the limit
  %[1]s
}
`, egressRules.String())
}

func testAccAWSSecurityGroupConfigCidrBlockRuleLimit(egressStartIndex, egressRulesCount int) string {
	var cidrBlocks strings.Builder
	for i := egressStartIndex; i < egressRulesCount+egressStartIndex; i++ {
		fmt.Fprintf(&cidrBlocks, `
		"${cidrhost("10.1.0.0/16", %[1]d)}/32",
`, i)
	}

	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-limit"
  }
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_rule_limit"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test"
  }

  egress {
    protocol  = "tcp"
    from_port = "80"
    to_port   = "80"
    # cidr_blocks to exhaust the limit
    cidr_blocks = [
		%s
    ]
  }
}
`, cidrBlocks.String())
}

const testAccAWSSecurityGroupConfigEmptyRuleDescription = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-empty-rule-description"
  }
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_desc_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`

const testAccAWSSecurityGroupConfigIpv6 = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-ipv6"
  }
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol         = "6"
    from_port        = 80
    to_port          = 8000
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    protocol         = "tcp"
    from_port        = 80
    to_port          = 8000
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
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
  name        = "tf-acc-sg-race-revoke-primary"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.sg-race-revoke.id

  tags = {
    Name = "tf-acc-revoke-test-primary"
  }
}

resource "aws_security_group" "secondary" {
  name        = "tf-acc-sg-race-revoke-secondary"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.sg-race-revoke.id

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
  name        = "tf-acc-sg-race-revoke-primary"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.sg-race-revoke.id

  tags = {
    Name = "tf-acc-revoke-test-primary"
  }

  revoke_rules_on_delete = false
}

resource "aws_security_group" "secondary" {
  name        = "tf-acc-sg-race-revoke-secondary"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.sg-race-revoke.id

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
  name        = "tf-acc-sg-race-revoke-primary"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.sg-race-revoke.id

  tags = {
    Name = "tf-acc-revoke-test-primary"
  }

  revoke_rules_on_delete = true
}

resource "aws_security_group" "secondary" {
  name        = "tf-acc-sg-race-revoke-secondary"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.sg-race-revoke.id

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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 9000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["0.0.0.0/0", "10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = "%s"
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol  = "tcp"
    from_port = 80
    to_port   = 8000
    self      = true
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "50"
    from_port   = 0
    to_port     = 0
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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 800
    to_port     = 800
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 8000
    security_groups = [aws_security_group.test.id]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

func testAccAWSAWSSecurityGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAWSSecurityGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccAWSSecurityGroupConfig_generatedName = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-generated-name"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.foo.id

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

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.tf_sg_egress_test.id

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

func testAccAWSSecurityGroupConfigClassic(rName string) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used in the terraform acceptance tests"
}
`, rName))
}

func testAccAWSSecurityGroupConfigName(name string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-security-group-name"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}
`, name)
}

func testAccAWSSecurityGroupConfigNamePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-security-group-name-prefix"
  }
}

resource "aws_security_group" "test" {
  name_prefix = %[1]q
  vpc_id      = aws_vpc.test.id
}
`, namePrefix)
}

func testAccAWSSecurityGroupConfig_drift() string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = "tf_acc_%d"
  description = "Used in the terraform acceptance tests"

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

resource "aws_security_group" "test2" {
  name        = "tf_acc_%d"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group" "test" {
  name        = "tf_acc_%d"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

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
    security_groups = [aws_security_group.test2.id]
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
    security_groups = [aws_security_group.test2.id]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`, acctest.RandInt(), acctest.RandInt())
}

const testAccAWSSecurityGroupInvalidIngressCidr = `
resource "aws_security_group" "test" {
  name        = "testing-foo"
  description = "foo-testing"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["1.2.3.4/33"]
  }
}
`

const testAccAWSSecurityGroupInvalidEgressCidr = `
resource "aws_security_group" "test" {
  name        = "testing-foo"
  description = "foo-testing"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["1.2.3.4/33"]
  }
}
`

const testAccAWSSecurityGroupInvalidIPv6IngressCidr = `
resource "aws_security_group" "test" {
  name        = "testing-foo"
  description = "foo-testing"

  ingress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/244"]
  }
}
`

const testAccAWSSecurityGroupInvalidIPv6EgressCidr = `
resource "aws_security_group" "test" {
  name        = "testing-foo"
  description = "foo-testing"

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/244"]
  }
}
`

const testAccAWSSecurityGroupCombindCIDRandGroups = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-combine-rand-groups"
  }
}

resource "aws_security_group" "two" {
  name   = "tf-test-1"
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-test-1"
  }
}

resource "aws_security_group" "one" {
  name   = "tf-test-2"
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-test-w"
  }
}

resource "aws_security_group" "three" {
  name   = "tf-test-3"
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-test-3"
  }
}

resource "aws_security_group" "test" {
  name   = "tf-mix-test"
  vpc_id = aws_vpc.foo.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16", "10.1.0.0/16", "10.7.0.0/16"]

    security_groups = [
      aws_security_group.one.id,
      aws_security_group.two.id,
      aws_security_group.three.id,
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

resource "aws_security_group" "test2" {
  name        = "tf_other_acc_tests"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

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
    security_groups = [aws_security_group.test2.id]
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

func testAccAWSSecurityGroupConfig_ingressWithCidrAndSGs_classic(rName string) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_security_group" "test2" {
  name        = "%[1]s-2"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
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
    security_groups = [aws_security_group.test2.name]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`, rName))
}

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
  vpc_id = aws_vpc.main.id
}

resource "aws_security_group" "jump" {
  name   = "test-jump"
  vpc_id = aws_vpc.main.id
}

resource "aws_security_group" "provision" {
  name   = "test-provision"
  vpc_id = aws_vpc.main.id
}

resource "aws_security_group" "nat" {
  vpc_id      = aws_vpc.main.id
  name        = "nat"
  description = "For nat servers "

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.jump.id]
  }

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.provision.id]
  }
}
`

const testAccAWSSecurityGroupConfig_allowAll = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-allow-all"
  }
}

resource "aws_security_group" "test" {
  name        = "allow_all"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_all" {
  type        = "ingress"
  from_port   = 0
  to_port     = 65535
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_all-1" {
  type      = "ingress"
  from_port = 65534
  to_port   = 65535
  protocol  = "tcp"

  self              = true
  security_group_id = aws_security_group.test.id
}
`

const testAccAWSSecurityGroupConfig_sourceSecurityGroup = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-source-sg"
  }
}

resource "aws_security_group" "test" {
  name   = "test group 1"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group" "test2" {
  name   = "test group 2"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group" "test3" {
  name   = "test group 3"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_test2" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test.id
  security_group_id        = aws_security_group.test2.id
}

resource "aws_security_group_rule" "allow_test3" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test.id
  security_group_id        = aws_security_group.test3.id
}
`

const testAccAWSSecurityGroupConfig_IPRangeAndSecurityGroupWithSameRules = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-import-ip-range-and-sg"
  }
}

resource "aws_security_group" "test" {
  name   = "test group 1"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group" "test2" {
  name   = "test group 2"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_security_group" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test2.id
  security_group_id        = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  cidr_blocks       = ["10.0.0.0/32"]
  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_ipv6_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  ipv6_cidr_blocks  = ["::/0"]
  security_group_id = aws_security_group.test.id
}
`

const testAccAWSSecurityGroupConfig_IPRangesWithSameRules = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-import-ip-ranges"
  }
}

resource "aws_security_group" "test" {
  name   = "test group 1"
  vpc_id = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  cidr_blocks       = ["10.0.0.0/32"]
  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_ipv6_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  ipv6_cidr_blocks  = ["::/0"]
  security_group_id = aws_security_group.test.id
}
`

const testAccAWSSecurityGroupConfigIpv4andIpv6Egress = `
resource "aws_vpc" "foo" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-security-group-ipv4-and-ipv6-egress"
  }
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.foo.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/0"]
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
  vpc_id = aws_vpc.tf_sg_prefix_list_egress_test.id
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.tf_sg_prefix_list_egress_test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.default.id]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_prefix_list_egress"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.tf_sg_prefix_list_egress_test.id

  egress {
    protocol        = "-1"
    from_port       = 0
    to_port         = 0
    prefix_list_ids = [aws_vpc_endpoint.test.prefix_list_id]
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
  vpc_id = aws_vpc.tf_sg_prefix_list_ingress_test.id
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.tf_sg_prefix_list_ingress_test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.default.id]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_prefix_list_ingress"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.tf_sg_prefix_list_ingress_test.id

  ingress {
    protocol        = "-1"
    from_port       = 0
    to_port         = 0
    prefix_list_ids = [aws_vpc_endpoint.test.prefix_list_id]
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
    Name = var.name
  }
}

resource "aws_route_table" "default" {
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.default.id]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_security_group" "source1" {
  name        = "${var.name}-source1"
  description = "terraform acceptance test for security group as source1"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group" "source2" {
  name        = "${var.name}-source2"
  description = "terraform acceptance test for security group as source2"
  vpc_id      = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name        = var.name
  description = "terraform acceptance test for security group"
  vpc_id      = aws_vpc.test.id

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
    security_groups = [aws_security_group.source1.id, aws_security_group.source2.id]
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
    prefix_list_ids = [aws_vpc_endpoint.test.prefix_list_id]
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
  name   = "terraform_acceptance_test_drop_rules_ref0"
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test_ref1" {
  name   = "terraform_acceptance_test_drop_rules_ref1"
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_drop_rules"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test"
  }

  ingress {
    protocol  = "tcp"
    from_port = "80"
    to_port   = "80"
    security_groups = [
      aws_security_group.test_ref0.id,
      aws_security_group.test_ref1.id,
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
  name   = "terraform_acceptance_test_drop_rules_ref0"
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test_ref1" {
  name   = "terraform_acceptance_test_drop_rules_ref1"
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name        = "terraform_acceptance_test_drop_rules"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test"
  }

  ingress {
    protocol  = "tcp"
    from_port = "80"
    to_port   = "80"
    security_groups = [
      aws_security_group.test_ref0.id,
      aws_security_group.test_ref1.id,
      "sg-malformed", # non-existent rule to trigger API error
    ]
  }
}
`

func testAccAWSSecurityGroupConfigEgressConfigModeBlocks() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-egress-config-mode"
  }
}

resource "aws_security_group" "test" {
  tags = {
    Name = "terraform-testacc-security-group-egress-config-mode"
  }

  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "tcp"
    to_port     = 0
  }

  egress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "udp"
    to_port     = 0
  }
}
`
}

func testAccAWSSecurityGroupConfigEgressConfigModeNoBlocks() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-egress-config-mode"
  }
}

resource "aws_security_group" "test" {
  tags = {
    Name = "terraform-testacc-security-group-egress-config-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccAWSSecurityGroupConfigEgressConfigModeZeroed() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-egress-config-mode"
  }
}

resource "aws_security_group" "test" {
  egress = []

  tags = {
    Name = "terraform-testacc-security-group-egress-config-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccAWSSecurityGroupConfigIngressConfigModeBlocks() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-ingress-config-mode"
  }
}

resource "aws_security_group" "test" {
  tags = {
    Name = "terraform-testacc-security-group-ingress-config-mode"
  }

  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "tcp"
    to_port     = 0
  }

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "udp"
    to_port     = 0
  }
}
`
}

func testAccAWSSecurityGroupConfigIngressConfigModeNoBlocks() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-ingress-config-mode"
  }
}

resource "aws_security_group" "test" {
  tags = {
    Name = "terraform-testacc-security-group-ingress-config-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}

func testAccAWSSecurityGroupConfigIngressConfigModeZeroed() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-ingress-config-mode"
  }
}

resource "aws_security_group" "test" {
  ingress = []

  tags = {
    Name = "terraform-testacc-security-group-ingress-config-mode"
  }

  vpc_id = aws_vpc.test.id
}
`
}
