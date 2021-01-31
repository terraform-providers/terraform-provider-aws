package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEc2TrafficMirrorFilter_basic(t *testing.T) {
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorFilter(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test Disable DNS service
			{
				Config: testAccTrafficMirrorFilterConfigWithoutDNS(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, "network_services"),
				),
			},
			// Test Enable DNS service
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "1"),
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

func TestAccAWSEc2TrafficMirrorFilter_tags(t *testing.T) {
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorFilter(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficMirrorFilterConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
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
				Config: testAccTrafficMirrorFilterConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTrafficMirrorFilterConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEc2TrafficMirrorFilter_disappears(t *testing.T) {
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TrafficMirrorFilter(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TrafficMirrorFilterExists(resourceName, &v),
					testAccCheckAWSEc2TrafficMirrorFilterDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEc2TrafficMirrorFilterExists(name string, traffic *ec2.TrafficMirrorFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return fmt.Errorf("Traffic mirror filter %s not found", rs.Primary.ID)
		}

		*traffic = *out.TrafficMirrorFilters[0]

		return nil
	}
}

func testAccCheckAWSEc2TrafficMirrorFilterDisappears(traffic *ec2.TrafficMirrorFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err := conn.DeleteTrafficMirrorFilter(&ec2.DeleteTrafficMirrorFilterInput{
			TrafficMirrorFilterId: traffic.TrafficMirrorFilterId,
		})

		return err
	}
}

func testAccTrafficMirrorFilterConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  description = "%s"

  network_services = ["amazon-dns"]
}
`, description)
}

func testAccTrafficMirrorFilterConfigWithoutDNS(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  description = "%s"
}
`, description)
}

func testAccTrafficMirrorFilterConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccTrafficMirrorFilterConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPreCheckAWSEc2TrafficMirrorFilter(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	_, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{})

	if testAccPreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckAWSEc2TrafficMirrorFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_filter" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if isAWSErr(err, "InvalidTrafficMirrorFilterId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(out.TrafficMirrorFilters) != 0 {
			return fmt.Errorf("Traffic mirror filter %s still not destroyed", rs.Primary.ID)
		}
	}

	return nil
}
