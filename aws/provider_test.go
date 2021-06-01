package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	// Provider name for single configuration testing
	ProviderNameAws = "aws"

	// Provider name for alternate configuration testing
	ProviderNameAwsAlternate = "awsalternate"

	// Provider name for alternate account and alternate region configuration testing
	ProviderNameAwsAlternateAccountAlternateRegion = "awsalternateaccountalternateregion"

	// Provider name for alternate account and same region configuration testing
	ProviderNameAwsAlternateAccountSameRegion = "awsalternateaccountsameregion"

	// Provider name for same account and alternate region configuration testing
	ProviderNameAwsSameAccountAlternateRegion = "awssameaccountalternateregion"

	// Provider name for third configuration testing
	ProviderNameAwsThird = "awsthird"
)

const rfc3339RegexPattern = `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?([Zz]|([+-]([01][0-9]|2[0-3]):[0-5][0-9]))$`
const uuidRegexPattern = `[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[ab89][a-f0-9]{3}-[a-f0-9]{12}`

// TestAccSkip implements a wrapper for (*testing.T).Skip() to prevent unused linting reports
//
// Reference: https://github.com/dominikh/go-tools/issues/633#issuecomment-606560616
var TestAccSkip = func(t *testing.T, message string) {
	t.Skip(message)
}

// testAccProviders is a static map containing only the main provider instance.
//
// Deprecated: Terraform Plugin SDK version 2 uses TestCase.ProviderFactories
// but supports this value in TestCase.Providers for backwards compatibility.
// In the future Providers: testAccProviders will be changed to
// ProviderFactories: testAccProviderFactories
var testAccProviders map[string]*schema.Provider

// testAccProviderFactories is a static map containing only the main provider instance
//
// Use other testAccProviderFactories functions, such as testAccProviderFactoriesAlternate,
// for tests requiring special provider configurations.
var testAccProviderFactories map[string]func() (*schema.Provider, error)

// testAccProvider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProvider *schema.Provider

// testAccProviderConfigure ensures testAccProvider is only configured once
//
// The testAccPreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// testAccProvider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

func init() {
	testAccProvider = Provider()

	testAccProviders = map[string]*schema.Provider{
		ProviderNameAws: testAccProvider,
	}

	// Always allocate a new provider instance each invocation, otherwise gRPC
	// ProviderConfigure() can overwrite configuration during concurrent testing.
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderNameAws: func() (*schema.Provider, error) { return Provider(), nil }, //nolint:unparam
	}
}

// testAccProviderFactoriesInit creates ProviderFactories for the provider under testing.
func testAccProviderFactoriesInit(providers *[]*schema.Provider, providerNames []string) map[string]func() (*schema.Provider, error) {
	var factories = make(map[string]func() (*schema.Provider, error), len(providerNames))

	for _, name := range providerNames {
		p := Provider()

		factories[name] = func() (*schema.Provider, error) { //nolint:unparam
			return p, nil
		}

		if providers != nil {
			*providers = append(*providers, p)
		}
	}

	return factories
}

// testAccProviderFactoriesInternal creates ProviderFactories for provider configuration testing
//
// This should only be used for TestAccAWSProvider_ tests which need to
// reference the provider instance itself. Other testing should use
// testAccProviderFactories or other related functions.
func testAccProviderFactoriesInternal(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return testAccProviderFactoriesInit(providers, []string{ProviderNameAws})
}

// testAccProviderFactoriesAlternate creates ProviderFactories for cross-account and cross-region configurations
//
// For cross-region testing: Typically paired with testAccMultipleRegionPreCheck and testAccAlternateRegionProviderConfig.
//
// For cross-account testing: Typically paired with testAccAlternateAccountPreCheck and testAccAlternateAccountProviderConfig.
func testAccProviderFactoriesAlternate(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return testAccProviderFactoriesInit(providers, []string{
		ProviderNameAws,
		ProviderNameAwsAlternate,
	})
}

// testAccProviderFactoriesAlternateAccountAndAlternateRegion creates ProviderFactories for cross-account and cross-region configurations
//
// Usage typically paired with testAccMultipleRegionPreCheck, testAccAlternateAccountPreCheck,
// and testAccAlternateAccountAndAlternateRegionProviderConfig.
func testAccProviderFactoriesAlternateAccountAndAlternateRegion(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return testAccProviderFactoriesInit(providers, []string{
		ProviderNameAws,
		ProviderNameAwsAlternateAccountAlternateRegion,
		ProviderNameAwsAlternateAccountSameRegion,
		ProviderNameAwsSameAccountAlternateRegion,
	})
}

