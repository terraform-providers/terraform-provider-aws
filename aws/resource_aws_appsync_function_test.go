package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsAppsyncFunction_basic(t *testing.T) {
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", acctest.RandString(8))
	rName3 := fmt.Sprintf("tfexample%s", acctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appsync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppsyncFunctionConfig(rName1, rName2, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName, &config),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile("apis/.+/functions/.+")),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "data_source", "aws_appsync_datasource.test", "name"),
				),
			},
			{
				Config: testAccAWSAppsyncFunctionConfig(rName1, rName3, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
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

func TestAccAwsAppsyncFunction_description(t *testing.T) {
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", acctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appsync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppsyncFunctionConfigDescription(rName1, rName2, testAccGetRegion(), "test description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "description", "test description 1"),
				),
			},
			{
				Config: testAccAWSAppsyncFunctionConfigDescription(rName1, rName2, testAccGetRegion(), "test description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "description", "test description 2"),
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

func TestAccAwsAppsyncFunction_responseMappingTemplate(t *testing.T) {
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", acctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appsync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppsyncFunctionConfigResponseMappingTemplate(rName1, rName2, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName, &config),
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

func TestAccAwsAppsyncFunction_disappears(t *testing.T) {
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", acctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appsync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppsyncFunctionConfig(rName1, rName2, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName, &config),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppsyncFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppsyncFunction_multipleFunctions(t *testing.T) {
	var function appsync.FunctionConfiguration
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt())
	resourceName := "aws_appsync_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(appsync.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, appsync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncFunction_multipleFunctions(rName, testAccGetRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncFunctionExists(resourceName+"1", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"2", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"3", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"4", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"5", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"6", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"7", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"8", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"9", &function),
					testAccCheckAwsAppsyncFunctionExists(resourceName+"10", &function),
				),
			},
		},
	})
}

func testAccCheckAwsAppsyncFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_function" {
			continue
		}

		apiID, functionID, err := decodeAppsyncFunctionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &appsync.GetFunctionInput{
			ApiId:      aws.String(apiID),
			FunctionId: aws.String(functionID),
		}

		_, err = conn.GetFunction(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncFunctionExists(name string, config *appsync.FunctionConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		apiID, functionID, err := decodeAppsyncFunctionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &appsync.GetFunctionInput{
			ApiId:      aws.String(apiID),
			FunctionId: aws.String(functionID),
		}

		output, err := conn.GetFunction(input)

		if err != nil {
			return err
		}

		*config = *output.FunctionConfiguration

		return nil
	}
}

func testAccAWSAppsyncFunctionConfig(r1, r2, region string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = "%[2]s"
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, testAccAppsyncDatasourceConfig_DynamoDBConfig_Region(r1, region), r2)
}

func testAccAWSAppsyncFunctionConfigDescription(r1, r2, region, description string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = "%[2]s"
  description              = "%[3]s"
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, testAccAppsyncDatasourceConfig_DynamoDBConfig_Region(r1, region), r2, description)
}

func testAccAWSAppsyncFunctionConfigResponseMappingTemplate(r1, r2, region string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = "%[2]s"
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, testAccAppsyncDatasourceConfig_DynamoDBConfig_Region(r1, region), r2)
}

func testAccAppsyncFunction_multipleFunctions(rName, region string) string {
	var functionResources string
	for i := 1; i <= 10; i++ {
		functionResources = functionResources + fmt.Sprintf(`
		resource "aws_appsync_function" "test%d" {
			api_id                   = aws_appsync_graphql_api.test.id
			data_source              = aws_appsync_datasource.test.name
			name                     = "test%d"
			request_mapping_template = <<EOF
		{
			"version": "2018-05-29",
			"method": "GET",
			"resourcePath": "/",
			"params":{
				"headers": $utils.http.copyheaders($ctx.request.headers)
			}
		}
		EOF

			response_mapping_template = <<EOF
		#if($ctx.result.statusCode == 200)
			$ctx.result.body
		#else
			$utils.appendError($ctx.result.body, $ctx.result.statusCode)
		#end
		EOF
		}
`, i, i)
	}

	return testAccAppsyncDatasourceConfig_DynamoDBConfig_Region(rName, region) + functionResources
}
