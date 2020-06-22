package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestCleanRecordName(t *testing.T) {
	cases := []struct {
		Input, Output string
	}{
		{"www.nonexample.com", "www.nonexample.com"},
		{"\\052.nonexample.com", "*.nonexample.com"},
		{"\\100.nonexample.com", "@.nonexample.com"},
		{"\\043.nonexample.com", "#.nonexample.com"},
		{"nonexample.com", "nonexample.com"},
	}

	for _, tc := range cases {
		actual := cleanRecordName(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestExpandRecordName(t *testing.T) {
	cases := []struct {
		Input, Output string
	}{
		{"www", "www.nonexample.com"},
		{"www.", "www.nonexample.com"},
		{"dev.www", "dev.www.nonexample.com"},
		{"*", "*.nonexample.com"},
		{"nonexample.com", "nonexample.com"},
		{"test.nonexample.com", "test.nonexample.com"},
		{"test.nonexample.com.", "test.nonexample.com"},
	}

	zone_name := "nonexample.com"
	for _, tc := range cases {
		actual := expandRecordName(tc.Input, zone_name)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestNormalizeAwsAliasName(t *testing.T) {
	cases := []struct {
		Input, Output string
	}{
		{"www.nonexample.com", "www.nonexample.com"},
		{"www.nonexample.com.", "www.nonexample.com"},
		{"dualstack.name-123456789.region.elb.amazonaws.com", "name-123456789.region.elb.amazonaws.com"},
		{"dualstack.test-987654321.region.elb.amazonaws.com", "test-987654321.region.elb.amazonaws.com"},
		{"dualstacktest.com", "dualstacktest.com"},
		{"NAME-123456789.region.elb.amazonaws.com", "name-123456789.region.elb.amazonaws.com"},
	}

	for _, tc := range cases {
		actual := normalizeAwsAliasName(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestParseRecordId(t *testing.T) {
	cases := []struct {
		Input, Zone, Name, Type, Set string
	}{
		{"ABCDEF_test.notexample.com_A", "ABCDEF", "test.notexample.com", "A", ""},
		{"ABCDEF_test.notexample.com._A", "ABCDEF", "test.notexample.com", "A", ""},
		{"ABCDEF_test.notexample.com_A_set1", "ABCDEF", "test.notexample.com", "A", "set1"},
		{"ABCDEF__underscore.notexample.com_A", "ABCDEF", "_underscore.notexample.com", "A", ""},
		{"ABCDEF__underscore.notexample.com_A_set1", "ABCDEF", "_underscore.notexample.com", "A", "set1"},
	}

	for _, tc := range cases {
		parts := parseRecordId(tc.Input)
		if parts[0] != tc.Zone {
			t.Fatalf("input: %s\noutput: %s\nexpected:%s", tc.Input, parts[0], tc.Zone)
		}
		if parts[1] != tc.Name {
			t.Fatalf("input: %s\noutput: %s\nexpected:%s", tc.Input, parts[1], tc.Name)
		}
		if parts[2] != tc.Type {
			t.Fatalf("input: %s\noutput: %s\nexpected:%s", tc.Input, parts[2], tc.Type)
		}
		if parts[3] != tc.Set {
			t.Fatalf("input: %s\noutput: %s\nexpected:%s", tc.Input, parts[3], tc.Set)
		}
	}
}

func TestAccAWSRoute53Record_basic(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_underscored(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.underscore"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigUnderscoreInName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_disappears(t *testing.T) {
	var record1 route53.ResourceRecordSet
	var zone1 route53.GetHostedZoneOutput
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.main", &zone1),
					testAccCheckRoute53RecordExists(resourceName, &record1),
					testAccCheckRoute53RecordDisappears(&zone1, &record1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53Record_disappears_MultipleRecords(t *testing.T) {
	var record1, record2, record3, record4, record5 route53.ResourceRecordSet
	var zone1 route53.GetHostedZoneOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigMultiple,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.test", &zone1),
					testAccCheckRoute53RecordExists("aws_route53_record.test.0", &record1),
					testAccCheckRoute53RecordExists("aws_route53_record.test.1", &record2),
					testAccCheckRoute53RecordExists("aws_route53_record.test.2", &record3),
					testAccCheckRoute53RecordExists("aws_route53_record.test.3", &record4),
					testAccCheckRoute53RecordExists("aws_route53_record.test.4", &record5),
					testAccCheckRoute53RecordDisappears(&zone1, &record1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53Record_basic_fqdn(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfig_fqdn,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},

			// Ensure that changing the name to include a trailing "dot" results in
			// nothing happening, because the name is stripped of trailing dots on
			// save. Otherwise, an update would occur and due to the
			// create_before_destroy, the record would actually be destroyed, and a
			// non-empty plan would appear, and the record will fail to exist in
			// testAccCheckRoute53RecordExists
			{
				Config: testAccRoute53RecordConfig_fqdn_no_op,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record2),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_txtSupport(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"zone_id"}, // just for this test
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigTXT,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight", "zone_id"},
			},
		},
	})
}

func TestAccAWSRoute53Record_spfSupport(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		IDRefreshName:       resourceName,
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckRoute53RecordDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigSPF,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
					resource.TestCheckResourceAttr(
						resourceName, "records.2930149397", "include:notexample.com"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_caaSupport(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		IDRefreshName:       resourceName,
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckRoute53RecordDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigCAA,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
					resource.TestCheckResourceAttr(
						resourceName, "records.2965463512", "0 issue \"exampleca.com;\""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_generatesSuffix(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigSuffix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_wildcard(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.wildcard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53WildCardRecordConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRoute53WildCardRecordConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record2),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_failover(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.www-primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53FailoverCNAMERecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
					testAccCheckRoute53RecordExists("aws_route53_record.www-secondary", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_weighted_basic(t *testing.T) {
	var record1, record2, record3 route53.ResourceRecordSet
	resourceName := "aws_route53_record.www-live"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route53_record.www-live",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53WeightedCNAMERecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists("aws_route53_record.www-dev", &record1),
					testAccCheckRoute53RecordExists(resourceName, &record2),
					testAccCheckRoute53RecordExists("aws_route53_record.www-off", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_weighted_to_simple_basic(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.www-server1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccRoute53RecordConfigWithWeightedRoutingPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
			{
				Config: testaccRoute53RecordConfigWithSimpleRoutingPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_Alias_Elb(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.alias"

	rs := acctest.RandString(10)
	config := fmt.Sprintf(testAccRoute53RecordConfigAliasElb, rs)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_Alias_S3(t *testing.T) {
	var record1 route53.ResourceRecordSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_record.alias"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigAliasS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_Alias_VpcEndpoint(t *testing.T) {
	var record1 route53.ResourceRecordSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_record.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoute53RecordConfigAliasCustomVpcEndpointSwappedAliasAttributes(rName),
				ExpectError: regexp.MustCompile(`expected length of`),
			},
			{
				Config: testAccRoute53RecordConfigCustomVpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_Alias_Uppercase(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.alias"

	rs := acctest.RandString(10)
	config := fmt.Sprintf(testAccRoute53RecordConfigAliasElbUppercase, rs)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_weighted_alias(t *testing.T) {
	var record1, record2, record3, record4, record5, record6 route53.ResourceRecordSet
	resourceName := "aws_route53_record.elb_weighted_alias_live"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53WeightedElbAliasRecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
					testAccCheckRoute53RecordExists("aws_route53_record.elb_weighted_alias_dev", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},

			{
				Config: testAccRoute53WeightedR53AliasRecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists("aws_route53_record.green_origin", &record3),
					testAccCheckRoute53RecordExists("aws_route53_record.r53_weighted_alias_live", &record4),
					testAccCheckRoute53RecordExists("aws_route53_record.blue_origin", &record5),
					testAccCheckRoute53RecordExists("aws_route53_record.r53_weighted_alias_dev", &record6),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_geolocation_basic(t *testing.T) {
	var record1, record2, record3, record4 route53.ResourceRecordSet
	resourceName := "aws_route53_record.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53GeolocationCNAMERecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists("aws_route53_record.default", &record1),
					testAccCheckRoute53RecordExists("aws_route53_record.california", &record2),
					testAccCheckRoute53RecordExists("aws_route53_record.oceania", &record3),
					testAccCheckRoute53RecordExists("aws_route53_record.denmark", &record4),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_latency_basic(t *testing.T) {
	var record1, record2, record3 route53.ResourceRecordSet
	resourceName := "aws_route53_record.us-east-1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53LatencyCNAMERecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
					testAccCheckRoute53RecordExists("aws_route53_record.eu-west-1", &record2),
					testAccCheckRoute53RecordExists("aws_route53_record.ap-northeast-1", &record3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_TypeChange(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.sample"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordTypeChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRoute53RecordTypeChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record2),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_SetIdentifierChange(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.basic_to_weighted"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordSetIdentifierChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRoute53RecordSetIdentifierChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record2),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_AliasChange(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.elb_alias_change"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordAliasChangePre,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},

			// Cause a change, which will trigger a refresh
			{
				Config: testAccRoute53RecordAliasChangePost,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record2),
				),
			},
		},
	})
}

func TestAccAWSRoute53Record_empty(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigEmptyName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/8423
func TestAccAWSRoute53Record_longTXTrecord(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.long_txt"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53RecordConfigLongTxtRecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists(resourceName, &record1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_multivalue_answer_basic(t *testing.T) {
	var record1, record2 route53.ResourceRecordSet
	resourceName := "aws_route53_record.www-server1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53MultiValueAnswerARecord,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53RecordExists("aws_route53_record.www-server1", &record1),
					testAccCheckRoute53RecordExists("aws_route53_record.www-server2", &record2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func TestAccAWSRoute53Record_allowOverwrite(t *testing.T) {
	var record1 route53.ResourceRecordSet
	resourceName := "aws_route53_record.overwriting"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53RecordDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoute53RecordConfig_allowOverwrite(false),
				ExpectError: regexp.MustCompile(`Tried to create resource record set \[name='www.notexample.com.', type='A'] but it already exists`),
			},
			{
				Config: testAccRoute53RecordConfig_allowOverwrite(true),
				Check:  resource.ComposeTestCheckFunc(testAccCheckRoute53RecordExists("aws_route53_record.overwriting", &record1)),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_overwrite", "weight"},
			},
		},
	})
}

func testAccCheckRoute53RecordDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_record" {
			continue
		}

		parts := parseRecordId(rs.Primary.ID)
		zone := parts[0]
		name := parts[1]
		rType := parts[2]

		en := expandRecordName(name, "notexample.com")

		lopts := &route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(cleanZoneID(zone)),
			StartRecordName: aws.String(en),
			StartRecordType: aws.String(rType),
		}

		resp, err := conn.ListResourceRecordSets(lopts)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// if NoSuchHostedZone, then all the things are destroyed
				if awsErr.Code() == "NoSuchHostedZone" {
					return nil
				}
			}
			return err
		}
		if len(resp.ResourceRecordSets) == 0 {
			return nil
		}
		rec := resp.ResourceRecordSets[0]
		if FQDN(*rec.Name) == FQDN(name) && *rec.Type == rType {
			return fmt.Errorf("Record still exists: %#v", rec)
		}
	}
	return nil
}

func testAccCheckRoute53RecordDisappears(zone *route53.GetHostedZoneOutput, resourceRecordSet *route53.ResourceRecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn

		input := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: zone.HostedZone.Id,
			ChangeBatch: &route53.ChangeBatch{
				Comment: aws.String("Deleted by Terraform"),
				Changes: []*route53.Change{
					{
						Action:            aws.String(route53.ChangeActionDelete),
						ResourceRecordSet: resourceRecordSet,
					},
				},
			},
		}

		respRaw, err := deleteRoute53RecordSet(conn, input)
		if err != nil {
			return fmt.Errorf("error deleting resource record set: %s", err)
		}

		changeInfo := respRaw.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo
		if changeInfo == nil {
			return nil
		}

		if err := waitForRoute53RecordSetToSync(conn, cleanChangeID(*changeInfo.Id)); err != nil {
			return fmt.Errorf("error waiting for resource record set deletion: %s", err)
		}

		return nil
	}
}

func testAccCheckRoute53RecordExists(n string, resourceRecordSet *route53.ResourceRecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No hosted zone ID is set")
		}

		parts := parseRecordId(rs.Primary.ID)
		zone := parts[0]
		name := parts[1]
		rType := parts[2]

		en := expandRecordName(name, "notexample.com")

		lopts := &route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(cleanZoneID(zone)),
			StartRecordName: aws.String(en),
			StartRecordType: aws.String(rType),
		}

		resp, err := conn.ListResourceRecordSets(lopts)
		if err != nil {
			return err
		}
		if len(resp.ResourceRecordSets) == 0 {
			return fmt.Errorf("Record does not exist")
		}

		// rec := resp.ResourceRecordSets[0]
		for _, rec := range resp.ResourceRecordSets {
			recName := cleanRecordName(*rec.Name)
			if FQDN(strings.ToLower(recName)) == FQDN(strings.ToLower(en)) && *rec.Type == rType {
				*resourceRecordSet = *rec

				return nil
			}
		}
		return fmt.Errorf("Record does not exist: %#v", rs.Primary.ID)
	}
}

func testAccRoute53RecordConfig_allowOverwrite(allowOverwrite bool) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "notexample.com."
}

resource "aws_route53_record" "default" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www.notexample.com"
  type = "A"
  ttl = "30"
  records = ["127.0.0.1"]
}

resource "aws_route53_record" "overwriting" {
  depends_on = ["aws_route53_record.default"]

  allow_overwrite = %v
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www.notexample.com"
  type = "A"
  ttl = "30"
  records = ["127.0.0.1"]
}
`, allowOverwrite)
}

const testAccRoute53RecordConfig = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "www.NOTexamplE.com"
	type = "A"
	ttl = "30"
	records = ["127.0.0.1", "127.0.0.27"]
}
`

const testAccRoute53RecordConfigMultiple = `
resource "aws_route53_zone" "test" {
  name = "notexample.com"
}

resource "aws_route53_record" "test" {
  count = 5

  name    = "record${count.index}.${aws_route53_zone.test.name}"
  records = ["127.0.0.${count.index}"]
  ttl     = "30"
  type    = "A"
  zone_id = "${aws_route53_zone.test.zone_id}"
}
`

const testAccRoute53RecordConfig_fqdn = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "default" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name    = "www.NOTexamplE.com"
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]

  lifecycle {
    create_before_destroy = true
  }
}
`

const testAccRoute53RecordConfig_fqdn_no_op = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "default" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name    = "www.NOTexamplE.com."
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]

  lifecycle {
    create_before_destroy = true
  }
}
`

const testAccRoute53RecordConfigSuffix = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "subdomain"
	type = "A"
	ttl = "30"
	records = ["127.0.0.1", "127.0.0.27"]
}
`

const testAccRoute53WildCardRecordConfig = `
resource "aws_route53_zone" "main" {
    name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "subdomain"
	type = "A"
	ttl = "30"
	records = ["127.0.0.1", "127.0.0.27"]
}

resource "aws_route53_record" "wildcard" {
    zone_id = "${aws_route53_zone.main.zone_id}"
    name = "*.notexample.com"
    type = "A"
    ttl = "30"
    records = ["127.0.0.1"]
}
`

const testAccRoute53WildCardRecordConfigUpdate = `
resource "aws_route53_zone" "main" {
    name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "subdomain"
	type = "A"
	ttl = "30"
	records = ["127.0.0.1", "127.0.0.27"]
}

resource "aws_route53_record" "wildcard" {
    zone_id = "${aws_route53_zone.main.zone_id}"
    name = "*.notexample.com"
    type = "A"
    ttl = "60"
    records = ["127.0.0.1"]
}
`
const testAccRoute53RecordConfigTXT = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "/hostedzone/${aws_route53_zone.main.zone_id}"
	name = "subdomain"
	type = "TXT"
	ttl = "30"
	records = ["lalalala"]
}
`
const testAccRoute53RecordConfigSPF = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "test"
	type = "SPF"
	ttl = "30"
	records = ["include:notexample.com"]
}

`
const testAccRoute53RecordConfigCAA = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "default" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "test"
	type = "CAA"
	ttl = "30"
	records = ["0 issue \"exampleca.com;\""]
}
`

const testAccRoute53FailoverCNAMERecord = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_health_check" "foo" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"

  tags = {
    Name = "tf-test-health-check"
   }
}

resource "aws_route53_record" "www-primary" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  failover_routing_policy {
    type = "PRIMARY"
  }
  health_check_id = "${aws_route53_health_check.foo.id}"
  set_identifier = "www-primary"
  records = ["primary.notexample.com"]
}

resource "aws_route53_record" "www-secondary" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  failover_routing_policy {
    type = "SECONDARY"
  }
  set_identifier = "www-secondary"
  records = ["secondary.notexample.com"]
}
`

const testAccRoute53WeightedCNAMERecord = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "www-dev" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  weighted_routing_policy {
	weight = 10
  }
  set_identifier = "dev"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "www-live" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  weighted_routing_policy {
	weight = 90
  }
  set_identifier = "live"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "www-off" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  weighted_routing_policy {
	weight = 0
  }
  set_identifier = "off"
  records = ["dev.notexample.com"]
}
`

const testAccRoute53GeolocationCNAMERecord = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "default" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  geolocation_routing_policy {
    country = "*"
  }
  set_identifier = "Default"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "california" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  geolocation_routing_policy {
    country = "US"
    subdivision = "CA"
  }
  set_identifier = "California"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "oceania" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  geolocation_routing_policy {
    continent = "OC"
  }
  set_identifier = "Oceania"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "denmark" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  geolocation_routing_policy {
    country = "DK"
  }
  set_identifier = "Denmark"
  records = ["dev.notexample.com"]
}
`

const testAccRoute53LatencyCNAMERecord = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "us-east-1" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  latency_routing_policy {
    region = "us-east-1"
  }
  set_identifier = "us-east-1"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "eu-west-1" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  latency_routing_policy {
    region = "eu-west-1"
  }
  set_identifier = "eu-west-1"
  records = ["dev.notexample.com"]
}

resource "aws_route53_record" "ap-northeast-1" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"
  ttl = "5"
  latency_routing_policy {
    region = "ap-northeast-1"
  }
  set_identifier = "ap-northeast-1"
  records = ["dev.notexample.com"]
}
`

