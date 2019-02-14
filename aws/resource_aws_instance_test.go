package aws

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_instance", &resource.Sweeper{
		Name: "aws_instance",
		F:    testSweepInstances,
	})
}

func testSweepInstances(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeInstancesPages(&ec2.DescribeInstancesInput{}, func(page *ec2.DescribeInstancesOutput, isLast bool) bool {
		if len(page.Reservations) == 0 {
			log.Print("[DEBUG] No EC2 Instances to sweep")
			return false
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				var nameTag string
				id := aws.StringValue(instance.InstanceId)

				if instance.State != nil && aws.StringValue(instance.State.Name) == ec2.InstanceStateNameTerminated {
					log.Printf("[INFO] Skipping terminated EC2 Instance: %s", id)
					continue
				}

				for _, instanceTag := range instance.Tags {
					if aws.StringValue(instanceTag.Key) == "Name" {
						nameTag = aws.StringValue(instanceTag.Value)
						break
					}
				}

				if !strings.HasPrefix(nameTag, "tf-acc-test-") {
					log.Printf("[INFO] Skipping EC2 Instance: %s", id)
					continue
				}

				log.Printf("[INFO] Terminating EC2 Instance: %s", id)
				err := awsTerminateInstance(conn, id, 5*time.Minute)
				if err != nil {
					log.Printf("[ERROR] Error terminating EC2 Instance (%s): %s", id, err)
				}
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving EC2 Instances: %s", err)
	}

	return nil
}

func TestFetchRootDevice(t *testing.T) {
	cases := []struct {
		label  string
		images []*ec2.Image
		name   string
	}{
		{
			"device name in mappings",
			[]*ec2.Image{{
				RootDeviceType: aws.String("ebs"),
				RootDeviceName: aws.String("/dev/xvda"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					{DeviceName: aws.String("/dev/xvdb")},
					{DeviceName: aws.String("/dev/xvda")},
				},
			}},
			"/dev/xvda",
		},
		{
			"device name not in mappings",
			[]*ec2.Image{{
				RootDeviceType: aws.String("ebs"),
				RootDeviceName: aws.String("/dev/xvda"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					{DeviceName: aws.String("/dev/xvdb")},
					{DeviceName: aws.String("/dev/xvdc")},
				},
			}},
			"/dev/xvdb",
		},
		{
			"no images",
			[]*ec2.Image{},
			"",
		},
	}

	sess, err := session.NewSession(nil)
	if err != nil {
		t.Errorf("Error new session: %s", err)
	}

	conn := ec2.New(sess)

	for _, tc := range cases {
		t.Run(fmt.Sprintf(tc.label), func(t *testing.T) {
			conn.Handlers.Clear()
			conn.Handlers.Send.PushBack(func(r *request.Request) {
				data := r.Data.(*ec2.DescribeImagesOutput)
				data.Images = tc.images
			})
			name, err := fetchRootDeviceName("ami-123", conn)
			if err != nil {
				t.Errorf("Error fetching device name: %s", err)
			}
			if tc.name != aws.StringValue(name) {
				t.Errorf("Expected name %s, got %s", tc.name, aws.StringValue(name))
			}
		})
	}
}

func TestAccAWSInstance_importBasic(t *testing.T) {
	resourceName := "aws_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigVPC,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address", "user_data"},
			},
		},
	})
}

func TestAccAWSInstance_importInDefaultVpcBySgName(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInDefaultVpcBySgName(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSInstance_importInDefaultVpcBySgId(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInDefaultVpcBySgId(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSInstance_importInEc2Classic(t *testing.T) {
	resourceName := "aws_instance.foo"
	rInt := acctest.RandInt()

	// EC2 Classic enabled
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInEc2Classic(rInt),
			},

			{
				Config:                  testAccInstanceConfigInEc2Classic(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "source_dest_check"},
			},
		},
	})
}

func TestAccAWSInstance_basic(t *testing.T) {
	var v ec2.Instance
	var vol *ec2.Volume

	rInt := acctest.RandInt()

	testCheck := func(rInt int) func(*terraform.State) error {
		return func(*terraform.State) error {
			if *v.Placement.AvailabilityZone != "us-west-2a" {
				return fmt.Errorf("bad availability zone: %#v", *v.Placement.AvailabilityZone)
			}

			if len(v.SecurityGroups) == 0 {
				return fmt.Errorf("no security groups: %#v", v.SecurityGroups)
			}
			if *v.SecurityGroups[0].GroupName != fmt.Sprintf("tf_test_%d", rInt) {
				return fmt.Errorf("no security groups: %#v", v.SecurityGroups)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Create a volume to cover #1249
			{
				// Need a resource in this config so the provisioner will be available
				Config: testAccInstanceConfig_pre(rInt),
				Check: func(*terraform.State) error {
					conn := testAccProvider.Meta().(*AWSClient).ec2conn
					var err error
					vol, err = conn.CreateVolume(&ec2.CreateVolumeInput{
						AvailabilityZone: aws.String("us-west-2a"),
						Size:             aws.Int64(int64(5)),
					})
					return err
				},
			},

			{
				Config: testAccInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					testCheck(rInt),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"user_data",
						"3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.#", "0"),
					resource.TestMatchResourceAttr(
						"aws_instance.foo",
						"arn",
						regexp.MustCompile(`^arn:[^:]+:ec2:[^:]+:\d{12}:instance/i-.+`)),
				),
			},

			// We repeat the exact same test so that we can be sure
			// that the user data hash stuff is working without generating
			// an incorrect diff.
			{
				Config: testAccInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					testCheck(rInt),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"user_data",
						"3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.#", "0"),
				),
			},

			// Clean up volume created above
			{
				Config: testAccInstanceConfig(rInt),
				Check: func(*terraform.State) error {
					conn := testAccProvider.Meta().(*AWSClient).ec2conn
					_, err := conn.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: vol.VolumeId})
					return err
				},
			},
		},
	})
}

func TestAccAWSInstance_userDataBase64(t *testing.T) {
	var v ec2.Instance

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithUserDataBase64(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"user_data_base64",
						"aGVsbG8gd29ybGQ="),
				),
			},
		},
	})
}

