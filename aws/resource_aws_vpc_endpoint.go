package aws

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsVpcEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointCreate,
		Read:   resourceAwsVpcEndpointRead,
		Update: resourceAwsVpcEndpointUpdate,
		Delete: resourceAwsVpcEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"auto_accept": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"requester_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tagsSchema(),
			"vpc_endpoint_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.VpcEndpointTypeGateway,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.VpcEndpointTypeGateway,
					ec2.VpcEndpointTypeInterface,
				}, false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsVpcEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	if d.Get("vpc_endpoint_type").(string) == ec2.VpcEndpointTypeInterface &&
		d.Get("security_group_ids").(*schema.Set).Len() == 0 {
		return errors.New("An Interface VPC Endpoint must always have at least one Security Group")
	}

	conn := meta.(*AWSClient).ec2conn

	req := &ec2.CreateVpcEndpointInput{
		VpcId:             aws.String(d.Get("vpc_id").(string)),
		VpcEndpointType:   aws.String(d.Get("vpc_endpoint_type").(string)),
		ServiceName:       aws.String(d.Get("service_name").(string)),
		PrivateDnsEnabled: aws.Bool(d.Get("private_dns_enabled").(bool)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), "vpc-endpoint"),
	}

	if v, ok := d.GetOk("policy"); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		req.PolicyDocument = aws.String(policy)
	}

	setVpcEndpointCreateList(d, "route_table_ids", &req.RouteTableIds)
	setVpcEndpointCreateList(d, "subnet_ids", &req.SubnetIds)
	setVpcEndpointCreateList(d, "security_group_ids", &req.SecurityGroupIds)

	log.Printf("[DEBUG] Creating VPC Endpoint: %#v", req)
	resp, err := conn.CreateVpcEndpoint(req)
	if err != nil {
		return fmt.Errorf("Error creating VPC Endpoint: %s", err)
	}

	vpce := resp.VpcEndpoint
	d.SetId(aws.StringValue(vpce.VpcEndpointId))

	if _, ok := d.GetOk("auto_accept"); ok && aws.StringValue(vpce.State) == "pendingAcceptance" {
		if err := vpcEndpointAccept(conn, d.Id(), aws.StringValue(vpce.ServiceName)); err != nil {
			return err
		}
	}

	if err := vpcEndpointWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAwsVpcEndpointRead(d, meta)
}

func resourceAwsVpcEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	vpceRaw, state, err := vpcEndpointStateRefresh(conn, d.Id())()
	if err != nil && state != "failed" {
		return fmt.Errorf("error reading VPC Endpoint (%s): %s", d.Id(), err)
	}

	terminalStates := map[string]bool{
		"deleted":  true,
		"deleting": true,
		"failed":   true,
		"expired":  true,
		"rejected": true,
	}
	if _, ok := terminalStates[state]; ok {
		log.Printf("[WARN] VPC Endpoint (%s) in state (%s), removing from state", d.Id(), state)
		d.SetId("")
		return nil
	}

	vpce := vpceRaw.(*ec2.VpcEndpoint)

	serviceName := aws.StringValue(vpce.ServiceName)
	d.Set("service_name", serviceName)
	d.Set("state", vpce.State)
	d.Set("vpc_id", vpce.VpcId)

	respPl, err := conn.DescribePrefixLists(&ec2.DescribePrefixListsInput{
		Filters: buildEC2AttributeFilterList(map[string]string{
			"prefix-list-name": serviceName,
		}),
	})
	if err != nil {
		return fmt.Errorf("error reading Prefix List (%s): %s", serviceName, err)
	}
	if respPl == nil || len(respPl.PrefixLists) == 0 {
		d.Set("cidr_blocks", []interface{}{})
	} else if len(respPl.PrefixLists) > 1 {
		return fmt.Errorf("multiple prefix lists associated with the service name '%s'. Unexpected", serviceName)
	} else {
		pl := respPl.PrefixLists[0]

		d.Set("prefix_list_id", pl.PrefixListId)
		err = d.Set("cidr_blocks", flattenStringList(pl.Cidrs))
		if err != nil {
			return fmt.Errorf("error setting cidr_blocks: %s", err)
		}
	}

	err = d.Set("dns_entry", flattenVpcEndpointDnsEntries(vpce.DnsEntries))
	if err != nil {
		return fmt.Errorf("error setting dns_entry: %s", err)
	}
	err = d.Set("network_interface_ids", flattenStringSet(vpce.NetworkInterfaceIds))
	if err != nil {
		return fmt.Errorf("error setting network_interface_ids: %s", err)
	}
	d.Set("owner_id", vpce.OwnerId)
	policy, err := structure.NormalizeJsonString(aws.StringValue(vpce.PolicyDocument))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %s", err)
	}
	d.Set("policy", policy)
	d.Set("private_dns_enabled", vpce.PrivateDnsEnabled)
	err = d.Set("route_table_ids", flattenStringSet(vpce.RouteTableIds))
	if err != nil {
		return fmt.Errorf("error setting route_table_ids: %s", err)
	}
	d.Set("requester_managed", vpce.RequesterManaged)
	err = d.Set("security_group_ids", flattenVpcEndpointSecurityGroupIds(vpce.Groups))
	if err != nil {
		return fmt.Errorf("error setting security_group_ids: %s", err)
	}
	err = d.Set("subnet_ids", flattenStringSet(vpce.SubnetIds))
	if err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	// VPC endpoints don't have types in GovCloud, so set type to default if empty
	if vpceType := aws.StringValue(vpce.VpcEndpointType); vpceType == "" {
		d.Set("vpc_endpoint_type", ec2.VpcEndpointTypeGateway)
	} else {
		d.Set("vpc_endpoint_type", vpceType)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(vpce.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsVpcEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if _, ok := d.GetOk("auto_accept"); ok && d.Get("state").(string) == "pendingAcceptance" {
		if err := vpcEndpointAccept(conn, d.Id(), d.Get("service_name").(string)); err != nil {
			return err
		}
	}

	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
	}

	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy"))
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}

		if policy == "" {
			req.ResetPolicy = aws.Bool(true)
		} else {
			req.PolicyDocument = aws.String(policy)
		}
	}

	setVpcEndpointUpdateLists(d, "route_table_ids", &req.AddRouteTableIds, &req.RemoveRouteTableIds)
	setVpcEndpointUpdateLists(d, "subnet_ids", &req.AddSubnetIds, &req.RemoveSubnetIds)
	setVpcEndpointUpdateLists(d, "security_group_ids", &req.AddSecurityGroupIds, &req.RemoveSecurityGroupIds)

	if d.HasChange("private_dns_enabled") {
		req.PrivateDnsEnabled = aws.Bool(d.Get("private_dns_enabled").(bool))
	}

	log.Printf("[DEBUG] Updating VPC Endpoint: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(req); err != nil {
		return fmt.Errorf("Error updating VPC Endpoint: %s", err)
	}

	if err := vpcEndpointWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsVpcEndpointRead(d, meta)
}

func resourceAwsVpcEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting VPC Endpoint: %s", d.Id())
	_, err := conn.DeleteVpcEndpoints(&ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		if isAWSErr(err, "InvalidVpcEndpointId.NotFound", "") {
			log.Printf("[DEBUG] VPC Endpoint %s is already gone", d.Id())
		} else {
			return fmt.Errorf("Error deleting VPC Endpoint: %s", err)
		}
	}

	if err := vpcEndpointWaitUntilDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

