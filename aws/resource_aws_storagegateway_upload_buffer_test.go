package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/finder"
)

func TestDecodeStorageGatewayUploadBufferID(t *testing.T) {
	var testCases = []struct {
		Input              string
		ExpectedGatewayARN string
		ExpectedDiskID     string
		ErrCount           int
	}{
		{
			Input:              "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0", //lintignore:AWSAT003,AWSAT005
			ExpectedGatewayARN: "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678",                               //lintignore:AWSAT003,AWSAT005
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
			Input:    "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678", //lintignore:AWSAT003,AWSAT005
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
		gatewayARN, diskID, err := decodeStorageGatewayUploadBufferID(tc.Input)
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

func TestAccAWSStorageGatewayUploadBuffer_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_upload_buffer.test"
	localDiskDataSourceName := "data.aws_storagegateway_local_disk.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, storagegateway.EndpointsID),
		Providers:  testAccProviders,
		// Storage Gateway API does not support removing upload buffers,
		// but we want to ensure other resources are removed.
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayUploadBufferConfigDiskId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayUploadBufferExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "disk_id", localDiskDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "disk_path", localDiskDataSourceName, "disk_path"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17809
func TestAccAWSStorageGatewayUploadBuffer_DiskPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_upload_buffer.test"
	localDiskDataSourceName := "data.aws_storagegateway_local_disk.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, storagegateway.EndpointsID),
		Providers:  testAccProviders,
		// Storage Gateway API does not support removing upload buffers,
		// but we want to ensure other resources are removed.
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayUploadBufferConfigDiskPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayUploadBufferExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "disk_id", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "disk_path", localDiskDataSourceName, "disk_path"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
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

func testAccCheckAWSStorageGatewayUploadBufferExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

		gatewayARN, diskID, err := decodeStorageGatewayUploadBufferID(rs.Primary.ID)
		if err != nil {
			return err
		}

		foundDiskID, err := finder.UploadBufferDisk(conn, gatewayARN, diskID)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway Upload Buffer (%s): %w", rs.Primary.ID, err)
		}

		if foundDiskID == nil {
			return fmt.Errorf("Storage Gateway Upload Buffer (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSStorageGatewayUploadBufferConfigDiskId(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_upload_buffer" "test" {
  disk_id     = data.aws_storagegateway_local_disk.test.id
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName)
}

func testAccAWSStorageGatewayUploadBufferConfigDiskPath(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_upload_buffer" "test" {
  disk_path   = data.aws_storagegateway_local_disk.test.disk_path
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName)
}
