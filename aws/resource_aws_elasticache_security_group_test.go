package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_security_group", &resource.Sweeper{
		Name: "aws_elasticache_security_group",
		F:    testSweepElasticacheCacheSecurityGroups,
		Dependencies: []string{
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
		},
	})
}

func testSweepElasticacheCacheSecurityGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).elasticacheconn

	err = conn.DescribeCacheSecurityGroupsPages(&elasticache.DescribeCacheSecurityGroupsInput{}, func(page *elasticache.DescribeCacheSecurityGroupsOutput, isLast bool) bool {
		if len(page.CacheSecurityGroups) == 0 {
			log.Print("[DEBUG] No Elasticache Cache Security Groups to sweep")
			return false
		}

		for _, securityGroup := range page.CacheSecurityGroups {
			name := aws.StringValue(securityGroup.CacheSecurityGroupName)

			if name == "default" {
				log.Printf("[INFO] Skipping Elasticache Cache Security Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Elasticache Cache Security Group: %s", name)
			_, err := conn.DeleteCacheSecurityGroup(&elasticache.DeleteCacheSecurityGroupInput{
				CacheSecurityGroupName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Cache Security Group (%s): %s", name, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Cache Security Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Cache Security Groups: %s", err)
	}
	return nil
}

func TestAccAWSElasticacheSecurityGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSElasticacheSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheSecurityGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
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

func testAccCheckAWSElasticacheSecurityGroupDestroy(s *terraform.State) error {
	conn := testAccProviderEc2Classic.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_security_group" {
			continue
		}
		res, err := conn.DescribeCacheSecurityGroups(&elasticache.DescribeCacheSecurityGroupsInput{
			CacheSecurityGroupName: aws.String(rs.Primary.ID),
		})
		if awserr, ok := err.(awserr.Error); ok && awserr.Code() == "CacheSecurityGroupNotFound" {
			continue
		}

		if len(res.CacheSecurityGroups) > 0 {
			return fmt.Errorf("cache security group still exists")
		}
		return err
	}
	return nil
}

func testAccCheckAWSElasticacheSecurityGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache security group ID is set")
		}

		conn := testAccProviderEc2Classic.Meta().(*AWSClient).elasticacheconn
		_, err := conn.DescribeCacheSecurityGroups(&elasticache.DescribeCacheSecurityGroupsInput{
			CacheSecurityGroupName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("CacheSecurityGroup error: %v", err)
		}
		return nil
	}
}

func testAccAWSElasticacheSecurityGroupConfig(rName string) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_security_group" "test" {
  name                 = %[1]q
  security_group_names = [aws_security_group.test.name]
}
`, rName))
}