const testAccRoute53RecordConfigAliasElb = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "alias" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "A"

  alias {
  	zone_id = "${aws_elb.main.zone_id}"
  	name = "${aws_elb.main.dns_name}"
  	evaluate_target_health = true
  }
}

resource "aws_elb" "main" {
  name = "foobar-terraform-elb-%s"
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 80
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

const testAccRoute53RecordConfigAliasElbUppercase = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "alias" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "A"

  alias {
  	zone_id = "${aws_elb.main.zone_id}"
  	name = "${aws_elb.main.dns_name}"
  	evaluate_target_health = true
  }
}

resource "aws_elb" "main" {
  name = "FOOBAR-TERRAFORM-ELB-%s"
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 80
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}
`

func testAccRoute53RecordConfigAliasS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_s3_bucket" "website" {
  bucket = %q
  acl    = "public-read"

  website {
    index_document = "index.html"
  }
}

resource "aws_route53_record" "alias" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name    = "www"
  type    = "A"

  alias {
    zone_id                = "${aws_s3_bucket.website.hosted_zone_id}"
    name                   = "${aws_s3_bucket.website.website_domain}"
    evaluate_target_health = true
  }
}
`, rName)
}

func testAccRoute53CustomVpcEndpointBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = ["${aws_subnet.test.id}"]
}

