package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/waiter"
)

func init() {
	resource.AddTestSweepers("aws_eks_addon", &resource.Sweeper{
		Name: "aws_eks_addon",
		F:    testSweepEksAddon,
	})
}

func testSweepEksAddon(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).eksconn
	ctx := context.TODO()
	var sweeperErrs *multierror.Error

	input := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	err = conn.ListClustersPagesWithContext(ctx, input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			clusterName := aws.StringValue(cluster)
			input := &eks.ListAddonsInput{
				ClusterName: aws.String(clusterName),
			}
			err := conn.ListAddonsPagesWithContext(ctx, input, func(page *eks.ListAddonsOutput, lastPage bool) bool {
				for _, addon := range page.Addons {
					addonName := aws.StringValue(addon)
					log.Printf("[INFO] Deleting EKS Addon %s from Cluster %s", addonName, clusterName)
					input := &eks.DeleteAddonInput{
						AddonName:   aws.String(addonName),
						ClusterName: aws.String(clusterName),
					}

					_, err := conn.DeleteAddonWithContext(ctx, input)

					if err != nil && !tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting EKS Addon %s from Cluster %s: %w", addonName, clusterName, err))
						continue
					}

					if _, err := waiter.EksAddonDeleted(ctx, conn, clusterName, addonName); err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error waiting for EKS Addon %s deletion: %w", addonName, err))
						continue
					}
				}
				return true
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Addons for Cluster %s: %w", clusterName, err))
			}
		}

		return true
	})
	if testSweepSkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping EKS Addon sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving EKS Clusters: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSEksAddon_basic(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_eks_cluster.test"
	addonResourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddon_Basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, addonResourceName, &addon),
					testAccMatchResourceAttrRegionalARN(addonResourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("addon/%s/%s/.+$", rName, addonName))),
					resource.TestCheckResourceAttrPair(addonResourceName, "cluster_name", clusterResourceName, "name"),
					resource.TestCheckResourceAttr(addonResourceName, "addon_name", addonName),
					resource.TestCheckResourceAttrSet(addonResourceName, "addon_version"),
					resource.TestCheckResourceAttr(addonResourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      addonResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEksAddon_disappears(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddon_Basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEksAddon(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksAddon_disappears_Cluster(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddon_Basic(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					testAccCheckAWSEksClusterDisappears(ctx, &addon),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksAddon_AddonVersion(t *testing.T) {
	var addon1, addon2 eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	addonVersion1 := "v1.6.3-eksbuild.1"
	addonVersion2 := "v1.7.5-eksbuild.1"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigAddonVersion(rName, addonName, addonVersion1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "addon_version", addonVersion1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAWSEksAddonConfigAddonVersion(rName, addonName, addonVersion2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "addon_version", addonVersion2),
				),
			},
		},
	})
}

func TestAccAWSEksAddon_ResolveConflicts(t *testing.T) {
	var addon1, addon2 eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigResolveConflicts(rName, addonName, eks.ResolveConflictsNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsNone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAWSEksAddonConfigResolveConflicts(rName, addonName, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsOverwrite),
				),
			},
		},
	})
}

func TestAccAWSEksAddon_ServiceAccountRoleArn(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	serviceRoleResourceName := "aws_iam_role.test-service-role"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigServiceAccountRoleArn(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", serviceRoleResourceName, "arn"),
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

func TestAccAWSEksAddon_Tags(t *testing.T) {
	var addon1, addon2, addon3 eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksAddonConfigTags2(rName, addonName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEksAddon_defaultTags_providerOnly(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAWSEksAddon_Basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags2("providerkey1", "providervalue1", "providerkey2", "providervalue2"),
					testAccAWSEksAddon_Basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccAWSEksAddon_Basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSEksAddon_defaultTags_updateToProviderOnly(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("key1", "value1"),
					testAccAWSEksAddon_Basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccAWSEksAddon_defaultTags_updateToResourceOnly(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("key1", "value1"),
					testAccAWSEksAddon_Basic(rName, addonName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccAWSEksAddon_defaultTags_providerAndResource_nonOverlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "resourcekey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccAWSEksAddonConfigTags2(rName, addonName, "resourcekey1", "resourcevalue1", "resourcekey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey2", "resourcevalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("providerkey2", "providervalue2"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey3", "resourcevalue3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey3", "resourcevalue3"),
				),
			},
		},
	})
}

func TestAccAWSEksAddon_defaultTags_providerAndResource_overlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "overlapkey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags2("overlapkey1", "providervalue1", "overlapkey2", "providervalue2"),
					testAccAWSEksAddonConfigTags2(rName, addonName, "overlapkey1", "resourcevalue1", "overlapkey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey2", "resourcevalue2"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "overlapkey1", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue2"),
				),
			},
		},
	})
}

