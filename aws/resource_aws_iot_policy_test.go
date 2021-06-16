package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iot_policy", &resource.Sweeper{
		Name: "aws_iot_policy",
		F:    testSweepIotPolicies,
		Dependencies: []string{
			"aws_iot_policy_attachment",
		},
	})
}

func testSweepIotPolicies(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).iotconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListPoliciesInput{}

	err = conn.ListPoliciesPages(input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.Policies {
			r := resourceAwsIotPolicy()
			d := r.Data(nil)

			d.SetId(aws.StringValue(policy.PolicyName))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy for %s: %w", region, err))
	}

	if err := testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Policy for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Policy sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIoTPolicy_basic(t *testing.T) {
	var v iot.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTPolicyConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "default_version_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccAWSIoTPolicy_disappears(t *testing.T) {
	var v iot.GetPolicyOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iot.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTPolicyConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTPolicyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIotPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSIoTPolicyDestroy_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_policy" {
			continue
		}

		// Try to find the Policy
		GetPolicyOpts := &iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.Attributes["name"]),
		}

		resp, err := conn.GetPolicy(GetPolicyOpts)

		if err == nil {
			if resp.PolicyName != nil {
				return fmt.Errorf("IoT Policy still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}

	return nil
}

func testAccCheckAWSIoTPolicyExists(n string, v *iot.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn

		resp, err := conn.GetPolicy(&iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccAWSIoTPolicyConfigInitialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
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
