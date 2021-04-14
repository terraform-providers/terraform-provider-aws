package aws

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ecs_service", &resource.Sweeper{
		Name: "aws_ecs_service",
		F:    testSweepEcsServices,
	})
}

func testSweepEcsServices(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ecsconn

	err = conn.ListClustersPages(&ecs.ListClustersInput{}, func(page *ecs.ListClustersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, clusterARNPtr := range page.ClusterArns {
			input := &ecs.ListServicesInput{
				Cluster: clusterARNPtr,
			}

			err = conn.ListServicesPages(input, func(page *ecs.ListServicesOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, serviceARNPtr := range page.ServiceArns {
					describeServicesInput := &ecs.DescribeServicesInput{
						Cluster:  clusterARNPtr,
						Services: []*string{serviceARNPtr},
					}
					serviceARN := aws.StringValue(serviceARNPtr)

					log.Printf("[DEBUG] Describing ECS Service: %s", serviceARN)
					describeServicesOutput, err := conn.DescribeServices(describeServicesInput)

					if isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") {
						continue
					}

					if err != nil {
						log.Printf("[ERROR] Error describing ECS Service (%s): %s", serviceARN, err)
						continue
					}

					if describeServicesOutput == nil || len(describeServicesOutput.Services) == 0 {
						continue
					}

					service := describeServicesOutput.Services[0]

					if aws.StringValue(service.Status) == "INACTIVE" {
						continue
					}

					deleteServiceInput := &ecs.DeleteServiceInput{
						Cluster: service.ClusterArn,
						Force:   aws.Bool(true),
						Service: service.ServiceArn,
					}

					log.Printf("[INFO] Deleting ECS Service: %s", serviceARN)
					_, err = conn.DeleteService(deleteServiceInput)

					if err != nil {
						log.Printf("[ERROR] Error deleting ECS Service (%s): %s", serviceARN, err)
					}
				}

				return !isLast
			})
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Service sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving ECS Services: %s", err)
	}

	return nil
}

func TestAccAWSEcsService_withARN(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},

			{
				Config: testAccAWSEcsServiceModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_basicImport(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"
	importInput := fmt.Sprintf("%s/%s", rName, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithFamilyAndRevision(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
			// Test existent resource import
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
				// wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"wait_for_steady_state"},
			},
			// Test non-existent resource import
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/nonexistent", rName),
				ImportState:       true,
				ImportStateVerify: false,
				ExpectError:       regexp.MustCompile(`(Please verify the ID is correct|Cannot import non-existent remote object)`),
			},
		},
	})
}

func TestAccAWSEcsService_disappears(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcsService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcsService_withUnnormalizedPlacementStrategy(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithInterchangeablePlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withCapacityProviderStrategy(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithCapacityProviderStrategy(rName, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
			{
				Config: testAccAWSEcsServiceWithCapacityProviderStrategy(rName, 10, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withMultipleCapacityProviderStrategies(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithMultipleCapacityProviderStrategies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withFamilyAndRevision(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithFamilyAndRevision(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},

			{
				Config: testAccAWSEcsServiceWithFamilyAndRevisionModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2427
func TestAccAWSEcsService_withRenamedCluster(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithRenamedCluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "cluster", "aws_ecs_cluster.default", "arn"),
				),
			},

			{
				Config: testAccAWSEcsServiceWithRenamedCluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttrPair(resourceName, "cluster", "aws_ecs_cluster.default", "arn"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_healthCheckGracePeriodSeconds(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEcsService_healthCheckGracePeriodSeconds(rName, -1),
				ExpectError: regexp.MustCompile(`expected health_check_grace_period_seconds to be in the range`),
			},
			{
				Config:      testAccAWSEcsService_healthCheckGracePeriodSeconds(rName, math.MaxInt32+1),
				ExpectError: regexp.MustCompile(`expected health_check_grace_period_seconds to be in the range`),
			},
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "300"),
				),
			},
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(rName, 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "600"),
				),
			},
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(rName, math.MaxInt32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "2147483647"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withIamRole(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withIamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withDeploymentController_Type_CodeDeploy(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigDeploymentControllerTypeCodeDeploy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", ecs.DeploymentControllerTypeCodeDeploy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
		},
	})
}

func TestAccAWSEcsService_withDeploymentController_Type_External(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigDeploymentControllerTypeExternal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_controller.0.type", ecs.DeploymentControllerTypeExternal),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"wait_for_steady_state"},
			},
		},
	})
}

