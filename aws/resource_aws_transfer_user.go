package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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

			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferUserName,
			},
		},
	}
}

func resourceAwsTransferUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
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

	if attr, ok := d.GetOk("policy"); ok {
		createOpts.Policy = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("tags"); ok {
		createOpts.Tags = keyvaluetags.New(attr.(map[string]interface{})).IgnoreAws().TransferTags()
	}

	log.Printf("[DEBUG] Create Transfer User Option: %#v", createOpts)

	_, err := conn.CreateUser(createOpts)
	if err != nil {
		return fmt.Errorf("error creating Transfer User: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serverID, userName))

	return resourceAwsTransferUserRead(d, meta)
}

func resourceAwsTransferUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	serverID, userName, err := decodeTransferUserId(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %s", err)
	}

	descOpts := &transfer.DescribeUserInput{
		UserName: aws.String(userName),
		ServerId: aws.String(serverID),
	}

	log.Printf("[DEBUG] Describe Transfer User Option: %#v", descOpts)

	resp, err := conn.DescribeUser(descOpts)
	if err != nil {
		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Transfer User (%s) for Server (%s) not found, removing from state", userName, serverID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Transfer User (%s): %s", d.Id(), err)
	}

	d.Set("server_id", resp.ServerId)
	d.Set("user_name", resp.User.UserName)
	d.Set("arn", resp.User.Arn)
	d.Set("home_directory", resp.User.HomeDirectory)
	d.Set("home_directory_type", resp.User.HomeDirectoryType)
	d.Set("policy", resp.User.Policy)
	d.Set("role", resp.User.Role)

	if err := d.Set("home_directory_mappings", flattenAwsTransferHomeDirectoryMappings(resp.User.HomeDirectoryMappings)); err != nil {
		return fmt.Errorf("Error setting home_directory_mappings: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.TransferKeyValueTags(resp.User.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}
	return nil
}

func resourceAwsTransferUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	updateFlag := false
	serverID, userName, err := decodeTransferUserId(d.Id())
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
			return fmt.Errorf("error updating Transfer User (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.TransferUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsTransferUserRead(d, meta)
}

func resourceAwsTransferUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	serverID, userName, err := decodeTransferUserId(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %s", err)
	}

	delOpts := &transfer.DeleteUserInput{
		UserName: aws.String(userName),
		ServerId: aws.String(serverID),
	}

	log.Printf("[DEBUG] Delete Transfer User Option: %#v", delOpts)

	_, err = conn.DeleteUser(delOpts)
	if err != nil {
		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Transfer User (%s) for Server(%s): %s", userName, serverID, err)
	}

	if err := waitForTransferUserDeletion(conn, serverID, userName); err != nil {
		return fmt.Errorf("error waiting for Transfer User (%s) for Server (%s): %s", userName, serverID, err)
	}

	return nil
}

func decodeTransferUserId(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected SERVERID/USERNAME", id)
	}
	return idParts[0], idParts[1], nil
}

func waitForTransferUserDeletion(conn *transfer.Transfer, serverID, userName string) error {
	params := &transfer.DescribeUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeUser(params)

		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("Transfer User (%s) for Server (%s) still exists", userName, serverID))
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DescribeUser(params)
	}
	if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error decoding transfer user ID: %s", err)
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