func TestAccAWSEksAddon_defaultTags_providerAndResource_duplicateTag(t *testing.T) {
	var providers []*schema.Provider

	rName := acctest.RandomWithPrefix("tf-acc-test")
	addonName := "vpc-cni"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "overlapkey", "overlapvalue"),
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`"tags" are identical to those in the "default_tags" configuration block`),
			},
		},
	})
}

func TestAccAWSEksAddon_defaultAndIgnoreTags(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					testAccCheckEksAddonUpdateTags(&addon, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSEksAddon_ignoreTags(t *testing.T) {
	var providers []*schema.Provider
	var addon eks.Addon

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, eks.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					testAccCheckEksAddonUpdateTags(&addon, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: composeConfig(
					testAccProviderConfigIgnoreTagsKeys1("ignorekey1"),
					testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckAWSEksAddonExists(ctx context.Context, resourceName string, addon *eks.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no EKS Addon ID is set")
		}

		clusterName, addonName, err := resourceAwsEksAddonParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if err != nil {
			return err
		}

		if output == nil || output.Addon == nil {
			return fmt.Errorf("EKS Addon (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Addon.AddonName) != addonName {
			return fmt.Errorf("EKS Addon (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Addon.ClusterName) != clusterName {
			return fmt.Errorf("EKS Addon (%s) not found", rs.Primary.ID)
		}

		*addon = *output.Addon

		return nil
	}
}

func testAccCheckAWSEksAddonDestroy(s *terraform.State) error {
	ctx := context.TODO()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_addon" {
			continue
		}

		clusterName, addonName, err := resourceAwsEksAddonParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn

		// Handle eventual consistency
		err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
			output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
				AddonName:   aws.String(addonName),
				ClusterName: aws.String(clusterName),
			})

			if err != nil {
				if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
					return nil
				}
				return resource.NonRetryableError(err)
			}

			if output != nil && output.Addon != nil && aws.StringValue(output.Addon.AddonName) == addonName {
				return resource.RetryableError(fmt.Errorf("EKS Addon (%s) still exists", rs.Primary.ID))
			}

			return nil
		})

		return err
	}

	return nil
}

func testAccCheckAWSEksClusterDisappears(ctx context.Context, addon *eks.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DeleteClusterInput{
			Name: addon.ClusterName,
		}

		_, err := conn.DeleteClusterWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return err
		}

		return waitForDeleteEksCluster(conn, aws.StringValue(addon.ClusterName), 30*time.Minute)
	}
}

func testAccPreCheckAWSEksAddon(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).eksconn

	input := &eks.DescribeAddonVersionsInput{}

	_, err := conn.DescribeAddonVersions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckEksAddonUpdateTags(addon *eks.Addon, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		return keyvaluetags.EksUpdateTags(conn, aws.StringValue(addon.AddonArn), oldTags, newTags)
	}
}

func testAccAWSEksAddonConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name                          = "terraform-testacc-eks-cluster-base"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = "terraform-testacc-eks-cluster-base"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName)
}

func testAccAWSEksAddon_Basic(rName, addonName string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q
}
`, rName, addonName))
}

func testAccAWSEksAddonConfigAddonVersion(rName, addonName, addonVersion string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  addon_version     = %[3]q
  resolve_conflicts = "OVERWRITE"
}
`, rName, addonName, addonVersion))
}

func testAccAWSEksAddonConfigResolveConflicts(rName, addonName, resolveConflicts string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  resolve_conflicts = %[3]q
}
`, rName, addonName, resolveConflicts))
}

func testAccAWSEksAddonConfigServiceAccountRoleArn(rName, addonName string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_iam_role" "test-service-role" {
  name               = "test-service-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_eks_addon" "test" {
  cluster_name             = aws_eks_cluster.test.name
  addon_name               = %[2]q
  service_account_role_arn = aws_iam_role.test-service-role.arn
}
`, rName, addonName))
}

func testAccAWSEksAddonConfigTags1(rName, addonName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, addonName, tagKey1, tagValue1))
}

func testAccAWSEksAddonConfigTags2(rName, addonName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, addonName, tagKey1, tagValue1, tagKey2, tagValue2))
}
