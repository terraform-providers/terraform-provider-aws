package aws

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func init() {
	resource.AddTestSweepers("aws_s3_bucket_object", &resource.Sweeper{
		Name: "aws_s3_bucket_object",
		F:    testSweepS3BucketObjects,
	})
}

func testSweepS3BucketObjects(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).s3connUriCleaningDisabled
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBuckets(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Bucket Objects sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Bucket Objects: %s", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Bucket Objects to sweep")
		return nil
	}

	for _, bucket := range output.Buckets {
		bucketName := aws.StringValue(bucket.Name)

		hasPrefix := false
		prefixes := []string{"mybucket.", "mylogs.", "tf-acc", "tf-object-test", "tf-test", "tf-emr-bootstrap"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(bucketName, prefix) {
				hasPrefix = true
				break
			}
		}

		if !hasPrefix {
			log.Printf("[INFO] Skipping S3 Bucket: %s", bucketName)
			continue
		}

		bucketRegion, err := testS3BucketRegion(conn, bucketName)

		if err != nil {
			log.Printf("[ERROR] Error getting S3 Bucket (%s) Location: %s", bucketName, err)
			continue
		}

		if bucketRegion != region {
			log.Printf("[INFO] Skipping S3 Bucket (%s) in different region: %s", bucketName, bucketRegion)
			continue
		}

		objectLockEnabled, err := testS3BucketObjectLockEnabled(conn, bucketName)

		if err != nil {
			log.Printf("[ERROR] Error getting S3 Bucket (%s) Object Lock: %s", bucketName, err)
			continue
		}

		// Delete everything including locked objects. Ignore any object errors.
		err = deleteAllS3ObjectVersions(conn, bucketName, "", objectLockEnabled, true)

		if err != nil {
			return fmt.Errorf("error listing S3 Bucket (%s) Objects: %s", bucketName, err)
		}
	}

	return nil
}

