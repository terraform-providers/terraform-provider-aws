package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/acctest"
)

func TestAccAWSWafGeoMatchSet_basic(t *testing.T) {
	var v waf.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafGeoMatchSetConfig(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &v),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "name", geoMatchSet),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.384465307.type", "Country"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.384465307.value", "US"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.1991628426.type", "Country"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.1991628426.value", "CA"),
				),
			},
		},
	})
}

func TestAccAWSWafGeoMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", acctest.RandString(5))
	geoMatchSetNewName := fmt.Sprintf("geoMatchSetNewName-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafGeoMatchSetConfig(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &before),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "name", geoMatchSet),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.#", "2"),
				),
			},
			{
				Config: testAccAWSWafGeoMatchSetConfigChangeName(geoMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &after),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "name", geoMatchSetNewName),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSWafGeoMatchSet_disappears(t *testing.T) {
	var v waf.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafGeoMatchSetConfig(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &v),
					testAccCheckAWSWafGeoMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafGeoMatchSet_changeConstraints(t *testing.T) {
	var before, after waf.GeoMatchSet
	setName := fmt.Sprintf("geoMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafGeoMatchSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &before),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.384465307.type", "Country"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.384465307.value", "US"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.1991628426.type", "Country"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.1991628426.value", "CA"),
				),
			},
			{
				Config: testAccAWSWafGeoMatchSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &after),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.1174390936.type", "Country"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.1174390936.value", "RU"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.4046309957.type", "Country"),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.4046309957.value", "CN"),
				),
			},
		},
	})
}

func TestAccAWSWafGeoMatchSet_noConstraints(t *testing.T) {
	var ipset waf.GeoMatchSet
	setName := fmt.Sprintf("geoMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafGeoMatchSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &ipset),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_waf_geo_match_set.geo_match_set", "geo_match_constraint.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafGeoMatchSetDisappears(v *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
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
			return fmt.Errorf("Error updating GeoMatchSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteGeoMatchSetInput{
				ChangeToken:   token,
				GeoMatchSetId: v.GeoMatchSetId,
			}
			return conn.DeleteGeoMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error deleting GeoMatchSet: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafGeoMatchSetExists(n string, v *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF GeoMatchSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
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

		return fmt.Errorf("WAF GeoMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafGeoMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_geo_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetGeoMatchSet(
			&waf.GetGeoMatchSetInput{
				GeoMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.GeoMatchSet.GeoMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF GeoMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the GeoMatchSet is already destroyed
		if isAWSErr(err, "WAFNonexistentItemException", "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafGeoMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"
  geo_match_constraint {
    type = "Country"
    value = "US"
  }

  geo_match_constraint {
    type = "Country"
    value = "CA"
  }
}`, name)
}

func testAccAWSWafGeoMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"
  geo_match_constraint {
    type = "Country"
    value = "US"
  }

  geo_match_constraint {
    type = "Country"
    value = "CA"
  }
}`, name)
}

func testAccAWSWafGeoMatchSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"
  geo_match_constraint {
    type = "Country"
    value = "RU"
  }

  geo_match_constraint {
    type = "Country"
    value = "CN"
  }
}`, name)
}

func testAccAWSWafGeoMatchSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"
}`, name)
}
