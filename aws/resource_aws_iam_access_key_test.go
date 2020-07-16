package aws

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"
)

func TestAccAWSAccessKey_basic(t *testing.T) {
	var conf iam.AccessKeyMetadata
	rName := fmt.Sprintf("test-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists("aws_iam_access_key.a_key", &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Active"),
					resource.TestCheckResourceAttrSet("aws_iam_access_key.a_key", "secret"),
				),
			},
		},
	})
}

func TestAccAWSAccessKey_pgpencrypted(t *testing.T) {
	var conf iam.AccessKeyMetadata
	rName := fmt.Sprintf("test-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig_pgpencrypted(rName, testPubKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists("aws_iam_access_key.a_key", &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Active"),
					testDecryptSecretKeyAndTest("aws_iam_access_key.a_key", testPrivKey1),
					resource.TestCheckNoResourceAttr(
						"aws_iam_access_key.a_key", "secret"),
					resource.TestCheckResourceAttrSet(
						"aws_iam_access_key.a_key", "encrypted_secret"),
					resource.TestCheckResourceAttrSet(
						"aws_iam_access_key.a_key", "key_fingerprint"),
				),
			},
		},
	})
}

func TestAccAWSAccessKey_rsaencrypted(t *testing.T) {
	var conf iam.AccessKeyMetadata
	rName := fmt.Sprintf("test-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig_rsaencrypted(rName, testRSAPubKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists("aws_iam_access_key.a_key", &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Active"),
					testDecryptSecretKeyAndTest("aws_iam_access_key.a_key", testRSAPrivKey1),
					resource.TestCheckNoResourceAttr(
						"aws_iam_access_key.a_key", "secret"),
					resource.TestCheckResourceAttrSet(
						"aws_iam_access_key.a_key", "encrypted_secret"),
					resource.TestCheckNoResourceAttr(
						"aws_iam_access_key.a_key", "key_fingerprint"),
				),
			},
		},
	})
}

func TestAccAWSAccessKey_inactive(t *testing.T) {
	var conf iam.AccessKeyMetadata
	rName := fmt.Sprintf("test-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists("aws_iam_access_key.a_key", &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Active"),
					resource.TestCheckResourceAttrSet("aws_iam_access_key.a_key", "secret"),
				),
			},
			{
				Config: testAccAWSAccessKeyConfig_inactive(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists("aws_iam_access_key.a_key", &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Inactive"),
					resource.TestCheckResourceAttrSet("aws_iam_access_key.a_key", "secret"),
				),
			},
		},
	})
}

func testAccCheckAWSAccessKeyDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_access_key" {
			continue
		}

		// Try to get access key
		resp, err := iamconn.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.AccessKeyMetadata) > 0 {
				return fmt.Errorf("still exist.")
			}
			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSAccessKeyExists(n string, res *iam.AccessKeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		name := rs.Primary.Attributes["user"]

		resp, err := iamconn.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: aws.String(name),
		})
		if err != nil {
			return err
		}

		if len(resp.AccessKeyMetadata) != 1 ||
			*resp.AccessKeyMetadata[0].UserName != name {
			return fmt.Errorf("User not found not found")
		}

		*res = *resp.AccessKeyMetadata[0]

		return nil
	}
}

func testAccCheckAWSAccessKeyAttributes(accessKeyMetadata *iam.AccessKeyMetadata, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.Contains(*accessKeyMetadata.UserName, "test-user") {
			return fmt.Errorf("Bad username: %s", *accessKeyMetadata.UserName)
		}

		if *accessKeyMetadata.Status != status {
			return fmt.Errorf("Bad status: %s", *accessKeyMetadata.Status)
		}

		return nil
	}
}

func testDecryptSecretKeyAndTest(nAccessKey, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keyResource, ok := s.RootModule().Resources[nAccessKey]
		if !ok {
			return fmt.Errorf("Not found: %s", nAccessKey)
		}

		password, ok := keyResource.Primary.Attributes["encrypted_secret"]
		if !ok {
			return errors.New("No password in state")
		}

		// We can't verify that the decrypted password is correct, because we don't
		// have it. We can verify that decrypting it does not error
		if strings.HasPrefix(key, "-----BEGIN RSA PRIVATE KEY-----") {
			_, err := rsaOAEPDecryptBytes(password, key)
			if err != nil {
				return fmt.Errorf("Error decrypting password: %s", err)
			}
		} else {
			_, err := pgpkeys.DecryptBytes(password, key)
			if err != nil {
				return fmt.Errorf("Error decrypting password: %s", err)
			}
		}
		return nil
	}
}

func testAccAWSAccessKeyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "a_user" {
  name = "%s"
}

resource "aws_iam_access_key" "a_key" {
  user = "${aws_iam_user.a_user.name}"
}
`, rName)
}

func testAccAWSAccessKeyConfig_pgpencrypted(rName, key string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "a_user" {
  name = "%s"
}

resource "aws_iam_access_key" "a_key" {
  user = "${aws_iam_user.a_user.name}"

  pgp_key = <<EOF
%s
EOF
}
`, rName, key)
}

func testAccAWSAccessKeyConfig_rsaencrypted(rName, key string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "a_user" {
  name = "%s"
}

resource "aws_iam_access_key" "a_key" {
  user = "${aws_iam_user.a_user.name}"

  rsa_key = <<EOF
%s
EOF
}
`, rName, key)
}

func testAccAWSAccessKeyConfig_inactive(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "a_user" {
  name = "%s"
}

resource "aws_iam_access_key" "a_key" {
  user   = "${aws_iam_user.a_user.name}"
  status = "Inactive"
}
`, rName)
}

func TestSesSmtpPasswordFromSecretKeySigV4(t *testing.T) {
	cases := []struct {
		Region   string
		Input    string
		Expected string
	}{
		{"eu-central-1", "some+secret+key", "BMXhUYlu5Z3gSXVQORxlVa7XPaz91aGWdfHxvkOZdWZ2"},
		{"eu-central-1", "another+secret+key", "BBbphbrQmrKMx42d1N6+C7VINYEBGI5v9VsZeTxwskfh"},
		{"us-west-1", "some+secret+key", "BH+jbMzper5WwlwUar9E1ySBqHa9whi0GPo+sJ0mVYJj"},
		{"us-west-1", "another+secret+key", "BKVmjjMDFk/qqw8EROW99bjCS65PF8WKvK5bSr4Y6EqF"},
	}

	for _, tc := range cases {
		actual, err := sesSmtpPasswordFromSecretKeySigV4(&tc.Input, tc.Region)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if actual != tc.Expected {
			t.Fatalf("%q: expected %q, got %q", tc.Input, tc.Expected, actual)
		}
	}
}

func TestSesSmtpPasswordFromSecretKeySigV2(t *testing.T) {
	cases := []struct {
		Input    string
		Expected string
	}{
		{"some+secret+key", "AnkqhOiWEcszZZzTMCQbOY1sPGoLFgMH9zhp4eNgSjo4"},
		{"another+secret+key", "Akwqr0Giwi8FsQFgW3DXWCC2DiiQ/jZjqLDWK8TeTBgL"},
	}

	for _, tc := range cases {
		actual, err := sesSmtpPasswordFromSecretKeySigV2(&tc.Input)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if actual != tc.Expected {
			t.Fatalf("%q: expected %q, got %q", tc.Input, tc.Expected, actual)
		}
	}
}

func rsaOAEPDecryptBytes(cipherText string, key string) ([]byte, error) {
	rsaPrivKey, err := pemDecodePrivateKey(key)
	if err != nil {
		return nil, err
	}

	b64, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptOAEP(sha512.New(), rand.Reader, rsaPrivKey, b64, nil)
}

func pemDecodePrivateKey(key string) (*rsa.PrivateKey, error) {
	privPem, _ := pem.Decode([]byte(key))

	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("error parsing pem encoded private key: %s", privPem.Type)
	}

	return x509.ParsePKCS1PrivateKey(privPem.Bytes)
}

