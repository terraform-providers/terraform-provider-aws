package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDataSyncLocationNfs() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataSyncLocationNfsCreate,
		Read:   resourceAwsDataSyncLocationNfsRead,
		Update: resourceAwsDataSyncLocationNfsUpdate,
		Delete: resourceAwsDataSyncLocationNfsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"on_prem_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"agent_arns": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateArn,
							},
						},
					},
				},
			},
			"mount_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:         schema.TypeString,
							Default:      datasync.NfsVersionAutomatic,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(datasync.NfsVersion_Values(), false),
						},
					},
				},
			},
			"server_hostname": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
				// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDataSyncLocationNfsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationNfsInput{
		OnPremConfig:   expandDataSyncOnPremConfig(d.Get("on_prem_config").([]interface{})),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           tags.IgnoreAws().DatasyncTags(),
	}

	if v, ok := d.GetOk("mount_options"); ok {
		input.MountOptions = expandDataSyncNfsMountOptions(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DataSync Location NFS: %s", input)
	output, err := conn.CreateLocationNfs(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location NFS: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceAwsDataSyncLocationNfsRead(d, meta)
}

func resourceAwsDataSyncLocationNfsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeLocationNfsInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location NFS: %s", input)
	output, err := conn.DescribeLocationNfs(input)

	if isAWSErr(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location NFS %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location NFS (%s): %w", d.Id(), err)
	}

	subdirectory, err := dataSyncParseLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return fmt.Errorf("error parsing Location NFS (%s) URI (%s): %w", d.Id(), aws.StringValue(output.LocationUri), err)
	}

	d.Set("arn", output.LocationArn)

	if err := d.Set("on_prem_config", flattenDataSyncOnPremConfig(output.OnPremConfig)); err != nil {
		return fmt.Errorf("error setting on_prem_config: %w", err)
	}

	if err := d.Set("mount_options", flattenDataSyncNfsMountOptions(output.MountOptions)); err != nil {
		return fmt.Errorf("error setting mount_options: %w", err)
	}

	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	tags, err := keyvaluetags.DatasyncListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location NFS (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsDataSyncLocationNfsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationNfsInput{
			LocationArn:  aws.String(d.Id()),
			OnPremConfig: expandDataSyncOnPremConfig(d.Get("on_prem_config").([]interface{})),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
		}

		if v, ok := d.GetOk("mount_options"); ok {
			input.MountOptions = expandDataSyncNfsMountOptions(v.([]interface{}))
		}

		_, err := conn.UpdateLocationNfs(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Location NFS (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location NFS (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsDataSyncLocationNfsRead(d, meta)
}

func resourceAwsDataSyncLocationNfsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location NFS: %s", input)
	_, err := conn.DeleteLocation(input)

	if isAWSErr(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location NFS (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDataSyncNfsMountOptions(l []interface{}) *datasync.NfsMountOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	nfsMountOptions := &datasync.NfsMountOptions{
		Version: aws.String(m["version"].(string)),
	}

	return nfsMountOptions
}

func flattenDataSyncNfsMountOptions(mountOptions *datasync.NfsMountOptions) []interface{} {
	if mountOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"version": aws.StringValue(mountOptions.Version),
	}

	return []interface{}{m}
}