func TestAccAWSInstance_GP2IopsDevice(t *testing.T) {
	var v ec2.Instance

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {

			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/sda1"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sda1")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"ephemeral_block_device", "user_data"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGP2IopsDevice,
				//Config: testAccInstanceConfigBlockDevices,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.iops", "100"),
					testCheck(),
				),
			},
		},
	})
}

func TestAccAWSInstance_GP2WithIopsValue(t *testing.T) {
	var v ec2.Instance
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"ephemeral_block_device", "user_data"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGP2WithIopsValue,
				Check:  testAccCheckInstanceExists("aws_instance.foo", &v),
			},
			{
				Config:             testAccInstanceGP2WithIopsValue,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSInstance_blockDevices(t *testing.T) {
	var v ec2.Instance

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {

			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/sda1"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sda1")
			}

			// Check if the secondary block device exists.
			if _, ok := blockDevices["/dev/sdb"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdb")
			}

			// Check if the third block device exists.
			if _, ok := blockDevices["/dev/sdc"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdc")
			}

			// Check if the encrypted block device exists
			if _, ok := blockDevices["/dev/sdd"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdd")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"ephemeral_block_device"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigBlockDevices,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.#", "1"),
					resource.TestMatchResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2576023345.device_name", "/dev/sdb"),
					resource.TestMatchResourceAttr(
						"aws_instance.foo", "ebs_block_device.2576023345.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2576023345.volume_size", "9"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2576023345.volume_type", "standard"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2554893574.device_name", "/dev/sdc"),
					resource.TestMatchResourceAttr(
						"aws_instance.foo", "ebs_block_device.2554893574.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2554893574.volume_size", "10"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2554893574.volume_type", "io1"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2554893574.iops", "100"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2634515331.device_name", "/dev/sdd"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2634515331.encrypted", "true"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.2634515331.volume_size", "12"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.1692014856.device_name", "/dev/sde"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.1692014856.virtual_name", "ephemeral0"),
					testCheck(),
				),
			},
		},
	})
}

func TestAccAWSInstance_rootInstanceStore(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "aws_instance" "foo" {
						# us-west-2
						# Amazon Linux HVM Instance Store 64-bit (2016.09.0)
						# https://aws.amazon.com/amazon-linux-ami
						ami = "ami-44c36524"

						# Only certain instance types support ephemeral root instance stores.
						# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html
						instance_type = "m3.medium"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ami", "ami-44c36524"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "instance_type", "m3.medium"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSInstance_noAMIEphemeralDevices(t *testing.T) {
	var v ec2.Instance

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {

			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/sda1"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sda1")
			}

			// Check if the secondary block not exists.
			if _, ok := blockDevices["/dev/sdb"]; ok {
				return fmt.Errorf("block device exist: /dev/sdb")
			}

			// Check if the third block device not exists.
			if _, ok := blockDevices["/dev/sdc"]; ok {
				return fmt.Errorf("block device exist: /dev/sdc")
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"ephemeral_block_device"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "aws_instance" "foo" {
						# us-west-2
						ami = "ami-01f05461"  // This AMI (Ubuntu) contains two ephemerals

						instance_type = "c3.large"

						root_block_device {
							volume_type = "gp2"
							volume_size = 11
						}
						ephemeral_block_device {
							device_name = "/dev/sdb"
							no_device = true
						}
						ephemeral_block_device {
							device_name = "/dev/sdc"
							no_device = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ami", "ami-01f05461"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "instance_type", "c3.large"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.172787947.device_name", "/dev/sdb"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.172787947.no_device", "true"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.3336996981.device_name", "/dev/sdc"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "ephemeral_block_device.3336996981.no_device", "true"),
					testCheck(),
				),
			},
		},
	})
}

func TestAccAWSInstance_sourceDestCheck(t *testing.T) {
	var v ec2.Instance

	testCheck := func(enabled bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.SourceDestCheck == nil {
				return fmt.Errorf("bad source_dest_check: got nil")
			}
			if *v.SourceDestCheck != enabled {
				return fmt.Errorf("bad source_dest_check: %#v", *v.SourceDestCheck)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigSourceDestDisable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheck(false),
				),
			},

			{
				Config: testAccInstanceConfigSourceDestEnable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheck(true),
				),
			},

			{
				Config: testAccInstanceConfigSourceDestDisable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheck(false),
				),
			},
		},
	})
}

func TestAccAWSInstance_disableApiTermination(t *testing.T) {
	var v ec2.Instance

	checkDisableApiTermination := func(expected bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			conn := testAccProvider.Meta().(*AWSClient).ec2conn
			r, err := conn.DescribeInstanceAttribute(&ec2.DescribeInstanceAttributeInput{
				InstanceId: v.InstanceId,
				Attribute:  aws.String("disableApiTermination"),
			})
			if err != nil {
				return err
			}
			got := *r.DisableApiTermination.Value
			if got != expected {
				return fmt.Errorf("expected: %t, got: %t", expected, got)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigDisableAPITermination(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					checkDisableApiTermination(true),
				),
			},

			{
				Config: testAccInstanceConfigDisableAPITermination(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					checkDisableApiTermination(false),
				),
			},
		},
	})
}

func TestAccAWSInstance_vpc(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"user_data",
						"562a3e32810edf6ff09994f050f12e799452379d"),
				),
			},
		},
	})
}

func TestAccAWSInstance_placementGroup(t *testing.T) {
	var v ec2.Instance
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPlacementGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"placement_group",
						fmt.Sprintf("testAccInstanceConfigPlacementGroup_%s", rStr)),
				),
			},
		},
	})
}

func TestAccAWSInstance_ipv6_supportAddressCount(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigIpv6Support,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"ipv6_address_count",
						"1"),
				),
			},
		},
	})
}

func TestAccAWSInstance_ipv6AddressCountAndSingleAddressCausesError(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfigIpv6ErrorConfig,
				ExpectError: regexp.MustCompile("Only 1 of `ipv6_address_count` or `ipv6_addresses` can be specified"),
			},
		},
	})
}

func TestAccAWSInstance_ipv6_supportAddressCountWithIpv4(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigIpv6SupportWithIpv4,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo",
						"ipv6_address_count",
						"1"),
				),
			},
		},
	})
}

