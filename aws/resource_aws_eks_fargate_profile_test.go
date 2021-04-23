package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/eks"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_eks_fargate_profile", &resource.Sweeper{
		Name: "aws_eks_fargate_profile",
		F:    testSweepEksFargateProfiles,
	})
}

func testSweepEksFargateProfiles(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).eksconn

	var errors error
	input := &eks.ListClustersInput{}
	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			clusterName := aws.StringValue(cluster)
			input := &eks.ListFargateProfilesInput{
				ClusterName: cluster,
			}
			err := conn.ListFargateProfilesPages(input, func(page *eks.ListFargateProfilesOutput, lastPage bool) bool {
				for _, profile := range page.FargateProfileNames {
					profileName := aws.StringValue(profile)
					log.Printf("[INFO] Deleting Fargate Profile %q", profileName)
					input := &eks.DeleteFargateProfileInput{
						ClusterName:        cluster,
						FargateProfileName: profile,
					}
					_, err := conn.DeleteFargateProfile(input)

					if err != nil && !isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
						errors = multierror.Append(errors, fmt.Errorf("error deleting EKS Fargate Profile %q: %w", profileName, err))
						continue
					}

					if err := waitForEksFargateProfileDeletion(conn, clusterName, profileName, 10*time.Minute); err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EKS Fargate Profile %q deletion: %w", profileName, err))
						continue
					}
				}
				return true
			})
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error listing Fargate Profiles for EKS Cluster %s: %w", clusterName, err))
			}
		}

		return true
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Clusters sweep for %s: %s", region, err)
		return errors // In case we have completed some pages, but had errors
	}
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EKS Clusters: %w", err))
	}

	return errors
}

func TestAccAWSEksFargateProfile_basic(t *testing.T) {
	var fargateProfile eks.FargateProfile
	rName := acctest.RandomWithPrefix("tf-acc-test")
	eksClusterResourceName := "aws_eks_cluster.test"
	iamRoleResourceName := "aws_iam_role.pod"
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksFargateProfile(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksFargateProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksFargateProfileConfigFargateProfileName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName, &fargateProfile),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("fargateprofile/%[1]s/%[1]s/.+", rName))),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", eksClusterResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "fargate_profile_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "pod_execution_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", eks.FargateProfileStatusActive),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSEksFargateProfile_disappears(t *testing.T) {
	var fargateProfile eks.FargateProfile
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksFargateProfile(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksFargateProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksFargateProfileConfigFargateProfileName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName, &fargateProfile),
					testAccCheckAWSEksFargateProfileDisappears(&fargateProfile),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksFargateProfile_Multi_Profile(t *testing.T) {
	var fargateProfile eks.FargateProfile
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName1 := "aws_eks_fargate_profile.test.0"
	resourceName2 := "aws_eks_fargate_profile.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksFargateProfile(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksFargateProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksFargateProfileConfigFargateProfileMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName1, &fargateProfile),
					testAccCheckAWSEksFargateProfileExists(resourceName2, &fargateProfile),
				),
			},
		},
	})
}

