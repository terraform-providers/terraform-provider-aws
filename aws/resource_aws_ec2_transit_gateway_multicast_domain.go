package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayMulticastDomainCreate,
		Read:   resourceAwsEc2TransitGatewayMulticastDomainRead,
		Update: resourceAwsEc2TransitGatewayMulticastDomainUpdate,
		Delete: resourceAwsEc2TransitGatewayMulticastDomainDelete,

		Schema: map[string]*schema.Schema{
			"transit_gateway_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"tags": tagsSchema(),
			"association": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transit_gateway_attachment_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceAwsEc2TransitGatewayMulticastDomainAssociationsHash,
			},
			"members": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceAwsEc2TransitGatewayMulticastDomainGroupsHash,
			},
			"sources": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceAwsEc2TransitGatewayMulticastDomainGroupsHash,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayMulticastDomainCreate(d *schema.ResourceData, meta interface{}) error {
	// create the domain
	conn := meta.(*AWSClient).ec2conn
	input := &ec2.CreateTransitGatewayMulticastDomainInput{
		TransitGatewayId: aws.String(d.Get("transit_gateway_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}),
			ec2.ResourceTypeTransitGatewayMulticastDomain),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Domain: %s", input)
	output, err := conn.CreateTransitGatewayMulticastDomain(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Multicast Domain: %s", err)
	}

	id := aws.StringValue(output.TransitGatewayMulticastDomain.TransitGatewayMulticastDomainId)
	d.SetId(id)

	// wait for the domain to become available
	if err := waitForEc2TransitGatewayMulticastDomainCreation(conn, id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) availability: %s", id, err)
	}

	return resourceAwsEc2TransitGatewayMulticastDomainUpdate(d, meta)
}

func resourceAwsEc2TransitGatewayMulticastDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	multicastDomain, err := ec2DescribeTransitGatewayMulticastDomain(conn, id)
	if isAWSErr(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		return resourceAwsEc2TransitGatewayMulticastDomainNotFound(d)
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if multicastDomain == nil {
		return resourceAwsEc2TransitGatewayMulticastDomainNotFound(d)
	}

	if err := resourceAwsEc2TransitGatewayMulticastDomainAssociationsRead(d, meta); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsRead(d, meta, true); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsRead(d, meta, false); err != nil {
		return err
	}

	if aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleting ||
		aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleted {
		log.Printf(
			"[WARN] EC2 Transit Gateway Multicast Domain (%s) in deleted state (%s), removing from state",
			d.Id(), aws.StringValue(multicastDomain.State))
		d.SetId("")
		return nil
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(multicastDomain.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error settings tags: %s", err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, id, o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Multicast Domain (%s) tags: %s", id, err)
		}
	}

	if err := resourceAwsEc2TransitGatewayMulticastDomainAssociationsUpdate(d, meta); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsUpdate(d, meta, true); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsUpdate(d, meta, false); err != nil {
		return err
	}

	return resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta)
}