resource "aws_default_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = ["${aws_lb.test.id}"]
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = ["${aws_default_security_group.test.id}"]
  service_name        = "${aws_vpc_endpoint_service.test.service_name}"
  subnet_ids          = ["${aws_subnet.test.id}"]
  vpc_endpoint_type   = "Interface"
  vpc_id              = "${aws_vpc.test.id}"
}

resource "aws_route53_zone" "test" {
  name = "notexample.com"

  vpc {
    vpc_id = "${aws_vpc.test.id}"
  }
}
`, rName)
}

func testAccRoute53RecordConfigAliasCustomVpcEndpointSwappedAliasAttributes(rName string) string {
	return testAccRoute53CustomVpcEndpointBase(rName) + fmt.Sprintf(`
resource "aws_route53_record" "test" {
  name    = "test"
  type    = "A"
  zone_id = "${aws_route53_zone.test.zone_id}"

  alias {
    evaluate_target_health = false
    name                   = "${lookup(aws_vpc_endpoint.test.dns_entry[0], "hosted_zone_id")}"
    zone_id                = "${lookup(aws_vpc_endpoint.test.dns_entry[0], "dns_name")}"
  }
}
`)
}

func testAccRoute53RecordConfigCustomVpcEndpoint(rName string) string {
	return testAccRoute53CustomVpcEndpointBase(rName) + fmt.Sprintf(`
