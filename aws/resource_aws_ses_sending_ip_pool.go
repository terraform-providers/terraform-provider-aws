package aws

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSesSendingIpPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesSendingIpPoolCreate,
		Read:   resourceAwsSesSendingIpPoolRead,
		Update: resourceAwsSesSendingIpPoolUpdate,
		Delete: resourceAwsSesSendingIpPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsSesSendingIpPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesv2Conn

	poolName := d.Get("name").(string)
	_, err := conn.CreateDedicatedIpPool(&sesv2.CreateDedicatedIpPoolInput{
		PoolName: aws.String(poolName),
	})
	if err != nil {
		return fmt.Errorf("Error creating SESv2 ip pool: %s", err)
	}

	d.SetId(poolName)

	// Set other properties of the sending pool
	if err := resourceAwsSesSendingIpPoolUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsSesSendingIpPoolRead(d, meta)
}

func resourceAwsSesSendingIpPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesv2Conn

	response, err := conn.ListDedicatedIpPools(&sesv2.ListDedicatedIpPoolsInput{
		PageSize: aws.Int64(100),
	})
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error reading SESv2 ip pool: %v", err)
	}
	for i := range response.DedicatedIpPools {
		if n := *response.DedicatedIpPools[i]; strings.EqualFold(n, d.Id()) {
			d.Set("name", d.Id())

			ips, err := readDedicatedIPs(conn, d.Id())
			if err != nil {
				return err
			}
			return d.Set("ips", ips)
		}
	}
	return fmt.Errorf("unable to find %s sending pool", d.Id())
}

func readDedicatedIPs(conn *sesv2.SESV2, poolName string) ([]string, error) {
	resp, err := conn.GetDedicatedIps(&sesv2.GetDedicatedIpsInput{
		PageSize: aws.Int64(100),
		PoolName: aws.String(poolName),
	})
	if err != nil {
		return nil, err
	}

	var out []string
	for i := range resp.DedicatedIps {
		if ip := resp.DedicatedIps[i]; ip != nil && ip.Ip != nil {
			out = append(out, *ip.Ip)
		}
	}
	sort.Strings(out)
	return out, nil
}

func resourceAwsSesSendingIpPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesv2Conn

	name := d.Get("name").(string)

	if v, ok := d.GetOk("ips"); ok {
		ips := v.([]interface{})
		for i := range ips {
			_, err := conn.PutDedicatedIpInPool(&sesv2.PutDedicatedIpInPoolInput{
				DestinationPoolName: aws.String(name),
				Ip:                  aws.String(ips[i].(string)),
			})
			if err != nil {
				return fmt.Errorf("Error adding IP to pool: %v", err)
			}
		}
	}

	// Read all IP's on the pool
	ips, err := readDedicatedIPs(conn, d.Id())
	if err != nil {
		return err
	}
	return d.Set("ips", ips)
}

func resourceAwsSesSendingIpPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesv2Conn
	log.Printf("[DEBUG] SES Delete Sending IP Pool: id=%s name=%s", d.Id(), d.Get("name").(string))
	_, err := conn.DeleteDedicatedIpPool(&sesv2.DeleteDedicatedIpPoolInput{
		PoolName: aws.String(d.Id()),
	})
	return err
}