const testRSAPubKey1 = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA7N5eaBspDrcmSeSsyAw4
3SJ/mRSdkLSljyX41txY8pOYrPwnIX3ZjkSUsTma+tBuJbb61BIjJEUM1DA0RIXp
/Q18FFgObMIXraSh8yGleCIn5R34zppR7nudv+RwkOAmSHch5mQ3HtNX5OIVY1Gr
PP/cXSpuclkWNLMrKKxTpJOYG7L2jplBo3aerpgYcK+awaO3AOUnbZ4MCpF+4InW
X5souXxg/1HwPk+nZdlBQG6hjIcEqR5rZYQJXZDABCOAyvQ19OSjVoLqsMPWCf5M
N+LyU/qVrixZVEt5pqYMgPkaEz0q5haIN3AgdMRUw/PxjbDqysMFvBIthp0SIBsq
9HP7reGJ45q1/gtpRp5hKKg5WyLWghGcFISszVkiLNP3wVCT2REPpsT7QDL8DGsp
wo6RbVQ1nwjem8M2fbKouFQ0lRgkLpbBBIbSlbn5X7JXF83BA7FbDgkKEjHrT3YQ
pGmjpPONaMzL3/obLEPViuIFXc5WW5OjX3BhsGiiJYFuS0s/0hunF/fmmjL7w0MO
CPAfLXIBRD9uowKBNsxf6K9shnK7kbrplxAwY2f2XcsDx8dCqAvuBormvEJLV8Rp
NH3CgSfNSmP+TH2cfl7fDFIy1Ce6dfl3yVfxIV+ajoK5iZDSI5kARyEextH5fo/2
fFNXTe7PkOnsOKDHSQfkqEMCAwEAAQ==
-----END PUBLIC KEY-----
`

const testRSAPrivKey1 = `-----BEGIN RSA PRIVATE KEY-----
MIIJKgIBAAKCAgEA7N5eaBspDrcmSeSsyAw43SJ/mRSdkLSljyX41txY8pOYrPwn
IX3ZjkSUsTma+tBuJbb61BIjJEUM1DA0RIXp/Q18FFgObMIXraSh8yGleCIn5R34
zppR7nudv+RwkOAmSHch5mQ3HtNX5OIVY1GrPP/cXSpuclkWNLMrKKxTpJOYG7L2
jplBo3aerpgYcK+awaO3AOUnbZ4MCpF+4InWX5souXxg/1HwPk+nZdlBQG6hjIcE
qR5rZYQJXZDABCOAyvQ19OSjVoLqsMPWCf5MN+LyU/qVrixZVEt5pqYMgPkaEz0q
5haIN3AgdMRUw/PxjbDqysMFvBIthp0SIBsq9HP7reGJ45q1/gtpRp5hKKg5WyLW
ghGcFISszVkiLNP3wVCT2REPpsT7QDL8DGspwo6RbVQ1nwjem8M2fbKouFQ0lRgk
LpbBBIbSlbn5X7JXF83BA7FbDgkKEjHrT3YQpGmjpPONaMzL3/obLEPViuIFXc5W
W5OjX3BhsGiiJYFuS0s/0hunF/fmmjL7w0MOCPAfLXIBRD9uowKBNsxf6K9shnK7
kbrplxAwY2f2XcsDx8dCqAvuBormvEJLV8RpNH3CgSfNSmP+TH2cfl7fDFIy1Ce6
dfl3yVfxIV+ajoK5iZDSI5kARyEextH5fo/2fFNXTe7PkOnsOKDHSQfkqEMCAwEA
AQKCAgEAwpMMBALDoFXsuuiA0jfQAj8Th+E6aaMrGML9fSo2WtXJpdfgIQ/rRYWq
i0ahu4S55ns/4jMf8OxT1H3ggaVrh7arBV8sQkTSBI7nhfxOm7ebBAex2a1Evl2H
QRlbKncmm4JZM5OA/+5mFhttrE9rFcmr8FApt/7cUeAYBOpCL0AaxC4ngQ39sFSB
lzRTZ0WpH4Xnj3GuCMq6Y3gPPE2d7p1bP7sfRry9V8JA5VYo1s/KHtDOkEnvuM1U
kCqWwQ+U/aLMK+YhErCqLxg/26esXoArxbZjfFbr7mWtgaqHPO7jb4hgk+9QaBHQ
Z3rbkrfx5g29YWAAdvSLtzeRqxQGHIHcwBuy6jlFpe0j2y6uB4nSkevQ1nqmBeYd
CY+2/MKs2P3YizWhJ9AtUUp1T5WJlUntSBs0DJpp1aw71e0LUAuv7DhkGHiFyxkn
j2zCpvP+hMfshk5TFSGVdoSTVq+P7wLtgP35S3I5YAbvoZMUxhq2eeQOJQrg5XCl
+VNizoHFCGUFTC1j6NKhl8GLJ8ILbWU97o9DbJh9rmM8lTXPS1dLqpMyH/4hp6Br
ltlkDOXRMdg5HVDoEmrb34Ly/Kcp1yHFP/s4LIOdGEGPOyBXzkgbsPw8HPVjxIXR
a0l+Qja9JmGltPfiACSvJn0cGEjMUq7xv+Wdrw8gRIpLVy9zJdECggEBAO4/xtTt
lphAG6/EB/+g3ndkZQHkcqfAYcl9KBfenk4DwG4513pDSudDRDd/KKK1Rp4GBMYL
VbnHojxdBHYgX/sjndrsahfd4UsTssUCxyrt8nEw5YgmSN54NeY5T9bpmXnHUkBC
LWmmEK/kG2pEe/pLzWN1trOHjfb3wL94VbSgMRzkP0n2CYt6dUm2XOsygxDNRNod
CTWeIPFXV+lvt1ELo6PN7uOpzxhtUcXZPlRHTpioznSRGz7iTXLHJfMJOfhAALPt
iGJfkqETgRSsppAfd52WZHKkCGjx3f5jO64BZysPqBYI5Xcb3P+P1e40gLLN77f6
2401iw08yBZrCbsCggEBAP6EQuGDb5WUZKEwBpEzzcZI5qJaFJIXFFKuCZP75bU/
BPmlJhNO/SFH7JOMja4Yg2Xx2d8CQK5BQN/gEYFry0X26X6ySHOT9ZyPURtAmUSH
xV1kyC1HkMsXMqV+mPmSL42eefPuT97bXpkXp6LYmHt7jXKayxzJRR6vn/JMe5/N
qzPLYOUfCjx/MSQgo92t8Qsfu11xz1nTSl071+xCzPoy+qz46PJRXM18khoMAxb5
YWdYh2zCNuCmxDr6dv8wieKhY3BBKoOQfUa1XzGAyQBCq9WHzfnEoxYzKWbxkKxZ
nP+012eGGLhBvANexttcNoxi7pKtGnSS8GKdV39QTxkCggEBAOIrPQ+Jc/qYcTAM
AiPTjRz9+z0upwFEihACdfsi5FjfCuAYnMt98i9UFgAxAFxlheIhiIqQ3BJ+xy5S
hxe0aCk6iHH5GEYL5gGlc4G+v1+rfwmhB2SWI3q91zz0jyxPmdiXNSe3KvEuKo0v
GC9rT94t845Fguku97/JNfsNux67K6RnsQT/QdRcrdcJ/W9xBib/FiuQgNubf15c
MiJyYS3YxMGNjwgkfWqM7KHLN3Y+MwiMx2C1F448upUJJKdwzM0zxPcJuPzaCZJL
t6/urjgHB0BcMoL54NnvMXR6s37d6hhgxooUa/EETGl3G/kDcmFLShP3WlDU0WoB
l1hgyF0CggEAYo+PYstGU9OCYJU7hdFc13N1tNtTaft8CESTOvZqTxTXdWGJ7CJD
jjbPG5hraUbe44STzXOO3qwwVkHsJYU70505cLHTssZSb81kKDyM3egB+xfDGR3E
qZETMNlkngkJVztOmLLpxTCIYpqxdTORYQhIj1/4Ve5vUOHL+8W4ffrkWItiu3eY
vDK5Mfdd3cO1O4yPPzGKjYtwGcjJ5hk8TYueXYuKkgQF/yFFZsbOD4CFQsMatnhD
Th3mkbxahpoiW1wKKPdjYk165f3onj/0FqC68FbF4fpO+ZLYbqAPWV7emHtMiy93
tSrnhxqVwW7lRNou7ygPvaMFafrqXkgYSQKCAQEA31NDiwgM+Atr2B+D2bUPH2kB
mqNOjBENCOM92F4IZ/Ohh+fK+4KgyV5L71lcXJxJZjoXkiv3hK5Ol08CWETRu/GL
fsqkEW5vy0FU0mgg4iUrp/sulA4fwmQ2wOOp4gLgqVmJ/GYioGitGyvroEeD0ez7
YDTuNq4iPdKvs8it+194aZTO15WOspjQ4K/l6np+RM5WUmZLO8n075XY86/RU6YW
E6R/IjQ0OH2kFHV5zhqmTeW/FI7wNB096pkoKG6eyB/MdTW8O1EAFAqHClvVrAEy
/3d6Q5gqsieofWP9eHwB9HEol5mxmYz6MH6cAffpIEV3V+6Fth/+TzjhseGlCQ==
-----END RSA PRIVATE KEY-----
`
