package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSWafRegionalGeoMatchSet_basic(t *testing.T) {
	var v waf.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	geoMatchSet := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfig(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", geoMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "US",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "CA",
					}),
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

func TestAccAWSWafRegionalGeoMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	geoMatchSet := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	geoMatchSetNewName := fmt.Sprintf("geoMatchSetNewName-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfig(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, "name", geoMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
				),
			},
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfigChangeName(geoMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &after),
					testAccCheckAWSWafGeoMatchSetIdDiffers(&before, &after),
					resource.TestCheckResourceAttr(
						resourceName, "name", geoMatchSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
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

func TestAccAWSWafRegionalGeoMatchSet_disappears(t *testing.T) {
	var v waf.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	geoMatchSet := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfig(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &v),
					testAccCheckAWSWafRegionalGeoMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalGeoMatchSet_changeConstraints(t *testing.T) {
	var before, after waf.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	setName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, "name", setName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "US",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "CA",
					}),
				),
			},
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, "name", setName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "RU",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "CN",
					}),
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

func TestAccAWSWafRegionalGeoMatchSet_noConstraints(t *testing.T) {
	var ipset waf.GeoMatchSet
	resourceName := "aws_wafregional_geo_match_set.test"
	setName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalGeoMatchSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalGeoMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(
						resourceName, "name", setName),
					resource.TestCheckResourceAttr(
						resourceName, "geo_match_constraint.#", "0"),
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

func testAccCheckAWSWafGeoMatchSetIdDiffers(before, after *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.GeoMatchSetId == *after.GeoMatchSetId {
			return fmt.Errorf("Expected different IDs, given %q for both sets", *before.GeoMatchSetId)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalGeoMatchSetDisappears(v *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateGeoMatchSetInput{
				ChangeToken:   token,
				GeoMatchSetId: v.GeoMatchSetId,
			}

			for _, geoMatchConstraint := range v.GeoMatchConstraints {
				geoMatchConstraintUpdate := &waf.GeoMatchSetUpdate{
					Action: aws.String("DELETE"),
					GeoMatchConstraint: &waf.GeoMatchConstraint{
						Type:  geoMatchConstraint.Type,
						Value: geoMatchConstraint.Value,
					},
				}
				req.Updates = append(req.Updates, geoMatchConstraintUpdate)
			}
			return conn.UpdateGeoMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regional Geo Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteGeoMatchSetInput{
				ChangeToken:   token,
				GeoMatchSetId: v.GeoMatchSetId,
			}
			return conn.DeleteGeoMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting WAF Regional Geo Match Set: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalGeoMatchSetExists(n string, v *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regional Geo Match Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetGeoMatchSet(&waf.GetGeoMatchSetInput{
			GeoMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.GeoMatchSet.GeoMatchSetId == rs.Primary.ID {
			*v = *resp.GeoMatchSet
			return nil
		}

		return fmt.Errorf("WAF Regional Geo Match Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegionalGeoMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_geo_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetGeoMatchSet(
			&waf.GetGeoMatchSetInput{
				GeoMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.GeoMatchSet.GeoMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF Regional Geo Match Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the WAF Regional Geo Match Set is already destroyed
		if isAWSErr(err, "WAFNonexistentItemException", "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafRegionalGeoMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_geo_match_set" "test" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
`, name)
}

func testAccAWSWafRegionalGeoMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_geo_match_set" "test" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
`, name)
}

func testAccAWSWafRegionalGeoMatchSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_geo_match_set" "test" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "RU"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CN"
  }
}
`, name)
}

func testAccAWSWafRegionalGeoMatchSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_geo_match_set" "test" {
  name = "%s"
}
`, name)
}