resource "aws_route53_record" "test" {
  name    = "test"
  type    = "A"
  zone_id = "${aws_route53_zone.test.zone_id}"

  alias {
    evaluate_target_health = false
    name                   = "${lookup(aws_vpc_endpoint.test.dns_entry[0], "dns_name")}"
    zone_id                = "${lookup(aws_vpc_endpoint.test.dns_entry[0], "hosted_zone_id")}"
  }
}
`)
}

const testAccRoute53WeightedElbAliasRecord = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_elb" "live" {
  name = "foobar-terraform-elb-live"
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 80
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

resource "aws_route53_record" "elb_weighted_alias_live" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "A"

  weighted_routing_policy {
	weight = 90
  }
  set_identifier = "live"

  alias {
    zone_id = "${aws_elb.live.zone_id}"
    name = "${aws_elb.live.dns_name}"
    evaluate_target_health = true
  }
}

resource "aws_elb" "dev" {
  name = "foobar-terraform-elb-dev"
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 80
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

resource "aws_route53_record" "elb_weighted_alias_dev" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "A"

  weighted_routing_policy {
	weight = 10
  }
  set_identifier = "dev"

  alias {
    zone_id = "${aws_elb.dev.zone_id}"
    name = "${aws_elb.dev.dns_name}"
    evaluate_target_health = true
  }
}
`