func TestAccAWSEcsService_withDeploymentValues(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithDeploymentValues(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_maximum_percent", "200"),
					resource.TestCheckResourceAttr(resourceName, "deployment_minimum_healthy_percent", "100"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/6315
func TestAccAWSEcsService_withDeploymentMinimumZeroMaximumOneHundred(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigDeploymentPercents(rName, 0, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_maximum_percent", "100"),
					resource.TestCheckResourceAttr(resourceName, "deployment_minimum_healthy_percent", "0"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withDeploymentCircuitBreaker(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigDeploymentCircuitBreaker(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "deployment_circuit_breaker.0.rollback", "true"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3444
func TestAccAWSEcsService_withLbChanges(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withLbChanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
			{
				Config: testAccAWSEcsService_withLbChanges_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3361
func TestAccAWSEcsService_withEcsClusterName(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithEcsClusterName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "cluster", rName),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withAlb(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithAlb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withMultipleTargetGroups(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithMultipleTargetGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withForceNewDeployment(t *testing.T) {
	var service1, service2 ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "0"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithForceNewDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service2),
					testAccCheckAWSEcsServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withPlacementStrategy(t *testing.T) {
	var service1, service2, service3, service4 ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "0"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithPlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service2),
					testAccCheckAWSEcsServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithRandomPlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service3),
					testAccCheckAWSEcsServiceNotRecreated(&service2, &service3),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "random"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", ""),
				),
			},
			{
				Config: testAccAWSEcsServiceWithMultiPlacementStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service4),
					testAccCheckAWSEcsServiceNotRecreated(&service3, &service4),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.0.field", "memory"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.1.type", "spread"),
					resource.TestCheckResourceAttr(resourceName, "ordered_placement_strategy.1.field", "instanceId"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13146
func TestAccAWSEcsService_withPlacementStrategy_Type_Missing(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEcsServiceWithPlacementStrategyType(rName, ""),
				ExpectError: regexp.MustCompile(`expected ordered_placement_strategy.0.type to be one of`),
			},
		},
	})
}

func TestAccAWSEcsService_withPlacementConstraints(t *testing.T) {
	var service1, service2 ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithPlacementConstraint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithPlacementConstraintEmptyExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service2),
					testAccCheckAWSEcsServiceNotRecreated(&service1, &service2),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withPlacementConstraints_emptyExpression(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithPlacementConstraintEmptyExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeFargate(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "LATEST"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "true"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeFargateAndPlatformVersion(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargateAndPlatformVersion(rName, "1.3.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.3.0"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargateAndPlatformVersion(rName, "LATEST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "LATEST"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargateAndPlatformVersion(rName, "1.4.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.4.0"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeFargateAndWaitForSteadyState(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				// Wait for the ECS Cluster to reach a steady state w/specified count
				Config: testAccAWSEcsServiceWithLaunchTypeFargateAndWait(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeFargateAndUpdateWaitForSteadyState(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargateWithoutWait(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "false"),
				),
			},
			{
				// Modify desired count and wait for the ECS Cluster to reach steady state
				Config: testAccAWSEcsServiceWithLaunchTypeFargateAndWait(rName, 2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "true"),
				),
			},
			{
				// Modify desired count without wait
				Config: testAccAWSEcsServiceWithLaunchTypeFargateAndWait(rName, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "false"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeEC2AndNetworkConfiguration(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithNetworkConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithNetworkConfiguration_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withDaemonSchedulingStrategy(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithDaemonSchedulingStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "DAEMON"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withDaemonSchedulingStrategySetDeploymentMinimum(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithDaemonSchedulingStrategySetDeploymentMinimum(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "DAEMON"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withReplicaSchedulingStrategy(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithReplicaSchedulingStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "scheduling_strategy", "REPLICA"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withServiceRegistries(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withServiceRegistries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withServiceRegistries_container(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withServiceRegistries_container(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withServiceRegistriesChanges(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	serviceDiscoveryName := acctest.RandomWithPrefix("tf-acc-test")
	updatedServiceDiscoveryName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withServiceRegistriesChanges(rName, serviceDiscoveryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
			{
				Config: testAccAWSEcsService_withServiceRegistriesChanges(rName, updatedServiceDiscoveryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_Tags(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
				// Resource currently defaults to importing task_definition as family:revision
				// and wait_for_steady_state is not read from API
				ImportStateVerifyIgnore: []string{"task_definition", "wait_for_steady_state"},
			},
			{
				Config: testAccAWSEcsServiceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEcsServiceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_ManagedTags(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigManagedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "enable_ecs_managed_tags", "true"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_PropagateTags(t *testing.T) {
	var first, second, third ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigPropagateTags(rName, "SERVICE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", ecs.PropagateTagsService),
				),
			},
			{
				Config: testAccAWSEcsServiceConfigPropagateTags(rName, "TASK_DEFINITION"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", ecs.PropagateTagsTaskDefinition),
				),
			},
			{
				Config: testAccAWSEcsServiceConfigManagedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "propagate_tags", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_ExecuteCommand(t *testing.T) {
	var service ecs.Service
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceConfigExecuteCommand(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "enable_execute_command", "true"),
				),
			},
			{
				Config: testAccAWSEcsServiceConfigExecuteCommand(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "enable_execute_command", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSEcsServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_service" {
			continue
		}

		out, err := conn.DescribeServices(&ecs.DescribeServicesInput{
			Services: []*string{aws.String(rs.Primary.ID)},
			Cluster:  aws.String(rs.Primary.Attributes["cluster"]),
		})

		if err == nil {
			if len(out.Services) > 0 {
				var activeServices []*ecs.Service
				for _, svc := range out.Services {
					if *svc.Status != "INACTIVE" {
						activeServices = append(activeServices, svc)
					}
				}
				if len(activeServices) == 0 {
					return nil
				}

				return fmt.Errorf("ECS service still exists:\n%#v", activeServices)
			}
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSEcsServiceExists(name string, service *ecs.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		input := &ecs.DescribeServicesInput{
			Cluster:  aws.String(rs.Primary.Attributes["cluster"]),
			Services: []*string{aws.String(rs.Primary.ID)},
		}
		var output *ecs.DescribeServicesOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.DescribeServices(input)

			if err != nil {
				if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			if len(output.Services) == 0 {
				return resource.RetryableError(fmt.Errorf("service not found: %s", rs.Primary.ID))
			}

			return nil
		})

		if err != nil {
			return err
		}

		*service = *output.Services[0]

		return nil
	}
}

func testAccCheckAWSEcsServiceNotRecreated(i, j *ecs.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return fmt.Errorf("ECS Service (%s) unexpectedly recreated", aws.StringValue(j.ServiceArn))
		}

		return nil
	}
}

func testAccAWSEcsService(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}

func testAccAWSEcsServiceModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 2
}
`, rName)
}

func testAccAWSEcsServiceWithLaunchTypeFargateWithoutWait(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = element(aws_subnet.test.*.id, count.index)
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0",
    ]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }
}
`, rName)
}

func testAccAWSEcsServiceWithLaunchTypeFargateAndWait(rName string, desiredCount int, waitForSteadyState bool) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = element(aws_subnet.test.*.id, count.index)
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0",
    ]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = %d
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  wait_for_steady_state = %t
}

`, rName, desiredCount, waitForSteadyState)
}

func testAccAWSEcsServiceWithInterchangeablePlacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    field = "host"
    type  = "spread"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithCapacityProviderStrategy(rName string, weight, base int) string {
	return composeConfig(testAccAWSEcsCapacityProviderConfigBase(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}

resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.test.name
    weight            = %[2]d
    base              = %[3]d
  }
}
`, rName, weight, base))
}

func testAccAWSEcsServiceWithMultipleCapacityProviderStrategies(rName string) string {
	return composeConfig(testAccAWSEcsClusterCapacityProviders(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  network_configuration {
    security_groups  = [aws_security_group.allow_all.id]
    subnets          = [aws_subnet.test.id]
    assign_public_ip = false
  }

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_security_group" "allow_all" {
  name        = %[1]q
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSEcsServiceWithForceNewDeployment(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster              = aws_ecs_cluster.default.id
  desired_count        = 1
  force_new_deployment = true
  name                 = %[1]q
  task_definition      = aws_ecs_task_definition.test.arn

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithPlacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithPlacementStrategyType(rName string, placementStrategyType string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 1
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  ordered_placement_strategy {
    type = %[1]q
  }
}
`, rName, placementStrategyType)
}

func testAccAWSEcsServiceWithRandomPlacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    type = "random"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithMultiPlacementStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  ordered_placement_strategy {
    type  = "binpack"
    field = "memory"
  }

  ordered_placement_strategy {
    field = "host"
    type  = "spread"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithPlacementConstraint(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [${data.aws_availability_zones.available.names[0]}]"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithPlacementConstraintEmptyExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  placement_constraints {
    type = "distinctInstance"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithLaunchTypeFargate(rName string, assignPublicIP bool) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "allow_all_a" {
  name        = "%[1]s-1"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_security_group" "allow_all_b" {
  name        = "%[1]s-2"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.allow_all_a.id, aws_security_group.allow_all_b.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = %t
  }
}
`, rName, assignPublicIP)
}

func testAccAWSEcsServiceWithLaunchTypeFargateAndPlatformVersion(rName, platformVersion string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "allow_all_a" {
  name        = "%[1]s-1"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_security_group" "allow_all_b" {
  name        = "%[1]s-2"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name             = %[1]q
  cluster          = aws_ecs_cluster.test.id
  task_definition  = aws_ecs_task_definition.test.arn
  desired_count    = 1
  launch_type      = "FARGATE"
  platform_version = %[2]q

  network_configuration {
    security_groups  = [aws_security_group.allow_all_a.id, aws_security_group.allow_all_b.id]
    subnets          = aws_subnet.test[*].id
    assign_public_ip = false
  }
}
`, rName, platformVersion)
}

func testAccAWSEcsService_healthCheckGracePeriodSeconds(rName string, healthCheckGracePeriodSeconds int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:Describe*",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:RegisterTargets"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_ecs_service" "test" {
  name                              = %[1]q
  cluster                           = aws_ecs_cluster.test.id
  task_definition                   = aws_ecs_task_definition.test.arn
  desired_count                     = 1
  health_check_grace_period_seconds = %d
  iam_role                          = aws_iam_role.ecs_service.name

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName, healthCheckGracePeriodSeconds)
}

func testAccAWSEcsService_withIamRole(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

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
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {"AWS": "*"},
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:*",
        "ec2:*",
        "ecs:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_elb" "test" {
  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = 8080
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  iam_role        = aws_iam_role.ecs_service.name

  load_balancer {
    elb_name       = aws_elb.test.id
    container_name = "ghost"
    container_port = "2368"
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName)
}

func testAccAWSEcsServiceWithDeploymentValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}

func testAccAWSEcsService_withLbChangesBase(rName, image, containerName string, containerPort, hostPort int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

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
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": %[2]q,
    "memory": 128,
    "name": %[3]q,
    "portMappings": [
      {
        "containerPort": %[4]d,
        "hostPort": %[5]d
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {"AWS": "*"},
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:*",
        "ec2:*",
        "ecs:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_elb" "test" {
  internal = true
  subnets  = aws_subnet.test[*].id

  listener {
    instance_port     = %[5]d
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  iam_role        = aws_iam_role.ecs_service.name

  load_balancer {
    elb_name       = aws_elb.test.id
    container_name = %[3]q
    container_port = %[4]d
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName, image, containerName, containerPort, hostPort)
}

func testAccAWSEcsService_withLbChanges(rName string) string {
	return testAccAWSEcsService_withLbChangesBase(rName, "ghost:latest", "ghost", 2368, 8080)
}

func testAccAWSEcsService_withLbChanges_modified(rName string) string {
	return testAccAWSEcsService_withLbChangesBase(rName, "nginx:latest", "nginx", 80, 8080)
}

func testAccAWSEcsServiceWithFamilyAndRevision(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccAWSEcsServiceWithFamilyAndRevisionModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccAWSEcsServiceWithRenamedCluster(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.id
  task_definition = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  desired_count   = 1
}
`, rName)
}

func testAccAWSEcsServiceWithEcsClusterName(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "jenkins:latest",
    "memory": 128,
    "name": "jenkins"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.default.name
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}
`, rName)
}

func testAccAWSEcsServiceWithAlb(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = %[1]q
  role = aws_iam_role.ecs_service.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:Describe*",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:RegisterTargets"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  iam_role        = aws_iam_role.ecs_service.name

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }

  depends_on = [aws_iam_role_policy.ecs_service]
}
`, rName)
}

func testAccAWSEcsServiceWithMultipleTargetGroups(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 2368
      },
      {
        "containerPort": 4501,
        "hostPort": 4501
      }
    ]
  }
]
DEFINITION
}

resource "aws_lb_target_group" "test" {
  name     = "${aws_lb.test.name}1"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group" "static" {
  name     = "${aws_lb.test.name}2"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.static.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.static.id
    container_name   = "ghost"
    container_port   = "4501"
  }
}
`, rName)
}

func testAccAWSEcsServiceWithNetworkConfigurationBase(rName, securityGroups string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "allow_all_a" {
  name        = "%[1]s-1"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_security_group" "allow_all_b" {
  name        = "%[1]s-2"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                = %[1]q
  network_mode          = "awsvpc"
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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  network_configuration {
    security_groups = [%[2]s]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName, securityGroups)
}

func testAccAWSEcsServiceWithNetworkConfiguration(rName string) string {
	return testAccAWSEcsServiceWithNetworkConfigurationBase(rName, "aws_security_group.allow_all_a.id, aws_security_group.allow_all_b.id")
}

func testAccAWSEcsServiceWithNetworkConfiguration_modified(rName string) string {
	return testAccAWSEcsServiceWithNetworkConfigurationBase(rName, "aws_security_group.allow_all_a.id")
}

func testAccAWSEcsService_withServiceRegistries(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  state = "available"

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
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.test.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[1]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_registries {
    port         = 34567
    registry_arn = aws_service_discovery_service.test.arn
  }

  network_configuration {
    security_groups = [aws_security_group.test.id]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName)
}

func testAccAWSEcsService_withServiceRegistries_container(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  state = "available"

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
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.test.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[1]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb",
    "portMappings": [
    {
      "hostPort": 0,
      "protocol": "tcp",
      "containerPort": 27017
    }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_registries {
    container_name = "mongodb"
    container_port = 27017
    registry_arn   = aws_service_discovery_service.test.arn
  }
}
`, rName)
}

func testAccAWSEcsService_withServiceRegistriesChanges(rName, discoveryName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  state = "available"

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
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.test.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[2]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[2]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

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
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  service_registries {
    port         = 34567
    registry_arn = aws_service_discovery_service.test.arn
  }

  network_configuration {
    security_groups = [aws_security_group.test.id]
    subnets         = aws_subnet.test[*].id
  }
}
`, rName, discoveryName)
}

func testAccAWSEcsServiceWithDaemonSchedulingStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                = %[1]q
  cluster             = aws_ecs_cluster.default.id
  task_definition     = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy = "DAEMON"
}
`, rName)
}

func testAccAWSEcsServiceWithDaemonSchedulingStrategySetDeploymentMinimum(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                               = %[1]q
  cluster                            = aws_ecs_cluster.default.id
  task_definition                    = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy                = "DAEMON"
  deployment_minimum_healthy_percent = "50"
}
`, rName)
}

func testAccAWSEcsServiceConfigDeploymentControllerTypeCodeDeploy(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

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
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "test",
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 0
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  deployment_controller {
    type = "CODE_DEPLOY"
  }

  load_balancer {
    container_name   = "test"
    container_port   = "80"
    target_group_arn = aws_lb_target_group.test.id
  }
}
`, rName)
}

func testAccAWSEcsServiceConfigDeploymentControllerTypeExternal(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_service" "test" {
  cluster       = aws_ecs_cluster.test.id
  desired_count = 0
  name          = %[1]q

  deployment_controller {
    type = "EXTERNAL"
  }
}
`, rName)
}

func testAccAWSEcsServiceConfigDeploymentPercents(rName string, deploymentMinimumHealthyPercent, deploymentMaximumPercent int) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster                            = aws_ecs_cluster.test.id
  deployment_maximum_percent         = %[2]d
  deployment_minimum_healthy_percent = %[3]d
  desired_count                      = 1
  name                               = %[1]q
  task_definition                    = aws_ecs_task_definition.test.arn
}
`, rName, deploymentMaximumPercent, deploymentMinimumHealthyPercent)
}

