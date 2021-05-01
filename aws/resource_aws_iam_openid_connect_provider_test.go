package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iam_openid_connect_provider", &resource.Sweeper{
		Name: "aws_iam_openid_connect_provider",
		F:    testSweepIamOpenIDConnectProvider,
	})
}

func testSweepIamOpenIDConnectProvider(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).iamconn

	var sweeperErrs *multierror.Error

	out, err := conn.ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})

	for _, oidcProvider := range out.OpenIDConnectProviderList {
		arn := aws.StringValue(oidcProvider.Arn)

		r := resourceAwsIamOpenIDConnectProvider()
		d := r.Data(nil)
		d.SetId(arn)
		err := r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM OIDC Provider (%s): %w", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM OIDC Provider sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM OIDC Providers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSIAMOpenIDConnectProvider_basic(t *testing.T) {
	rString := acctest.RandString(5)
	url := "accounts.testle.com/" + rString
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMOpenIDConnectProviderConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProvider(resourceName),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("oidc-provider/%s", url)),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.0",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMOpenIDConnectProviderConfig_modified(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.0",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.0", "cf23df2207d99a74fbe169e3eba035e633b65d94"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.1", "c784713d6f9cb67b55dd84f4e4af7832d42b8f55"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSIAMOpenIDConnectProvider_tags(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMOpenIDConnectProviderConfigTags1(rString, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccIAMOpenIDConnectProviderConfigTags2(rString, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIAMOpenIDConnectProviderConfigTags1(rString, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSIAMOpenIDConnectProvider_disappears(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMOpenIDConnectProviderConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProvider(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIamOpenIDConnectProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIAMOpenIDConnectProviderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_openid_connect_provider" {
			continue
		}

		input := &iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetOpenIDConnectProvider(input)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				// none found, that's good
				return nil
			}
			return fmt.Errorf("Error reading IAM OpenID Connect Provider, out: %s, err: %w", out, err)
		}

		if out != nil {
			return fmt.Errorf("Found IAM OpenID Connect Provider, expected none: %s", out)
		}
	}

	return nil
}

func testAccCheckIAMOpenIDConnectProvider(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not Found: %s", id)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		_, err := conn.GetOpenIDConnectProvider(&iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccIAMOpenIDConnectProviderConfig(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []
}
`, rString)
}

func testAccIAMOpenIDConnectProviderConfig_modified(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94", "c784713d6f9cb67b55dd84f4e4af7832d42b8f55"]
}
`, rString)
}

func testAccIAMOpenIDConnectProviderConfigTags1(rString, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []

  tags = {
    %[2]q = %[3]q
  }
}
`, rString, tagKey1, tagValue1)
}

func testAccIAMOpenIDConnectProviderConfigTags2(rString, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rString, tagKey1, tagValue1, tagKey2, tagValue2)
}
