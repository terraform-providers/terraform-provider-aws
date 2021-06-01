package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSEIP_Filter(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_Id(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigId,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}
func TestAccDataSourceAWSEIP_PublicIP_EC2Classic(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigPublicIpEc2Classic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_PublicIP_VPC(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigPublicIpVpc,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_Tags(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_NetworkInterface(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigNetworkInterface,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_dns", resourceName, "private_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_Instance(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigInstance,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_id", resourceName, "instance"),
					resource.TestCheckResourceAttrPair(dataSourceName, "association_id", resourceName, "association_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_CarrierIP(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSWavelengthZoneAvailable(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigCarrierIP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "carrier_ip", resourceName, "carrier_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSEIP_CustomerOwnedIpv4Pool(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSEIPConfigCustomerOwnedIpv4Pool(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ipv4_pool", dataSourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ip", dataSourceName, "customer_owned_ip"),
				),
			},
		},
	})
}

func testAccDataSourceAWSEIPConfigCustomerOwnedIpv4Pool() string {
	return `
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  vpc                      = true
}

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`
}

func testAccDataSourceAWSEIPConfigFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eip" "test" {
  filter {
    name   = "tag:Name"
    values = [aws_eip.test.tags.Name]
  }
}
`, rName)
}

const testAccDataSourceAWSEIPConfigId = `
resource "aws_eip" "test" {
  vpc = true
}

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`

func testAccDataSourceAWSEIPConfigPublicIpEc2Classic() string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		`
resource "aws_eip" "test" {}

data "aws_eip" "test" {
  public_ip = aws_eip.test.public_ip
}
`)
}

const testAccDataSourceAWSEIPConfigPublicIpVpc = `
resource "aws_eip" "test" {
  vpc = true
}

data "aws_eip" "test" {
  public_ip = aws_eip.test.public_ip
}
`

func testAccDataSourceAWSEIPConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eip" "test" {
  tags = {
    Name = aws_eip.test.tags["Name"]
  }
}
`, rName)
}

const testAccDataSourceAWSEIPConfigNetworkInterface = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_eip" "test" {
  vpc               = true
  network_interface = aws_network_interface.test.id
}

data "aws_eip" "test" {
  filter {
    name   = "network-interface-id"
    values = [aws_eip.test.network_interface]
  }
}
`

var testAccDataSourceAWSEIPConfigInstance = composeConfig(
	testAccAvailableAZsNoOptInDefaultExcludeConfig(), `
resource "aws_vpc" "test" {
  cidr_block = "10.2.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.2.0.0/24"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

data "aws_ami" "test" {
  most_recent = true
  name_regex  = "^amzn-ami.*ecs-optimized$"

  owners = [
    "amazon",
  ]
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.test.id
  subnet_id     = aws_subnet.test.id
  instance_type = "t2.micro"
}

resource "aws_eip" "test" {
  vpc      = true
  instance = aws_instance.test.id
}

data "aws_eip" "test" {
  filter {
    name   = "instance-id"
    values = [aws_eip.test.instance]
  }
}
`)

func testAccDataSourceAWSEIPConfigCarrierIP(rName string) string {
	return composeConfig(
		testAccAvailableAZsWavelengthZonesDefaultExcludeConfig(),
		fmt.Sprintf(`
data "aws_availability_zone" "available" {
  name = data.aws_availability_zones.available.names[0]
}

resource "aws_eip" "test" {
  vpc                  = true
  network_border_group = data.aws_availability_zone.available.network_border_group

  tags = {
    Name = %[1]q
  }
}

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`, rName))
}