// testAccProviderFactoriesMultipleRegion creates ProviderFactories for the number of region configurations
//
// Usage typically paired with testAccMultipleRegionPreCheck and testAccMultipleRegionProviderConfig.
func testAccProviderFactoriesMultipleRegion(providers *[]*schema.Provider, regions int) map[string]func() (*schema.Provider, error) {
	providerNames := []string{
		ProviderNameAws,
		ProviderNameAwsAlternate,
	}

	if regions >= 3 {
		providerNames = append(providerNames, ProviderNameAwsThird)
	}

	return testAccProviderFactoriesInit(providers, providerNames)
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

// testAccPreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It allows
// test configurations to omit a provider configuration with region and ensures
// testing functions that attempt to call AWS APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func testAccPreCheck(t *testing.T) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderConfigure.Do(func() {
		if os.Getenv("AWS_PROFILE") == "" && os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI") == "" {
			t.Fatal("AWS_ACCESS_KEY_ID, AWS_PROFILE, or AWS_CONTAINER_CREDENTIALS_FULL_URI must be set for acceptance tests")
		}

		if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
			t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
		}

		// Setting the AWS_DEFAULT_REGION environment variable here allows all tests to omit
		// a provider configuration with a region. This defaults to us-west-2 for provider
		// developer simplicity and has been in the codebase for a very long time.
		//
		// This handling must be preserved until either:
		//   * AWS_DEFAULT_REGION is required and checked above (should mention us-west-2 default)
		//   * Region is automatically handled via shared AWS configuration file and still verified
		region := testAccGetRegion()
		os.Setenv("AWS_DEFAULT_REGION", region)

		err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}

// testAccAwsProviderAccountID returns the account ID of an AWS provider
func testAccAwsProviderAccountID(provider *schema.Provider) string {
	if provider == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: empty provider")
		return ""
	}
	if provider.Meta() == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: unconfigured provider")
		return ""
	}
	client, ok := provider.Meta().(*AWSClient)
	if !ok {
		log.Print("[DEBUG] Unable to read account ID from test provider: non-AWS or unconfigured AWS provider")
		return ""
	}
	return client.accountid
}

// testAccCheckResourceAttrAccountID ensures the Terraform state exactly matches the account ID
func testAccCheckResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(resourceName, attributeName, testAccGetAccountID())(s)
	}
}

// testAccCheckResourceAttrRegionalARN ensures the Terraform state exactly matches a formatted ARN with region
func testAccCheckResourceAttrRegionalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrRegionalARNNoAccount ensures the Terraform state exactly matches a formatted ARN with region but without account ID
func testAccCheckResourceAttrRegionalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrRegionalARNAccountID ensures the Terraform state exactly matches a formatted ARN with region and specific account ID
func testAccCheckResourceAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: accountID,
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrRegionalHostname ensures the Terraform state exactly matches a formatted DNS hostname with region and partition DNS suffix
func testAccCheckResourceAttrRegionalHostname(resourceName, attributeName, serviceName, hostnamePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s.%s", hostnamePrefix, serviceName, testAccGetRegion(), testAccGetPartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// testAccCheckResourceAttrRegionalHostnameService ensures the Terraform state exactly matches a service DNS hostname with region and partition DNS suffix
//
// For example: ec2.us-west-2.amazonaws.com
func testAccCheckResourceAttrRegionalHostnameService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s", serviceName, testAccGetRegion(), testAccGetPartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// testAccCheckResourceAttrRegionalReverseDnsService ensures the Terraform state exactly matches a service reverse DNS hostname with region and partition DNS suffix
//
// For example: com.amazonaws.us-west-2.s3
func testAccCheckResourceAttrRegionalReverseDnsService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reverseDns := fmt.Sprintf("%s.%s.%s", testAccGetPartitionReverseDNSPrefix(), testAccGetRegion(), serviceName)

		return resource.TestCheckResourceAttr(resourceName, attributeName, reverseDns)(s)
	}
}

// testAccCheckResourceAttrHostnameWithPort ensures the Terraform state regexp matches a formatted DNS hostname with prefix, partition DNS suffix, and given port
func testAccCheckResourceAttrHostnameWithPort(resourceName, attributeName, serviceName, hostnamePrefix string, port int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// kafka broker example: "ec2-12-345-678-901.compute-1.amazonaws.com:2345"
		hostname := fmt.Sprintf("%s.%s.%s:%d", hostnamePrefix, serviceName, testAccGetPartitionDNSSuffix(), port)

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// testAccMatchResourceAttrAccountID ensures the Terraform state regexp matches an account ID
func testAccMatchResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(`^\d{12}$`))(s)
	}
}

