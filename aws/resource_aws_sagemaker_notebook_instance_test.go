package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix = "terraform-testacc-"

func TestAccAWSSagemakerNotebookInstance_basic(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					testAccCheckAWSSagemakerNotebookInstanceName(&notebook, notebookName),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "name", notebookName),
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
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "instance_type", "ml.t2.medium"),
				),
			},

			{
				Config: testAccAWSSagemakerNotebookInstanceUpdateConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "instance_type", "ml.m4.xlarge"),
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

func TestAccAWSSagemakerNotebookInstance_tags(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceTagsConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),
					testAccCheckAWSSagemakerNotebookInstanceTags(&notebook, "foo", "bar"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_notebook_instance.foo", "name", notebookName),
					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.foo", "bar"),
				),
			},

			{
				Config: testAccAWSSagemakerNotebookInstanceTagsUpdateConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists("aws_sagemaker_notebook_instance.foo", &notebook),
					testAccCheckAWSSagemakerNotebookInstanceTags(&notebook, "foo", ""),
					testAccCheckAWSSagemakerNotebookInstanceTags(&notebook, "bar", "baz"),

					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_notebook_instance.foo", "tags.bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerNotebookInstance_disappears(t *testing.T) {
	var notebook sagemaker.DescribeNotebookInstanceOutput
	notebookName := resource.PrefixedUniqueId(sagemakerTestAccSagemakerNotebookInstanceResourceNamePrefix)
	var resourceName = "aws_sagemaker_notebook_instance.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerNotebookInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerNotebookInstanceConfig(notebookName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerNotebookInstanceExists(resourceName, &notebook),
					testAccCheckAWSSagemakerNotebookInstanceDisappears(&notebook),
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

		if *notebookInstance.NotebookInstanceName == rs.Primary.ID {
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

func testAccCheckAWSSagemakerNotebookInstanceDisappears(instance *sagemaker.DescribeNotebookInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		if *instance.NotebookInstanceStatus != sagemaker.NotebookInstanceStatusFailed && *instance.NotebookInstanceStatus != sagemaker.NotebookInstanceStatusStopped {
			if err := stopSagemakerNotebookInstance(conn, *instance.NotebookInstanceName); err != nil {
				return err
			}
		}

		deleteOpts := &sagemaker.DeleteNotebookInstanceInput{
			NotebookInstanceName: instance.NotebookInstanceName,
		}

		if _, err := conn.DeleteNotebookInstance(deleteOpts); err != nil {
			return fmt.Errorf("error trying to delete sagemaker notebook instance (%s): %s", aws.StringValue(instance.NotebookInstanceName), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{
				sagemaker.NotebookInstanceStatusDeleting,
			},
			Target:  []string{""},
			Refresh: sagemakerNotebookInstanceStateRefreshFunc(conn, *instance.NotebookInstanceName),
			Timeout: 10 * time.Minute,
		}
		_, err := stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to delete: %s", aws.StringValue(instance.NotebookInstanceName), err)
		}

		return nil
	}
}

func testAccCheckAWSSagemakerNotebookInstanceName(notebook *sagemaker.DescribeNotebookInstanceOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		notebookName := notebook.NotebookInstanceName
		if *notebookName != expected {
			return fmt.Errorf("Bad Notebook Instance name: %s", *notebook.NotebookInstanceName)
		}

		return nil
	}
}

func testAccCheckAWSSagemakerNotebookInstanceTags(notebook *sagemaker.DescribeNotebookInstanceOutput, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		ts, err := conn.ListTags(&sagemaker.ListTagsInput{
			ResourceArn: notebook.NotebookInstanceArn,
		})
		if err != nil {
			return fmt.Errorf("Error listing tags: %s", err)
		}

		m := tagsToMapSagemaker(ts.Tags)
		v, ok := m[key]
		if value != "" && !ok {
			return fmt.Errorf("Missing tag: %s", key)
		} else if value == "" && ok {
			return fmt.Errorf("Extra tag: %s", key)
		}
		if value == "" {
			return nil
		}

		if v != value {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}

		return nil
	}
}

func testAccAWSSagemakerNotebookInstanceConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.t2.medium"
}

resource "aws_iam_role" "foo" {
	name = "%s"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName, notebookName)
}

func testAccAWSSagemakerNotebookInstanceUpdateConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.m4.xlarge"
}

resource "aws_iam_role" "foo" {
	name = "%s"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName, notebookName)
}

func testAccAWSSagemakerNotebookInstanceTagsConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.t2.medium"
	tags {
		foo = "bar"
	}
}

resource "aws_iam_role" "foo" {
	name = "%s"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName, notebookName)
}

func testAccAWSSagemakerNotebookInstanceTagsUpdateConfig(notebookName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance" "foo" {
	name = "%s"
	role_arn = "${aws_iam_role.foo.arn}"
	instance_type = "ml.t2.medium"
	tags {
		bar = "baz"
	}
}

resource "aws_iam_role" "foo" {
	name = "%s"
	path = "/"
	assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
	statement {
		actions = [ "sts:AssumeRole" ]
		principals {
			type = "Service"
			identifiers = [ "sagemaker.amazonaws.com" ]
		}
	}
}
`, notebookName, notebookName)
}
