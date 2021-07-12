package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcsAccountDefaultSetting_containerInstanceLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "containerInstanceLongArnFormat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					testAccMatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsAccountDefaultSetting_serviceLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "serviceLongArnFormat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					testAccMatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsAccountDefaultSetting_taskLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "taskLongArnFormat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					testAccMatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsAccountDefaultSetting_awsvpcTrunking(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "awsvpcTrunking"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					testAccMatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsAccountDefaultSetting_containerInsights(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "containerInsights"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					testAccMatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSEcsDefaultSettingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_account_setting_default" {
			continue
		}

		name := rs.Primary.Attributes["name"]

		input := &ecs.ListAccountSettingsInput{
			Name:              aws.String(name),
			EffectiveSettings: aws.Bool(true),
		}

		resp, err := conn.ListAccountSettings(input)

		if isAWSErr(err, ecs.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		for _, value := range resp.Settings {
			if aws.StringValue(value.Value) != "disabled" {
				return fmt.Errorf("[Destroy Error] Account Settings (%s) Still enabled", aws.StringValue(value.Name))
			}
		}

	}

	return nil
}

func testAccAWSEcsDefaultSettingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
	name = %q
    value = "enabled"
}
`, rName)
}