func vpcEndpointAccept(conn *ec2.EC2, vpceId, svcName string) error {
	describeSvcReq := &ec2.DescribeVpcEndpointServiceConfigurationsInput{}
	describeSvcReq.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"service-name": svcName,
		},
	)

	describeSvcResp, err := conn.DescribeVpcEndpointServiceConfigurations(describeSvcReq)
	if err != nil {
		return fmt.Errorf("Error reading VPC Endpoint Service: %s", err)
	}
	if describeSvcResp == nil || len(describeSvcResp.ServiceConfigurations) == 0 {
		return fmt.Errorf("No matching VPC Endpoint Service found")
	}

	acceptEpReq := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      describeSvcResp.ServiceConfigurations[0].ServiceId,
		VpcEndpointIds: aws.StringSlice([]string{vpceId}),
	}

	log.Printf("[DEBUG] Accepting VPC Endpoint connection: %#v", acceptEpReq)
	_, err = conn.AcceptVpcEndpointConnections(acceptEpReq)
	if err != nil {
		return fmt.Errorf("Error accepting VPC Endpoint connection: %s", err)
	}

	return nil
}

func vpcEndpointStateRefresh(conn *ec2.EC2, vpceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Reading VPC Endpoint: %s", vpceId)
		resp, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: aws.StringSlice([]string{vpceId}),
		})
		if err != nil {
			if isAWSErr(err, "InvalidVpcEndpointId.NotFound", "") {
				return "", "deleted", nil
			}

			return nil, "", err
		}

		n := len(resp.VpcEndpoints)
		switch n {
		case 0:
			return "", "deleted", nil

		case 1:
			vpce := resp.VpcEndpoints[0]
			state := aws.StringValue(vpce.State)
			// No use in retrying if the endpoint is in a failed state.
			if state == "failed" {
				return nil, state, errors.New("VPC Endpoint is in a failed state")
			}
			return vpce, state, nil

		default:
			return nil, "", fmt.Errorf("Found %d VPC Endpoints for %s, expected 1", n, vpceId)
		}
	}
}

func vpcEndpointWaitUntilAvailable(conn *ec2.EC2, vpceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"available", "pendingAcceptance"},
		Refresh:    vpcEndpointStateRefresh(conn, vpceId),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for VPC Endpoint (%s) to become available: %s", vpceId, err)
	}

	return nil
}

func vpcEndpointWaitUntilDeleted(conn *ec2.EC2, vpceId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"available", "pending", "deleting"},
		Target:     []string{"deleted"},
		Refresh:    vpcEndpointStateRefresh(conn, vpceId),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func setVpcEndpointCreateList(d *schema.ResourceData, key string, c *[]*string) {
	if v, ok := d.GetOk(key); ok {
		list := v.(*schema.Set).List()
		if len(list) > 0 {
			*c = expandStringList(list)
		}
	}
}

func setVpcEndpointUpdateLists(d *schema.ResourceData, key string, a, r *[]*string) {
	if d.HasChange(key) {
		o, n := d.GetChange(key)
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		add := expandStringList(ns.Difference(os).List())
		if len(add) > 0 {
			*a = add
		}

		remove := expandStringList(os.Difference(ns).List())
		if len(remove) > 0 {
			*r = remove
		}
	}
}

func flattenVpcEndpointDnsEntries(dnsEntries []*ec2.DnsEntry) []interface{} {
	vDnsEntries := []interface{}{}

	for _, dnsEntry := range dnsEntries {
		vDnsEntries = append(vDnsEntries, map[string]interface{}{
			"dns_name":       aws.StringValue(dnsEntry.DnsName),
			"hosted_zone_id": aws.StringValue(dnsEntry.HostedZoneId),
		})
	}

	return vDnsEntries
}

func flattenVpcEndpointSecurityGroupIds(groups []*ec2.SecurityGroupIdentifier) *schema.Set {
	vSecurityGroupIds := []interface{}{}

	for _, group := range groups {
		vSecurityGroupIds = append(vSecurityGroupIds, aws.StringValue(group.GroupId))
	}

	return schema.NewSet(schema.HashString, vSecurityGroupIds)
}
