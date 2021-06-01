package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfglue "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue"
)

func resourceAwsGluePartition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGluePartitionCreate,
		Read:   resourceAwsGluePartitionRead,
		Update: resourceAwsGluePartitionUpdate,
		Delete: resourceAwsGluePartitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"table_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"partition_values": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_descriptor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"comment": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"compressed": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"input_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"number_of_buckets": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"output_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ser_de_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"parameters": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"serialization_library": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"skewed_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"skewed_column_names": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_values": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_value_location_maps": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"sort_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"column": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sort_order": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"stored_as_sub_directories": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_analyzed_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_accessed_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGluePartitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	dbName := d.Get("database_name").(string)
	tableName := d.Get("table_name").(string)
	values := d.Get("partition_values").(*schema.Set)

	input := &glue.CreatePartitionInput{
		CatalogId:      aws.String(catalogID),
		DatabaseName:   aws.String(dbName),
		TableName:      aws.String(tableName),
		PartitionInput: expandGluePartitionInput(d),
	}

	log.Printf("[DEBUG] Creating Glue Partition: %#v", input)
	_, err := conn.CreatePartition(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Partition: %w", err)
	}

	d.SetId(tfglue.CreateAwsGluePartitionID(catalogID, dbName, tableName, values))

	return resourceAwsGluePartitionRead(d, meta)
}

func resourceAwsGluePartitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, values, err := tfglue.ReadAwsGluePartitionID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading Glue Partition: %s", d.Id())
	input := &glue.GetPartitionInput{
		CatalogId:       aws.String(catalogID),
		DatabaseName:    aws.String(dbName),
		TableName:       aws.String(tableName),
		PartitionValues: aws.StringSlice(values),
	}

	out, err := conn.GetPartition(input)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Partition (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading Glue Partition: %w", err)
	}

	partition := out.Partition

	d.Set("table_name", partition.TableName)
	d.Set("catalog_id", catalogID)
	d.Set("database_name", partition.DatabaseName)
	d.Set("partition_values", flattenStringSet(partition.Values))

	if partition.LastAccessTime != nil {
		d.Set("last_accessed_time", partition.LastAccessTime.Format(time.RFC3339))
	}

	if partition.LastAnalyzedTime != nil {
		d.Set("last_analyzed_time", partition.LastAnalyzedTime.Format(time.RFC3339))
	}

	if partition.CreationTime != nil {
		d.Set("creation_time", partition.CreationTime.Format(time.RFC3339))
	}

	if err := d.Set("storage_descriptor", flattenGlueStorageDescriptor(partition.StorageDescriptor)); err != nil {
		return fmt.Errorf("error setting storage_descriptor: %w", err)
	}

	if err := d.Set("parameters", aws.StringValueMap(partition.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %w", err)
	}

	return nil
}

func resourceAwsGluePartitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, values, err := tfglue.ReadAwsGluePartitionID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.UpdatePartitionInput{
		CatalogId:          aws.String(catalogID),
		DatabaseName:       aws.String(dbName),
		TableName:          aws.String(tableName),
		PartitionInput:     expandGluePartitionInput(d),
		PartitionValueList: aws.StringSlice(values),
	}

	log.Printf("[DEBUG] Updating Glue Partition: %#v", input)
	if _, err := conn.UpdatePartition(input); err != nil {
		return fmt.Errorf("error updating Glue Partition: %w", err)
	}

	return resourceAwsGluePartitionRead(d, meta)
}

func resourceAwsGluePartitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	catalogID, dbName, tableName, values, tableErr := tfglue.ReadAwsGluePartitionID(d.Id())
	if tableErr != nil {
		return tableErr
	}

	log.Printf("[DEBUG] Deleting Glue Partition: %s", d.Id())
	_, err := conn.DeletePartition(&glue.DeletePartitionInput{
		CatalogId:       aws.String(catalogID),
		TableName:       aws.String(tableName),
		DatabaseName:    aws.String(dbName),
		PartitionValues: aws.StringSlice(values),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glue Partition: %w", err)
	}
	return nil
}

func expandGluePartitionInput(d *schema.ResourceData) *glue.PartitionInput {
	tableInput := &glue.PartitionInput{}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		tableInput.StorageDescriptor = expandGlueStorageDescriptor(v.([]interface{}))
	}

	if v, ok := d.GetOk("parameters"); ok {
		tableInput.Parameters = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("partition_values"); ok && v.(*schema.Set).Len() > 0 {
		tableInput.Values = expandStringSet(v.(*schema.Set))
	}

	return tableInput
}
