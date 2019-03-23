package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworkscm"
)

func TestAccAWSOpsworksChefImportBasic(t *testing.T) {
	name := acctest.RandString(10)

	resourceName := fmt.Sprintf("aws_opsworks_chef.%s", name)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksChefDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksChefConfigCreate(name),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsOpsworksChefConfigCreate(name string) string {
	return fmt.Sprintf(`
		resource "aws_vpc" "tf-acc" {
			cidr_block = "10.3.5.0/24"
			tags = {
				Name = "terraform-testacc-opsworks-chef-create"
			}
		}

		// per https://s3.amazonaws.com/opsworks-cm-us-east-1-prod-default-assets/misc/opsworks-cm-roles.yaml
		// per https://docs.aws.amazon.com/sdk-for-go/api/service/opsworkscm/
		resource "aws_iam_role" "opsworks_chef_instance_role" {
			name = "opsworks_chef_instance_role"
			assume_role_policy = "${data.aws_iam_policy_document.opsworks_chef_instance_assumerole.json}"
			path = "/"
		}
		
		data "aws_iam_policy_document" "opsworks_chef_instance_assumerole" {
			statement {
				actions = ["sts:AssumeRole"]
				principals {
					type = "Service"
					identifiers = ["ec2.amazonaws.com"]
				}
			}
		}
		
		resource "aws_iam_role_policy_attachment" "opswork_chef_instance_role_ssm" {
			role = "${aws_iam_role.opsworks_chef_instance_role.name}"
			policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM"
		}
		
		resource "aws_iam_role_policy_attachment" "opsworks_chef_instance_role_profile" {
			role = "${aws_iam_role.opsworks_chef_instance_role.name}"
			policy_arn = "arn:aws:iam::aws:policy/AWSOpsWorksCMInstanceProfileRole"
		}

		resource "aws_iam_instance_profile" "opsworks_chef_instance_profile" {
			name = "opsworks_chef_instance_profile"
			role = "${aws_iam_role.opsworks_chef_instance_role.name}"
		}
		
		resource "aws_iam_role" "opsworks_chef_service" {
			name = "opsworks_chef_service_role"
			assume_role_policy = "${data.aws_iam_policy_document.opsworks_chef_service_assumerole.json}"
			path = "/"
		}
		
		data "aws_iam_policy_document" "opsworks_chef_service_assumerole" {
			statement {
				actions = ["sts:AssumeRole"]
				principals {
					type = "Service"
					identifiers = ["opsworks-cm.amazonaws.com"]
				}
			}
		}
		
		resource "aws_iam_role_policy_attachment" "opsworks_chef_service_role" {
			role       = "${aws_iam_role.opsworks_chef_service.name}"
			policy_arn = "arn:aws:iam::aws:policy/service-role/AWSOpsWorksCMServiceRole"
		}
		
		resource "aws_security_group" "opsworks_chef" {
			name = "opsworks_chef"
			description = "security group for opsworks chef server"
			vpc_id = "${aws_vpc.tf-acc.id}"

			ingress {
				protocol = "TCP"
				from_port = 0
				to_port = 443
				cidr_blocks = ["0.0.0.0/0"]
			}
			
			ingress {
				protocol = "TCP"
				from_port = 0
				to_port = 22
				cidr_blocks = ["0.0.0.0/0"]
			}
		}

		resource "aws_subnet" "tf-acc" {
			vpc_id = "${aws_vpc.tf-acc.id}"
			cidr_block = "${aws_vpc.tf-acc.cidr_block}"
			availability-zone = "us-west-2a"
			tags = {
				Name = "tf-acc-opsworks-chef-create"
			}
		}

		resource "tls_private_key" "opsworks_chef_rsa_key" {
			algorithm = "RSA"
		}
		
		resource "random_string" "opsworks_chef_admin_password" {
			length = 24
			min_lower = 1
			min_upper = 1
			min_numeric = 1
			min_special = 1
			override_special = "!/@#$%%^&+=_"
		}

		resource "aws_opsworks_chef" "%s" {
			chef_pivotal_key = "${tls_private_key.opsworks_chef_rsa_key.private_key_pem}"
			chef_delivery_admin_password = "${random_string.opsworks_chef_admin_password.result}"
			instance_profile_arn = "${aws_iam_instance_profile.opsworks_chef_instance.arn}"
			instance_type = "t2.medium"
			preferred_backup_window = "Mon:08:00"
			preferred_maintenance_window = "Sun:08:00"
			security_group_ids = ["${aws_security_group.opsworks_chef.id}"]
			service_role_arn = "${aws_iam_role.opsworks_chef_service.arn}"
			subnet_ids = ["${aws_subnet.tf-acc.id}"]
		}
		`, name)
}

func testAccCheckAwsOpsworksChefDestroy(s *terraform.State) error {
	opsworkscmconn := testAccProvider.Meta().(*AWSClient).opsworkscmconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_chef" {
			continue
		}

		req := &opsworkscm.DescribeServersInput{
			ServerName: aws.String(rs.Primary.ID),
		}

		_, err := opsworkscmconn.DescribeServers(req)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() == "ResourceNotFoundException" {
					// not found, all good
					return nil
				}
			}
			return err
		}
	}
	// TODO: implement
	return fmt.Errorf("fall through error for OpsWorks Chef test")
}
