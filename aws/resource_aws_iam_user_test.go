package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pquerna/otp/totp"
)

func TestValidateIamUserName(t *testing.T) {
	validNames := []string{
		"test-user",
		"test_user",
		"testuser123",
		"TestUser",
		"Test-User",
		"test.user",
		"test.123,user",
		"testuser@hashicorp",
		"test+user@hashicorp.com",
	}
	for _, v := range validNames {
		_, errors := validateAwsIamUserName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid IAM User name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"!",
		"/",
		" ",
		":",
		";",
		"test name",
		"/slash-at-the-beginning",
		"slash-at-the-end/",
	}
	for _, v := range invalidNames {
		_, errors := validateAwsIamUserName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid IAM User name", v)
		}
	}
}

func init() {
	resource.AddTestSweepers("aws_iam_user", &resource.Sweeper{
		Name: "aws_iam_user",
		F:    testSweepIamUsers,
	})
}

func testSweepIamUsers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).iamconn
	prefixes := []string{
		"test-user",
		"tf-acc-test",
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

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM User sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving IAM Users: %s", err)
	}

	if len(users) == 0 {
		log.Print("[DEBUG] No IAM Users to sweep")
		return nil
	}

	for _, user := range users {
		username := aws.StringValue(user.UserName)
		log.Printf("[DEBUG] Deleting IAM User: %s", username)

		if err := deleteAwsIamUserGroupMemberships(conn, username); err != nil {
			return fmt.Errorf("error removing IAM User (%s) group memberships: %s", username, err)
		}

		if err := deleteAwsIamUserAccessKeys(conn, username); err != nil {
			return fmt.Errorf("error removing IAM User (%s) access keys: %s", username, err)
		}

		if err := deleteAwsIamUserSSHKeys(conn, username); err != nil {
			return fmt.Errorf("error removing IAM User (%s) SSH keys: %s", username, err)
		}

		if err := deleteAwsIamUserMFADevices(conn, username); err != nil {
			return fmt.Errorf("error removing IAM User (%s) MFA devices: %s", username, err)
		}

		if err := deleteAwsIamUserLoginProfile(conn, username); err != nil {
			return fmt.Errorf("error removing IAM User (%s) login profile: %s", username, err)
		}

		input := &iam.DeleteUserInput{
			UserName: aws.String(username),
		}

		_, err := conn.DeleteUser(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("Error deleting IAM User (%s): %s", username, err)
		}
	}

	return nil
}

func TestAccAWSUser_importBasic(t *testing.T) {
	resourceName := "aws_iam_user.user"

	n := fmt.Sprintf("test-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(n, "/"),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_basic(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path1 := "/"
	path2 := "/path2/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(name1, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
					testAccCheckAWSUserAttributes(&conf, name1, "/"),
				),
			},
			{
				Config: testAccAWSUserConfig(name2, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
					testAccCheckAWSUserAttributes(&conf, name2, "/path2/"),
				),
			},
		},
	})
}

func TestAccAWSUser_disappears(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					testAccCheckAWSUserDisappears(&user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_AccessKey(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfigForceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					testAccCheckAWSUserCreatesAccessKey(&user),
				),
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_LoginProfile(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfigForceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					testAccCheckAWSUserCreatesLoginProfile(&user),
				),
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_MFADevice(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfigForceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					testAccCheckAWSUserCreatesMFADevice(&user),
				),
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_SSHKey(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfigForceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					testAccCheckAWSUserUploadsSSHKey(&user),
				),
			},
		},
	})
}

func TestAccAWSUser_nameChange(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path := "/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(name1, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
			{
				Config: testAccAWSUserConfig(name2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccAWSUser_pathChange(t *testing.T) {
	var conf iam.GetUserOutput

	name := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path1 := "/"
	path2 := "/updated/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig(name, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
			{
				Config: testAccAWSUserConfig(name, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccAWSUser_permissionsBoundary(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.user"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", testAccGetPartition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", testAccGetPartition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckAWSUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			// Test update
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
					testAccCheckAWSUserPermissionsBoundary(&user, permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// Test removal
			{
				Config: testAccAWSUserConfig(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckAWSUserPermissionsBoundary(&user, ""),
				),
			},
			// Test addition
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckAWSUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccAWSUserConfig_permissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckAWSUserPermissionsBoundary(&user, ""),
				),
			},
		},
	})
}

func TestAccAWSUser_tags(t *testing.T) {
	var user iam.GetUserOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSUserConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test-Name"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-tag2"),
				),
			},
			{
				Config: testAccAWSUserConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-tagUpdate"),
				),
			},
		},
	})
}

func testAccCheckAWSUserDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user" {
			continue
		}

		// Try to get user
		_, err := iamconn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		if !isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSUserExists(n string, res *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckAWSUserAttributes(user *iam.GetUserOutput, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *user.User.UserName != name {
			return fmt.Errorf("Bad name: %s", *user.User.UserName)
		}

		if *user.User.Path != path {
			return fmt.Errorf("Bad path: %s", *user.User.Path)
		}

		return nil
	}
}

func testAccCheckAWSUserDisappears(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		userName := aws.StringValue(getUserOutput.User.UserName)

		_, err := iamconn.DeleteUser(&iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("error deleting user %q: %s", userName, err)
		}

		return nil
	}
}

func testAccCheckAWSUserPermissionsBoundary(getUserOutput *iam.GetUserOutput, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if getUserOutput.User.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *getUserOutput.User.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckAWSUserCreatesAccessKey(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		input := &iam.CreateAccessKeyInput{
			UserName: getUserOutput.User.UserName,
		}

		if _, err := iamconn.CreateAccessKey(input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Access Key: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckAWSUserCreatesLoginProfile(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		input := &iam.CreateLoginProfileInput{
			Password: aws.String(generateIAMPassword(32)),
			UserName: getUserOutput.User.UserName,
		}

		if _, err := iamconn.CreateLoginProfile(input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Login Profile: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckAWSUserCreatesMFADevice(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		createVirtualMFADeviceInput := &iam.CreateVirtualMFADeviceInput{
			Path:                 getUserOutput.User.Path,
			VirtualMFADeviceName: getUserOutput.User.UserName,
		}

		createVirtualMFADeviceOutput, err := iamconn.CreateVirtualMFADevice(createVirtualMFADeviceInput)
		if err != nil {
			return fmt.Errorf("error creating IAM User (%s) Virtual MFA Device: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		secret := string(createVirtualMFADeviceOutput.VirtualMFADevice.Base32StringSeed)
		authenticationCode1, err := totp.GenerateCode(secret, time.Now().Add(time.Duration(-30*time.Second)))
		if err != nil {
			return fmt.Errorf("error generating Virtual MFA Device authentication code 1: %s", err)
		}
		authenticationCode2, err := totp.GenerateCode(secret, time.Now())
		if err != nil {
			return fmt.Errorf("error generating Virtual MFA Device authentication code 2: %s", err)
		}

		enableVirtualMFADeviceInput := &iam.EnableMFADeviceInput{
			AuthenticationCode1: aws.String(authenticationCode1),
			AuthenticationCode2: aws.String(authenticationCode2),
			SerialNumber:        createVirtualMFADeviceOutput.VirtualMFADevice.SerialNumber,
			UserName:            getUserOutput.User.UserName,
		}

		if _, err := iamconn.EnableMFADevice(enableVirtualMFADeviceInput); err != nil {
			return fmt.Errorf("error enabling IAM User (%s) Virtual MFA Device: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckAWSUserUploadsSSHKey(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		sshKey, err := ioutil.ReadFile("./test-fixtures/public-ssh-key.pub")
		if err != nil {
			return fmt.Errorf("error reading SSH fixture: %s", err)
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		input := &iam.UploadSSHPublicKeyInput{
			UserName:         getUserOutput.User.UserName,
			SSHPublicKeyBody: aws.String(string(sshKey)),
		}

		_, err = iamconn.UploadSSHPublicKey(input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) SSH key: %s", *getUserOutput.User.UserName, err)
		}

		return nil
	}
}

func testAccAWSUserConfig(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %q
  path = %q
}
`, rName, path)
}

func testAccAWSUserConfig_permissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name                 = %q
  permissions_boundary = %q
}
`, rName, permissionsBoundary)
}

func testAccAWSUserConfigForceDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  force_destroy = true
  name = %q
}
`, rName)
}

func testAccAWSUserConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %q
  tags = {
    Name = "test-Name"
    tag2 = "test-tag2"
  }
}
`, rName)
}

func testAccAWSUserConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %q
  tags = {
    tag2 = "test-tagUpdate"
  }
}
`, rName)
}
