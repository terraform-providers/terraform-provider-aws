package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cognitoidentityprovider/finder"
)

func TestAccAWSCognitoUserPoolUICustomization_AllClients_CSS(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	css := ".label-customizable {font-weight: 400;}"
	cssUpdated := ".label-customizable {font-weight: 100;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
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

func TestAccAWSCognitoUserPoolUICustomization_AllClients_Disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"

	css := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCognitoUserPoolUICustomization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPoolUICustomization_AllClients_ImageFile(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	filename := "testdata/service/cognitoidentityprovider/logo.png"
	updatedFilename := "testdata/service/cognitoidentityprovider/logo_modified.png"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_Image(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_Image(rName, updatedFilename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
		},
	})
}

func TestAccAWSCognitoUserPoolUICustomization_AllClients_CSSAndImageFile(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	css := ".label-customizable {font-weight: 400;}"
	filename := "testdata/service/cognitoidentityprovider/logo.png"
	updatedFilename := "testdata/service/cognitoidentityprovider/logo_modified.png"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSSAndImage(rName, css, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_Image(rName, updatedFilename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
		},
	})
}

func TestAccAWSCognitoUserPoolUICustomization_Client_CSS(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	css := ".label-customizable {font-weight: 400;}"
	cssUpdated := ".label-customizable {font-weight: 100;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", clientResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_CSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "css_version"),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", clientResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
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

func TestAccAWSCognitoUserPoolUICustomization_Client_Disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"

	css := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCognitoUserPoolUICustomization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPoolUICustomization_Client_Image(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	filename := "testdata/service/cognitoidentityprovider/logo.png"
	updatedFilename := "testdata/service/cognitoidentityprovider/logo_modified.png"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_Image(rName, filename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", clientResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_Image(rName, updatedFilename),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", clientResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "image_url"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_file"},
			},
		},
	})
}

func TestAccAWSCognitoUserPoolUICustomization_ClientAndAll_CSS(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.ui_all"
	clientUIResourceName := "aws_cognito_user_pool_ui_customization.ui_client"

	clientResourceName := "aws_cognito_user_pool_client.test"
	userPoolResourceName := "aws_cognito_user_pool.test"

	allCSS := ".label-customizable {font-weight: 400;}"
	clientCSS := ".label-customizable {font-weight: 100;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				// Test UI Customization settings shared by ALL and a specific client
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_ClientAndAllCustomizations_CSS(rName, allCSS, allCSS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					testAccCheckAWSCognitoUserPoolUICustomizationExists(clientUIResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "css", allCSS),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
					resource.TestCheckResourceAttrPair(clientUIResourceName, "client_id", clientResourceName, "id"),
					resource.TestCheckResourceAttrSet(clientUIResourceName, "creation_date"),
					resource.TestCheckResourceAttr(clientUIResourceName, "css", allCSS),
					resource.TestCheckResourceAttrSet(clientUIResourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(clientUIResourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      clientUIResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test UI Customization settings overridden for the client
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_ClientAndAllCustomizations_CSS(rName, allCSS, clientCSS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					testAccCheckAWSCognitoUserPoolUICustomizationExists(clientUIResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "css", allCSS),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", userPoolResourceName, "id"),
					resource.TestCheckResourceAttrPair(clientUIResourceName, "client_id", clientResourceName, "id"),
					resource.TestCheckResourceAttrSet(clientUIResourceName, "creation_date"),
					resource.TestCheckResourceAttr(clientUIResourceName, "css", clientCSS),
					resource.TestCheckResourceAttrSet(clientUIResourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(clientUIResourceName, "user_pool_id", userPoolResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      clientUIResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCognitoUserPoolUICustomization_UpdateClientToAll_CSS(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	css := ".label-customizable {font-weight: 100;}"
	cssUpdated := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", clientResourceName, "id"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
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

func TestAccAWSCognitoUserPoolUICustomization_UpdateAllToClient_CSS(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cognito_user_pool_ui_customization.test"
	clientResourceName := "aws_cognito_user_pool_client.test"

	css := ".label-customizable {font-weight: 100;}"
	cssUpdated := ".label-customizable {font-weight: 400;}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserPoolUICustomizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, css),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", css),
					resource.TestCheckResourceAttr(resourceName, "client_id", "ALL"),
				),
			},
			{
				Config: testAccAWSCognitoUserPoolUICustomizationConfig_Client_CSS(rName, cssUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserPoolUICustomizationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "css", cssUpdated),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", clientResourceName, "id"),
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

func testAccCheckAWSCognitoUserPoolUICustomizationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_ui_customization" {
			continue
		}

		userPoolId, clientId, err := parseCognitoUserPoolUICustomizationID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Cognito User Pool UI customization ID (%s): %w", rs.Primary.ID, err)
		}

		output, err := finder.CognitoUserPoolUICustomization(conn, userPoolId, clientId)

		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			continue
		}

		// Catch cases where the User Pool Domain has been destroyed, effectively eliminating
		// a UI customization; calls to GetUICustomization will fail
		if tfawserr.ErrMessageContains(err, cognitoidentityprovider.ErrCodeInvalidParameterException, "There has to be an existing domain associated with this user pool") {
			continue
		}

		if err != nil {
			return err
		}

		if testAccAWSCognitoUserPoolUICustomizationExists(output) {
			return fmt.Errorf("Cognito User Pool UI Customization (UserPoolId: %s, ClientId: %s) still exists", userPoolId, clientId)
		}
	}

	return nil
}

func testAccCheckAWSCognitoUserPoolUICustomizationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Client ID set")
		}

		userPoolId, clientId, err := parseCognitoUserPoolUICustomizationID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Cognito User Pool UI customization ID (%s): %w", rs.Primary.ID, err)
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		output, err := finder.CognitoUserPoolUICustomization(conn, userPoolId, clientId)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Cognito User Pool UI customization (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSS(rName, css string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  css = %q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, css)
}

func testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_Image(rName, filename string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  image_file = filebase64(%q)

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, filename)
}

