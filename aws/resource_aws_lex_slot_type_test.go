package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lex_slot_type", &resource.Sweeper{
		Name:         "aws_lex_slot_type",
		F:            testSweepLexSlotTypes,
		Dependencies: []string{"aws_lex_intent"},
	})
}

func testSweepLexSlotTypes(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).lexmodelconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetSlotTypesInput{}

	err = conn.GetSlotTypesPages(input, func(page *lexmodelbuildingservice.GetSlotTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, slotType := range page.SlotTypes {
			r := resourceAwsLexSlotType()
			d := r.Data(nil)

			d.SetId(aws.StringValue(slotType.Name))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Lex Slot Type for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Slot Type for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Lex Slot Type sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAwsLexSlotType_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, "1"),
					resource.TestCheckResourceAttr(rName, "create_version", "false"),
					resource.TestCheckResourceAttr(rName, "description", ""),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "enumeration_value.*", map[string]string{
						"value": "lilies",
					}),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Lirium"),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Martagon"),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttr(rName, "version", LexSlotTypeVersionLatest),
					testAccCheckResourceAttrRfc3339(rName, "created_date"),
					testAccCheckResourceAttrRfc3339(rName, "last_updated_date"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_createVersion(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, "1"),
					resource.TestCheckResourceAttr(rName, "version", LexSlotTypeVersionLatest),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_withVersion(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					testAccCheckAwsLexSlotTypeExistsWithVersion(rName, "1", &v),
					resource.TestCheckResourceAttr(rName, "version", "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_description(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "description", ""),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeUpdateConfig_description(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "description", "Types of flowers to pick up"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_enumerationValues(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_enumerationValues(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "enumeration_value.*", map[string]string{
						"value": "tulips",
					}),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Eduardoregelia"),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Podonix"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_name(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID1 := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testSlotTypeID2 := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID1),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID2),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_valueSelectionStrategy(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_valueSelectionStrategy(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyTopResolution),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_disappears(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLexSlotType(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsLexSlotTypeExistsWithVersion(rName, slotTypeVersion string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex slot type ID is set")
		}

		var err error
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		output, err = conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(rs.Primary.ID),
			Version: aws.String(slotTypeVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return fmt.Errorf("error slot type %q version %s not found", rs.Primary.ID, slotTypeVersion)
		}
		if err != nil {
			return fmt.Errorf("error getting slot type %q version %s: %w", rs.Primary.ID, slotTypeVersion, err)
		}

		return nil
	}
}

func testAccCheckAwsLexSlotTypeExists(rName string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return testAccCheckAwsLexSlotTypeExistsWithVersion(rName, LexSlotTypeVersionLatest, output)
}

func testAccCheckAwsLexSlotTypeNotExists(slotTypeName, slotTypeVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(slotTypeName),
			Version: aws.String(slotTypeVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting slot type %s version %s: %s", slotTypeName, slotTypeVersion, err)
		}

		return fmt.Errorf("error slot type %s version %s exists", slotTypeName, slotTypeVersion)
	}
}

func testAccCheckAwsLexSlotTypeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lex_slot_type" {
			continue
		}

		output, err := conn.GetSlotTypeVersions(&lexmodelbuildingservice.GetSlotTypeVersionsInput{
			Name: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		if output == nil || len(output.SlotTypes) == 0 {
			return nil
		}

		return fmt.Errorf("Lex slot type %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAwsLexSlotTypeConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  name = "%s"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
}
`, rName)
}

func testAccAwsLexSlotTypeConfig_withVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  name = "%s"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
  create_version = true
}
`, rName)
}

func testAccAwsLexSlotTypeUpdateConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  description = "Types of flowers to pick up"
  name        = "%s"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
}
`, rName)
}

func testAccAwsLexSlotTypeConfig_enumerationValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]

    value = "lilies"
  }

  enumeration_value {
    synonyms = [
      "Eduardoregelia",
      "Podonix",
    ]

    value = "tulips"
  }

  name = "%s"
}
`, rName)
}

func testAccAwsLexSlotTypeConfig_valueSelectionStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  name                     = "%s"
  value_selection_strategy = "TOP_RESOLUTION"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
}
`, rName)
}
