package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSNetworkAclAssociation(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_network_acl.acl_a",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkAclAssoc,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetIsAssociatedWithAcl("aws_network_acl.acl_a", "aws_subnet.sunet_a"),
				),
			},
		},
	})
}

const testAccAWSNetworkAclAssoc = `
resource "aws_vpc" "testespvpc" {
  cidr_block = "10.1.0.0/16"
    tags {
       Name = "testAccAWSNetworkAclEsp"
  }
}

 resource "aws_network_acl" "acl_a" {
   vpc_id = "${aws_vpc.testespvpc.id}"

   tags {
     Name = "terraform test"
   }
 }

 resource "aws_subnet" "sunet_a" {
   vpc_id = "${aws_vpc.testespvpc.id}"
   cidr_block = "10.0.33.0/24"
   tags {
     Name = "terraform test"
   }
 }

 resource "aws_network_acl_association" "test" {
   network_acl_id = "${aws_network_acl.acl_a.id}"
   subnet_id = "${aws_subnet.subnet_a.id}"
 }
}`
