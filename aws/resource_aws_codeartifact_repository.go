package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCodeArtifactRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeArtifactRepositoryCreate,
		Read:   resourceAwsCodeArtifactRepositoryRead,
		Update: resourceAwsCodeArtifactRepositoryUpdate,
		Delete: resourceAwsCodeArtifactRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"upstream": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"external_connections": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"package_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"administrator_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCodeArtifactRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Creating CodeArtifact Repository")

	params := &codeartifact.CreateRepositoryInput{
		Repository: aws.String(d.Get("repository").(string)),
		Domain:     aws.String(d.Get("domain").(string)),
		Tags:       keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CodeartifactTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("upstream"); ok {
		params.Upstreams = expandCodeArtifactUpstreams(v.([]interface{}))
	}

	res, err := conn.CreateRepository(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Repository: %w", err)
	}

	repo := res.Repository
	d.SetId(aws.StringValue(repo.Arn))

	if v, ok := d.GetOk("external_connections"); ok {
		externalConnection := v.([]interface{})[0].(map[string]interface{})
		input := &codeartifact.AssociateExternalConnectionInput{
			Domain:             repo.DomainName,
			Repository:         repo.Name,
			DomainOwner:        repo.DomainOwner,
			ExternalConnection: aws.String(externalConnection["external_connection_name"].(string)),
		}

		_, err := conn.AssociateExternalConnection(input)
		if err != nil {
			return fmt.Errorf("error associating external connection to CodeArtifact repository: %w", err)
		}
	}

	return resourceAwsCodeArtifactRepositoryRead(d, meta)
}

func resourceAwsCodeArtifactRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Updating CodeArtifact Repository")

	needsUpdate := false
	params := &codeartifact.UpdateRepositoryInput{
		Repository:  aws.String(d.Get("repository").(string)),
		Domain:      aws.String(d.Get("domain").(string)),
		DomainOwner: aws.String(d.Get("domain_owner").(string)),
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			params.Description = aws.String(v.(string))
			needsUpdate = true
		}
	}

	if d.HasChange("upstream") {
		if v, ok := d.GetOk("upstream"); ok {
			params.Upstreams = expandCodeArtifactUpstreams(v.([]interface{}))
			needsUpdate = true
		}
	}

	if needsUpdate {
		_, err := conn.UpdateRepository(params)
		if err != nil {
			return fmt.Errorf("error updating CodeArtifact Repository: %w", err)
		}
	}

	if d.HasChange("external_connections") {
		if v, ok := d.GetOk("external_connections"); ok {
			externalConnection := v.([]interface{})[0].(map[string]interface{})
			input := &codeartifact.AssociateExternalConnectionInput{
				Repository:         aws.String(d.Get("repository").(string)),
				Domain:             aws.String(d.Get("domain").(string)),
				DomainOwner:        aws.String(d.Get("domain_owner").(string)),
				ExternalConnection: aws.String(externalConnection["external_connection_name"].(string)),
			}

			_, err := conn.AssociateExternalConnection(input)
			if err != nil {
				return fmt.Errorf("error associating external connection to CodeArtifact repository: %w", err)
			}
		} else {
			oldConn, _ := d.GetChange("external_connections")
			externalConnection := oldConn.([]interface{})[0].(map[string]interface{})
			input := &codeartifact.DisassociateExternalConnectionInput{
				Repository:         aws.String(d.Get("repository").(string)),
				Domain:             aws.String(d.Get("domain").(string)),
				DomainOwner:        aws.String(d.Get("domain_owner").(string)),
				ExternalConnection: aws.String(externalConnection["external_connection_name"].(string)),
			}

			_, err := conn.DisassociateExternalConnection(input)
			if err != nil {
				return fmt.Errorf("error disassociating external connection to CodeArtifact repository: %w", err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.CodeartifactUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating CodeArtifact Repository (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsCodeArtifactRepositoryRead(d, meta)
}

func resourceAwsCodeArtifactRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading CodeArtifact Repository: %s", d.Id())

	owner, domain, repo, err := decodeCodeArtifactRepositoryID(d.Id())
	if err != nil {
		return err
	}
	sm, err := conn.DescribeRepository(&codeartifact.DescribeRepositoryInput{
		Repository:  aws.String(repo),
		Domain:      aws.String(domain),
		DomainOwner: aws.String(owner),
	})
	if err != nil {
		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CodeArtifact Repository %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading CodeArtifact Repository (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(sm.Repository.Arn)
	d.Set("repository", sm.Repository.Name)
	d.Set("arn", arn)
	d.Set("domain_owner", sm.Repository.DomainOwner)
	d.Set("domain", sm.Repository.DomainName)
	d.Set("administrator_account", sm.Repository.AdministratorAccount)
	d.Set("description", sm.Repository.Description)

	if sm.Repository.Upstreams != nil {
		if err := d.Set("upstream", flattenCodeArtifactUpstreams(sm.Repository.Upstreams)); err != nil {
			return fmt.Errorf("[WARN] Error setting upstream: %w", err)
		}
	}

	if sm.Repository.ExternalConnections != nil {
		if err := d.Set("external_connections", flattenCodeArtifactExternalConnections(sm.Repository.ExternalConnections)); err != nil {
			return fmt.Errorf("[WARN] Error setting external_connections: %w", err)
		}
	}

	tags, err := keyvaluetags.CodeartifactListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CodeArtifact Repository (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsCodeArtifactRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Printf("[DEBUG] Deleting CodeArtifact Repository: %s", d.Id())

	owner, domain, repo, err := decodeCodeArtifactRepositoryID(d.Id())
	if err != nil {
		return err
	}
	input := &codeartifact.DeleteRepositoryInput{
		Repository:  aws.String(repo),
		Domain:      aws.String(domain),
		DomainOwner: aws.String(owner),
	}

	_, err = conn.DeleteRepository(input)

	if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Repository (%s): %w", d.Id(), err)
	}

	return nil
}

func expandCodeArtifactUpstreams(l []interface{}) []*codeartifact.UpstreamRepository {
	upstreams := []*codeartifact.UpstreamRepository{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})
		upstream := &codeartifact.UpstreamRepository{
			RepositoryName: aws.String(m["repository_name"].(string)),
		}

		upstreams = append(upstreams, upstream)
	}

	return upstreams
}

func flattenCodeArtifactUpstreams(upstreams []*codeartifact.UpstreamRepositoryInfo) []interface{} {
	if len(upstreams) == 0 {
		return nil
	}

	var ls []interface{}

	for _, upstream := range upstreams {
		m := map[string]interface{}{
			"repository_name": aws.StringValue(upstream.RepositoryName),
		}

		ls = append(ls, m)
	}

	return ls
}

func flattenCodeArtifactExternalConnections(connections []*codeartifact.RepositoryExternalConnectionInfo) []interface{} {
	if len(connections) == 0 {
		return nil
	}

	var ls []interface{}

	for _, connection := range connections {
		m := map[string]interface{}{
			"external_connection_name": aws.StringValue(connection.ExternalConnectionName),
			"package_format":           aws.StringValue(connection.PackageFormat),
			"status":                   aws.StringValue(connection.Status),
		}

		ls = append(ls, m)
	}

	return ls
}

func decodeCodeArtifactRepositoryID(id string) (string, string, string, error) {
	repoArn, err := arn.Parse(id)
	if err != nil {
		return "", "", "", err
	}

	idParts := strings.Split(strings.TrimPrefix(repoArn.Resource, "repository/"), "/")
	if len(idParts) != 2 {
		return "", "", "", fmt.Errorf("expected resource part of arn in format DomainName/RepositoryName, received: %s", repoArn.Resource)
	}
	return repoArn.AccountID, idParts[0], idParts[1], nil
}
