package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMPatchBaseline_basic(t *testing.T) {
	var before, after ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ssm", regexp.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB123456"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_compliance_level", ssm.PatchComplianceLevelCritical),
					resource.TestCheckResourceAttr(resourceName, "description", "Baseline containing all updates approved for production systems"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_enable_non_security", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMPatchBaselineBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ssm", regexp.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB123456"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB456789"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("updated-patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_compliance_level", ssm.PatchComplianceLevelHigh),
					resource.TestCheckResourceAttr(resourceName, "description", "Baseline containing all updates approved for production systems - August 2017"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					func(*terraform.State) error {
						if aws.StringValue(before.BaselineId) != aws.StringValue(after.BaselineId) {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_tags(t *testing.T) {
	var patch ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfigTags1(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &patch),
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
				Config: testAccAWSSSMPatchBaselineBasicConfigTags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &patch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSSMPatchBaselineBasicConfigTags1(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &patch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_disappears(t *testing.T) {
	var identity ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &identity),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsmPatchBaseline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_OperatingSystem(t *testing.T) {
	var before, after ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineConfigWithOperatingSystem(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", ssm.PatchComplianceLevelCritical),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.enable_non_security", "true"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMPatchBaselineConfigWithOperatingSystemUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", ssm.PatchComplianceLevelInformational),
					resource.TestCheckResourceAttr(resourceName, "operating_system", ssm.OperatingSystemWindows),
					testAccCheckAwsSsmPatchBaselineRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_ApproveUntilDateParam(t *testing.T) {
	var before, after ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineConfigWithApproveUntilDate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_until_date", "2020-01-01"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", ssm.PatchComplianceLevelCritical),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.enable_non_security", "true"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMPatchBaselineConfigWithApproveUntilDateUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_until_date", "2020-02-02"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", ssm.PatchComplianceLevelCritical),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX"),
					func(*terraform.State) error {
						if aws.StringValue(before.BaselineId) != aws.StringValue(after.BaselineId) {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_Sources(t *testing.T) {
	var before, after ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineConfigWithSource(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", "My-AL2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.0.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.0", "AmazonLinux2017.09"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMPatchBaselineConfigWithSourceUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "source.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", "My-AL2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.0.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.0", "AmazonLinux2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.1.name", "My-AL2018.03"),
					resource.TestCheckResourceAttr(resourceName, "source.1.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.1.products.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.1.products.0", "AmazonLinux2018.03"),
					func(*terraform.State) error {
						if aws.StringValue(before.BaselineId) != aws.StringValue(after.BaselineId) {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAWSSSMPatchBaseline_ApprovedPatchesNonSec(t *testing.T) {
	var ssmPatch ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfigApprovedPatchesNonSec(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &ssmPatch),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_enable_non_security", "true"),
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

func TestAccAWSSSMPatchBaseline_RejectPatchesAction(t *testing.T) {
	var ssmPatch ssm.PatchBaselineIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchBaselineBasicConfigRejectPatchesAction(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchBaselineExists(resourceName, &ssmPatch),
					resource.TestCheckResourceAttr(resourceName, "rejected_patches_action", "ALLOW_AS_DEPENDENCY"),
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
resource "aws_ssm_patch_baseline" "test" {
  name                              = "patch-baseline-%s"
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName)
}

func testAccAWSSSMPatchBaselineBasicConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = %[1]q
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSSMPatchBaselineBasicConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = %[1]q
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSSMPatchBaselineBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "updated-patch-baseline-%s"
  description                       = "Baseline containing all updates approved for production systems - August 2017"
  approved_patches                  = ["KB123456", "KB456789"]
  approved_patches_compliance_level = "HIGH"
}
`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithOperatingSystem(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = "patch-baseline-%s"
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_after_days  = 7
    enable_non_security = true
    compliance_level    = "CRITICAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["AmazonLinux2016.03", "AmazonLinux2016.09", "AmazonLinux2017.03", "AmazonLinux2017.09"]
    }

    patch_filter {
      key    = "SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithOperatingSystemUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = "patch-baseline-%s"
  operating_system = "WINDOWS"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_after_days = 7
    compliance_level   = "INFORMATIONAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2012R2"]
    }

    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithApproveUntilDate(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_until_date  = "2020-01-01"
    enable_non_security = true
    compliance_level    = "CRITICAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["AmazonLinux2016.03", "AmazonLinux2016.09", "AmazonLinux2017.03", "AmazonLinux2017.09"]
    }

    patch_filter {
      key    = "SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithApproveUntilDateUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_until_date  = "2020-02-02"
    enable_non_security = true
    compliance_level    = "CRITICAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["AmazonLinux2016.03", "AmazonLinux2016.09", "AmazonLinux2017.03", "AmazonLinux2017.09"]
    }

    patch_filter {
      key    = "SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = %[1]q
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches_compliance_level = "CRITICAL"
  approved_patches                  = ["test123"]
  operating_system                  = "AMAZON_LINUX"

  source {
    name          = "My-AL2017.09"
    configuration = "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"
    products      = ["AmazonLinux2017.09"]
  }
}
`, rName)
}

func testAccAWSSSMPatchBaselineConfigWithSourceUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = %[1]q
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches_compliance_level = "CRITICAL"
  approved_patches                  = ["test123"]
  operating_system                  = "AMAZON_LINUX"

  source {
    name          = "My-AL2017.09"
    configuration = "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"
    products      = ["AmazonLinux2017.09"]
  }

  source {
    name          = "My-AL2018.03"
    configuration = "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"
    products      = ["AmazonLinux2018.03"]
  }
}
`, rName)
}

func testAccAWSSSMPatchBaselineBasicConfigApprovedPatchesNonSec(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                                 = %q
  operating_system                     = "AMAZON_LINUX"
  description                          = "Baseline containing all updates approved for production systems"
  approved_patches                     = ["KB123456"]
  approved_patches_compliance_level    = "CRITICAL"
  approved_patches_enable_non_security = true
}
`, rName)
}

func testAccAWSSSMPatchBaselineBasicConfigRejectPatchesAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "patch-baseline-%s"
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
  rejected_patches_action           = "ALLOW_AS_DEPENDENCY"
}
`, rName)
}
