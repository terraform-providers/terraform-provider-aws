package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directoryservice/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directoryservice/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsDirectoryServiceDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDirectoryServiceDirectoryCreate,
		Read:   resourceAwsDirectoryServiceDirectoryRead,
		Update: resourceAwsDirectoryServiceDirectoryUpdate,
		Delete: resourceAwsDirectoryServiceDirectoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"size": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directoryservice.DirectorySizeLarge,
					directoryservice.DirectorySizeSmall,
				}, false),
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"vpc_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"connect_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"customer_username": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
							Set: schema.HashString,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  directoryservice.DirectoryTypeSimpleAd,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directoryservice.DirectoryTypeAdconnector,
					directoryservice.DirectoryTypeMicrosoftAd,
					directoryservice.DirectoryTypeSimpleAd,
					directoryservice.DirectoryTypeSharedMicrosoftAd,
				}, false),
			},
			"edition": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directoryservice.DirectoryEditionEnterprise,
					directoryservice.DirectoryEditionStandard,
				}, false),
			},
			"radius_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(directoryservice.RadiusAuthenticationProtocol_Values(), false),
						},
						"label": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"retries": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 10),
						},
						"servers": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
							Set: schema.HashString,
						},
						"timeout": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"secret": {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(8, 512),
						},
						"same_username": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func buildVpcSettings(d *schema.ResourceData) (vpcSettings *directoryservice.DirectoryVpcSettings, err error) {
	v, ok := d.GetOk("vpc_settings")
	if !ok {
		return nil, fmt.Errorf("vpc_settings is required for type = SimpleAD or MicrosoftAD")
	}
	settings := v.([]interface{})
	s := settings[0].(map[string]interface{})
	var subnetIds []*string
	for _, id := range s["subnet_ids"].(*schema.Set).List() {
		subnetIds = append(subnetIds, aws.String(id.(string)))
	}

	vpcSettings = &directoryservice.DirectoryVpcSettings{
		SubnetIds: subnetIds,
		VpcId:     aws.String(s["vpc_id"].(string)),
	}

	return vpcSettings, nil
}

func buildConnectSettings(d *schema.ResourceData) (connectSettings *directoryservice.DirectoryConnectSettings, err error) {
	v, ok := d.GetOk("connect_settings")
	if !ok {
		return nil, fmt.Errorf("connect_settings is required for type = ADConnector")
	}
	settings := v.([]interface{})
	s := settings[0].(map[string]interface{})

	var subnetIds []*string
	for _, id := range s["subnet_ids"].(*schema.Set).List() {
		subnetIds = append(subnetIds, aws.String(id.(string)))
	}

	var customerDnsIps []*string
	for _, id := range s["customer_dns_ips"].(*schema.Set).List() {
		customerDnsIps = append(customerDnsIps, aws.String(id.(string)))
	}

	connectSettings = &directoryservice.DirectoryConnectSettings{
		CustomerDnsIps:   customerDnsIps,
		CustomerUserName: aws.String(s["customer_username"].(string)),
		SubnetIds:        subnetIds,
		VpcId:            aws.String(s["vpc_id"].(string)),
	}

	return connectSettings, nil
}

