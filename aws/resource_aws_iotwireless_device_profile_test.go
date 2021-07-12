package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotwireless"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIotWirelessDeviceProfile_basic(t *testing.T) {
	resourceName := "aws_iotwireless_device_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iotwireless.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotWirelessDeviceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotWirelessDeviceProfileConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotWirelessDeviceProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSIotWirelessDeviceProfile_Tags(t *testing.T) {
	resourceName := "aws_iotwireless_device_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iotwireless.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotWirelessDeviceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotWirelessDeviceProfileConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotWirelessDeviceProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccAWSIotWirelessDeviceProfileConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotWirelessDeviceProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccAWSIotWirelessDeviceProfileConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotWirelessDeviceProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
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

func testAccCheckAWSIotWirelessDeviceProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotwirelessconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotwireless_device_profile" {
			continue
		}

		output, err := conn.GetDeviceProfile(&iotwireless.GetDeviceProfileInput{
			Id: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, iotwireless.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Arn != nil {
			return fmt.Errorf("IoT Wireless Device Profile (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSIotWirelessDeviceProfileExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotwirelessconn

		output, err := conn.GetDeviceProfile(&iotwireless.GetDeviceProfileInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil || output.Arn == nil {
			return fmt.Errorf("IoT Wireless Device Profile (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSIotWirelessDeviceProfileConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iotwireless_device_profile" "test" {
  name = %[1]q

  lorawan {
    mac_version         = "1.0.3"
    reg_params_revision = "Regional Parameters v1.0.3rA"
    max_eirp            = 15
    max_duty_cycle      = 10
    rf_region           = "AU915"
    supports_join       = true
    supports_32bit_fcnt = true
  }
}
`, rName)
}

func testAccAWSIotWirelessDeviceProfileConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iotwireless_device_profile" "test" {
  name = %[1]q

  lorawan {
    mac_version         = "1.0.3"
    reg_params_revision = "Regional Parameters v1.0.3rA"
    max_eirp            = 15
    max_duty_cycle      = 10
    rf_region           = "AU915"
    supports_join       = true
    supports_32bit_fcnt = true
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSIotWirelessDeviceProfileConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iotwireless_device_profile" "test" {
  name = %[1]q

  lorawan {
    mac_version         = "1.0.3"
    reg_params_revision = "Regional Parameters v1.0.3rA"
    max_eirp            = 15
    max_duty_cycle      = 10
    rf_region           = "AU915"
    supports_join       = true
    supports_32bit_fcnt = true
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
