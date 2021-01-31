package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/waiter"
)

func resourceAwsStorageGatewaySmbFileShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsStorageGatewaySmbFileShareCreate,
		Read:   resourceAwsStorageGatewaySmbFileShareRead,
		Update: resourceAwsStorageGatewaySmbFileShareUpdate,
		Delete: resourceAwsStorageGatewaySmbFileShareDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "ActiveDirectory",
				ValidateFunc: validation.StringInSlice([]string{
					"ActiveDirectory",
					"GuestAccess",
				}, false),
			},
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"default_storage_class": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "S3_STANDARD",
				ValidateFunc: validation.StringInSlice([]string{
					"S3_ONEZONE_IA",
					"S3_STANDARD_IA",
					"S3_STANDARD",
					"S3_INTELLIGENT_TIERING",
				}, false),
			},
			"fileshare_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_share_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"guess_mime_type_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"invalid_user_list": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
				RequiredWith: []string{"kms_encrypted"},
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"object_acl": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.ObjectACLPrivate,
				ValidateFunc: validation.StringInSlice(storagegateway.ObjectACL_Values(), false),
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(300, 2592000),
						},
					},
				},
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"requester_pays": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"smb_acl_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"access_based_enumeration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"valid_user_list": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"admin_user_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"case_sensitivity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.CaseSensitivityClientSpecified,
				ValidateFunc: validation.StringInSlice(storagegateway.CaseSensitivity_Values(), false),
			},
			"notification_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "{}",
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^\{[\w\s:\{\}\[\]"]*}$`), ""),
					validation.StringLenBetween(2, 100),
				),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsStorageGatewaySmbFileShareCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.CreateSMBFileShareInput{
		Authentication:       aws.String(d.Get("authentication").(string)),
		ClientToken:          aws.String(resource.UniqueId()),
		DefaultStorageClass:  aws.String(d.Get("default_storage_class").(string)),
		GatewayARN:           aws.String(d.Get("gateway_arn").(string)),
		GuessMIMETypeEnabled: aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
		InvalidUserList:      expandStringSet(d.Get("invalid_user_list").(*schema.Set)),
		KMSEncrypted:         aws.Bool(d.Get("kms_encrypted").(bool)),
		LocationARN:          aws.String(d.Get("location_arn").(string)),
		ObjectACL:            aws.String(d.Get("object_acl").(string)),
		ReadOnly:             aws.Bool(d.Get("read_only").(bool)),
		RequesterPays:        aws.Bool(d.Get("requester_pays").(bool)),
		Role:                 aws.String(d.Get("role_arn").(string)),
		CaseSensitivity:      aws.String(d.Get("case_sensitivity").(string)),
		ValidUserList:        expandStringSet(d.Get("valid_user_list").(*schema.Set)),
		AdminUserList:        expandStringSet(d.Get("admin_user_list").(*schema.Set)),
		Tags:                 keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().StoragegatewayTags(),
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_share_name"); ok {
		input.FileShareName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("smb_acl_enabled"); ok {
		input.SMBACLEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("access_based_enumeration"); ok {
		input.AccessBasedEnumeration = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandStorageGatewayNfsFileShareCacheAttributes(v.([]interface{}))
	}

	if v, ok := d.GetOk("notification_policy"); ok {
		input.NotificationPolicy = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Storage Gateway SMB File Share: %#v", input)
	output, err := conn.CreateSMBFileShare(input)
	if err != nil {
		return fmt.Errorf("error creating Storage Gateway SMB File Share: %w", err)
	}

	d.SetId(aws.StringValue(output.FileShareARN))

	if _, err = waiter.SmbFileShareAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway SMB File Share (%q) to be Available: %w", d.Id(), err)
	}

	return resourceAwsStorageGatewaySmbFileShareRead(d, meta)
}

func resourceAwsStorageGatewaySmbFileShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &storagegateway.DescribeSMBFileSharesInput{
		FileShareARNList: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway SMB File Share: %#v", input)
	output, err := conn.DescribeSMBFileShares(input)
	if err != nil {
		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
			log.Printf("[WARN] Storage Gateway SMB File Share %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Storage Gateway SMB File Share: %w", err)
	}

	if output == nil || len(output.SMBFileShareInfoList) == 0 || output.SMBFileShareInfoList[0] == nil {
		log.Printf("[WARN] Storage Gateway SMB File Share %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	fileshare := output.SMBFileShareInfoList[0]

	arn := fileshare.FileShareARN
	d.Set("arn", arn)
	d.Set("authentication", fileshare.Authentication)
	d.Set("default_storage_class", fileshare.DefaultStorageClass)
	d.Set("fileshare_id", fileshare.FileShareId)
	d.Set("file_share_name", fileshare.FileShareName)
	d.Set("gateway_arn", fileshare.GatewayARN)
	d.Set("guess_mime_type_enabled", fileshare.GuessMIMETypeEnabled)
	d.Set("case_sensitivity", fileshare.CaseSensitivity)

	if err := d.Set("invalid_user_list", flattenStringSet(fileshare.InvalidUserList)); err != nil {
		return fmt.Errorf("error setting invalid_user_list: %w", err)
	}

	if err := d.Set("cache_attributes", flattenStorageGatewayNfsFileShareCacheAttributes(fileshare.CacheAttributes)); err != nil {
		return fmt.Errorf("error setting cache_attributes: %w", err)
	}

	d.Set("kms_encrypted", fileshare.KMSEncrypted)
	d.Set("kms_key_arn", fileshare.KMSKey)
	d.Set("location_arn", fileshare.LocationARN)
	d.Set("object_acl", fileshare.ObjectACL)
	d.Set("path", fileshare.Path)
	d.Set("read_only", fileshare.ReadOnly)
	d.Set("requester_pays", fileshare.RequesterPays)
	d.Set("role_arn", fileshare.Role)
	d.Set("audit_destination_arn", fileshare.AuditDestinationARN)
	d.Set("smb_acl_enabled", fileshare.SMBACLEnabled)
	d.Set("access_based_enumeration", fileshare.AccessBasedEnumeration)
	d.Set("notification_policy", fileshare.NotificationPolicy)

	if err := d.Set("valid_user_list", flattenStringSet(fileshare.ValidUserList)); err != nil {
		return fmt.Errorf("error setting valid_user_list: %w", err)
	}

	if err := d.Set("admin_user_list", flattenStringSet(fileshare.AdminUserList)); err != nil {
		return fmt.Errorf("error setting admin_user_list: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.StoragegatewayKeyValueTags(fileshare.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsStorageGatewaySmbFileShareUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	if d.HasChanges("admin_user_list", "default_storage_class", "guess_mime_type_enabled", "invalid_user_list",
		"kms_encrypted", "object_acl", "read_only", "requester_pays", "requester_pays",
		"valid_user_list", "kms_key_arn", "audit_destination_arn", "smb_acl_enabled", "cache_attributes",
		"case_sensitivity", "file_share_name", "notification_policy", "access_based_enumeration") {
		input := &storagegateway.UpdateSMBFileShareInput{
			DefaultStorageClass:    aws.String(d.Get("default_storage_class").(string)),
			FileShareARN:           aws.String(d.Id()),
			GuessMIMETypeEnabled:   aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
			InvalidUserList:        expandStringSet(d.Get("invalid_user_list").(*schema.Set)),
			KMSEncrypted:           aws.Bool(d.Get("kms_encrypted").(bool)),
			ObjectACL:              aws.String(d.Get("object_acl").(string)),
			ReadOnly:               aws.Bool(d.Get("read_only").(bool)),
			RequesterPays:          aws.Bool(d.Get("requester_pays").(bool)),
			ValidUserList:          expandStringSet(d.Get("valid_user_list").(*schema.Set)),
			AdminUserList:          expandStringSet(d.Get("admin_user_list").(*schema.Set)),
			SMBACLEnabled:          aws.Bool(d.Get("smb_acl_enabled").(bool)),
			CaseSensitivity:        aws.String(d.Get("case_sensitivity").(string)),
			AccessBasedEnumeration: aws.Bool(d.Get("access_based_enumeration").(bool)),
		}

		if v, ok := d.GetOk("kms_key_arn"); ok {
			input.KMSKey = aws.String(v.(string))
		}

		if v, ok := d.GetOk("notification_policy"); ok {
			input.NotificationPolicy = aws.String(v.(string))
		}

		if v, ok := d.GetOk("audit_destination_arn"); ok {
			input.AuditDestinationARN = aws.String(v.(string))
		}

		if v, ok := d.GetOk("file_share_name"); ok {
			input.FileShareName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandStorageGatewayNfsFileShareCacheAttributes(v.([]interface{}))
		}

		log.Printf("[DEBUG] Updating Storage Gateway SMB File Share: %#v", input)
		_, err := conn.UpdateSMBFileShare(input)
		if err != nil {
			return fmt.Errorf("error updating Storage Gateway SMB File Share: %w", err)
		}

		if _, err = waiter.SmbFileShareAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway SMB File Share (%q) to be Available: %w", d.Id(), err)
		}
	}

	return resourceAwsStorageGatewaySmbFileShareRead(d, meta)
}

func resourceAwsStorageGatewaySmbFileShareDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DeleteFileShareInput{
		FileShareARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway SMB File Share: %#v", input)
	_, err := conn.DeleteFileShare(input)
	if err != nil {
		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
			return nil
		}
		return fmt.Errorf("error deleting Storage Gateway SMB File Share: %w", err)
	}

	if _, err = waiter.SmbFileShareDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway SMB File Share (%q) to be Deleted: %w", d.Id(), err)
	}
	if err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error waiting for Storage Gateway SMB File Share deletion: %w", err)
	}

	return nil
}
