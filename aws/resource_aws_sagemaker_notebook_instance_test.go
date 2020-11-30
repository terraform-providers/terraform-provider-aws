package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_notebook_instance", &resource.Sweeper{
		Name: "aws_sagemaker_notebook_instance",
		F:    testSweepSagemakerNotebookInstances,
	})
}

func testSweepSagemakerNotebookInstances(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListNotebookInstancesPages(&sagemaker.ListNotebookInstancesInput{}, func(page *sagemaker.ListNotebookInstancesOutput, lastPage bool) bool {
		for _, instance := range page.NotebookInstances {
			name := aws.StringValue(instance.NotebookInstanceName)
			status := aws.StringValue(instance.NotebookInstanceStatus)

			input := &sagemaker.DeleteNotebookInstanceInput{
				NotebookInstanceName: instance.NotebookInstanceName,
			}

			log.Printf("[INFO] Stopping SageMaker Notebook Instance: %s", name)
			if status != sagemaker.NotebookInstanceStatusFailed && status != sagemaker.NotebookInstanceStatusStopped {
				if err := stopSagemakerNotebookInstance(conn, name); err != nil {
					log.Printf("[ERROR] Error stopping SageMaker Notebook Instance (%s): %s", name, err)
					continue
				}
			}

			log.Printf("[INFO] Deleting SageMaker Notebook Instance: %s", name)
			if _, err := conn.DeleteNotebookInstance(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Notebook Instance (%s): %s", name, err)
				continue
			}

			if _, err := waiter.NotebookInstanceDeleted(conn, name); err != nil {
				log.Printf("error waiting for sagemaker notebook instance (%s) to delete: %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Notebook Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Notebook Instances: %s", err)
	}

	return nil
}

func TestAccAWSSagemakerNotebookInstance_basic(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.t2.medium"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
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

func TestAccAWSSagemakerNotebookInstance_update(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.t2.medium"),
				),
			},

			{
				Config: testAccAWSSagemakerNotebookInstanceUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "ml.m4.xlarge"),
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

func TestAccAWSSagemakerNotebookInstance_volumesize(t *testing.T) {
	var notebook1, notebook2, notebook3 sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook1),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "5"),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook2),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "8"),
					testAccCheckAWSSagemakerNotebookInstanceNotRecreated(&notebook1, &notebook2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook3),
					resource.TestCheckResourceAttr(resourceName, "volume_size", "5"),
					testAccCheckAWSSagemakerNotebookInstanceRecreated(&notebook2, &notebook3),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_LifecycleConfigName(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"
	sagemakerLifecycleConfigResourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigLifecycleConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_config_name", ""),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigLifecycleConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "lifecycle_config_name", sagemakerLifecycleConfigResourceName, "name"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_tags(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
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
				Config: testAccAWSSagemakerNotebookInstanceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_kms(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceKMSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "id"),
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

func TestAccAWSSagemakerNotebookInstance_disappears(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerNotebookInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerNotebookInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_notebook_instance" {
			continue
		}

		describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(rs.Primary.ID),
		}
		notebookInstance, err := conn.DescribeNotebookInstance(describeNotebookInput)
		if err != nil {
			return nil
		}

		if aws.StringValue(notebookInstance.NotebookInstanceName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker notebook instance %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerNotebookInstanceExists(n string, notebook *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Notebook Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeNotebookInstance(opts)
		if err != nil {
			return err
		}

		*notebook = *resp

		return nil
	}
}

func testAccCheckAWSSagemakerNotebookInstanceNotRecreated(i, j *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("Sagemaker Notebook Instance was recreated")
		}

		return nil
	}
}

func testAccCheckAWSSagemakerNotebookInstanceRecreated(i, j *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) == aws.TimeValue(j.CreationTime) {
			return errors.New("Sagemaker Notebook Instance was not recreated")
		}

		return nil
	}
}

