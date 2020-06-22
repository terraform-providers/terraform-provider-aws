package aws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	route53ResolverEndpointStatusDeleted = "DELETED"
)

func resourceAwsRoute53ResolverEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverEndpointCreate,
		Read:   resourceAwsRoute53ResolverEndpointRead,
		Update: resourceAwsRoute53ResolverEndpointUpdate,
		Delete: resourceAwsRoute53ResolverEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.ResolverEndpointDirectionInbound,
					route53resolver.ResolverEndpointDirectionOutbound,
				}, false),
			},

			"ip_address": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 2,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"ip_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: route53ResolverEndpointHashIpAddress,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 64,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRoute53ResolverName,
			},

			"tags": tagsSchema(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"host_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsRoute53ResolverEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	req := &route53resolver.CreateResolverEndpointInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-endpoint-")),
		Direction:        aws.String(d.Get("direction").(string)),
		IpAddresses:      expandRoute53ResolverEndpointIpAddresses(d.Get("ip_address").(*schema.Set)),
		SecurityGroupIds: expandStringSet(d.Get("security_group_ids").(*schema.Set)),
	}
	if v, ok := d.GetOk("name"); ok {
		req.Name = aws.String(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		req.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Route53resolverTags()
	}

	log.Printf("[DEBUG] Creating Route53 Resolver endpoint: %#v", req)
	resp, err := conn.CreateResolverEndpoint(req)
	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver endpoint: %s", err)
	}

	d.SetId(aws.StringValue(resp.ResolverEndpoint.Id))

	err = route53ResolverEndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutCreate),
		[]string{route53resolver.ResolverEndpointStatusCreating},
		[]string{route53resolver.ResolverEndpointStatusOperational})
	if err != nil {
		return err
	}

	return resourceAwsRoute53ResolverEndpointRead(d, meta)
}

func resourceAwsRoute53ResolverEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	epRaw, state, err := route53ResolverEndpointRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver endpoint (%s): %s", d.Id(), err)
	}
	if state == route53ResolverEndpointStatusDeleted {
		log.Printf("[WARN] Route53 Resolver endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	ep := epRaw.(*route53resolver.ResolverEndpoint)
	d.Set("arn", ep.Arn)
	d.Set("direction", ep.Direction)
	d.Set("host_vpc_id", ep.HostVPCId)
	d.Set("name", ep.Name)
	if err := d.Set("security_group_ids", flattenStringSet(ep.SecurityGroupIds)); err != nil {
		return err
	}

	ipAddresses := []interface{}{}
	req := &route53resolver.ListResolverEndpointIpAddressesInput{
		ResolverEndpointId: aws.String(d.Id()),
	}
	for {
		resp, err := conn.ListResolverEndpointIpAddresses(req)
		if err != nil {
			return fmt.Errorf("error getting Route53 Resolver endpoint (%s) IP addresses: %s", d.Id(), err)
		}

		ipAddresses = append(ipAddresses, flattenRoute53ResolverEndpointIpAddresses(resp.IpAddresses)...)

		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	if err := d.Set("ip_address", schema.NewSet(route53ResolverEndpointHashIpAddress, ipAddresses)); err != nil {
		return err
	}

	tags, err := keyvaluetags.Route53resolverListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver endpoint (%s): %s", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsRoute53ResolverEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	if d.HasChange("name") {
		req := &route53resolver.UpdateResolverEndpointInput{
			ResolverEndpointId: aws.String(d.Id()),
			Name:               aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating Route53 Resolver endpoint: %#v", req)
		_, err := conn.UpdateResolverEndpoint(req)
		if err != nil {
			return fmt.Errorf("error updating Route53 Resolver endpoint (%s): %s", d.Id(), err)
		}

		err = route53ResolverEndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
			[]string{route53resolver.ResolverEndpointStatusUpdating},
			[]string{route53resolver.ResolverEndpointStatusOperational})
		if err != nil {
			return err
		}
	}

	if d.HasChange("ip_address") {
		oraw, nraw := d.GetChange("ip_address")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)
		del := o.Difference(n).List()
		add := n.Difference(o).List()

		// Add new before deleting old so number of IP addresses doesn't drop below 2.
		for _, v := range add {
			_, err := conn.AssociateResolverEndpointIpAddress(&route53resolver.AssociateResolverEndpointIpAddressInput{
				ResolverEndpointId: aws.String(d.Id()),
				IpAddress:          expandRoute53ResolverEndpointIpAddressUpdate(v),
			})
			if err != nil {
				return fmt.Errorf("error associating Route53 Resolver endpoint (%s) IP address: %s", d.Id(), err)
			}

			err = route53ResolverEndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
				[]string{route53resolver.ResolverEndpointStatusUpdating},
				[]string{route53resolver.ResolverEndpointStatusOperational})
			if err != nil {
				return err
			}
		}

		for _, v := range del {
			_, err := conn.DisassociateResolverEndpointIpAddress(&route53resolver.DisassociateResolverEndpointIpAddressInput{
				ResolverEndpointId: aws.String(d.Id()),
				IpAddress:          expandRoute53ResolverEndpointIpAddressUpdate(v),
			})
			if err != nil {
				return fmt.Errorf("error disassociating Route53 Resolver endpoint (%s) IP address: %s", d.Id(), err)
			}

			err = route53ResolverEndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
				[]string{route53resolver.ResolverEndpointStatusUpdating},
				[]string{route53resolver.ResolverEndpointStatusOperational})
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Route53resolverUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver endpoint (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsRoute53ResolverEndpointRead(d, meta)
}

func resourceAwsRoute53ResolverEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route53 Resolver endpoint: %s", d.Id())
	_, err := conn.DeleteResolverEndpoint(&route53resolver.DeleteResolverEndpointInput{
		ResolverEndpointId: aws.String(d.Id()),
	})
	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver endpoint (%s): %s", d.Id(), err)
	}

	err = route53ResolverEndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{route53resolver.ResolverEndpointStatusDeleting},
		[]string{route53ResolverEndpointStatusDeleted})
	if err != nil {
		return err
	}

	return nil
}

func route53ResolverEndpointRefresh(conn *route53resolver.Route53Resolver, epId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetResolverEndpoint(&route53resolver.GetResolverEndpointInput{
			ResolverEndpointId: aws.String(epId),
		})
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			return &route53resolver.ResolverEndpoint{}, route53ResolverEndpointStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		if statusMessage := aws.StringValue(resp.ResolverEndpoint.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Route 53 Resolver endpoint (%s) status message: %s", epId, statusMessage)
		}

		return resp.ResolverEndpoint, aws.StringValue(resp.ResolverEndpoint.Status), nil
	}
}

func route53ResolverEndpointWaitUntilTargetState(conn *route53resolver.Route53Resolver, epId string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    route53ResolverEndpointRefresh(conn, epId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver endpoint (%s) to reach target state: %s", epId, err)
	}

	return nil
}

func route53ResolverEndpointHashIpAddress(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	return hashcode.String(buf.String())
}
