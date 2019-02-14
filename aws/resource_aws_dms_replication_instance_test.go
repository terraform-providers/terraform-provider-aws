package aws

import (
	"fmt"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDmsReplicationInstance_Basic(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_ReplicationInstanceClass(rName, "dms.t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "50"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_private_ips.#", "1"),
					// ARN resource is its own unique identifier
					resource.TestCheckResourceAttrSet(resourceName, "replication_instance_arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", "dms.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_AllocatedStorage(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_AllocatedStorage(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_AllocatedStorage(rName, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allocated_storage", "6"),
				),
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_AutoMinorVersionUpgrade(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_AutoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_AutoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_AutoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_AvailabilityZone(t *testing.T) {
	dataSourceName := "data.aws_availability_zones.available"
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_AvailabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "names.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_EngineVersion(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	// This acceptance test is designed to test engine version upgrades.
	// Over time, DMS replication instance engine versions are deprecated
	// so they will eventually error on resource creation, e.g.
	//   InvalidParameterValueException: No replication engine found with version: 2.4.2
	// During the PreCheck, we will find candidate engine versions from the
	// orderable replication instances and generate the TestStep.
	// We prefer this method over creating a plural data source that
	// seems impractical for real world usage.
	testSteps := []resource.TestStep{
		{},
		{},
		{
			ResourceName:            resourceName,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"apply_immediately"},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)

			engineVersions := testAccAWSDmsReplicationInstanceEngineVersionsPreCheck(t)

			testSteps[0] = resource.TestStep{
				Config: testAccAWSDmsReplicationInstanceConfig_EngineVersion(rName, engineVersions[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersions[0]),
				),
			}
			testSteps[1] = resource.TestStep{
				Config: testAccAWSDmsReplicationInstanceConfig_EngineVersion(rName, engineVersions[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_version", engineVersions[1]),
				),
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps:        testSteps,
	})
}

func TestAccAWSDmsReplicationInstance_KmsKeyArn(t *testing.T) {
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_KmsKeyArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_MultiAz(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_MultiAz(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_MultiAz(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "false"),
				),
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_MultiAz(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "multi_az", "true"),
				),
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_PreferredMaintenanceWindow(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_PreferredMaintenanceWindow(rName, "sun:00:30-sun:02:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "sun:00:30-sun:02:30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_PreferredMaintenanceWindow(rName, "mon:00:30-mon:02:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "preferred_maintenance_window", "mon:00:30-mon:02:30"),
				),
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_PubliclyAccessible(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_PubliclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_public_ips.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_ReplicationInstanceClass(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_ReplicationInstanceClass(rName, "dms.t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", "dms.t2.micro"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_ReplicationInstanceClass(rName, "dms.t2.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_instance_class", "dms.t2.small"),
				),
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_ReplicationSubnetGroupId(t *testing.T) {
	dmsReplicationSubnetGroupResourceName := "aws_dms_replication_subnet_group.test"
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_ReplicationSubnetGroupId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "replication_subnet_group_id", dmsReplicationSubnetGroupResourceName, "replication_subnet_group_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_Tags(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_Tags_One(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_Tags_Two(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDmsReplicationInstanceConfig_Tags_One(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDmsReplicationInstance_VpcSecurityGroupIds(t *testing.T) {
	resourceName := "aws_dms_replication_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsReplicationInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsReplicationInstanceConfig_VpcSecurityGroupIds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDmsReplicationInstanceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func testAccCheckAWSDmsReplicationInstanceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeReplicationInstances(&dms.DescribeReplicationInstancesInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-instance-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("DMS replication instance error: %v", err)
		}
		if resp == nil || len(resp.ReplicationInstances) == 0 {
			return fmt.Errorf("DMS replication instance not found")
		}

		return nil
	}
}

func testAccCheckAWSDmsReplicationInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_replication_instance" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn

		resp, err := conn.DescribeReplicationInstances(&dms.DescribeReplicationInstancesInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-instance-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if isAWSErr(err, dms.ErrCodeResourceNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			for _, replicationInstance := range resp.ReplicationInstances {
				if aws.StringValue(replicationInstance.ReplicationInstanceIdentifier) == rs.Primary.ID {
					return fmt.Errorf("DMS Replication Instance (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}

	return nil
}

// Ensure at least two engine versions of the replication instance class are available
func testAccAWSDmsReplicationInstanceEngineVersionsPreCheck(t *testing.T) []string {
	conn := testAccProvider.Meta().(*AWSClient).dmsconn

	// Gather all orderable DMS replication instances of the instance class
	// used in the acceptance testing. Not currently available as an input
	// parameter to the describe API call.
	var orderableReplicationInstances []*dms.OrderableReplicationInstance
	input := &dms.DescribeOrderableReplicationInstancesInput{}
	replicationInstanceClass := "dms.t2.micro"

	err := conn.DescribeOrderableReplicationInstancesPages(input, func(output *dms.DescribeOrderableReplicationInstancesOutput, lastPage bool) bool {
		for _, orderableReplicationInstance := range output.OrderableReplicationInstances {
			if orderableReplicationInstance == nil {
				continue
			}

			if aws.StringValue(orderableReplicationInstance.ReplicationInstanceClass) == replicationInstanceClass {
				orderableReplicationInstances = append(orderableReplicationInstances, orderableReplicationInstance)
			}
		}

		return !lastPage
	})

	if err != nil {
		t.Fatalf("error describing DMS orderable replication instances: %s", err)
	}

	// Ensure we have enough
	if len(orderableReplicationInstances) < 2 {
		t.Fatalf("found (%d) DMS orderable replication instances for instance class (%s), need at least 2", len(orderableReplicationInstances), replicationInstanceClass)
	}

	// Sort them ascending
	sort.Slice(orderableReplicationInstances, func(i, j int) bool {
		iEngineVersion, err := gversion.NewVersion(aws.StringValue(orderableReplicationInstances[i].EngineVersion))

		if err != nil {
			t.Fatalf("error converting (%s) to go-version: %s", aws.StringValue(orderableReplicationInstances[i].EngineVersion), err)
		}

		jEngineVersion, err := gversion.NewVersion(aws.StringValue(orderableReplicationInstances[j].EngineVersion))

		if err != nil {
			t.Fatalf("error converting (%s) to go-version: %s", aws.StringValue(orderableReplicationInstances[j].EngineVersion), err)
		}

		return iEngineVersion.LessThan(jEngineVersion)
	})

	engineVersions := make([]string, len(orderableReplicationInstances))

	for i, orderableReplicationInstance := range orderableReplicationInstances {
		engineVersions[i] = aws.StringValue(orderableReplicationInstance.EngineVersion)
	}

	return engineVersions
}

func testAccAWSDmsReplicationInstanceConfig_AllocatedStorage(rName string, allocatedStorage int) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  allocated_storage          = %d
  apply_immediately          = true
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, allocatedStorage, rName)
}

func testAccAWSDmsReplicationInstanceConfig_AutoMinorVersionUpgrade(rName string, autoMinorVersionUpgrade bool) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  auto_minor_version_upgrade = %t
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, autoMinorVersionUpgrade, rName)
}

func testAccAWSDmsReplicationInstanceConfig_AvailabilityZone(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  availability_zone          = "${data.aws_availability_zones.available.names[0]}"
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, rName)
}

func testAccAWSDmsReplicationInstanceConfig_EngineVersion(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  engine_version             = %q
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, engineVersion, rName)
}

func testAccAWSDmsReplicationInstanceConfig_KmsKeyArn(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  kms_key_arn                = "${aws_kms_key.test.arn}"
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, rName)
}

func testAccAWSDmsReplicationInstanceConfig_MultiAz(rName string, multiAz bool) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  multi_az                   = %t
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, multiAz, rName)
}

