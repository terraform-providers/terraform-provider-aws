package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEksFargateProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEksFargateProfileCreate,
		Read:   resourceAwsEksFargateProfileRead,
		Update: resourceAwsEksFargateProfileUpdate,
		Delete: resourceAwsEksFargateProfileDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"fargate_profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"pod_execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"selector": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEksFargateProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	clusterName := d.Get("cluster_name").(string)
	fargateProfileName := d.Get("fargate_profile_name").(string)
	id := fmt.Sprintf("%s:%s", clusterName, fargateProfileName)

	input := &eks.CreateFargateProfileInput{
		ClientRequestToken:  aws.String(resource.UniqueId()),
		ClusterName:         aws.String(clusterName),
		FargateProfileName:  aws.String(fargateProfileName),
		PodExecutionRoleArn: aws.String(d.Get("pod_execution_role_arn").(string)),
		Selectors:           expandEksFargateProfileSelectors(d.Get("selector").(*schema.Set).List()),
		Subnets:             expandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
	}

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateFargateProfile(input)

		// Retry for IAM eventual consistency on error:
		// InvalidParameterException: Misconfigured PodExecutionRole Trust Policy; Please add the eks-fargate-pods.amazonaws.com Service Principal
		if isAWSErr(err, eks.ErrCodeInvalidParameterException, "Misconfigured PodExecutionRole Trust Policy") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateFargateProfile(input)
	}

	if err != nil {
		return fmt.Errorf("error creating EKS Fargate Profile (%s): %s", id, err)
	}

	d.SetId(id)

	stateConf := resource.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusCreating},
		Target:  []string{eks.FargateProfileStatusActive},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: refreshEksFargateProfileStatus(conn, clusterName, fargateProfileName),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for EKS Fargate Profile (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsEksFargateProfileRead(d, meta)
}

func resourceAwsEksFargateProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterName, fargateProfileName, err := resourceAwsEksFargateProfileParseId(d.Id())
	if err != nil {
		return err
	}

	input := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	output, err := conn.DescribeFargateProfile(input)

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] EKS Fargate Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	fargateProfile := output.FargateProfile

	if fargateProfile == nil {
		log.Printf("[WARN] EKS Fargate Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", fargateProfile.FargateProfileArn)
	d.Set("cluster_name", fargateProfile.ClusterName)
	d.Set("fargate_profile_name", fargateProfile.FargateProfileName)
	d.Set("pod_execution_role_arn", fargateProfile.PodExecutionRoleArn)

	if err := d.Set("selector", flattenEksFargateProfileSelectors(fargateProfile.Selectors)); err != nil {
		return fmt.Errorf("error setting selector: %s", err)
	}

	d.Set("status", fargateProfile.Status)

	if err := d.Set("subnet_ids", aws.StringValueSlice(fargateProfile.Subnets)); err != nil {
		return fmt.Errorf("error setting subnets: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(fargateProfile.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEksFargateProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsEksFargateProfileRead(d, meta)
}

func resourceAwsEksFargateProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	clusterName, fargateProfileName, err := resourceAwsEksFargateProfileParseId(d.Id())
	if err != nil {
		return err
	}

	input := &eks.DeleteFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	_, err = conn.DeleteFargateProfile(input)

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EKS Fargate Profile (%s): %s", d.Id(), err)
	}

	if err := waitForEksFargateProfileDeletion(conn, clusterName, fargateProfileName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EKS Fargate Profile (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandEksFargateProfileSelectors(l []interface{}) []*eks.FargateProfileSelector {
	if len(l) == 0 {
		return nil
	}

	fargateProfileSelectors := make([]*eks.FargateProfileSelector, 0, len(l))

	for _, mRaw := range l {
		m, ok := mRaw.(map[string]interface{})

		if !ok {
			continue
		}

		fargateProfileSelector := &eks.FargateProfileSelector{}

		if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
			fargateProfileSelector.Labels = stringMapToPointers(v)
		}

		if v, ok := m["namespace"].(string); ok && v != "" {
			fargateProfileSelector.Namespace = aws.String(v)
		}

		fargateProfileSelectors = append(fargateProfileSelectors, fargateProfileSelector)
	}

	return fargateProfileSelectors
}

func flattenEksFargateProfileSelectors(fargateProfileSelectors []*eks.FargateProfileSelector) []map[string]interface{} {
	if len(fargateProfileSelectors) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(fargateProfileSelectors))

	for _, fargateProfileSelector := range fargateProfileSelectors {
		m := map[string]interface{}{
			"labels":    aws.StringValueMap(fargateProfileSelector.Labels),
			"namespace": aws.StringValue(fargateProfileSelector.Namespace),
		}

		l = append(l, m)
	}

	return l
}

func refreshEksFargateProfileStatus(conn *eks.EKS, clusterName string, fargateProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &eks.DescribeFargateProfileInput{
			ClusterName:        aws.String(clusterName),
			FargateProfileName: aws.String(fargateProfileName),
		}

		output, err := conn.DescribeFargateProfile(input)

		if err != nil {
			return "", "", err
		}

		fargateProfile := output.FargateProfile

		if fargateProfile == nil {
			return fargateProfile, "", fmt.Errorf("EKS Fargate Profile (%s:%s) missing", clusterName, fargateProfileName)
		}

		return fargateProfile, aws.StringValue(fargateProfile.Status), nil
	}
}

func waitForEksFargateProfileDeletion(conn *eks.EKS, clusterName string, fargateProfileName string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			eks.FargateProfileStatusActive,
			eks.FargateProfileStatusDeleting,
		},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: refreshEksFargateProfileStatus(conn, clusterName, fargateProfileName),
	}

	_, err := stateConf.WaitForState()

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	return err
}

func resourceAwsEksFargateProfileParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected cluster-name:fargate-profile-name", id)
	}

	return parts[0], parts[1], nil
}