func TestAccAWSInstance_multipleRegions(t *testing.T) {
	var v ec2.Instance

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckInstanceDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigMultipleRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExistsWithProvider("aws_instance.foo", &v,
						testAccAwsRegionProviderFunc("us-west-2", &providers)),
					testAccCheckInstanceExistsWithProvider("aws_instance.bar", &v,
						testAccAwsRegionProviderFunc("us-east-1", &providers)),
				),
			},
		},
	})
}

func TestAccAWSInstance_NetworkInstanceSecurityGroups(t *testing.T) {
	var v ec2.Instance

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo_instance",
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceNetworkInstanceSecurityGroups(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo_instance", &v),
				),
			},
		},
	})
}

func TestAccAWSInstance_NetworkInstanceRemovingAllSecurityGroups(t *testing.T) {
	var v ec2.Instance

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo_instance",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceNetworkInstanceVPCSecurityGroupIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo_instance", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo_instance", "security_groups.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo_instance", "vpc_security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccInstanceNetworkInstanceVPCRemoveSecurityGroupIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo_instance", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo_instance", "security_groups.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo_instance", "vpc_security_group_ids.#", "1"),
				),
				ExpectError: regexp.MustCompile(`VPC-based instances require at least one security group to be attached`),
			},
		},
	})
}

func TestAccAWSInstance_NetworkInstanceVPCSecurityGroupIDs(t *testing.T) {
	var v ec2.Instance

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo_instance",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceNetworkInstanceVPCSecurityGroupIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.foo_instance", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo_instance", "security_groups.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo_instance", "vpc_security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSInstance_tags(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testAccCheckTags(&v.Tags, "foo", "bar"),
					// Guard against regression of https://github.com/hashicorp/terraform/issues/914
					testAccCheckTags(&v.Tags, "#", ""),
				),
			},
			{
				Config: testAccCheckInstanceConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testAccCheckTags(&v.Tags, "foo", ""),
					testAccCheckTags(&v.Tags, "bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSInstance_volumeTags(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigNoVolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					resource.TestCheckNoResourceAttr(
						"aws_instance.foo", "volume_tags"),
				),
			},
			{
				Config: testAccCheckInstanceConfigWithVolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "volume_tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "volume_tags.Name", "acceptance-test-volume-tag"),
				),
			},
			{
				Config: testAccCheckInstanceConfigWithVolumeTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "volume_tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "volume_tags.Name", "acceptance-test-volume-tag"),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "volume_tags.Environment", "dev"),
				),
			},
			{
				Config: testAccCheckInstanceConfigNoVolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					resource.TestCheckNoResourceAttr(
						"aws_instance.foo", "volume_tags"),
				),
			},
		},
	})
}

func TestAccAWSInstance_volumeTagsComputed(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigWithAttachedVolume,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
				),
			},
		},
	})
}

func TestAccAWSInstance_instanceProfileChange(t *testing.T) {
	var v ec2.Instance
	rName := acctest.RandString(5)

	testCheckInstanceProfile := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.IamInstanceProfile == nil {
				return fmt.Errorf("Instance Profile is nil - we expected an InstanceProfile associated with the Instance")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithoutInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
				),
			},
			{
				Config: testAccInstanceConfigWithInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheckInstanceProfile(),
				),
			},
		},
	})
}

func TestAccAWSInstance_withIamInstanceProfile(t *testing.T) {
	var v ec2.Instance
	rName := acctest.RandString(5)

	testCheckInstanceProfile := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.IamInstanceProfile == nil {
				return fmt.Errorf("Instance Profile is nil - we expected an InstanceProfile associated with the Instance")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheckInstanceProfile(),
				),
			},
		},
	})
}

func TestAccAWSInstance_privateIP(t *testing.T) {
	var v ec2.Instance

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *v.PrivateIpAddress != "10.1.1.42" {
				return fmt.Errorf("bad private IP: %s", *v.PrivateIpAddress)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPrivateIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheckPrivateIP(),
				),
			},
		},
	})
}

func TestAccAWSInstance_associatePublicIPAndPrivateIP(t *testing.T) {
	var v ec2.Instance

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *v.PrivateIpAddress != "10.1.1.42" {
				return fmt.Errorf("bad private IP: %s", *v.PrivateIpAddress)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigAssociatePublicIPAndPrivateIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheckPrivateIP(),
				),
			},
		},
	})
}

// Guard against regression with KeyPairs
// https://github.com/hashicorp/terraform/issues/2302
func TestAccAWSInstance_keyPairCheck(t *testing.T) {
	var v ec2.Instance

	testCheckKeyPair := func(keyName string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.KeyName == nil {
				return fmt.Errorf("No Key Pair found, expected(%s)", keyName)
			}
			if v.KeyName != nil && *v.KeyName != keyName {
				return fmt.Errorf("Bad key name, expected (%s), got (%s)", keyName, *v.KeyName)
			}

			return nil
		}
	}

	keyPairName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_instance.foo",
		IDRefreshIgnore: []string{"source_dest_check"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigKeyPair(keyPairName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					testCheckKeyPair(keyPairName),
				),
			},
		},
	})
}

func TestAccAWSInstance_rootBlockDeviceMismatch(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigRootBlockDeviceMismatch,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_instance.foo", "root_block_device.0.volume_size", "13"),
				),
			},
		},
	})
}

// This test reproduces the bug here:
//   https://github.com/hashicorp/terraform/issues/1752
//
// I wish there were a way to exercise resources built with helper.Schema in a
// unit context, in which case this test could be moved there, but for now this
// will cover the bugfix.
//
// The following triggers "diffs didn't match during apply" without the fix in to
// set NewRemoved on the .# field when it changes to 0.
func TestAccAWSInstance_forceNewAndTagsDrift(t *testing.T) {
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigForceNewAndTagsDrift,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
					driftTags(&v),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccInstanceConfigForceNewAndTagsDrift_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &v),
				),
			},
		},
	})
}

