package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"mount_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Default:  datasync.SmbVersionAutomatic,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.SmbVersionAutomatic,
								datasync.SmbVersionSmb2,
								datasync.SmbVersionSmb3,
							}, false),
						},
					},
				},
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"server_hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subdirectory": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"tags": tagsSchema(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDataSyncLocationSmbCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.CreateLocationSmbInput{
		AgentArns:      expandStringSet(d.Get("agent_arns").(*schema.Set)),
		MountOptions:   expandDataSyncSmbMountOptions(d.Get("mount_options").([]interface{})),
		Password:       aws.String(d.Get("password").(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatasyncTags(),
		User:           aws.String(d.Get("user").(string)),
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location SMB: %s", input)
	output, err := conn.CreateLocationSmb(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location SMB: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceAwsDataSyncLocationSmbRead(d, meta)
}

func resourceAwsDataSyncLocationSmbRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
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
		return fmt.Errorf("error reading DataSync Location SMB (%s): %s", d.Id(), err)
	}

	tagsInput := &datasync.ListTagsForResourceInput{
		ResourceArn: output.LocationArn,
	}

	log.Printf("[DEBUG] Reading DataSync Location SMB tags: %s", tagsInput)
	tagsOutput, err := conn.ListTagsForResource(tagsInput)

	if err != nil {
		return fmt.Errorf("error reading DataSync Location SMB (%s) tags: %s", d.Id(), err)
	}

	subdirectory, err := dataSyncParseLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return fmt.Errorf("error parsing Location SMB (%s) URI (%s): %s", d.Id(), aws.StringValue(output.LocationUri), err)
	}

	d.Set("agent_arns", schema.NewSet(schema.HashString, flattenStringList(output.AgentArns)))

	d.Set("arn", output.LocationArn)

	d.Set("domain", output.Domain)

	if err := d.Set("mount_options", flattenDataSyncSmbMountOptions(output.MountOptions)); err != nil {
		return fmt.Errorf("error setting mount_options: %s", err)
	}

	d.Set("subdirectory", subdirectory)

	if err := d.Set("tags", keyvaluetags.DatasyncKeyValueTags(tagsOutput.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("user", output.User)

	d.Set("uri", output.LocationUri)

	return nil
}

func resourceAwsDataSyncLocationSmbUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Datasync SMB location (%s) tags: %s", d.Id(), err)
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
		return fmt.Errorf("error deleting DataSync Location SMB (%s): %s", d.Id(), err)
	}

	return nil
}
