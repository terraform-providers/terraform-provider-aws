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

func TestDecodeStorageGatewayCacheID(t *testing.T) {
	var testCases = []struct {
		Input              string
		ExpectedGatewayARN string
		ExpectedDiskID     string
		ErrCount           int
	}{
		{
			Input:              "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0",
			ExpectedGatewayARN: "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678",
			ExpectedDiskID:     "pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount:           0,
		},
		{
			Input:    "sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount: 1,
		},
		{
			Input:    "example:pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678",
			ErrCount: 1,
		},
		{
			Input:    "pci-0000:03:00.0-scsi-0:0:0:0",
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
		gatewayARN, diskID, err := decodeStorageGatewayCacheID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if gatewayARN != tc.ExpectedGatewayARN {
			t.Fatalf("expected %q to return Gateway ARN %q, received: %s", tc.Input, tc.ExpectedGatewayARN, gatewayARN)
		}
		if diskID != tc.ExpectedDiskID {
			t.Fatalf("expected %q to return Disk ID %q, received: %s", tc.Input, tc.ExpectedDiskID, diskID)
		}
	}
}

func TestAccAWSStorageGatewayCache_FileGateway(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_cache.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Storage Gateway API does not support removing caches
		// CheckDestroy: testAccCheckAWSStorageGatewayCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayCacheConfig_FileGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayCacheExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "disk_id"),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:gateway/sgw-.+$`)),
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

func TestAccAWSStorageGatewayCache_TapeAndVolumeGateway(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_cache.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Storage Gateway API does not support removing caches
		// CheckDestroy: testAccCheckAWSStorageGatewayCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayCacheConfig_TapeAndVolumeGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayCacheExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "disk_id"),
					resource.TestMatchResourceAttr(resourceName, "gateway_arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:gateway/sgw-.+$`)),
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

func testAccCheckAWSStorageGatewayCacheExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

		gatewayARN, diskID, err := decodeStorageGatewayCacheID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &storagegateway.DescribeCacheInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeCache(input)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway cache: %s", err)
		}

		if output == nil || len(output.DiskIds) == 0 {
			return fmt.Errorf("Storage Gateway cache %q not found", rs.Primary.ID)
		}

		for _, existingDiskID := range output.DiskIds {
			if aws.StringValue(existingDiskID) == diskID {
				return nil
			}
		}

		return fmt.Errorf("Storage Gateway cache %q not found", rs.Primary.ID)
	}
}

func testAccAWSStorageGatewayCacheConfig_FileGateway(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cache" "test" {
  disk_id     = "${data.aws_storagegateway_local_disk.test.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
`, rName)
}

func testAccAWSStorageGatewayCacheConfig_TapeAndVolumeGateway(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
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
`, rName)
}