func createDirectoryConnector(dsconn *directoryservice.DirectoryService, d *schema.ResourceData, meta interface{}) (directoryId string, err error) {
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := directoryservice.ConnectDirectoryInput{
		Name:     aws.String(d.Get("name").(string)),
		Password: aws.String(d.Get("password").(string)),
		Tags:     tags.IgnoreAws().DirectoryserviceTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("size"); ok {
		input.Size = aws.String(v.(string))
	} else {
		// Matching previous behavior of Default: "Large" for Size attribute
		input.Size = aws.String(directoryservice.DirectorySizeLarge)
	}
	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	input.ConnectSettings, err = buildConnectSettings(d)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Creating Directory Connector: %s", input)
	out, err := dsconn.ConnectDirectory(&input)
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Directory Connector created: %s", out)

	return *out.DirectoryId, nil
}

func createSimpleDirectoryService(dsconn *directoryservice.DirectoryService, d *schema.ResourceData, meta interface{}) (directoryId string, err error) {
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := directoryservice.CreateDirectoryInput{
		Name:     aws.String(d.Get("name").(string)),
		Password: aws.String(d.Get("password").(string)),
		Tags:     tags.IgnoreAws().DirectoryserviceTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("size"); ok {
		input.Size = aws.String(v.(string))
	} else {
		// Matching previous behavior of Default: "Large" for Size attribute
		input.Size = aws.String(directoryservice.DirectorySizeLarge)
	}
	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}

	input.VpcSettings, err = buildVpcSettings(d)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Creating Simple Directory Service: %s", input)
	out, err := dsconn.CreateDirectory(&input)
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Simple Directory Service created: %s", out)

	return *out.DirectoryId, nil
}

func createActiveDirectoryService(dsconn *directoryservice.DirectoryService, d *schema.ResourceData, meta interface{}) (directoryId string, err error) {
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := directoryservice.CreateMicrosoftADInput{
		Name:     aws.String(d.Get("name").(string)),
		Password: aws.String(d.Get("password").(string)),
		Tags:     tags.IgnoreAws().DirectoryserviceTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("short_name"); ok {
		input.ShortName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("edition"); ok {
		input.Edition = aws.String(v.(string))
	}

	input.VpcSettings, err = buildVpcSettings(d)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Creating Microsoft AD Directory Service: %s", input)
	out, err := dsconn.CreateMicrosoftAD(&input)
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Microsoft AD Directory Service created: %s", out)

	return *out.DirectoryId, nil
}

func enableDirectoryServiceSso(dsconn *directoryservice.DirectoryService, d *schema.ResourceData) error {
	if v, ok := d.GetOk("enable_sso"); ok && v.(bool) {
		log.Printf("[DEBUG] Enabling SSO for DS directory %q", d.Id())
		if _, err := dsconn.EnableSso(&directoryservice.EnableSsoInput{
			DirectoryId: aws.String(d.Id()),
		}); err != nil {
			return fmt.Errorf("Error Enabling SSO for DS directory %s: %s", d.Id(), err)
		}
	} else {
		log.Printf("[DEBUG] Disabling SSO for DS directory %q", d.Id())
		if _, err := dsconn.DisableSso(&directoryservice.DisableSsoInput{
			DirectoryId: aws.String(d.Id()),
		}); err != nil {
			return fmt.Errorf("Error Disabling SSO for DS directory %s: %s", d.Id(), err)
		}
	}

	return nil
}

func enableRadiusMfa(dsconn *directoryservice.DirectoryService, d *schema.ResourceData) error {
	if v, ok := d.GetOk("radius_settings"); ok {
		directoryType := d.Get("type").(string)

		if directoryType == directoryservice.DirectoryTypeSimpleAd {
			return fmt.Errorf("Multi-factor authentication is not available for Simple AD.")
		}

		settings := v.([]interface{})
		s := settings[0].(map[string]interface{})

		var radiusServers []*string
		for _, id := range s["servers"].(*schema.Set).List() {
			radiusServers = append(radiusServers, aws.String(id.(string)))
		}

		input := directoryservice.EnableRadiusInput{
			DirectoryId: aws.String(d.Id()),
			RadiusSettings: &directoryservice.RadiusSettings{
				AuthenticationProtocol: aws.String(s["protocol"].(string)),
				DisplayLabel:           aws.String(s["label"].(string)),
				RadiusPort:             aws.Int64(int64(s["port"].(int))),
				RadiusRetries:          aws.Int64(int64(s["retries"].(int))),
				RadiusServers:          radiusServers,
				RadiusTimeout:          aws.Int64(int64(s["timeout"].(int))),
				SharedSecret:           aws.String(s["secret"].(string)),
				UseSameUsername:        aws.Bool(s["same_username"].(bool)),
			},
		}

		log.Printf("[DEBUG] Enabling Radius MFA for DS directory %q", d.Id())
		if _, err := dsconn.EnableRadius(&input); err != nil {
			return fmt.Errorf("Error Enabling Radius MFA for DS directory %s: %s", d.Id(), err)
		}
	} else {
		log.Printf("[DEBUG] Disabling Radius MFA for DS directory %q", d.Id())
		if _, err := dsconn.DisableRadius(&directoryservice.DisableRadiusInput{
			DirectoryId: aws.String(d.Id()),
		}); err != nil {
			return fmt.Errorf("Error Disabling Radius MFA for DS directory %s: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsDirectoryServiceDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	dsconn := meta.(*AWSClient).dsconn

	var directoryId string
	var err error
	directoryType := d.Get("type").(string)

	if directoryType == directoryservice.DirectoryTypeAdconnector {
		directoryId, err = createDirectoryConnector(dsconn, d, meta)
	} else if directoryType == directoryservice.DirectoryTypeMicrosoftAd {
		directoryId, err = createActiveDirectoryService(dsconn, d, meta)
	} else if directoryType == directoryservice.DirectoryTypeSimpleAd {
		directoryId, err = createSimpleDirectoryService(dsconn, d, meta)
	}

	if err != nil {
		return err
	}

	d.SetId(directoryId)

	_, err = waiter.DirectoryCreated(dsconn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Directory Service Directory (%s) to create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("alias"); ok {
		input := directoryservice.CreateAliasInput{
			DirectoryId: aws.String(d.Id()),
			Alias:       aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Assigning alias %q to DS directory %q",
			v.(string), d.Id())
		out, err := dsconn.CreateAlias(&input)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Alias %q assigned to DS directory %q",
			*out.Alias, *out.DirectoryId)
	}

	if d.HasChange("enable_sso") {
		if err := enableDirectoryServiceSso(dsconn, d); err != nil {
			return err
		}
	}

	if d.HasChange("radius_settings") {
		if err := enableRadiusMfa(dsconn, d); err != nil {
			return err
		}
	}

	return resourceAwsDirectoryServiceDirectoryRead(d, meta)
}

func resourceAwsDirectoryServiceDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	dsconn := meta.(*AWSClient).dsconn

	if d.HasChange("enable_sso") {
		if err := enableDirectoryServiceSso(dsconn, d); err != nil {
			return err
		}
	}

	if d.HasChange("radius_settings") {
		if err := enableRadiusMfa(dsconn, d); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.DirectoryserviceUpdateTags(dsconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Directory Service Directory (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsDirectoryServiceDirectoryRead(d, meta)
}

func resourceAwsDirectoryServiceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	dsconn := meta.(*AWSClient).dsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	dir, err := finder.DirectoryByID(dsconn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Directory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Directory Service Directory (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received DS directory: %s", dir)

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	d.Set("description", dir.Description)

	if *dir.Type == directoryservice.DirectoryTypeAdconnector {
		d.Set("dns_ip_addresses", flattenStringSet(dir.ConnectSettings.ConnectIps))
	} else {
		d.Set("dns_ip_addresses", flattenStringSet(dir.DnsIpAddrs))
	}
	d.Set("name", dir.Name)
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("edition", dir.Edition)
	d.Set("type", dir.Type)

	if err := d.Set("vpc_settings", flattenDSVpcSettings(dir.VpcSettings)); err != nil {
		return fmt.Errorf("error setting VPC settings: %s", err)
	}

	if err := d.Set("connect_settings", flattenDSConnectSettings(dir.DnsIpAddrs, dir.ConnectSettings)); err != nil {
		return fmt.Errorf("error setting connect settings: %s", err)
	}

	d.Set("enable_sso", dir.SsoEnabled)

	radiusSecret := aws.String("")
	if v, ok := d.GetOk("radius_settings"); ok && len(v.([]interface{})) > 0 {
		settings := v.([]interface{})
		s := settings[0].(map[string]interface{})
		radiusSecret = aws.String(s["secret"].(string))
	}

	if err := d.Set("radius_settings", flattenDSRadiusSettings(dir.RadiusSettings, radiusSecret)); err != nil {
		return fmt.Errorf("error setting radius_settings: %s", err)
	}

	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("security_group_id", dir.ConnectSettings.SecurityGroupId)
	} else {
		d.Set("security_group_id", dir.VpcSettings.SecurityGroupId)
	}

	tags, err := keyvaluetags.DirectoryserviceListTags(dsconn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Directory Service Directory (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsDirectoryServiceDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dsconn

	input := &directoryservice.DeleteDirectoryInput{
		DirectoryId: aws.String(d.Id()),
	}

	_, err := conn.DeleteDirectory(input)

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Directory Service Directory (%s): %w", d.Id(), err)
	}

	_, err = waiter.DirectoryDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Directory Service Directory (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
