package aws

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elb", &resource.Sweeper{
		Name: "aws_elb",
		F:    testSweepELBs,
	})
}

func testSweepELBs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).elbconn

	prefixes := []string{
		"test-elb-",
	}

	err = conn.DescribeLoadBalancersPages(&elb.DescribeLoadBalancersInput{}, func(out *elb.DescribeLoadBalancersOutput, isLast bool) bool {
		if len(out.LoadBalancerDescriptions) == 0 {
			log.Println("[INFO] No ELBs found for sweeping")
			return false
		}

		for _, lb := range out.LoadBalancerDescriptions {
			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(*lb.LoadBalancerName, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping ELB: %s", *lb.LoadBalancerName)
				continue
			}
			log.Printf("[INFO] Deleting ELB: %s", *lb.LoadBalancerName)

			_, err := conn.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{
				LoadBalancerName: lb.LoadBalancerName,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete ELB %s: %s", *lb.LoadBalancerName, err)
				continue
			}
			err = cleanupELBNetworkInterfaces(client.(*AWSClient).ec2conn, *lb.LoadBalancerName)
			if err != nil {
				log.Printf("[WARN] Failed to cleanup ENIs for ELB %q: %s", *lb.LoadBalancerName, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ELB sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving ELBs: %s", err)
	}
	return nil
}

func TestAccAWSELB_importBasic(t *testing.T) {
	resourceName := "aws_elb.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSELB_basic(t *testing.T) {
	var conf elb.LoadBalancerDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testAccCheckAWSELBAttributes(&conf),
					resource.TestCheckResourceAttrSet("aws_elb.bar", "arn"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "subnets.#", "3"),
					// NOTE: Subnet IDs are different across AWS accounts and cannot be checked.
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "listener.206423021.lb_protocol", "http"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "cross_zone_load_balancing", "true"),
				),
			},
		},
	})
}

func TestAccAWSELB_disappears(t *testing.T) {
	var loadBalancer elb.LoadBalancerDescription
	resourceName := "aws_elb.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &loadBalancer),
					testAccCheckAWSELBDisappears(&loadBalancer),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSELB_fullCharacterRange(t *testing.T) {
	var conf elb.LoadBalancerDescription

	lbName := fmt.Sprintf("Tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAWSELBFullRangeOfCharacters, lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "name", lbName),
				),
			},
		},
	})
}

func TestAccAWSELB_AccessLogs_enabled(t *testing.T) {
	var conf elb.LoadBalancerDescription

	rName := fmt.Sprintf("terraform-access-logs-bucket-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBAccessLogs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
				),
			},

			{
				Config: testAccAWSELBAccessLogsOn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.0.interval", "5"),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.0.enabled", "true"),
				),
			},

			{
				Config: testAccAWSELBAccessLogs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSELB_AccessLogs_disabled(t *testing.T) {
	var conf elb.LoadBalancerDescription

	rName := fmt.Sprintf("terraform-access-logs-bucket-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBAccessLogs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
				),
			},

			{
				Config: testAccAWSELBAccessLogsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.0.bucket", rName),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.0.interval", "5"),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.0.enabled", "false"),
				),
			},

			{
				Config: testAccAWSELBAccessLogs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.foo", "access_logs.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSELB_namePrefix(t *testing.T) {
	var conf elb.LoadBalancerDescription
	nameRegex := regexp.MustCompile("^test-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELB_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.test", &conf),
					resource.TestMatchResourceAttr(
						"aws_elb.test", "name", nameRegex),
				),
			},
		},
	})
}

func TestAccAWSELB_generatedName(t *testing.T) {
	var conf elb.LoadBalancerDescription
	generatedNameRegexp := regexp.MustCompile("^tf-lb-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestMatchResourceAttr(
						"aws_elb.foo", "name", generatedNameRegexp),
				),
			},
		},
	})
}