const testAccRoute53WeightedR53AliasRecord = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "blue_origin" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "blue-origin"
  type = "CNAME"
  ttl = 5
  records = ["v1.terraform.io"]
}

resource "aws_route53_record" "r53_weighted_alias_live" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"

  weighted_routing_policy {
	weight = 90
  }
  set_identifier = "blue"

  alias {
    zone_id = "${aws_route53_zone.main.zone_id}"
    name = "${aws_route53_record.blue_origin.name}.${aws_route53_zone.main.name}"
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "green_origin" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "green-origin"
  type = "CNAME"
  ttl = 5
  records = ["v2.terraform.io"]
}

resource "aws_route53_record" "r53_weighted_alias_dev" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "CNAME"

  weighted_routing_policy {
	weight = 10
  }
  set_identifier = "green"

  alias {
    zone_id = "${aws_route53_zone.main.zone_id}"
    name = "${aws_route53_record.green_origin.name}.${aws_route53_zone.main.name}"
    evaluate_target_health = false
  }
}
`

const testAccRoute53RecordTypeChangePre = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "sample" {
	zone_id = "${aws_route53_zone.main.zone_id}"
  name = "sample"
  type = "CNAME"
  ttl = "30"
  records = ["www.terraform.io"]
}
`

const testAccRoute53RecordTypeChangePost = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "sample" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "sample"
  type = "A"
  ttl = "30"
  records = ["127.0.0.1", "8.8.8.8"]
}
`

const testAccRoute53RecordSetIdentifierChangePre = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "basic_to_weighted" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "sample"
  type = "A"
  ttl = "30"
  records = ["127.0.0.1", "8.8.8.8"]
}
`

