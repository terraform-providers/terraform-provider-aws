package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMDocument_basic(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_format", "JSON"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ssm", fmt.Sprintf("document/%s", name)),
					testAccCheckResourceAttrRfc3339(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "document_version"),
					resource.TestCheckResourceAttr(resourceName, "version_name", ""),
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

func TestAccAWSSSMDocument_Name(t *testing.T) {
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentBasicConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocumentBasicConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_target_type(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentBasicConfigTargetType(name, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocumentBasicConfigTargetType(name, "/AWS::EC2::Instance"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "/AWS::EC2::Instance"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_VersionName(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentBasicConfigVersionName(name, "release-1.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "version_name", "release-1.0.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocumentBasicConfigVersionName(name, "release-1.0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "version_name", "release-1.0.1"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_update(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocument20Config(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema_version", "2.0"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocument20UpdatedConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_version", "2"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_permission_public(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentPublicPermissionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", "all"),
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

func TestAccAWSSSMDocument_permission_private(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	ids := "123456789012"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentPrivatePermissionConfig(name, ids),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
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

func TestAccAWSSSMDocument_permission_batching(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	ids := "123456789012,123456789013,123456789014,123456789015,123456789016,123456789017,123456789018,123456789019,123456789020,123456789021,123456789022,123456789023,123456789024,123456789025,123456789026,123456789027,123456789028,123456789029,123456789030,123456789031,123456789032"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentPrivatePermissionConfig(name, ids),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
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

func TestAccAWSSSMDocument_permission_change(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	idsInitial := "123456789012,123456789013"
	idsRemove := "123456789012"
	idsAdd := "123456789012,123456789014"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentPrivatePermissionConfig(name, idsInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", idsInitial),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocumentPrivatePermissionConfig(name, idsRemove),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", idsRemove),
				),
			},
			{
				Config: testAccAWSSSMDocumentPrivatePermissionConfig(name, idsAdd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", idsAdd),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_params(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentParamConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.0.name", "commands"),
					resource.TestCheckResourceAttr(resourceName, "parameter.0.type", "StringList"),
					resource.TestCheckResourceAttr(resourceName, "parameter.1.name", "workingDirectory"),
					resource.TestCheckResourceAttr(resourceName, "parameter.1.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2.name", "executionTimeout"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2.type", "String"),
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

func TestAccAWSSSMDocument_automation(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentTypeAutomationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Automation"),
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

func TestAccAWSSSMDocument_package(t *testing.T) {
	name := acctest.RandString(10)
	rInt := acctest.RandInt()
	rInt2 := acctest.RandInt()
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentTypePackageConfig(name, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Package"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"attachments_source"}, // This doesn't work because the API doesn't provide attachments info directly
			},
			{
				Config: testAccAWSSSMDocumentTypePackageConfig(name, rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Package"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_SchemaVersion_1(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentConfigSchemaVersion1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema_version", "1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocumentConfigSchemaVersion1Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema_version", "1.0"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_session(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentTypeSessionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Session"),
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

func TestAccAWSSSMDocument_DocumentFormat_YAML(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	content1 := `
---
schemaVersion: '2.2'
description: Sample document
mainSteps:
- action: aws:runPowerShellScript
  name: runPowerShellScript
  inputs:
    runCommand:
      - hostname
`
	content2 := `
---
schemaVersion: '2.2'
description: Sample document
mainSteps:
- action: aws:runPowerShellScript
  name: runPowerShellScript
  inputs:
    runCommand:
      - Get-Process
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentConfig_DocumentFormat_YAML(name, content1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content", content1+"\n"),
					resource.TestCheckResourceAttr(resourceName, "document_format", "YAML"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMDocumentConfig_DocumentFormat_YAML(name, content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content", content2+"\n"),
					resource.TestCheckResourceAttr(resourceName, "document_format", "YAML"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_Tags(t *testing.T) {
	rName := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentConfig_Tags_Single(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
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
				Config: testAccAWSSSMDocumentConfig_Tags_Multiple(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSSMDocumentConfig_Tags_Single(rName, "key2", "value2updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2updated"),
				),
			},
		},
	})
}

func TestAccAWSSSMDocument_disappears(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_document.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMDocumentBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMDocumentExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsmDocument(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestValidateSSMDocumentPermissions(t *testing.T) {
	validValues := []map[string]interface{}{
		{
			"type":        "Share",
			"account_ids": "123456789012,123456789014",
		},
		{
			"type":        "Share",
			"account_ids": "all",
		},
	}

	for _, s := range validValues {
		errors := validateSSMDocumentPermissions(s)
		if len(errors) > 0 {
			t.Fatalf("%q should be valid SSM Document Permissions: %v", s, errors)
		}
	}

	invalidValues := []map[string]interface{}{
		{},
		{"type": ""},
		{"type": "Share"},
		{"account_ids": ""},
		{"account_ids": "all"},
		{"type": "", "account_ids": ""},
		{"type": "", "account_ids": "all"},
		{"type": "share", "account_ids": ""},
		{"type": "share", "account_ids": "all"},
		{"type": "private", "account_ids": "all"},
	}

	for _, s := range invalidValues {
		errors := validateSSMDocumentPermissions(s)
		if len(errors) == 0 {
			t.Fatalf("%q should not be valid SSM Document Permissions: %v", s, errors)
		}
	}
}

func testAccCheckAWSSSMDocumentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Document ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		_, err := conn.DescribeDocument(&ssm.DescribeDocumentInput{
			Name: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSSSMDocumentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_document" {
			continue
		}

		out, err := conn.DescribeDocument(&ssm.DescribeDocumentInput{
			Name: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			// InvalidDocument means it's gone, this is good
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == ssm.ErrCodeInvalidDocument {
				return nil
			}
			return err
		}

		if out != nil {
			return fmt.Errorf("Expected AWS SSM Document to be gone, but was still found")
		}

		return nil
	}

	return fmt.Errorf("Default error in SSM Document Test")
}

/*
Based on examples from here: https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/create-ssm-doc.html
*/

func testAccAWSSSMDocumentBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentBasicConfigTargetType(rName, typ string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "%s"
  document_type = "Command"
  target_type   = "%s"

  content = <<DOC
{
  "schemaVersion": "2.0",
  "description": "Sample version 2.0 document v2",
  "parameters": {},
  "mainSteps": [
    {
      "action": "aws:runPowerShellScript",
      "name": "runPowerShellScript",
      "inputs": {
        "runCommand": [
          "Get-Process"
        ]
      }
    }
  ]
}
DOC

}
`, rName, typ)
}

func testAccAWSSSMDocumentBasicConfigVersionName(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"
  version_name  = %[2]q

  content = <<DOC
    {
       "schemaVersion": "2.0",
       "description": "Sample version 2.0 document %[2]s",
       "parameters": {

       },
       "mainSteps": [
          {
             "action": "aws:runPowerShellScript",
             "name": "runPowerShellScript",
             "inputs": {
                "runCommand": [
                   "Get-Process"
                ]
             }
          }
       ]
    }
DOC
}
`, rName, version)
}

func testAccAWSSSMDocument20Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "2.0",
  "description": "Sample version 2.0 document v2",
  "parameters": {},
  "mainSteps": [
    {
      "action": "aws:runPowerShellScript",
      "name": "runPowerShellScript",
      "inputs": {
        "runCommand": [
          "Get-Process"
        ]
      }
    }
  ]
}
DOC

}
`, rName)
}

func testAccAWSSSMDocument20UpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "2.0",
  "description": "Sample version 2.0 document v2",
  "parameters": {},
  "mainSteps": [
    {
      "action": "aws:runPowerShellScript",
      "name": "runPowerShellScript",
      "inputs": {
        "runCommand": [
          "Get-Process -Verbose"
        ]
      }
    }
  ]
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentPublicPermissionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document-%s"
  document_type = "Command"

  permissions = {
    type        = "Share"
    account_ids = "all"
  }

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentPrivatePermissionConfig(rName string, rIds string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document-%s"
  document_type = "Command"

  permissions = {
    type        = "Share"
    account_ids = "%s"
  }

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}
`, rName, rIds)
}

func testAccAWSSSMDocumentParamConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Run a PowerShell script or specify the paths to scripts to run.",
  "parameters": {
    "commands": {
      "type": "StringList",
      "description": "(Required) Specify the commands to run or the paths to existing scripts on the instance.",
      "minItems": 1,
      "displayType": "textarea"
    },
    "workingDirectory": {
      "type": "String",
      "default": "",
      "description": "(Optional) The path to the working directory on your instance.",
      "maxChars": 4096
    },
    "executionTimeout": {
      "type": "String",
      "default": "3600",
      "description": "(Optional) The time in seconds for a command to be completed before it is considered to have failed. Default is 3600 (1 hour). Maximum is 28800 (8 hours).",
      "allowedPattern": "([1-9][0-9]{0,3})|(1[0-9]{1,4})|(2[0-7][0-9]{1,3})|(28[0-7][0-9]{1,2})|(28800)"
    }
  },
  "runtimeConfig": {
    "aws:runPowerShellScript": {
      "properties": [
        {
          "id": "0.aws:runPowerShellScript",
          "runCommand": "{{ commands }}",
          "workingDirectory": "{{ workingDirectory }}",
          "timeoutSeconds": "{{ executionTimeout }}"
        }
      ]
    }
  }
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentTypeAutomationConfig(rName string) string {
	return composeConfig(testAccLatestAmazonLinuxHvmEbsAmiConfig(), fmt.Sprintf(`
resource "aws_iam_instance_profile" "ssm_profile" {
  name = "ssm_profile-%[1]s"
  role = aws_iam_role.ssm_role.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "ssm_role" {
  name = "ssm_role-%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_ssm_document" "test" {
  name          = "test_document-%[1]s"
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.ssm_role.arn}",
  "mainSteps": [
    {
      "name": "startInstances",
      "action": "aws:runInstances",
      "timeoutSeconds": 1200,
      "maxAttempts": 1,
      "onFailure": "Abort",
      "inputs": {
        "ImageId": "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.ssm_profile.name}"
      }
    },
    {
      "name": "stopInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "stopped"
      }
    },
    {
      "name": "terminateInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "terminated"
      }
    }
  ]
}
DOC

}
`, rName))
}

func testAccAWSSSMDocumentTypePackageConfig(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "ssm_profile-%[1]s"
  role = aws_iam_role.test.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "ssm_role-%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-object-test-bucket-%[2]d"
}

resource "aws_s3_bucket_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "test.zip"
  source       = "test-fixtures/ssm-doc-acc-test.zip"
  content_type = "binary/octet-stream"
}

resource "aws_ssm_document" "test" {
  name          = "test_document-%[1]s"
  document_type = "Package"

  attachments_source {
    key    = "SourceUrl"
    values = ["s3://${aws_s3_bucket_object.test.bucket}"]
  }

  content = <<DOC
{
  "description": "Systems Manager Package Document Test",
  "schemaVersion": "2.0",
  "version": "0.1",
  "assumeRole": "${aws_iam_role.test.arn}",
  "files": {
    "test.zip": {
      "checksums": {
        "sha256": "${filesha256("test-fixtures/ssm-doc-acc-test.zip")}"
      }
    }
  },
  "packages": {
    "amazon": {
      "_any": {
        "x86_64": {
          "file": "${aws_s3_bucket_object.test.key}"
        }
      }
    }
  }
}
DOC

  depends_on = [aws_iam_role_policy.test]
}
`, rName, rInt)
}

func testAccAWSSSMDocumentTypeSessionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document-%s"
  document_type = "Session"

  content = <<DOC
{
  "schemaVersion": "1.0",
  "description": "Document to hold regional settings for Session Manager",
  "sessionType": "Standard_Stream",
  "inputs": {
    "s3BucketName": "test",
    "s3KeyPrefix": "test",
    "s3EncryptionEnabled": true,
    "cloudWatchLogGroupName": "/logs/sessions",
    "cloudWatchEncryptionEnabled": false
  }
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentConfig_DocumentFormat_YAML(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  document_format = "YAML"
  document_type   = "Command"
  name            = "test_document-%s"

  content = <<DOC
%s
DOC

}
`, rName, content)
}

func testAccAWSSSMDocumentConfigSchemaVersion1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %q
  document_type = "Session"

  content = <<DOC
{
  "schemaVersion": "1.0",
  "description": "Document to hold regional settings for Session Manager",
  "sessionType": "Standard_Stream",
  "inputs": {
    "s3BucketName": "test",
    "s3KeyPrefix": "test",
    "s3EncryptionEnabled": true,
    "cloudWatchLogGroupName": "/logs/sessions",
    "cloudWatchEncryptionEnabled": false
  }
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentConfigSchemaVersion1Update(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %q
  document_type = "Session"

  content = <<DOC
{
  "schemaVersion": "1.0",
  "description": "Document to hold regional settings for Session Manager",
  "sessionType": "Standard_Stream",
  "inputs": {
    "s3BucketName": "test",
    "s3KeyPrefix": "test",
    "s3EncryptionEnabled": true,
    "cloudWatchLogGroupName": "/logs/sessions-updated",
    "cloudWatchEncryptionEnabled": false
  }
}
DOC

}
`, rName)
}

func testAccAWSSSMDocumentConfig_Tags_Single(rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  document_type = "Command"
  name          = "test_document-%s"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

  tags = {
    %s = %q
  }
}
`, rName, key1, value1)
}

func testAccAWSSSMDocumentConfig_Tags_Multiple(rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  document_type = "Command"
  name          = "test_document-%s"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

  tags = {
    %s = %q
    %s = %q
  }
}
`, rName, key1, value1, key2, value2)
}
