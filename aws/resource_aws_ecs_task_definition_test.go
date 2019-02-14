package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEcsTaskDefinition_basic(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_basic_%s", rString)
	resourceName := "aws_ecs_task_definition.jenkins"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinition(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists(resourceName, &def),
				),
			},
			{
				Config: testAccAWSEcsTaskDefinitionModified(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists(resourceName, &def),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSEcsTaskDefinitionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2370
func TestAccAWSEcsTaskDefinition_withScratchVolume(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_with_scratch_volume_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithScratchVolume(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withDockerVolume(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_with_docker_volume_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithDockerVolumes(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.scope", "shared"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.autoprovision", "true"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.driver", "local"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.driver_opts.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.driver_opts.uid", "1000"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.driver_opts.device", "tmpfs"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.labels.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.labels.stack", "april"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.labels.environment", "test"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withDockerVolumeMinimalConfig(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_with_docker_volume_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithDockerVolumesMinimalConfig(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.scope", "task"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.driver", "local"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withTaskScopedDockerVolume(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_with_docker_volume_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithTaskScopedDockerVolume(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					testAccCheckAWSTaskDefinitionDockerVolumeConfigurationAutoprovisionNil(&def),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.scope", "task"),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "volume.584193650.docker_volume_configuration.0.driver", "local"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2694
func TestAccAWSEcsTaskDefinition_withEcsService(t *testing.T) {
	var def ecs.TaskDefinition
	var service ecs.Service

	rString := acctest.RandString(8)
	clusterName := fmt.Sprintf("tf_acc_cluster_with_ecs_service_%s", rString)
	svcName := fmt.Sprintf("tf_acc_td_with_ecs_service_%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_with_ecs_service_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithEcsService(clusterName, svcName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					testAccCheckAWSEcsServiceExists("aws_ecs_service.sleep-svc", &service),
				),
			},
			{
				Config: testAccAWSEcsTaskDefinitionWithEcsServiceModified(clusterName, svcName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					testAccCheckAWSEcsServiceExists("aws_ecs_service.sleep-svc", &service),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withTaskRoleArn(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_ecs_td_with_task_role_arn_%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-ecs-td-with-task-role-arn-%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_with_task_role_arn_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithTaskRoleArn(roleName, policyName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withNetworkMode(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_ecs_td_with_network_mode_%s", rString)
	policyName := fmt.Sprintf("tf_acc_ecs_td_with_network_mode_%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_with_network_mode_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithNetworkMode(roleName, policyName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "network_mode", "bridge"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withIPCMode(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_ecs_td_with_ipc_mode_%s", rString)
	policyName := fmt.Sprintf("tf_acc_ecs_td_with_ipc_mode_%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_with_ipc_mode_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithIpcMode(roleName, policyName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "ipc_mode", "host"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_withPidMode(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_ecs_td_with_pid_mode_%s", rString)
	policyName := fmt.Sprintf("tf_acc_ecs_td_with_pid_mode_%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_with_pid_mode_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionWithPidMode(roleName, policyName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.sleep", &def),
					resource.TestCheckResourceAttr(
						"aws_ecs_task_definition.sleep", "pid_mode", "host"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_constraint(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_constraint_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinition_constraint(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.jenkins", &def),
					resource.TestCheckResourceAttr("aws_ecs_task_definition.jenkins", "placement_constraints.#", "1"),
					testAccCheckAWSTaskDefinitionConstraintsAttrs(&def),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_changeVolumesForcesNewResource(t *testing.T) {
	var before ecs.TaskDefinition
	var after ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_change_vol_forces_new_resource_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinition(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.jenkins", &before),
				),
			},
			{
				Config: testAccAWSEcsTaskDefinitionUpdatedVolume(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.jenkins", &after),
					testAccCheckEcsTaskDefinitionRecreated(t, &before, &after),
				),
			},
		},
	})
}

// Regression for https://github.com/terraform-providers/terraform-provider-aws/issues/2336
func TestAccAWSEcsTaskDefinition_arrays(t *testing.T) {
	var conf ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_arrays_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionArrays(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.test", &conf),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_Fargate(t *testing.T) {
	var conf ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_fargate_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionFargate(tdName, `[{"protocol": "tcp", "containerPort": 8000}]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.fargate", &conf),
					resource.TestCheckResourceAttr("aws_ecs_task_definition.fargate", "requires_compatibilities.#", "1"),
					resource.TestCheckResourceAttr("aws_ecs_task_definition.fargate", "cpu", "256"),
					resource.TestCheckResourceAttr("aws_ecs_task_definition.fargate", "memory", "512"),
				),
			},
			{
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
				Config:             testAccAWSEcsTaskDefinitionFargate(tdName, `[{"protocol": "tcp", "containerPort": 8000, "hostPort": 8000}]`),
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_ExecutionRole(t *testing.T) {
	var conf ecs.TaskDefinition

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_ecs_td_execution_role_%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-ecs-td-execution-role-%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_execution_role_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionExecutionRole(roleName, policyName, tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.fargate", &conf),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3582#issuecomment-286409786
func TestAccAWSEcsTaskDefinition_Inactive(t *testing.T) {
	var def ecs.TaskDefinition

	rString := acctest.RandString(8)
	tdName := fmt.Sprintf("tf_acc_td_basic_%s", rString)

	markTaskDefinitionInactive := func() {
		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		_, err := conn.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
			TaskDefinition: aws.String(fmt.Sprintf("%s:1", tdName)),
		})

		if err != nil {
			t.Fatalf("error deregistering ECS Task Definition (%s): %s", tdName, err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinition(tdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists("aws_ecs_task_definition.jenkins", &def),
				),
			},
			{
				Config:    testAccAWSEcsTaskDefinition(tdName),
				PreConfig: markTaskDefinitionInactive,
				Check:     resource.TestCheckResourceAttr("aws_ecs_task_definition.jenkins", "revision", "2"), // should get re-created
			},
		},
	})
}

func TestAccAWSEcsTaskDefinition_Tags(t *testing.T) {
	var taskDefinition ecs.TaskDefinition
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskDefinitionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists(resourceName, &taskDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSEcsTaskDefinitionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists(resourceName, &taskDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEcsTaskDefinitionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskDefinitionExists(resourceName, &taskDefinition),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckEcsTaskDefinitionRecreated(t *testing.T,
	before, after *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Revision == *after.Revision {
			t.Fatalf("Expected change of TaskDefinition Revisions, but both were %v", before.Revision)
		}
		return nil
	}
}

func testAccCheckAWSTaskDefinitionConstraintsAttrs(def *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(def.PlacementConstraints) != 1 {
			return fmt.Errorf("Expected (1) placement_constraints, got (%d)", len(def.PlacementConstraints))
		}
		return nil
	}
}
func TestValidateAwsEcsTaskDefinitionContainerDefinitions(t *testing.T) {
	validDefinitions := []string{
		testValidateAwsEcsTaskDefinitionValidContainerDefinitions,
	}
	for _, v := range validDefinitions {
		_, errors := validateAwsEcsTaskDefinitionContainerDefinitions(v, "container_definitions")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid AWS ECS Task Definition Container Definitions: %q", v, errors)
		}
	}

	invalidDefinitions := []string{
		testValidateAwsEcsTaskDefinitionInvalidCommandContainerDefinitions,
	}
	for _, v := range invalidDefinitions {
		_, errors := validateAwsEcsTaskDefinitionContainerDefinitions(v, "container_definitions")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid AWS ECS Task Definition Container Definitions", v)
		}
	}
}

func testAccCheckAWSEcsTaskDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_task_definition" {
			continue
		}

		input := ecs.DescribeTaskDefinitionInput{
			TaskDefinition: aws.String(rs.Primary.Attributes["arn"]),
		}

		out, err := conn.DescribeTaskDefinition(&input)

		if err != nil {
			return err
		}

		if out.TaskDefinition != nil && *out.TaskDefinition.Status != "INACTIVE" {
			return fmt.Errorf("ECS task definition still exists:\n%#v", *out.TaskDefinition)
		}
	}

	return nil
}

func testAccCheckAWSEcsTaskDefinitionExists(name string, def *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		out, err := conn.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
			TaskDefinition: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			return err
		}

		*def = *out.TaskDefinition

		return nil
	}
}

func testAccCheckAWSTaskDefinitionDockerVolumeConfigurationAutoprovisionNil(def *ecs.TaskDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(def.Volumes) != 1 {
			return fmt.Errorf("Expected (1) volumes, got (%d)", len(def.Volumes))
		}
		config := def.Volumes[0].DockerVolumeConfiguration
		if config == nil {
			return fmt.Errorf("Expected docker_volume_configuration, got nil")
		}
		if config.Autoprovision != nil {
			return fmt.Errorf("Expected autoprovision to be nil, got %t", *config.Autoprovision)
		}
		return nil
	}
}

func testAccAWSEcsTaskDefinition_constraint(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }

	placement_constraints {
		type = "memberOf"
		expression = "attribute:ecs.availability-zone in [us-west-2a, us-west-2b]"
	}
}
`, tdName)
}

func testAccAWSEcsTaskDefinition(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionUpdatedVolume(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name = "jenkins-home"
    host_path = "/ecs/jenkins"
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionArrays(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "links": ["container1", "container2", "container3"],
      "portMappings": [
        {"containerPort": 80},
        {"containerPort": 81},
        {"containerPort": 82}
      ],
      "environment": [
        {"name": "VARNAME1", "value": "VARVAL1"},
        {"name": "VARNAME2", "value": "VARVAL2"},
        {"name": "VARNAME3", "value": "VARVAL3"}
      ],
      "extraHosts": [
        {"hostname": "host1", "ipAddress": "127.0.0.1"},
        {"hostname": "host2", "ipAddress": "127.0.0.2"},
        {"hostname": "host3", "ipAddress": "127.0.0.3"}
      ],
      "mountPoints": [
        {"sourceVolume": "vol1", "containerPath": "/vol1"},
        {"sourceVolume": "vol2", "containerPath": "/vol2"},
        {"sourceVolume": "vol3", "containerPath": "/vol3"}
      ],
      "volumesFrom": [
        {"sourceContainer": "container1"},
        {"sourceContainer": "container2"},
        {"sourceContainer": "container3"}
      ],
      "ulimits": [
        {
          "name": "core",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "cpu",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "fsize",
          "softLimit": 10, "hardLimit": 20
        }
      ],
      "linuxParameters": {
        "capabilities": {
          "add": ["AUDIT_CONTROL", "AUDIT_WRITE", "BLOCK_SUSPEND"],
          "drop": ["CHOWN", "IPC_LOCK", "KILL"]
        }
      },
      "devices": [
        {
          "hostPath": "/path1",
          "permissions": ["read", "write", "mknod"]
        },
        {
          "hostPath": "/path2",
          "permissions": ["read", "write"]
        },
        {
          "hostPath": "/path3",
          "permissions": ["read", "mknod"]
        }
      ],
      "dockerSecurityOptions": ["label:one", "label:two", "label:three"],
      "memory": 500,
      "cpu": 10
    },
    {
      "name": "container1",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container2",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container3",
      "image": "busybox",
      "memory": 100
    }
]
TASK_DEFINITION
  volume {
    name = "vol1"
    host_path = "/host/vol1"
  }
  volume {
    name = "vol2"
    host_path = "/host/vol2"
  }
  volume {
    name = "vol3"
    host_path = "/host/vol3"
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionFargate(tdName, portMappings string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "fargate" {
  family                   = "%s"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true,
    "portMappings": %s
  }
]
TASK_DEFINITION
}
`, tdName, portMappings)
}

func testAccAWSEcsTaskDefinitionExecutionRole(roleName, policyName, tdName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy" {
  name        = "%s"
  description = "A test policy"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${aws_iam_policy.policy.arn}"
}

resource "aws_ecs_task_definition" "fargate" {
  family                   = "%s"
  execution_role_arn       = "${aws_iam_role.role.arn}"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION
}
`, roleName, policyName, tdName)
}

func testAccAWSEcsTaskDefinitionWithScratchVolume(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "sleep" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = "database_scratch"
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionWithDockerVolumes(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "sleep" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = "database_scratch"
    docker_volume_configuration {
        driver = "local"
        scope  = "shared"
        driver_opts = {
            device = "tmpfs"
            uid    = "1000"
        }
        labels = {
            environment = "test"
            stack       = "april"
        }
        autoprovision = true
    }
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionWithDockerVolumesMinimalConfig(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "sleep" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = "database_scratch"
    docker_volume_configuration {
      autoprovision = true
    }
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionWithTaskScopedDockerVolume(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "sleep" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = "database_scratch"
    docker_volume_configuration {
      scope = "task"
    }
  }
}
`, tdName)
}

func testAccAWSEcsTaskDefinitionWithTaskRoleArn(roleName, policyName, tdName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role_test" {
	name = "%s"
	path = "/test/"
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "ec2.amazonaws.com"
			},
			"Effect": "Allow",
			"Sid": ""
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "role_test" {
	name = "%s"
	role = "${aws_iam_role.role_test.id}"
	policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:GetBucketLocation",
				"s3:ListAllMyBuckets"
			],
			"Resource": "arn:aws:s3:::*"
		}
	]
}
EOF
}

resource "aws_ecs_task_definition" "sleep" {
	family = "%s"
	task_role_arn = "${aws_iam_role.role_test.arn}"
	container_definitions = <<TASK_DEFINITION
[
	{
		"name": "sleep",
		"image": "busybox",
		"cpu": 10,
		"command": ["sleep","360"],
		"memory": 10,
		"essential": true
	}
]
TASK_DEFINITION
		volume {
		name = "database_scratch"
	}
}`, roleName, policyName, tdName)
}

func testAccAWSEcsTaskDefinitionWithIpcMode(roleName, policyName, tdName string) string {
	return fmt.Sprintf(`
 resource "aws_iam_role" "role_test" {
	 name = "%s"
	 path = "/test/"
	 assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Action": "sts:AssumeRole",
		 "Principal": {
			 "Service": "ec2.amazonaws.com"
		 },
		 "Effect": "Allow",
		 "Sid": ""
	 }
 ]
}
EOF
 }

 resource "aws_iam_role_policy" "role_test" {
	 name = "%s"
	 role = "${aws_iam_role.role_test.id}"
	 policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Effect": "Allow",
		 "Action": [
			 "s3:GetBucketLocation",
			 "s3:ListAllMyBuckets"
		 ],
		 "Resource": "arn:aws:s3:::*"
	 }
 ]
}
 EOF
 }

 resource "aws_ecs_task_definition" "sleep" {
	 family = "%s"
	 task_role_arn = "${aws_iam_role.role_test.arn}"
	 network_mode = "bridge"
	 ipc_mode = "host"
	 container_definitions = <<TASK_DEFINITION
[
 {
	 "name": "sleep",
	 "image": "busybox",
	 "cpu": 10,
	 "command": ["sleep","360"],
	 "memory": 10,
	 "essential": true
 }
]
TASK_DEFINITION

	 volume {
		 name = "database_scratch"
	 }
 }`, roleName, policyName, tdName)
}

func testAccAWSEcsTaskDefinitionWithPidMode(roleName, policyName, tdName string) string {
	return fmt.Sprintf(`
 resource "aws_iam_role" "role_test" {
	 name = "%s"
	 path = "/test/"
	 assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Action": "sts:AssumeRole",
		 "Principal": {
			 "Service": "ec2.amazonaws.com"
		 },
		 "Effect": "Allow",
		 "Sid": ""
	 }
 ]
}
EOF
 }

 resource "aws_iam_role_policy" "role_test" {
	 name = "%s"
	 role = "${aws_iam_role.role_test.id}"
	 policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Effect": "Allow",
		 "Action": [
			 "s3:GetBucketLocation",
			 "s3:ListAllMyBuckets"
		 ],
		 "Resource": "arn:aws:s3:::*"
	 }
 ]
}
 EOF
 }

 resource "aws_ecs_task_definition" "sleep" {
	 family = "%s"
	 task_role_arn = "${aws_iam_role.role_test.arn}"
	 network_mode = "bridge"
	 pid_mode = "host"
	 container_definitions = <<TASK_DEFINITION
[
 {
	 "name": "sleep",
	 "image": "busybox",
	 "cpu": 10,
	 "command": ["sleep","360"],
	 "memory": 10,
	 "essential": true
 }
]
TASK_DEFINITION

	 volume {
		 name = "database_scratch"
	 }
 }`, roleName, policyName, tdName)
}

func testAccAWSEcsTaskDefinitionWithNetworkMode(roleName, policyName, tdName string) string {
	return fmt.Sprintf(`
 resource "aws_iam_role" "role_test" {
	 name = "%s"
	 path = "/test/"
	 assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Action": "sts:AssumeRole",
		 "Principal": {
			 "Service": "ec2.amazonaws.com"
		 },
		 "Effect": "Allow",
		 "Sid": ""
	 }
 ]
}
EOF
 }

 resource "aws_iam_role_policy" "role_test" {
	 name = "%s"
	 role = "${aws_iam_role.role_test.id}"
	 policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
	 {
		 "Effect": "Allow",
		 "Action": [
			 "s3:GetBucketLocation",
			 "s3:ListAllMyBuckets"
		 ],
		 "Resource": "arn:aws:s3:::*"
	 }
 ]
}
 EOF
 }

 resource "aws_ecs_task_definition" "sleep" {
	 family = "%s"
	 task_role_arn = "${aws_iam_role.role_test.arn}"
	 network_mode = "bridge"
	 container_definitions = <<TASK_DEFINITION
[
 {
	 "name": "sleep",
	 "image": "busybox",
	 "cpu": 10,
	 "command": ["sleep","360"],
	 "memory": 10,
	 "essential": true
 }
]
TASK_DEFINITION

	 volume {
		 name = "database_scratch"
	 }
 }`, roleName, policyName, tdName)
}

func testAccAWSEcsTaskDefinitionWithEcsService(clusterName, svcName, tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_service" "sleep-svc" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.sleep.arn}"
  desired_count = 1
}

resource "aws_ecs_task_definition" "sleep" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = "database_scratch"
  }
}
`, clusterName, svcName, tdName)
}

func testAccAWSEcsTaskDefinitionWithEcsServiceModified(clusterName, svcName, tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_service" "sleep-svc" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.sleep.arn}"
  desired_count = 1
}

resource "aws_ecs_task_definition" "sleep" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 20,
    "command": ["sleep","360"],
    "memory": 50,
    "essential": true
  }
]
TASK_DEFINITION

  volume {
    name = "database_scratch"
  }
}
`, clusterName, svcName, tdName)
}

func testAccAWSEcsTaskDefinitionModified(tdName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 20,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }
}
`, tdName)
}

var testValidateAwsEcsTaskDefinitionValidContainerDefinitions = `
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": ["sleep","360"],
    "memory": 10,
    "essential": true
  }
]
`

var testValidateAwsEcsTaskDefinitionInvalidCommandContainerDefinitions = `
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 10,
    "command": "sleep 360",
    "memory": 10,
    "essential": true
  }
]
`

func testAccAWSEcsTaskDefinitionConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}

resource "aws_ecs_task_definition" "test" {
  family = %q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION

  tags = {
    %q = %q
  }
}
`, rName, rName, tag1Key, tag1Value)
}

func testAccAWSEcsTaskDefinitionConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}

resource "aws_ecs_task_definition" "test" {
  family = %q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAWSEcsTaskDefinitionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