// testAccMatchResourceAttrRegionalARN ensures the Terraform state regexp matches a formatted ARN with region
func testAccMatchResourceAttrRegionalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccMatchResourceAttrRegionalARNNoAccount ensures the Terraform state regexp matches a formatted ARN with region but without account ID
func testAccMatchResourceAttrRegionalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccMatchResourceAttrRegionalARNAccountID ensures the Terraform state regexp matches a formatted ARN with region and specific account ID
func testAccMatchResourceAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: accountID,
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccMatchResourceAttrRegionalHostname ensures the Terraform state regexp matches a formatted DNS hostname with region and partition DNS suffix
func testAccMatchResourceAttrRegionalHostname(resourceName, attributeName, serviceName string, hostnamePrefixRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostnameRegexpPattern := fmt.Sprintf("%s\\.%s\\.%s\\.%s$", hostnamePrefixRegexp.String(), serviceName, testAccGetRegion(), testAccGetPartitionDNSSuffix())

		hostnameRegexp, err := regexp.Compile(hostnameRegexpPattern)

		if err != nil {
			return fmt.Errorf("Unable to compile hostname regexp (%s): %w", hostnameRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, hostnameRegexp)(s)
	}
}

// testAccCheckResourceAttrGlobalARN ensures the Terraform state exactly matches a formatted ARN without region
func testAccCheckResourceAttrGlobalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrGlobalARNNoAccount ensures the Terraform state exactly matches a formatted ARN without region or account ID
func testAccCheckResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: testAccGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrGlobalARNAccountID ensures the Terraform state exactly matches a formatted ARN without region and with specific account ID
func testAccCheckResourceAttrGlobalARNAccountID(resourceName, attributeName, accountID, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: accountID,
			Partition: testAccGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccMatchResourceAttrGlobalARN ensures the Terraform state regexp matches a formatted ARN without region
func testAccMatchResourceAttrGlobalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount ensures the Terraform state exactly matches a formatted ARN with region without specifying the region or account
func testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: awsAccountIDRegexpInternalPattern,
			Partition: testAccGetPartition(),
			Region:    awsRegionRegexpInternalPattern,
			Resource:  arnResource,
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccMatchResourceAttrGlobalARNNoAccount ensures the Terraform state regexp matches a formatted ARN without region or account ID
func testAccMatchResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: testAccGetPartition(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccCheckResourceAttrRfc3339 ensures the Terraform state matches a RFC3339 value
// This TestCheckFunc will likely be moved to the Terraform Plugin SDK in the future.
func testAccCheckResourceAttrRfc3339(resourceName, attributeName string) resource.TestCheckFunc {
	return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(rfc3339RegexPattern))
}

// testAccCheckListHasSomeElementAttrPair is a TestCheckFunc which validates that the collection on the left has an element with an attribute value
// matching the value on the left
// Based on TestCheckResourceAttrPair from the Terraform SDK testing framework
func testAccCheckListHasSomeElementAttrPair(nameFirst string, resourceAttr string, elementAttr string, nameSecond string, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := primaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := primaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, ok := isSecond.Attributes[keySecond]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", nameSecond, keySecond)
		} else if vSecond == "" {
			return fmt.Errorf("%s: No value was set on attribute %q", nameSecond, keySecond)
		}

		attrsFirst := make([]string, 0, len(isFirst.Attributes))
		attrMatcher := regexp.MustCompile(fmt.Sprintf("%s\\.\\d+\\.%s", resourceAttr, elementAttr))
		for k := range isFirst.Attributes {
			if attrMatcher.MatchString(k) {
				attrsFirst = append(attrsFirst, k)
			}
		}

		found := false
		for _, attrName := range attrsFirst {
			vFirst := isFirst.Attributes[attrName]
			if vFirst == vSecond {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s: No element of %q found with attribute %q matching value %q set on %q of %s", nameFirst, resourceAttr, elementAttr, vSecond, keySecond, nameSecond)
		}

		return nil
	}
}

// testAccCheckResourceAttrEquivalentJSON is a TestCheckFunc that compares a JSON value with an expected value. Both JSON
// values are normalized before being compared.
func testAccCheckResourceAttrEquivalentJSON(resourceName, attributeName, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, resourceName)
		if err != nil {
			return err
		}

		v, ok := is.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", resourceName, attributeName)
		}

		vNormal, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("%s: Error normalizing JSON in %q: %w", resourceName, attributeName, err)
		}

		expectedNormal, err := structure.NormalizeJsonString(expectedJSON)
		if err != nil {
			return fmt.Errorf("Error normalizing expected JSON: %w", err)
		}

		if vNormal != expectedNormal {
			return fmt.Errorf("%s: Attribute %q expected\n%s\ngot\n%s", resourceName, attributeName, expectedJSON, v)
		}
		return nil
	}
}

