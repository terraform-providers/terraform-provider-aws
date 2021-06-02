package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_apprunner_service", &resource.Sweeper{
		Name: "aws_apprunner_service",
		F:    testSweepAppRunnerServices,
	})
}

func testSweepAppRunnerServices(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).apprunnerconn
	sweepResources := make([]*testSweepResource, 0)
	ctx := context.Background()
	var errs *multierror.Error

	input := &apprunner.ListServicesInput{}

	err = conn.ListServicesPagesWithContext(ctx, input, func(page *apprunner.ListServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, service := range page.ServiceSummaryList {
			if service == nil {
				continue
			}

			arn := aws.StringValue(service.ServiceArn)

			log.Printf("[INFO] Deleting App Runner Service: %s", arn)

			r := resourceAwsAppRunnerService()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Services: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Services for %s: %w", region, err))
	}

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Services sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func TestAccAwsAppRunnerService_ImageRepository_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`service/%s/.+`, rName))),
					testAccMatchResourceAttrRegionalARN(resourceName, "auto_scaling_configuration_arn", "apprunner", regexp.MustCompile(`autoscalingconfiguration/DefaultConfiguration/1/.+`)),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", apprunner.HealthCheckProtocolTcp),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.path", "/"),
					// Only check the following attribute values for health_check and instance configurations
					// are set as their defaults differ in the API documentation and API itself
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.interval"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.timeout"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.healthy_threshold"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.unhealthy_threshold"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.cpu"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.memory"),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttrSet(resourceName, "service_url"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.auto_deployments_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_identifier", "public.ecr.aws/jg/hello:latest"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_repository_type", apprunner.ImageRepositoryTypeEcrPublic),
					resource.TestCheckResourceAttr(resourceName, "status", apprunner.ServiceStatusRunning),
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

func TestAccAwsAppRunnerService_ImageRepository_AutoScalingConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"
	autoScalingResourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
				),
			},
			{
				Config: testAccAppRunnerService_imageRepository_autoScalingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_configuration_arn", autoScalingResourceName, "arn"),
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

func TestAccAwsAppRunnerService_ImageRepository_EncryptionConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository_encryptionConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation; EncryptionConfiguration (or lack thereof) Forces New resource
				Config: testAccAppRunnerService_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAwsAppRunnerService_ImageRepository_HealthCheckConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository_healthCheckConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", apprunner.HealthCheckProtocolTcp),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.timeout", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.unhealthy_threshold", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation; HealthConfiguration Forces New resource
				Config: testAccAppRunnerService_imageRepository_updateHealthCheckConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", apprunner.HealthCheckProtocolTcp),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.unhealthy_threshold", "4"),
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

func TestAccAwsAppRunnerService_ImageRepository_InstanceConfiguration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository_instanceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "1024"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "3072"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppRunnerService_imageRepository_updateInstanceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "2048"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "4096"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppRunnerService_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.cpu"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.memory"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19469
func TestAccAwsAppRunnerService_ImageRepository_RuntimeEnvironmentVars(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository_runtimeEnvVars(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.APP_NAME", rName),
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

func TestAccAwsAppRunnerService_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerService_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppRunnerService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppRunnerService_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerServiceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
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
				Config: testAccAppRunnerServiceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppRunnerServiceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsAppRunnerServiceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_service" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeServiceWithContext(context.Background(), input)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Service != nil && aws.StringValue(output.Service.Status) != apprunner.ServiceStatusDeleted {
			return fmt.Errorf("App Runner Service (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsAppRunnerServiceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeServiceWithContext(context.Background(), input)

		if err != nil {
			return err
		}

		if output == nil || output.Service == nil {
			return fmt.Errorf("App Runner Service (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckAppRunner(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).apprunnerconn
	ctx := context.Background()

	input := &apprunner.ListServicesInput{}

	_, err := conn.ListServicesWithContext(ctx, input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAppRunnerService_imageRepository(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccAppRunnerService_imageRepository_runtimeEnvVars(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
        runtime_environment_variables = {
          APP_NAME = %[1]q
        }
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccAppRunnerService_imageRepository_autoScalingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_service" "test" {
  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.test.arn

  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccAppRunnerService_imageRepository_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = %[1]q
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  encryption_configuration {
    kms_key = aws_kms_key.test.arn
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccAppRunnerService_imageRepository_healthCheckConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold = 2
    timeout           = 5
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccAppRunnerService_imageRepository_updateHealthCheckConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold   = 2
    timeout             = 10
    unhealthy_threshold = 4
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccAppRunnerIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "tasks.apprunner.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
`, rName)
}

func testAccAppRunnerService_imageRepository_instanceConfiguration(rName string) string {
	return composeConfig(
		testAccAppRunnerIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "1 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "3 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccAppRunnerService_imageRepository_updateInstanceConfiguration(rName string) string {
	return composeConfig(
		testAccAppRunnerIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "2 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "4 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccAppRunnerServiceConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppRunnerServiceConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/jg/hello:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