func TestAccAWSELB_generatesNameForZeroValue(t *testing.T) {
	var conf elb.LoadBalancerDescription
	generatedNameRegexp := regexp.MustCompile("^tf-lb-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELB_zeroValueName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.foo", &conf),
					resource.TestMatchResourceAttr(
						"aws_elb.foo", "name", generatedNameRegexp),
				),
			},
		},
	})
}

func TestAccAWSELB_availabilityZones(t *testing.T) {
	var conf elb.LoadBalancerDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.2050015877", "us-west-2c"),
				),
			},

			{
				Config: testAccAWSELBConfig_AvailabilityZonesUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "availability_zones.221770259", "us-west-2b"),
				),
			},
		},
	})
}

func TestAccAWSELB_tags(t *testing.T) {
	var conf elb.LoadBalancerDescription
	var td elb.TagDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testAccCheckAWSELBAttributes(&conf),
					testAccLoadTags(&conf, &td),
					testAccCheckELBTags(&td.Tags, "bar", "baz"),
				),
			},

			{
				Config: testAccAWSELBConfig_TagUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testAccCheckAWSELBAttributes(&conf),
					testAccLoadTags(&conf, &td),
					testAccCheckELBTags(&td.Tags, "foo", "bar"),
					testAccCheckELBTags(&td.Tags, "new", "type"),
				),
			},
		},
	})
}

func TestAccAWSELB_Listener_SSLCertificateID_IAMServerCertificate(t *testing.T) {
	var conf elb.LoadBalancerDescription
	rName := fmt.Sprintf("tf-acctest-%s", acctest.RandString(10))
	resourceName := "aws_elb.bar"

	testCheck := func(*terraform.State) error {
		if len(conf.ListenerDescriptions) != 1 {
			return fmt.Errorf(
				"TestAccAWSELB_iam_server_cert expected 1 listener, got %d",
				len(conf.ListenerDescriptions))
		}
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccELBConfig_Listener_IAMServerCertificate(rName, "tcp"),
				ExpectError: regexp.MustCompile(`ssl_certificate_id may be set only when protocol is 'https' or 'ssl'`),
			},
			{
				Config: testAccELBConfig_Listener_IAMServerCertificate(rName, "https"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					testCheck,
				),
			},
			{
				Config:      testAccELBConfig_Listener_IAMServerCertificate_AddInvalidListener(rName),
				ExpectError: regexp.MustCompile(`ssl_certificate_id may be set only when protocol is 'https' or 'ssl'`),
			},
		},
	})
}

func TestAccAWSELB_swap_subnets(t *testing.T) {
	var conf elb.LoadBalancerDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.ourapp",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig_subnets,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.ourapp", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.ourapp", "subnets.#", "2"),
				),
			},

			{
				Config: testAccAWSELBConfig_subnet_swap,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.ourapp", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.ourapp", "subnets.#", "2"),
				),
			},
		},
	})
}

func testAccLoadTags(conf *elb.LoadBalancerDescription, td *elb.TagDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elbconn

		describe, err := conn.DescribeTags(&elb.DescribeTagsInput{
			LoadBalancerNames: []*string{conf.LoadBalancerName},
		})

		if err != nil {
			return err
		}
		if len(describe.TagDescriptions) > 0 {
			*td = *describe.TagDescriptions[0]
		}
		return nil
	}
}

func TestAccAWSELB_InstanceAttaching(t *testing.T) {
	var conf elb.LoadBalancerDescription

	testCheckInstanceAttached := func(count int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(conf.Instances) != count {
				return fmt.Errorf("instance count does not match")
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testAccCheckAWSELBAttributes(&conf),
				),
			},

			{
				Config: testAccAWSELBConfigNewInstance,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(1),
				),
			},
		},
	})
}

