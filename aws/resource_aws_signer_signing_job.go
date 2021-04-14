package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSignerSigningJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSignerSigningJobCreate,
		Read:   resourceAwsSignerSigningJobRead,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"version": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"ignore_signing_job_failure": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"completed_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_invoker": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"requested_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"signature_expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signed_object": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSignerSigningJobCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn
	profileName := d.Get("profile_name")
	source := d.Get("source").([]interface{})
	destination := d.Get("destination").([]interface{})

	startSigningJobInput := &signer.StartSigningJobInput{
		ProfileName: aws.String(profileName.(string)),
		Source:      expandSignerSigningJobSource(source),
		Destination: expandSignerSigningJobDestination(destination),
	}

	log.Printf("[DEBUG] Starting Signer Signing Job using profile name %q.", profileName)
	startSigningJobOutput, err := conn.StartSigningJob(startSigningJobInput)
	if err != nil {
		return fmt.Errorf("error starting Signing Job: %w", err)
	}

	jobId := aws.StringValue(startSigningJobOutput.JobId)

	ignoreSigningJobFailure := d.Get("ignore_signing_job_failure").(bool)
	log.Printf("[DEBUG] Waiting for Signer Signing Job ID (%s) to complete.", jobId)
	waiterError := conn.WaitUntilSuccessfulSigningJobWithContext(aws.BackgroundContext(), &signer.DescribeSigningJobInput{
		JobId: aws.String(jobId),
	}, request.WithWaiterMaxAttempts(200), request.WithWaiterDelay(request.ConstantWaiterDelay(5*time.Second)))
	if waiterError != nil {
		if !ignoreSigningJobFailure || !tfawserr.ErrCodeEquals(waiterError, request.WaiterResourceNotReadyErrorCode) {
			return waiterError
		}
	}

	d.SetId(jobId)

	return resourceAwsSignerSigningJobRead(d, meta)
}

func resourceAwsSignerSigningJobRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn
	jobId := d.Id()

	describeSigningJobOutput, err := conn.DescribeSigningJob(&signer.DescribeSigningJobInput{
		JobId: aws.String(jobId),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, signer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Signer Signing Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Signer signing job (%s): %s", d.Id(), err)
	}

	if err := d.Set("job_id", describeSigningJobOutput.JobId); err != nil {
		return fmt.Errorf("error setting signer signing job id: %s", err)
	}

	if err := d.Set("completed_at", aws.TimeValue(describeSigningJobOutput.CompletedAt).Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting signer signing job completed at: %s", err)
	}

	if err := d.Set("created_at", aws.TimeValue(describeSigningJobOutput.CreatedAt).Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting signer signing job created at: %s", err)
	}

	if err := d.Set("job_invoker", describeSigningJobOutput.JobInvoker); err != nil {
		return fmt.Errorf("error setting signer signing job invoker: %s", err)
	}

	if err := d.Set("job_owner", describeSigningJobOutput.JobOwner); err != nil {
		return fmt.Errorf("error setting signer signing job owner: %s", err)
	}

	if err := d.Set("platform_display_name", describeSigningJobOutput.PlatformDisplayName); err != nil {
		return fmt.Errorf("error setting signer signing job platform display name: %s", err)
	}

	if err := d.Set("platform_id", describeSigningJobOutput.PlatformId); err != nil {
		return fmt.Errorf("error setting signer signing job platform id: %s", err)
	}

	if err := d.Set("profile_name", describeSigningJobOutput.ProfileName); err != nil {
		return fmt.Errorf("error setting signer signing job profile name: %s", err)
	}

	if err := d.Set("profile_version", describeSigningJobOutput.ProfileVersion); err != nil {
		return fmt.Errorf("error setting signer signing job profile version: %s", err)
	}

	if err := d.Set("requested_by", describeSigningJobOutput.RequestedBy); err != nil {
		return fmt.Errorf("error setting signer signing job requested by: %s", err)
	}

	if err := d.Set("revocation_record", flattenSignerSigningJobRevocationRecord(describeSigningJobOutput.RevocationRecord)); err != nil {
		return fmt.Errorf("error setting signer signing job revocation record: %s", err)
	}

	signatureExpiresAt := ""
	if describeSigningJobOutput.SignatureExpiresAt != nil {
		signatureExpiresAt = aws.TimeValue(describeSigningJobOutput.SignatureExpiresAt).Format(time.RFC3339)
	}
	if err := d.Set("signature_expires_at", signatureExpiresAt); err != nil {
		return fmt.Errorf("error setting signer signing job requested by: %s", err)
	}

	if err := d.Set("signed_object", flattenSignerSigningJobSignedObject(describeSigningJobOutput.SignedObject)); err != nil {
		return fmt.Errorf("error setting signer signing job signed object: %s", err)
	}

	if err := d.Set("source", flattenSignerSigningJobSource(describeSigningJobOutput.Source)); err != nil {
		return fmt.Errorf("error setting signer signing job source: %s", err)
	}

	if err := d.Set("status", describeSigningJobOutput.Status); err != nil {
		return fmt.Errorf("error setting signer signing job status: %s", err)
	}

	if err := d.Set("status_reason", describeSigningJobOutput.StatusReason); err != nil {
		return fmt.Errorf("error setting signer signing job status reason: %s", err)
	}

	return nil
}

func flattenSignerSigningJobRevocationRecord(apiObject *signer.SigningJobRevocationRecord) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Reason; v != nil {
		tfMap["reason"] = aws.StringValue(v)
	}

	if v := apiObject.RevokedAt; v != nil {
		tfMap["revoked_at"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedBy; v != nil {
		tfMap["revoked_by"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenSignerSigningJobSource(apiObject *signer.Source) []interface{} {
	if apiObject == nil || apiObject.S3 == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"s3": flattenSignerSigningJobS3Source(apiObject.S3),
	}

	return []interface{}{tfMap}
}

func flattenSignerSigningJobS3Source(apiObject *signer.S3Source) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		tfMap["bucket"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandSignerSigningJobSource(tfList []interface{}) *signer.Source {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var source *signer.Source
	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		source = &signer.Source{
			S3: expandSignerSigningJobS3Source(v),
		}
	}

	return source
}

func expandSignerSigningJobS3Source(tfList []interface{}) *signer.S3Source {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	s3Source := &signer.S3Source{}

	if v, ok := tfMap["bucket"].(string); ok {
		s3Source.BucketName = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok {
		s3Source.Key = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok {
		s3Source.Version = aws.String(v)
	}

	return s3Source
}

func expandSignerSigningJobDestination(tfList []interface{}) *signer.Destination {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var destination *signer.Destination
	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		destination = &signer.Destination{
			S3: expandSignerSigningJobS3Destination(v),
		}
	}

	return destination
}

func expandSignerSigningJobS3Destination(tfList []interface{}) *signer.S3Destination {
	if tfList == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	s3Destination := &signer.S3Destination{}

	if _, ok := tfMap["bucket"]; ok {
		s3Destination.BucketName = aws.String(tfMap["bucket"].(string))
	}

	if _, ok := tfMap["prefix"]; ok {
		s3Destination.Prefix = aws.String(tfMap["prefix"].(string))
	}

	return s3Destination
}

func flattenSignerSigningJobSignedObject(apiObject *signer.SignedObject) []interface{} {
	if apiObject == nil || apiObject.S3 == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"s3": flattenSignerSigningJobS3SignedObject(apiObject.S3),
	}

	return []interface{}{tfMap}
}

func flattenSignerSigningJobS3SignedObject(apiObject *signer.S3SignedObject) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		tfMap["bucket"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
