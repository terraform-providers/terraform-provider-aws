package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGreengrassLoggerDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassLoggerDefinitionCreate,
		Read:   resourceAwsGreengrassLoggerDefinitionRead,
		Update: resourceAwsGreengrassLoggerDefinitionUpdate,
		Delete: resourceAwsGreengrassLoggerDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amzn_client_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
			"latest_definition_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logger_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logger": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"component": {
										Type:     schema.TypeString,
										Required: true,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"level": {
										Type:     schema.TypeString,
										Required: true,
									},
									"space": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createLoggerDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("logger_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateLoggerDefinitionVersionInput{
		LoggerDefinitionId: aws.String(d.Id()),
	}

	if v := d.Get("amzn_client_token").(string); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	loggers := make([]*greengrass.Logger, 0)
	for _, loggerToCast := range rawData["logger"].(*schema.Set).List() {
		rawLogger := loggerToCast.(map[string]interface{})
		logger := &greengrass.Logger{
			Component: aws.String(rawLogger["component"].(string)),
			Id:        aws.String(rawLogger["id"].(string)),
			Level:     aws.String(rawLogger["level"].(string)),
			Type:      aws.String(rawLogger["type"].(string)),
		}
		if space, ok := rawLogger["space"]; ok {
			logger.Space = aws.Int64(int64(space.(int)))
		}

		loggers = append(loggers, logger)
	}
	params.Loggers = loggers

	log.Printf("[DEBUG] Creating Greengrass Logger Definition Version: %s", params)
	_, err := conn.CreateLoggerDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassLoggerDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateLoggerDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GreengrassTags()
	}

	log.Printf("[DEBUG] Creating Greengrass Logger Definition: %s", params)
	out, err := conn.CreateLoggerDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createLoggerDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassLoggerDefinitionRead(d, meta)
}

func setLoggerDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetLoggerDefinitionVersionInput{
		LoggerDefinitionId:        aws.String(d.Id()),
		LoggerDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetLoggerDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	d.Set("latest_definition_version_arn", *out.Arn)

	rawLoggerList := make([]map[string]interface{}, 0)
	for _, logger := range out.Definition.Loggers {
		rawLogger := make(map[string]interface{})
		rawLogger["component"] = *logger.Component
		rawLogger["id"] = *logger.Id
		rawLogger["level"] = *logger.Level
		rawLogger["type"] = *logger.Type

		if logger.Space != nil {
			rawLogger["space"] = *logger.Space
		}

		rawLoggerList = append(rawLoggerList, rawLogger)
	}

	rawVersion["logger"] = rawLoggerList

	d.Set("logger_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassLoggerDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetLoggerDefinitionInput{
		LoggerDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Logger Definition: %s", params)
	out, err := conn.GetLoggerDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Logger Definition: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	arn := *out.Arn
	tags, err := keyvaluetags.GreengrassListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if out.LatestVersion != nil {
		err = setLoggerDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassLoggerDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateLoggerDefinitionInput{
		Name:               aws.String(d.Get("name").(string)),
		LoggerDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateLoggerDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("logger_definition_version") {
		err = createLoggerDefinitionVersion(d, conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GreengrassUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGreengrassLoggerDefinitionRead(d, meta)
}

func resourceAwsGreengrassLoggerDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteLoggerDefinitionInput{
		LoggerDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Logger Definition: %s", params)

	_, err := conn.DeleteLoggerDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
