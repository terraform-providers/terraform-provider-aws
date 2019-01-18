package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSpotInstanceRequest_basic(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testCheckKeyPair(fmt.Sprintf("tmp-key-%d", rInt), &sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "instance_interruption_behaviour", "terminate"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_withLaunchGroup(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_withLaunchGroup(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testCheckKeyPair(fmt.Sprintf("tmp-key-%d", rInt), &sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "launch_group", "terraform-test-group"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_withBlockDuration(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_withBlockDuration(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testCheckKeyPair(fmt.Sprintf("tmp-key-%d", rInt), &sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "block_duration_minutes", "60"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_vpc(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfigVPC(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo_VPC", &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testCheckKeyPair(fmt.Sprintf("tmp-key-%d", rInt), &sir),
					testAccCheckAWSSpotInstanceRequestAttributesVPC(&sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo_VPC", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo_VPC", "spot_request_state", "active"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_validUntil(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()
	validUntil := testAccAWSSpotInstanceRequestValidUntil(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfigValidUntil(rInt, validUntil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequestAttributes(&sir),
					testCheckKeyPair(fmt.Sprintf("tmp-key-%d", rInt), &sir),
					testAccCheckAWSSpotInstanceRequestAttributesValidUntil(&sir, validUntil),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_withoutSpotPrice(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_withoutSpotPrice(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequestAttributesCheckSIRWithoutSpot(&sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_SubnetAndSGAndPublicIpAddress(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_SubnetAndSGAndPublicIpAddress(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequest_InstanceAttributes(&sir, rInt),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "associate_public_ip_address", "true"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_NetworkInterfaceAttributes(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_SubnetAndSGAndPublicIpAddress(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					testAccCheckAWSSpotInstanceRequest_InstanceAttributes(&sir, rInt),
					testAccCheckAWSSpotInstanceRequest_NetworkInterfaceAttributes(&sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "associate_public_ip_address", "true"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequest_getPasswordData(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestConfig_getPasswordData(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					resource.TestCheckResourceAttrSet("aws_spot_instance_request.foo", "password_data"),
				),
			},
		},
	})
}

func testCheckKeyPair(keyName string, sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if sir.LaunchSpecification.KeyName == nil {
			return fmt.Errorf("No Key Pair found, expected(%s)", keyName)
		}
		if sir.LaunchSpecification.KeyName != nil && *sir.LaunchSpecification.KeyName != keyName {
			return fmt.Errorf("Bad key name, expected (%s), got (%s)", keyName, *sir.LaunchSpecification.KeyName)
		}

		return nil
	}
}

func testAccAWSSpotInstanceRequestValidUntil(t *testing.T) string {
	return testAccAWSSpotInstanceRequestTime(t, "12h")
}

func testAccAWSSpotInstanceRequestTime(t *testing.T, duration string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(duration)
	if err != nil {
		t.Fatalf("err parsing time duration: %s", err)
	}
	return n.Add(d).Format(time.RFC3339)
}

func testAccCheckAWSSpotInstanceRequestDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_spot_instance_request" {
			continue
		}

		req := &ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeSpotInstanceRequests(req)
		var s *ec2.SpotInstanceRequest
		if err == nil {
			for _, sir := range resp.SpotInstanceRequests {
				if sir.SpotInstanceRequestId != nil && *sir.SpotInstanceRequestId == rs.Primary.ID {
					s = sir
				}
				continue
			}
		}

		if s == nil {
			// not found
			return nil
		}

		if *s.State == "canceled" || *s.State == "closed" {
			// Requests stick around for a while, so we make sure it's cancelled
			// or closed.
			return nil
		}

		// Verify the error is what we expect
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidSpotInstanceRequestID.NotFound" {
			return err
		}

		// Now check if the associated Spot Instance was also destroyed
		instId := rs.Primary.Attributes["spot_instance_id"]
		instResp, instErr := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(instId)},
		})
		if instErr == nil {
			if len(instResp.Reservations) > 0 {
				return fmt.Errorf("Instance still exists.")
			}

			return nil
		}

		// Verify the error is what we expect
		ec2err, ok = err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidInstanceID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSSpotInstanceRequestExists(
	n string, sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS subscription with that ARN exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		params := &ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{&rs.Primary.ID},
		}
		resp, err := conn.DescribeSpotInstanceRequests(params)

		if err != nil {
			return err
		}

		if v := len(resp.SpotInstanceRequests); v != 1 {
			return fmt.Errorf("Expected 1 request returned, got %d", v)
		}

		*sir = *resp.SpotInstanceRequests[0]

		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributes(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *sir.SpotPrice != "0.050000" {
			return fmt.Errorf("Unexpected spot price: %s", *sir.SpotPrice)
		}
		if *sir.State != "active" {
			return fmt.Errorf("Unexpected request state: %s", *sir.State)
		}
		if *sir.Status.Code != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", *sir.State)
		}
		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributesValidUntil(
	sir *ec2.SpotInstanceRequest, validUntil string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sir.ValidUntil.Format(time.RFC3339) != validUntil {
			return fmt.Errorf("Unexpected valid_until time: %s", sir.ValidUntil.String())
		}
		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributesCheckSIRWithoutSpot(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *sir.State != "active" {
			return fmt.Errorf("Unexpected request state: %s", *sir.State)
		}
		if *sir.Status.Code != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", *sir.State)
		}
		return nil
	}
}

func testAccCheckAWSSpotInstanceRequest_InstanceAttributes(
	sir *ec2.SpotInstanceRequest, rInt int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{sir.InstanceId},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
				return fmt.Errorf("Spot Instance not found")
			}
			return err
		}

		// If nothing was found, then return no state
		if len(resp.Reservations) == 0 {
			return fmt.Errorf("Spot Instance not found")
		}

		instance := resp.Reservations[0].Instances[0]

		var sgMatch bool
		for _, s := range instance.SecurityGroups {
			// Hardcoded name for the security group that should be added inside the
			// VPC
			if *s.GroupName == fmt.Sprintf("tf_test_sg_ssh-%d", rInt) {
				sgMatch = true
			}
		}

		if !sgMatch {
			return fmt.Errorf("Error in matching Spot Instance Security Group, expected 'tf_test_sg_ssh-%d', got %s", rInt, instance.SecurityGroups)
		}

		return nil
	}
}

func testAccCheckAWSSpotInstanceRequest_NetworkInterfaceAttributes(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		return nil
	}
}

func testAccCheckAWSSpotInstanceRequestAttributesVPC(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		ni := nis[0]

		if ni.SubnetId == nil {
			return fmt.Errorf("Expected SubnetId not be non-empty for %s as the instance belongs to a VPC", *sir.InstanceId)
		}
		return nil
	}
}

func TestAccAWSSpotInstanceRequestInterruptStop(t *testing.T) {
	var sir ec2.SpotInstanceRequest

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("stop"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "instance_interruption_behaviour", "stop"),
				),
			},
		},
	})
}

func TestAccAWSSpotInstanceRequestInterruptHibernate(t *testing.T) {
	var sir ec2.SpotInstanceRequest

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotInstanceRequestInterruptConfig("hibernate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotInstanceRequestExists(
						"aws_spot_instance_request.foo", &sir),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_instance_request.foo", "instance_interruption_behaviour", "hibernate"),
				),
			},
		},
	})
}

