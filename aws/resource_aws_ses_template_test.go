package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSesTemplate_basic(t *testing.T) {
	resourceName := "aws_ses_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var template ses.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
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

func TestAccAWSSesTemplate_Update(t *testing.T) {
	TestAccSkip(t, "Skip due to SES.UpdateTemplate eventual consistency issues")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_template.test"
	var template ses.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("template/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", "text"),
				),
			},
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html update"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
		},
	})
}

func TestAccAWSSesTemplate_disappears(t *testing.T) {
	resourceName := "aws_ses_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var template ses.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplateExists(resourceName, &template),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSesTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSesTemplateExists(pr string, template *ses.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sesconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		input := ses.GetTemplateInput{
			TemplateName: aws.String(rs.Primary.ID),
		}

		templateOutput, err := conn.GetTemplate(&input)
		if err != nil {
			return err
		}

		if templateOutput == nil || templateOutput.Template == nil {
			return fmt.Errorf("SES Template (%s) not found", rs.Primary.ID)
		}

		*template = *templateOutput.Template

		return nil
	}
}

func testAccCheckSesTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_template" {
			continue
		}
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			input := ses.GetTemplateInput{
				TemplateName: aws.String(rs.Primary.ID),
			}

			gto, err := conn.GetTemplate(&input)
			if err != nil {
				if isAWSErr(err, ses.ErrCodeTemplateDoesNotExistException, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			if gto.Template != nil {
				return resource.RetryableError(fmt.Errorf("Template exists: %v", gto.Template))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAwsSesTemplateResourceConfigBasic1(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = "%s"
  subject = "subject"
  html    = "html"
}
`, name)
}

func testAccCheckAwsSesTemplateResourceConfigBasic2(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = "%s"
  subject = "subject"
  html    = "html"
  text    = "text"
}
`, name)
}

func testAccCheckAwsSesTemplateResourceConfigBasic3(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = "%s"
  subject = "subject"
  html    = "html update"
}
`, name)
}
