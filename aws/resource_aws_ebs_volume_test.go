package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ebs_volume", &resource.Sweeper{
		Name: "aws_ebs_volume",
		Dependencies: []string{
			"aws_instance",
		},
		F: testSweepEbsVolumes,
	})
}

func testSweepEbsVolumes(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeVolumesPages(&ec2.DescribeVolumesInput{}, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		for _, volume := range page.Volumes {
			id := aws.StringValue(volume.VolumeId)

			if aws.StringValue(volume.State) != ec2.VolumeStateAvailable {
				log.Printf("[INFO] Skipping unavailable EC2 EBS Volume: %s", id)
				continue
			}

			input := &ec2.DeleteVolumeInput{
				VolumeId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 EBS Volume: %s", id)
			_, err := conn.DeleteVolume(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 EBS Volume (%s): %s", id, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 EBS Volume sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 EBS Volumes: %s", err)
	}

	return nil
}

// testAccErrorCheckSkipEBSVolume skips EBS volume tests that have error messages indicating unsupported features
func testAccErrorCheckSkipEBSVolume(t *testing.T) resource.ErrorCheckFunc {
	return testAccErrorCheckSkipMessagesContaining(t,
		"specified zone does not support multi-attach-enabled volumes",
		"Unsupported volume type",
	)
}

func TestAccAWSEBSVolume_basic(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_updateAttachedEbsVolume(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsAttachedVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsAttachedVolumeConfigUpdateSize,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "20"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateSize(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigUpdateSize,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateType(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigUpdateType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "sc1"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateIops_Io1(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithIopsIo1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigWithIopsIo1Updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "200"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateIops_Io2(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithIopsIo2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigWithIopsIo2Updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "200"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_kmsKey(t *testing.T) {
	var v ec2.Volume
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccAwsEbsVolumeConfigWithKmsKey, ri)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_NoIops(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithNoIops,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
func TestAccAWSEBSVolume_InvalidIopsForType(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEbsVolumeConfigWithInvalidIopsForType,
				ExpectError: regexp.MustCompile(`'iops' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccAWSEBSVolume_InvalidThroughputForType(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEbsVolumeConfigWithInvalidThroughputForType,
				ExpectError: regexp.MustCompile(`'throughput' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccAWSEBSVolume_withTags(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TerraformTest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_multiAttach(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigMultiAttach(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_outpost(t *testing.T) {
	var v ec2.Volume
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigOutpost(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_gp3_basic(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp3", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_gp3_iops(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp3", "4000", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "4000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "200"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp3", "5000", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "5000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "200"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_gp3_throughput(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp3", "", "400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "400"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp3", "", "600"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "600"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_gp3_to_gp2(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp3", "3000", "400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "400"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, "10", "gp2", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_snapshotID(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigSnapshotId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_snapshotIDAndSize(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		ErrorCheck:    testAccErrorCheckSkipEBSVolume(t),
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigSnapshotIdAndSize(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "20"),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEBSVolume_disappears(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEbsVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVolumeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_volume" {
			continue
		}

		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeVolumes(request)

		if isAWSErr(err, "InvalidVolume.NotFound", "") {
			continue
		}

		if err == nil {
			for _, volume := range resp.Volumes {
				if aws.StringValue(volume.VolumeId) == rs.Primary.ID {
					return fmt.Errorf("Volume still exists")
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckVolumeExists(n string, v *ec2.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeVolumes(request)
		if err == nil {
			if response.Volumes != nil && len(response.Volumes) > 0 {
				*v = *response.Volumes[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 volume %s", rs.Primary.ID)
	}
}

const testAccAwsEbsVolumeConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}
`

const testAccAwsEbsAttachedVolumeConfig = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
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

  owners = ["amazon"]
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.debian_jessie_latest.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name = "test-terraform"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  depends_on        = [aws_instance.test]
  availability_zone = aws_instance.test.availability_zone
  type              = "gp2"
  size              = "10"
}

resource "aws_volume_attachment" "test" {
  depends_on  = [aws_ebs_volume.test]
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`

const testAccAwsEbsAttachedVolumeConfigUpdateSize = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
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

  owners = ["amazon"]
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.debian_jessie_latest.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name = "test-terraform"
  }
}

resource "aws_ebs_volume" "test" {
  depends_on        = [aws_instance.test]
  availability_zone = aws_instance.test.availability_zone
  type              = "gp2"
  size              = "20"
}

resource "aws_volume_attachment" "test" {
  depends_on  = [aws_ebs_volume.test]
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`

const testAccAwsEbsVolumeConfigUpdateSize = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 10

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigUpdateType = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "sc1"
  size              = 500

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithIopsIo1 = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 4
  iops              = 100

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithIopsIo1Updated = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 4
  iops              = 200

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithIopsIo2 = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io2"
  size              = 4
  iops              = 100

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithIopsIo2Updated = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io2"
  size              = 4
  iops              = 200

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithKmsKey = `
resource "aws_kms_key" "test" {
  description = "Terraform acc test %d"
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
  kms_key_id        = aws_kms_key.test.arn
}
`

const testAccAwsEbsVolumeConfigWithTags = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccAwsEbsVolumeConfigWithNoIops = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  type              = "gp2"
  iops              = 0

  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccAwsEbsVolumeConfigWithInvalidIopsForType = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  iops              = 100

  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccAwsEbsVolumeConfigWithInvalidThroughputForType = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  iops              = 100
  throughput        = 500
  type              = "io1"

  tags = {
    Name = "TerraformTest"
  }
}
`

func testAccAwsEbsVolumeConfigOutpost() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  size              = 1
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = "tf-acc-volume-outpost"
  }
}
`
}

func testAccAwsEbsVolumeConfigMultiAttach(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  type                 = "io1"
  multi_attach_enabled = true
  size                 = 4
  iops                 = 100

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAwsEbsVolumeConfigSizeTypeIopsThroughput(rName, size, volumeType, iops, throughput string) string {
	if volumeType == "" {
		volumeType = "null"
	}
	if iops == "" {
		iops = "null"
	}
	if throughput == "" {
		throughput = "null"
	}

	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = %[2]s
  type              = %[3]q
  iops              = %[4]s
  throughput        = %[5]s

  tags = {
    Name = %[1]q
  }
}
`, rName, size, volumeType, iops, throughput))
}

func testAccAwsEbsVolumeConfigSnapshotId(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "source" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  snapshot_id       = aws_ebs_snapshot.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAwsEbsVolumeConfigSnapshotIdAndSize(rName string, size int) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "source" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  snapshot_id       = aws_ebs_snapshot.test.id
  size              = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, size))
}
