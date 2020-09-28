package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsSESIpSendingPool_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIpSendingPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESIpSendingPoolConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESIpSendingPoolExists("aws_ses_ip_sending_pool.test"),
				),
			},
		},
	})
}

func testAccCheckAwsSESIpSendingPoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesv2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_ip_sending_pool" {
			continue
		}

		response, err := conn.ListDedicatedIpPools(&sesv2.ListDedicatedIpPoolsInput{})
		if err != nil {
			return err
		}

		found := false
		for i := range response.DedicatedIpPools {
			if n := *response.DedicatedIpPools[i]; n == "sender" {
				found = true
			}
		}
		if found {
			return errors.New("The sending ip pool still exists")
		}
	}

	return nil
}

func testAccCheckAwsSESIpSendingPoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sesv2Conn

		response, err := conn.ListDedicatedIpPools(&sesv2.ListDedicatedIpPoolsInput{})
		if err != nil {
			return err
		}

		found := false
		for i := range response.DedicatedIpPools {
			if *response.DedicatedIpPools[i] == "sender" {
				found = true
			}
		}

		if !found {
			return errors.New("The sending ip pool was not created")
		}

		return nil
	}
}

const testAccAwsSESIpSendingPoolConfig = `
resource "aws_ses_sending_ip_pool" "test" {
  name = "sender"
}
`
