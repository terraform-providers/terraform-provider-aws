package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGreengrassConnectorDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassConnectorDefinitionCreate,
		Read:   resourceAwsGreengrassConnectorDefinitionRead,
		Update: resourceAwsGreengrassConnectorDefinitionUpdate,
		Delete: resourceAwsGreengrassConnectorDefinitionDelete,

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
			"connector_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connector_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
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

func convertInterfaceMapToStringMap(interfaceMap map[string]interface{}) map[string]*string {
	stringMap := make(map[string]*string)
	for k, v := range interfaceMap {
		strVal := v.(string)
		stringMap[k] = &strVal
	}
	return stringMap
}

func createConnectorDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("connector_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateConnectorDefinitionVersionInput{
		ConnectorDefinitionId: aws.String(d.Id()),
	}

	if v := d.Get("amzn_client_token").(string); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	connectors := make([]*greengrass.Connector, 0)
	for _, connectorToCast := range rawData["connector"].(*schema.Set).List() {
		rawConnector := connectorToCast.(map[string]interface{})
		connector := &greengrass.Connector{
			ConnectorArn: aws.String(rawConnector["connector_arn"].(string)),
			Id:           aws.String(rawConnector["id"].(string)),
			Parameters:   convertInterfaceMapToStringMap(rawConnector["parameters"].(map[string]interface{})),
		}
		connectors = append(connectors, connector)
	}
	params.Connectors = connectors

	log.Printf("[DEBUG] Creating Greengrass Connector Definition Version: %s", params)
	_, err := conn.CreateConnectorDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassConnectorDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateConnectorDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if rawTags := d.Get("tags").(map[string]interface{}); len(rawTags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GreengrassTags()
	}

	log.Printf("[DEBUG] Creating Greengrass Connector Definition: %s", params)
	out, err := conn.CreateConnectorDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createConnectorDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassConnectorDefinitionRead(d, meta)
}

func setConnectorDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetConnectorDefinitionVersionInput{
		ConnectorDefinitionId:        aws.String(d.Id()),
		ConnectorDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetConnectorDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	d.Set("latest_definition_version_arn", *out.Arn)

	rawConnectorList := make([]map[string]interface{}, 0)
	for _, connector := range out.Definition.Connectors {
		rawConnector := make(map[string]interface{})
		rawConnector["connector_arn"] = *connector.ConnectorArn
		rawConnector["id"] = *connector.Id

		parameters := make(map[string]string)
		for k, v := range connector.Parameters {
			parameters[k] = *v
		}
		rawConnector["parameters"] = parameters

		rawConnectorList = append(rawConnectorList, rawConnector)
	}

	rawVersion["connector"] = rawConnectorList

	d.Set("connector_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassConnectorDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetConnectorDefinitionInput{
		ConnectorDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Connector Definition: %s", params)
	out, err := conn.GetConnectorDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Connector Definition: %s", out)

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
		err = setConnectorDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassConnectorDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateConnectorDefinitionInput{
		Name:                  aws.String(d.Get("name").(string)),
		ConnectorDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateConnectorDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("connector_definition_version") {
		err = createConnectorDefinitionVersion(d, conn)
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

	return resourceAwsGreengrassConnectorDefinitionRead(d, meta)
}

func resourceAwsGreengrassConnectorDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteConnectorDefinitionInput{
		ConnectorDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Connector Definition: %s", params)

	_, err := conn.DeleteConnectorDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
