package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafregional"
)

func TestAccAWSWafRegionalWebAclAssociation_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWafRegionalWebAclAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWafRegionalWebAclAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWafRegionalWebAclAssociationExists("aws_wafregional_web_acl_association.foo"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAclAssociation_multipleAssociations(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWafRegionalWebAclAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWafRegionalWebAclAssociationConfig_multipleAssociations,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWafRegionalWebAclAssociationExists("aws_wafregional_web_acl_association.foo"),
					testAccCheckWafRegionalWebAclAssociationExists("aws_wafregional_web_acl_association.bar"),
				),
			},
		},
	})
}

func testAccCheckWafRegionalWebAclAssociationDestroy(s *terraform.State) error {
	return testAccCheckWafRegionalWebAclAssociationDestroyWithProvider(s, testAccProvider)
}

func testAccCheckWafRegionalWebAclAssociationDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).wafregionalconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_web_acl_association" {
			continue
		}

		webAclId, resourceArn := resourceAwsWafRegionalWebAclAssociationParseId(rs.Primary.ID)

		resp, err := conn.ListResourcesForWebACL(&wafregional.ListResourcesForWebACLInput{WebACLId: aws.String(webAclId)})
		if err != nil {
			found := false
			for _, listResourceArn := range resp.ResourceArns {
				if resourceArn == *listResourceArn {
					found = true
					break
				}
			}
			if found {
				return fmt.Errorf("WebACL: %v is still associated to resource: %v", webAclId, resourceArn)
			}
		}
	}
	return nil
}

func testAccCheckWafRegionalWebAclAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckWafRegionalWebAclAssociationExistsWithProvider(s, n, testAccProvider)
	}
}

func testAccCheckWafRegionalWebAclAssociationExistsWithProvider(s *terraform.State, n string, provider *schema.Provider) error {
	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return fmt.Errorf("Not found: %s", n)
	}

	if rs.Primary.ID == "" {
		return fmt.Errorf("No WebACL association ID is set")
	}

	webAclId, resourceArn := resourceAwsWafRegionalWebAclAssociationParseId(rs.Primary.ID)

	conn := provider.Meta().(*AWSClient).wafregionalconn
	resp, err := conn.ListResourcesForWebACL(&wafregional.ListResourcesForWebACLInput{WebACLId: aws.String(webAclId)})
	if err != nil {
		return fmt.Errorf("List Web ACL err: %v", err)
	}

	found := false
	for _, listResourceArn := range resp.ResourceArns {
		if resourceArn == *listResourceArn {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Web ACL association not found")
	}

	return nil
}

const testAccCheckWafRegionalWebAclAssociationConfig_basic = `
resource "aws_wafregional_rule" "foo" {
  name = "foo"
  metric_name = "foo"
}

resource "aws_wafregional_web_acl" "foo" {
  name = "foo"
  metric_name = "foo"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
      type = "COUNT"
    }
    priority = 100
    rule_id = "${aws_wafregional_rule.foo.id}"
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.1.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.1.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_alb" "foo" {
  internal = true
  subnets = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
}

resource "aws_wafregional_web_acl_association" "foo" {
  resource_arn = "${aws_alb.foo.arn}"
  web_acl_id = "${aws_wafregional_web_acl.foo.id}"
}
`

const testAccCheckWafRegionalWebAclAssociationConfig_multipleAssociations = testAccCheckWafRegionalWebAclAssociationConfig_basic + `
resource "aws_alb" "bar" {
  internal = true
  subnets = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
}

resource "aws_wafregional_web_acl_association" "bar" {
  resource_arn = "${aws_alb.bar.arn}"
  web_acl_id = "${aws_wafregional_web_acl.foo.id}"
}
`