const testAccRoute53RecordSetIdentifierChangePost = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "basic_to_weighted" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "sample"
  type = "A"
  ttl = "30"
  records = ["127.0.0.1", "8.8.8.8"]
  set_identifier = "cluster-a"
  weighted_routing_policy {
    weight = 100
  }
}
`

const testAccRoute53RecordAliasChangePre = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_elb" "alias_change" {
  name = "foobar-tf-elb-alias-change"
  availability_zones = ["us-west-2a"]

  listener {
    instance_port = 80
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

resource "aws_route53_record" "elb_alias_change" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "alias-change"
  type = "A"

  alias {
    zone_id = "${aws_elb.alias_change.zone_id}"
    name = "${aws_elb.alias_change.dns_name}"
    evaluate_target_health = true
  }
}
`

const testAccRoute53RecordAliasChangePost = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "elb_alias_change" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "alias-change"
  type = "CNAME"
  ttl = "30"
  records = ["www.terraform.io"]
}
`

const testAccRoute53RecordConfigEmptyName = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "empty" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = ""
	type = "A"
	ttl = "30"
	records = ["127.0.0.1"]
}
`

const testAccRoute53RecordConfigLongTxtRecord = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "long_txt" {
    zone_id = "${aws_route53_zone.main.zone_id}"
    name = "google.notexample.com"
    type = "TXT"
    ttl = "30"
    records = [
        "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAiajKNMp\" \"/A12roF4p3MBm9QxQu6GDsBlWUWFx8EaS8TCo3Qe8Cj0kTag1JMjzCC1s6oM0a43JhO6mp6z/"
    ]
}
`

const testAccRoute53RecordConfigUnderscoreInName = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "underscore" {
	zone_id = "${aws_route53_zone.main.zone_id}"
	name = "_underscore.notexample.com"
	type = "A"
	ttl = "30"
	records = ["127.0.0.1"]
}
`

