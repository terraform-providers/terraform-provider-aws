package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDBSecurityGroup_importBasic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rName := fmt.Sprintf("tf-acc-%s", acctest.RandString(5))
	resourceName := "aws_db_security_group.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBSecurityGroupConfig(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDBSecurityGroup_basic(t *testing.T) {
	var v rds.DBSecurityGroup

	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rName := fmt.Sprintf("tf-acc-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBSecurityGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBSecurityGroupExists("aws_db_security_group.bar", &v),
					testAccCheckAWSDBSecurityGroupAttributes(&v),
					resource.TestMatchResourceAttr("aws_db_security_group.bar", "arn", regexp.MustCompile(`^arn:[^:]+:rds:[^:]+:\d{12}:secgrp:.+`)),
					resource.TestCheckResourceAttr(
						"aws_db_security_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_security_group.bar", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"aws_db_security_group.bar", "ingress.3363517775.cidr", "10.0.0.1/24"),
					resource.TestCheckResourceAttr(
						"aws_db_security_group.bar", "ingress.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_db_security_group.bar", "tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSDBSecurityGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_security_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBSecurityGroups(
			&rds.DescribeDBSecurityGroupsInput{
				DBSecurityGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBSecurityGroups) != 0 &&
				*resp.DBSecurityGroups[0].DBSecurityGroupName == rs.Primary.ID {
				return fmt.Errorf("DB Security Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "DBSecurityGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDBSecurityGroupAttributes(group *rds.DBSecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(group.IPRanges) == 0 {
			return fmt.Errorf("no cidr: %#v", group.IPRanges)
		}

		if *group.IPRanges[0].CIDRIP != "10.0.0.1/24" {
			return fmt.Errorf("bad cidr: %#v", group.IPRanges)
		}

		statuses := make([]string, 0, len(group.IPRanges))
		for _, ips := range group.IPRanges {
			statuses = append(statuses, *ips.Status)
		}

		if statuses[0] != "authorized" {
			return fmt.Errorf("bad status: %#v", statuses)
		}

		return nil
	}
}

func testAccCheckAWSDBSecurityGroupExists(n string, v *rds.DBSecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Security Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBSecurityGroupsInput{
			DBSecurityGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBSecurityGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBSecurityGroups) != 1 ||
			*resp.DBSecurityGroups[0].DBSecurityGroupName != rs.Primary.ID {
			return fmt.Errorf("DB Security Group not found")
		}

		*v = *resp.DBSecurityGroups[0]

		return nil
	}
}

func testAccAWSDBSecurityGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_db_security_group" "bar" {
    name = "%s"

    ingress {
        cidr = "10.0.0.1/24"
    }

  tags = {
		foo = "bar"
    }
}
`, name)
}
