package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/service/rds"
)

func TestAccAWSDefaultDBSubnetGroup_basic(t *testing.T) {
	var v rds.DBSubnetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDefaultDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultDBSubnetGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_default_db_subnet_group.default", &v),
					resource.TestCheckResourceAttr(
						"aws_default_db_subnet_group.default", "name", "default"),
					resource.TestCheckResourceAttr(
						"aws_default_db_subnet_group.default", "description", "Managed by Terraform"),
					resource.TestMatchResourceAttr(
						"aws_default_db_subnet_group.default", "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:rds:[^:]+:\\d{12}:subgrp:default"))),
					resource.TestCheckResourceAttr(
						"aws_default_db_subnet_group.default", "tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckDefaultDBSubnetGroupDestroy(s *terraform.State) error {
	// We expect thid resource to still exist
	return nil
}

const testAccDefaultDBSubnetGroupConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_default_subnet" "default_az1" {
  availability_zone = "us-west-2a"
}

resource "aws_default_subnet" "default_az2" {
  availability_zone = "us-west-2b"
}

resource "aws_default_db_subnet_group" "default" {
  subnet_ids = ["${aws_default_subnet.default_az1.id}", "${aws_default_subnet.default_az2.id}"]
  tags {
    Name = "Default DB subnet group"
  }
}
`
