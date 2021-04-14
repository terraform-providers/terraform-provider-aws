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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumDeny,
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumAllow,
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumAuthenticate,
										}, true),
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
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumDeny,
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumAllow,
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumAuthenticate,
										}, true),
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
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.RedirectActionStatusCodeEnumHttp301,
											elbv2.RedirectActionStatusCodeEnumHttp302,
										}, false),
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
							ValidateFunc: validation.StringInSlice([]string{
								elbv2.ActionTypeEnumAuthenticateCognito,
								elbv2.ActionTypeEnumAuthenticateOidc,
								elbv2.ActionTypeEnumFixedResponse,
								elbv2.ActionTypeEnumForward,
								elbv2.ActionTypeEnumRedirect,
							}, true),
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

	lbArn := d.Get("load_balancer_arn").(string)

	params := &elbv2.CreateListenerInput{
		LoadBalancerArn: aws.String(lbArn),
	}

	if v, ok := d.GetOk("port"); ok {
		params.Port = aws.Int64(int64(v.(int)))
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

	defaultActions := d.Get("default_action").([]interface{})
	params.DefaultActions = make([]*elbv2.Action, len(defaultActions))
	for i, defaultAction := range defaultActions {
		defaultActionMap := defaultAction.(map[string]interface{})

		action := &elbv2.Action{
			Order: aws.Int64(int64(i + 1)),
			Type:  aws.String(defaultActionMap["type"].(string)),
		}

		if order, ok := defaultActionMap["order"]; ok && order.(int) != 0 {
			action.Order = aws.Int64(int64(order.(int)))
		}

		switch defaultActionMap["type"].(string) {
		case elbv2.ActionTypeEnumForward:
			if err := lbListenerRuleActionForward(defaultActionMap, action); err != nil {
				return err
			}

		case elbv2.ActionTypeEnumRedirect:
			redirectList := defaultActionMap["redirect"].([]interface{})

			if len(redirectList) == 1 {
				redirectMap := redirectList[0].(map[string]interface{})

				action.RedirectConfig = &elbv2.RedirectActionConfig{
					Host:       aws.String(redirectMap["host"].(string)),
					Path:       aws.String(redirectMap["path"].(string)),
					Port:       aws.String(redirectMap["port"].(string)),
					Protocol:   aws.String(redirectMap["protocol"].(string)),
					Query:      aws.String(redirectMap["query"].(string)),
					StatusCode: aws.String(redirectMap["status_code"].(string)),
				}
			} else {
				return errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
			}

		case elbv2.ActionTypeEnumFixedResponse:
			fixedResponseList := defaultActionMap["fixed_response"].([]interface{})

			if len(fixedResponseList) == 1 {
				fixedResponseMap := fixedResponseList[0].(map[string]interface{})

				action.FixedResponseConfig = &elbv2.FixedResponseActionConfig{
					ContentType: aws.String(fixedResponseMap["content_type"].(string)),
					MessageBody: aws.String(fixedResponseMap["message_body"].(string)),
					StatusCode:  aws.String(fixedResponseMap["status_code"].(string)),
				}
			} else {
				return errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
			}

		case elbv2.ActionTypeEnumAuthenticateCognito:
			authenticateCognitoList := defaultActionMap["authenticate_cognito"].([]interface{})

			if len(authenticateCognitoList) == 1 {
				authenticateCognitoMap := authenticateCognitoList[0].(map[string]interface{})

				authenticationRequestExtraParams := make(map[string]*string)
				for key, value := range authenticateCognitoMap["authentication_request_extra_params"].(map[string]interface{}) {
					authenticationRequestExtraParams[key] = aws.String(value.(string))
				}

				action.AuthenticateCognitoConfig = &elbv2.AuthenticateCognitoActionConfig{
					AuthenticationRequestExtraParams: authenticationRequestExtraParams,
					UserPoolArn:                      aws.String(authenticateCognitoMap["user_pool_arn"].(string)),
					UserPoolClientId:                 aws.String(authenticateCognitoMap["user_pool_client_id"].(string)),
					UserPoolDomain:                   aws.String(authenticateCognitoMap["user_pool_domain"].(string)),
				}

				if onUnauthenticatedRequest, ok := authenticateCognitoMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
					action.AuthenticateCognitoConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
				}
				if scope, ok := authenticateCognitoMap["scope"]; ok && scope != "" {
					action.AuthenticateCognitoConfig.Scope = aws.String(scope.(string))
				}
				if sessionCookieName, ok := authenticateCognitoMap["session_cookie_name"]; ok && sessionCookieName != "" {
					action.AuthenticateCognitoConfig.SessionCookieName = aws.String(sessionCookieName.(string))
				}
				if sessionTimeout, ok := authenticateCognitoMap["session_timeout"]; ok && sessionTimeout != 0 {
					action.AuthenticateCognitoConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
				}
			} else {
				return errors.New("for actions of type 'authenticate-cognito', you must specify a 'authenticate_cognito' block")
			}

		case elbv2.ActionTypeEnumAuthenticateOidc:
			authenticateOidcList := defaultActionMap["authenticate_oidc"].([]interface{})

			if len(authenticateOidcList) == 1 {
				authenticateOidcMap := authenticateOidcList[0].(map[string]interface{})

				authenticationRequestExtraParams := make(map[string]*string)
				for key, value := range authenticateOidcMap["authentication_request_extra_params"].(map[string]interface{}) {
					authenticationRequestExtraParams[key] = aws.String(value.(string))
				}

				action.AuthenticateOidcConfig = &elbv2.AuthenticateOidcActionConfig{
					AuthenticationRequestExtraParams: authenticationRequestExtraParams,
					AuthorizationEndpoint:            aws.String(authenticateOidcMap["authorization_endpoint"].(string)),
					ClientId:                         aws.String(authenticateOidcMap["client_id"].(string)),
					ClientSecret:                     aws.String(authenticateOidcMap["client_secret"].(string)),
					Issuer:                           aws.String(authenticateOidcMap["issuer"].(string)),
					TokenEndpoint:                    aws.String(authenticateOidcMap["token_endpoint"].(string)),
					UserInfoEndpoint:                 aws.String(authenticateOidcMap["user_info_endpoint"].(string)),
				}

				if onUnauthenticatedRequest, ok := authenticateOidcMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
					action.AuthenticateOidcConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
				}
				if scope, ok := authenticateOidcMap["scope"]; ok && scope != "" {
					action.AuthenticateOidcConfig.Scope = aws.String(scope.(string))
				}
				if sessionCookieName, ok := authenticateOidcMap["session_cookie_name"]; ok && sessionCookieName != "" {
					action.AuthenticateOidcConfig.SessionCookieName = aws.String(sessionCookieName.(string))
				}
				if sessionTimeout, ok := authenticateOidcMap["session_timeout"]; ok && sessionTimeout != 0 {
					action.AuthenticateOidcConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
				}
			} else {
				return errors.New("for actions of type 'authenticate-oidc', you must specify a 'authenticate_oidc' block")
			}
		}

		params.DefaultActions[i] = action
	}

	var output *elbv2.CreateListenerOutput

	err := resource.Retry(waiter.LoadBalancerListenerCreateTimeout, func() *resource.RetryError {
		var err error

		log.Printf("[DEBUG] Creating LB listener for ARN: %s", d.Get("load_balancer_arn").(string))
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
		return fmt.Errorf("error creating ELBv2 Listener (%s): %w", d.Get("load_balancer_arn").(string), err)
	}

	if output == nil || len(output.Listeners) == 0 {
		return fmt.Errorf("error creating ELBv2 Listener: no listeners returned in response")
	}

	d.SetId(aws.StringValue(output.Listeners[0].ListenerArn))

	return resourceAwsLbListenerRead(d, meta)
}

func resourceAwsLbListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	var output *elbv2.DescribeListenersOutput
	var input = &elbv2.DescribeListenersInput{
		ListenerArns: []*string{aws.String(d.Id())},
	}

	err := resource.Retry(waiter.LoadBalancerListenerReadTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.DescribeListeners(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.DescribeListeners(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elbv2.ErrCodeListenerNotFoundException) {
		log.Printf("[WARN] ELBv2 Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing ELBv2 Listener (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error describing ELBv2 Listener (%s): empty response", d.Id())
	}

	var listener *elbv2.Listener

	for _, l := range output.Listeners {
		if aws.StringValue(l.ListenerArn) == d.Id() {
			listener = l
			break
		}
	}

	if listener == nil {
		return fmt.Errorf("error describing ELBv2 Listener (%s): not found in response", d.Id())
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
	defaultActions := make([]interface{}, len(listener.DefaultActions))
	for i, defaultAction := range listener.DefaultActions {
		defaultActionMap := make(map[string]interface{})
		defaultActionMap["type"] = aws.StringValue(defaultAction.Type)
		defaultActionMap["order"] = aws.Int64Value(defaultAction.Order)

		switch aws.StringValue(defaultAction.Type) {
		case elbv2.ActionTypeEnumForward:
			if aws.StringValue(defaultAction.TargetGroupArn) != "" {
				defaultActionMap["target_group_arn"] = aws.StringValue(defaultAction.TargetGroupArn)
			} else {
				targetGroups := make([]map[string]interface{}, 0, len(defaultAction.ForwardConfig.TargetGroups))
				for _, targetGroup := range defaultAction.ForwardConfig.TargetGroups {
					targetGroups = append(targetGroups,
						map[string]interface{}{
							"arn":    aws.StringValue(targetGroup.TargetGroupArn),
							"weight": aws.Int64Value(targetGroup.Weight),
						},
					)
				}
				defaultActionMap["forward"] = []map[string]interface{}{
					{
						"target_group": targetGroups,
						"stickiness": []map[string]interface{}{
							{
								"enabled":  aws.BoolValue(defaultAction.ForwardConfig.TargetGroupStickinessConfig.Enabled),
								"duration": aws.Int64Value(defaultAction.ForwardConfig.TargetGroupStickinessConfig.DurationSeconds),
							},
						},
					},
				}
			}

		case elbv2.ActionTypeEnumRedirect:
			defaultActionMap["redirect"] = []map[string]interface{}{
				{
					"host":        aws.StringValue(defaultAction.RedirectConfig.Host),
					"path":        aws.StringValue(defaultAction.RedirectConfig.Path),
					"port":        aws.StringValue(defaultAction.RedirectConfig.Port),
					"protocol":    aws.StringValue(defaultAction.RedirectConfig.Protocol),
					"query":       aws.StringValue(defaultAction.RedirectConfig.Query),
					"status_code": aws.StringValue(defaultAction.RedirectConfig.StatusCode),
				},
			}

		case elbv2.ActionTypeEnumFixedResponse:
			defaultActionMap["fixed_response"] = []map[string]interface{}{
				{
					"content_type": aws.StringValue(defaultAction.FixedResponseConfig.ContentType),
					"message_body": aws.StringValue(defaultAction.FixedResponseConfig.MessageBody),
					"status_code":  aws.StringValue(defaultAction.FixedResponseConfig.StatusCode),
				},
			}

		case elbv2.ActionTypeEnumAuthenticateCognito:
			authenticationRequestExtraParams := make(map[string]interface{})
			for key, value := range defaultAction.AuthenticateCognitoConfig.AuthenticationRequestExtraParams {
				authenticationRequestExtraParams[key] = aws.StringValue(value)
			}
			defaultActionMap["authenticate_cognito"] = []map[string]interface{}{
				{
					"authentication_request_extra_params": authenticationRequestExtraParams,
					"on_unauthenticated_request":          aws.StringValue(defaultAction.AuthenticateCognitoConfig.OnUnauthenticatedRequest),
					"scope":                               aws.StringValue(defaultAction.AuthenticateCognitoConfig.Scope),
					"session_cookie_name":                 aws.StringValue(defaultAction.AuthenticateCognitoConfig.SessionCookieName),
					"session_timeout":                     aws.Int64Value(defaultAction.AuthenticateCognitoConfig.SessionTimeout),
					"user_pool_arn":                       aws.StringValue(defaultAction.AuthenticateCognitoConfig.UserPoolArn),
					"user_pool_client_id":                 aws.StringValue(defaultAction.AuthenticateCognitoConfig.UserPoolClientId),
					"user_pool_domain":                    aws.StringValue(defaultAction.AuthenticateCognitoConfig.UserPoolDomain),
				},
			}

		case elbv2.ActionTypeEnumAuthenticateOidc:
			authenticationRequestExtraParams := make(map[string]interface{})
			for key, value := range defaultAction.AuthenticateOidcConfig.AuthenticationRequestExtraParams {
				authenticationRequestExtraParams[key] = aws.StringValue(value)
			}

			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			clientSecret := d.Get("default_action." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret").(string)

			defaultActionMap["authenticate_oidc"] = []map[string]interface{}{
				{
					"authentication_request_extra_params": authenticationRequestExtraParams,
					"authorization_endpoint":              aws.StringValue(defaultAction.AuthenticateOidcConfig.AuthorizationEndpoint),
					"client_id":                           aws.StringValue(defaultAction.AuthenticateOidcConfig.ClientId),
					"client_secret":                       clientSecret,
					"issuer":                              aws.StringValue(defaultAction.AuthenticateOidcConfig.Issuer),
					"on_unauthenticated_request":          aws.StringValue(defaultAction.AuthenticateOidcConfig.OnUnauthenticatedRequest),
					"scope":                               aws.StringValue(defaultAction.AuthenticateOidcConfig.Scope),
					"session_cookie_name":                 aws.StringValue(defaultAction.AuthenticateOidcConfig.SessionCookieName),
					"session_timeout":                     aws.Int64Value(defaultAction.AuthenticateOidcConfig.SessionTimeout),
					"token_endpoint":                      aws.StringValue(defaultAction.AuthenticateOidcConfig.TokenEndpoint),
					"user_info_endpoint":                  aws.StringValue(defaultAction.AuthenticateOidcConfig.UserInfoEndpoint),
				},
			}
		}

		defaultActions[i] = defaultActionMap
	}
	if err := d.Set("default_action", defaultActions); err != nil {
		return fmt.Errorf("error setting default_action for ELBv2 listener (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsLbListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	params := &elbv2.ModifyListenerInput{
		ListenerArn: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("port"); ok {
		params.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("protocol"); ok {
		params.Protocol = aws.String(v.(string))
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

	if d.HasChange("default_action") {
		defaultActions := d.Get("default_action").([]interface{})
		params.DefaultActions = make([]*elbv2.Action, len(defaultActions))

		for i, defaultAction := range defaultActions {
			defaultActionMap := defaultAction.(map[string]interface{})

			action := &elbv2.Action{
				Order: aws.Int64(int64(i + 1)),
				Type:  aws.String(defaultActionMap["type"].(string)),
			}

			if order, ok := defaultActionMap["order"]; ok && order.(int) != 0 {
				action.Order = aws.Int64(int64(order.(int)))
			}

			switch defaultActionMap["type"].(string) {
			case elbv2.ActionTypeEnumForward:
				if err := lbListenerRuleActionForward(defaultActionMap, action); err != nil {
					return err
				}

			case elbv2.ActionTypeEnumRedirect:
				redirectList := defaultActionMap["redirect"].([]interface{})

				if len(redirectList) == 1 {
					redirectMap := redirectList[0].(map[string]interface{})

					action.RedirectConfig = &elbv2.RedirectActionConfig{
						Host:       aws.String(redirectMap["host"].(string)),
						Path:       aws.String(redirectMap["path"].(string)),
						Port:       aws.String(redirectMap["port"].(string)),
						Protocol:   aws.String(redirectMap["protocol"].(string)),
						Query:      aws.String(redirectMap["query"].(string)),
						StatusCode: aws.String(redirectMap["status_code"].(string)),
					}
				} else {
					return errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
				}

			case elbv2.ActionTypeEnumFixedResponse:
				fixedResponseList := defaultActionMap["fixed_response"].([]interface{})

				if len(fixedResponseList) == 1 {
					fixedResponseMap := fixedResponseList[0].(map[string]interface{})

					action.FixedResponseConfig = &elbv2.FixedResponseActionConfig{
						ContentType: aws.String(fixedResponseMap["content_type"].(string)),
						MessageBody: aws.String(fixedResponseMap["message_body"].(string)),
						StatusCode:  aws.String(fixedResponseMap["status_code"].(string)),
					}
				} else {
					return errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
				}

			case elbv2.ActionTypeEnumAuthenticateCognito:
				authenticateCognitoList := defaultActionMap["authenticate_cognito"].([]interface{})

				if len(authenticateCognitoList) == 1 {
					authenticateCognitoMap := authenticateCognitoList[0].(map[string]interface{})

					authenticationRequestExtraParams := make(map[string]*string)
					for key, value := range authenticateCognitoMap["authentication_request_extra_params"].(map[string]interface{}) {
						authenticationRequestExtraParams[key] = aws.String(value.(string))
					}

					action.AuthenticateCognitoConfig = &elbv2.AuthenticateCognitoActionConfig{
						AuthenticationRequestExtraParams: authenticationRequestExtraParams,
						UserPoolArn:                      aws.String(authenticateCognitoMap["user_pool_arn"].(string)),
						UserPoolClientId:                 aws.String(authenticateCognitoMap["user_pool_client_id"].(string)),
						UserPoolDomain:                   aws.String(authenticateCognitoMap["user_pool_domain"].(string)),
					}

					if onUnauthenticatedRequest, ok := authenticateCognitoMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
						action.AuthenticateCognitoConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
					}
					if scope, ok := authenticateCognitoMap["scope"]; ok && scope != "" {
						action.AuthenticateCognitoConfig.Scope = aws.String(scope.(string))
					}
					if sessionCookieName, ok := authenticateCognitoMap["session_cookie_name"]; ok && sessionCookieName != "" {
						action.AuthenticateCognitoConfig.SessionCookieName = aws.String(sessionCookieName.(string))
					}
					if sessionTimeout, ok := authenticateCognitoMap["session_timeout"]; ok && sessionTimeout != 0 {
						action.AuthenticateCognitoConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
					}
				} else {
					return errors.New("for actions of type 'authenticate-cognito', you must specify a 'authenticate_cognito' block")
				}

			case elbv2.ActionTypeEnumAuthenticateOidc:
				authenticateOidcList := defaultActionMap["authenticate_oidc"].([]interface{})

				if len(authenticateOidcList) == 1 {
					authenticateOidcMap := authenticateOidcList[0].(map[string]interface{})

					authenticationRequestExtraParams := make(map[string]*string)
					for key, value := range authenticateOidcMap["authentication_request_extra_params"].(map[string]interface{}) {
						authenticationRequestExtraParams[key] = aws.String(value.(string))
					}

					action.AuthenticateOidcConfig = &elbv2.AuthenticateOidcActionConfig{
						AuthenticationRequestExtraParams: authenticationRequestExtraParams,
						AuthorizationEndpoint:            aws.String(authenticateOidcMap["authorization_endpoint"].(string)),
						ClientId:                         aws.String(authenticateOidcMap["client_id"].(string)),
						ClientSecret:                     aws.String(authenticateOidcMap["client_secret"].(string)),
						Issuer:                           aws.String(authenticateOidcMap["issuer"].(string)),
						TokenEndpoint:                    aws.String(authenticateOidcMap["token_endpoint"].(string)),
						UserInfoEndpoint:                 aws.String(authenticateOidcMap["user_info_endpoint"].(string)),
					}

					if onUnauthenticatedRequest, ok := authenticateOidcMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
						action.AuthenticateOidcConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
					}
					if scope, ok := authenticateOidcMap["scope"]; ok && scope != "" {
						action.AuthenticateOidcConfig.Scope = aws.String(scope.(string))
					}
					if sessionCookieName, ok := authenticateOidcMap["session_cookie_name"]; ok && sessionCookieName != "" {
						action.AuthenticateOidcConfig.SessionCookieName = aws.String(sessionCookieName.(string))
					}
					if sessionTimeout, ok := authenticateOidcMap["session_timeout"]; ok && sessionTimeout != 0 {
						action.AuthenticateOidcConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
					}
				} else {
					return errors.New("for actions of type 'authenticate-oidc', you must specify a 'authenticate_oidc' block")
				}
			}

			params.DefaultActions[i] = action
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
