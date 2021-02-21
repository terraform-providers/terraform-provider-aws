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

func TestAccAWSStorageGatewayTapePool_basic(t *testing.T) {
	var TapePool storagegateway.PoolInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapePoolBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapePoolExists(resourceName, &TapePool),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`tapepool/pool-.+`)),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_time_in_days", "0"),
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

func TestAccAWSStorageGatewayTapePool_retention(t *testing.T) {
	var TapePool storagegateway.PoolInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapePoolRetentionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapePoolExists(resourceName, &TapePool),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`tapepool/pool-.+`)),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "GLACIER"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_type", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_time_in_days", "1"),
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

func TestAccAWSStorageGatewayTapePool_tags(t *testing.T) {
	var TapePool storagegateway.PoolInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapePoolConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapePoolExists(resourceName, &TapePool),
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
				Config: testAccAWSStorageGatewayTapePoolConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapePoolExists(resourceName, &TapePool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayTapePoolConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapePoolExists(resourceName, &TapePool),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayTapePool_disappears(t *testing.T) {
	var storedIscsiVolume storagegateway.PoolInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_tape_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayTapePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayTapePoolBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayTapePoolExists(resourceName, &storedIscsiVolume),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsStorageGatewayTapePool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSStorageGatewayTapePoolExists(resourceName string, TapePool *storagegateway.PoolInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

		input := &storagegateway.ListTapePoolsInput{
			PoolARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.ListTapePools(input)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway Tape Pool: %s", err)
		}

		if output == nil || len(output.PoolInfos) == 0 || output.PoolInfos[0] == nil || aws.StringValue(output.PoolInfos[0].PoolARN) != rs.Primary.ID {
			return fmt.Errorf("Storage Gateway Tape Pool %q not found", rs.Primary.ID)
		}

		*TapePool = *output.PoolInfos[0]

		return nil
	}
}

func testAccCheckAWSStorageGatewayTapePoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_tape_pool" {
			continue
		}

		input := &storagegateway.ListTapePoolsInput{
			PoolARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.ListTapePools(input)

		if err != nil {
			return err
		}

		if len(output.PoolInfos) != 0 {
			return fmt.Errorf("Storage Gateway Tape Pool %q not found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSStorageGatewayTapePoolBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name     = %[1]q
  storage_class = "GLACIER"
}
`, rName)
}

func testAccAWSStorageGatewayTapePoolRetentionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name                   = %[1]q
  storage_class               = "GLACIER"
  retention_lock_type         = "GOVERNANCE"
  retention_lock_time_in_days = 1
}
`, rName)
}

func testAccAWSStorageGatewayTapePoolConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name     = %[1]q
  storage_class = "GLACIER"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSStorageGatewayTapePoolConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name     = %[1]q
  storage_class = "GLACIER"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