func TestAccAWSS3BucketObject_noNameNoKey(t *testing.T) {
	bucketError := regexp.MustCompile(`bucket must not be empty`)
	keyError := regexp.MustCompile(`key must not be empty`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig:   func() {},
				Config:      testAccAWSS3BucketObjectConfigBasic("", "a key"),
				ExpectError: bucketError,
			},
			{
				PreConfig:   func() {},
				Config:      testAccAWSS3BucketObjectConfigBasic("a name", ""),
				ExpectError: keyError,
			},
		},
	})
}
func TestAccAWSS3BucketObject_empty(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigEmpty(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_source(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfigSource(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_content(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigContent(rInt, "some_bucket_content"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_etagEncryption(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectEtagEncryption(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "etag", "7b006ff4d70f68cc65061acf2f802e6f"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_contentBase64(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigContentBase64(rInt, base64.StdEncoding.EncodeToString([]byte("some_bucket_content"))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_withContentCharacteristics(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_withContentCharacteristics(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "binary/octet-stream"),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", "http://google.com"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_NonVersioned(t *testing.T) {
	sourceInitial := testAccAWSS3BucketObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)

	var originalObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAssumeRoleARNPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_NonVersioned(acctest.RandInt(), sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_updates(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	sourceInitial := testAccAWSS3BucketObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccAWSS3BucketObjectCreateTempFile(t, "modified object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, false, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "647d1d58e1011c743ec67d5e8af87b53"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, false, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &modifiedObj),
					testAccCheckAWSS3BucketObjectBody(&modifiedObj, "modified object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "1c7fd13df1515c2a13ad9eb068931f09"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_updateSameFile(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	startingData := "lane 8"
	changingData := "chicane"

	filename := testAccAWSS3BucketObjectCreateTempFile(t, startingData)
	defer os.Remove(filename)

	rewriteFile := func(*terraform.State) error {
		if err := ioutil.WriteFile(filename, []byte(changingData), 0644); err != nil {
			os.Remove(filename)
			t.Fatal(err)
		}
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, false, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, startingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "aa48b42f36a2652cbee40c30a5df7d25"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, false, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &modifiedObj),
					testAccCheckAWSS3BucketObjectBody(&modifiedObj, changingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "fafc05f8c4da0266a99154681ab86e8c"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_updatesWithVersioning(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	sourceInitial := testAccAWSS3BucketObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccAWSS3BucketObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, true, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, true, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &modifiedObj),
					testAccCheckAWSS3BucketObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&modifiedObj, &originalObj),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_updatesWithVersioningViaAccessPoint(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_object.test"
	accessPointResourceName := "aws_s3_access_point.test"

	sourceInitial := testAccAWSS3BucketObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccAWSS3BucketObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_updateableViaAccessPoint(rName, true, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", accessPointResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_updateableViaAccessPoint(rName, true, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &modifiedObj),
					testAccCheckAWSS3BucketObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&modifiedObj, &originalObj),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_kms(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withKMSId(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectSSE(resourceName, "aws:kms"),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_sse(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withSSE(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectSSE(resourceName, "AES256"),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_acl(t *testing.T) {
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_acl(rInt, "some_bucket_content", "private"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
					testAccCheckAWSS3BucketObjectAcl(resourceName, []string{"FULL_CONTROL"}),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_acl(rInt, "some_bucket_content", "public-read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					testAccCheckAWSS3BucketObjectAcl(resourceName, []string{"FULL_CONTROL", "READ"}),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_acl(rInt, "changed_some_bucket_content", "private"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "changed_some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
					testAccCheckAWSS3BucketObjectAcl(resourceName, []string{"FULL_CONTROL"}),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_metadata(t *testing.T) {
	rInt := acctest.RandInt()
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_withMetadata(rInt, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key2", "value2"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_withMetadata(rInt, "key1", "value1updated", "key3", "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key3", "value3"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfigEmpty(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_storageClass(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigContent(rInt, "some_bucket_content"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "STANDARD"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "STANDARD"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "REDUCED_REDUNDANCY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "REDUCED_REDUNDANCY"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "REDUCED_REDUNDANCY"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "GLACIER"),
				Check: resource.ComposeTestCheckFunc(
					// Can't GetObject on an object in Glacier without restoring it.
					resource.TestCheckResourceAttr(resourceName, "storage_class", "GLACIER"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "GLACIER"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "INTELLIGENT_TIERING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "INTELLIGENT_TIERING"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "INTELLIGENT_TIERING"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "DEEP_ARCHIVE"),
				Check: resource.ComposeTestCheckFunc(
					// 	Can't GetObject on an object in DEEP_ARCHIVE without restoring it.
					resource.TestCheckResourceAttr(resourceName, "storage_class", "DEEP_ARCHIVE"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "DEEP_ARCHIVE"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_tags(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	key := "test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withUpdatedTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withNoTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj4),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj4, &obj3),
					testAccCheckAWSS3BucketObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_tagsLeadingSlash(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	key := "/test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withUpdatedTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withNoTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj4),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj4, &obj3),
					testAccCheckAWSS3BucketObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_ObjectLockLegalHoldStartWithNone(t *testing.T) {
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_noObjectLockLegalHold(rInt, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockLegalHold(rInt, "stuff", "ON"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			// Remove legal hold but create a new object version to test force_destroy
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockLegalHold(rInt, "changed stuff", "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_ObjectLockLegalHoldStartWithOn(t *testing.T) {
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockLegalHold(rInt, "stuff", "ON"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "ON"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockLegalHold(rInt, "stuff", "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_ObjectLockRetentionStartWithNone(t *testing.T) {
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_noObjectLockRetention(rInt, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockRetention(rInt, "stuff", retainUntilDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate),
				),
			},
			// Remove retention period but create a new object version to test force_destroy
			{
				Config: testAccAWSS3BucketObjectConfig_noObjectLockRetention(rInt, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_ObjectLockRetentionStartWithSet(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	retainUntilDate1 := time.Now().UTC().AddDate(0, 0, 20).Format(time.RFC3339)
	retainUntilDate2 := time.Now().UTC().AddDate(0, 0, 30).Format(time.RFC3339)
	retainUntilDate3 := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockRetention(rInt, "stuff", retainUntilDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate1),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockRetention(rInt, "stuff", retainUntilDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate2),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_withObjectLockRetention(rInt, "stuff", retainUntilDate3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate3),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_noObjectLockRetention(rInt, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj4),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj4, &obj3),
					testAccCheckAWSS3BucketObjectBody(&obj4, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_legal_hold_status", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_defaultBucketSSE(t *testing.T) {
	var obj1 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_defaultBucketSSE(rInt, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_ignoreTags(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	key := "test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config: composeConfig(
					testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccAWSS3BucketObjectConfig_withNoTags(rInt, key, "stuff")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "stuff"),
					testAccCheckAWSS3BucketObjectUpdateTags(resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckAWSS3BucketObjectCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				PreConfig: func() {},
				Config: composeConfig(
					testAccProviderConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccAWSS3BucketObjectConfig_withTags(rInt, key, "stuff")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckAWSS3BucketObjectCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
						"Key1":       "AAA",
						"Key2":       "BBB",
						"Key3":       "CCC",
					}),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketObjectVersionIdDiffers(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if first.VersionId == nil {
			return fmt.Errorf("Expected first object to have VersionId: %s", first)
		}
		if second.VersionId == nil {
			return fmt.Errorf("Expected second object to have VersionId: %s", second)
		}

		if *first.VersionId == *second.VersionId {
			return fmt.Errorf("Expected Version IDs to differ, but they are equal (%s)", *first.VersionId)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectVersionIdEquals(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if first.VersionId == nil {
			return fmt.Errorf("Expected first object to have VersionId: %s", first)
		}
		if second.VersionId == nil {
			return fmt.Errorf("Expected second object to have VersionId: %s", second)
		}

		if *first.VersionId != *second.VersionId {
			return fmt.Errorf("Expected Version IDs to be equal, but they differ (%s, %s)", *first.VersionId, *second.VersionId)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectDestroy(s *terraform.State) error {
	s3conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_object" {
			continue
		}

		_, err := s3conn.HeadObject(
			&s3.HeadObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err == nil {
			return fmt.Errorf("AWS S3 Object still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSS3BucketObjectExists(n string, obj *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket Object ID is set")
		}

		s3conn := testAccProvider.Meta().(*AWSClient).s3conn
		out, err := s3conn.GetObject(
			&s3.GetObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err != nil {
			return fmt.Errorf("S3Bucket Object error: %s", err)
		}

		*obj = *out

		return nil
	}
}

func testAccCheckAWSS3BucketObjectBody(obj *s3.GetObjectOutput, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		body, err := ioutil.ReadAll(obj.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %s", err)
		}
		obj.Body.Close()

		if got := string(body); got != want {
			return fmt.Errorf("wrong result body %q; want %q", got, want)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectAcl(n string, expectedPerms []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := s3conn.GetObjectAcl(&s3.GetObjectAclInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Key:    aws.String(rs.Primary.Attributes["key"]),
		})

		if err != nil {
			return fmt.Errorf("GetObjectAcl error: %v", err)
		}

		var perms []string
		for _, v := range out.Grants {
			perms = append(perms, *v.Permission)
		}
		sort.Strings(perms)

		if !reflect.DeepEqual(perms, expectedPerms) {
			return fmt.Errorf("Expected ACL permissions to be %v, got %v", expectedPerms, perms)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectStorageClass(n, expectedClass string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := s3conn.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Key:    aws.String(rs.Primary.Attributes["key"]),
		})

		if err != nil {
			return fmt.Errorf("HeadObject error: %v", err)
		}

		// The "STANDARD" (which is also the default) storage
		// class when set would not be included in the results.
		storageClass := s3.StorageClassStandard
		if out.StorageClass != nil {
			storageClass = *out.StorageClass
		}

		if storageClass != expectedClass {
			return fmt.Errorf("Expected Storage Class to be %v, got %v",
				expectedClass, storageClass)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectSSE(n, expectedSSE string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := s3conn.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Key:    aws.String(rs.Primary.Attributes["key"]),
		})

		if err != nil {
			return fmt.Errorf("HeadObject error: %v", err)
		}

		if out.ServerSideEncryption == nil {
			return fmt.Errorf("Expected a non %v Server Side Encryption.", out.ServerSideEncryption)
		}

		sse := *out.ServerSideEncryption
		if sse != expectedSSE {
			return fmt.Errorf("Expected Server Side Encryption %v, got %v.",
				expectedSSE, sse)
		}

		return nil
	}
}

func testAccAWSS3BucketObjectCreateTempFile(t *testing.T, data string) string {
	tmpFile, err := ioutil.TempFile("", "tf-acc-s3-obj")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpFile.Name()

	err = ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		os.Remove(filename)
		t.Fatal(err)
	}

	return filename
}

func testAccCheckAWSS3BucketObjectUpdateTags(n string, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		return keyvaluetags.S3ObjectUpdateTags(conn, rs.Primary.Attributes["bucket"], rs.Primary.Attributes["key"], oldTags, newTags)
	}
}

func testAccCheckAWSS3BucketObjectCheckTags(n string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		got, err := keyvaluetags.S3ObjectListTags(conn, rs.Primary.Attributes["bucket"], rs.Primary.Attributes["key"])
		if err != nil {
			return err
		}

		want := keyvaluetags.New(expectedTags)
		if !reflect.DeepEqual(want, got) {
			return fmt.Errorf("Incorrect tags, want: %v got: %v", want, got)
		}

		return nil
	}
}

func testAccAWSS3BucketObjectConfigBasic(bucket, key string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_object" "object" {
  bucket = %q
  key    = %q
}
`, bucket, key)
}

func testAccAWSS3BucketObjectConfigEmpty(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "test-key"
}
`, randInt)
}

func testAccAWSS3BucketObjectConfigSource(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "test-key"
  source       = %q
  content_type = "binary/octet-stream"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_withContentCharacteristics(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket           = aws_s3_bucket.object_bucket.bucket
  key              = "test-key"
  source           = %q
  content_language = "en"
  content_type     = "binary/octet-stream"
  website_redirect = "http://google.com"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfigContent(randInt int, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = "test-key"
  content = %q
}
`, randInt, content)
}

func testAccAWSS3BucketObjectEtagEncryption(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_object" "object" {
  bucket                 = aws_s3_bucket.object_bucket.bucket
  key                    = "test-key"
  server_side_encryption = "AES256"
  source                 = %[2]q
  etag                   = filemd5(%[2]q)
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfigContentBase64(randInt int, contentBase64 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket         = aws_s3_bucket.object_bucket.bucket
  key            = "test-key"
  content_base64 = %q
}
`, randInt, contentBase64)
}

func testAccAWSS3BucketObjectConfig_updateable(randInt int, bucketVersioning bool, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket_3" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = %t
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[3]q
  etag   = filemd5(%[3]q)
}
`, randInt, bucketVersioning, source)
}

func testAccAWSS3BucketObjectConfig_updateableViaAccessPoint(rName string, bucketVersioning bool, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  versioning {
    enabled = %[2]t
  }
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_access_point.test.arn
  key    = "updateable-key"
  source = %[3]q
  etag   = filemd5(%[3]q)
}
`, rName, bucketVersioning, source)
}

func testAccAWSS3BucketObjectConfig_withKMSId(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "kms_key_1" {}

resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket     = aws_s3_bucket.object_bucket.bucket
  key        = "test-key"
  source     = %q
  kms_key_id = aws_kms_key.kms_key_1.arn
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_withSSE(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket                 = aws_s3_bucket.object_bucket.bucket
  key                    = "test-key"
  source                 = %q
  server_side_encryption = "AES256"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_acl(randInt int, content, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = "test-key"
  content = %q
  acl     = %q
}
`, randInt, content, acl)
}

func testAccAWSS3BucketObjectConfig_storageClass(randInt int, storage_class string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket        = aws_s3_bucket.object_bucket.bucket
  key           = "test-key"
  content       = "some_bucket_content"
  storage_class = %q
}
`, randInt, storage_class)
}

func testAccAWSS3BucketObjectConfig_withTags(randInt int, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = %q
  content = %q

  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, randInt, key, content)
}

func testAccAWSS3BucketObjectConfig_withUpdatedTags(randInt int, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = %q
  content = %q

  tags = {
    Key2 = "BBB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, randInt, key, content)
}

func testAccAWSS3BucketObjectConfig_withNoTags(randInt int, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = %q
  content = %q
}
`, randInt, key, content)
}

func testAccAWSS3BucketObjectConfig_withMetadata(randInt int, metadataKey1, metadataValue1, metadataKey2, metadataValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "test-key"

  metadata = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, randInt, metadataKey1, metadataValue1, metadataKey2, metadataValue2)
}

func testAccAWSS3BucketObjectConfig_noObjectLockLegalHold(randInt int, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket        = aws_s3_bucket.object_bucket.bucket
  key           = "test-key"
  content       = %q
  force_destroy = true
}
`, randInt, content)
}

func testAccAWSS3BucketObjectConfig_withObjectLockLegalHold(randInt int, content, legalHoldStatus string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket                        = aws_s3_bucket.object_bucket.bucket
  key                           = "test-key"
  content                       = %q
  object_lock_legal_hold_status = %q
  force_destroy                 = true
}
`, randInt, content, legalHoldStatus)
}

func testAccAWSS3BucketObjectConfig_noObjectLockRetention(randInt int, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket        = aws_s3_bucket.object_bucket.bucket
  key           = "test-key"
  content       = %q
  force_destroy = true
}
`, randInt, content)
}

func testAccAWSS3BucketObjectConfig_withObjectLockRetention(randInt int, content, retainUntilDate string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket                        = aws_s3_bucket.object_bucket.bucket
  key                           = "test-key"
  content                       = %q
  force_destroy                 = true
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = %q
}
`, randInt, content, retainUntilDate)
}

func testAccAWSS3BucketObjectConfig_NonVersioned(randInt int, source string) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowYeah",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    },
    {
      "Sid": "DenyStm1",
      "Effect": "Deny",
      "Action": [
        "s3:GetObjectVersion*",
        "s3:ListBucketVersions"
      ],
      "Resource": "*"
    }
  ]
}`

	return testAccProviderConfigAssumeRolePolicy(policy) + fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket_3" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[2]q
  etag   = filemd5(%[2]q)
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_defaultBucketSSE(randInt int, content string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test bucket objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.test.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = "test-key"
  content = %q
}
`, randInt, content)
}
