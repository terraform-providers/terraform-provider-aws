package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestParseTaskDefinition(t *testing.T) {
	cases := map[string]map[string]interface{}{
		"invalid": {
			"family":   "",
			"revision": "",
			"isValid":  false,
		},
		"invalidWithColon:": {
			"family":   "",
			"revision": "",
			"isValid":  false,
		},
		"1234": {
			"family":   "",
			"revision": "",
			"isValid":  false,
		},
		"invalid:aaa": {
			"family":   "",
			"revision": "",
			"isValid":  false,
		},
		"invalid=family:1": {
			"family":   "",
			"revision": "",
			"isValid":  false,
		},
		"invalid:name:1": {
			"family":   "",
			"revision": "",
			"isValid":  false,
		},
		"valid:1": {
			"family":   "valid",
			"revision": "1",
			"isValid":  true,
		},
		"abc12-def:54": {
			"family":   "abc12-def",
			"revision": "54",
			"isValid":  true,
		},
		"lorem_ip-sum:123": {
			"family":   "lorem_ip-sum",
			"revision": "123",
			"isValid":  true,
		},
		"lorem-ipsum:1": {
			"family":   "lorem-ipsum",
			"revision": "1",
			"isValid":  true,
		},
	}

	for input, expectedOutput := range cases {
		family, revision, err := parseTaskDefinition(input)
		isValid := expectedOutput["isValid"].(bool)
		if !isValid && err == nil {
			t.Fatalf("Task definition %s should fail", input)
		}

		expectedFamily := expectedOutput["family"].(string)
		if family != expectedFamily {
			t.Fatalf("Unexpected family (%#v) for task definition %s\n%#v", family, input, err)
		}
		expectedRevision := expectedOutput["revision"].(string)
		if revision != expectedRevision {
			t.Fatalf("Unexpected revision (%#v) for task definition %s\n%#v", revision, input, err)
		}
	}
}

func TestAccAWSEcsService_withARN(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-arn-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-arn-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-arn-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "service_registries.#", "0"),
				),
			},

			{
				Config: testAccAWSEcsServiceModified(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "service_registries.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_basicImport(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-%s", rString)

	resourceName := "aws_ecs_service.jenkins"
	importInput := fmt.Sprintf("%s/%s", clusterName, svcName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithFamilyAndRevision(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.jenkins", &service),
				),
			},
			// Test existent resource import
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test non-existent resource import
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/nonexistent", clusterName),
				ImportState:       true,
				ImportStateVerify: false,
				ExpectError:       regexp.MustCompile(`No ECS service found`),
			},
		},
	})
}

func TestAccAWSEcsService_withUnnormalizedPlacementStrategy(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-ups-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-ups-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-ups-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithInterchangeablePlacementStrategy(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withFamilyAndRevision(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-far-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-far-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-far-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithFamilyAndRevision(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.jenkins", &service),
				),
			},

			{
				Config: testAccAWSEcsServiceWithFamilyAndRevisionModified(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.jenkins", &service),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/2427
func TestAccAWSEcsService_withRenamedCluster(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-rc-%s", rString)
	uClusterName := fmt.Sprintf("tf-acc-cluster-svc-w-rc-updated-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-rc-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-rc-%s", rString)

	originalRegexp := regexp.MustCompile(
		"^arn:aws:ecs:[^:]+:[0-9]+:cluster/" + clusterName + "$")
	modifiedRegexp := regexp.MustCompile(
		"^arn:aws:ecs:[^:]+:[0-9]+:cluster/" + uClusterName + "$")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithRenamedCluster(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.ghost", &service),
					resource.TestMatchResourceAttr(
						"aws_ecs_service.ghost", "cluster", originalRegexp),
				),
			},

			{
				Config: testAccAWSEcsServiceWithRenamedCluster(uClusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.ghost", &service),
					resource.TestMatchResourceAttr(
						"aws_ecs_service.ghost", "cluster", modifiedRegexp),
				),
			},
		},
	})
}