func TestAccAWSELB_listener(t *testing.T) {
	var conf elb.LoadBalancerDescription
	resourceName := "aws_elb.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_protocol", "http"),
				),
			},
			{
				Config: testAccAWSELBConfigListener_multipleListeners,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "listener.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.829854800.instance_port", "22"),
					resource.TestCheckResourceAttr(resourceName, "listener.829854800.instance_protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "listener.829854800.lb_port", "22"),
					resource.TestCheckResourceAttr(resourceName, "listener.829854800.lb_protocol", "tcp"),
				),
			},
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_protocol", "http"),
				),
			},
			{
				Config: testAccAWSELBConfigListener_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "listener.3931999347.instance_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "listener.3931999347.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.3931999347.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.3931999347.lb_protocol", "http"),
				),
			},
			{
				PreConfig: func() {
					// Simulate out of band listener removal
					conn := testAccProvider.Meta().(*AWSClient).elbconn
					input := &elb.DeleteLoadBalancerListenersInput{
						LoadBalancerName:  conf.LoadBalancerName,
						LoadBalancerPorts: []*int64{aws.Int64(int64(80))},
					}
					if _, err := conn.DeleteLoadBalancerListeners(input); err != nil {
						t.Fatalf("Error deleting listener: %s", err)
					}
				},
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_protocol", "http"),
				),
			},
			{
				PreConfig: func() {
					// Simulate out of band listener addition
					conn := testAccProvider.Meta().(*AWSClient).elbconn
					input := &elb.CreateLoadBalancerListenersInput{
						LoadBalancerName: conf.LoadBalancerName,
						Listeners: []*elb.Listener{
							{
								InstancePort:     aws.Int64(int64(22)),
								InstanceProtocol: aws.String("tcp"),
								LoadBalancerPort: aws.Int64(int64(22)),
								Protocol:         aws.String("tcp"),
							},
						},
					}
					if _, err := conn.CreateLoadBalancerListeners(input); err != nil {
						t.Fatalf("Error creating listener: %s", err)
					}
				},
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.instance_protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "listener.206423021.lb_protocol", "http"),
				),
			},
		},
	})
}

func TestAccAWSELB_HealthCheck(t *testing.T) {
	var conf elb.LoadBalancerDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfigHealthCheck,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testAccCheckAWSELBAttributesHealthCheck(&conf),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.healthy_threshold", "5"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.target", "HTTP:8000/"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.timeout", "30"),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.interval", "60"),
				),
			},
		},
	})
}

func TestAccAWSELBUpdate_HealthCheck(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfigHealthCheck,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.healthy_threshold", "5"),
				),
			},
			{
				Config: testAccAWSELBConfigHealthCheck_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "health_check.0.healthy_threshold", "10"),
				),
			},
		},
	})
}

func TestAccAWSELB_Timeout(t *testing.T) {
	var conf elb.LoadBalancerDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfigIdleTimeout,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "idle_timeout", "200",
					),
				),
			},
		},
	})
}

func TestAccAWSELBUpdate_Timeout(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfigIdleTimeout,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "idle_timeout", "200",
					),
				),
			},
			{
				Config: testAccAWSELBConfigIdleTimeout_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "idle_timeout", "400",
					),
				),
			},
		},
	})
}

func TestAccAWSELB_ConnectionDraining(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfigConnectionDraining,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining", "true",
					),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining_timeout", "400",
					),
				),
			},
		},
	})
}

func TestAccAWSELBUpdate_ConnectionDraining(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfigConnectionDraining,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining", "true",
					),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining_timeout", "400",
					),
				),
			},
			{
				Config: testAccAWSELBConfigConnectionDraining_update_timeout,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining", "true",
					),
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining_timeout", "600",
					),
				),
			},
			{
				Config: testAccAWSELBConfigConnectionDraining_update_disable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "connection_draining", "false",
					),
				),
			},
		},
	})
}

func TestAccAWSELB_SecurityGroups(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBConfig,
				Check: resource.ComposeTestCheckFunc(
					// ELBs get a default security group
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "security_groups.#", "1",
					),
				),
			},
			{
				Config: testAccAWSELBConfigSecurityGroups,
				Check: resource.ComposeTestCheckFunc(
					// Count should still be one as we swap in a custom security group
					resource.TestCheckResourceAttr(
						"aws_elb.bar", "security_groups.#", "1",
					),
				),
			},
		},
	})
}

