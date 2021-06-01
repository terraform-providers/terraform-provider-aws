package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/waf/lister"
)

func init() {
	resource.AddTestSweepers("aws_waf_rule", &resource.Sweeper{
		Name: "aws_waf_rule",
		F:    testSweepWafRules,
		Dependencies: []string{
			"aws_waf_rule_group",
			"aws_waf_web_acl",
		},
	})
}

func testSweepWafRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListRulesInput{}

	err = lister.ListRulesPages(conn, input, func(page *waf.ListRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, rule := range page.Rules {
			id := aws.StringValue(rule.RuleId)

			r := resourceAwsWafRule()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in predicates attribute
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF Rule (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			// In case it was already deleted
			if d.Id() == "" {
				continue
			}

			err = r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAF Rule (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Rule sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Rules: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSWafRule_basic(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`rule/.+`)),
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

func TestAccAWSWafRule_changeNameForceNew(t *testing.T) {
	var before, after waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleNewName),
				),
			},
		},
	})
}

func TestAccAWSWafRule_disappears(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &v),
					testAccCheckAWSWafRuleDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var byteMatchSet waf.ByteMatchSet

	var before, after waf.Rule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists("aws_waf_ipset.ipset", &ipset),
					testAccCheckAWSWafRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					computeWafRulePredicateWithIpSet(&ipset, false, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccAWSWafRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists("aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckAWSWafRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					computeWafRulePredicateWithByteMatchSet(&byteMatchSet, true, "ByteMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "true",
						"type":    "ByteMatch",
					}),
				),
			},
		},
	})
}

func TestAccAWSWafRule_geoMatchSetPredicate(t *testing.T) {
	var geoMatchSet waf.GeoMatchSet

	var v waf.Rule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfig_geoMatchSetPredicate(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &geoMatchSet),
					testAccCheckAWSWafRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					computeWafRulePredicateWithGeoMatchSet(&geoMatchSet, true, "GeoMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "true",
						"type":    "GeoMatch",
					}),
				),
			},
		},
	})
}

// computeWafRulePredicateWithIpSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRulePredicateWithIpSet(ipSet *waf.IPSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRule().Schema["predicates"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			"data_id": *ipSet.IPSetId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

// computeWafRulePredicateWithByteMatchSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRulePredicateWithByteMatchSet(set *waf.ByteMatchSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRule().Schema["predicates"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			"data_id": *set.ByteMatchSetId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

// computeWafRulePredicateWithGeoMatchSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRulePredicateWithGeoMatchSet(set *waf.GeoMatchSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRule().Schema["predicates"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			"data_id": *set.GeoMatchSetId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

func TestAccAWSWafRule_noPredicates(t *testing.T) {
	var rule waf.Rule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSWafRule_Tags(t *testing.T) {
	var rule waf.Rule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleConfigTags1(ruleName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSWafRuleConfigTags2(ruleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSWafRuleConfigTags1(ruleName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccCheckAWSWafRuleDisappears(v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}

			for _, Predicate := range v.Predicates {
				Predicate := &waf.RuleUpdate{
					Action: aws.String("DELETE"),
					Predicate: &waf.Predicate{
						Negated: Predicate.Negated,
						Type:    Predicate.Type,
						DataId:  Predicate.DataId,
					},
				}
				req.Updates = append(req.Updates, Predicate)
			}

			return conn.UpdateRule(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}
			return conn.DeleteRule(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF Rule: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_rule" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRule(
			&waf.GetRuleInput{
				RuleId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.Rule.RuleId == rs.Primary.ID {
				return fmt.Errorf("WAF Rule %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Rule is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == waf.ErrCodeNonexistentItemException {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafRuleExists(n string, v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRule(&waf.GetRuleInput{
			RuleId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Rule.RuleId == rs.Primary.ID {
			*v = *resp.Rule
			return nil
		}

		return fmt.Errorf("WAF Rule (%s) not found", rs.Primary.ID)
	}
}

func testAccPreCheckAWSWaf(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).wafconn

	input := &waf.ListRulesInput{}

	_, err := conn.ListRules(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSWafRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, name, name)
}

func testAccAWSWafRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, name, name)
}

func testAccAWSWafRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_byte_match_set" "set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}

resource "aws_waf_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"

  predicates {
    data_id = aws_waf_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name, name, name, name)
}

func testAccAWSWafRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
}
`, name, name)
}

func testAccAWSWafRuleConfig_geoMatchSetPredicate(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }
}

resource "aws_waf_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"

  predicates {
    data_id = aws_waf_geo_match_set.geo_match_set.id
    negated = true
    type    = "GeoMatch"
  }
}
`, name, name, name)
}

func testAccAWSWafRuleConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %q = %q
  }
}
`, rName, rName, rName, tag1Key, tag1Value)
}

func testAccAWSWafRuleConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, rName, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
