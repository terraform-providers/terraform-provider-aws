package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/globalaccelerator/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_globalaccelerator_accelerator", &resource.Sweeper{
		Name: "aws_globalaccelerator_accelerator",
		F:    testSweepGlobalAcceleratorAccelerators,
	})
}

func testSweepGlobalAcceleratorAccelerators(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).globalacceleratorconn

	input := &globalaccelerator.ListAcceleratorsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListAccelerators(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Global Accelerator Accelerators: %s", err)
		}

		for _, accelerator := range output.Accelerators {
			arn := aws.StringValue(accelerator.AcceleratorArn)

			errs := sweepGlobalAcceleratorListeners(client, accelerator.AcceleratorArn)
			if errs != nil {
				sweeperErrs = multierror.Append(sweeperErrs, errs)
			}

			r := resourceAwsGlobalAcceleratorAccelerator()
			d := r.Data(nil)
			d.SetId(arn)
			err = r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Global Accelerator Accelerator (%s): %s", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepGlobalAcceleratorListeners(client interface{}, acceleratorArn *string) *multierror.Error {
	conn := client.(*AWSClient).globalacceleratorconn
	var sweeperErrs *multierror.Error

	log.Printf("[INFO] deleting Listeners for Accelerator %s", *acceleratorArn)
	listenersInput := &globalaccelerator.ListListenersInput{
		AcceleratorArn: acceleratorArn,
	}
	listenersOutput, err := conn.ListListeners(listenersInput)
	if err != nil {
		sweeperErr := fmt.Errorf("error listing Global Accelerator Listeners for Accelerator (%s): %s", *acceleratorArn, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	for _, listener := range listenersOutput.Listeners {
		errs := sweepGlobalAcceleratorEndpointGroups(client, listener.ListenerArn)
		if errs != nil {
			sweeperErrs = multierror.Append(sweeperErrs, errs)
		}

		arn := aws.StringValue(listener.ListenerArn)

		r := resourceAwsGlobalAcceleratorListener()
		d := r.Data(nil)
		d.SetId(arn)
		err = r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Global Accelerator listener (%s): %s", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs
}

func sweepGlobalAcceleratorEndpointGroups(client interface{}, listenerArn *string) *multierror.Error {
	conn := client.(*AWSClient).globalacceleratorconn
	var sweeperErrs *multierror.Error

	log.Printf("[INFO] deleting Endpoint Groups for Listener %s", *listenerArn)
	input := &globalaccelerator.ListEndpointGroupsInput{
		ListenerArn: listenerArn,
	}
	output, err := conn.ListEndpointGroups(input)
	if err != nil {
		sweeperErr := fmt.Errorf("error listing Global Accelerator Endpoint Groups for Listener (%s): %s", *listenerArn, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	for _, endpoint := range output.EndpointGroups {
		arn := aws.StringValue(endpoint.EndpointGroupArn)

		r := resourceAwsGlobalAcceleratorEndpointGroup()
		d := r.Data(nil)
		d.SetId(arn)
		err = r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Global Accelerator endpoint group (%s): %s", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs
}

func TestAccAwsGlobalAcceleratorAccelerator_basic(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	ipRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	dnsNameRegex := regexp.MustCompile(`^a[a-f0-9]{16}\.awsglobalaccelerator\.com$`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckGlobalAccelerator(t) },
		ErrorCheck:   testAccErrorCheck(t, globalaccelerator.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAcceleratorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", ""),
					resource.TestMatchResourceAttr(resourceName, "dns_name", dnsNameRegex),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", "Z2BJ6XQ5FK7U4H"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.0", ipRegex),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.1", ipRegex),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAwsGlobalAcceleratorAccelerator_disappears(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckGlobalAccelerator(t) },
		ErrorCheck:   testAccErrorCheck(t, globalaccelerator.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAcceleratorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGlobalAcceleratorAccelerator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsGlobalAcceleratorAccelerator_update(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	newName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckGlobalAccelerator(t) },
		ErrorCheck:   testAccErrorCheck(t, globalaccelerator.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigEnabled(newName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", newName),
				),
			},
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func TestAccAwsGlobalAcceleratorAccelerator_attributes(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckGlobalAccelerator(t) },
		ErrorCheck:   testAccErrorCheck(t, globalaccelerator.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigAttributes(rName, false, "flow-logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigAttributes(rName, true, "flow-logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
				),
			},
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigAttributes(rName, true, "flow-logs-updated/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs-updated/"),
				),
			},
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigAttributes(rName, false, "flow-logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
				),
			},
		},
	})
}

func TestAccAwsGlobalAcceleratorAccelerator_tags(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckGlobalAccelerator(t) },
		ErrorCheck:   testAccErrorCheck(t, globalaccelerator.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
				Config: testAccGlobalAcceleratorAcceleratorConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGlobalAcceleratorAcceleratorConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccPreCheckGlobalAccelerator(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

	input := &globalaccelerator.ListAcceleratorsInput{}

	_, err := conn.ListAccelerators(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckGlobalAcceleratorAcceleratorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		_, err := finder.AcceleratorByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckGlobalAcceleratorAcceleratorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_globalaccelerator_accelerator" {
			continue
		}

		_, err := finder.AcceleratorByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Global Accelerator Accelerator %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccGlobalAcceleratorAcceleratorConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name = %[1]q
}
`, rName)
}

func testAccGlobalAcceleratorAcceleratorConfigEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = %[2]t
}
`, rName, enabled)
}

func testAccGlobalAcceleratorAcceleratorConfigAttributes(rName string, flowLogsEnabled bool, flowLogsPrefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  acl           = "private"
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false

  attributes {
    flow_logs_enabled   = %[2]t
    flow_logs_s3_bucket = aws_s3_bucket.test.bucket
    flow_logs_s3_prefix = %[3]q
  }
}
`, rName, flowLogsEnabled, flowLogsPrefix)
}

func testAccGlobalAcceleratorAcceleratorConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGlobalAcceleratorAcceleratorConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