// Unit test for listeners hash
func TestResourceAwsElbListenerHash(t *testing.T) {
	cases := map[string]struct {
		Left  map[string]interface{}
		Right map[string]interface{}
		Match bool
	}{
		"protocols are case insensitive": {
			map[string]interface{}{
				"instance_port":     80,
				"instance_protocol": "TCP",
				"lb_port":           80,
				"lb_protocol":       "TCP",
			},
			map[string]interface{}{
				"instance_port":     80,
				"instance_protocol": "Tcp",
				"lb_port":           80,
				"lb_protocol":       "tcP",
			},
			true,
		},
	}

	for tn, tc := range cases {
		leftHash := resourceAwsElbListenerHash(tc.Left)
		rightHash := resourceAwsElbListenerHash(tc.Right)
		if leftHash == rightHash != tc.Match {
			t.Fatalf("%s: expected match: %t, but did not get it", tn, tc.Match)
		}
	}
}

func TestResourceAWSELB_validateElbNameCannotBeginWithHyphen(t *testing.T) {
	var elbName = "-Testing123"
	_, errors := validateElbName(elbName, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestResourceAWSELB_validateElbNameCanBeAnEmptyString(t *testing.T) {
	var elbName = ""
	_, errors := validateElbName(elbName, "SampleKey")

	if len(errors) != 0 {
		t.Fatalf("Expected the ELB Name to pass validation")
	}
}

func TestResourceAWSELB_validateElbNameCannotBeLongerThan32Characters(t *testing.T) {
	var elbName = "Testing123dddddddddddddddddddvvvv"
	_, errors := validateElbName(elbName, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestResourceAWSELB_validateElbNameCannotHaveSpecialCharacters(t *testing.T) {
	var elbName = "Testing123%%"
	_, errors := validateElbName(elbName, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestResourceAWSELB_validateElbNameCannotEndWithHyphen(t *testing.T) {
	var elbName = "Testing123-"
	_, errors := validateElbName(elbName, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestResourceAWSELB_validateAccessLogsInterval(t *testing.T) {
	type testCases struct {
		Value    int
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    0,
			ErrCount: 1,
		},
		{
			Value:    10,
			ErrCount: 1,
		},
		{
			Value:    -1,
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validateAccessLogsInterval(tc.Value, "interval")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

}

func TestResourceAWSELB_validateHealthCheckTarget(t *testing.T) {
	type testCase struct {
		Value    string
		ErrCount int
	}

	randomRunes := func(n int) string {
		rand.Seed(time.Now().UTC().UnixNano())

		// A complete set of modern Katakana characters.
		runes := []rune("アイウエオ" +
			"カキクケコガギグゲゴサシスセソザジズゼゾ" +
			"タチツテトダヂヅデドナニヌネノハヒフヘホ" +
			"バビブベボパピプペポマミムメモヤユヨラリ" +
			"ルレロワヰヱヲン")

		s := make([]rune, n)
		for i := range s {
			s[i] = runes[rand.Intn(len(runes))]
		}
		return string(s)
	}

	validCases := []testCase{
		{
			Value:    "TCP:1234",
			ErrCount: 0,
		},
		{
			Value:    "http:80/test",
			ErrCount: 0,
		},
		{
			Value:    fmt.Sprintf("HTTP:8080/%s", randomRunes(5)),
			ErrCount: 0,
		},
		{
			Value:    "SSL:8080",
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := validateHeathCheckTarget(tc.Value, "target")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}

	invalidCases := []testCase{
		{
			Value:    "",
			ErrCount: 1,
		},
		{
			Value:    "TCP:",
			ErrCount: 1,
		},
		{
			Value:    "TCP:1234/",
			ErrCount: 1,
		},
		{
			Value:    "SSL:8080/",
			ErrCount: 1,
		},
		{
			Value:    "HTTP:8080",
			ErrCount: 1,
		},
		{
			Value:    "incorrect-value",
			ErrCount: 1,
		},
		{
			Value:    "TCP:123456",
			ErrCount: 1,
		},
		{
			Value:    "incorrect:80/",
			ErrCount: 1,
		},
		{
			Value:    fmt.Sprintf("HTTP:8080/%s%s", randomString(512), randomRunes(512)),
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validateHeathCheckTarget(tc.Value, "target")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}
}

func testAccCheckAWSELBDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elb" {
			continue
		}

		describe, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.LoadBalancerDescriptions) != 0 &&
				*describe.LoadBalancerDescriptions[0].LoadBalancerName == rs.Primary.ID {
				return fmt.Errorf("ELB still exists")
			}
		}

		// Verify the error
		providerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		if providerErr.Code() != "LoadBalancerNotFound" {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}

	return nil
}

func testAccCheckAWSELBDisappears(loadBalancer *elb.LoadBalancerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elbconn

		input := elb.DeleteLoadBalancerInput{
			LoadBalancerName: loadBalancer.LoadBalancerName,
		}
		_, err := conn.DeleteLoadBalancer(&input)

		return err
	}
}

func testAccCheckAWSELBAttributes(conf *elb.LoadBalancerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		zones := []string{"us-west-2a", "us-west-2b", "us-west-2c"}
		azs := make([]string, 0, len(conf.AvailabilityZones))
		for _, x := range conf.AvailabilityZones {
			azs = append(azs, *x)
		}
		sort.StringSlice(azs).Sort()
		if !reflect.DeepEqual(azs, zones) {
			return fmt.Errorf("bad availability_zones")
		}

		l := elb.Listener{
			InstancePort:     aws.Int64(int64(8000)),
			InstanceProtocol: aws.String("HTTP"),
			LoadBalancerPort: aws.Int64(int64(80)),
			Protocol:         aws.String("HTTP"),
		}

		if !reflect.DeepEqual(conf.ListenerDescriptions[0].Listener, &l) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				conf.ListenerDescriptions[0].Listener,
				l)
		}

		if *conf.DNSName == "" {
			return fmt.Errorf("empty dns_name")
		}

		return nil
	}
}

func testAccCheckAWSELBAttributesHealthCheck(conf *elb.LoadBalancerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		zones := []string{"us-west-2a", "us-west-2b", "us-west-2c"}
		azs := make([]string, 0, len(conf.AvailabilityZones))
		for _, x := range conf.AvailabilityZones {
			azs = append(azs, *x)
		}
		sort.StringSlice(azs).Sort()
		if !reflect.DeepEqual(azs, zones) {
			return fmt.Errorf("bad availability_zones")
		}

		check := &elb.HealthCheck{
			Timeout:            aws.Int64(int64(30)),
			UnhealthyThreshold: aws.Int64(int64(5)),
			HealthyThreshold:   aws.Int64(int64(5)),
			Interval:           aws.Int64(int64(60)),
			Target:             aws.String("HTTP:8000/"),
		}

		if !reflect.DeepEqual(conf.HealthCheck, check) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				conf.HealthCheck,
				check)
		}

		if *conf.DNSName == "" {
			return fmt.Errorf("empty dns_name")
		}

		return nil
	}
}