func TestAccAWSEksFargateProfile_Selector_Labels(t *testing.T) {
	var fargateProfile1 eks.FargateProfile
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksFargateProfile(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksFargateProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksFargateProfileConfigSelectorLabels1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName, &fargateProfile1),
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

func TestAccAWSEksFargateProfile_Tags(t *testing.T) {
	var fargateProfile1, fargateProfile2, fargateProfile3 eks.FargateProfile
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_fargate_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksFargateProfile(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksFargateProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksFargateProfileConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName, &fargateProfile1),
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
				Config: testAccAWSEksFargateProfileConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName, &fargateProfile2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksFargateProfileConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksFargateProfileExists(resourceName, &fargateProfile3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEksFargateProfileExists(resourceName string, fargateProfile *eks.FargateProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Fargate Profile ID is set")
		}

		clusterName, fargateProfileName, err := resourceAwsEksFargateProfileParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DescribeFargateProfileInput{
			ClusterName:        aws.String(clusterName),
			FargateProfileName: aws.String(fargateProfileName),
		}

		output, err := conn.DescribeFargateProfile(input)

		if err != nil {
			return err
		}

		if output == nil || output.FargateProfile == nil {
			return fmt.Errorf("EKS Fargate Profile (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.FargateProfile.FargateProfileName) != fargateProfileName {
			return fmt.Errorf("EKS Fargate Profile (%s) not found", rs.Primary.ID)
		}

		if got, want := aws.StringValue(output.FargateProfile.Status), eks.FargateProfileStatusActive; got != want {
			return fmt.Errorf("EKS Fargate Profile (%s) not in %s status, got: %s", rs.Primary.ID, want, got)
		}

		*fargateProfile = *output.FargateProfile

		return nil
	}
}

func testAccCheckAWSEksFargateProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).eksconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_fargate_profile" {
			continue
		}

		clusterName, fargateProfileName, err := resourceAwsEksFargateProfileParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &eks.DescribeFargateProfileInput{
			ClusterName:        aws.String(clusterName),
			FargateProfileName: aws.String(fargateProfileName),
		}

		output, err := conn.DescribeFargateProfile(input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if output != nil && output.FargateProfile != nil && aws.StringValue(output.FargateProfile.FargateProfileName) == fargateProfileName {
			return fmt.Errorf("EKS Fargate Profile (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSEksFargateProfileDisappears(fargateProfile *eks.FargateProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DeleteFargateProfileInput{
			ClusterName:        fargateProfile.ClusterName,
			FargateProfileName: fargateProfile.FargateProfileName,
		}

		_, err := conn.DeleteFargateProfile(input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		return waitForEksFargateProfileDeletion(conn, aws.StringValue(fargateProfile.ClusterName), aws.StringValue(fargateProfile.FargateProfileName), 10*time.Minute)
	}
}

func testAccPreCheckAWSEksFargateProfile(t *testing.T) {
	// Most PreCheck functions try to use a list or describe API call to
	// determine service or functionality availability, however
	// ListFargateProfiles requires a valid ClusterName and does not indicate
	// that the functionality is unavailable in a region. The create API call
	// fails with same "ResourceNotFoundException: No cluster found" before
	// returning the definitive "InvalidRequestException: CreateFargateProfile
	// is not supported for region" error. We do not want to wait 20 minutes to
	// create and destroy an EKS Cluster just to find the real error, instead
	// we take the least desirable approach of hardcoding allowed regions.
	allowedRegions := []string{
		endpoints.ApEast1RegionID,
		endpoints.ApNortheast1RegionID,
		endpoints.ApNortheast2RegionID,
		endpoints.ApSouth1RegionID,
		endpoints.ApSoutheast1RegionID,
		endpoints.ApSoutheast2RegionID,
		endpoints.CaCentral1RegionID,
		endpoints.EuCentral1RegionID,
		endpoints.EuNorth1RegionID,
		endpoints.EuWest1RegionID,
		endpoints.EuWest2RegionID,
		endpoints.EuWest3RegionID,
		endpoints.MeSouth1RegionID,
		endpoints.SaEast1RegionID,
		endpoints.UsEast1RegionID,
		endpoints.UsEast2RegionID,
		endpoints.UsWest1RegionID,
		endpoints.UsWest2RegionID,
	}
	region := testAccProvider.Meta().(*AWSClient).region

	for _, allowedRegion := range allowedRegions {
		if region == allowedRegion {
			return
		}
	}

	message := fmt.Sprintf(`Test provider region (%s) not found in allowed EKS Fargate regions: %v

The allowed regions are hardcoded in the acceptance testing since dynamically determining the
functionality requires creating and destroying a real EKS Cluster, which is a lengthy process.
If this check is out of date, please create an issue in the Terraform AWS Provider
repository (https://github.com/hashicorp/terraform-provider-aws) or submit a PR to update the
check itself (testAccPreCheckAWSEksFargateProfile).

For the most up to date supported region information, see the EKS User Guide:
https://docs.aws.amazon.com/eks/latest/userguide/fargate.html
`, region, allowedRegions)

	t.Skipf("skipping acceptance testing:\n\n%s", message)
}

func testAccAWSEksFargateProfileConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "pod" {
  name = "%[1]s-pod"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks-fargate-pods.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "pod-AmazonEKSFargatePodExecutionRolePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"
  role       = aws_iam_role.pod.name
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                          = "tf-acc-test-eks-fargate-profile"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_main_route_table_association" "test" {
  route_table_id = aws_route_table.public.id
  vpc_id         = aws_vpc.test.id
}

resource "aws_subnet" "private" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 2)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = "tf-acc-test-eks-fargate-profile-private"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eip" "private" {
  count      = 2
  depends_on = [aws_internet_gateway.test]

  vpc = true
}

resource "aws_nat_gateway" "private" {
  count = 2

  allocation_id = aws_eip.private[count.index].id
  subnet_id     = aws_subnet.private[count.index].id
}

resource "aws_route_table" "private" {
  count = 2

  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.private[count.index].id
  }
}

resource "aws_route_table_association" "private" {
  count = 2

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

resource "aws_subnet" "public" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = "tf-acc-test-eks-fargate-profile-public"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.public[*].id
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy,
    aws_main_route_table_association.test,
  ]
}
`, rName)
}

func testAccAWSEksFargateProfileConfigFargateProfileName(rName string) string {
	return testAccAWSEksFargateProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName)
}

func testAccAWSEksFargateProfileConfigFargateProfileMultiple(rName string) string {
	return composeConfig(
		testAccAWSEksFargateProfileConfigBase(rName),
		fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  count = 2

  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = "%[1]s-${count.index}"
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName))
}

func testAccAWSEksFargateProfileConfigSelectorLabels1(rName, labelKey1, labelValue1 string) string {
	return testAccAWSEksFargateProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    labels = {
      %[2]q = %[3]q
    }
    namespace = "test"
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName, labelKey1, labelValue1)
}

func testAccAWSEksFargateProfileConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSEksFargateProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSEksFargateProfileConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSEksFargateProfileConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_fargate_profile" "test" {
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = %[1]q
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
