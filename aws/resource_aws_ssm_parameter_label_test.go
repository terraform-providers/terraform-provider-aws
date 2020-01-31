package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// Import test ignored for the timebeing. This is because labels are not returned by aws sdk after immediately attaching labels.

// func TestAccAWSSSMParameterLabel_importBasic(t *testing.T) {
// 	resourceName := "aws_ssm_parameter_label.foo"
// 	randName := acctest.RandString(5)
// 	randValue := acctest.RandString(5)

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSSSMParameterLabelBasicConfig(randName, "String", randValue, "label1"),
// 			},

// 			{
// 				ResourceName:            resourceName,
// 				ImportState:             true,
// 				ImportStateVerify:       true,
// 				ImportStateVerifyIgnore: []string{"overwrite"},
// 			},
// 		},
// 	})
// }

func TestAccAWSSSMParameterLabel_basic(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterLabelBasicConfig(name, "String", "bar", "label1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterLabelExists("aws_ssm_parameter_label.foo", &param, "label1"),
					resource.TestCheckResourceAttr("aws_ssm_parameter_label.foo", "labels.#", "1"),
					resource.TestCheckResourceAttr("aws_ssm_parameter_label.foo", "ssm_parameter_name", name),
					resource.TestCheckResourceAttr("aws_ssm_parameter_label.foo", "ssm_parameter_version", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSSSMParameterLabelExists(n string, param *ssm.Parameter, label string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Parameter Label ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		paramInput := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(fmt.Sprintf("%s:%s", rs.Primary.Attributes["ssm_parameter_name"], label)),
			},
			WithDecryption: aws.Bool(true),
		}

		resp, err := conn.GetParameters(paramInput)
		if err != nil {
			return err
		}

		if len(resp.Parameters) == 0 {
			return fmt.Errorf("Expected AWS SSM Parameter Label to be created, but wasn't found")
		}

		*param = *resp.Parameters[0]

		return nil
	}
}

func testAccAWSSSMParameterLabelBasicConfig(rName, pType, value string, label string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "foo" {
	name  = "%s"
	type  = "%s"
	value = "%s"	
	}

resource "aws_ssm_parameter_label" "foo" {
	ssm_parameter_name  = aws_ssm_parameter.foo.name
	ssm_parameter_version = aws_ssm_parameter.foo.version
	labels = ["%s"]
}
`, rName, pType, value, label)
}
