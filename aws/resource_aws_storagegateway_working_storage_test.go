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

func TestDecodeStorageGatewayWorkingStorageID(t *testing.T) {
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
		gatewayARN, diskID, err := decodeStorageGatewayWorkingStorageID(tc.Input)
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

func TestAccAWSStorageGatewayWorkingStorage_Basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_working_storage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// Storage Gateway API does not support removing working storages
		// CheckDestroy: testAccCheckAWSStorageGatewayWorkingStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayWorkingStorageConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayWorkingStorageExists(resourceName),
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

func testAccCheckAWSStorageGatewayWorkingStorageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

		gatewayARN, diskID, err := decodeStorageGatewayWorkingStorageID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &storagegateway.DescribeWorkingStorageInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeWorkingStorage(input)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway working storage: %s", err)
		}

		if output == nil || len(output.DiskIds) == 0 {
			return fmt.Errorf("Storage Gateway working storage %q not found", rs.Primary.ID)
		}

		for _, existingDiskID := range output.DiskIds {
			if aws.StringValue(existingDiskID) == diskID {
				return nil
			}
		}

		return fmt.Errorf("Storage Gateway working storage %q not found", rs.Primary.ID)
	}
}

func testAccAWSStorageGatewayWorkingStorageConfig_Basic(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName) + fmt.Sprintf(`
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

resource "aws_storagegateway_working_storage" "test" {
  disk_id     = "${data.aws_storagegateway_local_disk.test.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
`, rName)
}
