package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	GlueDevEndpointResourcePrefix = "tf-acc-test"
	publicKey1                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 foo1@bar.com"
	publicKey2                    = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAq6U3HQYC4g8WzU147gZZ7CKQH8TgYn3chZGRPxaGmHW1RUwsyEs0nmombmIhwxudhJ4ehjqXsDLoQpd6+c7BuLgTMvbv8LgE9LX53vnljFe1dsObsr/fYLvpU9LTlo8HgHAqO5ibNdrAUvV31ronzCZhms/Gyfdaue88Fd0/YnsZVGeOZPayRkdOHSpqme2CBrpa8myBeL1CWl0LkDG4+YCURjbaelfyZlIApLYKy3FcCan9XQFKaL32MJZwCgzfOvWIMtYcU8QtXMgnA3/I3gXk8YDUJv5P4lj0s/PJXuTM8DygVAUtebNwPuinS7wwonm5FXcWMuVGsVpG5K7FGQ== foo2@bar.com"
	publicKey3                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtCk0lzMj1gPEOjdfQ37AIxCyETqJBubaMWuB4bgvGHp8LvEghr2YDl2bml1JrE1EOcZhPnIwgyucryXKA959sTUlgbvaFN7vmpVze56Q9tVU6BJQxOdaRoy5FcQMET9LB6SdbXk+V4CkDMsQNaFXezpg98HgCj+V7+bBWsfI6U63IESlWKK7kraCom8EWxkQk4mk9fizE2I+KrtiqN4xcah02LFG6IMnS+Xy3CDhcpZeYzWOV6zhcf675UJOdg/pLgQbUhhiwTOJFgRo8IcvE3iBrRMz508ppx6vLLr8J+3B8ujykc+/3ZSGfQfx6rO+OuSskhG5FLI6icbQBtBzf foo3@bar.com"
	publicKey4                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 foo4@bar.com"
)

func init() {
	resource.AddTestSweepers("aws_glue_dev_endpoint", &resource.Sweeper{
		Name: "aws_glue_dev_endpoint",
		F:    testSweepGlueDevEndpoint,
	})
}