func resourceAwsEc2TransitGatewayMulticastDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	if err := resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta); err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
	}

	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsDeregisterAll(d, meta, true); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsDeregisterAll(d, meta, false); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainDisassociateAll(d, meta); err != nil {
		return err
	}

	input := &ec2.DeleteTransitGatewayMulticastDomainInput{
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	_, err := conn.DeleteTransitGatewayMulticastDomain(input)
	if isAWSErr(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if err := waitForEc2TransitGatewayMulticastDomainDeletion(conn, id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) deletion: %s", id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainNotFound(d *schema.ResourceData) error {
	log.Printf("[WARN] EC2 Transit Gateway (%s) not found, removing from state", d.Id())
	d.SetId("")
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	groupedRemoteAssocSet, err := resourceAwsEc2TransitGatewayMulticastDomainAssociationsReadRaw(conn, id)
	if err != nil {
		return err
	}

	localAssocSet := d.Get("association").(*schema.Set).List()

	// match local associations to remote
	assocSet := &schema.Set{F: resourceAwsEc2TransitGatewayMulticastDomainAssociationsHash}
	for _, localAssoc := range localAssocSet {
		assocData := localAssoc.(map[string]interface{})
		attachmentID := assocData["transit_gateway_attachment_id"].(string)
		subnetIDs := assocData["subnet_ids"].(*schema.Set)

		c, exists := groupedRemoteAssocSet[attachmentID]
		if !exists {
			continue
		}

		container := c.(map[string]interface{})
		remoteSubnetIDs := container["subnet_ids"].(*schema.Set)
		if len(remoteSubnetIDs.List()) == 0 {
			continue
		}

		newSubnetIDs := subnetIDs.Intersection(remoteSubnetIDs)
		if len(newSubnetIDs.List()) == 0 {
			continue
		}

		container["subnet_ids"] = remoteSubnetIDs.Difference(newSubnetIDs)
		groupedRemoteAssocSet[attachmentID] = container
		assocData["subnet_ids"] = newSubnetIDs
		assocSet.Add(assocData)
	}

	// add remaining remote associations
	for attachmentID, elem := range groupedRemoteAssocSet {
		subnetIDs := elem.(map[string]interface{})["subnet_ids"].(*schema.Set)
		if len(subnetIDs.List()) == 0 {
			continue
		}
		assocSet.Add(resourceAwsEc2TransitGatewayMulticastDomainAssociationMake(attachmentID, subnetIDs))
	}

	d.Set("association", assocSet)

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsRead(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	groupedRemoteGroups, err := resourceAwsEc2TransitGatewayMulticastDomainGroupsReadRaw(conn, id, member)
	if err != nil {
		return err
	}

	groupType := resourceAwsEc2TransitGatewayMulticastDomainGroupType(member)
	localGroups := d.Get(groupType).(*schema.Set).List()

	groups := &schema.Set{F: resourceAwsEc2TransitGatewayMulticastDomainGroupsHash}
	for _, localGroup := range localGroups {
		groupData := localGroup.(map[string]interface{})
		groupIP := groupData["group_ip_address"].(string)
		netIDs := groupData["network_interface_ids"].(*schema.Set)

		c, exists := groupedRemoteGroups[groupIP]
		if !exists {
			continue
		}

		container := c.(map[string]interface{})
		remoteNetIDs := container["network_interface_ids"].(*schema.Set)
		if len(remoteNetIDs.List()) == 0 {
			continue
		}

		newNetIDs := netIDs.Intersection(remoteNetIDs)
		if len(newNetIDs.List()) == 0 {
			continue
		}

		container["network_interface_ids"] = remoteNetIDs.Difference(newNetIDs)
		groupedRemoteGroups[groupIP] = container
		groupData["network_interface_ids"] = newNetIDs
		groups.Add(groupData)
	}

	for groupIP, elem := range groupedRemoteGroups {
		netIDs := elem.(map[string]interface{})["network_interface_ids"].(*schema.Set)
		if len(netIDs.List()) == 0 {
			continue
		}
		groups.Add(resourceAwsEc2TransitGatewayMulticastDomainGroupMake(groupIP, netIDs))
	}

	d.Set(groupType, groups)

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupMake(groupIP string, netIDs *schema.Set) map[string]interface{} {
	group := make(map[string]interface{})
	group["group_ip_address"] = groupIP
	group["network_interface_ids"] = netIDs
	return group
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationMake(attachID string, subnetIDs *schema.Set) map[string]interface{} {
	assoc := make(map[string]interface{})
	assoc["transit_gateway_attachment_id"] = attachID
	assoc["subnet_ids"] = subnetIDs
	return assoc
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsReadRaw(conn *ec2.EC2, domainID string, member bool) (map[string]interface{}, error) {
	remoteGroups, err := ec2SearchTransitGatewayMulticastDomainGroupsByType(conn, domainID, member)
	if err != nil {
		return nil, fmt.Errorf(
			"error getting EC2 Transit Gateway Multicast Domain (%s) groups: %s", domainID, err)
	}

	groupedRemoteGroups := make(map[string]interface{})
	for _, remoteGroup := range remoteGroups {
		groupIP := aws.StringValue(remoteGroup.GroupIpAddress)
		netID := aws.StringValue(remoteGroup.NetworkInterfaceId)
		if c, exists := groupedRemoteGroups[groupIP]; exists {
			container := c.(map[string]interface{})
			container["network_interface_ids"].(*schema.Set).Add(netID)
			continue
		}
		newSet := &schema.Set{F: schema.HashString}
		newSet.Add(netID)
		groupedRemoteGroups[groupIP] = resourceAwsEc2TransitGatewayMulticastDomainGroupMake(groupIP, newSet)
	}

	return groupedRemoteGroups, nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsReadRaw(conn *ec2.EC2, domainID string) (map[string]interface{}, error) {
	remoteAssocSet, err := ec2GetTransitGatewayMulticastDomainAssociations(conn, domainID)
	if err != nil {
		return nil, fmt.Errorf(
			"error getting EC2 Transit Gateway Multicast Domain (%s) associations: %s", domainID, err)
	}

	groupedAssocSet := make(map[string]interface{})
	for _, remoteAssoc := range remoteAssocSet {
		attachID := aws.StringValue(remoteAssoc.TransitGatewayAttachmentId)
		subnetID := aws.StringValue(remoteAssoc.Subnet.SubnetId)
		if c, exists := groupedAssocSet[attachID]; exists {
			container := c.(map[string]interface{})
			container["subnet_ids"].(*schema.Set).Add(subnetID)
			continue
		}
		newSet := &schema.Set{F: schema.HashString}
		newSet.Add(subnetID)
		groupedAssocSet[attachID] = resourceAwsEc2TransitGatewayMulticastDomainAssociationMake(attachID, newSet)
	}

	return groupedAssocSet, nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	if !d.HasChange("association") {
		return nil
	}

	o, n := d.GetChange("association")
	old := resourceAwsEc2TransitGatewayMulticastDomainAssociationsCompress(o.(*schema.Set).Difference(n.(*schema.Set)))
	nw := resourceAwsEc2TransitGatewayMulticastDomainAssociationsCompress(n.(*schema.Set).Difference(o.(*schema.Set)))

	if old.HashEqual(nw) {
		log.Printf(
			"[DEBUG] Flattened EC2 Transit Gateway Multicast Domain assoctiation configuration was " +
				"determined to be equivalent")
		d.Set("association", n)
		return nil
	}

	// disassociate old associations
	for _, assoc := range old.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainDisassociate(conn, id, assocData); err != nil {
			return err
		}
	}

	// save current state
	associations := o.(*schema.Set).Intersection(n.(*schema.Set))
	d.Set("association", associations)

	// associate new subnets
	for _, assoc := range nw.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainAssociate(conn, id, assocData); err != nil {
			return err
		}
		associations.Add(assoc)
		d.Set("association", associations)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsUpdate(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	key := resourceAwsEc2TransitGatewayMulticastDomainGroupType(member)

	if !d.HasChange(key) {
		return nil
	}

	o, n := d.GetChange(key)
	old := resourceAwsEc2TransitGatewayMulticastDomainGroupsCompress(o.(*schema.Set).Difference(n.(*schema.Set)))
	nw := resourceAwsEc2TransitGatewayMulticastDomainGroupsCompress(n.(*schema.Set).Difference(o.(*schema.Set)))

	if old.HashEqual(nw) {
		log.Printf(
			"[DEBUG] Flattened EC2 Transit Gateway Multicast Domain group configuration was determined to be " +
				"equivalent")
		d.Set(key, n)
		return nil
	}

	// remove old groups
	for _, group := range old.List() {
		groupData := group.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
			return err
		}
	}

	// save current state
	groups := o.(*schema.Set).Intersection(n.(*schema.Set))
	d.Set(key, groups)

	// register new groups
	for _, group := range nw.List() {
		groupData := group.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainGroupRegister(conn, id, groupData, member); err != nil {
			return err
		}
		groups.Add(group)
		d.Set(key, groups)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainDisassociate(conn *ec2.EC2, id string, assocData map[string]interface{}) error {
	tgwAttachmentID := assocData["transit_gateway_attachment_id"].(string)
	subnetIDs := expandStringSet(assocData["subnet_ids"].(*schema.Set))
	input := &ec2.DisassociateTransitGatewayMulticastDomainInput{
		SubnetIds:                       subnetIDs,
		TransitGatewayAttachmentId:      aws.String(tgwAttachmentID),
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf(
		"[DEBUG] Disassociating subnets from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	if _, err := conn.DisassociateTransitGatewayMulticastDomain(input); err != nil {
		return fmt.Errorf(
			"error disassociating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
	}

	// wait for subnets to be disassociated
	if err := waitForEc2TransitGatewayMulticastDomainDisassociation(conn, id, subnetIDs); err != nil {
		return fmt.Errorf(
			"error waiting for EC2 Transit Gateway Multicast Domain (%s) to disassociate subnets: %s",
			id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociate(conn *ec2.EC2, id string, assocData map[string]interface{}) error {
	tgwAttachmentID := assocData["transit_gateway_attachment_id"].(string)
	subnetIDs := expandStringSet(assocData["subnet_ids"].(*schema.Set))
	input := &ec2.AssociateTransitGatewayMulticastDomainInput{
		SubnetIds:                       subnetIDs,
		TransitGatewayAttachmentId:      aws.String(tgwAttachmentID),
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf(
		"[DEBUG] Associating subnets to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	if _, err := conn.AssociateTransitGatewayMulticastDomain(input); err != nil {
		return fmt.Errorf(
			"error associating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
	}

	// wait for associations
	if err := waitForEc2TransitGatewayMulticastDomainAssociation(conn, id, subnetIDs); err != nil {
		return fmt.Errorf(
			"error waiting for EC2 Transit Gateway Multicast Domain (%s) associations: %s", id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupDeregister(conn *ec2.EC2, id string, groupData map[string]interface{}, member bool) error {
	// Note: for some reason AWS made the decision to logically separate "members" from "sources" in
	// register/deregister; however, they are unified in "search"
	if member {
		input := &ec2.DeregisterTransitGatewayMulticastGroupMembersInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Deregistering members from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.DeregisterTransitGatewayMulticastGroupMembers(input); err != nil {
			return fmt.Errorf(
				"error removing member from EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
		}
	} else {
		input := &ec2.DeregisterTransitGatewayMulticastGroupSourcesInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Deregistering sources from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.DeregisterTransitGatewayMulticastGroupSources(input); err != nil {
			return fmt.Errorf(
				"error removing source from EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
		}
	}

	if err := waitForEc2TransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
		return fmt.Errorf(
			"error validating EC2 Transit Gateway Multicast Domain (%s) group deregistration: %s", id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupRegister(conn *ec2.EC2, id string, groupData map[string]interface{}, member bool) error {
	// Note: for some reason AWS made the decision to logically separate "members" from "sources" in
	// register/deregister; however, they are unified in "search"
	if member {
		input := &ec2.RegisterTransitGatewayMulticastGroupMembersInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Registering members to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.RegisterTransitGatewayMulticastGroupMembers(input); err != nil {
			return fmt.Errorf(
				"error registering EC2 Transit Gateway Multicast Domain (%s) members: %s", id, err)
		}
	} else {
		input := &ec2.RegisterTransitGatewayMulticastGroupSourcesInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Registering sources to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.RegisterTransitGatewayMulticastGroupSources(input); err != nil {
			return fmt.Errorf(
				"error registering EC2 Transit Gateway Multicast Domain (%s) sources: %s", id, err)
		}
	}

	if err := waitForEc2TransitGatewayMulticastDomainGroupRegister(conn, id, groupData, member); err != nil {
		return fmt.Errorf(
			"error validating EC2 Transit Gateway Multicast Domain (%s) group registration: %s", id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainDisassociateAll(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	// disassociate subnets
	assocSet := d.Get("association").(*schema.Set)
	for _, assoc := range assocSet.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainDisassociate(conn, id, assocData); err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsDeregisterAll(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	key := resourceAwsEc2TransitGatewayMulticastDomainGroupType(member)
	// deregister groups
	groups := d.Get(key).(*schema.Set)
	for _, group := range groups.List() {
		groupData := group.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsHash(meta interface{}) int {
	m, castOk := meta.(map[string]interface{})
	if !castOk {
		return 0
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", m["transit_gateway_attachment_id"].(string)))
	for _, subnetID := range m["subnet_ids"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s,", subnetID.(string)))
	}
	return hashcode.String(buf.String())
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsHash(meta interface{}) int {
	m, castOk := meta.(map[string]interface{})
	if !castOk {
		return 0
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", m["group_ip_address"].(string)))
	for _, networkID := range m["network_interface_ids"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s,", networkID.(string)))
	}
	return hashcode.String(buf.String())
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupType(member bool) string {
	if member {
		return "members"
	}
	return "sources"
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsCompress(assocSet *schema.Set) *schema.Set {
	return resourceAwsEc2TransitGatewayMulticastDomainCompress(
		assocSet, "transit_gateway_attachment_id", "subnet_ids",
		resourceAwsEc2TransitGatewayMulticastDomainAssociationsHash)
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsCompress(groupSet *schema.Set) *schema.Set {
	return resourceAwsEc2TransitGatewayMulticastDomainCompress(
		groupSet, "group_ip_address", "network_interface_ids",
		resourceAwsEc2TransitGatewayMulticastDomainGroupsHash)
}

func resourceAwsEc2TransitGatewayMulticastDomainCompress(set *schema.Set, groupingKey string, compressKey string, setFunc schema.SchemaSetFunc) *schema.Set {
	groups := make(map[string][]*string)
	for _, elem := range set.List() {
		data := elem.(map[string]interface{})
		groupVal := data[groupingKey].(string)
		val := expandStringSet(data[compressKey].(*schema.Set))
		if existing, exists := groups[groupVal]; exists {
			groups[groupVal] = append(existing, val...)
			continue
		}
		groups[groupVal] = val
	}

	compressed := &schema.Set{F: setFunc}
	for group, values := range groups {
		data := make(map[string]interface{})
		data[groupingKey] = group
		data[compressKey] = flattenStringSet(values)
		compressed.Add(data)
	}

	return compressed
}