// Copied and inlined from the SDK testing code
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", name)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s", name)
	}

	return is, nil
}

// testAccGetAccountID returns the account ID of testAccProvider
// Must be used within a resource.TestCheckFunc
func testAccGetAccountID() string {
	return testAccAwsProviderAccountID(testAccProvider)
}

func testAccGetRegion() string {
	v := os.Getenv("AWS_DEFAULT_REGION")
	if v == "" {
		return "us-west-2" // lintignore:AWSAT003
	}
	return v
}

func testAccGetAlternateRegion() string {
	v := os.Getenv("AWS_ALTERNATE_REGION")
	if v == "" {
		return "us-east-1" // lintignore:AWSAT003
	}
	return v
}

func testAccGetThirdRegion() string {
	v := os.Getenv("AWS_THIRD_REGION")
	if v == "" {
		return "us-east-2" // lintignore:AWSAT003
	}
	return v
}

func testAccGetPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func testAccGetPartitionDNSSuffix() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		return partition.DNSSuffix()
	}
	return "amazonaws.com"
}

func testAccGetPartitionReverseDNSPrefix() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		dnsParts := strings.Split(partition.DNSSuffix(), ".")
		sort.Sort(sort.Reverse(sort.StringSlice(dnsParts)))
		return strings.Join(dnsParts, ".")
	}

	return "com.amazonaws"
}

func testAccGetAlternateRegionPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetAlternateRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func testAccGetThirdRegionPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetThirdRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func testAccAlternateAccountPreCheck(t *testing.T) {
	if os.Getenv("AWS_ALTERNATE_PROFILE") == "" && os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID") == "" {
		t.Fatal("AWS_ALTERNATE_ACCESS_KEY_ID or AWS_ALTERNATE_PROFILE must be set for acceptance tests")
	}

	if os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID") != "" && os.Getenv("AWS_ALTERNATE_SECRET_ACCESS_KEY") == "" {
		t.Fatal("AWS_ALTERNATE_SECRET_ACCESS_KEY must be set for acceptance tests")
	}
}

func testAccEC2VPCOnlyPreCheck(t *testing.T) {
	client := testAccProvider.Meta().(*AWSClient)
	platforms := client.supportedplatforms
	region := client.region
	if hasEc2Classic(platforms) {
		t.Skipf("This test can only in regions without EC2 Classic, platforms available in %s: %q",
			region, platforms)
	}
}

func testAccPartitionHasServicePreCheck(serviceId string, t *testing.T) {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		if _, ok := partition.Services()[serviceId]; !ok {
			t.Skip(fmt.Sprintf("skipping tests; partition %s does not support %s service", partition.ID(), serviceId))
		}
	}
}

func testAccMultipleRegionPreCheck(t *testing.T, regions int) {
	if testAccGetRegion() == testAccGetAlternateRegion() {
		t.Fatal("AWS_DEFAULT_REGION and AWS_ALTERNATE_REGION must be set to different values for acceptance tests")
	}

	if testAccGetPartition() != testAccGetAlternateRegionPartition() {
		t.Fatalf("AWS_ALTERNATE_REGION partition (%s) does not match AWS_DEFAULT_REGION partition (%s)", testAccGetAlternateRegionPartition(), testAccGetPartition())
	}

	if regions >= 3 {
		if testAccGetRegion() == testAccGetThirdRegion() {
			t.Fatal("AWS_DEFAULT_REGION and AWS_THIRD_REGION must be set to different values for acceptance tests")
		}

		if testAccGetAlternateRegion() == testAccGetThirdRegion() {
			t.Fatal("AWS_ALTERNATE_REGION and AWS_THIRD_REGION must be set to different values for acceptance tests")
		}

		if testAccGetPartition() != testAccGetThirdRegionPartition() {
			t.Fatalf("AWS_THIRD_REGION partition (%s) does not match AWS_DEFAULT_REGION partition (%s)", testAccGetThirdRegionPartition(), testAccGetPartition())
		}
	}

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		if len(partition.Regions()) < regions {
			t.Skipf("skipping tests; partition includes %d regions, %d expected", len(partition.Regions()), regions)
		}
	}
}

// testAccRegionPreCheck checks that the test region is the specified region.
func testAccRegionPreCheck(t *testing.T, region string) {
	if testAccGetRegion() != region {
		t.Skipf("skipping tests; AWS_DEFAULT_REGION (%s) does not equal %s", testAccGetRegion(), region)
	}
}

// testAccPartitionPreCheck checks that the test partition is the specified partition.
func testAccPartitionPreCheck(partition string, t *testing.T) {
	if testAccGetPartition() != partition {
		t.Skipf("skipping tests; current partition (%s) does not equal %s", testAccGetPartition(), partition)
	}
}

func testAccOrganizationsAccountPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		return
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
	t.Skip("skipping tests; this AWS account must not be an existing member of an AWS Organization")
}

func testAccOrganizationsEnabledPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		t.Skip("this AWS account must be an existing member of an AWS Organization")
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
}

func testAccPreCheckIamServiceLinkedRole(t *testing.T, pathPrefix string) {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	input := &iam.ListRolesInput{
		PathPrefix: aws.String(pathPrefix),
	}

	var role *iam.Role
	err := conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, r := range page.Roles {
			role = r
			break
		}

		return !lastPage
	})

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping tests: %s", err)
	}

	if err != nil {
		t.Fatalf("error listing IAM roles: %s", err)
	}

	if role == nil {
		t.Skipf("skipping tests; missing IAM service-linked role %s. Please create the role and retry", pathPrefix)
	}
}

func testAccEnvironmentVariableSetPreCheck(variable string, t *testing.T) {
	if os.Getenv(variable) == "" {
		t.Skipf("skipping tests; environment variable %s must be set", variable)
	}
}

func testAccAlternateAccountProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[3]q
}
`, os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID"), os.Getenv("AWS_ALTERNATE_PROFILE"), os.Getenv("AWS_ALTERNATE_SECRET_ACCESS_KEY"))
}

func testAccAlternateAccountAlternateRegionProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}
`, os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID"), os.Getenv("AWS_ALTERNATE_PROFILE"), testAccGetAlternateRegion(), os.Getenv("AWS_ALTERNATE_SECRET_ACCESS_KEY"))
}

// When testing needs to distinguish a second region and second account in the same region
// e.g. cross-region functionality with RAM shared subnets
func testAccAlternateAccountAndAlternateRegionProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternateaccountalternateregion" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}

provider "awsalternateaccountsameregion" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[4]q
}

provider "awssameaccountalternateregion" {
  region = %[3]q
}
`, os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID"), os.Getenv("AWS_ALTERNATE_PROFILE"), testAccGetAlternateRegion(), os.Getenv("AWS_ALTERNATE_SECRET_ACCESS_KEY"))
}

// Deprecated: Use testAccMultipleRegionProviderConfig instead
func testAccAlternateRegionProviderConfig() string {
	return testAccNamedRegionalProviderConfig(ProviderNameAwsAlternate, testAccGetAlternateRegion())
}

func testAccMultipleRegionProviderConfig(regions int) string {
	var config strings.Builder

	config.WriteString(testAccNamedRegionalProviderConfig(ProviderNameAwsAlternate, testAccGetAlternateRegion()))

	if regions >= 3 {
		config.WriteString(testAccNamedRegionalProviderConfig(ProviderNameAwsThird, testAccGetThirdRegion()))
	}

	return config.String()
}

func testAccProviderConfigIgnoreTagsKeyPrefixes1(keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }
}
`, keyPrefix1)
}

func testAccProviderConfigIgnoreTagsKeys1(key1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1)
}

// testAccNamedRegionalProviderConfig creates a new provider named configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func testAccNamedRegionalProviderConfig(providerName string, region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  region = %[2]q
}
`, providerName, region)
}

// testAccRegionalProviderConfig creates a new provider configuration with a region.
//
// This can only be used for single provider configuration testing as it
// overwrites the "aws" provider configuration.
func testAccRegionalProviderConfig(region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  region = %[1]q
}
`, region)
}

func testAccAwsRegionProviderFunc(region string, providers *[]*schema.Provider) func() *schema.Provider {
	return func() *schema.Provider {
		if region == "" {
			log.Println("[DEBUG] No region given")
			return nil
		}
		if providers == nil {
			log.Println("[DEBUG] No providers given")
			return nil
		}

		log.Printf("[DEBUG] Checking providers for AWS region: %s", region)
		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider == nil || provider.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider")
				continue
			}

			// Ignore if Meta is not AWSClient, this will happen for other providers
			client, ok := provider.Meta().(*AWSClient)
			if !ok {
				log.Printf("[DEBUG] Skipping non-AWS provider")
				continue
			}

			clientRegion := client.region
			log.Printf("[DEBUG] Checking AWS provider region %q against %q", clientRegion, region)
			if clientRegion == region {
				log.Printf("[DEBUG] Found AWS provider with region: %s", region)
				return provider
			}
		}

		log.Printf("[DEBUG] No suitable provider found for %q in %d providers", region, len(*providers))
		return nil
	}
}

