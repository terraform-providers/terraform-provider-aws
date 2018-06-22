package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsS3BucketReplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketReplicationPut,
		Read:   resourceAwsS3BucketReplicationRead,
		Update: resourceAwsS3BucketReplicationPut,
		Delete: resourceAwsS3BucketReplicationDelete,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"replication_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rules": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      rulesHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateS3BucketReplicationRuleId,
									},
									"destination": {
										Type:     schema.TypeSet,
										MaxItems: 1,
										MinItems: 1,
										Required: true,
										Set:      destinationHash,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateArn,
												},
												"storage_class": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateS3BucketReplicationDestinationStorageClass,
												},
												"replica_kms_key_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"source_selection_criteria": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										MaxItems: 1,
										Set:      sourceSelectionCriteriaHash,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"sse_kms_encrypted_objects": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 1,
													MaxItems: 1,
													Set:      sourceSseKmsObjectsHash,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"enabled": {
																Type:     schema.TypeBool,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"prefix": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateS3BucketReplicationRulePrefix,
									},
									"status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateS3BucketReplicationRuleStatus,
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

func resourceAwsS3BucketReplicationPut(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	replicationConfiguration := d.Get("replication_configuration").([]interface{})
	c := replicationConfiguration[0].(map[string]interface{})

	rc := buildAwsS3BucketReplicationConfiguration(c)

	params := &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: rc,
	}

	log.Printf("[DEBUG] S3 put bucket replication configuration: %#v", params)

	_, err := s3conn.PutBucketReplication(params)

	if err != nil {
		return fmt.Errorf("Error putting S3 replication configuration: %s", err)
	}

	d.SetId(bucket)
	return resourceAwsS3BucketReplicationRead(d, meta)
}

func resourceAwsS3BucketReplicationRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	log.Printf("[DEBUG] S3 bucket replication, reading for bucket: %s", d.Id())

	replication, err := s3conn.GetBucketReplication(&s3.GetBucketReplicationInput{
		Bucket: aws.String(d.Id()),
	})

	if err != nil {
		if awsError, ok := err.(awserr.RequestFailure); ok && awsError.StatusCode() != 404 {
			return err
		}
	}

	log.Printf("[DEBUG] S3 bucket: %s, read replication configuration: %v", d.Id(), replication)

	if r := replication.ReplicationConfiguration; r != nil {
		if err := d.Set("replication_configuration", flattenAwsS3BucketReplicationConfiguration(replication.ReplicationConfiguration)); err != nil {
			log.Printf("[DEBUG] Error setting replication configuration: %s", err)
			return err
		}
	}

	return nil
}

func resourceAwsS3BucketReplicationDelete(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	log.Printf("[DEBUG] S3 bucket: %s, delete replication configuration", bucket)

	_, err := s3conn.DeleteBucketReplication(&s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
			return nil
		}
		return fmt.Errorf("Error deleting S3 replication configuration: %s", err)
	}

	d.SetId("")
	return nil
}