func testAccAWSCognitoUserPoolUICustomizationConfig_AllClients_CSSAndImage(rName, css, filename string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  css        = %q
  image_file = filebase64(%q)

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, css, filename)
}

func testAccAWSCognitoUserPoolUICustomizationConfig_Client_CSS(rName, css string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  client_id = aws_cognito_user_pool_client.test.id
  css       = %q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state 
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, css)
}

func testAccAWSCognitoUserPoolUICustomizationConfig_Client_Image(rName, filename string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "test" {
  client_id  = aws_cognito_user_pool_client.test.id
  image_file = filebase64(%q)

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, filename)
}

func testAccAWSCognitoUserPoolUICustomizationConfig_ClientAndAllCustomizations_CSS(rName, allCSS, clientCSS string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_ui_customization" "ui_all" {
  css = %q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}

resource "aws_cognito_user_pool_ui_customization" "ui_client" {
  client_id = aws_cognito_user_pool_client.test.id
  css       = %q

  # Refer to the aws_cognito_user_pool_domain resource's
  # user_pool_id attribute to ensure it is in an 'Active' state
  user_pool_id = aws_cognito_user_pool_domain.test.user_pool_id
}
`, rName, allCSS, clientCSS)
}

// testAccAWSCognitoUserPoolUICustomizationExists validates the API object such that
// we define resource existence when the object is non-nil and
// at least one of the object's fields are non-nil with the exception of CSSVersion
// which remains as an artifact even after UI customization removal
func testAccAWSCognitoUserPoolUICustomizationExists(ui *cognitoidentityprovider.UICustomizationType) bool {
	if ui == nil {
		return false
	}

	if ui.CSS != nil {
		return true
	}

	if ui.CreationDate != nil {
		return true
	}

	if ui.ImageUrl != nil {
		return true
	}

	if ui.LastModifiedDate != nil {
		return true
	}

	return false
}
