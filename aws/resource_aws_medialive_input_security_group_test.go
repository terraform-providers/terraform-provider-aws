package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"
)

func TestAccAWSMediaLiveInputSecurityGroup_basic(t *testing.T) {
	var v medialive.DescribeInputSecurityGroupOutput
	resourceName := "aws_medialive_input_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaLiveInputSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMediaLiveInputSecurityGroupConfig("10.0.0.0/8"),
				Check: resource.ComposeTestCheckFunc(
					testAccMediaLiveInputSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.0.cidr", "10.0.0.0/8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsMediaLiveInputSecurityGroupConfig("10.1.0.0/8"),
				Check: resource.ComposeTestCheckFunc(
					testAccMediaLiveInputSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.0.cidr", "10.1.0.0/8"),
				),
			},
			{
				Config: testAccAwsMediaLiveInputSecurityGroupConfigWithMultiple("10.1.0.0/8", "10.2.0.0/8"),
				Check: resource.ComposeTestCheckFunc(
					testAccMediaLiveInputSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.0.cidr", "10.1.0.0/8"),
					resource.TestCheckResourceAttr(
						resourceName, "whitelist_rule.1.cidr", "10.2.0.0/8"),
				),
			},
		},
	})
}

func TestAccAWSMediaLiveInputSecurityGroup_tags(t *testing.T) {
	var v medialive.DescribeInputSecurityGroupOutput
	resourceName := "aws_medialive_input_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaLiveInputSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMediaLiveInputSecurityGroupConfigWithTag("10.0.0.0/8", "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccMediaLiveInputSecurityGroupExists(resourceName, &v),
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
				Config: testAccAwsMediaLiveInputSecurityGroupConfigWithTag2("10.0.0.0/8", "key1", "value1update", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccMediaLiveInputSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1update"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsMediaLiveInputSecurityGroupConfigWithTag("10.0.0.0/8", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccMediaLiveInputSecurityGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsMediaLiveInputSecurityGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_medialive_input_security_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).medialiveconn
		resp, err := conn.DescribeInputSecurityGroup(&medialive.DescribeInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, medialive.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return fmt.Errorf("Error Describing Media Live Input Security Group: %s", err)
		}

		if aws.StringValue(resp.Id) == rs.Primary.ID {
			if aws.StringValue(resp.State) == medialive.InputSecurityGroupStateDeleted {
				continue
			}
			return fmt.Errorf("Media Live Input Security Group %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccMediaLiveInputSecurityGroupExists(n string, v *medialive.DescribeInputSecurityGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Media Live Input Security Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).medialiveconn
		resp, err := conn.DescribeInputSecurityGroup(&medialive.DescribeInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if aws.StringValue(resp.Id) == rs.Primary.ID {
			*v = *resp
			return nil
		}

		return fmt.Errorf("Workspaces IP Group (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsMediaLiveInputSecurityGroupConfig(cidr string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rule {
    cidr = %[1]q
  }
}
`, cidr)
}

func testAccAwsMediaLiveInputSecurityGroupConfigWithMultiple(cidr1, cidr2 string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rule {
    cidr = %[1]q
  }

  whitelist_rule {
    cidr = %[2]q
  }
}
`, cidr1, cidr2)
}

func testAccAwsMediaLiveInputSecurityGroupConfigWithTag(cidr, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rule {
    cidr = %[1]q
  }

  tags = {
    %[2]q = %[3]q  
  }
}
`, cidr, tagKey, tagValue)
}

func testAccAwsMediaLiveInputSecurityGroupConfigWithTag2(cidr, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rule {
    cidr = %[1]q
  }

  tags = {
	%[2]q = %[3]q
	%[4]q = %[5]q 
  }
}
`, cidr, tagKey1, tagValue1, tagKey2, tagValue2)
}