func TestAccAWSInstance_changeInstanceType(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithSmallInstanceType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &before),
				),
			},
			{
				Config: testAccInstanceConfigUpdateInstanceType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &after),
					testAccCheckInstanceNotRecreated(
						t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSInstance_primaryNetworkInterface(t *testing.T) {
	var instance ec2.Instance
	var ini ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPrimaryNetworkInterface,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &instance),
					testAccCheckAWSENIExists("aws_network_interface.bar", &ini),
					resource.TestCheckResourceAttr("aws_instance.foo", "network_interface.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSInstance_primaryNetworkInterfaceSourceDestCheck(t *testing.T) {
	var instance ec2.Instance
	var ini ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPrimaryNetworkInterfaceSourceDestCheck,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &instance),
					testAccCheckAWSENIExists("aws_network_interface.bar", &ini),
					resource.TestCheckResourceAttr("aws_instance.foo", "source_dest_check", "false"),
				),
			},
		},
	})
}

func TestAccAWSInstance_addSecondaryInterface(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	var iniPrimary ec2.NetworkInterface
	var iniSecondary ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigAddSecondaryNetworkInterfaceBefore,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &before),
					testAccCheckAWSENIExists("aws_network_interface.primary", &iniPrimary),
					resource.TestCheckResourceAttr("aws_instance.foo", "network_interface.#", "1"),
				),
			},
			{
				Config: testAccInstanceConfigAddSecondaryNetworkInterfaceAfter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &after),
					testAccCheckAWSENIExists("aws_network_interface.secondary", &iniSecondary),
					resource.TestCheckResourceAttr("aws_instance.foo", "network_interface.#", "1"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/3205
func TestAccAWSInstance_addSecurityGroupNetworkInterface(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigAddSecurityGroupBefore,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &before),
					resource.TestCheckResourceAttr("aws_instance.foo", "vpc_security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccInstanceConfigAddSecurityGroupAfter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &after),
					resource.TestCheckResourceAttr("aws_instance.foo", "vpc_security_group_ids.#", "2"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_defaultPrivate(t *testing.T) {
	var before ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_defaultPrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resName, "public_ip", ""),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_defaultPublic(t *testing.T) {
	var before ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_defaultPublic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resName, "public_ip"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_explicitPublic(t *testing.T) {
	var before ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_explicitPublic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resName, "public_ip"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_explicitPrivate(t *testing.T) {
	var before ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_explicitPrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resName, "public_ip", ""),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_overridePublic(t *testing.T) {
	var before ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_overridePublic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resName, "public_ip"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_overridePrivate(t *testing.T) {
	var before ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_overridePrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resName, "public_ip", ""),
				),
			},
		},
	})
}

func TestAccAWSInstance_getPasswordData_falseToTrue(t *testing.T) {
	var before, after ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "get_password_data", "false"),
					resource.TestCheckResourceAttr(resName, "password_data", ""),
				),
			},
			{
				Config: testAccInstanceConfig_getPasswordData(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &after),
					testAccCheckInstanceNotRecreated(t, &before, &after),
					resource.TestCheckResourceAttr(resName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(resName, "password_data"),
				),
			},
		},
	})
}

func TestAccAWSInstance_getPasswordData_trueToFalse(t *testing.T) {
	var before, after ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(resName, "password_data"),
				),
			},
			{
				Config: testAccInstanceConfig_getPasswordData(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &after),
					testAccCheckInstanceNotRecreated(t, &before, &after),
					resource.TestCheckResourceAttr(resName, "get_password_data", "false"),
					resource.TestCheckResourceAttr(resName, "password_data", ""),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_unspecifiedDefaultsToStandard(t *testing.T) {
	var instance ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_standardCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &first),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &second),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_unlimitedCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &first),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &second),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_updateCpuCredits(t *testing.T) {
	var first, second, third ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &first),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &second),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &third),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_isNotAppliedToNonBurstable(t *testing.T) {
	var instance ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_isNotAppliedToNonBurstable(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &instance),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_unspecifiedDefaultsToUnlimited(t *testing.T) {
	var instance ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_standardCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &first),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &second),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_unlimitedCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &first),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &second),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_updateCpuCredits(t *testing.T) {
	var first, second, third ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &first),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &second),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &third),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_standardCpuCredits_t2Tot3Taint(t *testing.T) {
	var before, after ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "standard"),
				),
				Taint: []string{resName},
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_unlimitedCpuCredits_t2Tot3Taint(t *testing.T) {
	var before, after ec2.Instance
	resName := "aws_instance.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
				Taint: []string{resName},
			},
		},
	})
}

func TestAccAWSInstance_disappears(t *testing.T) {
	var conf ec2.Instance
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("aws_instance.foo", &conf),
					testAccCheckInstanceDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSInstance_UserData_EmptyStringToUnspecified(t *testing.T) {
	var instance ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_UserData_EmptyString(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			// Switching should show no difference
			{
				Config:             testAccInstanceConfig_UserData_Unspecified(rInt),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccAWSInstance_UserData_UnspecifiedToEmptyString(t *testing.T) {
	var instance ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_UserData_Unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			// Switching should show no difference
			{
				Config:             testAccInstanceConfig_UserData_EmptyString(rInt),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func testAccCheckInstanceNotRecreated(t *testing.T,
	before, after *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.InstanceId != *after.InstanceId {
			t.Fatalf("AWS Instance IDs have changed. Before %s. After %s", *before.InstanceId, *after.InstanceId)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	return testAccCheckInstanceDestroyWithProvider(s, testAccProvider)
}

func testAccCheckInstanceDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_instance" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			for _, r := range resp.Reservations {
				for _, i := range r.Instances {
					if i.State != nil && *i.State.Name != "terminated" {
						return fmt.Errorf("Found unterminated instance: %s", i)
					}
				}
			}
		}

		// Verify the error is what we want
		if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidInstanceID.NotFound" {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckInstanceExists(n string, i *ec2.Instance) resource.TestCheckFunc {
	return testAccCheckInstanceExistsWithProvider(n, i, func() *schema.Provider { return testAccProvider })
}

func testAccCheckInstanceExistsWithProvider(n string, i *ec2.Instance, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		provider := providerF()

		conn := provider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.Reservations) > 0 {
			*i = *resp.Reservations[0].Instances[0]
			return nil
		}

		return fmt.Errorf("Instance not found")
	}
}

func testAccCheckInstanceDisappears(conf *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		params := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{conf.InstanceId},
		}

		if _, err := conn.TerminateInstances(params); err != nil {
			return err
		}

		return waitForInstanceDeletion(conn, *conf.InstanceId, 10*time.Minute)
	}
}

func TestInstanceTenancySchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["tenancy"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ForceNew: true,
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func TestInstanceHostIDSchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["host_id"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ForceNew: true,
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func TestInstanceCpuCoreCountSchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["cpu_core_count"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
		ForceNew: true,
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func TestInstanceCpuThreadsPerCoreSchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["cpu_threads_per_core"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
		ForceNew: true,
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func driftTags(instance *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err := conn.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{instance.InstanceId},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Drift"),
					Value: aws.String("Happens"),
				},
			},
		})
		return err
	}
}

func testAccInstanceConfigInDefaultVpcBySgName(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_vpc" "default" {
	default = true
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
	vpc_id = "${data.aws_vpc.default.id}"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "t2.micro"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}

func testAccInstanceConfigInDefaultVpcBySgId(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_vpc" "default" {
	default = true
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
	vpc_id = "${data.aws_vpc.default.id}"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.sg.id}"]
}
`, rInt)
}

func testAccInstanceConfigInEc2Classic(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["paravirtual"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "sg" {
  name = "tf_acc_test_%d"
  description = "Test security group"
}

resource "aws_instance" "foo" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "m3.medium"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}

func testAccInstanceConfig_pre(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_test_foo" {
	name = "tf_test_%d"
	description = "foo"

	ingress {
		protocol = "icmp"
		from_port = -1
		to_port = -1
		cidr_blocks = ["0.0.0.0/0"]
	}
}
`, rInt)
}

func testAccInstanceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_test_foo" {
	name = "tf_test_%d"
	description = "foo"

	ingress {
		protocol = "icmp"
		from_port = -1
		to_port = -1
		cidr_blocks = ["0.0.0.0/0"]
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-4fccb37f"
	availability_zone = "us-west-2a"

	instance_type = "m1.small"
	security_groups = ["${aws_security_group.tf_test_foo.name}"]
	user_data = "foo:-with-character's"
}
`, rInt)
}

func testAccInstanceConfigWithUserDataBase64(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_test_foo" {
	name = "tf_test_%d"
	description = "foo"

	ingress {
		protocol = "icmp"
		from_port = -1
		to_port = -1
		cidr_blocks = ["0.0.0.0/0"]
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-4fccb37f"
	availability_zone = "us-west-2a"

	instance_type = "m1.small"
	security_groups = ["${aws_security_group.tf_test_foo.name}"]
	user_data_base64 = "${base64encode("hello world")}"
}
`, rInt)
}

const testAccInstanceConfigWithSmallInstanceType = `
resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-55a7ea65"
	availability_zone = "us-west-2a"

	instance_type = "m3.medium"

	tags = {
	    Name = "tf-acctest"
	}
}
`

const testAccInstanceConfigUpdateInstanceType = `
resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-55a7ea65"
	availability_zone = "us-west-2a"

	instance_type = "m3.large"

	tags = {
	    Name = "tf-acctest"
	}
}
`

const testAccInstanceGP2IopsDevice = `
resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-55a7ea65"

	# In order to attach an encrypted volume to an instance you need to have an
	# m3.medium or larger. See "Supported Instance Types" in:
	# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
	}
}
`

const testAccInstanceGP2WithIopsValue = `
resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-55a7ea65"

	# In order to attach an encrypted volume to an instance you need to have an
	# m3.medium or larger. See "Supported Instance Types" in:
	# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
        # configured explicitly
		iops        = 10
	}
}
`

const testAccInstanceConfigBlockDevices = `
resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-55a7ea65"

	# In order to attach an encrypted volume to an instance you need to have an
	# m3.medium or larger. See "Supported Instance Types" in:
	# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
	}
	ebs_block_device {
		device_name = "/dev/sdb"
		volume_size = 9
	}
	ebs_block_device {
		device_name = "/dev/sdc"
		volume_size = 10
		volume_type = "io1"
		iops = 100
	}

	# Encrypted ebs block device
	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}
}
`

const testAccInstanceConfigSourceDestEnable = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-source-dest-enable"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-source-dest-enable"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.foo.id}"
}
`

const testAccInstanceConfigSourceDestDisable = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-source-dest-disable"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-source-dest-disable"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.foo.id}"
	source_dest_check = false
}
`

func testAccInstanceConfigDisableAPITermination(val bool) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "foo" {
		cidr_block = "10.1.0.0/16"
	tags = {
			Name = "terraform-testacc-instance-disable-api-termination"
		}
	}

	resource "aws_subnet" "foo" {
		cidr_block = "10.1.1.0/24"
		vpc_id = "${aws_vpc.foo.id}"
	tags = {
			Name = "tf-acc-instance-disable-api-termination"
		}
	}

	resource "aws_instance" "foo" {
		# us-west-2
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		subnet_id = "${aws_subnet.foo.id}"
		disable_api_termination = %t
	}
	`, val)
}

const testAccInstanceConfigVPC = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-vpc"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-vpc"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.foo.id}"
	associate_public_ip_address = true
	tenancy = "dedicated"
	# pre-encoded base64 data
	user_data = "3dc39dda39be1205215e776bad998da361a5955d"
}
`

func testAccInstanceConfigPlacementGroup(rStr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-placement-group"
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
  	Name = "tf-acc-instance-placement-group"
  }
}

