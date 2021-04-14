package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/codebuild/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/codebuild/waiter"
)

func resourceAwsCodeBuildReportGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildReportGroupCreate,
		Read:   resourceAwsCodeBuildReportGroupRead,
		Update: resourceAwsCodeBuildReportGroupUpdate,
		Delete: resourceAwsCodeBuildReportGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codebuild.ReportType_Values(), false),
			},
			"export_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ReportExportConfigType_Values(), false),
						},
						"s3_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Required: true,
									},
									"encryption_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"encryption_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"packaging": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      codebuild.ReportPackagingTypeNone,
										ValidateFunc: validation.StringInSlice(codebuild.ReportPackagingType_Values(), false),
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_reports": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCodeBuildReportGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn
	createOpts := &codebuild.CreateReportGroupInput{
		Name:         aws.String(d.Get("name").(string)),
		Type:         aws.String(d.Get("type").(string)),
		ExportConfig: expandAwsCodeBuildReportGroupExportConfig(d.Get("export_config").([]interface{})),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CodebuildTags(),
	}

	resp, err := conn.CreateReportGroup(createOpts)
	if err != nil {
		return fmt.Errorf("error creating CodeBuild Report Group: %w", err)
	}

	d.SetId(aws.StringValue(resp.ReportGroup.Arn))

	return resourceAwsCodeBuildReportGroupRead(d, meta)
}

func resourceAwsCodeBuildReportGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	reportGroup, err := finder.ReportGroupByArn(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error Listing CodeBuild Report Groups: %w", err)
	}

	if reportGroup == nil {
		log.Printf("[WARN] CodeBuild Report Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", reportGroup.Arn)
	d.Set("type", reportGroup.Type)
	d.Set("name", reportGroup.Name)

	if err := d.Set("created", reportGroup.Created.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting created: %w", err)
	}

	if err := d.Set("export_config", flattenAwsCodeBuildReportGroupExportConfig(reportGroup.ExportConfig)); err != nil {
		return fmt.Errorf("error setting export config: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.CodebuildKeyValueTags(reportGroup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsCodeBuildReportGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	input := &codebuild.UpdateReportGroupInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("export_config") {
		input.ExportConfig = expandAwsCodeBuildReportGroupExportConfig(d.Get("export_config").([]interface{}))
	}

	if d.HasChange("tags") {
		input.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CodebuildTags()
	}

	_, err := conn.UpdateReportGroup(input)
	if err != nil {
		return fmt.Errorf("error updating CodeBuild Report Group: %w", err)
	}

	return resourceAwsCodeBuildReportGroupRead(d, meta)
}

func resourceAwsCodeBuildReportGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	deleteOpts := &codebuild.DeleteReportGroupInput{
		Arn:           aws.String(d.Id()),
		DeleteReports: aws.Bool(d.Get("delete_reports").(bool)),
	}

	if _, err := conn.DeleteReportGroup(deleteOpts); err != nil {
		return fmt.Errorf("error deleting CodeBuild Report Group (%s): %w", d.Id(), err)
	}

	if _, err := waiter.ReportGroupDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error while waiting for CodeBuild Report Group (%s) to become deleted: %w", d.Id(), err)
	}

	return nil
}

func expandAwsCodeBuildReportGroupExportConfig(config []interface{}) *codebuild.ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	exportConfig := &codebuild.ReportExportConfig{}

	if v, ok := s["type"]; ok {
		exportConfig.ExportConfigType = aws.String(v.(string))
	}

	if v, ok := s["s3_destination"]; ok {
		exportConfig.S3Destination = expandAwsCodeBuildReportGroupS3ReportExportConfig(v.([]interface{}))
	}

	return exportConfig
}

func flattenAwsCodeBuildReportGroupExportConfig(config *codebuild.ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["s3_destination"] = flattenAwsCodeBuildReportGroupS3ReportExportConfig(config.S3Destination)
	settings["type"] = aws.StringValue(config.ExportConfigType)

	return []map[string]interface{}{settings}
}

func expandAwsCodeBuildReportGroupS3ReportExportConfig(config []interface{}) *codebuild.S3ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	s3ReportExportConfig := &codebuild.S3ReportExportConfig{}

	if v, ok := s["bucket"]; ok {
		s3ReportExportConfig.Bucket = aws.String(v.(string))
	}
	if v, ok := s["encryption_disabled"]; ok {
		s3ReportExportConfig.EncryptionDisabled = aws.Bool(v.(bool))
	}

	if v, ok := s["encryption_key"]; ok {
		s3ReportExportConfig.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := s["packaging"]; ok {
		s3ReportExportConfig.Packaging = aws.String(v.(string))
	}

	if v, ok := s["path"]; ok {
		s3ReportExportConfig.Path = aws.String(v.(string))
	}

	return s3ReportExportConfig
}

func flattenAwsCodeBuildReportGroupS3ReportExportConfig(config *codebuild.S3ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["path"] = aws.StringValue(config.Path)
	settings["bucket"] = aws.StringValue(config.Bucket)
	settings["packaging"] = aws.StringValue(config.Packaging)
	settings["encryption_disabled"] = aws.BoolValue(config.EncryptionDisabled)

	if config.EncryptionKey != nil {
		settings["encryption_key"] = aws.StringValue(config.EncryptionKey)
	}

	return []map[string]interface{}{settings}
}