func testAccAWSDmsReplicationInstanceConfig_PreferredMaintenanceWindow(rName, preferredMaintenanceWindow string) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately            = true
  preferred_maintenance_window = %q
  replication_instance_class   = "dms.t2.micro"
  replication_instance_id      = %q
}
`, preferredMaintenanceWindow, rName)
}

func testAccAWSDmsReplicationInstanceConfig_PubliclyAccessible(rName string, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately          = true
  publicly_accessible        = %t
  replication_instance_class = "dms.t2.micro"
  replication_instance_id    = %q
}
`, publiclyAccessible, rName)
}

func testAccAWSDmsReplicationInstanceConfig_ReplicationInstanceClass(rName, replicationInstanceClass string) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately            = true
  replication_instance_class   = %q
  replication_instance_id      = %q
}
`, replicationInstanceClass, rName)
}

func testAccAWSDmsReplicationInstanceConfig_ReplicationSubnetGroupId(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "${aws_vpc.test.tags["Name"]}"
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = %q
  replication_subnet_group_id          = %q
  subnet_ids                           = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = "dms.t2.micro"
  replication_instance_id     = %q
  replication_subnet_group_id = "${aws_dms_replication_subnet_group.test.replication_subnet_group_id}"
}
`, rName, rName, rName, rName)
}

func testAccAWSDmsReplicationInstanceConfig_Tags_One(rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately            = true
  replication_instance_class   = "dms.t2.micro"
  replication_instance_id      = %q

  tags = {
    %q = %q
  }
}
`, rName, key1, value1)
}

func testAccAWSDmsReplicationInstanceConfig_Tags_Two(rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_dms_replication_instance" "test" {
  apply_immediately            = true
  replication_instance_class   = "dms.t2.micro"
  replication_instance_id      = %q

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, key1, value1, key2, value2)
}

func testAccAWSDmsReplicationInstanceConfig_VpcSecurityGroupIds(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_security_group" "test" {
  name   = "${aws_vpc.test.tags["Name"]}"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "${aws_vpc.test.tags["Name"]}"
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = %q
  replication_subnet_group_id          = %q
  subnet_ids                           = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = "dms.t2.micro"
  replication_instance_id     = %q
  replication_subnet_group_id = "${aws_dms_replication_subnet_group.test.replication_subnet_group_id}"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
}
`, rName, rName, rName, rName)
}
