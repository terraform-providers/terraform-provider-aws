package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestParseStorageGatewayVolumeGatewayARNAndTargetNameFromARN(t *testing.T) {
	var testCases = []struct {
		Input              string
		ExpectedGatewayARN string
		ExpectedTargetName string
		ErrCount           int
	}{
		{
			Input:              "arn:aws:storagegateway:us-east-2:111122223333:gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName",
			ExpectedGatewayARN: "arn:aws:storagegateway:us-east-2:111122223333:gateway/sgw-12A3456B",
			ExpectedTargetName: "TargetName",
			ErrCount:           0,
		},
		{
			Input:    "gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:storagegateway:us-east-2:111122223333:target/iqn.1997-05.com.amazon:TargetName",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678",
			ErrCount: 1,
		},
		{
			Input:    "TargetName",
			ErrCount: 1,
		},
		{
			Input:    "gateway/sgw-12345678",
			ErrCount: 1,
		},
		{
			Input:    "sgw-12345678",
			ErrCount: 1,
		},
	}

	for _, tc := range testCases {
		gatewayARN, targetName, err := parseStorageGatewayVolumeGatewayARNAndTargetNameFromARN(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if gatewayARN != tc.ExpectedGatewayARN {
			t.Fatalf("expected %q to return Gateway ARN %q, received: %s", tc.Input, tc.ExpectedGatewayARN, gatewayARN)
		}
		if targetName != tc.ExpectedTargetName {
			t.Fatalf("expected %q to return Disk ID %q, received: %s", tc.Input, tc.ExpectedTargetName, targetName)
		}
	}
}

func TestAccAWSStorageGatewayCachedIscsiVolume_Basic(t *testing.T) {
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayCachedIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayCachedIscsiVolumeConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayCachedIscsiVolumeExists(resourceName, &cachedIscsiVolume),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", "false"),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestMatchResourceAttr(resourceName, "target_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:storagegateway:[^:]+:\\d{12}:gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "volume_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "5368709120"),
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

func TestAccAWSStorageGatewayCachedIscsiVolume_SnapshotId(t *testing.T) {
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayCachedIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayCachedIscsiVolumeConfig_SnapshotId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayCachedIscsiVolumeExists(resourceName, &cachedIscsiVolume),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", "false"),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "snapshot_id", regexp.MustCompile(`^snap-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "target_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:storagegateway:[^:]+:\\d{12}:gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "volume_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "5368709120"),
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

func TestAccAWSStorageGatewayCachedIscsiVolume_SourceVolumeArn(t *testing.T) {
	t.Skip("This test can cause Storage Gateway 2.0.10.0 to enter an irrecoverable state during volume deletion.")
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayCachedIscsiVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayCachedIscsiVolumeConfig_SourceVolumeArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayCachedIscsiVolumeExists(resourceName, &cachedIscsiVolume),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestMatchResourceAttr(resourceName, "source_volume_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "target_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:storagegateway:[^:]+:\\d{12}:gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "volume_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:\d{12}:gateway/sgw-.+/volume/vol-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "1073741824"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_volume_arn"},
			},
		},
	})
}

func testAccCheckAWSStorageGatewayCachedIscsiVolumeExists(resourceName string, cachedIscsiVolume *storagegateway.CachediSCSIVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

		input := &storagegateway.DescribeCachediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeCachediSCSIVolumes(input)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway cached iSCSI volume: %s", err)
		}

		if output == nil || len(output.CachediSCSIVolumes) == 0 || output.CachediSCSIVolumes[0] == nil || aws.StringValue(output.CachediSCSIVolumes[0].VolumeARN) != rs.Primary.ID {
			return fmt.Errorf("Storage Gateway cached iSCSI volume %q not found", rs.Primary.ID)
		}

		*cachedIscsiVolume = *output.CachediSCSIVolumes[0]

		return nil
	}
}

func testAccCheckAWSStorageGatewayCachedIscsiVolumeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_cached_iscsi_volume" {
			continue
		}

		input := &storagegateway.DescribeCachediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeCachediSCSIVolumes(input)

		if err != nil {
			if isAWSErrStorageGatewayGatewayNotFound(err) {
				return nil
			}
			if isAWSErr(err, storagegateway.ErrorCodeVolumeNotFound, "") {
				return nil
			}
			return err
		}

		if output != nil && len(output.CachediSCSIVolumes) > 0 && output.CachediSCSIVolumes[0] != nil && aws.StringValue(output.CachediSCSIVolumes[0].VolumeARN) == rs.Primary.ID {
			return fmt.Errorf("Storage Gateway cached iSCSI volume %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSStorageGatewayCachedIscsiVolumeConfig_Basic(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = 10
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cache" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "0b68f77a-709b-4c79-ad9d-d7728014b291" => "/dev/xvdc" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.
  lifecycle {
    ignore_changes = ["disk_id"]
  }

  disk_id     = "${data.aws_storagegateway_local_disk.test.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = "${aws_storagegateway_cache.test.gateway_arn}"
  network_interface_id = "${aws_instance.test.private_ip}"
  target_name          = %q
  volume_size_in_bytes = 5368709120
}
`, rName, rName)
}

func testAccAWSStorageGatewayCachedIscsiVolumeConfig_SnapshotId(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "cachevolume" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = 10
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.cachevolume.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cache" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "0b68f77a-709b-4c79-ad9d-d7728014b291" => "/dev/xvdc" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.
  lifecycle {
    ignore_changes = ["disk_id"]
  }

  disk_id     = "${data.aws_storagegateway_local_disk.test.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_ebs_volume" "snapvolume" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = 5
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.snapvolume.id}"

  tags = {
    Name = %q
  }
}

resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = "${aws_storagegateway_cache.test.gateway_arn}"
  network_interface_id = "${aws_instance.test.private_ip}"
  snapshot_id          = "${aws_ebs_snapshot.test.id}"
  target_name          = %q
  volume_size_in_bytes = "${aws_ebs_snapshot.test.volume_size * 1024 * 1024 * 1024}"
}
`, rName, rName, rName, rName)
}

func testAccAWSStorageGatewayCachedIscsiVolumeConfig_SourceVolumeArn(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName) + fmt.Sprintf(`
data "aws_storagegateway_local_disk" "uploadbuffer" {
  disk_path   = "/dev/xvdb"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_upload_buffer" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "0b68f77a-709b-4c79-ad9d-d7728014b291" => "/dev/xvdc" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.
  lifecycle {
    ignore_changes = ["disk_id"]
  }

  disk_id     = "${data.aws_storagegateway_local_disk.uploadbuffer.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = 10
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cache" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "0b68f77a-709b-4c79-ad9d-d7728014b291" => "/dev/xvdc" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.
  lifecycle {
    ignore_changes = ["disk_id"]
  }

  disk_id     = "${data.aws_storagegateway_local_disk.test.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cached_iscsi_volume" "source" {
  gateway_arn          = "${aws_storagegateway_cache.test.gateway_arn}"
  network_interface_id = "${aws_instance.test.private_ip}"
  target_name          = "%s-source"
  volume_size_in_bytes = 1073741824
}

resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = "${aws_storagegateway_cache.test.gateway_arn}"
  network_interface_id = "${aws_instance.test.private_ip}"
  source_volume_arn    = "${aws_storagegateway_cached_iscsi_volume.source.arn}"
  target_name          = %q
  volume_size_in_bytes = "${aws_storagegateway_cached_iscsi_volume.source.volume_size_in_bytes}"
}
`, rName, rName, rName)
}
