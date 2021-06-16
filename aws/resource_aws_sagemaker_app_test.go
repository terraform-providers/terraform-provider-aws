package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_app", &resource.Sweeper{
		Name: "aws_sagemaker_app",
		F:    testSweepSagemakerApps,
	})
}

func testSweepSagemakerApps(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sagemakerconn
	var sweeperErrs *multierror.Error

	err = conn.ListAppsPages(&sagemaker.ListAppsInput{}, func(page *sagemaker.ListAppsOutput, lastPage bool) bool {
		for _, app := range page.Apps {

			if aws.StringValue(app.Status) == sagemaker.AppStatusDeleted {
				continue
			}

			r := resourceAwsSagemakerApp()
			d := r.Data(nil)
			d.SetId(aws.StringValue(app.AppName))
			d.Set("app_name", app.AppName)
			d.Set("app_type", app.AppType)
			d.Set("domain_id", app.DomainId)
			d.Set("user_profile_name", app.UserProfileName)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Sagemaker Apps: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAWSSagemakerApp_basic(t *testing.T) {
	var app sagemaker.DescribeAppOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", "aws_sagemaker_domain.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "user_profile_name", "aws_sagemaker_user_profile.test", "user_profile_name"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
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

func testAccAWSSagemakerApp_resourceSpec(t *testing.T) {
	var app sagemaker.DescribeAppOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppResourceSpecConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_spec.0.instance_type", "system"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_spec.0.sagemaker_image_arn"),
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

func testAccAWSSagemakerApp_tags(t *testing.T) {
	var app sagemaker.DescribeAppOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppExists(resourceName, &app),
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
				Config: testAccAWSSagemakerAppConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerAppConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSSagemakerApp_disappears(t *testing.T) {
	var app sagemaker.DescribeAppOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerAppBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerAppExists(resourceName, &app),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_app" {
			continue
		}

		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]
		appType := rs.Primary.Attributes["app_type"]
		appName := rs.Primary.Attributes["app_name"]

		app, err := finder.AppByName(conn, domainID, userProfileName, appType, appName)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker App (%s): %w", rs.Primary.ID, err)
		}

		appArn := aws.StringValue(app.AppArn)
		if appArn == rs.Primary.ID && aws.StringValue(app.Status) != sagemaker.AppStatusDeleted {
			return fmt.Errorf("SageMaker App %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerAppExists(n string, app *sagemaker.DescribeAppOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		domainID := rs.Primary.Attributes["domain_id"]
		userProfileName := rs.Primary.Attributes["user_profile_name"]
		appType := rs.Primary.Attributes["app_type"]
		appName := rs.Primary.Attributes["app_name"]

		resp, err := finder.AppByName(conn, domainID, userProfileName, appType, appName)
		if err != nil {
			return err
		}

		*app = *resp

		return nil
	}
}

func testAccAWSSagemakerAppConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # Sagemaker compute resources are not available at usw2-az4.
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}

resource "aws_sagemaker_user_profile" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerAppBasicConfig(rName string) string {
	return testAccAWSSagemakerAppConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"
}
`, rName)
}

func testAccAWSSagemakerAppConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerAppConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerAppConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerAppConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSagemakerAppResourceSpecConfig(rName string) string {
	return testAccAWSSagemakerAppConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_app" "test" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = aws_sagemaker_user_profile.test.user_profile_name
  app_name          = %[1]q
  app_type          = "JupyterServer"

  resource_spec {
    instance_type = "system"
  }
}
`, rName)
}
