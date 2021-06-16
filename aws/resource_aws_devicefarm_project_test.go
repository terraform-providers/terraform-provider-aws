package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDeviceFarmProject_basic(t *testing.T) {
	var proj devicefarm.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_devicefarm_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			testAccRegionPreCheck(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   testAccErrorCheck(t, devicefarm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`project:.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFarmProjectConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`project:.+`)),
				),
			},
		},
	})
}

func TestAccAWSDeviceFarmProject_timeout(t *testing.T) {
	var proj devicefarm.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			testAccRegionPreCheck(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   testAccErrorCheck(t, devicefarm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectConfigDefaultJobTimeout(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "default_job_timeout_minutes", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFarmProjectConfigDefaultJobTimeout(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "default_job_timeout_minutes", "20"),
				),
			},
		},
	})
}

func TestAccAWSDeviceFarmProject_tags(t *testing.T) {
	var proj devicefarm.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			testAccRegionPreCheck(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   testAccErrorCheck(t, devicefarm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
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
				Config: testAccDeviceFarmProjectConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeviceFarmProjectConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDeviceFarmProject_disappears(t *testing.T) {
	var proj devicefarm.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(devicefarm.EndpointsID, t)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			testAccRegionPreCheck(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   testAccErrorCheck(t, devicefarm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmProjectExists(resourceName, &proj),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDevicefarmProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFarmProjectExists(n string, v *devicefarm.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).devicefarmconn
		resp, err := conn.GetProject(
			&devicefarm.GetProjectInput{Arn: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if resp.Project == nil {
			return fmt.Errorf("DeviceFarmProject not found")
		}

		*v = *resp.Project

		return nil
	}
}

func testAccCheckDeviceFarmProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).devicefarmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_devicefarm_project" {
			continue
		}

		// Try to find the resource
		resp, err := conn.GetProject(
			&devicefarm.GetProjectInput{Arn: aws.String(rs.Primary.ID)})
		if err == nil {
			if resp.Project != nil {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		if isAWSErr(err, devicefarm.ErrCodeNotFoundException, "") {
			return nil
		}
	}

	return nil
}

func testAccDeviceFarmProjectConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_project" "test" {
  name = %[1]q
}
`, rName)
}

func testAccDeviceFarmProjectConfigDefaultJobTimeout(rName string, timeout int) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_project" "test" {
  name                        = %[1]q
  default_job_timeout_minutes = %[2]d
}
`, rName, timeout)
}

func testAccDeviceFarmProjectConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_project" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDeviceFarmProjectConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_project" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
