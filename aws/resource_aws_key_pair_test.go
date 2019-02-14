package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_key_pair", &resource.Sweeper{
		Name: "aws_key_pair",
		Dependencies: []string{
			"aws_elastic_beanstalk_environment",
			"aws_instance",
			"aws_spot_fleet_request",
		},
		F: testSweepKeyPairs,
	})
}

func testSweepKeyPairs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	ec2conn := client.(*AWSClient).ec2conn

	log.Printf("Destroying the tmp keys in (%s)", client.(*AWSClient).region)

	resp, err := ec2conn.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("key-name"),
				Values: []*string{
					aws.String("tf-acctest*"),
					aws.String("tf_acc*"),
					aws.String("tmp-key*"),
				},
			},
		},
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Key Pair sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing key pairs in Sweeper: %s", err)
	}

	keyPairs := resp.KeyPairs
	for _, d := range keyPairs {
		_, err := ec2conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{
			KeyName: d.KeyName,
		})

		if err != nil {
			return fmt.Errorf("Error deleting key pairs in Sweeper: %s", err)
		}
	}
	return nil
}

func TestAccAWSKeyPair_basic(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	fingerprint := "d7:ff:a6:63:18:64:9c:57:a1:ee:ca:a4:ad:c2:81:62"
	resourceName := "aws_key_pair.a_key_pair"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKeyPairConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKeyPairExists(resourceName, &keyPair),
					testAccCheckAWSKeyPairFingerprint(&keyPair, fingerprint),
					resource.TestCheckResourceAttr(resourceName, "fingerprint", fingerprint),
					resource.TestCheckResourceAttr(resourceName, "key_name", "tf-acc-key-pair"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
		},
	})
}

func TestAccAWSKeyPair_generatedName(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.a_key_pair"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKeyPairConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKeyPairExists(resourceName, &keyPair),
					testAccCheckAWSKeyPairKeyNamePrefix(&keyPair, "terraform-"),
					resource.TestMatchResourceAttr(resourceName, "key_name", regexp.MustCompile(`^terraform-`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key"},
			},
		},
	})
}

func TestAccAWSKeyPair_namePrefix(t *testing.T) {
	var keyPair ec2.KeyPairInfo
	resourceName := "aws_key_pair.a_key_pair"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_key_pair.a_key_pair",
		IDRefreshIgnore: []string{"key_name_prefix"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSKeyPairPrefixNameConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKeyPairExists(resourceName, &keyPair),
					testAccCheckAWSKeyPairKeyNamePrefix(&keyPair, "baz-"),
					resource.TestMatchResourceAttr(resourceName, "key_name", regexp.MustCompile(`^baz-`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_name_prefix", "public_key"},
			},
		},
	})
}

func testAccCheckAWSKeyPairDestroy(s *terraform.State) error {
	ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_key_pair" {
			continue
		}

		// Try to find key pair
		resp, err := ec2conn.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
			KeyNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.KeyPairs) > 0 {
				return fmt.Errorf("still exist.")
			}
			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidKeyPair.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSKeyPairFingerprint(conf *ec2.KeyPairInfo, expectedFingerprint string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(conf.KeyFingerprint) != expectedFingerprint {
			return fmt.Errorf("incorrect fingerprint. expected %s, got %s", expectedFingerprint, aws.StringValue(conf.KeyFingerprint))
		}
		return nil
	}
}

func testAccCheckAWSKeyPairKeyNamePrefix(conf *ec2.KeyPairInfo, namePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(aws.StringValue(conf.KeyName), namePrefix) {
			return fmt.Errorf("incorrect key name. expected %s prefix, got %s", namePrefix, aws.StringValue(conf.KeyName))
		}
		return nil
	}
}

func testAccCheckAWSKeyPairExists(n string, res *ec2.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KeyPair name is set")
		}

		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := ec2conn.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
			KeyNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.KeyPairs) != 1 ||
			*resp.KeyPairs[0].KeyName != rs.Primary.ID {
			return fmt.Errorf("KeyPair not found")
		}

		*res = *resp.KeyPairs[0]

		return nil
	}
}

const testAccAWSKeyPairConfig = `
resource "aws_key_pair" "a_key_pair" {
  key_name   = "tf-acc-key-pair"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
`

const testAccAWSKeyPairConfig_generatedName = `
resource "aws_key_pair" "a_key_pair" {
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
`

const testAccCheckAWSKeyPairPrefixNameConfig = `
resource "aws_key_pair" "a_key_pair" {
  key_name_prefix   = "baz-"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
`
