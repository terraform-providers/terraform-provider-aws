package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSesTemplate_Basic(t *testing.T) {
	name := acctest.RandString(5)
	var template ses.Template
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplate("aws_ses_template.test", &template),
					resource.TestCheckResourceAttr("aws_ses_template.test", "name", name),
					resource.TestCheckResourceAttr("aws_ses_template.test", "html", "html"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "subject", "subject"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "text", ""),
				),
			},
		},
	})
}

func TestAccAWSSesTemplate_Update(t *testing.T) {
	t.Skipf("Skip due to SES.UpdateTemplate eventual consistency issues")
	name := acctest.RandString(5)
	var template ses.Template
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplate("aws_ses_template.test", &template),
					resource.TestCheckResourceAttr("aws_ses_template.test", "name", name),
					resource.TestCheckResourceAttr("aws_ses_template.test", "html", "html"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "subject", "subject"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "text", ""),
				),
			},
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplate("aws_ses_template.test", &template),
					resource.TestCheckResourceAttr("aws_ses_template.test", "name", name),
					resource.TestCheckResourceAttr("aws_ses_template.test", "html", "html"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "subject", "subject"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "text", "text"),
				),
			},
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic3(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSesTemplate("aws_ses_template.test", &template),
					resource.TestCheckResourceAttr("aws_ses_template.test", "name", name),
					resource.TestCheckResourceAttr("aws_ses_template.test", "html", "html update"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "subject", "subject"),
					resource.TestCheckResourceAttr("aws_ses_template.test", "text", ""),
				),
			},
		},
	})
}

func TestAccAWSSesTemplate_Import(t *testing.T) {
	resourceName := "aws_ses_template.test"

	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSesTemplateResourceConfigBasic1(name),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSesTemplate(pr string, template *ses.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sesConn
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

		_, err := conn.GetTemplate(&input)
		return err
	}
}

func testAccCheckSesTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesConn

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
				if awsErr, ok := err.(awserr.Error); ok && (awsErr.Code() == "TemplateDoesNotExist") {
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
  name = "%s"
  subject = "subject"
  html = "html"
}
`, name)
}

func testAccCheckAwsSesTemplateResourceConfigBasic2(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name = "%s"
  subject = "subject"
  html = "html"
  text = "text"
}
`, name)
}

func testAccCheckAwsSesTemplateResourceConfigBasic3(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name = "%s"
  subject = "subject"
  html = "html update"
}
`, name)
}