func testAccAWSEcsServiceConfigDeploymentCircuitBreaker(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 1
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }
}
`, rName)
}

func testAccAWSEcsServiceConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 0
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSEcsServiceConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 0
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAWSEcsServiceConfigManagedTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
}

resource "aws_ecs_service" "test" {
  cluster                 = aws_ecs_cluster.test.id
  desired_count           = 0
  name                    = %[1]q
  task_definition         = aws_ecs_task_definition.test.arn
  enable_ecs_managed_tags = true

  tags = {
    tag-key = "tag-value"
  }
}
`, rName)
}

func testAccAWSEcsServiceConfigPropagateTags(rName, propagate string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

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
    tag-key = "task-def"
  }
}

resource "aws_ecs_service" "test" {
  cluster                 = aws_ecs_cluster.test.id
  desired_count           = 0
  name                    = %[1]q
  task_definition         = aws_ecs_task_definition.test.arn
  enable_ecs_managed_tags = true
  propagate_tags          = %[2]q

  tags = {
    tag-key = "service"
  }
}
`, rName, propagate)
}

func testAccAWSEcsServiceWithReplicaSchedulingStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "ghost:latest",
    "memory": 128,
    "name": "ghost"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name                = %[1]q
  cluster             = aws_ecs_cluster.default.id
  task_definition     = "${aws_ecs_task_definition.test.family}:${aws_ecs_task_definition.test.revision}"
  scheduling_strategy = "REPLICA"
  desired_count       = 1
}
`, rName)
}

func testAccAWSEcsServiceConfigExecuteCommand(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

  inline_policy {
    name = "exec_policy"

    policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
       {
       "Effect": "Allow",
       "Action": [
            "ssmmessages:CreateControlChannel",
            "ssmmessages:CreateDataChannel",
            "ssmmessages:OpenControlChannel",
            "ssmmessages:OpenDataChannel"
       ],
      "Resource": "*"
      }
   ]
}
EOF
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  task_role_arn = aws_iam_role.test.arn

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
}

resource "aws_ecs_service" "test" {
  cluster                = aws_ecs_cluster.test.id
  desired_count          = 0
  name                   = %[1]q
  task_definition        = aws_ecs_task_definition.test.arn
  enable_execute_command = %[2]t
}
`, rName, enable)
}
