package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_networkmanager_site", &resource.Sweeper{
		Name: "aws_networkmanager_site",
		F:    testSweepNetworkManagerSite,
	})
}

func testSweepNetworkManagerSite(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkmanagerconn
	var sweeperErrs *multierror.Error

	err = conn.GetSitesPages(&networkmanager.GetSitesInput{},
		func(page *networkmanager.GetSitesOutput, lastPage bool) bool {
			for _, site := range page.Sites {
				input := &networkmanager.DeleteSiteInput{
					GlobalNetworkId: site.GlobalNetworkId,
					SiteId:          site.SiteId,
				}
				id := aws.StringValue(site.SiteId)
				globalNetworkID := aws.StringValue(site.GlobalNetworkId)

				log.Printf("[INFO] Deleting Network Manager Site: %s", id)
				_, err := conn.DeleteSite(input)

				if isAWSErr(err, "InvalidSiteID.NotFound", "") {
					continue
				}

				if err != nil {
					sweeperErr := fmt.Errorf("failed to delete Network Manager Site %s: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := waitForNetworkManagerSiteDeletion(conn, globalNetworkID, id); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Network Manager Site (%s) deletion: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Site sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Network Manager Sites: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSNetworkManagerSite_basic(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	gloablNetworkResourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkManagerSiteConfig("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerSiteExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSNetworkManagerSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkManagerSiteConfig_Update("test updated", "18.0029784", "-76.7897987"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerSiteExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "test updated"),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", ""),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "18.0029784"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "-76.7897987"),
				),
			},
		},
	})
}

func TestAccAWSNetworkManagerSite_tags(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkManagerSiteConfigTags1(description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSNetworkManagerSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkManagerSiteConfigTags2(description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNetworkManagerSiteConfigTags1(description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsNetworkManagerSiteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_site" {
			continue
		}

		site, err := networkmanagerDescribeSite(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)
		if err != nil {
			if isAWSErr(err, networkmanager.ErrCodeValidationException, "") {
				return nil
			}
			return err
		}

		if site == nil {
			continue
		}

		return fmt.Errorf("Expected Site to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsNetworkManagerSiteExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

		site, err := networkmanagerDescribeSite(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		if site == nil {
			return fmt.Errorf("Network Manager Site not found")
		}

		if aws.StringValue(site.State) != networkmanager.SiteStateAvailable && aws.StringValue(site.State) != networkmanager.SiteStatePending {
			return fmt.Errorf("Network Manager Site (%s) exists in (%s) state", rs.Primary.ID, aws.StringValue(site.State))
		}

		return err
	}
}

func testAccNetworkManagerSiteConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id
}
`, description)
}

func testAccNetworkManagerSiteConfigTags1(description, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id

  tags = {
  	%q = %q
  }
}
`, description, tagKey1, tagValue1)
}

func testAccNetworkManagerSiteConfigTags2(description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id

  tags = {
  	%q = %q
	%q = %q
  }
}
`, description, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccNetworkManagerSiteConfig_Update(description, latitude, longitude string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = %q
 global_network_id = aws_networkmanager_global_network.test.id

 location {
  latitude  = %q	
  longitude = %q
 }
}
`, description, latitude, longitude)
}

func testAccAWSNetworkManagerSiteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
