package aws

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/mitchellh/go-homedir"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsS3BucketObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketObjectCreate,
		Read:   resourceAwsS3BucketObjectRead,
		Update: resourceAwsS3BucketObjectUpdate,
		Delete: resourceAwsS3BucketObjectDelete,

		CustomizeDiff: resourceAwsS3BucketObjectCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"acl": {
				Type:     schema.TypeString,
				Default:  s3.ObjectCannedACLPrivate,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectCannedACLPrivate,
					s3.ObjectCannedACLPublicRead,
					s3.ObjectCannedACLPublicReadWrite,
					s3.ObjectCannedACLAuthenticatedRead,
					s3.ObjectCannedACLAwsExecRead,
					s3.ObjectCannedACLBucketOwnerRead,
					s3.ObjectCannedACLBucketOwnerFullControl,
				}, false),
			},

			"cache_control": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"content_disposition": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"content_encoding": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"content_language": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"metadata": {
				Type:         schema.TypeMap,
				ValidateFunc: validateMetadataIsLowerCase,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
			},

			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"source": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"content", "content_base64"},
			},

			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source", "content_base64"},
			},

			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source", "content"},
			},

			"storage_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectStorageClassStandard,
					s3.ObjectStorageClassReducedRedundancy,
					s3.ObjectStorageClassGlacier,
					s3.ObjectStorageClassStandardIa,
					s3.ObjectStorageClassOnezoneIa,
					s3.ObjectStorageClassIntelligentTiering,
					s3.ObjectStorageClassDeepArchive,
				}, false),
			},

			"server_side_encryption": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					s3.ServerSideEncryptionAes256,
					s3.ServerSideEncryptionAwsKms,
				}, false),
				Computed: true,
			},

			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"etag": {
				Type: schema.TypeString,
				// This will conflict with SSE-C and SSE-KMS encryption and multi-part upload
				// if/when it's actually implemented. The Etag then won't match raw-file MD5.
				// See http://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonResponseHeaders.html
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"kms_key_id"},
			},

			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"website_redirect": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"object_lock_legal_hold_status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectLockLegalHoldStatusOn,
					s3.ObjectLockLegalHoldStatusOff,
				}, false),
			},

			"object_lock_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectLockModeGovernance,
					s3.ObjectLockModeCompliance,
				}, false),
			},

			"object_lock_retain_until_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
		},
	}
}

