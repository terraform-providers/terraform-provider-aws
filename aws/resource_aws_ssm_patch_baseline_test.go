package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMPatchBaseline_basic(t *testing.T) {
	var before, after ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists("aws_ssm_patch_baseline.foo", &before),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches.2062620480", "KB123456"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "name", fmt.Sprintf("patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches_compliance_level", ssm.PatchComplianceLevelCritical),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "description", "Baseline containing all updates approved for production systems"),
				),
			},
			{
				Config: testAccAWSSSMPatchBaselineBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists("aws_ssm_patch_baseline.foo", &after),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches.2062620480", "KB123456"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches.2291496788", "KB456789"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "name", fmt.Sprintf("updated-patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approved_patches_compliance_level", ssm.PatchComplianceLevelHigh),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "description", "Baseline containing all updates approved for production systems - August 2017"),
					func(*terraform.State) error {
						if *before.BaselineId != *after.BaselineId {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_disappears(t *testing.T) {
	var identity ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &identity),
					testAccCheckAWSSSMPatchBaselineDisappears(&identity),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMPatchBaselineWithOperatingSystem(t *testing.T) {
	var before, after ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineConfigWithOperatingSystem(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists("aws_ssm_patch_baseline.foo", &before),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.patch_filter.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.compliance_level", ssm.PatchComplianceLevelCritical),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.enable_non_security", "true"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "operating_system", "AMAZON_LINUX"),
				),
			},
			{
				Config: testAccAWSSSMPatchBaselineConfigWithOperatingSystemUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists("aws_ssm_patch_baseline.foo", &after),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.patch_filter.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "approval_rule.0.compliance_level", ssm.PatchComplianceLevelInformational),
					resource.TestCheckResourceAttr(
						"aws_ssm_patch_baseline.foo", "operating_system", ssm.OperatingSystemWindows),
					testAccCheckAwsSsmPatchBaselineRecreated(t, &before, &after),
				),
			},
		},
	})
}

func testAccCheckAwsSsmPatchBaselineRecreated(t *testing.T,
	before, after *ssm.PatchBaselineIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.BaselineId == *after.BaselineId {
			t.Fatalf("Expected change of SSM Patch Baseline IDs, but both were %v", *before.BaselineId)
		}
		return nil
	}
}

func testAccCheckAWSSSMPatchBaselineExists(n string, patch *ssm.PatchBaselineIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Patch Baseline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribePatchBaselines(&ssm.DescribePatchBaselinesInput{
			Filters: []*ssm.PatchOrchestratorFilter{
				{
					Key:    aws.String("NAME_PREFIX"),
					Values: []*string{aws.String(rs.Primary.Attributes["name"])},
				},
			},
		})

		for _, i := range resp.BaselineIdentities {
			if *i.BaselineId == rs.Primary.ID {
				*patch = *i
				return nil
			}
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("No AWS SSM Patch Baseline found")
	}
}

func testAccCheckAWSSSMPatchBaselineDisappears(patch *ssm.PatchBaselineIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		id := aws.StringValue(patch.BaselineId)
		params := &ssm.DeletePatchBaselineInput{
			BaselineId: aws.String(id),
		}

		_, err := conn.DeletePatchBaseline(params)
		if err != nil {
			return fmt.Errorf("error deleting Patch Baseline %s: %s", id, err)
		}

		return nil
	}
}

func testAccCheckAWSSSMPatchBaselineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_patch_baseline" {
			continue
		}

		out, err := conn.DescribePatchBaselines(&ssm.DescribePatchBaselinesInput{
			Filters: []*ssm.PatchOrchestratorFilter{
				{
					Key:    aws.String("NAME_PREFIX"),
					Values: []*string{aws.String(rs.Primary.Attributes["name"])},
				},
			},
		})

		if err != nil {
			return err
		}

		if len(out.BaselineIdentities) > 0 {
			return fmt.Errorf("Expected AWS SSM Patch Baseline to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSSSMPatchBaselineBasicConfig(rName string) string {
	return fmt.Sprintf(`

resource "aws_ssm_patch_baseline" "foo" {
  name  = "patch-baseline-%s"
  description = "Baseline containing all updates approved for production systems"
  approved_patches = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}

`, rName)
}

func testAccAWSSSMPatchBaselineBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`

resource "aws_ssm_patch_baseline" "foo" {
  name  = "updated-patch-baseline-%s"
  description = "Baseline containing all updates approved for production systems - August 2017"
  approved_patches = ["KB123456","KB456789"]
  approved_patches_compliance_level = "HIGH"
}

`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithOperatingSystem(rName string) string {
	return fmt.Sprintf(`

resource "aws_ssm_patch_baseline" "foo" {
  name  = "patch-baseline-%s"
  operating_system = "AMAZON_LINUX"
  description = "Baseline containing all updates approved for production systems"
  approval_rule {
  	approve_after_days = 7
	enable_non_security = true
  	compliance_level = "CRITICAL"

  	patch_filter {
		key = "PRODUCT"
		values = ["AmazonLinux2016.03","AmazonLinux2016.09","AmazonLinux2017.03","AmazonLinux2017.09"]
  	}

  	patch_filter {
		key = "SEVERITY"
		values = ["Critical","Important"]
  	}
  }
}

`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithOperatingSystemUpdated(rName string) string {
	return fmt.Sprintf(`

resource "aws_ssm_patch_baseline" "foo" {
  name  = "patch-baseline-%s"
  operating_system = "WINDOWS"
  description = "Baseline containing all updates approved for production systems"
  approval_rule {
  	approve_after_days = 7
  	compliance_level = "INFORMATIONAL"

  	patch_filter {
		key = "PRODUCT"
		values = ["WindowsServer2012R2"]
  	}

  	patch_filter {
		key = "MSRC_SEVERITY"
		values = ["Critical","Important"]
  	}
  }
}

`, rName)
}
