package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMAssociation_basic(t *testing.T) {
	name := fmt.Sprintf("tf-acc-ssm-association-%s", acctest.RandString(10))

	deleteSsmAssociaton := func() {
		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn
		ssmconn := testAccProvider.Meta().(*AWSClient).ssmconn

		ins, err := ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []*string{aws.String(name)},
				},
			},
		})
		if err != nil {
			t.Fatalf("Error getting instance with tag:Name %s: %s", name, err)
		}
		if len(ins.Reservations) == 0 || len(ins.Reservations[0].Instances) == 0 {
			t.Fatalf("No instance exists with tag:Name %s", name)
		}
		instanceId := ins.Reservations[0].Instances[0].InstanceId

		_, err = ssmconn.DeleteAssociation(&ssm.DeleteAssociationInput{
			Name:       aws.String(name),
			InstanceId: instanceId,
		})
		if err != nil {
			t.Fatalf("Error deleting ssm association %s/%s: %s", name, aws.StringValue(instanceId), err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
				),
			},
			{
				PreConfig: deleteSsmAssociaton,
				Config:    testAccAWSSSMAssociationBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withTargets(t *testing.T) {
	name := acctest.RandString(10)
	oneTarget := `
	targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }`
	twoTargets := `
	targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
  targets {
    key = "tag:ExtraName"
    values = ["acceptanceTest"]
  }`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(name, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.0.values.0", "acceptanceTest"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(name, twoTargets),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.0.values.0", "acceptanceTest"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.1.key", "tag:ExtraName"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.1.values.0", "acceptanceTest"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(name, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "targets.0.values.0", "acceptanceTest"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withParameters(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParameters(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParametersUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAssociationName(t *testing.T) {
	assocName1 := acctest.RandString(10)
	assocName2 := acctest.RandString(10)
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "association_name", assocName1),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "association_name", assocName2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAssociationNameAndScheduleExpression(t *testing.T) {
	assocName := acctest.RandString(10)
	rName := acctest.RandString(5)
	resourceName := "aws_ssm_association.test"
	scheduleExpression1 := "cron(0 16 ? * TUE *)"
	scheduleExpression2 := "cron(0 16 ? * WED *)"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, assocName, scheduleExpression1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression1),
				),
			},
			{
				Config: testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, assocName, scheduleExpression2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withDocumentVersion(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithDocumentVersion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "document_version", "1"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withOutputLocation(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocation(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-%s", name)),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateBucketName(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-updated-%s", name)),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateKeyPrefix(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-updated-%s", name)),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "output_location.0.s3_key_prefix", "UpdatedAssociation"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withScheduleExpression(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithScheduleExpression(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "schedule_expression", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithScheduleExpressionUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists("aws_ssm_association.foo"),
					resource.TestCheckResourceAttr(
						"aws_ssm_association.foo", "schedule_expression", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func testAccCheckAWSSSMAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Assosciation ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		_, err := conn.DescribeAssociation(&ssm.DescribeAssociationInput{
			AssociationId: aws.String(rs.Primary.Attributes["association_id"]),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == "AssociationDoesNotExist" {
				return nil
			}
			return err
		}

		return nil
	}
}

func testAccCheckAWSSSMAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_association" {
			continue
		}

		out, err := conn.DescribeAssociation(&ssm.DescribeAssociationInput{
			AssociationId: aws.String(rs.Primary.Attributes["association_id"]),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == "AssociationDoesNotExist" {
				return nil
			}
			return err
		}

		if out != nil {
			return fmt.Errorf("Expected AWS SSM Association to be gone, but was still found")
		}
	}

	return fmt.Errorf("Default error in SSM Association Test")
}

func testAccAWSSSMAssociationBasicConfigWithParameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<-DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {
	  "Directory": {
		"description":"(Optional) The path to the working directory on your instance.",
		"default":"",
		"type": "String",
		"maxChars": 4096
	  }
	},
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
  DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  parameters = {
  	Directory = "myWorkSpace"
  }
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithParametersUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<-DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {
	  "Directory": {
		"description":"(Optional) The path to the working directory on your instance.",
		"default":"",
		"type": "String",
		"maxChars": 4096
	  }
	},
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
  DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  parameters = {
  	Directory = "myWorkSpaceUpdated"
  }
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithTargets(rName, targetsStr string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  %s
}`, rName, targetsStr)
}

func testAccAWSSSMAssociationBasicConfig(rName string) string {
	return fmt.Sprintf(`
variable "name" { default = "%s" }

data "aws_availability_zones" "available" {}

data "aws_ami" "amzn" {
  most_recent      = true
  owners     = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "${var.name}"
  }
}

resource "aws_subnet" "first" {
  vpc_id = "${aws_vpc.main.id}"
  cidr_block = "10.0.0.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_security_group" "tf_test_foo" {
  name = "${var.name}"
  description = "foo"
  vpc_id = "${aws_vpc.main.id}"
  ingress {
    protocol = "icmp"
    from_port = -1
    to_port = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "foo" {
  ami = "${data.aws_ami.amzn.image_id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  instance_type = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.tf_test_foo.id}"]
  subnet_id = "${aws_subnet.first.id}"
  tags = {
    Name = "${var.name}"
  }
}

resource "aws_ssm_document" "foo_document" {
  name    = "${var.name}"
	document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name        = "${var.name}"
  instance_id = "${aws_instance.foo.id}"
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithDocumentVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name    = "test_document_association-%s"
	document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name        = "test_document_association-%s"
  document_version = "${aws_ssm_document.foo_document.latest_version}"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithScheduleExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  schedule_expression = "cron(0 16 ? * TUE *)"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithScheduleExpressionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  schedule_expression = "cron(0 16 ? * WED *)"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocation(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
  output_location {
    s3_bucket_name = "${aws_s3_bucket.output_location.id}"
    s3_key_prefix = "SSMAssociation"
  }
}`, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateBucketName(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket = "tf-acc-test-ssmoutput-updated-%s"
  force_destroy = true
}

resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
  output_location {
    s3_bucket_name = "${aws_s3_bucket.output_location_updated.id}"
    s3_key_prefix = "SSMAssociation"
  }
}`, rName, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateKeyPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket = "tf-acc-test-ssmoutput-updated-%s"
  force_destroy = true
}

resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
  output_location {
    s3_bucket_name = "${aws_s3_bucket.output_location_updated.id}"
    s3_key_prefix = "UpdatedAssociation"
  }
}`, rName, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {
    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name = "${aws_ssm_document.foo_document.name}"
  association_name = "%s"
  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName)
}

func testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, associationName, scheduleExpression string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name = "test_document_association-%s"
  document_type = "Command"
  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {
    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "test" {
  association_name    = %q
  name                = "${aws_ssm_document.test.name}"
  schedule_expression = %q

  targets {
    key = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, associationName, scheduleExpression)
}
