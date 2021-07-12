package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iam_service_specific_credential", &resource.Sweeper{
		Name: "aws_iam_service_specific_credential",
		F:    testSweepIamServiceSpecificCredentials,
	})
}

func testSweepIamServiceSpecificCredentials(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).iamconn

	var sweeperErrs *multierror.Error

	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}
	users := make([]*iam.User, 0)

	err = conn.ListUsersPages(&iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}

		return !lastPage
	})

	for _, user := range users {
		userName := aws.StringValue(user.UserName)
		input := &iam.ListServiceSpecificCredentialsInput{
			UserName: aws.String(userName),
		}

		creds, err := conn.ListServiceSpecificCredentials(input)
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM Service Specific Credentials for IAM user (%s): %w", userName, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, cred := range creds.ServiceSpecificCredentials {
			credId := aws.StringValue(cred.ServiceSpecificCredentialId)
			r := resourceAwsIamServiceSpecificCredential()
			d := r.Data(nil)
			d.Set("service_specific_credential_id", credId)
			d.Set("user_name", userName)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Service Specific Credential (%s): %w", credId, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}
	}
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM SAML Provider sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM SAML Providers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSIAMServiceSpecificCredential_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_service_specific_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServiceSpecificCredentialBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMServiceSpecificCredentialExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "user_name", "aws_iam_user.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttrSet(resourceName, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service_specific_credential_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"service_password",
				},
			},
		},
	})
}

func TestAccAWSIAMServiceSpecificCredential_status(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_service_specific_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServiceSpecificCredentialConfigStatus(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMServiceSpecificCredentialExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"service_password",
				},
			},
			{
				Config: testAccIAMServiceSpecificCredentialConfigStatus(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMServiceSpecificCredentialExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
				),
			},
			{
				Config: testAccIAMServiceSpecificCredentialConfigStatus(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMServiceSpecificCredentialExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
				),
			},
		},
	})
}

func TestAccAWSIAMServiceSpecificCredential_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_service_specific_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServiceSpecificCredentialBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMServiceSpecificCredentialExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIamServiceSpecificCredential(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIAMServiceSpecificCredentialDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_service_specific_credential" {
			continue
		}

		serviceName, userName, err := decodeAwsIamServiceSpecificCredential(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &iam.ListServiceSpecificCredentialsInput{
			ServiceName: aws.String(serviceName),
			UserName:    aws.String(userName),
		}

		out, err := conn.ListServiceSpecificCredentials(input)
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error reading IAM Service Specific Credential (%s): %w", rs.Primary.ID, err)
		}

		if out == nil || len(out.ServiceSpecificCredentials) == 0 {
			return fmt.Errorf("error reading IAM Service Specific Credential: no results found")
		}

		if len(out.ServiceSpecificCredentials) > 1 {
			return fmt.Errorf("error reading IAM Service Specific Credential: multiple results found, try adjusting search criteria")
		}
	}

	return nil
}

func testAccCheckIAMServiceSpecificCredentialExists(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not Found: %s", id)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		serviceName, userName, err := decodeAwsIamServiceSpecificCredential(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &iam.ListServiceSpecificCredentialsInput{
			ServiceName: aws.String(serviceName),
			UserName:    aws.String(userName),
		}

		_, err = conn.ListServiceSpecificCredentials(input)
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}

		return err
	}
}

func testAccIAMServiceSpecificCredentialBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}
`, rName)
}

func testAccIAMServiceSpecificCredentialConfigStatus(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
  status       = %[2]q
}
`, rName, status)
}
