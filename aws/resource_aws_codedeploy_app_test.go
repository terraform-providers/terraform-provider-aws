package aws

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_codedeploy_app", &resource.Sweeper{
		Name: "aws_codedeploy_app",
		F:    testSweepCodeDeployApps,
	})
}

func testSweepCodeDeployApps(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).codedeployconn
	input := &codedeploy.ListApplicationsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListApplicationsPages(input, func(page *codedeploy.ListApplicationsOutput, lastPage bool) bool {
		for _, app := range page.Applications {
			if app == nil {
				continue
			}

			appName := aws.StringValue(app)
			r := resourceAwsCodeDeployApp()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", "xxxx", appName))
			err = r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeDeploy Application (%s): %w", appName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeDeploy Application sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeDeploy Applications: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCodeDeployApp_basic(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codedeploy", fmt.Sprintf(`application:%s`, rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "linked_to_github", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
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
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
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

func TestAccAWSCodeDeployApp_tags(t *testing.T) {
	var application codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application),
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
				Config: testAccAWSCodeDeployAppConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCodeDeployAppConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCodeDeployApp_disappears(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, codedeploy.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeDeployAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeDeployAppConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeDeployAppExists(resourceName, &application1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeDeployApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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
		if aws.TimeValue(i.CreateTime).Equal(aws.TimeValue(j.CreateTime)) {
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

func testAccAWSCodeDeployAppConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSCodeDeployAppConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
