package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iam_instance_profile", &resource.Sweeper{
		Name:         "aws_iam_instance_profile",
		F:            testSweepIamInstanceProfile,
		Dependencies: []string{"aws_iam_role"},
	})
}

func testSweepIamInstanceProfile(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).iamconn

	var sweeperErrs *multierror.Error

	err = conn.ListInstanceProfilesPages(&iam.ListInstanceProfilesInput{}, func(page *iam.ListInstanceProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceProfile := range page.InstanceProfiles {
			name := aws.StringValue(instanceProfile.InstanceProfileName)

			if !iamRoleNameFilter(name) {
				log.Printf("[INFO] Skipping IAM Instance Profile (%s): no match on allow-list", name)
				continue
			}

			r := resourceAwsIamInstanceProfile()
			d := r.Data(nil)
			d.SetId(name)

			roles := instanceProfile.Roles
			if r := len(roles); r > 1 {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("unexpected number of roles for IAM Instance Profile (%s): %d", name, r))
			} else if r == 1 {
				d.Set("role", roles[0].RoleName)
			}

			log.Printf("[INFO] Sweeping IAM Instance Profile %q", name)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting IAM Instance Profile (%s): %w", name, err))
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Instance Profile sweep for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IAM Instance Profiles: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSIAMInstanceProfile_basic(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamInstanceProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("instance-profile/test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSIAMInstanceProfile_withoutRole(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamInstanceProfileConfigWithoutRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSIAMInstanceProfile_tags(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamInstanceProfileConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
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
				Config: testAccAwsIamInstanceProfileConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsIamInstanceProfileConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSIAMInstanceProfile_namePrefix(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	rName := acctest.RandString(5)
	resourceName := "aws_iam_instance_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInstanceProfilePrefixNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
					testAccCheckAWSInstanceProfileGeneratedNamePrefix(
						resourceName, "test-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSIAMInstanceProfile_disappears(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamInstanceProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIamInstanceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMInstanceProfile_disappears_role(t *testing.T) {
	var conf iam.GetInstanceProfileOutput
	resourceName := "aws_iam_instance_profile.test"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamInstanceProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInstanceProfileExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIamRole(), "aws_iam_role.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSInstanceProfileGeneratedNamePrefix(resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

func testAccCheckAWSInstanceProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_instance_profile" {
			continue
		}

		// Try to get role
		_, err := conn.GetInstanceProfile(&iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckAWSInstanceProfileExists(n string, res *iam.GetInstanceProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Instance Profile name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := conn.GetInstanceProfile(&iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccAwsIamInstanceProfileBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "test-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccAwsIamInstanceProfileConfig(rName string) string {
	return testAccAwsIamInstanceProfileBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%[1]s"
  role = aws_iam_role.test.name
}
`, rName)
}

func testAccAwsIamInstanceProfileConfigWithoutRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%s"
}
`, rName)
}

func testAccAWSInstanceProfilePrefixNameConfig(rName string) string {
	return testAccAwsIamInstanceProfileBaseConfig(rName) + `
resource "aws_iam_instance_profile" "test" {
  name_prefix = "test-"
  role        = aws_iam_role.test.name
}
`
}

func testAccAwsIamInstanceProfileConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAwsIamInstanceProfileBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%[1]s"
  role = aws_iam_role.test.name

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsIamInstanceProfileConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAwsIamInstanceProfileBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = "test-%[1]s"
  role = aws_iam_role.test.name

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