resource "aws_placement_group" "foo" {
  name = "testAccInstanceConfigPlacementGroup_%s"
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "foo" {
  # us-west-2
  ami = "ami-55a7ea65"
  instance_type = "c3.large"
  subnet_id = "${aws_subnet.foo.id}"
  associate_public_ip_address = true
  placement_group = "${aws_placement_group.foo.name}"
  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"
}
`, rStr)
}

const testAccInstanceConfigIpv6ErrorConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-instance-ipv6-err"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.foo.ipv6_cidr_block, 8, 1)}"
	tags = {
		Name = "tf-acc-instance-ipv6-err"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.foo.id}"
	ipv6_addresses = ["2600:1f14:bb2:e501::10"]
	ipv6_address_count = 1
	tags = {
		Name = "tf-ipv6-instance-acc-test"
	}
}
`

const testAccInstanceConfigIpv6Support = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-instance-ipv6-support"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.foo.ipv6_cidr_block, 8, 1)}"
	tags = {
		Name = "tf-acc-instance-ipv6-support"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.foo.id}"

	ipv6_address_count = 1
	tags = {
		Name = "tf-ipv6-instance-acc-test"
	}
}
`

const testAccInstanceConfigIpv6SupportWithIpv4 = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-instance-ipv6-support-with-ipv4"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.foo.ipv6_cidr_block, 8, 1)}"
	tags = {
		Name = "tf-acc-instance-ipv6-support-with-ipv4"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.foo.id}"

	associate_public_ip_address = true
	ipv6_address_count = 1
	tags = {
		Name = "tf-ipv6-instance-acc-test"
	}
}
`

const testAccInstanceConfigMultipleRegions = `
provider "aws" {
	alias = "west"
	region = "us-west-2"
}

provider "aws" {
	alias = "east"
	region = "us-east-1"
}

resource "aws_instance" "foo" {
	# us-west-2
	provider = "aws.west"
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
}

resource "aws_instance" "bar" {
	# us-east-1
	provider = "aws.east"
	ami = "ami-8c6ea9e4"
	instance_type = "m1.small"
}
`

const testAccCheckInstanceConfigTags = `
resource "aws_instance" "foo" {
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	tags = {
		foo = "bar"
	}
}
`

const testAccCheckInstanceConfigWithAttachedVolume = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["debian-jessie-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  owners = ["379101102735"] # Debian
}

resource "aws_instance" "foo" {
  ami           = "${data.aws_ami.debian_jessie_latest.id}"
  instance_type = "t2.medium"

  root_block_device {
    delete_on_termination = true
    volume_size           = "10"
    volume_type           = "standard"
  }

  tags = {
    Name    = "test-terraform"
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.foo.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = "test-terraform"
  }
}

resource "aws_volume_attachment" "test" {
  device_name = "/dev/xvdg"
  volume_id   = "${aws_ebs_volume.test.id}"
  instance_id = "${aws_instance.foo.id}"
}
`

const testAccCheckInstanceConfigNoVolumeTags = `
resource "aws_instance" "foo" {
	ami = "ami-55a7ea65"

	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
	}
	ebs_block_device {
		device_name = "/dev/sdb"
		volume_size = 9
	}
	ebs_block_device {
		device_name = "/dev/sdc"
		volume_size = 10
		volume_type = "io1"
		iops = 100
	}

	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}
}
`

const testAccCheckInstanceConfigWithVolumeTags = `
resource "aws_instance" "foo" {
	ami = "ami-55a7ea65"

	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
	}
	ebs_block_device {
		device_name = "/dev/sdb"
		volume_size = 9
	}
	ebs_block_device {
		device_name = "/dev/sdc"
		volume_size = 10
		volume_type = "io1"
		iops = 100
	}

	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}

	volume_tags {
		Name = "acceptance-test-volume-tag"
	}
}
`

const testAccCheckInstanceConfigWithVolumeTagsUpdate = `
resource "aws_instance" "foo" {
	ami = "ami-55a7ea65"

	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
	}
	ebs_block_device {
		device_name = "/dev/sdb"
		volume_size = 9
	}
	ebs_block_device {
		device_name = "/dev/sdc"
		volume_size = 10
		volume_type = "io1"
		iops = 100
	}

	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}

	volume_tags {
		Name = "acceptance-test-volume-tag"
		Environment = "dev"
	}
}
`

const testAccCheckInstanceConfigTagsUpdate = `
resource "aws_instance" "foo" {
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	tags = {
		bar = "baz"
	}
}
`

func testAccInstanceConfigWithoutInstanceProfile(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
	name = "test-%s"
	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_instance" "foo" {
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	tags = {
		bar = "baz"
	}
}`, rName)
}

func testAccInstanceConfigWithInstanceProfile(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
	name = "test-%s"
	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_instance_profile" "test" {
	name = "test-%s"
	roles = ["${aws_iam_role.test.name}"]
}

resource "aws_instance" "foo" {
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	iam_instance_profile = "${aws_iam_instance_profile.test.name}"
	tags = {
		bar = "baz"
	}
}`, rName, rName)
}

const testAccInstanceConfigPrivateIP = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-private-ip"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-private-ip"
	}
}

resource "aws_instance" "foo" {
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.foo.id}"
	private_ip = "10.1.1.42"
}
`

const testAccInstanceConfigAssociatePublicIPAndPrivateIP = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-public-ip-and-private-ip"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-public-ip-and-private-ip"
	}
}

resource "aws_instance" "foo" {
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.foo.id}"
	associate_public_ip_address = true
	private_ip = "10.1.1.42"
}
`

func testAccInstanceNetworkInstanceSecurityGroups(rInt int) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-network-security-groups"
	}
}

resource "aws_security_group" "tf_test_foo" {
  name = "tf_test_%d"
  description = "foo"
  vpc_id="${aws_vpc.foo.id}"

  ingress {
    protocol = "icmp"
    from_port = -1
    to_port = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
  	Name = "tf-acc-instance-network-security-groups"
  }
}

resource "aws_instance" "foo_instance" {
  ami = "ami-21f78e11"
  instance_type = "t1.micro"
  vpc_security_group_ids = ["${aws_security_group.tf_test_foo.id}"]
  subnet_id = "${aws_subnet.foo.id}"
  associate_public_ip_address = true
	depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_eip" "foo_eip" {
  instance = "${aws_instance.foo_instance.id}"
  vpc = true
	depends_on = ["aws_internet_gateway.gw"]
}
`, rInt)
}

func testAccInstanceNetworkInstanceVPCSecurityGroupIDs(rInt int) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-network-vpc-sg-ids"
	}
}

