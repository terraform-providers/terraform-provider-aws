package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsLexSlotType() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexSlotTypeRead,

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enumeration_value": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"synonyms": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
				),
			},
			"value_selection_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  LexSlotTypeVersionLatest,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`\$LATEST|[0-9]+`), ""),
				),
			},
		},
	}
}

func dataSourceAwsLexSlotTypeRead(d *schema.ResourceData, meta interface{}) error {
	slotTypeName := d.Get("name").(string)

	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(slotTypeName),
		Version: aws.String(d.Get("version").(string)),
	})
	if err != nil {
		return fmt.Errorf("error getting slot type %s: %w", slotTypeName, err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("enumeration_value", flattenLexEnumerationValues(resp.EnumerationValues))
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)
	d.Set("value_selection_strategy", resp.ValueSelectionStrategy)
	d.Set("version", resp.Version)

	d.SetId(slotTypeName)

	return nil
}
