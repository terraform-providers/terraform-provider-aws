package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsWafv2RegexPatternSet_Basic(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RegexPatternSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.1641128102.regex_string", "one"),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.265355341.regex_string", "two"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RegexPatternSetConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.1641128102.regex_string", "one"),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.265355341.regex_string", "two"),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.2339296779.regex_string", "three"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSWafv2RegexPatternSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RegexPatternSet_Disappears(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RegexPatternSetConfig_Minimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2RegexPatternSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsWafv2RegexPatternSet_Minimal(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RegexPatternSetConfig_Minimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "0"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RegexPatternSet_ChangeNameForceNew(t *testing.T) {
	var before, after wafv2.RegexPatternSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNewName := fmt.Sprintf("regex-pattern-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RegexPatternSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "2"),
				),
			},
			{
				Config: testAccAwsWafv2RegexPatternSetConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rNewName),
					resource.TestCheckResourceAttr(resourceName, "description", rNewName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "2"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RegexPatternSet_Tags(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2RegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RegexPatternSetConfig_OneTag(rName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSWafv2RegexPatternSetImportStateIdFunc(resourceName),
			},
			{
				Config: testAccAwsWafv2RegexPatternSetConfig_TwoTags(rName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2RegexPatternSetConfig_OneTag(rName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2RegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func testAccCheckAWSWafv2RegexPatternSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_regex_pattern_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRegexPatternSet(
			&wafv2.GetRegexPatternSetInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if resp != nil && resp.RegexPatternSet != nil && aws.StringValue(resp.RegexPatternSet.Id) == rs.Primary.ID {
				return fmt.Errorf("WAFv2 RegexPatternSet %s still exists", rs.Primary.ID)
			}
			return nil
		}

		// Return nil if the RegexPatternSet is already destroyed
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafv2RegexPatternSetExists(n string, v *wafv2.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 RegexPatternSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRegexPatternSet(&wafv2.GetRegexPatternSetInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.RegexPatternSet == nil {
			return fmt.Errorf("Error getting WAFv2 RegexPatternSet for %s", rs.Primary.ID)
		}

		if aws.StringValue(resp.RegexPatternSet.Id) == rs.Primary.ID {
			*v = *resp.RegexPatternSet
			return nil
		}

		return fmt.Errorf("WAFv2 RegexPatternSet (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWafv2RegexPatternSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

	regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }
}
`, name, name)
}

func testAccAwsWafv2RegexPatternSetConfig_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "Updated"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  regular_expression {
    regex_string = "three"
  }
}
`, name)
}

func testAccAwsWafv2RegexPatternSetConfig_Minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "%s"
  scope = "REGIONAL"
}
`, name)
}

func testAccAwsWafv2RegexPatternSetConfig_OneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  tags = {
    "%s" = "%s"
  }
}
`, name, name, tagKey, tagValue)
}

func testAccAwsWafv2RegexPatternSetConfig_TwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAWSWafv2RegexPatternSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
