package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elastic_beanstalk_application", &resource.Sweeper{
		Name:         "aws_elastic_beanstalk_application",
		Dependencies: []string{"aws_elastic_beanstalk_environment"},
		F:            testSweepElasticBeanstalkApplications,
	})
}

func testSweepElasticBeanstalkApplications(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	beanstalkconn := client.(*AWSClient).elasticbeanstalkconn

	resp, err := beanstalkconn.DescribeApplications(&elasticbeanstalk.DescribeApplicationsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elastic Beanstalk Application sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving beanstalk application: %w", err)
	}

	if len(resp.Applications) == 0 {
		log.Print("[DEBUG] No aws beanstalk applications to sweep")
		return nil
	}

	var errors error
	for _, bsa := range resp.Applications {
		applicationName := aws.StringValue(bsa.ApplicationName)
		_, err := beanstalkconn.DeleteApplication(
			&elasticbeanstalk.DeleteApplicationInput{
				ApplicationName: bsa.ApplicationName,
			})
		if err != nil {
			if isAWSErr(err, "InvalidConfiguration.NotFound", "") || isAWSErr(err, "ValidationError", "") {
				log.Printf("[DEBUG] beanstalk application %q not found", applicationName)
				continue
			}

			errors = multierror.Append(fmt.Errorf("error deleting Elastic Beanstalk Application %q: %w", applicationName, err))
		}
	}

	return errors
}

func TestAccAWSElasticBeanstalkApplication_basic(t *testing.T) {
	resourceName := "aws_elastic_beanstalk_application.tftest"
	config := fmt.Sprintf("tf-test-name-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkAppImportConfig(config),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccBeanstalkAppImportConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"
}
`, name)
}

func TestAccAWSBeanstalkApp_basic(t *testing.T) {
	var app elasticbeanstalk.ApplicationDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkAppConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists("aws_elastic_beanstalk_application.tftest", &app),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkApp_appversionlifecycle(t *testing.T) {
	var app elasticbeanstalk.ApplicationDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkAppConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3"),
				),
			},
			{
				Config: testAccBeanstalkAppConfigWithMaxAge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.#", "1"),
					resource.TestCheckResourceAttrPair(
						"aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role",
						"aws_iam_role.beanstalk_service", "arn"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days", "90"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count", "0"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3", "true"),
				),
			},
			{
				Config: testAccBeanstalkAppConfigWithMaxCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.#", "1"),
					resource.TestCheckResourceAttrPair(
						"aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role",
						"aws_iam_role.beanstalk_service", "arn"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days", "0"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count", "10"),
					resource.TestCheckResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3", "false"),
				),
			},
			{
				Config: testAccBeanstalkAppConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists("aws_elastic_beanstalk_application.tftest", &app),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.service_role"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_age_in_days"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.max_count"),
					resource.TestCheckNoResourceAttr("aws_elastic_beanstalk_application.tftest", "appversion_lifecycle.0.delete_source_from_s3"),
				),
			},
		},
	})
}

func TestAccAWSBeanstalkApp_tags(t *testing.T) {
	var app elasticbeanstalk.ApplicationDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elastic_beanstalk_application.tftest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBeanstalkAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkAppConfigWithTags(rName, "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "test1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "test2"),
				),
			},
			{
				Config: testAccBeanstalkAppConfigWithTags(rName, "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
			{
				Config: testAccBeanstalkAppConfigWithAddTags(rName, "updateTest1", "updateTest2", "addTest3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
					resource.TestCheckResourceAttr(resourceName, "tags.thirdTag", "addTest3"),
				),
			},
			{
				Config: testAccBeanstalkAppConfigWithTags(rName, "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBeanstalkAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
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

func testAccCheckBeanstalkAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastic_beanstalk_application" {
			continue
		}

		// Try to find the application
		DescribeBeanstalkAppOpts := &elasticbeanstalk.DescribeApplicationsInput{
			ApplicationNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeApplications(DescribeBeanstalkAppOpts)
		if err == nil {
			if len(resp.Applications) > 0 {
				return fmt.Errorf("Elastic Beanstalk Application still exists.")
			}
			return nil
		}

		if !isAWSErr(err, "InvalidBeanstalkAppID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckBeanstalkAppExists(n string, app *elasticbeanstalk.ApplicationDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk app ID is not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn
		DescribeBeanstalkAppOpts := &elasticbeanstalk.DescribeApplicationsInput{
			ApplicationNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeApplications(DescribeBeanstalkAppOpts)
		if err != nil {
			return err
		}
		if len(resp.Applications) == 0 {
			return fmt.Errorf("Elastic Beanstalk Application not found.")
		}

		*app = *resp.Applications[0]

		return nil
	}
}

func testAccBeanstalkAppConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"
}
`, rName)
}

func testAccBeanstalkAppServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "beanstalk_service" {
  name = "%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticbeanstalk.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "elasticbeanstalk"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "beanstalk_service" {
  name = "%[1]s"
  role = aws_iam_role.beanstalk_service.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowOperations",
            "Effect": "Allow",
            "Action": [
                "iam:PassRole"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
EOF
}
`, rName)
}

func testAccBeanstalkAppConfigWithMaxAge(rName string) string {
	return testAccBeanstalkAppServiceRole(rName) + fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  appversion_lifecycle {
    service_role          = aws_iam_role.beanstalk_service.arn
    max_age_in_days       = 90
    delete_source_from_s3 = true
  }
}
`, rName)
}

func testAccBeanstalkAppConfigWithMaxCount(rName string) string {
	return testAccBeanstalkAppServiceRole(rName) + fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  appversion_lifecycle {
    service_role          = aws_iam_role.beanstalk_service.arn
    max_count             = 10
    delete_source_from_s3 = false
  }
}
`, rName)
}

func testAccBeanstalkAppConfigWithTags(rName, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  tags = {
    firstTag  = "%s"
    secondTag = "%s"
  }
}
`, rName, tag1, tag2)
}

func testAccBeanstalkAppConfigWithAddTags(rName, tag1, tag2, tag3 string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "%s"
  description = "tf-test-desc"

  tags = {
    firstTag  = "%s"
    secondTag = "%s"
    thirdTag  = "%s"
  }
}
`, rName, tag1, tag2, tag3)
}
