package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSStorageGatewayStoredIscsiVolume_basic(t *testing.T) {
	var storedIscsiVolume storagegateway.StorediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayStoredIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					resource.TestCheckResourceAttr(resourceName, "preserve_existing_data", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "disk_id", "data.aws_storagegateway_local_disk.test", "id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", "aws_storagegateway_gateway.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", "aws_instance.test", "private_ip"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "target_arn", "storagegateway", regexp.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-+`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "10737418240"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
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

func TestAccAWSStorageGatewayStoredIscsiVolume_kms(t *testing.T) {
	var storedIscsiVolume storagegateway.StorediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayStoredIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigKMSEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key", keyResourceName, "arn"),
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

func TestAccAWSStorageGatewayStoredIscsiVolume_tags(t *testing.T) {
	var storedIscsiVolume storagegateway.StorediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayStoredIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayStoredIscsiVolume_snapshotId(t *testing.T) {
	var storedIscsiVolume storagegateway.StorediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayStoredIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigSnapshotId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", "aws_storagegateway_gateway.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", "aws_instance.test", "private_ip"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_id", "aws_ebs_snapshot.test", "id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "target_arn", "storagegateway", regexp.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-+`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "10737418240"),
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

func TestAccAWSStorageGatewayStoredIscsiVolume_disappears(t *testing.T) {
	var storedIscsiVolume storagegateway.StorediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayStoredIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayStoredIscsiVolumeConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName, &storedIscsiVolume),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsStorageGatewayStoredIscsiVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSStorageGatewayStoredIscsiVolumeExists(resourceName string, storedIscsiVolume *storagegateway.StorediSCSIVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

		input := &storagegateway.DescribeStorediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeStorediSCSIVolumes(input)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway stored iSCSI volume: %w", err)
		}

		if output == nil || len(output.StorediSCSIVolumes) == 0 || output.StorediSCSIVolumes[0] == nil || aws.StringValue(output.StorediSCSIVolumes[0].VolumeARN) != rs.Primary.ID {
			return fmt.Errorf("Storage Gateway stored iSCSI volume %q not found", rs.Primary.ID)
		}

		*storedIscsiVolume = *output.StorediSCSIVolumes[0]

		return nil
	}
}

func testAccCheckAWSStorageGatewayStoredIscsiVolumeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_stored_iscsi_volume" {
			continue
		}

		input := &storagegateway.DescribeStorediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeStorediSCSIVolumes(input)

		if err != nil {
			if isAWSErrStorageGatewayGatewayNotFound(err) {
				return nil
			}
			if isAWSErr(err, storagegateway.ErrorCodeVolumeNotFound, "") {
				return nil
			}
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
				return nil
			}
			return err
		}

		if output != nil && len(output.StorediSCSIVolumes) > 0 && output.StorediSCSIVolumes[0] != nil && aws.StringValue(output.StorediSCSIVolumes[0].VolumeARN) == rs.Primary.ID {
			return fmt.Errorf("Storage Gateway stored iSCSI volume %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSStorageGatewayStoredIscsiVolumeConfigBase(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "buffer" {
  availability_zone = aws_instance.test.availability_zone
  size              = 10
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "buffer" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.buffer.id
}

data "aws_storagegateway_local_disk" "buffer" {
  disk_node   = aws_volume_attachment.buffer.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_working_storage" "buffer" {
  disk_id     = data.aws_storagegateway_local_disk.buffer.id
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = 10
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName)
}

func testAccAWSStorageGatewayStoredIscsiVolumeConfigBasic(rName string) string {
	return testAccAWSStorageGatewayStoredIscsiVolumeConfigBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.disk_id

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName)
}

func testAccAWSStorageGatewayStoredIscsiVolumeConfigKMSEncrypted(rName string) string {
	return testAccAWSStorageGatewayStoredIscsiVolumeConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %[1]s"
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

resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id
  kms_encrypted          = true
  kms_key                = aws_kms_key.test.arn

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName)
}

func testAccAWSStorageGatewayStoredIscsiVolumeConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSStorageGatewayStoredIscsiVolumeConfigBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSStorageGatewayStoredIscsiVolumeConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSStorageGatewayStoredIscsiVolumeConfigBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSStorageGatewayStoredIscsiVolumeConfigSnapshotId(rName string) string {
	return testAccAWSStorageGatewayStoredIscsiVolumeConfigBase(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "snapvolume" {
  availability_zone = aws_instance.test.availability_zone
  size              = 5
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.snapvolume.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  snapshot_id            = aws_ebs_snapshot.test.id
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName)
}
