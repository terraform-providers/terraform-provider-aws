package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAMI_basic(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccAWSAMI_description(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	desc := acctest.RandomWithPrefix("desc")
	descUpdated := acctest.RandomWithPrefix("desc-updated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfigDesc(rName, desc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
			{
				Config: testAccAmiConfigDesc(rName, descUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descUpdated),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAMI_disappears(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAmi(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAMI_EphemeralBlockDevices(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfigEphemeralBlockDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name":  "/dev/sdb",
						"virtual_name": "ephemeral0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name":  "/dev/sdc",
						"virtual_name": "ephemeral1",
					}),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccAWSAMI_Gp3BlockDevice(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfigGp3BlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "false",
						"device_name":           "/dev/sdb",
						"encrypted":             "true",
						"iops":                  "100",
						"throughput":            "500",
						"volume_size":           "10",
						"volume_type":           "gp3",
					}),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccAWSAMI_tags(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
			{
				Config: testAccAmiConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAmiConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAmiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami" {
			continue
		}

		// Try to find the AMI
		log.Printf("AMI-ID: %s", rs.Primary.ID)
		DescribeAmiOpts := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeImages(DescribeAmiOpts)
		if err != nil {
			if isAWSErr(err, "InvalidAMIID", "NotFound") {
				log.Printf("[DEBUG] AMI not found, passing")
				return nil
			}
			return err
		}

		if len(resp.Images) > 0 {
			state := resp.Images[0].State
			return fmt.Errorf("AMI %s still exists in the state: %s.", *resp.Images[0].ImageId, *state)
		}
	}
	return nil
}

func testAccCheckAmiExists(n string, ami *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("AMI Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMI ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		var resp *ec2.DescribeImagesOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			opts := &ec2.DescribeImagesInput{
				ImageIds: []*string{aws.String(rs.Primary.ID)},
			}
			var err error
			resp, err = conn.DescribeImages(opts)
			if err != nil {
				// This can be just eventual consistency
				if isAWSErr(err, "InvalidAMIID.NotFound", "") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("Unable to find AMI after retries: %s", err)
		}

		if len(resp.Images) == 0 {
			return fmt.Errorf("AMI not found")
		}
		*ami = *resp.Images[0]
		return nil
	}
}

func testAccAmiConfigBase(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 8

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAmiConfigBasic(rName string) string {
	return composeConfig(
		testAccAmiConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName))
}

func testAccAmiConfigDesc(rName, desc string) string {
	return composeConfig(
		testAccAmiConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"
  description         = %[2]q

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName, desc))
}

func testAccAmiConfigEphemeralBlockDevices(rName string) string {
	return composeConfig(
		testAccAmiConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  ephemeral_block_device {
    device_name  = "/dev/sdb"
    virtual_name = "ephemeral0"
  }

  ephemeral_block_device {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral1"
  }
}
`, rName))
}

func testAccAmiConfigGp3BlockDevice(rName string) string {
	return composeConfig(
		testAccAmiConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = false
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  ebs_block_device {
    delete_on_termination = false
    device_name           = "/dev/sdb"
    encrypted             = true
    iops                  = 100
    throughput            = 500
    volume_size           = 10
    volume_type           = "gp3"
  }
}
`, rName))
}

func testAccAmiConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAmiConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAmiConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAmiConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
