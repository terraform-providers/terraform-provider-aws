package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsGlobalAcceleratorListener_basic(t *testing.T) {
	resourceName := "aws_globalaccelerator_listener.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckGlobalAcceleratorListenerDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorListener_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorListenerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "client_affinity", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "port_range.3309144275.from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "port_range.3309144275.to_port", "81"),
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

func TestAccAwsGlobalAcceleratorListener_update(t *testing.T) {
	resourceName := "aws_globalaccelerator_listener.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckGlobalAcceleratorListenerDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorListener_basic(rInt),
			},
			{
				Config: testAccGlobalAcceleratorListener_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorListenerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "client_affinity", "SOURCE_IP"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "UDP"),
					resource.TestCheckResourceAttr(resourceName, "port_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "port_range.3922064764.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "port_range.3922064764.to_port", "444"),
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

func testAccCheckGlobalAcceleratorListenerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		accelerator, err := resourceAwsGlobalAcceleratorListenerRetrieve(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if accelerator == nil {
			return fmt.Errorf("Global Accelerator listener not found")
		}

		return nil
	}
}

func testAccCheckGlobalAcceleratorListenerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_globalaccelerator_listener" {
			continue
		}

		accelerator, err := resourceAwsGlobalAcceleratorListenerRetrieve(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if accelerator != nil {
			return fmt.Errorf("Global Accelerator listener still exists")
		}
	}
	return nil
}

func testAccGlobalAcceleratorListener_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = "${aws_globalaccelerator_accelerator.example.id}"
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 81
  }
}
`, rInt)
}

func testAccGlobalAcceleratorListener_update(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = "${aws_globalaccelerator_accelerator.example.id}"
  client_affinity = "SOURCE_IP"
  protocol        = "UDP"

  port_range {
    from_port = 443
    to_port   = 444
  }
}
`, rInt)
}
