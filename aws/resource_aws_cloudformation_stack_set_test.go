package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudformation_stack_set", &resource.Sweeper{
		Name: "aws_cloudformation_stack_set",
		Dependencies: []string{
			"aws_cloudformation_stack_set_instance",
		},
		F: testSweepCloudformationStackSets,
	})
}

func testSweepCloudformationStackSets(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).cfconn
	stackSets, err := listCloudFormationStackSets(conn)

	if testSweepSkipSweepError(err) || isAWSErr(err, "ValidationError", "AWS CloudFormation StackSets is not supported") {
		log.Printf("[WARN] Skipping CloudFormation StackSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets: %w", err)
	}

	var sweeperErrs *multierror.Error

	for _, stackSet := range stackSets {
		input := &cloudformation.DeleteStackSetInput{
			StackSetName: stackSet.StackSetName,
		}
		name := aws.StringValue(stackSet.StackSetName)

		log.Printf("[INFO] Deleting CloudFormation StackSet: %s", name)
		_, err := conn.DeleteStackSet(input)

		if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
			continue
		}

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting CloudFormation StackSet (%s): %w", name, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFormationStackSet_basic(t *testing.T) {
	var stackSet1 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName, "arn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudformation", regexp.MustCompile(`stackset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "AWSCloudFormationStackSetExecutionRole"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "stack_set_id", regexp.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccAWSCloudFormationStackSetTemplateBodyVpc(rName)+"\n"),
					resource.TestCheckNoResourceAttr(resourceName, "template_url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_disappears(t *testing.T) {
	var stackSet1 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationStackSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_AdministrationRoleArn(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamRoleResourceName1 := "aws_iam_role.test1"
	iamRoleResourceName2 := "aws_iam_role.test2"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigAdministrationRoleArn1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName1, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigAdministrationRoleArn2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_Description(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_ExecutionRoleName(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigExecutionRoleName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigExecutionRoleName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "name2"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_Name(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudFormationStackSetConfigName(""),
				ExpectError: regexp.MustCompile(`expected length`),
			},
			{
				Config:      testAccAWSCloudFormationStackSetConfigName(acctest.RandStringFromCharSet(129, acctest.CharSetAlpha)),
				ExpectError: regexp.MustCompile(`(cannot be longer|expected length)`),
			},
			{
				Config:      testAccAWSCloudFormationStackSetConfigName("1"),
				ExpectError: regexp.MustCompile(`must begin with alphabetic character`),
			},
			{
				Config:      testAccAWSCloudFormationStackSetConfigName("a_b"),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric and hyphen characters`),
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigName(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_Parameters(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigParameters1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigParameters2(rName, "value1updated", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigParameters1(rName, "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1updated"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_Parameters_Default(t *testing.T) {
	TestAccSkip(t, "this resource does not currently ignore unconfigured CloudFormation template parameters with the Default property")
	// Additional references:
	//  * https://github.com/hashicorp/terraform/issues/18863

	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigParametersDefault0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "defaultvalue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigParametersDefault1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigParametersDefault0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "defaultvalue"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_Parameters_NoEcho(t *testing.T) {
	TestAccSkip(t, "this resource does not currently ignore CloudFormation template parameters with the NoEcho property")
	// Additional references:
	//  * https://github.com/hashicorp/terraform-provider-aws/issues/55

	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigParametersNoEcho1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "****"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigParametersNoEcho1(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "****"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_Tags(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigTags1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigTags2(rName, "value1updated", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigTags1(rName, "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1updated"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_TemplateBody(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigTemplateBody(rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName+"1")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccAWSCloudFormationStackSetTemplateBodyVpc(rName+"1")+"\n"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigTemplateBody(rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName+"2")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccAWSCloudFormationStackSetTemplateBodyVpc(rName+"2")+"\n"),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStackSet_TemplateUrl(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetConfigTemplateUrl1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetConfigTemplateUrl2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(resourceName, &stackSet2),
					testAccCheckCloudFormationStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackSetExists(resourceName string, stackSet *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).cfconn

		input := &cloudformation.DescribeStackSetInput{
			StackSetName: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeStackSet(input)

		if err != nil {
			return err
		}

		if output == nil || output.StackSet == nil {
			return fmt.Errorf("CloudFormation StackSet (%s) not found", rs.Primary.ID)
		}

		*stackSet = *output.StackSet

		return nil
	}
}

func testAccCheckAWSCloudFormationStackSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack_set" {
			continue
		}

		input := cloudformation.DescribeStackSetInput{
			StackSetName: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeStackSet(&input)

		if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		if output != nil && output.StackSet != nil {
			return fmt.Errorf("CloudFormation StackSet (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudFormationStackSetNotRecreated(i, j *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackSetId) != aws.StringValue(j.StackSetId) {
			return fmt.Errorf("CloudFormation StackSet (%s) recreated", aws.StringValue(i.StackSetName))
		}

		return nil
	}
}

func testAccCheckCloudFormationStackSetRecreated(i, j *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackSetId) == aws.StringValue(j.StackSetId) {
			return fmt.Errorf("CloudFormation StackSet (%s) not recreated", aws.StringValue(i.StackSetName))
		}

		return nil
	}
}

func testAccPreCheckAWSCloudFormationStackSet(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).cfconn

	input := &cloudformation.ListStackSetsInput{}
	_, err := conn.ListStackSets(input)

	if testAccPreCheckSkipError(err) || isAWSErr(err, "ValidationError", "AWS CloudFormation StackSets is not supported") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSCloudFormationStackSetTemplateBodyParameters1(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccAWSCloudFormationStackSetTemplateBodyParameters2(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
  Parameter2:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Parameter2Value:
    Value: !Ref Parameter2
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccAWSCloudFormationStackSetTemplateBodyParametersDefault1(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
    Default: defaultvalue
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccAWSCloudFormationStackSetTemplateBodyParametersNoEcho1(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
    NoEcho: true
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccAWSCloudFormationStackSetTemplateBodyVpc(rName string) string {
	return fmt.Sprintf(`
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccAWSCloudFormationStackSetConfigAdministrationRoleArn1(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test1" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s1"
}

resource "aws_iam_role" "test2" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s2"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test1.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName))
}

func testAccAWSCloudFormationStackSetConfigAdministrationRoleArn2(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test1" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s1"
}

resource "aws_iam_role" "test2" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s2"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test2.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName))
}

func testAccAWSCloudFormationStackSetConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  description             = %[3]q
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName), description)
}

func testAccAWSCloudFormationStackSetConfigExecutionRoleName(rName, executionRoleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  execution_role_name     = %[3]q
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName), executionRoleName)
}

func testAccAWSCloudFormationStackSetConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName))
}

func testAccAWSCloudFormationStackSetConfigParameters1(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyParameters1(rName), value1)
}

func testAccAWSCloudFormationStackSetConfigParameters2(rName, value1, value2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
    Parameter2 = %[4]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyParameters2(rName), value1, value2)
}

func testAccAWSCloudFormationStackSetConfigParametersDefault0(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyParametersDefault1(rName))
}

func testAccAWSCloudFormationStackSetConfigParametersDefault1(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyParametersDefault1(rName), value1)
}

func testAccAWSCloudFormationStackSetConfigParametersNoEcho1(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyParametersNoEcho1(rName), value1)
}

func testAccAWSCloudFormationStackSetConfigTags1(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  tags = {
    Key1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName), value1)
}

func testAccAWSCloudFormationStackSetConfigTags2(rName, value1, value2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  tags = {
    Key1 = %[3]q
    Key2 = %[4]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName), value1, value2)
}

func testAccAWSCloudFormationStackSetConfigTemplateBody(rName, templateBody string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, templateBody)
}

func testAccAWSCloudFormationStackSetConfigTemplateUrl1(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  acl    = "public-read"
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  acl    = "public-read"
  bucket = aws_s3_bucket.test.bucket

  content = <<CONTENT
%[2]s
CONTENT

  key = "%[1]s-template1.yml"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q
  template_url            = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName+"1"))
}

func testAccAWSCloudFormationStackSetConfigTemplateUrl2(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  acl    = "public-read"
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  acl    = "public-read"
  bucket = aws_s3_bucket.test.bucket

  content = <<CONTENT
%[2]s
CONTENT

  key = "%[1]s-template2.yml"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q
  template_url            = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_bucket_object.test.key}"
}
`, rName, testAccAWSCloudFormationStackSetTemplateBodyVpc(rName+"2"))
}