func resourceAwsS3BucketObjectPut(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	var body io.ReadSeeker

	if v, ok := d.GetOk("source"); ok {
		source := v.(string)
		path, err := homedir.Expand(source)
		if err != nil {
			return fmt.Errorf("Error expanding homedir in source (%s): %s", source, err)
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening S3 bucket object source (%s): %s", path, err)
		}

		body = file
		defer func() {
			err := file.Close()
			if err != nil {
				log.Printf("[WARN] Error closing S3 bucket object source (%s): %s", path, err)
			}
		}()
	} else if v, ok := d.GetOk("content"); ok {
		content := v.(string)
		body = bytes.NewReader([]byte(content))
	} else if v, ok := d.GetOk("content_base64"); ok {
		content := v.(string)
		// We can't do streaming decoding here (with base64.NewDecoder) because
		// the AWS SDK requires an io.ReadSeeker but a base64 decoder can't seek.
		contentRaw, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return fmt.Errorf("error decoding content_base64: %s", err)
		}
		body = bytes.NewReader(contentRaw)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	putInput := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String(d.Get("acl").(string)),
		Body:   body,
	}

	if v, ok := d.GetOk("storage_class"); ok {
		putInput.StorageClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_control"); ok {
		putInput.CacheControl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_type"); ok {
		putInput.ContentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metadata"); ok {
		putInput.Metadata = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("content_encoding"); ok {
		putInput.ContentEncoding = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_language"); ok {
		putInput.ContentLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_disposition"); ok {
		putInput.ContentDisposition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_side_encryption"); ok {
		putInput.ServerSideEncryption = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		putInput.SSEKMSKeyId = aws.String(v.(string))
		putInput.ServerSideEncryption = aws.String(s3.ServerSideEncryptionAwsKms)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		// The tag-set must be encoded as URL Query parameters.
		putInput.Tagging = aws.String(keyvaluetags.New(v).IgnoreAws().UrlEncode())
	}

	if v, ok := d.GetOk("website_redirect"); ok {
		putInput.WebsiteRedirectLocation = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_legal_hold_status"); ok {
		putInput.ObjectLockLegalHoldStatus = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_mode"); ok {
		putInput.ObjectLockMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_retain_until_date"); ok {
		putInput.ObjectLockRetainUntilDate = expandS3ObjectLockRetainUntilDate(v.(string))
	}

	if _, err := s3conn.PutObject(putInput); err != nil {
		return fmt.Errorf("Error putting object in S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(key)
	return resourceAwsS3BucketObjectRead(d, meta)
}

func resourceAwsS3BucketObjectCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsS3BucketObjectPut(d, meta)
}

func resourceAwsS3BucketObjectRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	resp, err := s3conn.HeadObject(
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		// If S3 returns a 404 Request Failure, mark the object as destroyed
		if awsErr, ok := err.(awserr.RequestFailure); ok && awsErr.StatusCode() == 404 {
			d.SetId("")
			log.Printf("[WARN] Error Reading Object (%s), object not found (HTTP status 404)", key)
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Reading S3 Bucket Object meta: %s", resp)

	d.Set("cache_control", resp.CacheControl)
	d.Set("content_disposition", resp.ContentDisposition)
	d.Set("content_encoding", resp.ContentEncoding)
	d.Set("content_language", resp.ContentLanguage)
	d.Set("content_type", resp.ContentType)
	metadata := pointersMapToStringList(resp.Metadata)

	// AWS Go SDK capitalizes metadata, this is a workaround. https://github.com/aws/aws-sdk-go/issues/445
	for k, v := range metadata {
		delete(metadata, k)
		metadata[strings.ToLower(k)] = v
	}

	if err := d.Set("metadata", metadata); err != nil {
		return fmt.Errorf("error setting metadata: %s", err)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("server_side_encryption", resp.ServerSideEncryption)
	d.Set("website_redirect", resp.WebsiteRedirectLocation)
	d.Set("object_lock_legal_hold_status", resp.ObjectLockLegalHoldStatus)
	d.Set("object_lock_mode", resp.ObjectLockMode)
	d.Set("object_lock_retain_until_date", flattenS3ObjectLockRetainUntilDate(resp.ObjectLockRetainUntilDate))

	// Only set non-default KMS key ID (one that doesn't match default)
	if resp.SSEKMSKeyId != nil {
		// retrieve S3 KMS Default Master Key
		kmsconn := meta.(*AWSClient).kmsconn
		kmsresp, err := kmsconn.DescribeKey(&kms.DescribeKeyInput{
			KeyId: aws.String("alias/aws/s3"),
		})
		if err != nil {
			return fmt.Errorf("Failed to describe default S3 KMS key (alias/aws/s3): %s", err)
		}

		if *resp.SSEKMSKeyId != *kmsresp.KeyMetadata.Arn {
			log.Printf("[DEBUG] S3 object is encrypted using a non-default KMS Key ID: %s", *resp.SSEKMSKeyId)
			d.Set("kms_key_id", resp.SSEKMSKeyId)
		}
	}
	// See https://forums.aws.amazon.com/thread.jspa?threadID=44003
	d.Set("etag", strings.Trim(aws.StringValue(resp.ETag), `"`))

	// The "STANDARD" (which is also the default) storage
	// class when set would not be included in the results.
	d.Set("storage_class", s3.StorageClassStandard)
	if resp.StorageClass != nil {
		d.Set("storage_class", resp.StorageClass)
	}

	// Retry due to S3 eventual consistency
	tags, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return keyvaluetags.S3ObjectListTags(s3conn, bucket, key)
	})

	if err != nil {
		return fmt.Errorf("error listing tags for S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	if err := d.Set("tags", tags.(keyvaluetags.KeyValueTags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsS3BucketObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	// Changes to any of these attributes requires creation of a new object version (if bucket is versioned):
	for _, key := range []string{
		"cache_control",
		"content_base64",
		"content_disposition",
		"content_encoding",
		"content_language",
		"content_type",
		"content",
		"etag",
		"kms_key_id",
		"metadata",
		"server_side_encryption",
		"source",
		"storage_class",
		"website_redirect",
	} {
		if d.HasChange(key) {
			return resourceAwsS3BucketObjectPut(d, meta)
		}
	}

	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	if d.HasChange("acl") {
		_, err := conn.PutObjectAcl(&s3.PutObjectAclInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			ACL:    aws.String(d.Get("acl").(string)),
		})
		if err != nil {
			return fmt.Errorf("error putting S3 object ACL: %s", err)
		}
	}

	if d.HasChange("object_lock_legal_hold_status") {
		_, err := conn.PutObjectLegalHold(&s3.PutObjectLegalHoldInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			LegalHold: &s3.ObjectLockLegalHold{
				Status: aws.String(d.Get("object_lock_legal_hold_status").(string)),
			},
		})
		if err != nil {
			return fmt.Errorf("error putting S3 object lock legal hold: %s", err)
		}
	}

	if d.HasChanges("object_lock_mode", "object_lock_retain_until_date") {
		req := &s3.PutObjectRetentionInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Retention: &s3.ObjectLockRetention{
				Mode:            aws.String(d.Get("object_lock_mode").(string)),
				RetainUntilDate: expandS3ObjectLockRetainUntilDate(d.Get("object_lock_retain_until_date").(string)),
			},
		}

		// Bypass required to lower or clear retain-until date.
		if d.HasChange("object_lock_retain_until_date") {
			oraw, nraw := d.GetChange("object_lock_retain_until_date")
			o := expandS3ObjectLockRetainUntilDate(oraw.(string))
			n := expandS3ObjectLockRetainUntilDate(nraw.(string))
			if n == nil || (o != nil && n.Before(*o)) {
				req.BypassGovernanceRetention = aws.Bool(true)
			}
		}

		_, err := conn.PutObjectRetention(req)
		if err != nil {
			return fmt.Errorf("error putting S3 object lock retention: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.S3ObjectUpdateTags(conn, bucket, key, o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsS3BucketObjectRead(d, meta)
}

func resourceAwsS3BucketObjectDelete(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	// We are effectively ignoring any leading '/' in the key name as aws.Config.DisableRestProtocolURICleaning is false
	key = strings.TrimPrefix(key, "/")

	var err error
	if _, ok := d.GetOk("version_id"); ok {
		err = deleteAllS3ObjectVersions(s3conn, bucket, key, d.Get("force_destroy").(bool), false)
	} else {
		err = deleteS3ObjectVersion(s3conn, bucket, key, "", false)
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	return nil
}

func validateMetadataIsLowerCase(v interface{}, k string) (ws []string, errors []error) {
	value := v.(map[string]interface{})

	for k := range value {
		if k != strings.ToLower(k) {
			errors = append(errors, fmt.Errorf(
				"Metadata must be lowercase only. Offending key: %q", k))
		}
	}
	return
}

func resourceAwsS3BucketObjectCustomizeDiff(d *schema.ResourceDiff, meta interface{}) error {
	if d.HasChange("etag") {
		d.SetNewComputed("version_id")
	}

	return nil
}

// deleteAllS3ObjectVersions deletes all versions of a specified key from an S3 bucket.
// If key is empty then all versions of all objects are deleted.
// Set force to true to override any S3 object lock protections on object lock enabled buckets.
func deleteAllS3ObjectVersions(conn *s3.S3, bucketName, key string, force, ignoreObjectErrors bool) error {
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}
	if key != "" {
		input.Prefix = aws.String(key)
	}

	var lastErr error
	err := conn.ListObjectVersionsPages(input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, objectVersion := range page.Versions {
			objectKey := aws.StringValue(objectVersion.Key)
			objectVersionID := aws.StringValue(objectVersion.VersionId)

			if key != "" && key != objectKey {
				continue
			}

			err := deleteS3ObjectVersion(conn, bucketName, objectKey, objectVersionID, force)
			if isAWSErr(err, "AccessDenied", "") && force {
				// Remove any legal hold.
				resp, err := conn.HeadObject(&s3.HeadObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       objectVersion.Key,
					VersionId: objectVersion.VersionId,
				})

				if err != nil {
					log.Printf("[ERROR] Error getting S3 Bucket (%s) Object (%s) Version (%s) metadata: %s", bucketName, objectKey, objectVersionID, err)
					lastErr = err
					continue
				}

				if aws.StringValue(resp.ObjectLockLegalHoldStatus) == s3.ObjectLockLegalHoldStatusOn {
					_, err := conn.PutObjectLegalHold(&s3.PutObjectLegalHoldInput{
						Bucket:    aws.String(bucketName),
						Key:       objectVersion.Key,
						VersionId: objectVersion.VersionId,
						LegalHold: &s3.ObjectLockLegalHold{
							Status: aws.String(s3.ObjectLockLegalHoldStatusOff),
						},
					})

					if err != nil {
						log.Printf("[ERROR] Error putting S3 Bucket (%s) Object (%s) Version(%s) legal hold: %s", bucketName, objectKey, objectVersionID, err)
						lastErr = err
						continue
					}

					// Attempt to delete again.
					err = deleteS3ObjectVersion(conn, bucketName, objectKey, objectVersionID, force)

					if err != nil {
						lastErr = err
					}

					continue
				}

				// AccessDenied for another reason.
				lastErr = fmt.Errorf("AccessDenied deleting S3 Bucket (%s) Object (%s) Version: %s", bucketName, objectKey, objectVersionID)
				continue
			}

			if err != nil {
				lastErr = err
			}
		}

		return !lastPage
	})

	if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
		err = nil
	}

	if err != nil {
		return err
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return fmt.Errorf("error deleting at least one object version, last error: %s", lastErr)
		}

		lastErr = nil
	}

	err = conn.ListObjectVersionsPages(input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deleteMarker := range page.DeleteMarkers {
			deleteMarkerKey := aws.StringValue(deleteMarker.Key)
			deleteMarkerVersionID := aws.StringValue(deleteMarker.VersionId)

			if key != "" && key != deleteMarkerKey {
				continue
			}

			// Delete markers have no object lock protections.
			err := deleteS3ObjectVersion(conn, bucketName, deleteMarkerKey, deleteMarkerVersionID, false)

			if err != nil {
				lastErr = err
			}
		}

		return !lastPage
	})

	if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
		err = nil
	}

	if err != nil {
		return err
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return fmt.Errorf("error deleting at least one object delete marker, last error: %s", lastErr)
		}

		lastErr = nil
	}

	return nil
}

// deleteS3ObjectVersion deletes a specific bucket object version.
// Set force to true to override any S3 object lock protections.
func deleteS3ObjectVersion(conn *s3.S3, b, k, v string, force bool) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	}

	if v != "" {
		input.VersionId = aws.String(v)
	}

	if force {
		input.BypassGovernanceRetention = aws.Bool(true)
	}

	log.Printf("[INFO] Deleting S3 Bucket (%s) Object (%s) Version: %s", b, k, v)
	_, err := conn.DeleteObject(input)

	if err != nil {
		log.Printf("[WARN] Error deleting S3 Bucket (%s) Object (%s) Version (%s): %s", b, k, v, err)
	}

	if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, s3.ErrCodeNoSuchKey, "") {
		return nil
	}

	return err
}

func expandS3ObjectLockRetainUntilDate(v string) *time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}

	return aws.Time(t)
}

func flattenS3ObjectLockRetainUntilDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format(time.RFC3339)
}
