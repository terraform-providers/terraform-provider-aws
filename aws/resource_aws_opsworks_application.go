package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsOpsworksApplication() *schema.Resource {
	return &schema.Resource{

		Create: resourceAwsOpsworksApplicationCreate,
		Read:   resourceAwsOpsworksApplicationRead,
		Update: resourceAwsOpsworksApplicationUpdate,
		Delete: resourceAwsOpsworksApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					opsworks.AppTypeAwsFlowRuby,
					opsworks.AppTypeJava,
					opsworks.AppTypeRails,
					opsworks.AppTypePhp,
					opsworks.AppTypeNodejs,
					opsworks.AppTypeStatic,
					opsworks.AppTypeOther,
				}, false),
			},
			"stack_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// TODO: the following 4 vals are really part of the Attributes array. We should validate that only ones relevant to the chosen type are set, perhaps. (what is the default type? how do they map?)
			"document_root": {
				Type:     schema.TypeString,
				Optional: true,
				//Default:  "public",
			},
			"rails_env": {
				Type:     schema.TypeString,
				Optional: true,
				//Default:  "production",
			},
			"auto_bundle_on_deploy": {
				Type:     schema.TypeString,
				Optional: true,
				//Default:  true,
			},
			"aws_flow_ruby_settings": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"app_source": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								opsworks.SourceTypeGit,
								opsworks.SourceTypeSvn,
								opsworks.SourceTypeArchive,
								opsworks.SourceTypeS3,
								"other",
							}, false),
						},

						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},

						"revision": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"ssh_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"data_source_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AutoSelectOpsworksMysqlInstance",
					"OpsworksMysqlInstance",
					"RdsDbInstance",
					"None",
				}, false),
			},
			"data_source_database_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_source_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domains": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"environment": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secure": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"enable_ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ssl_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				//Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(v interface{}) string {
								switch v := v.(type) {
								case string:
									return strings.TrimSpace(v)
								default:
									return ""
								}
							},
						},
						"private_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							StateFunc: func(v interface{}) string {
								switch v := v.(type) {
								case string:
									return strings.TrimSpace(v)
								default:
									return ""
								}
							},
						},
						"chain": {
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(v interface{}) string {
								switch v := v.(type) {
								case string:
									return strings.TrimSpace(v)
								default:
									return ""
								}
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsOpsworksApplicationValidate(d *schema.ResourceData) error {
	appSourceCount := d.Get("app_source.#").(int)
	if appSourceCount > 1 {
		return fmt.Errorf("Only one app_source is permitted.")
	}

	sslCount := d.Get("ssl_configuration.#").(int)
	if sslCount > 1 {
		return fmt.Errorf("Only one ssl_configuration is permitted.")
	}

	if d.Get("type") == opsworks.AppTypeNodejs || d.Get("type") == opsworks.AppTypeJava {
		// allowed attributes: none
		if d.Get("document_root").(string) != "" || d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" || d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("No additional attributes are allowed for app type '%s'.", d.Get("type").(string))
		}
	} else if d.Get("type") == opsworks.AppTypeRails {
		// allowed attributes: document_root, rails_env, auto_bundle_on_deploy
		if d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("Only 'document_root, rails_env, auto_bundle_on_deploy' are allowed for app type '%s'.", opsworks.AppTypeRails)
		}
		// rails_env is required
		if _, ok := d.GetOk("rails_env"); !ok {
			return fmt.Errorf("Set rails_env must be set if type is set to rails.")
		}
	} else if d.Get("type") == opsworks.AppTypePhp || d.Get("type") == opsworks.AppTypeStatic || d.Get("type") == opsworks.AppTypeOther {
		log.Printf("[DEBUG] the app type is : %s", d.Get("type").(string))
		log.Printf("[DEBUG] the attributes are: document_root '%s', rails_env '%s', auto_bundle_on_deploy '%s', aws_flow_ruby_settings '%s'", d.Get("document_root").(string), d.Get("rails_env").(string), d.Get("auto_bundle_on_deploy").(string), d.Get("aws_flow_ruby_settings").(string))
		// allowed attributes: document_root
		if d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" || d.Get("aws_flow_ruby_settings").(string) != "" {
			return fmt.Errorf("Only 'document_root' is allowed for app type '%s'.", d.Get("type").(string))
		}
	} else if d.Get("type") == opsworks.AppTypeAwsFlowRuby {
		// allowed attributes: aws_flow_ruby_settings
		if d.Get("document_root").(string) != "" || d.Get("rails_env").(string) != "" || d.Get("auto_bundle_on_deploy").(string) != "" {
			return fmt.Errorf("Only 'aws_flow_ruby_settings' is allowed for app type '%s'.", d.Get("type").(string))
		}
	}

	return nil
}

func resourceAwsOpsworksApplicationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	req := &opsworks.DescribeAppsInput{
		AppIds: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks app: %s", d.Id())

	resp, err := client.DescribeApps(req)
	if err != nil {
		if isAWSErr(err, opsworks.ErrCodeResourceNotFoundException, "") {
			log.Printf("[INFO] App not found: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	app := resp.Apps[0]
	log.Printf("[DEBUG] Opsworks Application: %#v", app)

	d.Set("name", app.Name)
	d.Set("stack_id", app.StackId)
	d.Set("type", app.Type)
	d.Set("description", app.Description)
	d.Set("domains", flattenStringList(app.Domains))
	d.Set("enable_ssl", app.EnableSsl)
	err = resourceAwsOpsworksSetApplicationSsl(d, app.SslConfiguration)
	if err != nil {
		return err
	}
	err = resourceAwsOpsworksSetApplicationSource(d, app.AppSource)
	if err != nil {
		return err
	}
	resourceAwsOpsworksSetApplicationDataSources(d, app.DataSources)
	resourceAwsOpsworksSetApplicationEnvironmentVariable(d, app.Environment)
	resourceAwsOpsworksSetApplicationAttributes(d, app.Attributes)

	return nil
}

func resourceAwsOpsworksApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	err := resourceAwsOpsworksApplicationValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.CreateAppInput{
		Name:             aws.String(d.Get("name").(string)),
		Shortname:        aws.String(d.Get("short_name").(string)),
		StackId:          aws.String(d.Get("stack_id").(string)),
		Type:             aws.String(d.Get("type").(string)),
		Description:      aws.String(d.Get("description").(string)),
		Domains:          expandStringList(d.Get("domains").([]interface{})),
		EnableSsl:        aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: resourceAwsOpsworksApplicationSsl(d),
		AppSource:        resourceAwsOpsworksApplicationSource(d),
		DataSources:      resourceAwsOpsworksApplicationDataSources(d),
		Environment:      resourceAwsOpsworksApplicationEnvironmentVariable(d),
		Attributes:       resourceAwsOpsworksApplicationAttributes(d),
	}

	resp, err := client.CreateApp(req)
	if err != nil {
		return fmt.Errorf("Error creating OpsWorks application: %s", err)
	}

	appID := *resp.AppId
	d.SetId(appID)

	return resourceAwsOpsworksApplicationRead(d, meta)
}

func resourceAwsOpsworksApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	err := resourceAwsOpsworksApplicationValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.UpdateAppInput{
		AppId:            aws.String(d.Id()),
		Name:             aws.String(d.Get("name").(string)),
		Type:             aws.String(d.Get("type").(string)),
		Description:      aws.String(d.Get("description").(string)),
		Domains:          expandStringList(d.Get("domains").([]interface{})),
		EnableSsl:        aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: resourceAwsOpsworksApplicationSsl(d),
		AppSource:        resourceAwsOpsworksApplicationSource(d),
		DataSources:      resourceAwsOpsworksApplicationDataSources(d),
		Environment:      resourceAwsOpsworksApplicationEnvironmentVariable(d),
		Attributes:       resourceAwsOpsworksApplicationAttributes(d),
	}

	log.Printf("[DEBUG] Updating OpsWorks layer: %s", d.Id())

	_, err = client.UpdateApp(req)
	if err != nil {
		return fmt.Errorf("Error updating OpsWorks app: %s", err)
	}

	return resourceAwsOpsworksApplicationRead(d, meta)
}

func resourceAwsOpsworksApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	req := &opsworks.DeleteAppInput{
		AppId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks application: %s", d.Id())

	_, err := client.DeleteApp(req)
	return err
}

func resourceAwsOpsworksFindEnvironmentVariable(key string, vs []*opsworks.EnvironmentVariable) *opsworks.EnvironmentVariable {
	for _, v := range vs {
		if aws.StringValue(v.Key) == key {
			return v
		}
	}
	return nil
}

func resourceAwsOpsworksSetApplicationEnvironmentVariable(d *schema.ResourceData, vs []*opsworks.EnvironmentVariable) {
	if len(vs) == 0 {
		d.Set("environment", nil)
		return
	}

	// sensitive variables are returned obfuscated from the API, this creates a
	// permadiff between the obfuscated API response and the config value. We
	// start with the existing state so it can passthrough when the key is secure
	values := d.Get("environment").(*schema.Set).List()

	for i := 0; i < len(values); i++ {
		value := values[i].(map[string]interface{})
		if v := resourceAwsOpsworksFindEnvironmentVariable(value["key"].(string), vs); v != nil {
			if !aws.BoolValue(v.Secure) {
				value["secure"] = aws.BoolValue(v.Secure)
				value["key"] = aws.StringValue(v.Key)
				value["value"] = aws.StringValue(v.Value)
				values[i] = value
			}
		} else {
			// delete if not found in API response
			values = append(values[:i], values[i+1:]...)
		}
	}

	d.Set("environment", values)
}

func resourceAwsOpsworksApplicationEnvironmentVariable(d *schema.ResourceData) []*opsworks.EnvironmentVariable {
	environmentVariables := d.Get("environment").(*schema.Set).List()
	result := make([]*opsworks.EnvironmentVariable, len(environmentVariables))

	for i := 0; i < len(environmentVariables); i++ {
		env := environmentVariables[i].(map[string]interface{})

		result[i] = &opsworks.EnvironmentVariable{
			Key:    aws.String(env["key"].(string)),
			Value:  aws.String(env["value"].(string)),
			Secure: aws.Bool(env["secure"].(bool)),
		}
	}
	return result
}

func resourceAwsOpsworksApplicationSource(d *schema.ResourceData) *opsworks.Source {
	count := d.Get("app_source.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.Source{
		Type:     aws.String(d.Get("app_source.0.type").(string)),
		Url:      aws.String(d.Get("app_source.0.url").(string)),
		Username: aws.String(d.Get("app_source.0.username").(string)),
		Password: aws.String(d.Get("app_source.0.password").(string)),
		Revision: aws.String(d.Get("app_source.0.revision").(string)),
		SshKey:   aws.String(d.Get("app_source.0.ssh_key").(string)),
	}
}

func resourceAwsOpsworksSetApplicationSource(d *schema.ResourceData, v *opsworks.Source) error {
	nv := make([]interface{}, 0, 1)
	if v != nil {
		m := make(map[string]interface{})
		if v.Type != nil {
			m["type"] = *v.Type
		}
		if v.Url != nil {
			m["url"] = *v.Url
		}
		if v.Username != nil {
			m["username"] = *v.Username
		}
		if v.Revision != nil {
			m["revision"] = *v.Revision
		}

		// v.Password and v.SshKey will, on read, contain the placeholder string
		// "*****FILTERED*****", so we ignore it on read and let persist
		// the value already in the state.
		m["password"] = d.Get("app_source.0.password").(string)
		m["ssh_key"] = d.Get("app_source.0.ssh_key").(string)

		nv = append(nv, m)
	}

	err := d.Set("app_source", nv)
	if err != nil {
		// should never happen
		return err
	}
	return nil
}

func resourceAwsOpsworksApplicationDataSources(d *schema.ResourceData) []*opsworks.DataSource {
	arn := d.Get("data_source_arn").(string)
	databaseName := d.Get("data_source_database_name").(string)
	databaseType := d.Get("data_source_type").(string)

	result := make([]*opsworks.DataSource, 1)

	if len(arn) > 0 || len(databaseName) > 0 || len(databaseType) > 0 {
		result[0] = &opsworks.DataSource{
			Arn:          aws.String(arn),
			DatabaseName: aws.String(databaseName),
			Type:         aws.String(databaseType),
		}
	}
	return result
}

func resourceAwsOpsworksSetApplicationDataSources(d *schema.ResourceData, v []*opsworks.DataSource) {
	d.Set("data_source_arn", nil)
	d.Set("data_source_database_name", nil)
	d.Set("data_source_type", nil)

	if len(v) == 0 {
		return
	}

	d.Set("data_source_arn", v[0].Arn)
	d.Set("data_source_database_name", v[0].DatabaseName)
	d.Set("data_source_type", v[0].Type)
}

func resourceAwsOpsworksApplicationSsl(d *schema.ResourceData) *opsworks.SslConfiguration {
	count := d.Get("ssl_configuration.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.SslConfiguration{
		PrivateKey:  aws.String(d.Get("ssl_configuration.0.private_key").(string)),
		Certificate: aws.String(d.Get("ssl_configuration.0.certificate").(string)),
		Chain:       aws.String(d.Get("ssl_configuration.0.chain").(string)),
	}
}

func resourceAwsOpsworksSetApplicationSsl(d *schema.ResourceData, v *opsworks.SslConfiguration) error {
	nv := make([]interface{}, 0, 1)
	set := false
	if v != nil {
		m := make(map[string]interface{})
		if v.PrivateKey != nil {
			m["private_key"] = *v.PrivateKey
			set = true
		}
		if v.Certificate != nil {
			m["certificate"] = *v.Certificate
			set = true
		}
		if v.Chain != nil {
			m["chain"] = *v.Chain
			set = true
		}
		if set {
			nv = append(nv, m)
		}
	}

	err := d.Set("ssl_configuration", nv)
	if err != nil {
		// should never happen
		return err
	}
	return nil
}

func resourceAwsOpsworksApplicationAttributes(d *schema.ResourceData) map[string]*string {
	attributes := make(map[string]*string)

	if val := d.Get("document_root").(string); len(val) > 0 {
		attributes[opsworks.AppAttributesKeysDocumentRoot] = aws.String(val)
	}
	if val := d.Get("aws_flow_ruby_settings").(string); len(val) > 0 {
		attributes[opsworks.AppAttributesKeysAwsFlowRubySettings] = aws.String(val)
	}
	if val := d.Get("rails_env").(string); len(val) > 0 {
		attributes[opsworks.AppAttributesKeysRailsEnv] = aws.String(val)
	}
	if val := d.Get("auto_bundle_on_deploy").(string); len(val) > 0 {
		if val == "1" {
			val = "true"
		} else if val == "0" {
			val = "false"
		}
		attributes[opsworks.AppAttributesKeysAutoBundleOnDeploy] = aws.String(val)
	}

	return attributes
}

func resourceAwsOpsworksSetApplicationAttributes(d *schema.ResourceData, v map[string]*string) {
	d.Set("document_root", nil)
	d.Set("rails_env", nil)
	d.Set("aws_flow_ruby_settings", nil)
	d.Set("auto_bundle_on_deploy", nil)

	if d.Get("type") == opsworks.AppTypeNodejs || d.Get("type") == opsworks.AppTypeJava {
		return
	} else if d.Get("type") == opsworks.AppTypeRails {
		if val, ok := v[opsworks.AppAttributesKeysDocumentRoot]; ok {
			d.Set("document_root", val)
		}
		if val, ok := v[opsworks.AppAttributesKeysRailsEnv]; ok {
			d.Set("rails_env", val)
		}
		if val, ok := v[opsworks.AppAttributesKeysAutoBundleOnDeploy]; ok {
			d.Set("auto_bundle_on_deploy", val)
		}
		return
	} else if d.Get("type") == opsworks.AppTypePhp || d.Get("type") == opsworks.AppTypeStatic || d.Get("type") == opsworks.AppTypeOther {
		if val, ok := v[opsworks.AppAttributesKeysDocumentRoot]; ok {
			d.Set("document_root", val)
		}
		return
	} else if d.Get("type") == opsworks.AppTypeAwsFlowRuby {
		if val, ok := v[opsworks.AppAttributesKeysAwsFlowRubySettings]; ok {
			d.Set("aws_flow_ruby_settings", val)
		}
		return
	}

}