func testAccAWSSpotInstanceRequestConfig(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_key_pair" "debugging" {
		key_name = "tmp-key-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	}

	resource "aws_spot_instance_request" "foo" {
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		key_name = "${aws_key_pair.debugging.key_name}"

		// base price is $0.044 hourly, so bidding above that should theoretically
		// always fulfill
		spot_price = "0.05"

		// we wait for fulfillment because we want to inspect the launched instance
		// and verify termination behavior
		wait_for_fulfillment = true

	tags = {
			Name = "terraform-test"
		}
	}`, rInt)
}

func testAccAWSSpotInstanceRequestConfigValidUntil(rInt int, validUntil string) string {
	return fmt.Sprintf(`
	resource "aws_key_pair" "debugging" {
		key_name = "tmp-key-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	}

	resource "aws_spot_instance_request" "foo" {
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		key_name = "${aws_key_pair.debugging.key_name}"

		// base price is $0.044 hourly, so bidding above that should theoretically
		// always fulfill
		spot_price = "0.05"

		// The end date and time of the request, the default end date is 7 days from the current date.
		// so 12 hours from the current time will be valid time for valid_until.
		valid_until = "%s"

		// we wait for fulfillment because we want to inspect the launched instance
		// and verify termination behavior
		wait_for_fulfillment = true

	tags = {
			Name = "terraform-test"
		}
	}`, rInt, validUntil)
}

func testAccAWSSpotInstanceRequestConfig_withoutSpotPrice(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_key_pair" "debugging" {
		key_name = "tmp-key-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	}

	resource "aws_spot_instance_request" "foo" {
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		key_name = "${aws_key_pair.debugging.key_name}"

		# no spot price so AWS *should* default max bid to current on-demand price

		# we wait for fulfillment because we want to inspect the launched instance
		# and verify termination behavior
		wait_for_fulfillment = true

	tags = {
			Name = "terraform-test"
		}
	}`, rInt)
}

