package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// initialize sweeper
func init() {
	resource.AddTestSweepers("aws_beanstalk_application", &resource.Sweeper{
		Name:         "aws_beanstalk_application",
		Dependencies: []string{"aws_beanstalk_environment"},
		F:            testSweepBeanstalkApplications,
	})
}

func testSweepBeanstalkApplications(region string) error {
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
		return fmt.Errorf("Error retrieving beanstalk application: %s", err)
	}

	if len(resp.Applications) == 0 {
		log.Print("[DEBUG] No aws beanstalk applications to sweep")
		return nil
	}

	for _, bsa := range resp.Applications {
		var testOptGroup bool
		for _, testName := range []string{
			"terraform-",
			"tf-test-",
			"tf_acc_",
			"tf-acc-",
		} {
			if strings.HasPrefix(*bsa.ApplicationName, testName) {
				testOptGroup = true
			}
		}

		if !testOptGroup {
			continue
		}

		_, err := beanstalkconn.DeleteApplication(
			&elasticbeanstalk.DeleteApplicationInput{
				ApplicationName: bsa.ApplicationName,
			})
		if err != nil {
			elasticbeanstalkerr, ok := err.(awserr.Error)
			if ok && (elasticbeanstalkerr.Code() == "InvalidConfiguration.NotFound" || elasticbeanstalkerr.Code() == "ValidationError") {
				log.Printf("[DEBUG] beanstalk application (%s) not found", *bsa.ApplicationName)
				return nil
			}

			return err
		}
	}

	return nil
}

func TestAWSElasticBeanstalkApplication_importBasic(t *testing.T) {
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
	return fmt.Sprintf(`resource "aws_elastic_beanstalk_application" "tftest" {
	  name = "%s"
	  description = "tf-test-desc"
	}`, name)
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

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidBeanstalkAppID.NotFound" {
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
  name = "%s"
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
  role = "${aws_iam_role.beanstalk_service.id}"

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
  name = "%s"
  description = "tf-test-desc"
	appversion_lifecycle {
		service_role = "${aws_iam_role.beanstalk_service.arn}"
		max_age_in_days = 90
		delete_source_from_s3 = true
	}
}
`, rName)
}

func testAccBeanstalkAppConfigWithMaxCount(rName string) string {
	return testAccBeanstalkAppServiceRole(rName) + fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "tftest" {
  name = "%s"
  description = "tf-test-desc"
	appversion_lifecycle {
		service_role = "${aws_iam_role.beanstalk_service.arn}"
		max_count = 10
		delete_source_from_s3 = false
	}
}
`, rName)
}
