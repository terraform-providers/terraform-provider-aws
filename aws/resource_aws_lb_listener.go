package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elbv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elbv2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsLbListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbListenerCreate,
		Read:   resourceAwsLbListenerRead,
		Update: resourceAwsLbListenerUpdate,
		Delete: resourceAwsLbListenerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.Sequence(
			SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"alpn_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"HTTP1Only",
					"HTTP2Only",
					"HTTP2Optional",
					"HTTP2Preferred",
					"None",
				}, true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authenticate_cognito": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(elbv2.ActionTypeEnumAuthenticateCognito),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice(
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnum_Values(),
											true,
										),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"user_pool_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"user_pool_client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool_domain": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"authenticate_oidc": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(elbv2.ActionTypeEnumAuthenticateOidc),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice(
											elbv2.AuthenticateOidcActionConditionalBehaviorEnum_Values(),
											true,
										),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"token_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_info_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"fixed_response": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(elbv2.ActionTypeEnumFixedResponse),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"text/plain",
											"text/css",
											"text/html",
											"application/javascript",
											"application/json",
										}, false),
									},
									"message_body": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"status_code": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[245]\d\d$`), ""),
									},
								},
							},
						},
						"forward": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(elbv2.ActionTypeEnumForward),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_group": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 5,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"arn": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateArn,
												},
												"weight": {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(0, 999),
													Default:      1,
													Optional:     true,
												},
											},
										},
									},
									"stickiness": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"duration": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 604800),
												},
												"enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
											},
										},
									},
								},
							},
						},
						"order": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 50000),
						},
						"redirect": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(elbv2.ActionTypeEnumRedirect),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{host}",
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/#{path}",
									},
									"port": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{port}",
									},
									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{protocol}",
										ValidateFunc: validation.StringInSlice([]string{
											"#{protocol}",
											"HTTP",
											"HTTPS",
										}, false),
									},
									"query": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{query}",
									},
									"status_code": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice(
											elbv2.RedirectActionStatusCodeEnum_Values(),
											false,
										),
									},
								},
							},
						},
						"target_group_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfDefaultActionTypeNot(elbv2.ActionTypeEnumForward),
							ValidateFunc:     validateArn,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								elbv2.ActionTypeEnum_Values(),
								true,
							),
						},
					},
				},
			},
			"load_balancer_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
				ValidateFunc: validation.StringInSlice(elbv2.ProtocolEnum_Values(), true),
			},
			"ssl_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func suppressIfDefaultActionTypeNot(t string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		take := 2
		i := strings.IndexFunc(k, func(r rune) bool {
			if r == '.' {
				take -= 1
				return take == 0
			}
			return false
		})
		at := k[:i+1] + "type"
		return d.Get(at).(string) != t
	}
}

func resourceAwsLbListenerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	lbArn := d.Get("load_balancer_arn").(string)

	params := &elbv2.CreateListenerInput{
		LoadBalancerArn: aws.String(lbArn),
	}

	if v, ok := d.GetOk("port"); ok {
		params.Port = aws.Int64(int64(v.(int)))
	}

	if len(tags) > 0 {
		params.Tags = tags.IgnoreAws().Elbv2Tags()
	}

	if v, ok := d.GetOk("protocol"); ok {
		params.Protocol = aws.String(v.(string))
	} else if strings.Contains(lbArn, "loadbalancer/app/") {
		// Keep previous default of HTTP for Application Load Balancers
		params.Protocol = aws.String(elbv2.ProtocolEnumHttp)
	}

	if sslPolicy, ok := d.GetOk("ssl_policy"); ok {
		params.SslPolicy = aws.String(sslPolicy.(string))
	}

	if certificateArn, ok := d.GetOk("certificate_arn"); ok {
		params.Certificates = make([]*elbv2.Certificate, 1)
		params.Certificates[0] = &elbv2.Certificate{
			CertificateArn: aws.String(certificateArn.(string)),
		}
	}

	if alpnPolicy, ok := d.GetOk("alpn_policy"); ok {
		params.AlpnPolicy = make([]*string, 1)
		params.AlpnPolicy[0] = aws.String(alpnPolicy.(string))
	}

	if v, ok := d.GetOk("default_action"); ok && len(v.([]interface{})) > 0 {
		var err error
		params.DefaultActions, err = expandLbListenerActions(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("error creating ELBv2 Listener for ARN (%s): %w", lbArn, err)
		}
	}

	var output *elbv2.CreateListenerOutput

	err := resource.Retry(waiter.LoadBalancerListenerCreateTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateListener(params)

		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeCertificateNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateListener(params)
	}

	if err != nil {
		return fmt.Errorf("error creating ELBv2 Listener (%s): %w", lbArn, err)
	}

	if output == nil || len(output.Listeners) == 0 {
		return fmt.Errorf("error creating ELBv2 Listener: no listeners returned in response")
	}

	d.SetId(aws.StringValue(output.Listeners[0].ListenerArn))

	return resourceAwsLbListenerRead(d, meta)
}

func resourceAwsLbListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var listener *elbv2.Listener

	err := resource.Retry(waiter.LoadBalancerListenerReadTimeout, func() *resource.RetryError {
		var err error
		listener, err = finder.ListenerByARN(conn, d.Id())

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		listener, err = finder.ListenerByARN(conn, d.Id())
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
		log.Printf("[WARN] ELBv2 Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing ELBv2 Listener (%s): %w", d.Id(), err)
	}

	if listener == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error describing ELBv2 Listener (%s): empty response", d.Id())
		}
		log.Printf("[WARN] ELBv2 Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", listener.ListenerArn)
	d.Set("load_balancer_arn", listener.LoadBalancerArn)
	d.Set("port", listener.Port)
	d.Set("protocol", listener.Protocol)
	d.Set("ssl_policy", listener.SslPolicy)

	if listener.Certificates != nil && len(listener.Certificates) == 1 && listener.Certificates[0] != nil {
		d.Set("certificate_arn", listener.Certificates[0].CertificateArn)
	}

	if listener.AlpnPolicy != nil && len(listener.AlpnPolicy) == 1 && listener.AlpnPolicy[0] != nil {
		d.Set("alpn_policy", listener.AlpnPolicy[0])
	}

	sort.Slice(listener.DefaultActions, func(i, j int) bool {
		return aws.Int64Value(listener.DefaultActions[i].Order) < aws.Int64Value(listener.DefaultActions[j].Order)
	})

	if err := d.Set("default_action", flattenLbListenerActions(d, listener.DefaultActions)); err != nil {
		return fmt.Errorf("error setting default_action for ELBv2 listener (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.Elbv2ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for (%s): %w", d.Id(), err)
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

func resourceAwsLbListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	if d.HasChangesExcept("tags", "tags_all") {
		params := &elbv2.ModifyListenerInput{
			ListenerArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("port"); ok {
			params.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("protocol"); ok {
			params.Protocol = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ssl_policy"); ok {
			params.SslPolicy = aws.String(v.(string))
		}

		if v, ok := d.GetOk("certificate_arn"); ok {
			params.Certificates = make([]*elbv2.Certificate, 1)
			params.Certificates[0] = &elbv2.Certificate{
				CertificateArn: aws.String(v.(string)),
			}
		}

		if v, ok := d.GetOk("alpn_policy"); ok {
			params.AlpnPolicy = aws.StringSlice([]string{v.(string)})
		}

		if d.HasChange("default_action") {
			var err error
			params.DefaultActions, err = expandLbListenerActions(d.Get("default_action").([]interface{}))
			if err != nil {
				return fmt.Errorf("error updating ELBv2 Listener (%s): %w", d.Id(), err)
			}
		}

		err := resource.Retry(waiter.LoadBalancerListenerUpdateTimeout, func() *resource.RetryError {
			_, err := conn.ModifyListener(params)

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeCertificateNotFoundException) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.ModifyListener(params)
		}

		if err != nil {
			return fmt.Errorf("error modifying ELBv2 Listener (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := resource.Retry(waiter.LoadBalancerTagPropagationTimeout, func() *resource.RetryError {
			err := keyvaluetags.Elbv2UpdateTags(conn, d.Id(), o, n)

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) ||
				tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
				log.Printf("[DEBUG] Retrying tagging of LB Listener (%s) after error: %s", d.Id(), err)
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			err = keyvaluetags.Elbv2UpdateTags(conn, d.Id(), o, n)
		}

		if err != nil {
			return fmt.Errorf("error updating LB (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsLbListenerRead(d, meta)
}

func resourceAwsLbListenerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	_, err := conn.DeleteListener(&elbv2.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting Listener (%s): %w", d.Id(), err)
	}

	return nil
}

func expandLbListenerActions(l []interface{}) ([]*elbv2.Action, error) {
	if len(l) == 0 {
		return nil, nil
	}

	var actions []*elbv2.Action
	var err error

	for i, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		action := &elbv2.Action{
			Order: aws.Int64(int64(i + 1)),
			Type:  aws.String(tfMap["type"].(string)),
		}

		if order, ok := tfMap["order"].(int); ok && order != 0 {
			action.Order = aws.Int64(int64(order))
		}

		switch tfMap["type"].(string) {
		case elbv2.ActionTypeEnumForward:
			if v, ok := tfMap["target_group_arn"].(string); ok && v != "" {
				action.TargetGroupArn = aws.String(v)
			} else if v, ok := tfMap["forward"].([]interface{}); ok {
				action.ForwardConfig = expandLbListenerActionForwardConfig(v)
			} else {
				err = errors.New("for actions of type 'forward', you must specify a 'forward' block or 'target_group_arn'")
			}

		case elbv2.ActionTypeEnumRedirect:
			if v, ok := tfMap["redirect"].([]interface{}); ok {
				action.RedirectConfig = expandLbListenerRedirectActionConfig(v)
			} else {
				err = errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
			}

		case elbv2.ActionTypeEnumFixedResponse:
			if v, ok := tfMap["fixed_response"].([]interface{}); ok {
				action.FixedResponseConfig = expandLbListenerFixedResponseConfig(v)
			} else {
				err = errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
			}

		case elbv2.ActionTypeEnumAuthenticateCognito:
			if v, ok := tfMap["authenticate_cognito"].([]interface{}); ok {
				action.AuthenticateCognitoConfig = expandLbListenerAuthenticateCognitoConfig(v)
			} else {
				err = errors.New("for actions of type 'authenticate-cognito', you must specify a 'authenticate_cognito' block")
			}

		case elbv2.ActionTypeEnumAuthenticateOidc:
			if v, ok := tfMap["authenticate_oidc"].([]interface{}); ok {
				action.AuthenticateOidcConfig = expandLbListenerAuthenticateOidcConfig(v)
			} else {
				err = errors.New("for actions of type 'authenticate-oidc', you must specify a 'authenticate_oidc' block")
			}
		}

		actions = append(actions, action)
	}

	return actions, err
}

func expandLbListenerAuthenticateCognitoConfig(l []interface{}) *elbv2.AuthenticateCognitoActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &elbv2.AuthenticateCognitoActionConfig{
		AuthenticationRequestExtraParams: expandStringMap(tfMap["authentication_request_extra_params"].(map[string]interface{})),
		UserPoolArn:                      aws.String(tfMap["user_pool_arn"].(string)),
		UserPoolClientId:                 aws.String(tfMap["user_pool_client_id"].(string)),
		UserPoolDomain:                   aws.String(tfMap["user_pool_domain"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = aws.String(v)
	}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}

	if v, ok := tfMap["session_cookie_name"].(string); ok && v != "" {
		config.SessionCookieName = aws.String(v)
	}

	if v, ok := tfMap["session_timeout"].(int); ok && v != 0 {
		config.SessionTimeout = aws.Int64(int64(v))
	}

	return config
}

func expandLbListenerAuthenticateOidcConfig(l []interface{}) *elbv2.AuthenticateOidcActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &elbv2.AuthenticateOidcActionConfig{
		AuthenticationRequestExtraParams: expandStringMap(tfMap["authentication_request_extra_params"].(map[string]interface{})),
		AuthorizationEndpoint:            aws.String(tfMap["authorization_endpoint"].(string)),
		ClientId:                         aws.String(tfMap["client_id"].(string)),
		ClientSecret:                     aws.String(tfMap["client_secret"].(string)),
		Issuer:                           aws.String(tfMap["issuer"].(string)),
		TokenEndpoint:                    aws.String(tfMap["token_endpoint"].(string)),
		UserInfoEndpoint:                 aws.String(tfMap["user_info_endpoint"].(string)),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = aws.String(v)
	}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}

	if v, ok := tfMap["session_cookie_name"].(string); ok && v != "" {
		config.SessionCookieName = aws.String(v)
	}

	if v, ok := tfMap["session_timeout"].(int); ok && v != 0 {
		config.SessionTimeout = aws.Int64(int64(v))
	}

	return config
}

func expandLbListenerFixedResponseConfig(l []interface{}) *elbv2.FixedResponseActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &elbv2.FixedResponseActionConfig{
		ContentType: aws.String(tfMap["content_type"].(string)),
		MessageBody: aws.String(tfMap["message_body"].(string)),
		StatusCode:  aws.String(tfMap["status_code"].(string)),
	}
}

func expandLbListenerRedirectActionConfig(l []interface{}) *elbv2.RedirectActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &elbv2.RedirectActionConfig{
		Host:       aws.String(tfMap["host"].(string)),
		Path:       aws.String(tfMap["path"].(string)),
		Port:       aws.String(tfMap["port"].(string)),
		Protocol:   aws.String(tfMap["protocol"].(string)),
		Query:      aws.String(tfMap["query"].(string)),
		StatusCode: aws.String(tfMap["status_code"].(string)),
	}
}

func expandLbListenerActionForwardConfig(l []interface{}) *elbv2.ForwardActionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &elbv2.ForwardActionConfig{}

	if v, ok := tfMap["target_group"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetGroups = expandLbListenerActionForwardConfigTargetGroups(v.List())
	}

	if v, ok := tfMap["stickiness"].([]interface{}); ok && len(v) > 0 {
		config.TargetGroupStickinessConfig = expandLbListenerActionForwardConfigTargetGroupStickinessConfig(v)
	}

	return config
}

func expandLbListenerActionForwardConfigTargetGroups(l []interface{}) []*elbv2.TargetGroupTuple {
	if len(l) == 0 {
		return nil
	}

	var groups []*elbv2.TargetGroupTuple

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue

		}

		group := &elbv2.TargetGroupTuple{
			TargetGroupArn: aws.String(tfMap["arn"].(string)),
			Weight:         aws.Int64(int64(tfMap["weight"].(int))),
		}

		groups = append(groups, group)
	}

	return groups
}

func expandLbListenerActionForwardConfigTargetGroupStickinessConfig(l []interface{}) *elbv2.TargetGroupStickinessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &elbv2.TargetGroupStickinessConfig{
		Enabled:         aws.Bool(tfMap["enabled"].(bool)),
		DurationSeconds: aws.Int64(int64(tfMap["duration"].(int))),
	}
}

func flattenLbListenerActions(d *schema.ResourceData, Actions []*elbv2.Action) []interface{} {
	if len(Actions) == 0 {
		return []interface{}{}
	}

	var vActions []interface{}

	for i, action := range Actions {
		m := map[string]interface{}{
			"type":  aws.StringValue(action.Type),
			"order": aws.Int64Value(action.Order),
		}

		switch aws.StringValue(action.Type) {
		case elbv2.ActionTypeEnumForward:
			if aws.StringValue(action.TargetGroupArn) != "" {
				m["target_group_arn"] = aws.StringValue(action.TargetGroupArn)
			} else {
				m["forward"] = flattenLbListenerActionForwardConfig(action.ForwardConfig)
			}

		case elbv2.ActionTypeEnumRedirect:
			m["redirect"] = flattenLbListenerActionRedirectConfig(action.RedirectConfig)

		case elbv2.ActionTypeEnumFixedResponse:
			m["fixed_response"] = flattenLbListenerActionFixedResponseConfig(action.FixedResponseConfig)

		case elbv2.ActionTypeEnumAuthenticateCognito:
			m["authenticate_cognito"] = flattenLbListenerActionAuthenticateCognitoConfig(action.AuthenticateCognitoConfig)

		case elbv2.ActionTypeEnumAuthenticateOidc:
			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			var clientSecret string
			if v, ok := d.GetOk("default_action." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret"); ok {
				clientSecret = v.(string)
			}

			m["authenticate_oidc"] = flattenLbListenerActionAuthenticateOidcConfig(action.AuthenticateOidcConfig, clientSecret)
		}

		vActions = append(vActions, m)
	}

	return vActions
}

func flattenLbListenerActionAuthenticateOidcConfig(config *elbv2.AuthenticateOidcActionConfig, clientSecret string) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": aws.StringValueMap(config.AuthenticationRequestExtraParams),
		"authorization_endpoint":              aws.StringValue(config.AuthorizationEndpoint),
		"client_id":                           aws.StringValue(config.ClientId),
		"client_secret":                       clientSecret,
		"issuer":                              aws.StringValue(config.Issuer),
		"on_unauthenticated_request":          aws.StringValue(config.OnUnauthenticatedRequest),
		"scope":                               aws.StringValue(config.Scope),
		"session_cookie_name":                 aws.StringValue(config.SessionCookieName),
		"session_timeout":                     aws.Int64Value(config.SessionTimeout),
		"token_endpoint":                      aws.StringValue(config.TokenEndpoint),
		"user_info_endpoint":                  aws.StringValue(config.UserInfoEndpoint),
	}

	return []interface{}{m}
}

func flattenLbListenerActionAuthenticateCognitoConfig(config *elbv2.AuthenticateCognitoActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": aws.StringValueMap(config.AuthenticationRequestExtraParams),
		"on_unauthenticated_request":          aws.StringValue(config.OnUnauthenticatedRequest),
		"scope":                               aws.StringValue(config.Scope),
		"session_cookie_name":                 aws.StringValue(config.SessionCookieName),
		"session_timeout":                     aws.Int64Value(config.SessionTimeout),
		"user_pool_arn":                       aws.StringValue(config.UserPoolArn),
		"user_pool_client_id":                 aws.StringValue(config.UserPoolClientId),
		"user_pool_domain":                    aws.StringValue(config.UserPoolDomain),
	}

	return []interface{}{m}
}

func flattenLbListenerActionFixedResponseConfig(config *elbv2.FixedResponseActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"content_type": aws.StringValue(config.ContentType),
		"message_body": aws.StringValue(config.MessageBody),
		"status_code":  aws.StringValue(config.StatusCode),
	}

	return []interface{}{m}
}

func flattenLbListenerActionForwardConfig(config *elbv2.ForwardActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"target_group": flattenLbListenerActionForwardConfigTargetGroups(config.TargetGroups),
		"stickiness":   flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config.TargetGroupStickinessConfig),
	}

	return []interface{}{m}
}

func flattenLbListenerActionForwardConfigTargetGroups(groups []*elbv2.TargetGroupTuple) []interface{} {
	if len(groups) == 0 {
		return []interface{}{}
	}

	var vGroups []interface{}

	for _, group := range groups {
		if group == nil {
			continue
		}

		m := map[string]interface{}{
			"arn":    aws.StringValue(group.TargetGroupArn),
			"weight": aws.Int64Value(group.Weight),
		}

		vGroups = append(vGroups, m)
	}

	return vGroups
}

func flattenLbListenerActionForwardConfigTargetGroupStickinessConfig(config *elbv2.TargetGroupStickinessConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":  aws.BoolValue(config.Enabled),
		"duration": aws.Int64Value(config.DurationSeconds),
	}

	return []interface{}{m}
}

func flattenLbListenerActionRedirectConfig(config *elbv2.RedirectActionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"host":        aws.StringValue(config.Host),
		"path":        aws.StringValue(config.Path),
		"port":        aws.StringValue(config.Port),
		"protocol":    aws.StringValue(config.Protocol),
		"query":       aws.StringValue(config.Query),
		"status_code": aws.StringValue(config.StatusCode),
	}

	return []interface{}{m}
}