func testAccAWSSpotInstanceRequestConfig_withLaunchGroup(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_key_pair" "debugging" {
		key_name = "tmp-key-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	}

	resource "aws_spot_instance_request" "foo" {
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		key_name = "${aws_key_pair.debugging.key_name}"

		// base price is $0.044 hourly, so bidding above that should theoretically
		// always fulfill
		spot_price = "0.05"

		// we wait for fulfillment because we want to inspect the launched instance
		// and verify termination behavior
		wait_for_fulfillment = true

		launch_group = "terraform-test-group"

	tags = {
			Name = "terraform-test"
		}
	}`, rInt)
}

func testAccAWSSpotInstanceRequestConfig_withBlockDuration(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_key_pair" "debugging" {
		key_name = "tmp-key-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	}

	resource "aws_spot_instance_request" "foo" {
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		key_name = "${aws_key_pair.debugging.key_name}"

		// base price is $0.044 hourly, so bidding above that should theoretically
		// always fulfill
		spot_price = "0.05"

		// we wait for fulfillment because we want to inspect the launched instance
		// and verify termination behavior
		wait_for_fulfillment = true

		block_duration_minutes = 60

	tags = {
			Name = "terraform-test"
		}
	}`, rInt)
}

func testAccAWSSpotInstanceRequestConfigVPC(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "foo_VPC" {
		cidr_block = "10.1.0.0/16"
	tags = {
			Name = "terraform-testacc-spot-instance-request-vpc"
		}
	}

	resource "aws_subnet" "foo_VPC" {
		cidr_block = "10.1.1.0/24"
		vpc_id = "${aws_vpc.foo_VPC.id}"
	tags = {
			Name = "tf-acc-spot-instance-request-vpc"
		}
	}

	resource "aws_key_pair" "debugging" {
		key_name = "tmp-key-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
	}

	resource "aws_spot_instance_request" "foo_VPC" {
		ami = "ami-4fccb37f"
		instance_type = "m1.small"
		key_name = "${aws_key_pair.debugging.key_name}"

		// base price is $0.044 hourly, so bidding above that should theoretically
		// always fulfill
		spot_price = "0.05"

		// VPC settings
		subnet_id = "${aws_subnet.foo_VPC.id}"

		// we wait for fulfillment because we want to inspect the launched instance
		// and verify termination behavior
		wait_for_fulfillment = true

	tags = {
			Name = "terraform-test-VPC"
		}
	}`, rInt)
}

func testAccAWSSpotInstanceRequestConfig_SubnetAndSGAndPublicIpAddress(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_spot_instance_request" "foo" {
		ami                         = "ami-4fccb37f"
		instance_type               = "m1.small"
		spot_price                  = "0.05"
		wait_for_fulfillment        = true
		subnet_id                   = "${aws_subnet.tf_test_subnet.id}"
		vpc_security_group_ids      = ["${aws_security_group.tf_test_sg_ssh.id}"]
	  associate_public_ip_address = true
	}

	resource "aws_vpc" "default" {
		cidr_block           = "10.0.0.0/16"
		enable_dns_hostnames = true

	tags = {
			Name = "terraform-testacc-spot-instance-request-subnet-and-sg-public-ip"
		}
	}

	resource "aws_subnet" "tf_test_subnet" {
		vpc_id                  = "${aws_vpc.default.id}"
		cidr_block              = "10.0.0.0/24"
		map_public_ip_on_launch = true

	tags = {
			Name = "tf-acc-spot-instance-request-subnet-and-sg-public-ip"
		}
	}

	resource "aws_security_group" "tf_test_sg_ssh" {
		name        = "tf_test_sg_ssh-%d"
		description = "tf_test_sg_ssh"
		vpc_id      = "${aws_vpc.default.id}"

	tags = {
			Name = "tf_test_sg_ssh-%d"
		}
	}`, rInt, rInt)
}

