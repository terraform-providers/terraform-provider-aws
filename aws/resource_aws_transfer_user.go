package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tftransfer "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsTransferUser() *schema.Resource {

	return &schema.Resource{
		Create: resourceAwsTransferUserCreate,
		Read:   resourceAwsTransferUserRead,
		Update: resourceAwsTransferUserUpdate,
		Delete: resourceAwsTransferUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},

			"home_directory_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.HomeDirectoryTypePath,
				ValidateFunc: validation.StringInSlice([]string{transfer.HomeDirectoryTypePath, transfer.HomeDirectoryTypeLogical}, false),
			},

			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"posix_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
						},
					},
				},
			},

			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferServerID,
			},

			"tags": tagsSchema(),

			"tags_all": tagsSchemaComputed(),

			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferUserName,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsTransferUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))
	userName := d.Get("user_name").(string)
	serverID := d.Get("server_id").(string)

	createOpts := &transfer.CreateUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
		Role:     aws.String(d.Get("role").(string)),
	}

	if attr, ok := d.GetOk("home_directory"); ok {
		createOpts.HomeDirectory = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("home_directory_type"); ok {
		createOpts.HomeDirectoryType = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("home_directory_mappings"); ok {
		createOpts.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("posix_profile"); ok {
		createOpts.PosixProfile = expandTransferUserPosixUser(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("policy"); ok {
		createOpts.Policy = aws.String(attr.(string))
	}

	if len(tags) > 0 {
		createOpts.Tags = tags.IgnoreAws().TransferTags()
	}

	log.Printf("[DEBUG] Create Transfer User Option: %#v", createOpts)

	_, err := conn.CreateUser(createOpts)
	if err != nil {
		return fmt.Errorf("error creating Transfer User: %s", err)
	}

	d.SetId(tftransfer.UserCreateResourceID(serverID, userName))

	return resourceAwsTransferUserRead(d, meta)
}

func resourceAwsTransferUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	serverID, userName, err := tftransfer.UserParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %s", err)
	}

	resp, err := finder.UserByID(conn, serverID, userName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Transfer User (%s): %w", d.Id(), err)
	}

	user := resp.User
	d.Set("server_id", resp.ServerId)
	d.Set("user_name", user.UserName)
	d.Set("arn", user.Arn)
	d.Set("home_directory", user.HomeDirectory)
	d.Set("home_directory_type", user.HomeDirectoryType)
	d.Set("policy", user.Policy)
	d.Set("role", user.Role)

	if err := d.Set("home_directory_mappings", flattenAwsTransferHomeDirectoryMappings(user.HomeDirectoryMappings)); err != nil {
		return fmt.Errorf("Error setting home_directory_mappings: %w", err)
	}

	if err := d.Set("posix_profile", flattenTransferUserPosixUser(user.PosixProfile)); err != nil {
		return fmt.Errorf("Error setting posix_profile: %w", err)
	}
	tags := keyvaluetags.TransferKeyValueTags(user.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}
	return nil
}

func resourceAwsTransferUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	updateFlag := false
	serverID, userName, err := tftransfer.UserParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %s", err)
	}

	updateOpts := &transfer.UpdateUserInput{
		UserName: aws.String(userName),
		ServerId: aws.String(serverID),
	}

	if d.HasChange("home_directory") {
		updateOpts.HomeDirectory = aws.String(d.Get("home_directory").(string))
		updateFlag = true
	}

	if d.HasChange("home_directory_mappings") {
		updateOpts.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(d.Get("home_directory_mappings").([]interface{}))
		updateFlag = true
	}

	if d.HasChange("posix_profile") {
		updateOpts.PosixProfile = expandTransferUserPosixUser(d.Get("posix_profile").([]interface{}))
		updateFlag = true
	}

	if d.HasChange("home_directory_type") {
		updateOpts.HomeDirectoryType = aws.String(d.Get("home_directory_type").(string))
		updateFlag = true
	}

	if d.HasChange("policy") {
		updateOpts.Policy = aws.String(d.Get("policy").(string))
		updateFlag = true
	}

	if d.HasChange("role") {
		updateOpts.Role = aws.String(d.Get("role").(string))
		updateFlag = true
	}

	if updateFlag {
		_, err := conn.UpdateUser(updateOpts)
		if err != nil {
			if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
				log.Printf("[WARN] Transfer User (%s) for Server (%s) not found, removing from state", userName, serverID)
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error updating Transfer User (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.TransferUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsTransferUserRead(d, meta)
}

func resourceAwsTransferUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	serverID, userName, err := tftransfer.UserParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %w", err)
	}

	return transferUserDelete(conn, serverID, userName)
}

// transferUserDelete attempts to delete a transfer user.
func transferUserDelete(conn *transfer.Transfer, serverID, userName string) error {
	id := fmt.Sprintf("%s/%s", serverID, userName)
	input := &transfer.DeleteUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	log.Printf("[INFO] Deleting Transfer User: %s", id)
	_, err := conn.DeleteUser(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Transfer User (%s): %w", id, err)
	}

	_, err = waiter.UserDeleted(conn, serverID, userName)

	if err != nil {
		return fmt.Errorf("error waiting for Transfer User (%s) delete: %w", id, err)
	}

	return nil
}

func expandAwsTransferHomeDirectoryMappings(in []interface{}) []*transfer.HomeDirectoryMapEntry {
	mappings := make([]*transfer.HomeDirectoryMapEntry, 0)

	for _, tConfig := range in {
		config := tConfig.(map[string]interface{})

		m := &transfer.HomeDirectoryMapEntry{
			Entry:  aws.String(config["entry"].(string)),
			Target: aws.String(config["target"].(string)),
		}

		mappings = append(mappings, m)
	}

	return mappings
}

func flattenAwsTransferHomeDirectoryMappings(mappings []*transfer.HomeDirectoryMapEntry) []interface{} {
	l := make([]interface{}, len(mappings))
	for i, m := range mappings {
		l[i] = map[string]interface{}{
			"entry":  aws.StringValue(m.Entry),
			"target": aws.StringValue(m.Target),
		}
	}
	return l
}

func expandTransferUserPosixUser(pUser []interface{}) *transfer.PosixProfile {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	posixUser := &transfer.PosixProfile{
		Gid: aws.Int64(int64(m["gid"].(int))),
		Uid: aws.Int64(int64(m["uid"].(int))),
	}

	if v, ok := m["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		posixUser.SecondaryGids = expandInt64Set(v)
	}

	return posixUser
}

func flattenTransferUserPosixUser(posixUser *transfer.PosixProfile) []interface{} {
	if posixUser == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"gid":            aws.Int64Value(posixUser.Gid),
		"uid":            aws.Int64Value(posixUser.Uid),
		"secondary_gids": aws.Int64ValueSlice(posixUser.SecondaryGids),
	}

	return []interface{}{m}
}
