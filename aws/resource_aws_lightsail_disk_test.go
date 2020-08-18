package aws

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLightsailDisk_basic(t *testing.T) {
	var conf lightsail.Disk
	diskName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSLightsailDisk(t) },
		IDRefreshName: "aws_lightsail_disk.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDiskConfig_basic(diskName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDiskExists("aws_lightsail_disk.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "created_at"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "name"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "size"),
					resource.TestCheckResourceAttr("aws_lightsail_disk.test", "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDisk_Tags(t *testing.T) {
	var conf lightsail.Disk
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		IDRefreshName: "aws_lightsail_disk.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLightsailDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDiskConfig_tags1(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDiskExists("aws_lightsail_disk.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "created_at"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "name"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "size"),
					resource.TestCheckResourceAttr("aws_lightsail_disk.test", "tags.%", "1"),
				),
			},
			{
				Config: testAccAWSLightsailDiskConfig_tags2(lightsailName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDiskExists("aws_lightsail_disk.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "availability_zone"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "created_at"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "name"),
					resource.TestCheckResourceAttrSet("aws_lightsail_disk.test", "size"),
					resource.TestCheckResourceAttr("aws_lightsail_disk.test", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccAWSLightsailDisk_disapear(t *testing.T) {
	var conf lightsail.Disk
	lightsailName := fmt.Sprintf("tf-test-lightsail-%d", acctest.RandInt())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Disk
		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		_, err := conn.DeleteDisk(&lightsail.DeleteDiskInput{
			DiskName: aws.String(lightsailName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Disk in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSLightsail(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDiskConfig_basic(lightsailName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailDiskExists("aws_lightsail_disk.test", &conf),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLightsailDiskExists(n string, res *lightsail.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailDisk ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respDisk, err := conn.GetDisk(&lightsail.GetDiskInput{
			DiskName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respDisk == nil || respDisk.Disk == nil {
			return fmt.Errorf("Disk (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *respDisk.Disk
		return nil
	}
}

func testAccCheckAWSLightsailDiskDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_disk" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn

		respDisk, err := conn.GetDisk(&lightsail.GetDiskInput{
			DiskName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err == nil {
			if respDisk.Disk != nil {
				return fmt.Errorf("LightsailDisk %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccPreCheckAWSLightsailDisk(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).lightsailconn

	input := &lightsail.GetDisksInput{}

	_, err := conn.GetDisks(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSLightsailDiskConfig_basic(diskName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  	state = "available"
}

resource "aws_lightsail_disk" "test" {
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
	name = "%s"
	size = 8
}

`, diskName)
}

func testAccAWSLightsailDiskConfig_tags1(diskName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  	state = "available"
}

resource "aws_lightsail_disk" "test" {
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
	name = "%s"
	size = 8
	  tags = {
		Name = "tf-test"
	  }
}

`, diskName)
}

func testAccAWSLightsailDiskConfig_tags2(diskName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_lightsail_disk" "test" {
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
	name = "%s"
	size = 8
	  tags = {
		Name = "tf-test"	
    	ExtraName = "tf-test"
  }
}
`, diskName)
}
