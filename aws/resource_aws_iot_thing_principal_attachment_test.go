package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIotThingPrincipalAttachment_basic(t *testing.T) {
	thingName := acctest.RandomWithPrefix("tf-acc")
	thingName2 := acctest.RandomWithPrefix("tf-acc2")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingPrincipalAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingPrincipalAttachmentConfig(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccAWSIotThingPrincipalAttachmentConfigUpdate1(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att2"),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName2, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccAWSIotThingPrincipalAttachmentConfigUpdate2(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName2, true, []string{}),
				),
			},
			{
				Config: testAccAWSIotThingPrincipalAttachmentConfigUpdate3(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att2"),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert", "aws_iot_certificate.cert2"}),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName2, false, []string{}),
				),
			},
			{
				Config: testAccAWSIotThingPrincipalAttachmentConfigUpdate4(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att2"),
					testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert2"}),
				),
			},
		},
	})
}

func testAccCheckAWSIotThingPrincipalAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_principal_attachment" {
			continue
		}

		principal := rs.Primary.Attributes["principal"]
		thing := rs.Primary.Attributes["thing"]

		found, err := getIoTThingPricipalAttachment(conn, thing, principal)

		if err != nil {
			return fmt.Errorf("Error: Failed listing principals for thing (%s): %s", thing, err)
		}

		if !found {
			continue
		}

		return fmt.Errorf("IOT Thing Principal Attachment (%s) still exists", rs.Primary.Attributes["id"])
	}

	return nil
}

func testAccCheckAWSIotThingPrincipalAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No attachment")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		thing := rs.Primary.Attributes["thing"]
		principal := rs.Primary.Attributes["principal"]

		found, err := getIoTThingPricipalAttachment(conn, thing, principal)

		if err != nil {
			return fmt.Errorf("Error: Failed listing principals for thing (%s), resource (%s): %s", thing, n, err)
		}

		if !found {
			return fmt.Errorf("Error: Principal (%s) is not attached to thing (%s), resource (%s)", principal, thing, n)
		}

		return nil
	}
}

func testAccCheckAWSIotThingPrincipalAttachmentStatus(thingName string, exists bool, principals []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotconn

		principalARNs := make(map[string]string)

		for _, p := range principals {
			pr, ok := s.RootModule().Resources[p]
			if !ok {
				return fmt.Errorf("Not found: %s", p)
			}
			principalARNs[pr.Primary.Attributes["arn"]] = p
		}

		thing, err := conn.DescribeThing(&iot.DescribeThingInput{
			ThingName: aws.String(thingName),
		})

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			if exists {
				return fmt.Errorf("Error: Thing (%s) exists, but expected to be removed", thingName)
			} else {
				return nil
			}
		} else if err != nil {
			return fmt.Errorf("Error: cannot describe thing %s: %s", thingName, err)
		} else if !exists {
			return fmt.Errorf("Error: Thing (%s) does not exist, but expected to be", thingName)
		}

		res, err := conn.ListThingPrincipals(&iot.ListThingPrincipalsInput{
			ThingName: aws.String(thingName),
		})

		if err != nil {
			return fmt.Errorf("Error: Cannot list thing (%s) principals: %s", thingName, err)
		}

		if len(res.Principals) != len(principalARNs) {
			return fmt.Errorf("Error: Thing (%s) has wrong number of principals attached", thing)
		}

		for _, p := range res.Principals {
			if principal, ok := principalARNs[aws.StringValue(p)]; !ok {
				return fmt.Errorf("Error: Principal %s is not attached to thing %s", principal, thingName)
			}
		}

		return nil
	}
}

func testAccAWSIotThingPrincipalAttachmentConfig(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert.arn}"
}
`, thingName)
}

func testAccAWSIotThingPrincipalAttachmentConfigUpdate1(thingName, thingName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing" "thing2" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert.arn}"
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing = "${aws_iot_thing.thing2.name}"
  principal = "${aws_iot_certificate.cert.arn}"
}
`, thingName, thingName2)
}

func testAccAWSIotThingPrincipalAttachmentConfigUpdate2(thingName, thingName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing" "thing2" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert.arn}"
}
`, thingName, thingName2)
}

func testAccAWSIotThingPrincipalAttachmentConfigUpdate3(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_certificate" "cert2" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert.arn}"
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert2.arn}"
}
`, thingName)
}

func testAccAWSIotThingPrincipalAttachmentConfigUpdate4(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert2" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert2.arn}"
}
`, thingName)
}
