package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsGlueCatalogDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogDatabaseCreate,
		Read:   resourceAwsGlueCatalogDatabaseRead,
		Update: resourceAwsGlueCatalogDatabaseUpdate,
		Delete: resourceAwsGlueCatalogDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexp.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"location_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"target_database": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsGlueCatalogDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	name := d.Get("name").(string)

	dbInput := &glue.DatabaseInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		dbInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("location_uri"); ok {
		dbInput.LocationUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		dbInput.Parameters = expandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("target_database"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		dbInput.TargetDatabase = expandGlueDatabaseTargetDatabase(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &glue.CreateDatabaseInput{
		CatalogId:     aws.String(catalogID),
		DatabaseInput: dbInput,
	}

	_, err := conn.CreateDatabase(input)
	if err != nil {
		return fmt.Errorf("Error creating Catalog Database: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, name))

	return resourceAwsGlueCatalogDatabaseRead(d, meta)
}

func resourceAwsGlueCatalogDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, name, err := readAwsGlueCatalogID(d.Id())
	if err != nil {
		return err
	}

	dbUpdateInput := &glue.UpdateDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	dbInput := &glue.DatabaseInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		dbInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("location_uri"); ok {
		dbInput.LocationUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		dbInput.Parameters = expandStringMap(v.(map[string]interface{}))
	}

	dbUpdateInput.DatabaseInput = dbInput

	if d.HasChanges("description", "location_uri", "parameters") {
		if _, err := conn.UpdateDatabase(dbUpdateInput); err != nil {
			return err
		}
	}

	return resourceAwsGlueCatalogDatabaseRead(d, meta)
}

func resourceAwsGlueCatalogDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, name, err := readAwsGlueCatalogID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.GetDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	out, err := conn.GetDatabase(input)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Catalog Database (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Glue Catalog Database: %s", err.Error())
	}

	database := out.Database
	databaseArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("database/%s", aws.StringValue(database.Name)),
	}.String()
	d.Set("arn", databaseArn)
	d.Set("name", database.Name)
	d.Set("catalog_id", database.CatalogId)
	d.Set("description", database.Description)
	d.Set("location_uri", database.LocationUri)
	d.Set("parameters", aws.StringValueMap(database.Parameters))

	if database.TargetDatabase != nil {
		if err := d.Set("target_database", []interface{}{flattenGlueDatabaseTargetDatabase(database.TargetDatabase)}); err != nil {
			return fmt.Errorf("error setting target_database: %w", err)
		}
	} else {
		d.Set("target_database", nil)
	}

	return nil
}

func resourceAwsGlueCatalogDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Glue Catalog Database: %s", d.Id())
	_, err := conn.DeleteDatabase(&glue.DeleteDatabaseInput{
		Name:      aws.String(d.Get("name").(string)),
		CatalogId: aws.String(d.Get("catalog_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Catalog Database: %w", err)
	}
	return nil
}

func readAwsGlueCatalogID(id string) (catalogID string, name string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected CATALOG-ID:DATABASE-NAME", id)
	}
	return idParts[0], idParts[1], nil
}

func createAwsGlueCatalogID(d *schema.ResourceData, accountid string) (catalogID string) {
	if rawCatalogID, ok := d.GetOkExists("catalog_id"); ok {
		catalogID = rawCatalogID.(string)
	} else {
		catalogID = accountid
	}
	return
}

func expandGlueDatabaseTargetDatabase(tfMap map[string]interface{}) *glue.DatabaseIdentifier {
	if tfMap == nil {
		return nil
	}

	apiObject := &glue.DatabaseIdentifier{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	return apiObject
}

func flattenGlueDatabaseTargetDatabase(apiObject *glue.DatabaseIdentifier) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	return tfMap
}