func testAccCheckResourceDisappears(provider *schema.Provider, resource *schema.Resource, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if resourceState.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", resourceName)
		}

		if resource.DeleteContext != nil {
			diags := resource.DeleteContext(context.Background(), resource.Data(resourceState.Primary), provider.Meta())

			for i := range diags {
				if diags[i].Severity == diag.Error {
					return fmt.Errorf("error deleting resource: %s", diags[i].Summary)
				}
			}

			return nil
		}

		return resource.Delete(resource.Data(resourceState.Primary), provider.Meta())
	}
}

func testAccCheckWithProviders(f func(*terraform.State, *schema.Provider) error, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		numberOfProviders := len(*providers)
		for i, provider := range *providers {
			if provider.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider %d (total: %d)", i, numberOfProviders)
				continue
			}
			log.Printf("[DEBUG] Calling check with provider %d (total: %d)", i, numberOfProviders)
			if err := f(s, provider); err != nil {
				return err
			}
		}
		return nil
	}
}

// testAccErrorCheckSkipMessagesContaining skips tests based on error messages that indicate unsupported features
func testAccErrorCheckSkipMessagesContaining(t *testing.T, messages ...string) resource.ErrorCheckFunc {
	return func(err error) error {
		if err == nil {
			return err
		}

		for _, message := range messages {
			if strings.Contains(err.Error(), message) {
				t.Skipf("skipping test for %s/%s: %s", testAccGetPartition(), testAccGetRegion(), err.Error())
			}
		}

		return err
	}
}

