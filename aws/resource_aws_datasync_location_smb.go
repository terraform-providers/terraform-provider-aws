package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDataSyncLocationSmb() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataSyncLocationSmbCreate,
		Read:   resourceAwsDataSyncLocationSmbRead,
		Update: resourceAwsDataSyncLocationSmbUpdate,
		Delete: resourceAwsDataSyncLocationSmbDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"domain": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
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
							Default:      datasync.SmbVersionAutomatic,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.SmbVersion_Values(), false),
						},
					},
				},
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 104),
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
				/*// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
				*/
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDataSyncLocationSmbCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationSmbInput{
		AgentArns:      expandStringSet(d.Get("agent_arns").(*schema.Set)),
		MountOptions:   expandDataSyncSmbMountOptions(d.Get("mount_options").([]interface{})),
		Password:       aws.String(d.Get("password").(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           tags.IgnoreAws().DatasyncTags(),
		User:           aws.String(d.Get("user").(string)),
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location SMB: %s", input)
	output, err := conn.CreateLocationSmb(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location SMB: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceAwsDataSyncLocationSmbRead(d, meta)
}

func resourceAwsDataSyncLocationSmbRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeLocationSmbInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location SMB: %s", input)
	output, err := conn.DescribeLocationSmb(input)

	if isAWSErr(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location SMB %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location SMB (%s): %w", d.Id(), err)
	}

	tagsInput := &datasync.ListTagsForResourceInput{
		ResourceArn: output.LocationArn,
	}

	log.Printf("[DEBUG] Reading DataSync Location SMB tags: %s", tagsInput)
	tagsOutput, err := conn.ListTagsForResource(tagsInput)

	if err != nil {
		return fmt.Errorf("error reading DataSync Location SMB (%s) tags: %w", d.Id(), err)
	}

	subdirectory, err := dataSyncParseLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return fmt.Errorf("error parsing Location SMB (%s) URI (%s): %w", d.Id(), aws.StringValue(output.LocationUri), err)
	}

	d.Set("agent_arns", flattenStringSet(output.AgentArns))

	d.Set("arn", output.LocationArn)

	d.Set("domain", output.Domain)

	if err := d.Set("mount_options", flattenDataSyncSmbMountOptions(output.MountOptions)); err != nil {
		return fmt.Errorf("error setting mount_options: %w", err)
	}

	d.Set("subdirectory", subdirectory)

	tags := keyvaluetags.DatasyncKeyValueTags(tagsOutput.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("user", output.User)

	d.Set("uri", output.LocationUri)

	return nil
}

func resourceAwsDataSyncLocationSmbUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationSmbInput{
			LocationArn:  aws.String(d.Id()),
			AgentArns:    expandStringSet(d.Get("agent_arns").(*schema.Set)),
			MountOptions: expandDataSyncSmbMountOptions(d.Get("mount_options").([]interface{})),
			Password:     aws.String(d.Get("password").(string)),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
			User:         aws.String(d.Get("user").(string)),
		}

		if v, ok := d.GetOk("domain"); ok {
			input.Domain = aws.String(v.(string))
		}

		_, err := conn.UpdateLocationSmb(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Location SMB (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Datasync SMB location (%s) tags: %w", d.Id(), err)
		}
	}
	return resourceAwsDataSyncLocationSmbRead(d, meta)
}

func resourceAwsDataSyncLocationSmbDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location SMB: %s", input)
	_, err := conn.DeleteLocation(input)

	if isAWSErr(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location SMB (%s): %w", d.Id(), err)
	}

	return nil
}