func testAccCheckAWSELBExists(n string, res *elb.LoadBalancerDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ELB ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbconn

		describe, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.LoadBalancerDescriptions) != 1 ||
			*describe.LoadBalancerDescriptions[0].LoadBalancerName != rs.Primary.ID {
			return fmt.Errorf("ELB not found")
		}

		*res = *describe.LoadBalancerDescriptions[0]

		// Confirm source_security_group_id for ELBs in a VPC
		// 	See https://github.com/hashicorp/terraform/pull/3780
		if res.VPCId != nil {
			sgid := rs.Primary.Attributes["source_security_group_id"]
			if sgid == "" {
				return fmt.Errorf("Expected to find source_security_group_id for ELB, but was empty")
			}
		}

		return nil
	}
}

const testAccAWSELBConfig = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    // Protocol should be case insensitive
    lb_protocol = "HttP"
  }

	tags = {
		bar = "baz"
	}

  cross_zone_load_balancing = true
}
`

const testAccAWSELBFullRangeOfCharacters = `
resource "aws_elb" "foo" {
  name = "%s"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

const testAccAWSELBAccessLogs = `
resource "aws_elb" "foo" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

func testAccAWSELBAccessLogsOn(r string) string {
	return fmt.Sprintf(`
# an S3 bucket configured for Access logs
# The 797873946194 is the AWS ID for us-west-2, so this test
# must be ran in us-west-2
resource "aws_s3_bucket" "acceslogs_bucket" {
  bucket = "%s"
  acl = "private"
  force_destroy = true
  policy = <<EOF
{
  "Id": "Policy1446577137248",
  "Statement": [
    {
      "Action": "s3:PutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::797873946194:root"
      },
      "Resource": "arn:aws:s3:::%s/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}

resource "aws_elb" "foo" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

	access_logs {
		interval = 5
		bucket = "${aws_s3_bucket.acceslogs_bucket.bucket}"
	}
}
`, r, r)
}

func testAccAWSELBAccessLogsDisabled(r string) string {
	return fmt.Sprintf(`
# an S3 bucket configured for Access logs
# The 797873946194 is the AWS ID for us-west-2, so this test
# must be ran in us-west-2
resource "aws_s3_bucket" "acceslogs_bucket" {
  bucket = "%s"
  acl = "private"
  force_destroy = true
  policy = <<EOF
{
  "Id": "Policy1446577137248",
  "Statement": [
    {
      "Action": "s3:PutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::797873946194:root"
      },
      "Resource": "arn:aws:s3:::%s/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}

resource "aws_elb" "foo" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

	access_logs {
		interval = 5
		bucket = "${aws_s3_bucket.acceslogs_bucket.bucket}"
		enabled = false
	}
}
`, r, r)
}

const testAccAWSELB_namePrefix = `
resource "aws_elb" "test" {
  name_prefix = "test-"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

const testAccAWSELBGeneratedName = `
resource "aws_elb" "foo" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

const testAccAWSELB_zeroValueName = `
resource "aws_elb" "foo" {
  name               = ""
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

# See https://github.com/terraform-providers/terraform-provider-aws/issues/2498
output "lb_name" {
  value = "${aws_elb.foo.name}"
}
`

const testAccAWSELBConfig_AvailabilityZonesUpdate = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

const testAccAWSELBConfig_TagUpdate = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

	tags = {
		foo = "bar"
		new = "type"
	}

  cross_zone_load_balancing = true
}
`

const testAccAWSELBConfigNewInstance = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  instances = ["${aws_instance.foo.id}"]
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-043a5034"
	instance_type = "t1.micro"
}
`

const testAccAWSELBConfigHealthCheck = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  health_check {
    healthy_threshold = 5
    unhealthy_threshold = 5
    target = "HTTP:8000/"
    interval = 60
    timeout = 30
  }
}
`

const testAccAWSELBConfigHealthCheck_update = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  health_check {
    healthy_threshold = 10
    unhealthy_threshold = 5
    target = "HTTP:8000/"
    interval = 60
    timeout = 30
  }
}
`

const testAccAWSELBConfigListener_update = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8080
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

const testAccAWSELBConfigListener_multipleListeners = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  listener {
    instance_port = 22
    instance_protocol = "tcp"
    lb_port = 22
    lb_protocol = "tcp"
  }
}
`

const testAccAWSELBConfigIdleTimeout = `
resource "aws_elb" "bar" {
	availability_zones = ["us-west-2a"]

	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}

	idle_timeout = 200
}
`

const testAccAWSELBConfigIdleTimeout_update = `
resource "aws_elb" "bar" {
	availability_zones = ["us-west-2a"]

	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}

	idle_timeout = 400
}
`

const testAccAWSELBConfigConnectionDraining = `
resource "aws_elb" "bar" {
	availability_zones = ["us-west-2a"]

	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}

	connection_draining = true
	connection_draining_timeout = 400
}
`

const testAccAWSELBConfigConnectionDraining_update_timeout = `
resource "aws_elb" "bar" {
	availability_zones = ["us-west-2a"]

	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}

	connection_draining = true
	connection_draining_timeout = 600
}
`

const testAccAWSELBConfigConnectionDraining_update_disable = `
resource "aws_elb" "bar" {
	availability_zones = ["us-west-2a"]

	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}

	connection_draining = false
}
`

const testAccAWSELBConfigSecurityGroups = `
resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  security_groups = ["${aws_security_group.bar.id}"]
}

resource "aws_security_group" "bar" {
  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

	tags = {
		Name = "tf_elb_sg_test"
	}
}
`

func testAccELBConfig_Listener_IAMServerCertificate(certName, lbProtocol string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

%[1]s

resource "aws_iam_server_certificate" "test_cert" {
  name             = "%[2]s"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
  private_key      = "${tls_private_key.example.private_key_pem}"
}

resource "aws_elb" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]

  listener {
    instance_port      = 443
    instance_protocol  = "%[3]s"
    lb_port            = 443
    lb_protocol        = "%[3]s"
    ssl_certificate_id = "${aws_iam_server_certificate.test_cert.arn}"
  }
}
`, testAccTLSServerCert, certName, lbProtocol)
}

func testAccELBConfig_Listener_IAMServerCertificate_AddInvalidListener(certName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

%[1]s

resource "aws_iam_server_certificate" "test_cert" {
  name             = "%[2]s"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
  private_key      = "${tls_private_key.example.private_key_pem}"
}

resource "aws_elb" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]

  listener {
    instance_port      = 443
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = "${aws_iam_server_certificate.test_cert.arn}"
  }

  # lb_protocol tcp and ssl_certificate_id is not valid
  listener {
    instance_port      = 8443
    instance_protocol  = "tcp"
    lb_port            = 8443
    lb_protocol        = "tcp"
    ssl_certificate_id = "${aws_iam_server_certificate.test_cert.arn}"
  }
}
`, testAccTLSServerCert, certName)
}

const testAccAWSELBConfig_subnets = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "azelb" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-elb-subnets"
  }
}

