package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	elasticacheEmptyDescription = " "
)

const (
	elasticacheGlobalReplicationGroupRegionPrefixFormat = "[[:alpha:]]{5}-"
)

const (
	GlobalReplicationGroupMemberRolePrimary   = "PRIMARY"
	GlobalReplicationGroupMemberRoleSecondary = "SECONDARY"
)

func resourceAwsElasticacheGlobalReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticacheGlobalReplicationGroupCreate,
		Read:   resourceAwsElasticacheGlobalReplicationGroupRead,
		Update: resourceAwsElasticacheGlobalReplicationGroupUpdate,
		Delete: resourceAwsElasticacheGlobalReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				re := regexp.MustCompile("^" + elasticacheGlobalReplicationGroupRegionPrefixFormat)
				d.Set("global_replication_group_id_suffix", re.ReplaceAllLiteralString(d.Id(), ""))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auth_token_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cache_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Leaving space for `engine_version` for creation and updating.
			// `engine_version` cannot be used for returning the version because, starting with Redis 6,
			// version configuration is major-version-only: `engine_version = "6.x"`, while `actual_engine_version`
			// will be e.g. `6.0.5`
			// See also https://github.com/hashicorp/terraform-provider-aws/issues/15625
			"actual_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_replication_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_replication_group_id_suffix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_replication_group_description": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: elasticacheDescriptionDiffSuppress,
				StateFunc:        elasticacheDescriptionStateFunc,
			},
			// global_replication_group_members cannot be correctly implemented because any secondary
			// replication groups will be added after this resource completes.
			// "global_replication_group_members": {
			// 	Type:     schema.TypeSet,
			// 	Computed: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"replication_group_id": {
			// 				Type:     schema.TypeString,
			// 				Computed: true,
			// 			},
			// 			"replication_group_region": {
			// 				Type:     schema.TypeString,
			// 				Computed: true,
			// 			},
			// 			"role": {
			// 				Type:     schema.TypeString,
			// 				Computed: true,
			// 			},
			// 		},
			// 	},
			// },
			"primary_replication_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateReplicationGroupID,
			},
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func elasticacheDescriptionDiffSuppress(_, old, new string, d *schema.ResourceData) bool {
	if (old == elasticacheEmptyDescription && new == "") || (old == "" && new == elasticacheEmptyDescription) {
		return true
	}
	return false
}

func elasticacheDescriptionStateFunc(v interface{}) string {
	s := v.(string)
	if s == "" {
		return elasticacheEmptyDescription
	}
	return s
}

func resourceAwsElasticacheGlobalReplicationGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	input := &elasticache.CreateGlobalReplicationGroupInput{
		GlobalReplicationGroupIdSuffix: aws.String(d.Get("global_replication_group_id_suffix").(string)),
		PrimaryReplicationGroupId:      aws.String(d.Get("primary_replication_group_id").(string)),
	}

	if v, ok := d.GetOk("global_replication_group_description"); ok {
		input.GlobalReplicationGroupDescription = aws.String(v.(string))
	}

	output, err := conn.CreateGlobalReplicationGroup(input)
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Global Replication Group: %w", err)
	}

	d.SetId(aws.StringValue(output.GlobalReplicationGroup.GlobalReplicationGroupId))

	if _, err := waiter.GlobalReplicationGroupAvailable(conn, d.Id(), waiter.GlobalReplicationGroupDefaultCreatedTimeout); err != nil {
		return fmt.Errorf("error waiting for ElastiCache Global Replication Group (%s) availability: %w", d.Id(), err)
	}

	return resourceAwsElasticacheGlobalReplicationGroupRead(d, meta)
}

func resourceAwsElasticacheGlobalReplicationGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	globalReplicationGroup, err := finder.GlobalReplicationGroupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading ElastiCache Replication Group: %w", err)
	}

	if aws.StringValue(globalReplicationGroup.Status) == "deleting" || aws.StringValue(globalReplicationGroup.Status) == "deleted" {
		log.Printf("[WARN] ElastiCache Global Replication Group (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(globalReplicationGroup.Status))
		d.SetId("")
		return nil
	}

	d.Set("arn", globalReplicationGroup.ARN)
	d.Set("at_rest_encryption_enabled", globalReplicationGroup.AtRestEncryptionEnabled)
	d.Set("auth_token_enabled", globalReplicationGroup.AuthTokenEnabled)
	d.Set("cache_node_type", globalReplicationGroup.CacheNodeType)
	d.Set("cluster_enabled", globalReplicationGroup.ClusterEnabled)
	d.Set("engine", globalReplicationGroup.Engine)
	d.Set("actual_engine_version", globalReplicationGroup.EngineVersion)
	d.Set("global_replication_group_description", globalReplicationGroup.GlobalReplicationGroupDescription)
	d.Set("global_replication_group_id", globalReplicationGroup.GlobalReplicationGroupId)
	d.Set("transit_encryption_enabled", globalReplicationGroup.TransitEncryptionEnabled)

	d.Set("primary_replication_group_id", flattenElasticacheGlobalReplicationGroupPrimaryGroupID(globalReplicationGroup.Members))

	return nil
}

func resourceAwsElasticacheGlobalReplicationGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	// Only one field can be changed per request
	updaters := map[string]elasticacheGlobalReplicationGroupUpdater{}
	if !d.IsNewResource() {
		updaters["global_replication_group_description"] = func(input *elasticache.ModifyGlobalReplicationGroupInput) {
			input.GlobalReplicationGroupDescription = aws.String(d.Get("global_replication_group_description").(string))
		}
	}

	for k, f := range updaters {
		if d.HasChange(k) {
			if err := updateElasticacheGlobalReplicationGroup(conn, d.Id(), f); err != nil {
				return fmt.Errorf("error updating ElastiCache Global Replication Group (%s): %w", d.Id(), err)
			}
		}
	}

	return resourceAwsElasticacheGlobalReplicationGroupRead(d, meta)
}

type elasticacheGlobalReplicationGroupUpdater func(input *elasticache.ModifyGlobalReplicationGroupInput)

func updateElasticacheGlobalReplicationGroup(conn *elasticache.ElastiCache, id string, f elasticacheGlobalReplicationGroupUpdater) error {
	input := &elasticache.ModifyGlobalReplicationGroupInput{
		ApplyImmediately:         aws.Bool(true),
		GlobalReplicationGroupId: aws.String(id),
	}
	f(input)

	if _, err := conn.ModifyGlobalReplicationGroup(input); err != nil {
		return err
	}

	if _, err := waiter.GlobalReplicationGroupAvailable(conn, id, waiter.GlobalReplicationGroupDefaultUpdatedTimeout); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}

func resourceAwsElasticacheGlobalReplicationGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	err := deleteElasticacheGlobalReplicationGroup(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Global Replication Group: %w", err)
	}

	return nil
}

func deleteElasticacheGlobalReplicationGroup(conn *elasticache.ElastiCache, id string) error {
	input := &elasticache.DeleteGlobalReplicationGroupInput{
		GlobalReplicationGroupId:      aws.String(id),
		RetainPrimaryReplicationGroup: aws.Bool(true),
	}

	// Using Update timeout because the Global Replication Group could be in the middle of an update operation
	err := resource.Retry(waiter.GlobalReplicationGroupDefaultUpdatedTimeout, func() *resource.RetryError {
		_, err := conn.DeleteGlobalReplicationGroup(input)
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
			return resource.NonRetryableError(&resource.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}
		if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidGlobalReplicationGroupStateFault, "is not empty") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteGlobalReplicationGroup(input)
	}
	if tfresource.NotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if _, err := waiter.GlobalReplicationGroupDeleted(conn, id); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}

func flattenElasticacheGlobalReplicationGroupPrimaryGroupID(members []*elasticache.GlobalReplicationGroupMember) string {
	for _, member := range members {
		if aws.StringValue(member.Role) == GlobalReplicationGroupMemberRolePrimary {
			return aws.StringValue(member.ReplicationGroupId)
		}
	}
	return ""
}