// Check service API call error for reasons to skip acceptance testing
// These include missing API endpoints and unsupported API calls
func testAccPreCheckSkipError(err error) bool {
	// GovCloud has endpoints that respond with (no message provided after the error code):
	// AccessDeniedException:
	// Ignore these API endpoints that exist but are not officially enabled
	if isAWSErr(err, "AccessDeniedException", "") {
		return true
	}
	// Ignore missing API endpoints
	if isAWSErr(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if isAWSErr(err, "UnknownOperationException", "") {
		return true
	}
	if isAWSErr(err, "UnsupportedOperation", "") {
		return true
	}
	if isAWSErr(err, "InvalidInputException", "Unknown operation") {
		return true
	}
	if isAWSErr(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	return false
}

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func testSweepSkipSweepError(err error) bool {
	// Ignore missing API endpoints
	if isAWSErr(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if isAWSErr(err, "UnsupportedOperation", "") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if isAWSErr(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if isAWSErr(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// GovCloud has endpoints that respond with (no message provided):
	// AccessDeniedException:
	// Since acceptance test sweepers are best effort and this response is very common,
	// we allow bypassing this error globally instead of individual test sweeper fixes.
	if isAWSErr(err, "AccessDeniedException", "") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if isAWSErr(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: The action DescribeTransitGatewayAttachments is not valid for this web service
	if isAWSErr(err, "InvalidAction", "is not valid") {
		return true
	}
	// For example from GovCloud SES.SetActiveReceiptRuleSet.
	if isAWSErr(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	return false
}

// Check sweeper API call error for reasons to skip a specific resource
// These include AccessDenied or AccessDeniedException for individual resources, e.g. managed by central IT
func testSweepSkipResourceError(err error) bool {
	// Since acceptance test sweepers are best effort, we allow bypassing this error globally
	// instead of individual test sweeper fixes.
	return tfawserr.ErrCodeContains(err, "AccessDenied")
}

func TestAccAWSProvider_Endpoints(t *testing.T) {
	var providers []*schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, endpointServiceName := range endpointServiceNames {
		endpoints.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", endpointServiceName, endpointServiceName))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigEndpoints(endpoints.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderEndpoints(&providers),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_EmptyConfigurationBlock(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeys(&providers, []string{}),
					testAccCheckAWSProviderIgnoreTagsKeyPrefixes(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_KeyPrefixes_None(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsKeyPrefixes0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeyPrefixes(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_KeyPrefixes_One(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsKeyPrefixes1("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeyPrefixes(&providers, []string{"test"}),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_KeyPrefixes_Multiple(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsKeyPrefixes2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeyPrefixes(&providers, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_Keys_None(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsKeys0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeys(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_Keys_One(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsKeys1("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeys(&providers, []string{"test"}),
				),
			},
		},
	})
}

func TestAccAWSProvider_IgnoreTags_Keys_Multiple(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigIgnoreTagsKeys2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderIgnoreTagsKeys(&providers, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccAWSProvider_Region_AwsChina(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigRegion("cn-northwest-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderDnsSuffix(&providers, "amazonaws.com.cn"),
					testAccCheckAWSProviderPartition(&providers, "aws-cn"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSProvider_Region_AwsCommercial(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigRegion("us-west-2"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderDnsSuffix(&providers, "amazonaws.com"),
					testAccCheckAWSProviderPartition(&providers, "aws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSProvider_Region_AwsGovCloudUs(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigRegion("us-gov-west-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderDnsSuffix(&providers, "amazonaws.com"),
					testAccCheckAWSProviderPartition(&providers, "aws-us-gov"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSProvider_AssumeRole_Empty(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSProviderConfigAssumeRoleEmpty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCallerIdentityAccountId("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

func testAccCheckAWSProviderDnsSuffix(providers *[]*schema.Provider, expectedDnsSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerDnsSuffix := provider.Meta().(*AWSClient).dnsSuffix

			if providerDnsSuffix != expectedDnsSuffix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedDnsSuffix, providerDnsSuffix)
			}
		}

		return nil
	}
}

func testAccCheckAWSProviderEndpoints(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		// Match AWSClient struct field names to endpoint configuration names
		endpointFieldNameF := func(endpoint string) func(string) bool {
			return func(name string) bool {
				switch endpoint {
				case "applicationautoscaling":
					endpoint = "appautoscaling"
				case "budgets":
					endpoint = "budget"
				case "cloudformation":
					endpoint = "cf"
				case "cloudhsm":
					endpoint = "cloudhsmv2"
				case "cognitoidentity":
					endpoint = "cognito"
				case "configservice":
					endpoint = "config"
				case "cur":
					endpoint = "costandusagereport"
				case "directconnect":
					endpoint = "dx"
				case "lexmodels":
					endpoint = "lexmodel"
				case "route53":
					endpoint = "r53"
				case "sdb":
					endpoint = "simpledb"
				case "serverlessrepo":
					endpoint = "serverlessapplicationrepository"
				case "servicecatalog":
					endpoint = "sc"
				case "servicediscovery":
					endpoint = "sd"
				case "stepfunctions":
					endpoint = "sfn"
				}

				return name == fmt.Sprintf("%sconn", endpoint)
			}
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)

			for _, endpointServiceName := range endpointServiceNames {
				providerClientField := reflect.Indirect(reflect.ValueOf(providerClient)).FieldByNameFunc(endpointFieldNameF(endpointServiceName))

				if !providerClientField.IsValid() {
					return fmt.Errorf("unable to match AWSClient struct field name for endpoint name: %s", endpointServiceName)
				}

				actualEndpoint := reflect.Indirect(reflect.Indirect(providerClientField).FieldByName("Config").FieldByName("Endpoint")).String()
				expectedEndpoint := fmt.Sprintf("http://%s", endpointServiceName)

				if actualEndpoint != expectedEndpoint {
					return fmt.Errorf("expected endpoint (%s) value (%s), got: %s", endpointServiceName, expectedEndpoint, actualEndpoint)
				}
			}
		}

		return nil
	}
}

func testAccCheckAWSProviderIgnoreTagsKeyPrefixes(providers *[]*schema.Provider, expectedKeyPrefixes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)
			ignoreTagsConfig := providerClient.IgnoreTagsConfig

			if ignoreTagsConfig == nil || ignoreTagsConfig.KeyPrefixes == nil {
				if len(expectedKeyPrefixes) != 0 {
					return fmt.Errorf("expected key_prefixes (%d) length, got: 0", len(expectedKeyPrefixes))
				}

				continue
			}

			actualKeyPrefixes := ignoreTagsConfig.KeyPrefixes.Keys()

			if len(actualKeyPrefixes) != len(expectedKeyPrefixes) {
				return fmt.Errorf("expected key_prefixes (%d) length, got: %d", len(expectedKeyPrefixes), len(actualKeyPrefixes))
			}

			for _, expectedElement := range expectedKeyPrefixes {
				var found bool

				for _, actualElement := range actualKeyPrefixes {
					if actualElement == expectedElement {
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("expected key_prefixes element, but was missing: %s", expectedElement)
				}
			}

			for _, actualElement := range actualKeyPrefixes {
				var found bool

				for _, expectedElement := range expectedKeyPrefixes {
					if actualElement == expectedElement {
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("unexpected key_prefixes element: %s", actualElement)
				}
			}
		}

		return nil
	}
}

func testAccCheckAWSProviderIgnoreTagsKeys(providers *[]*schema.Provider, expectedKeys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)
			ignoreTagsConfig := providerClient.IgnoreTagsConfig

			if ignoreTagsConfig == nil || ignoreTagsConfig.Keys == nil {
				if len(expectedKeys) != 0 {
					return fmt.Errorf("expected keys (%d) length, got: 0", len(expectedKeys))
				}

				continue
			}

			actualKeys := ignoreTagsConfig.Keys.Keys()

			if len(actualKeys) != len(expectedKeys) {
				return fmt.Errorf("expected keys (%d) length, got: %d", len(expectedKeys), len(actualKeys))
			}

			for _, expectedElement := range expectedKeys {
				var found bool

				for _, actualElement := range actualKeys {
					if actualElement == expectedElement {
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("expected keys element, but was missing: %s", expectedElement)
				}
			}

			for _, actualElement := range actualKeys {
				var found bool

				for _, expectedElement := range expectedKeys {
					if actualElement == expectedElement {
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("unexpected keys element: %s", actualElement)
				}
			}
		}

		return nil
	}
}

func testAccCheckAWSProviderPartition(providers *[]*schema.Provider, expectedPartition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerPartition := provider.Meta().(*AWSClient).partition

			if providerPartition != expectedPartition {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedPartition, providerPartition)
			}
		}

		return nil
	}
}

// testAccPreCheckEc2ClassicOrHasDefaultVpcWithDefaultSubnets checks that the test region has either
// - The EC2-Classic platform available, or
// - A default VPC with default subnets.
// This check is useful to ensure that an instance can be launched without specifying a subnet.
func testAccPreCheckEc2ClassicOrHasDefaultVpcWithDefaultSubnets(t *testing.T) {
	client := testAccProvider.Meta().(*AWSClient)

	if !hasEc2Classic(client.supportedplatforms) && !(testAccHasDefaultVpc(t) && testAccDefaultSubnetCount(t) > 0) {
		t.Skipf("skipping tests; %s does not have EC2-Classic or a default VPC with default subnets", client.region)
	}
}

// testAccHasDefaultVpc returns whether the current AWS region has a default VPC.
func testAccHasDefaultVpc(t *testing.T) bool {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	resp, err := conn.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{
		AttributeNames: aws.StringSlice([]string{ec2.AccountAttributeNameDefaultVpc}),
	})
	if testAccPreCheckSkipError(err) ||
		len(resp.AccountAttributes) == 0 ||
		len(resp.AccountAttributes[0].AttributeValues) == 0 ||
		aws.StringValue(resp.AccountAttributes[0].AttributeValues[0].AttributeValue) == "none" {
		return false
	}
	if err != nil {
		t.Fatalf("error describing EC2 account attributes: %s", err)
	}

	return true
}

// testAccDefaultSubnetCount returns the number of default subnets in the current region's default VPC.
func testAccDefaultSubnetCount(t *testing.T) int {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeSubnetsInput{
		Filters: buildEC2AttributeFilterList(map[string]string{
			"defaultForAz": "true",
		}),
	}
	output, err := conn.DescribeSubnets(input)
	if testAccPreCheckSkipError(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("error describing default subnets: %s", err)
	}

	return len(output.Subnets)
}

func testAccAWSProviderConfigEndpoints(endpoints string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    %[1]s
  }
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`, endpoints)
}

func testAccAWSProviderConfigIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return `
provider "aws" {
  ignore_tags {}

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`
}

func testAccAWSProviderConfigIgnoreTagsKeyPrefixes0() string {
	//lintignore:AT004
	return `
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`
}

func testAccAWSProviderConfigIgnoreTagsKeyPrefixes1(tagPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`, tagPrefix1)
}

func testAccAWSProviderConfigIgnoreTagsKeyPrefixes2(tagPrefix1, tagPrefix2 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q, %[2]q]
  }

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`, tagPrefix1, tagPrefix2)
}

func testAccAWSProviderConfigIgnoreTagsKeys0() string {
	//lintignore:AT004
	return `
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`
}

func testAccAWSProviderConfigIgnoreTagsKeys1(tag1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`, tag1)
}

func testAccAWSProviderConfigIgnoreTagsKeys2(tag1, tag2 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q, %[2]q]
  }

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`, tag1, tag2)
}

func testAccAWSProviderConfigRegion(region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  region                      = %[1]q
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`, region)
}

func testAccAssumeRoleARNPreCheck(t *testing.T) {
	v := os.Getenv("TF_ACC_ASSUME_ROLE_ARN")
	if v == "" {
		t.Skip("skipping tests; TF_ACC_ASSUME_ROLE_ARN must be set")
	}
}

func testAccProviderConfigAssumeRolePolicy(policy string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn = %q
    policy   = %q
  }
}
`, os.Getenv("TF_ACC_ASSUME_ROLE_ARN"), policy)
}

const testAccCheckAWSProviderConfigAssumeRoleEmpty = `
provider "aws" {
  assume_role {
  }
}

data "aws_caller_identity" "current" {}
` //lintignore:AT004

// composeConfig can be called to concatenate multiple strings to build test configurations
func composeConfig(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}