func testAccAWSSpotInstanceRequestConfig_getPasswordData(rInt int) string {
	return fmt.Sprintf(`
	# Find latest Microsoft Windows Server 2016 Core image (Amazon deletes old ones)
	data "aws_ami" "win2016core" {
		most_recent = true

		filter {
			name = "owner-alias"
			values = ["amazon"]
		}

		filter {
			name = "name"
			values = ["Windows_Server-2016-English-Core-Base-*"]
		}
	}

	resource "aws_key_pair" "foo" {
		key_name = "tf-acctest-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAq6U3HQYC4g8WzU147gZZ7CKQH8TgYn3chZGRPxaGmHW1RUwsyEs0nmombmIhwxudhJ4ehjqXsDLoQpd6+c7BuLgTMvbv8LgE9LX53vnljFe1dsObsr/fYLvpU9LTlo8HgHAqO5ibNdrAUvV31ronzCZhms/Gyfdaue88Fd0/YnsZVGeOZPayRkdOHSpqme2CBrpa8myBeL1CWl0LkDG4+YCURjbaelfyZlIApLYKy3FcCan9XQFKaL32MJZwCgzfOvWIMtYcU8QtXMgnA3/I3gXk8YDUJv5P4lj0s/PJXuTM8DygVAUtebNwPuinS7wwonm5FXcWMuVGsVpG5K7FGQ== tf-acc-winpasswordtest"
	}

	resource "aws_spot_instance_request" "foo" {
		ami                  = "${data.aws_ami.win2016core.id}"
		instance_type        = "m1.small"
		spot_price           = "0.05"
		key_name             = "${aws_key_pair.foo.key_name}"
		wait_for_fulfillment = true
		get_password_data    = true
	}
	`, rInt)
}

func testAccAWSSpotInstanceRequestInterruptConfig(interruption_behavior string) string {
	return fmt.Sprintf(`
	resource "aws_spot_instance_request" "foo" {
		ami = "ami-19e92861"
		instance_type = "c5.large"

		// base price is $0.067 hourly, so bidding above that should theoretically
		// always fulfill
		spot_price = "0.07"

		// we wait for fulfillment because we want to inspect the launched instance
		// and verify termination behavior
		wait_for_fulfillment = true

		instance_interruption_behaviour = "%s"
	}`, interruption_behavior)
}