resource "aws_security_group" "tf_test_foo" {
  name = "tf_test_%d"
  description = "foo"
  vpc_id="${aws_vpc.foo.id}"

  ingress {
    protocol = "icmp"
    from_port = -1
    to_port = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
  	Name = "tf-acc-instance-network-vpc-sg-ids"
  }
}

resource "aws_instance" "foo_instance" {
  ami = "ami-21f78e11"
  instance_type = "t1.micro"
  vpc_security_group_ids = ["${aws_security_group.tf_test_foo.id}"]
  subnet_id = "${aws_subnet.foo.id}"
	depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_eip" "foo_eip" {
  instance = "${aws_instance.foo_instance.id}"
  vpc = true
	depends_on = ["aws_internet_gateway.gw"]
}
`, rInt)
}

func testAccInstanceNetworkInstanceVPCRemoveSecurityGroupIDs(rInt int) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
        Name = "terraform-testacc-instance-network-vpc-sg-ids"
    }
}

resource "aws_security_group" "tf_test_foo" {
  name = "tf_test_%d"
  description = "foo"
  vpc_id="${aws_vpc.foo.id}"

  ingress {
    protocol = "icmp"
    from_port = -1
    to_port = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
  	Name = "tf-acc-instance-network-vpc-sg-ids"
  }
}

resource "aws_instance" "foo_instance" {
  ami = "ami-21f78e11"
  instance_type = "t1.micro"
  vpc_security_group_ids = []
  subnet_id = "${aws_subnet.foo.id}"
    depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_eip" "foo_eip" {
  instance = "${aws_instance.foo_instance.id}"
  vpc = true
    depends_on = ["aws_internet_gateway.gw"]
}
`, rInt)
}

func testAccInstanceConfigKeyPair(keyPairName string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

resource "aws_key_pair" "debugging" {
	key_name = "%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_instance" "foo" {
  ami = "ami-408c7f28"
  instance_type = "t1.micro"
  key_name = "${aws_key_pair.debugging.key_name}"
	tags = {
		Name = "testAccInstanceConfigKeyPair_TestAMI"
	}
}
`, keyPairName)
}

const testAccInstanceConfigRootBlockDeviceMismatch = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-root-block-device-mismatch"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-root-block-device-mismatch"
	}
}

resource "aws_instance" "foo" {
	// This is an AMI with RootDeviceName: "/dev/sda1"; actual root: "/dev/sda"
	ami = "ami-ef5b69df"
	instance_type = "t1.micro"
	subnet_id = "${aws_subnet.foo.id}"
	root_block_device {
		volume_size = 13
	}
}
`

const testAccInstanceConfigForceNewAndTagsDrift = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-force-new-and-tags-drift"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-force-new-and-tags-drift"
	}
}

resource "aws_instance" "foo" {
	ami = "ami-22b9a343"
	instance_type = "t2.nano"
	subnet_id = "${aws_subnet.foo.id}"
}
`

const testAccInstanceConfigForceNewAndTagsDrift_Update = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-force-new-and-tags-drift"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-instance-force-new-and-tags-drift"
	}
}

resource "aws_instance" "foo" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.foo.id}"
}
`

const testAccInstanceConfigPrimaryNetworkInterface = `
resource "aws_vpc" "foo" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-primary-network-iface"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-primary-network-iface"
  }
}

resource "aws_network_interface" "bar" {
  subnet_id = "${aws_subnet.foo.id}"
  private_ips = ["172.16.10.100"]
  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_instance" "foo" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.bar.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigPrimaryNetworkInterfaceSourceDestCheck = `
resource "aws_vpc" "foo" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-primary-network-iface-source-dest-check"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-primary-network-iface-source-dest-check"
  }
}

resource "aws_network_interface" "bar" {
  subnet_id = "${aws_subnet.foo.id}"
  private_ips = ["172.16.10.100"]
  source_dest_check = false
  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_instance" "foo" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.bar.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigAddSecondaryNetworkInterfaceBefore = `
resource "aws_vpc" "foo" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-add-secondary-network-iface"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-add-secondary-network-iface"
  }
}

resource "aws_network_interface" "primary" {
  subnet_id = "${aws_subnet.foo.id}"
  private_ips = ["172.16.10.100"]
  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_network_interface" "secondary" {
  subnet_id = "${aws_subnet.foo.id}"
  private_ips = ["172.16.10.101"]
  tags = {
    Name = "secondary_network_interface"
  }
}

resource "aws_instance" "foo" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.primary.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigAddSecondaryNetworkInterfaceAfter = `
resource "aws_vpc" "foo" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-add-secondary-network-iface"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-add-secondary-network-iface"
  }
}

resource "aws_network_interface" "primary" {
  subnet_id = "${aws_subnet.foo.id}"
  private_ips = ["172.16.10.100"]
  tags = {
    Name = "primary_network_interface"
  }
}

// Attach previously created network interface, observe no state diff on instance resource
resource "aws_network_interface" "secondary" {
  subnet_id = "${aws_subnet.foo.id}"
  private_ips = ["172.16.10.101"]
  tags = {
    Name = "secondary_network_interface"
  }
  attachment {
    instance = "${aws_instance.foo.id}"
    device_index = 1
  }
}

resource "aws_instance" "foo" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.primary.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigAddSecurityGroupBefore = `
resource "aws_vpc" "foo" {
    cidr_block = "172.16.0.0/16"
  tags = {
        Name = "terraform-testacc-instance-add-security-group"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "172.16.11.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-bar"
    }
}

resource "aws_security_group" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  description = "foo"
  name = "foo"
}

resource "aws_security_group" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
  description = "bar"
  name = "bar"
}

resource "aws_instance" "foo" {
    ami = "ami-c5eabbf5"
    instance_type = "t2.micro"
    subnet_id = "${aws_subnet.bar.id}"
    associate_public_ip_address = false
    vpc_security_group_ids = [
      "${aws_security_group.foo.id}"
    ]
  tags = {
        Name = "foo-instance-sg-add-test"
    }
}