func testSweepGlueDevEndpoint(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	input := &glue.GetDevEndpointsInput{}
	err = conn.GetDevEndpointsPages(input, func(page *glue.GetDevEndpointsOutput, lastPage bool) bool {
		if len(page.DevEndpoints) == 0 {
			log.Printf("[INFO] No Glue Dev Endpoints to sweep")
			return false
		}
		for _, endpoint := range page.DevEndpoints {
			name := aws.StringValue(endpoint.EndpointName)
			if !strings.HasPrefix(name, GlueDevEndpointResourcePrefix) {
				log.Printf("[INFO] Skipping Glue Dev Endpoint: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Glue Dev Endpoint: %s", name)
			_, err := conn.DeleteDevEndpoint(&glue.DeleteDevEndpointInput{
				EndpointName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Dev Endpoint %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Dev Endpoint sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Glue Dev Endpoint: %s", err)
	}

	return nil
}

func TestAccGlueDevEndpoint_Basic(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("devEndpoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "5"),
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

func TestAccGlueDevEndpoint_Arguments(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_Arguments(rName, "--arg1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlueDevEndpointConfig_Arguments2(rName, "--arg1", "value1updated", "--arg2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg2", "value2"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_Arguments(rName, "--arg2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueDevEndpoint_ExtraJarsS3Path(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	extraJarsS3Path := "foo"
	extraJarsS3PathUpdated := "bar"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName, extraJarsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "extra_jars_s3_path", extraJarsS3Path),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName, extraJarsS3PathUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "extra_jars_s3_path", extraJarsS3PathUpdated),
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

func TestAccGlueDevEndpoint_ExtraPythonLibsS3Path(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	extraPythonLibsS3Path := "foo"
	extraPythonLibsS3PathUpdated := "bar"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName, extraPythonLibsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "extra_python_libs_s3_path", extraPythonLibsS3Path),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName, extraPythonLibsS3PathUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "extra_python_libs_s3_path", extraPythonLibsS3PathUpdated),
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

func TestAccGlueDevEndpoint_GlueVersion(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccGlueDevEndpointConfig_GlueVersion(rName, "1"),
				ExpectError: regexp.MustCompile(`must match version pattern X.X`),
			},
			{
				Config: testAccGlueDevEndpointConfig_GlueVersion(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_GlueVersion(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
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

func TestAccGlueDevEndpoint_NumberOfNodes(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccGlueDevEndpointConfig_NumberOfNodes(rName, 1),
				ExpectError: regexp.MustCompile(`expected number_of_nodes to be at least`),
			},
			{
				Config: testAccGlueDevEndpointConfig_NumberOfNodes(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "2"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_NumberOfNodes(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "5"),
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

func TestAccGlueDevEndpoint_NumberOfWorkers(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccGlueDevEndpointConfig_NumberOfWorkers(rName, 1),
				ExpectError: regexp.MustCompile(`expected number_of_workers to be at least`),
			},
			{
				Config: testAccGlueDevEndpointConfig_NumberOfWorkers(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", "2"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_NumberOfWorkers(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", "5"),
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

func TestAccGlueDevEndpoint_PublicKey(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_PublicKey(rName, publicKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey1),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_PublicKey(rName, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey2),
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

func TestAccGlueDevEndpoint_PublicKeys(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_PublicKeys2(rName, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "2"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_PublicKeys3(rName, publicKey1, publicKey3, publicKey4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "3"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_PublicKeys4(rName, publicKey1, publicKey1, publicKey3, publicKey4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "3"),
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

func TestAccGlueDevEndpoint_SecurityConfiguration(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_SecurityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", rName),
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

// Note: Either none or both of subnetId and securityGroupIds must be specified.
func TestAccGlueDevEndpoint_SubnetID_SecurityGroupIDs(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_SubnetID_SecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "aws_subnet.test", "availability_zone"),
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

func TestAccGlueDevEndpoint_Tags(t *testing.T) {
	var endpoint1, endpoint2, endpoint3 glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueDevEndpointConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint1),
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
				Config: testAccAWSGlueDevEndpointConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGlueDevEndpointConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueDevEndpoint_WorkerType(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_WorkerType_Standard(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "worker_type", glue.WorkerTypeStandard),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_WorkerType(rName, glue.WorkerTypeG1x),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "worker_type", glue.WorkerTypeG1x),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_WorkerType(rName, glue.WorkerTypeG2x),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "worker_type", glue.WorkerTypeG2x),
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

func TestAccGlueDevEndpoint_disappears(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlueDevEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGlueDevEndpointExists(resourceName string, endpoint *glue.DevEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := conn.GetDevEndpoint(&glue.GetDevEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("no Glue Dev Endpoint")
		}

		*endpoint = *output.DevEndpoint

		return nil
	}
}

func testAccCheckAWSGlueDevEndpointDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_dev_endpoint" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := conn.GetDevEndpoint(&glue.GetDevEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				return nil
			}
			return err
		}

		endpoint := output.DevEndpoint
		if endpoint != nil && aws.StringValue(endpoint.EndpointName) == rs.Primary.ID {
			return fmt.Errorf("the Glue Dev Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccGlueDevEndpointConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = "AWSGlueServiceRole-%[1]s"
  assume_role_policy = data.aws_iam_policy_document.service.json
}

data "aws_iam_policy_document" "service" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "glue_service_role" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}

data "aws_partition" "current" {}
`, rName)
}

func testAccGlueDevEndpointConfig_Basic(rName string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccGlueDevEndpointConfig_Arguments(rName, argKey, argValue string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  arguments = {
    %[2]q = %[3]q
  }
}
`, rName, argKey, argValue)
}

func testAccGlueDevEndpointConfig_Arguments2(rName, argKey1, argValue1, argKey2, argValue2 string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  arguments = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, argKey1, argValue1, argKey2, argValue2)
}

func testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName string, extraJarsS3Path string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name               = %q
  role_arn           = aws_iam_role.test.arn
  extra_jars_s3_path = %q
}
`, rName, extraJarsS3Path)
}

func testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName string, extraPythonLibsS3Path string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name                      = %q
  role_arn                  = aws_iam_role.test.arn
  extra_python_libs_s3_path = %q
}
`, rName, extraPythonLibsS3Path)
}

func testAccGlueDevEndpointConfig_GlueVersion(rName string, glueVersion string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  glue_version = %[2]q
}
`, rName, glueVersion)
}

func testAccGlueDevEndpointConfig_NumberOfNodes(rName string, numberOfNodes int) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name            = %q
  role_arn        = aws_iam_role.test.arn
  number_of_nodes = %d
}
`, rName, numberOfNodes)
}

func testAccGlueDevEndpointConfig_NumberOfWorkers(rName string, numberOfWorkers int) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name              = %q
  role_arn          = aws_iam_role.test.arn
  worker_type       = "G.1X"
  number_of_workers = %d
}
`, rName, numberOfWorkers)
}

func testAccGlueDevEndpointConfig_PublicKey(rName string, publicKey string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name       = %q
  role_arn   = aws_iam_role.test.arn
  public_key = "%s"
}
`, rName, publicKey)
}

func testAccGlueDevEndpointConfig_PublicKeys2(rName string, publicKey1 string, publicKey2 string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  public_keys = [%[2]q, %[3]q]
}
`, rName, publicKey1, publicKey2)
}

func testAccGlueDevEndpointConfig_PublicKeys3(rName string, publicKey1 string, publicKey2 string, publicKey3 string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  public_keys = [%[2]q, %[3]q, %[4]q]
}
`, rName, publicKey1, publicKey2, publicKey3)
}

func testAccGlueDevEndpointConfig_PublicKeys4(rName string, publicKey1 string, publicKey2 string, publicKey3 string, publicKey4 string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  public_keys = [%[2]q, %[3]q, %[4]q, %[5]q]
}
`, rName, publicKey1, publicKey2, publicKey3, publicKey4)
}

func testAccGlueDevEndpointConfig_SecurityConfiguration(rName string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name                   = %[1]q
  role_arn               = aws_iam_role.test.arn
  security_configuration = aws_glue_security_configuration.test.name
}

resource "aws_glue_security_configuration" "test" {
  name = %[1]q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccGlueDevEndpointConfig_SubnetID_SecurityGroupIDs(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), testAccGlueDevEndpointConfig_Base(rName), fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name               = %[1]q
  role_arn           = aws_iam_role.test.arn
  subnet_id          = aws_subnet.test.id
  security_group_ids = [aws_security_group.test.id]
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = data.aws_vpc_endpoint_service.s3.service_name
}

data "aws_vpc_endpoint_service" "s3" {
  service      = "s3"
  service_type = "Gateway"
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
  route_table_id  = aws_vpc.test.main_route_table_id
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }

  timeouts {
    delete = "40m"
  }
  depends_on = [aws_iam_role_policy_attachment.glue_service_role]
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  timeouts {
    delete = "40m"
  }
  depends_on = [aws_iam_role_policy_attachment.glue_service_role]
}
`, rName))
}

func testAccAWSGlueDevEndpointConfig_Tags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSGlueJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGlueDevEndpointConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSGlueJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccGlueDevEndpointConfig_WorkerType(rName, workerType string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = %[2]q
  number_of_workers = 2
}
`, rName, workerType)
}

func testAccGlueDevEndpointConfig_WorkerType_Standard(rName string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"
  number_of_workers = 2
}
`, rName)
}
