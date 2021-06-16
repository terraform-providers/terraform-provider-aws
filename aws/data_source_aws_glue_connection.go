package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsGlueConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsGlueConnectionRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_properties": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"match_criteria": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"physical_connection_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsGlueConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).glueconn
	id := d.Get("id").(string)
	catalogID, connectionName, err := decodeGlueConnectionID(id)
	if err != nil {
		return diag.Errorf("error decoding Glue Connection %s: %s", id, err)
	}
	input := &glue.GetConnectionInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(connectionName),
	}
	output, err := conn.GetConnection(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return diag.Errorf("error Glue Connection (%s) not found", id)
		}
		return diag.Errorf("error reading Glue Connection (%s): %s", id, err)
	}

	connection := output.Connection
	d.SetId(id)
	d.Set("catalog_id", catalogID)
	d.Set("connection_type", connection.ConnectionType)
	d.Set("name", connection.Name)
	d.Set("description", connection.Description)

	connectionArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("connection/%s", connectionName),
	}.String()
	d.Set("arn", connectionArn)

	if err := d.Set("connection_properties", aws.StringValueMap(connection.ConnectionProperties)); err != nil {
		return diag.Errorf("error setting connection_properties: %s", err)
	}

	if err := d.Set("physical_connection_requirements", flattenGluePhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return diag.Errorf("error setting physical_connection_requirements: %s", err)
	}

	if err := d.Set("match_criteria", flattenStringList(connection.MatchCriteria)); err != nil {
		return diag.Errorf("error setting match_criteria: %s", err)
	}

	return nil
}