resource "aws_network_interface" "bar" {
    subnet_id = "${aws_subnet.foo.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.foo.id}"]
    attachment {
        instance = "${aws_instance.foo.id}"
        device_index = 1
    }
  tags = {
        Name = "bar_interface"
    }
}
`

const testAccInstanceConfigAddSecurityGroupAfter = `
resource "aws_vpc" "foo" {
    cidr_block = "172.16.0.0/16"
  tags = {
        Name = "terraform-testacc-instance-add-security-group"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "172.16.11.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-bar"
    }
}

resource "aws_security_group" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  description = "foo"
  name = "foo"
}

resource "aws_security_group" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
  description = "bar"
  name = "bar"
}

resource "aws_instance" "foo" {
    ami = "ami-c5eabbf5"
    instance_type = "t2.micro"
    subnet_id = "${aws_subnet.bar.id}"
    associate_public_ip_address = false
    vpc_security_group_ids = [
      "${aws_security_group.foo.id}",
      "${aws_security_group.bar.id}"
    ]
  tags = {
        Name = "foo-instance-sg-add-test"
    }
}

resource "aws_network_interface" "bar" {
    subnet_id = "${aws_subnet.foo.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.foo.id}"]
    attachment {
        instance = "${aws_instance.foo.id}"
        device_index = 1
    }
  tags = {
        Name = "bar_interface"
    }
}
`

func testAccInstanceConfig_associatePublic_defaultPrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-associate-public-default-private"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = false
  tags = {
    Name = "tf-acc-instance-associate-public-default-private"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.public_subnet.id}"
  tags = {
    Name = "tf-acctest-%d"
  }
}`, rInt)
}

func testAccInstanceConfig_associatePublic_defaultPublic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-associate-public-default-public"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = true
  tags = {
    Name = "tf-acc-instance-associate-public-default-public"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.public_subnet.id}"
  tags = {
    Name = "tf-acctest-%d"
  }
}`, rInt)
}

func testAccInstanceConfig_associatePublic_explicitPublic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-associate-public-explicit-public"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = true
  tags = {
    Name = "tf-acc-instance-associate-public-explicit-public"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = true
  tags = {
    Name = "tf-acctest-%d"
  }
}`, rInt)
}

func testAccInstanceConfig_associatePublic_explicitPrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-associate-public-explicit-private"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = false
  tags = {
    Name = "tf-acc-instance-associate-public-explicit-private"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = false
  tags = {
    Name = "tf-acctest-%d"
  }
}`, rInt)
}

func testAccInstanceConfig_associatePublic_overridePublic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-associate-public-override-public"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = false
  tags = {
    Name = "tf-acc-instance-associate-public-override-public"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = true
  tags = {
    Name = "tf-acctest-%d"
  }
}`, rInt)
}

func testAccInstanceConfig_associatePublic_overridePrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-associate-public-override-private"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = true
  tags = {
    Name = "tf-acc-instance-associate-public-override-private"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = false
  tags = {
    Name = "tf-acctest-%d"
  }
}`, rInt)
}

func testAccInstanceConfig_getPasswordData(val bool, rInt int) string {
	return fmt.Sprintf(`
	# Find latest Microsoft Windows Server 2016 Core image (Amazon deletes old ones)
	data "aws_ami" "win2016core" {
		most_recent = true

		filter {
			name = "owner-alias"
			values = ["amazon"]
		}

		filter {
			name = "name"
			values = ["Windows_Server-2016-English-Core-Base-*"]
		}
	}

	resource "aws_key_pair" "foo" {
		key_name = "tf-acctest-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAq6U3HQYC4g8WzU147gZZ7CKQH8TgYn3chZGRPxaGmHW1RUwsyEs0nmombmIhwxudhJ4ehjqXsDLoQpd6+c7BuLgTMvbv8LgE9LX53vnljFe1dsObsr/fYLvpU9LTlo8HgHAqO5ibNdrAUvV31ronzCZhms/Gyfdaue88Fd0/YnsZVGeOZPayRkdOHSpqme2CBrpa8myBeL1CWl0LkDG4+YCURjbaelfyZlIApLYKy3FcCan9XQFKaL32MJZwCgzfOvWIMtYcU8QtXMgnA3/I3gXk8YDUJv5P4lj0s/PJXuTM8DygVAUtebNwPuinS7wwonm5FXcWMuVGsVpG5K7FGQ== tf-acc-winpasswordtest"
	}

	resource "aws_instance" "foo" {
		ami = "${data.aws_ami.win2016core.id}"
		instance_type = "t2.medium"
		key_name = "${aws_key_pair.foo.key_name}"

		get_password_data = %t
	}
	`, rInt, val)
}

func testAccInstanceConfig_creditSpecification_unspecified(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.my_subnet.id}"
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unspecified_t3(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-51537029" # us-west-2
  instance_type = "t3.micro"
  subnet_id = "${aws_subnet.my_subnet.id}"
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.my_subnet.id}"
  credit_specification {
    cpu_credits = "standard"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-51537029" # us-west-2
  instance_type = "t3.micro"
  subnet_id = "${aws_subnet.my_subnet.id}"
  credit_specification {
    cpu_credits = "standard"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.my_subnet.id}"
  credit_specification {
    cpu_credits = "unlimited"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-51537029" # us-west-2
  instance_type = "t3.micro"
  subnet_id = "${aws_subnet.my_subnet.id}"
  credit_specification {
    cpu_credits = "unlimited"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_isNotAppliedToNonBurstable(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id = "${aws_vpc.my_vpc.id}"
  cidr_block = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "foo" {
  ami = "ami-22b9a343" # us-west-2
  instance_type = "m1.small"
  subnet_id = "${aws_subnet.my_subnet.id}"
  credit_specification {
    cpu_credits = "standard"
  }
}
`, rInt)
}

func testAccInstanceConfig_UserData_Base(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "test" {
  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "172.16.0.0/24"

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt, rInt)
}

func testAccInstanceConfig_UserData_Unspecified(rInt int) string {
	return testAccInstanceConfig_UserData_Base(rInt) + `
resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.test.id}"
}
`
}

func testAccInstanceConfig_UserData_EmptyString(rInt int) string {
	return testAccInstanceConfig_UserData_Base(rInt) + `
resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.test.id}"
  user_data     = ""
}
`
}
