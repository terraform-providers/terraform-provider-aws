package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestFindRegionByEc2Endpoint(t *testing.T) {
	var testCases = []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "does-not-exist",
			ErrCount: 1,
		},
		{
			Value:    "ec2.does-not-exist.amazonaws.com",
			ErrCount: 1,
		},
		{
			Value:    "us-east-1",
			ErrCount: 1,
		},
		{
			Value:    "ec2.us-east-1.amazonaws.com",
			ErrCount: 0,
		},
	}

	for _, tc := range testCases {
		_, err := findRegionByEc2Endpoint(tc.Value)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Value, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Value)
		}
	}
}

func TestFindRegionByName(t *testing.T) {
	var testCases = []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "does-not-exist",
			ErrCount: 1,
		},
		{
			Value:    "ec2.us-east-1.amazonaws.com",
			ErrCount: 1,
		},
		{
			Value:    "us-east-1",
			ErrCount: 0,
		},
	}

	for _, tc := range testCases {
		_, err := findRegionByName(tc.Value)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Value, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Value)
		}
	}
}

func TestAccDataSourceAwsRegion_basic(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", "ec2.us-east-1.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "name", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "description", "US East (N. Virginia)"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_endpoint(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	endpoint1 := "ec2.us-east-1.amazonaws.com"
	endpoint2 := "ec2.us-east-2.amazonaws.com"
	name1 := "us-east-1"
	name2 := "us-east-2"
	description1 := "US East (N. Virginia)"
	description2 := "US East (Ohio)"
	resourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_endpoint(endpoint1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config: testAccDataSourceAwsRegionConfig_endpoint(endpoint2),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
				),
			},
			{
				Config:      testAccDataSourceAwsRegionConfig_endpoint("does-not-exist"),
				ExpectError: regexp.MustCompile(`region not found for endpoint: does-not-exist`),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_endpointAndName(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	endpoint1 := "ec2.us-east-1.amazonaws.com"
	endpoint2 := "ec2.us-east-2.amazonaws.com"
	name1 := "us-east-1"
	name2 := "us-east-2"
	description1 := "US East (N. Virginia)"
	description2 := "US East (Ohio)"
	resourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_endpointAndName(endpoint1, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config: testAccDataSourceAwsRegionConfig_endpointAndName(endpoint2, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
				),
			},
			{
				Config: testAccDataSourceAwsRegionConfig_endpointAndName(endpoint1, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config:      testAccDataSourceAwsRegionConfig_endpointAndName(endpoint1, name2),
				ExpectError: regexp.MustCompile(`multiple regions matched`),
			},
			{
				Config:      testAccDataSourceAwsRegionConfig_endpointAndName(endpoint2, name1),
				ExpectError: regexp.MustCompile(`multiple regions matched`),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_name(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	endpoint1 := "ec2.us-east-1.amazonaws.com"
	endpoint2 := "ec2.us-east-2.amazonaws.com"
	name1 := "us-east-1"
	name2 := "us-east-2"
	description1 := "US East (N. Virginia)"
	description2 := "US East (Ohio)"
	resourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config: testAccDataSourceAwsRegionConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
				),
			},
			{
				Config:      testAccDataSourceAwsRegionConfig_name("does-not-exist"),
				ExpectError: regexp.MustCompile(`region not found for name: does-not-exist`),
			},
		},
	})
}

func testAccDataSourceAwsRegionCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

const testAccDataSourceAwsRegionConfig_empty = `
data "aws_region" "test" {}
`

func testAccDataSourceAwsRegionConfig_endpoint(endpoint string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  endpoint = "%s"
}
`, endpoint)
}

func testAccDataSourceAwsRegionConfig_endpointAndName(endpoint, name string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  endpoint = "%s"
  name     = "%s"
}
`, endpoint, name)
}

func testAccDataSourceAwsRegionConfig_name(name string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  name = "%s"
}
`, name)
}
