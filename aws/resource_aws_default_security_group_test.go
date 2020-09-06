package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSDefaultSecurityGroup_Vpc_basic(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSecurityGroupConfig_Vpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDefaultSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "default"),
					resource.TestCheckResourceAttr(resourceName, "description", "default VPC security group"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					testAccCheckAWSDefaultSecurityGroupARN(resourceName, &group),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-test"),
				),
			},
			{
				Config:   testAccAWSDefaultSecurityGroupConfig_Vpc,
				PlanOnly: true,
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

func TestAccAWSDefaultSecurityGroup_Vpc_empty(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSecurityGroupConfig_Vpc_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDefaultSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
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

func TestAccAWSDefaultSecurityGroup_Classic_basic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSecurityGroupConfig_Classic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDefaultSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "default"),
					resource.TestCheckResourceAttr(resourceName, "description", "default group"),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					testAccCheckAWSDefaultSecurityGroupARN(resourceName, &group),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-test"),
				),
			},
			{
				Config:   testAccAWSDefaultSecurityGroupConfig_Classic,
				PlanOnly: true,
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

func TestAccAWSDefaultSecurityGroup_Classic_empty(t *testing.T) {

	TestAccSkip(t, "This resource does not currently clear tags when adopting the resource")
	// Additional references:
	//  * https://github.com/terraform-providers/terraform-provider-aws/issues/14631

	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSecurityGroupConfig_Classic_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDefaultSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSDefaultSecurityGroupDestroy(s *terraform.State) error {
	// We expect Security Group to still exist
	return nil
}

func testAccCheckAWSDefaultSecurityGroupExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
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

func testAccCheckAWSDefaultSecurityGroupARN(resourceName string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ec2", fmt.Sprintf("security-group/%s", aws.StringValue(group.GroupId)))(s)
	}
}

const testAccAWSDefaultSecurityGroupConfig_Vpc = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  
  tags = {
    Name = "terraform-testacc-default-security-group"
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
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

  tags = {
    Name = "tf-acc-test"
  }
}
`

const testAccAWSDefaultSecurityGroupConfig_Vpc_empty = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-security-group"
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
`

const testAccAWSDefaultSecurityGroupConfig_Classic = `
resource "aws_default_security_group" "test" {
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`

const testAccAWSDefaultSecurityGroupConfig_Classic_empty = `
resource "aws_default_security_group" "test" {
  # No attributes set.
}
`

func TestAWSDefaultSecurityGroupMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0": {
			StateVersion: 0,
			Attributes: map[string]string{
				"name": "test",
			},
			Expected: map[string]string{
				"name":                   "test",
				"revoke_rules_on_delete": "false",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsDefaultSecurityGroupMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}

func TestAWSDefaultSecurityGroupMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceAwsDefaultSecurityGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = resourceAwsDefaultSecurityGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