func TestAccAWSEcsService_healthCheckGracePeriodSeconds(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	vpcNameTag := "terraform-testacc-ecs-service-health-check-grace-period"
	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-hcgps-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-hcgps-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-svc-w-hcgps-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-svc-w-hcgps-%s", rString)
	tgName := fmt.Sprintf("tf-acc-tg-svc-w-hcgps-%s", rString)
	lbName := fmt.Sprintf("tf-acc-lb-svc-w-hcgps-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-hcgps-%s", rString)

	resourceName := "aws_ecs_service.with_alb"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(vpcNameTag, clusterName, tdName,
					roleName, policyName, tgName, lbName, svcName, -1),
				ExpectError: regexp.MustCompile(`must be between 0 and 7200`),
			},
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(vpcNameTag, clusterName, tdName,
					roleName, policyName, tgName, lbName, svcName, 7201),
				ExpectError: regexp.MustCompile(`must be between 0 and 7200`),
			},
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(vpcNameTag, clusterName, tdName,
					roleName, policyName, tgName, lbName, svcName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "300"),
				),
			},
			{
				Config: testAccAWSEcsService_healthCheckGracePeriodSeconds(vpcNameTag, clusterName, tdName,
					roleName, policyName, tgName, lbName, svcName, 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period_seconds", "600"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withIamRole(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-iam-role-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-iam-role-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-svc-w-iam-role-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-svc-w-iam-role-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-iam-role-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withIamRole(clusterName, tdName, roleName, policyName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.ghost", &service),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withDeploymentValues(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-dv-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-dv-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-dv-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithDeploymentValues(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr(
						"aws_ecs_service.mongo", "deployment_maximum_percent", "200"),
					resource.TestCheckResourceAttr(
						"aws_ecs_service.mongo", "deployment_minimum_healthy_percent", "100"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3444
func TestAccAWSEcsService_withLbChanges(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-lbc-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-lbc-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-svc-w-lbc-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-svc-w-lbc-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-lbc-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withLbChanges(clusterName, tdName, roleName, policyName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.with_lb_changes", &service),
				),
			},
			{
				Config: testAccAWSEcsService_withLbChanges_modified(clusterName, tdName, roleName, policyName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.with_lb_changes", &service),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/3361
func TestAccAWSEcsService_withEcsClusterName(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-cluster-name-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-cluster-name-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-cluster-name-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithEcsClusterName(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.jenkins", &service),
					resource.TestCheckResourceAttr(
						"aws_ecs_service.jenkins", "cluster", clusterName),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withAlb(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-alb-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-alb-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-svc-w-alb-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-svc-w-alb-%s", rString)
	tgName := fmt.Sprintf("tf-acc-tg-svc-w-alb-%s", rString)
	lbName := fmt.Sprintf("tf-acc-lb-svc-w-alb-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-alb-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithAlb(clusterName, tdName, roleName, policyName, tgName, lbName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.with_alb", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.with_alb", "load_balancer.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withPlacementStrategy(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-ps-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-ps-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-ps-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.#", "0"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithPlacementStrategy(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.#", "1"),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.0.field", "memory"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithMultiPlacementStrategy(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.#", "2"),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.0.type", "binpack"),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.0.field", "memory"),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.1.type", "spread"),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "ordered_placement_strategy.1.field", "instanceId"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withPlacementConstraints(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-pc-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-pc-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-pc-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithPlacementConstraint(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withPlacementConstraints_emptyExpression(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-pc-ee-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-pc-ee-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-pc-ee-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithPlacementConstraintEmptyExpression(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.mongo", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.mongo", "placement_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeFargate(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	sg1Name := fmt.Sprintf("tf-acc-sg-1-svc-w-ltf-%s", rString)
	sg2Name := fmt.Sprintf("tf-acc-sg-2-svc-w-ltf-%s", rString)
	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-ltf-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-ltf-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-ltf-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargate(sg1Name, sg2Name, clusterName, tdName, svcName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.main", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.subnets.#", "2"),
				),
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargate(sg1Name, sg2Name, clusterName, tdName, svcName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.main", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.assign_public_ip", "true"),
				),
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithLaunchTypeFargate(sg1Name, sg2Name, clusterName, tdName, svcName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.main", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.assign_public_ip", "false"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withLaunchTypeEC2AndNetworkConfiguration(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	sg1Name := fmt.Sprintf("tf-acc-sg-1-svc-w-nc-%s", rString)
	sg2Name := fmt.Sprintf("tf-acc-sg-2-svc-w-nc-%s", rString)
	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-nc-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-nc-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-nc-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsServiceWithNetworkConfiguration(sg1Name, sg2Name, clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.main", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.subnets.#", "2"),
				),
			},
			{
				Config: testAccAWSEcsServiceWithNetworkConfiguration_modified(sg1Name, sg2Name, clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.main", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr("aws_ecs_service.main", "network_configuration.0.subnets.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSEcsService_withServiceRegistries(t *testing.T) {
	var service ecs.Service
	rString := acctest.RandString(8)

	clusterName := fmt.Sprintf("tf-acc-cluster-svc-w-ups-%s", rString)
	tdName := fmt.Sprintf("tf-acc-td-svc-w-ups-%s", rString)
	svcName := fmt.Sprintf("tf-acc-svc-w-ups-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsService_withServiceRegistries(rString, clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsServiceExists("aws_ecs_service.test", &service),
					resource.TestCheckResourceAttr("aws_ecs_service.test", "service_registries.#", "1"),
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

func testAccAWSEcsService(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceModified(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 2
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithInterchangeablePlacementStrategy(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  placement_strategy {
    field = "host"
    type = "spread"
  }
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithPlacementStrategy(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  ordered_placement_strategy {
    type = "binpack"
    field = "memory"
  }
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithMultiPlacementStrategy(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  ordered_placement_strategy {
    type = "binpack"
    field = "memory"
  }
  ordered_placement_strategy {
    field = "host"
    type = "spread"
  }
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithPlacementConstraint(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  placement_constraints {
    type = "memberOf"
    expression = "attribute:ecs.availability-zone in [us-west-2a, us-west-2b]"
  }
}
	`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithPlacementConstraintEmptyExpression(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}
resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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
resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  placement_constraints {
    type = "distinctInstance"
  }
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithLaunchTypeFargate(sg1Name, sg2Name, clusterName, tdName, svcName, assignPublicIP string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}
data "aws_availability_zones" "available" {}

resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags {
    Name = "terraform-testacc-ecs-service-with-launch-type-fargate"
  }
}

resource "aws_subnet" "main" {
  count = 2
  cidr_block = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id = "${aws_vpc.main.id}"
  tags {
    Name = "tf-acc-ecs-service-with-launch-type-fargate"
  }
}

resource "aws_security_group" "allow_all_a" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}

resource "aws_security_group" "allow_all_b" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}

resource "aws_ecs_cluster" "main" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
  network_mode = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu = "256"
  memory = "512"

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

resource "aws_ecs_service" "main" {
  name = "%s"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  launch_type = "FARGATE"
  network_configuration {
    security_groups = ["${aws_security_group.allow_all_a.id}", "${aws_security_group.allow_all_b.id}"]
    subnets = ["${aws_subnet.main.*.id}"]
    assign_public_ip = %s
  }
}
`, sg1Name, sg2Name, clusterName, tdName, svcName, assignPublicIP)
}

func testAccAWSEcsService_healthCheckGracePeriodSeconds(vpcNameTag, clusterName, tdName, roleName, policyName,
	tgName, lbName, svcName string, healthCheckGracePeriodSeconds int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags {
    Name = "%s"
  }
}

resource "aws_subnet" "main" {
  count = 2
  cidr_block = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id = "${aws_vpc.main.id}"
  tags {
    Name = "tf-acc-ecs-service-health-check-grace-period"
  }
}

resource "aws_ecs_cluster" "main" {
  name = "%s"
}

resource "aws_ecs_task_definition" "with_lb_changes" {
  family = "%s"
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
  name = "%s"
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
  name = "%s"
  role = "${aws_iam_role.ecs_service.name}"
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
  name = "%s"
  port = 80
  protocol = "HTTP"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_lb" "main" {
  name     = "%s"
  internal = true
  subnets  = ["${aws_subnet.main.*.id}"]
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.main.id}"
  port = "80"
  protocol = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.id}"
    type             = "forward"
  }
}

resource "aws_ecs_service" "with_alb" {
  name = "%s"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.with_lb_changes.arn}"
  desired_count = 1
  health_check_grace_period_seconds = %d
  iam_role = "${aws_iam_role.ecs_service.name}"

  load_balancer {
    target_group_arn = "${aws_lb_target_group.test.id}"
    container_name = "ghost"
    container_port = "2368"
  }

  depends_on = [
    "aws_iam_role_policy.ecs_service",
    "aws_lb_listener.front_end"
  ]
}
`, vpcNameTag, clusterName, tdName, roleName, policyName,
		tgName, lbName, svcName, healthCheckGracePeriodSeconds)
}

func testAccAWSEcsService_withIamRole(clusterName, tdName, roleName, policyName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "main" {
  name = "%s"
}

resource "aws_ecs_task_definition" "ghost" {
  family = "%s"
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
    name = "%s"
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
    name = "%s"
    role = "${aws_iam_role.ecs_service.name}"
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

resource "aws_elb" "main" {
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 8080
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

resource "aws_ecs_service" "ghost" {
  name = "%s"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.ghost.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_service.name}"

  load_balancer {
    elb_name = "${aws_elb.main.id}"
    container_name = "ghost"
    container_port = "2368"
  }

  depends_on = ["aws_iam_role_policy.ecs_service"]
}
`, clusterName, tdName, roleName, policyName, svcName)
}

func testAccAWSEcsServiceWithDeploymentValues(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "mongo" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
}
`, clusterName, tdName, svcName)
}

func tpl_testAccAWSEcsService_withLbChanges(clusterName, tdName, image,
	containerName string, containerPort, hostPort int, roleName, policyName string,
	instancePort int, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "main" {
  name = "%[1]s"
}

resource "aws_ecs_task_definition" "with_lb_changes" {
  family = "%[2]s"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "%[3]s",
    "memory": 128,
    "name": "%[4]s",
    "portMappings": [
      {
        "containerPort": %[5]d,
        "hostPort": %[6]d
      }
    ]
  }
]
DEFINITION
}

resource "aws_iam_role" "ecs_service" {
    name = "%[7]s"
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
    name = "%[8]s"
    role = "${aws_iam_role.ecs_service.name}"
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

resource "aws_elb" "main" {
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = %[6]d
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

resource "aws_ecs_service" "with_lb_changes" {
  name = "%[10]s"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.with_lb_changes.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_service.name}"

  load_balancer {
    elb_name = "${aws_elb.main.id}"
    container_name = "%[4]s"
    container_port = "%[5]d"
  }

  depends_on = ["aws_iam_role_policy.ecs_service"]
}
`, clusterName, tdName, image, containerName, containerPort, hostPort, roleName, policyName, instancePort, svcName)
}

func testAccAWSEcsService_withLbChanges(clusterName, tdName, roleName, policyName, svcName string) string {
	return tpl_testAccAWSEcsService_withLbChanges(
		clusterName, tdName, "ghost:latest", "ghost", 2368, 8080, roleName, policyName, 2368, svcName)
}

func testAccAWSEcsService_withLbChanges_modified(clusterName, tdName, roleName, policyName, svcName string) string {
	return tpl_testAccAWSEcsService_withLbChanges(
		clusterName, tdName, "nginx:latest", "nginx", 80, 8080, roleName, policyName, 80, svcName)
}

func testAccAWSEcsServiceWithFamilyAndRevision(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
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

resource "aws_ecs_service" "jenkins" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.jenkins.family}:${aws_ecs_task_definition.jenkins.revision}"
  desired_count = 1
}`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithFamilyAndRevisionModified(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
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

resource "aws_ecs_service" "jenkins" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.jenkins.family}:${aws_ecs_task_definition.jenkins.revision}"
  desired_count = 1
}`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithRenamedCluster(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}
resource "aws_ecs_task_definition" "ghost" {
  family = "%s"
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
resource "aws_ecs_service" "ghost" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.ghost.family}:${aws_ecs_task_definition.ghost.revision}"
  desired_count = 1
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithEcsClusterName(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "jenkins" {
  family = "%s"
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

resource "aws_ecs_service" "jenkins" {
  name = "%s"
  cluster = "${aws_ecs_cluster.default.name}"
  task_definition = "${aws_ecs_task_definition.jenkins.arn}"
  desired_count = 1
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsServiceWithAlb(clusterName, tdName, roleName, policyName, tgName, lbName, svcName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags {
    Name = "terraform-testacc-ecs-service-with-alb"
  }
}

resource "aws_subnet" "main" {
  count = 2
  cidr_block = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id = "${aws_vpc.main.id}"
  tags {
    Name = "tf-acc-ecs-service-with-alb"
  }
}

resource "aws_ecs_cluster" "main" {
  name = "%s"
}

resource "aws_ecs_task_definition" "with_lb_changes" {
  family = "%s"
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
    name = "%s"
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
    name = "%s"
    role = "${aws_iam_role.ecs_service.name}"
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
  name = "%s"
  port = 80
  protocol = "HTTP"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_lb" "main" {
  name            = "%s"
  internal        = true
  subnets         = ["${aws_subnet.main.*.id}"]
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.main.id}"
  port = "80"
  protocol = "HTTP"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.id}"
    type = "forward"
  }
}

resource "aws_ecs_service" "with_alb" {
  name = "%s"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.with_lb_changes.arn}"
  desired_count = 1
  iam_role = "${aws_iam_role.ecs_service.name}"

  load_balancer {
    target_group_arn = "${aws_lb_target_group.test.id}"
    container_name = "ghost"
    container_port = "2368"
  }

  depends_on = [
    "aws_iam_role_policy.ecs_service",
    "aws_lb_listener.front_end"
  ]
}
`, clusterName, tdName, roleName, policyName, tgName, lbName, svcName)
}

func testAccAWSEcsServiceWithNetworkConfiguration(sg1Name, sg2Name, clusterName, tdName, svcName string) string {
	return tpl_testAccAWSEcsServiceWithNetworkConfiguration(
		sg1Name, sg2Name, clusterName, tdName, svcName,
		`"${aws_security_group.allow_all_a.id}", "${aws_security_group.allow_all_b.id}"`,
	)
}
func testAccAWSEcsServiceWithNetworkConfiguration_modified(sg1Name, sg2Name, clusterName, tdName, svcName string) string {
	return tpl_testAccAWSEcsServiceWithNetworkConfiguration(
		sg1Name, sg2Name, clusterName, tdName, svcName,
		`"${aws_security_group.allow_all_a.id}"`,
	)
}

func tpl_testAccAWSEcsServiceWithNetworkConfiguration(sg1Name, sg2Name, clusterName, tdName, svcName string, securityGroups string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags {
    Name = "terraform-testacc-ecs-service-with-network-config"
  }
}

resource "aws_subnet" "main" {
  count = 2
  cidr_block = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id = "${aws_vpc.main.id}"
  tags {
    Name = "tf-acc-ecs-service-with-network-config"
  }
}

resource "aws_security_group" "allow_all_a" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}

resource "aws_security_group" "allow_all_b" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    protocol = "6"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}

resource "aws_ecs_cluster" "main" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
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

resource "aws_ecs_service" "main" {
  name = "%s"
  cluster = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
  network_configuration {
    security_groups = [%s]
    subnets = ["${aws_subnet.main.*.id}"]
  }
}
`, sg1Name, sg2Name, clusterName, tdName, svcName, securityGroups)
}

func testAccAWSEcsService_withServiceRegistries(rName, clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2
  cidr_block = "${cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.test.names[count.index]}"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_security_group" "test" {
  name        = "tf-acc-sg-%s"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol = "-1"
    from_port = 0
    to_port = 0
    cidr_blocks = ["${aws_vpc.test.cidr_block}"]
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "tf-acc-sd-%s.terraform.local"
  description = "test"
  vpc = "${aws_vpc.test.id}"
}

resource "aws_service_discovery_service" "test" {
  name = "tf-acc-sd-%s"
  dns_config {
    namespace_id = "${aws_service_discovery_private_dns_namespace.test.id}"
    dns_records {
      ttl = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = "%s"
}

resource "aws_ecs_task_definition" "test" {
  family = "%s"
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
  name = "%s"
  cluster = "${aws_ecs_cluster.test.id}"
  task_definition = "${aws_ecs_task_definition.test.arn}"
  desired_count = 1
  service_registries {
    port = 34567
    registry_arn = "${aws_service_discovery_service.test.arn}"
  }
  network_configuration {
    security_groups = ["${aws_security_group.test.id}"]
    subnets = ["${aws_subnet.test.*.id}"]
  }
}
`, rName, rName, rName, clusterName, tdName, svcName)
}
