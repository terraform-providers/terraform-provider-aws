package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeDeployApp_basic(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			// Import by ID
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by name
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeDeployApp_computePlatform(t *testing.T) {
	var application1, application2 codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigComputePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
				),
			},
			{
				Config: testAccAWSCodeDeployAppConfigComputePlatform(rName, "Server"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application2),
					testAccCheckAWSCodeDeployAppRecreated(&application1, &application2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
				),
			},
		},
	})
}

func TestAccAWSCodeDeployApp_computePlatform_ECS(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigComputePlatform(rName, "ECS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "ECS"),
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

func TestAccAWSCodeDeployApp_computePlatform_Lambda(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigComputePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
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

func TestAccAWSCodeDeployApp_name(t *testing.T) {
	var application1, application2 codedeploy.ApplicationInfo
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigName(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAWSCodeDeployAppConfigName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application2),
					testAccCheckAWSCodeDeployAppRecreated(&application1, &application2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
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

func testAccCheckAWSCodeDeployAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codedeployconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codedeploy_app" {
			continue
		}

		_, err := conn.GetApplication(&codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		})

		if isAWSErr(err, codedeploy.ErrCodeApplicationDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("still exists")
	}

	return nil
}

func testAccCheckAWSCodeDeployAppExists(name string, application *codedeploy.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codedeployconn

		input := &codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}

		output, err := conn.GetApplication(input)

		if err != nil {
			return err
		}

		if output == nil || output.Application == nil {
			return fmt.Errorf("error reading CodeDeploy Application (%s): empty response", rs.Primary.ID)
		}

		*application = *output.Application

		return nil
	}
}

func testAccCheckAWSCodeDeployAppRecreated(i, j *codedeploy.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime) == aws.TimeValue(j.CreateTime) {
			return errors.New("CodeDeploy Application was not recreated")
		}

		return nil
	}
}

func testAccAWSCodeDeployAppConfigComputePlatform(rName string, computePlatform string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  compute_platform = %q
  name             = %q
}
`, computePlatform, rName)
}

func testAccAWSCodeDeployAppConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %q
}
`, rName)
}