const testAccRoute53MultiValueAnswerARecord = `
resource "aws_route53_zone" "main" {
	name = "notexample.com"
}

resource "aws_route53_record" "www-server1" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "A"
  ttl = "5"
  multivalue_answer_routing_policy = true
  set_identifier = "server1"
  records = ["127.0.0.1"]
}

resource "aws_route53_record" "www-server2" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name = "www"
  type = "A"
  ttl = "5"
  multivalue_answer_routing_policy = true
  set_identifier = "server2"
  records = ["127.0.0.2"]
}
`

const testaccRoute53RecordConfigWithWeightedRoutingPolicy = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "www-server1" {
  zone_id = "${aws_route53_zone.main.zone_id}"
  name    = "www"
  type    = "A"

  weighted_routing_policy {
    weight = 5
  }

  ttl            = "300"
  set_identifier = "server1"
  records        = ["127.0.0.1"]
}
`

const testaccRoute53RecordConfigWithSimpleRoutingPolicy = `
resource "aws_route53_zone" "main" {
  name = "notexample.com"
}

resource "aws_route53_record" "www-server1" {
  zone_id        = "${aws_route53_zone.main.zone_id}"
  name           = "www"
  type           = "A"
  ttl            = "300"
  records        = ["127.0.0.1"]
}
`
