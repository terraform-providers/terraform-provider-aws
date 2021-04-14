package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsSignerSigningJob() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSignerSigningJobRead,

		Schema: map[string]*schema.Schema{
			"job_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"completed_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_invoker": {
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
			"profile_name": {
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
			"source": {
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
									"version": {
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

func dataSourceAwsSignerSigningJobRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn
	jobId := d.Get("job_id").(string)

	describeSigningJobOutput, err := conn.DescribeSigningJob(&signer.DescribeSigningJobInput{
		JobId: aws.String(jobId),
	})

	if err != nil {
		return fmt.Errorf("error reading Signer signing job (%s): %w", d.Id(), err)
	}

	if err := d.Set("completed_at", aws.TimeValue(describeSigningJobOutput.CompletedAt).Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting signer signing job completed at: %w", err)
	}

	if err := d.Set("created_at", aws.TimeValue(describeSigningJobOutput.CreatedAt).Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting signer signing job created at: %w", err)
	}

	if err := d.Set("job_invoker", describeSigningJobOutput.JobInvoker); err != nil {
		return fmt.Errorf("error setting signer signing job invoker: %w", err)
	}

	if err := d.Set("job_owner", describeSigningJobOutput.JobOwner); err != nil {
		return fmt.Errorf("error setting signer signing job owner: %w", err)
	}

	if err := d.Set("platform_display_name", describeSigningJobOutput.PlatformDisplayName); err != nil {
		return fmt.Errorf("error setting signer signing job platform display name: %w", err)
	}

	if err := d.Set("platform_id", describeSigningJobOutput.PlatformId); err != nil {
		return fmt.Errorf("error setting signer signing job platform id: %w", err)
	}

	if err := d.Set("profile_name", describeSigningJobOutput.ProfileName); err != nil {
		return fmt.Errorf("error setting signer signing job profile name: %w", err)
	}

	if err := d.Set("profile_version", describeSigningJobOutput.ProfileVersion); err != nil {
		return fmt.Errorf("error setting signer signing job profile version: %w", err)
	}

	if err := d.Set("requested_by", describeSigningJobOutput.RequestedBy); err != nil {
		return fmt.Errorf("error setting signer signing job requested by: %w", err)
	}

	if err := d.Set("revocation_record", flattenSignerSigningJobRevocationRecord(describeSigningJobOutput.RevocationRecord)); err != nil {
		return fmt.Errorf("error setting signer signing job revocation record: %w", err)
	}

	signatureExpiresAt := ""
	if describeSigningJobOutput.SignatureExpiresAt != nil {
		signatureExpiresAt = aws.TimeValue(describeSigningJobOutput.SignatureExpiresAt).Format(time.RFC3339)
	}
	if err := d.Set("signature_expires_at", signatureExpiresAt); err != nil {
		return fmt.Errorf("error setting signer signing job requested by: %w", err)
	}

	if err := d.Set("signed_object", flattenSignerSigningJobSignedObject(describeSigningJobOutput.SignedObject)); err != nil {
		return fmt.Errorf("error setting signer signing job signed object: %w", err)
	}

	if err := d.Set("source", flattenSignerSigningJobSource(describeSigningJobOutput.Source)); err != nil {
		return fmt.Errorf("error setting signer signing job source: %w", err)
	}

	if err := d.Set("status", describeSigningJobOutput.Status); err != nil {
		return fmt.Errorf("error setting signer signing job status: %w", err)
	}

	if err := d.Set("status_reason", describeSigningJobOutput.StatusReason); err != nil {
		return fmt.Errorf("error setting signer signing job status reason: %w", err)
	}

	d.SetId(aws.StringValue(describeSigningJobOutput.JobId))

	return nil
}
