package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_classifier", &resource.Sweeper{
		Name: "aws_glue_classifier",
		F:    testSweepGlueClassifiers,
	})
}

func testSweepGlueClassifiers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	prefixes := []string{
		"tf-acc-test-",
	}

	input := &glue.GetClassifiersInput{}
	err = conn.GetClassifiersPages(input, func(page *glue.GetClassifiersOutput, lastPage bool) bool {
		if len(page.Classifiers) == 0 {
			log.Printf("[INFO] No Glue Classifiers to sweep")
			return false
		}
		for _, classifier := range page.Classifiers {
			skip := true

			var name string
			if classifier.GrokClassifier != nil {
				name = aws.StringValue(classifier.GrokClassifier.Name)
			} else if classifier.JsonClassifier != nil {
				name = aws.StringValue(classifier.JsonClassifier.Name)
			} else if classifier.XMLClassifier != nil {
				name = aws.StringValue(classifier.XMLClassifier.Name)
			}
			if name == "" {
				log.Printf("[WARN] Unable to determine Glue Classifier name: %#v", classifier)
				continue
			}

			for _, prefix := range prefixes {
				if strings.HasPrefix(name, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping Glue Classifier: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Glue Classifier: %s", name)
			err := deleteGlueClassifier(conn, name)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Classifier %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Classifier sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Glue Classifiers: %s", err)
	}

	return nil
}

func TestAccAWSGlueClassifier_GrokClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueClassifierConfig_GrokClassifier(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_GrokClassifier(rName, "classification2", "pattern2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern2"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccAWSGlueClassifier_GrokClassifier_CustomPatterns(t *testing.T) {
	var classifier glue.Classifier

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueClassifierConfig_GrokClassifier_CustomPatterns(rName, "custompattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", "custompattern1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_GrokClassifier_CustomPatterns(rName, "custompattern2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", "custompattern2"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccAWSGlueClassifier_JsonClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueClassifierConfig_JsonClassifier(rName, "jsonpath1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_JsonClassifier(rName, "jsonpath2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
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

func TestAccAWSGlueClassifier_TypeChange(t *testing.T) {
	var classifier glue.Classifier

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueClassifierConfig_GrokClassifier(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_JsonClassifier(rName, "jsonpath1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.0.json_path", "jsonpath1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_XmlClassifier(rName, "classification1", "rowtag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag1"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_GrokClassifier(rName, "classification1", "pattern1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.custom_patterns", ""),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.0.grok_pattern", "pattern1"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSGlueClassifier_XmlClassifier(t *testing.T) {
	var classifier glue.Classifier

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_classifier.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueClassifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueClassifierConfig_XmlClassifier(rName, "classification1", "rowtag1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag1"),
				),
			},
			{
				Config: testAccAWSGlueClassifierConfig_XmlClassifier(rName, "classification2", "rowtag2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueClassifierExists(resourceName, &classifier),
					resource.TestCheckResourceAttr(resourceName, "grok_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "json_classifier.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.classification", "classification2"),
					resource.TestCheckResourceAttr(resourceName, "xml_classifier.0.row_tag", "rowtag2"),
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

func testAccCheckAWSGlueClassifierExists(resourceName string, classifier *glue.Classifier) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Classifier ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetClassifier(&glue.GetClassifierInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Classifier == nil {
			return fmt.Errorf("Glue Classifier (%s) not found", rs.Primary.ID)
		}

		*classifier = *output.Classifier
		return nil
	}
}

func testAccCheckAWSGlueClassifierDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_classifier" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetClassifier(&glue.GetClassifierInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}

		}

		classifier := output.Classifier
		if classifier != nil {
			return fmt.Errorf("Glue Classifier %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccAWSGlueClassifierConfig_GrokClassifier(rName, classification, grokPattern string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  grok_classifier {
    classification = "%s"
    grok_pattern   = "%s"
  }
}
`, rName, classification, grokPattern)
}

func testAccAWSGlueClassifierConfig_GrokClassifier_CustomPatterns(rName, customPatterns string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  grok_classifier {
    classification  = "classification"
    custom_patterns = "%s"
    grok_pattern    = "pattern"
  }
}
`, rName, customPatterns)
}

func testAccAWSGlueClassifierConfig_JsonClassifier(rName, jsonPath string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  json_classifier {
    json_path = "%s"
  }
}
`, rName, jsonPath)
}

func testAccAWSGlueClassifierConfig_XmlClassifier(rName, classification, rowTag string) string {
	return fmt.Sprintf(`
resource "aws_glue_classifier" "test" {
  name = "%s"

  xml_classifier {
    classification  = "%s"
    row_tag         = "%s"
  }
}
`, rName, classification, rowTag)
}