func TestAccAWSSagemakerNotebookInstance_root_access(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigRootAccess(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Disabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigRootAccess(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "root_access", "Enabled"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_direct_internet_access(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_notebook_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigDirectInternetAccess(rName, "Disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Disabled"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile("eni-.*")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigDirectInternetAccess(rName, "Enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "direct_internet_access", "Enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile("eni-.*")),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_default_code_repository(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigDefaultCodeRepository(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigDefaultCodeRepository(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_additional_code_repositories(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigAdditionalCodeRepository1(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "0"),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigAdditionalCodeRepository2(rName, "https://github.com/hashicorp/terraform-provider-aws.git", "https://github.com/hashicorp/terraform.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform.git"),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigAdditionalCodeRepository1(rName, "https://github.com/hashicorp/terraform-provider-aws.git"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "additional_code_repositories.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "additional_code_repositories.*", "https://github.com/hashicorp/terraform-provider-aws.git"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_default_code_repository_sagemakerRepo(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var resourceName = "aws_sagemaker_notebook_instance.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigDefaultCodeRepositorySageMakerRepo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "default_code_repository", "aws_sagemaker_code_repository.test", "code_repository_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_code_repository", ""),
				),
			},
			{
				Config: testAccAWSSagemakerNotebookInstanceConfigDefaultCodeRepositorySageMakerRepo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "default_code_repository", "aws_sagemaker_code_repository.test", "code_repository_name")),
			},
		},
	})
}

func testAccAWSSagemakerNotebookInstanceBaseConfig(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccAWSSagemakerNotebookInstanceBasicConfig(rName string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
}
`, rName)
}

func testAccAWSSagemakerNotebookInstanceUpdateConfig(rName string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.m4.xlarge"
}
`, rName)
}

func testAccAWSSagemakerNotebookInstanceConfigLifecycleConfigName(rName string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %[1]q
}

resource "aws_sagemaker_notebook_instance" "test" {
  instance_type         = "ml.t2.medium"
  lifecycle_config_name = aws_sagemaker_notebook_instance_lifecycle_configuration.test.name
  name                  = %[1]q
  role_arn              = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSSagemakerNotebookInstanceConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerNotebookInstanceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSagemakerNotebookInstanceConfigRootAccess(rName string, rootAccess string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  root_access   = %[2]q
}
`, rName, rootAccess)
}

func testAccAWSSagemakerNotebookInstanceConfigDirectInternetAccess(rName string, directInternetAccess string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) +
		fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                   = %[1]q
  role_arn               = aws_iam_role.test.arn
  instance_type          = "ml.t2.medium"
  security_groups        = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id
  direct_internet_access = %[2]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, directInternetAccess)
}

func testAccAWSSagemakerNotebookInstanceConfigVolume(rName string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  volume_size   = 8
}
  `, rName)
}

func testAccAWSSagemakerNotebookInstanceConfigDefaultCodeRepository(rName string, defaultCodeRepository string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = %[2]q
}
`, rName, defaultCodeRepository)
}

func testAccAWSSagemakerNotebookInstanceConfigAdditionalCodeRepository1(rName, repo1 string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                         = %[1]q
  role_arn                     = aws_iam_role.test.arn
  instance_type                = "ml.t2.medium"
  additional_code_repositories = ["%[2]s"]
}
`, rName, repo1)
}

func testAccAWSSagemakerNotebookInstanceConfigAdditionalCodeRepository2(rName, repo1, repo2 string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "test" {
  name                         = %[1]q
  role_arn                     = aws_iam_role.test.arn
  instance_type                = "ml.t2.medium"
  additional_code_repositories = ["%[2]s", "%[3]s"]
}
`, rName, repo1, repo2)
}

func testAccAWSSagemakerNotebookInstanceConfigDefaultCodeRepositorySageMakerRepo(rName string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}

resource "aws_sagemaker_notebook_instance" "test" {
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = aws_sagemaker_code_repository.test.code_repository_name
}
`, rName)
}

func testAccAWSSagemakerNotebookInstanceKMSConfig(rName string) string {
	return testAccAWSSagemakerNotebookInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_sagemaker_notebook_instance" "test" {
  name          = %[1]q
  role_arn      = aws_iam_role.test.arn
  instance_type = "ml.t2.medium"
  kms_key_id    = aws_kms_key.test.id
}
`, rName)
}
