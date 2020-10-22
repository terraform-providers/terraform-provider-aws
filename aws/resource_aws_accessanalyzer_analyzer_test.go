package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// This test can be run via the pattern: TestAccAWSAccessAnalyzer
func testAccAWSAccessAnalyzerAnalyzer_basic(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAccessAnalyzer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessAnalyzerAnalyzerConfigAnalyzerName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "analyzer_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "access-analyzer", fmt.Sprintf("analyzer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", accessanalyzer.TypeAccount),
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

// This test can be run via the pattern: TestAccAWSAccessAnalyzer
func testAccAWSAccessAnalyzerAnalyzer_disappears(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAccessAnalyzer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessAnalyzerAnalyzerConfigAnalyzerName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName, &analyzer),
					testAccCheckAwsAccessAnalyzerAnalyzerDisappears(&analyzer),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// This test can be run via the pattern: TestAccAWSAccessAnalyzer
func testAccAWSAccessAnalyzerAnalyzer_Tags(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAccessAnalyzer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessAnalyzerAnalyzerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName, &analyzer),
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
				Config: testAccAWSAccessAnalyzerAnalyzerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAccessAnalyzerAnalyzerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

// This test can be run via the pattern: TestAccAWSAccessAnalyzer
func testAccAWSAccessAnalyzerAnalyzer_Type_Organization(t *testing.T) {
	var analyzer accessanalyzer.AnalyzerSummary

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSAccessAnalyzer(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessAnalyzerAnalyzerConfigTypeOrganization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "type", accessanalyzer.TypeOrganization),
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

func testAccCheckAccessAnalyzerAnalyzerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).accessanalyzerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_accessanalyzer_analyzer" {
			continue
		}

		input := &accessanalyzer.GetAnalyzerInput{
			AnalyzerName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetAnalyzer(input)

		if isAWSErr(err, accessanalyzer.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("Access Analyzer Analyzer (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsAccessAnalyzerAnalyzerDisappears(analyzer *accessanalyzer.AnalyzerSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).accessanalyzerconn

		input := &accessanalyzer.DeleteAnalyzerInput{
			AnalyzerName: analyzer.Name,
		}

		_, err := conn.DeleteAnalyzer(input)

		return err
	}
}

func testAccCheckAwsAccessAnalyzerAnalyzerExists(resourceName string, analyzer *accessanalyzer.AnalyzerSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).accessanalyzerconn

		input := &accessanalyzer.GetAnalyzerInput{
			AnalyzerName: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetAnalyzer(input)

		if err != nil {
			return err
		}

		*analyzer = *output.Analyzer

		return nil
	}
}

func testAccAWSAccessAnalyzerAnalyzerConfigAnalyzerName(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
}
`, rName)
}

func testAccAWSAccessAnalyzerAnalyzerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAccessAnalyzerAnalyzerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAccessAnalyzerAnalyzerConfigTypeOrganization(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["access-analyzer.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_accessanalyzer_analyzer" "test" {
  depends_on = [aws_organizations_organization.test]

  analyzer_name = %[1]q
  type          = "ORGANIZATION"
}
`, rName)
}