resource "aws_subnet" "public_a_one" {
  vpc_id = "${aws_vpc.azelb.id}"

  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-elb-subnets-a-one"
  }
}

resource "aws_subnet" "public_b_one" {
  vpc_id = "${aws_vpc.azelb.id}"

  cidr_block        = "10.1.7.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-elb-subnets-b-one"
  }
}

resource "aws_subnet" "public_a_two" {
  vpc_id = "${aws_vpc.azelb.id}"

  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-elb-subnets-a-two"
  }
}

resource "aws_elb" "ourapp" {
  name = "terraform-asg-deployment-example"

  subnets = [
    "${aws_subnet.public_a_one.id}",
    "${aws_subnet.public_b_one.id}",
  ]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.azelb.id}"

  tags = {
    Name = "main"
  }
}
`

const testAccAWSELBConfig_subnet_swap = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "azelb" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-elb-subnet-swap"
  }
}

resource "aws_subnet" "public_a_one" {
  vpc_id = "${aws_vpc.azelb.id}"

  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-elb-subnet-swap-a-one"
  }
}

resource "aws_subnet" "public_b_one" {
  vpc_id = "${aws_vpc.azelb.id}"

  cidr_block        = "10.1.7.0/24"
  availability_zone = "us-west-2b"
  tags = {
    Name = "tf-acc-elb-subnet-swap-b-one"
  }
}

resource "aws_subnet" "public_a_two" {
  vpc_id = "${aws_vpc.azelb.id}"

  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-elb-subnet-swap-a-two"
  }
}

resource "aws_elb" "ourapp" {
  name = "terraform-asg-deployment-example"

  subnets = [
    "${aws_subnet.public_a_two.id}",
    "${aws_subnet.public_b_one.id}",
  ]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.azelb.id}"

  tags = {
    Name = "main"
  }
}
`
