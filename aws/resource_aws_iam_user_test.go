package aws

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pquerna/otp/totp"
)

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

	var sweeperErrs *multierror.Error
	for _, user := range users {
		username := aws.StringValue(user.UserName)
		log.Printf("[DEBUG] Deleting IAM User: %s", username)

		listUserPoliciesInput := &iam.ListUserPoliciesInput{
			UserName: user.UserName,
		}
		listUserPoliciesOutput, err := conn.ListUserPolicies(listUserPoliciesInput)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM User (%s) inline policies: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, inlinePolicyName := range listUserPoliciesOutput.PolicyNames {
			log.Printf("[DEBUG] Deleting IAM User (%s) inline policy %q", username, *inlinePolicyName)

			input := &iam.DeleteUserPolicyInput{
				PolicyName: inlinePolicyName,
				UserName:   user.UserName,
			}

			if _, err := conn.DeleteUserPolicy(input); err != nil {
				if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
					continue
				}
				sweeperErr := fmt.Errorf("error deleting IAM User (%s) inline policy %q: %s", username, *inlinePolicyName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		listAttachedUserPoliciesInput := &iam.ListAttachedUserPoliciesInput{
			UserName: user.UserName,
		}
		listAttachedUserPoliciesOutput, err := conn.ListAttachedUserPolicies(listAttachedUserPoliciesInput)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM User (%s) attached policies: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, attachedPolicy := range listAttachedUserPoliciesOutput.AttachedPolicies {
			policyARN := aws.StringValue(attachedPolicy.PolicyArn)

			log.Printf("[DEBUG] Detaching IAM User (%s) attached policy: %s", username, policyARN)

			if err := detachPolicyFromUser(conn, username, policyARN); err != nil {
				sweeperErr := fmt.Errorf("error detaching IAM User (%s) attached policy (%s): %s", username, policyARN, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if err := deleteAwsIamUserGroupMemberships(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) group memberships: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := deleteAwsIamUserAccessKeys(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) access keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := deleteAwsIamUserSSHKeys(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) SSH keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := deleteAwsIamUserVirtualMFADevices(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) virtual MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := deactivateAwsIamUserMFADevices(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := deleteAwsIamUserLoginProfile(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) login profile: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		input := &iam.DeleteUserInput{
			UserName: aws.String(username),
		}

		_, err = conn.DeleteUser(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM User (%s): %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSUser_basic(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path1 := "/"
	path2 := "/path2/"
	resourceName := "aws_iam_user.user"

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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
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

func TestAccAWSUser_ForceDestroy_SigningCertificate(t *testing.T) {
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
					testAccCheckAWSUserUploadSigningCertificate(&user),
				),
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

func TestAccAWSUser_nameChange(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", acctest.RandInt())
	path := "/"
	resourceName := "aws_iam_user.user"

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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
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
	resourceName := "aws_iam_user.user"

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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
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
		password, err := generateIAMPassword(32)
		if err != nil {
			return err
		}
		input := &iam.CreateLoginProfileInput{
			Password: aws.String(password),
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
		authenticationCode1, err := totp.GenerateCode(secret, time.Now().Add(-30*time.Second))
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

func testAccCheckAWSUserUploadSigningCertificate(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		signingCertificate, err := ioutil.ReadFile("./test-fixtures/iam-ssl-unix-line-endings.pem")
		if err != nil {
			return fmt.Errorf("error reading signing certificate fixture: %s", err)
		}
		input := &iam.UploadSigningCertificateInput{
			CertificateBody: aws.String(string(signingCertificate)),
			UserName:        getUserOutput.User.UserName,
		}

		if _, err := iamconn.UploadSigningCertificate(input); err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Signing Certificate : %s", aws.StringValue(getUserOutput.User.UserName), err)
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
  name          = %q
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
