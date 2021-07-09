package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/waiter"
)

func resourceAwsStorageGatewayFsxAssociateFileSystem() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsStorageGatewayFsxAssociateFileSystemCreate,
		Read:          resourceAwsStorageGatewayFsxAssociateFileSystemRead,
		Update:        resourceAwsStorageGatewayFsxAssociateFileSystemUpdate,
		Delete:        resourceAwsStorageGatewayFsxAssociateFileSystemDelete,
		CustomizeDiff: customdiff.Sequence(SetTagsDiff),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
				Default:      "",
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{0}),
								validation.IntBetween(300, 2592000),
							),
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^\w[\w\.\- ]*$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
		},
	}
}

func resourceAwsStorageGatewayFsxAssociateFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &storagegateway.AssociateFileSystemInput{
		ClientToken: aws.String(resource.UniqueId()),
		GatewayARN:  aws.String(d.Get("gateway_arn").(string)),
		LocationARN: aws.String(d.Get("location_arn").(string)),
		Password:    aws.String(d.Get("password").(string)),
		Tags:        tags.IgnoreAws().StoragegatewayTags(),
		UserName:    aws.String(d.Get("username").(string)),
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandStorageGatewayFsxAssociateFileSystemCacheAttributes(v.([]interface{}))
	}

	log.Printf("[DEBUG] Associating File System to Storage Gateway: %s", input)
	output, err := conn.AssociateFileSystem(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInvalidGatewayRequestException) {
			var igrex *storagegateway.InvalidGatewayRequestException
			if ok := errors.As(err, &igrex); ok {
				if err := igrex.Error_; err != nil {
					if aws.StringValue(err.ErrorCode) == "FileSystemAssociationNotFound" {
						log.Printf("[WARN] FSX File System %q not found, removing from state", d.Id())
						d.SetId("")
						return nil
					}
				}
			}
		}

		return fmt.Errorf("Error associating file system to storage gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.FileSystemAssociationARN))
	log.Printf("[INFO] Storage Gateway FSx File System Association ID: %s", d.Id())

	if _, err = waiter.FsxFileSystemAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway FSx File System Association (%q) to be Available: %w", d.Id(), err)
	}

	return resourceAwsStorageGatewayFsxAssociateFileSystemRead(d, meta)
}

func resourceAwsStorageGatewayFsxAssociateFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway FSx File Systems: %s", input)
	output, err := conn.DescribeFileSystemAssociations(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInvalidGatewayRequestException) {
			var igrex *storagegateway.InvalidGatewayRequestException
			if ok := errors.As(err, &igrex); ok {
				if err := igrex.Error_; err != nil {
					if aws.StringValue(err.ErrorCode) == "FileSystemAssociationNotFound" {
						log.Printf("[WARN] FSX File System %q not found, removing from state", d.Id())
						d.SetId("")
						return nil
					}
				}
			}
		}

		return fmt.Errorf("error reading Storage Gateway FSx File System: %w", err)
	}

	if output == nil || len(output.FileSystemAssociationInfoList) == 0 || output.FileSystemAssociationInfoList[0] == nil {
		log.Printf("[WARN] Storage Gateway FSx File System %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	filesystem := output.FileSystemAssociationInfoList[0]

	arn := filesystem.FileSystemAssociationARN
	d.Set("arn", arn)

	d.Set("audit_destination_arn", filesystem.AuditDestinationARN)
	d.Set("gateway_arn", filesystem.GatewayARN)
	d.Set("location_arn", filesystem.LocationARN)

	if err := d.Set("cache_attributes", flattenStorageGatewayFsxAssociateFileSystemCacheAttributes(filesystem.CacheAttributes)); err != nil {
		return fmt.Errorf("error setting cache_attributes: %w", err)
	}

	tags := keyvaluetags.StoragegatewayKeyValueTags(filesystem.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsStorageGatewayFsxAssociateFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	if d.HasChanges("cache_attributes", "audit_destination_arn", "username", "password") {

		input := &storagegateway.UpdateFileSystemAssociationInput{
			Password:                 aws.String(d.Get("password").(string)),
			UserName:                 aws.String(d.Get("username").(string)),
			FileSystemAssociationARN: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("audit_destination_arn"); ok {
			input.AuditDestinationARN = aws.String(v.(string))
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandStorageGatewayFsxAssociateFileSystemCacheAttributes(v.([]interface{}))
		}

		log.Printf("[DEBUG] Updating Storage Gateway FSx File System Association: %s", input)
		_, err := conn.UpdateFileSystemAssociation(input)
		if err != nil {
			return fmt.Errorf("error updating Storage Gateway FSx File System Association: %w", err)
		}

		if _, err = waiter.FsxFileSystemAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway FSx File System Association (%q) to be Available: %w", d.Id(), err)
		}
	}

	return resourceAwsStorageGatewayFsxAssociateFileSystemRead(d, meta)
}

func resourceAwsStorageGatewayFsxAssociateFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DisassociateFileSystemInput{
		FileSystemAssociationARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway File System Association: %s", input)
	_, err := conn.DisassociateFileSystem(input)
	if err != nil {
		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file system association") {
			log.Printf("[WARN] Storage Gateway FSx File System Association %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}
	if _, err = waiter.FsxFileSystemDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error waiting for Storage Gateway FSx File System Association (%q) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandStorageGatewayFsxAssociateFileSystemCacheAttributes(l []interface{}) *storagegateway.CacheAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ca := &storagegateway.CacheAttributes{
		CacheStaleTimeoutInSeconds: aws.Int64(int64(m["cache_stale_timeout_in_seconds"].(int))),
	}

	return ca
}

func flattenStorageGatewayFsxAssociateFileSystemCacheAttributes(ca *storagegateway.CacheAttributes) []interface{} {
	if ca == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cache_stale_timeout_in_seconds": aws.Int64Value(ca.CacheStaleTimeoutInSeconds),
	}

	return []interface{}{m}
}
