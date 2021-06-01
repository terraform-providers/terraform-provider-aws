package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appmesh/finder"
)

func resourceAwsAppmeshGatewayRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshGatewayRouteCreate,
		Read:   resourceAwsAppmeshGatewayRouteRead,
		Update: resourceAwsAppmeshGatewayRouteUpdate,
		Delete: resourceAwsAppmeshGatewayRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshGatewayRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"virtual_gateway_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grpc_route": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_service": {
																Type:     schema.TypeList,
																Required: true,
																MinItems: 1,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"virtual_service_name": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"service_name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
							ExactlyOneOf: []string{
								"spec.0.grpc_route",
								"spec.0.http2_route",
								"spec.0.http_route",
							},
						},

						"http2_route": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_service": {
																Type:     schema.TypeList,
																Required: true,
																MinItems: 1,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"virtual_service_name": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prefix": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
												},
											},
										},
									},
								},
							},
							ExactlyOneOf: []string{
								"spec.0.grpc_route",
								"spec.0.http2_route",
								"spec.0.http_route",
							},
						},

						"http_route": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_service": {
																Type:     schema.TypeList,
																Required: true,
																MinItems: 1,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"virtual_service_name": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prefix": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
												},
											},
										},
									},
								},
							},
							ExactlyOneOf: []string{
								"spec.0.grpc_route",
								"spec.0.http2_route",
								"spec.0.http_route",
							},
						},
					},
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAppmeshGatewayRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	input := &appmesh.CreateGatewayRouteInput{
		GatewayRouteName:   aws.String(d.Get("name").(string)),
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandAppmeshGatewayRouteSpec(d.Get("spec").([]interface{})),
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh gateway route: %s", input)
	output, err := conn.CreateGatewayRoute(input)

	if err != nil {
		return fmt.Errorf("error creating App Mesh gateway route: %w", err)
	}

	d.SetId(aws.StringValue(output.GatewayRoute.Metadata.Uid))

	return resourceAwsAppmeshGatewayRouteRead(d, meta)
}

func resourceAwsAppmeshGatewayRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	gatewayRoute, err := finder.GatewayRoute(conn, d.Get("mesh_name").(string), d.Get("virtual_gateway_name").(string), d.Get("name").(string), d.Get("mesh_owner").(string))

	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh gateway route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading App Mesh gateway route: %w", err)
	}

	if gatewayRoute == nil || aws.StringValue(gatewayRoute.Status.Status) == appmesh.GatewayRouteStatusCodeDeleted {
		log.Printf("[WARN] App Mesh gateway route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(gatewayRoute.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", gatewayRoute.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", gatewayRoute.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", gatewayRoute.MeshName)
	d.Set("mesh_owner", gatewayRoute.Metadata.MeshOwner)
	d.Set("name", gatewayRoute.GatewayRouteName)
	d.Set("resource_owner", gatewayRoute.Metadata.ResourceOwner)
	err = d.Set("spec", flattenAppmeshGatewayRouteSpec(gatewayRoute.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %w", err)
	}
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh gateway route (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppmeshGatewayRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		input := &appmesh.UpdateGatewayRouteInput{
			GatewayRouteName:   aws.String(d.Get("name").(string)),
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			Spec:               expandAppmeshGatewayRouteSpec(d.Get("spec").([]interface{})),
			VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh gateway route: %s", input)
		_, err := conn.UpdateGatewayRoute(input)

		if err != nil {
			return fmt.Errorf("error updating App Mesh gateway route (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh gateway route (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsAppmeshGatewayRouteRead(d, meta)
}

func resourceAwsAppmeshGatewayRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh gateway route (%s)", d.Id())
	_, err := conn.DeleteGatewayRoute(&appmesh.DeleteGatewayRouteInput{
		GatewayRouteName:   aws.String(d.Get("name").(string)),
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	})

	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting App Mesh gateway route (%s) : %w", d.Id(), err)
	}

	return nil
}

func resourceAwsAppmeshGatewayRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-gateway-name/gateway-route-name'", d.Id())
	}

	mesh := parts[0]
	vgName := parts[1]
	name := parts[2]
	log.Printf("[DEBUG] Importing App Mesh gateway route %s from mesh %s/virtual gateway %s ", name, mesh, vgName)

	conn := meta.(*AWSClient).appmeshconn

	gatewayRoute, err := finder.GatewayRoute(conn, mesh, vgName, name, "")

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(gatewayRoute.Metadata.Uid))
	d.Set("mesh_name", gatewayRoute.MeshName)
	d.Set("name", gatewayRoute.GatewayRouteName)
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	return []*schema.ResourceData{d}, nil
}

func expandAppmeshGatewayRouteSpec(vSpec []interface{}) *appmesh.GatewayRouteSpec {
	if len(vSpec) == 0 || vSpec[0] == nil {
		return nil
	}

	spec := &appmesh.GatewayRouteSpec{}

	mSpec := vSpec[0].(map[string]interface{})

	if vGrpcRoute, ok := mSpec["grpc_route"].([]interface{}); ok {
		spec.GrpcRoute = expandAppmeshGrpcGatewayRoute(vGrpcRoute)
	}

	if vHttp2Route, ok := mSpec["http2_route"].([]interface{}); ok {
		spec.Http2Route = expandAppmeshHttpGatewayRoute(vHttp2Route)
	}

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok {
		spec.HttpRoute = expandAppmeshHttpGatewayRoute(vHttpRoute)
	}

	return spec
}

func expandAppmeshGatewayRouteTarget(vRouteTarget []interface{}) *appmesh.GatewayRouteTarget {
	if len(vRouteTarget) == 0 || vRouteTarget[0] == nil {
		return nil
	}

	routeTarget := &appmesh.GatewayRouteTarget{}

	mRouteTarget := vRouteTarget[0].(map[string]interface{})

	if vVirtualService, ok := mRouteTarget["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
		virtualService := &appmesh.GatewayRouteVirtualService{}

		mVirtualService := vVirtualService[0].(map[string]interface{})

		if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok && vVirtualServiceName != "" {
			virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
		}

		routeTarget.VirtualService = virtualService
	}

	return routeTarget
}

func expandAppmeshGrpcGatewayRoute(vGrpcRoute []interface{}) *appmesh.GrpcGatewayRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	route := &appmesh.GrpcGatewayRoute{}

	mGrpcRoute := vGrpcRoute[0].(map[string]interface{})

	if vRouteAction, ok := mGrpcRoute["action"].([]interface{}); ok && len(vRouteAction) > 0 && vRouteAction[0] != nil {
		routeAction := &appmesh.GrpcGatewayRouteAction{}

		mRouteAction := vRouteAction[0].(map[string]interface{})

		if vRouteTarget, ok := mRouteAction["target"].([]interface{}); ok {
			routeAction.Target = expandAppmeshGatewayRouteTarget(vRouteTarget)
		}

		route.Action = routeAction
	}

	if vRouteMatch, ok := mGrpcRoute["match"].([]interface{}); ok && len(vRouteMatch) > 0 && vRouteMatch[0] != nil {
		routeMatch := &appmesh.GrpcGatewayRouteMatch{}

		mRouteMatch := vRouteMatch[0].(map[string]interface{})

		if vServiceName, ok := mRouteMatch["service_name"].(string); ok && vServiceName != "" {
			routeMatch.ServiceName = aws.String(vServiceName)
		}

		route.Match = routeMatch
	}

	return route
}

func expandAppmeshHttpGatewayRoute(vHttpRoute []interface{}) *appmesh.HttpGatewayRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	route := &appmesh.HttpGatewayRoute{}

	mHttpRoute := vHttpRoute[0].(map[string]interface{})

	if vRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vRouteAction) > 0 && vRouteAction[0] != nil {
		routeAction := &appmesh.HttpGatewayRouteAction{}

		mRouteAction := vRouteAction[0].(map[string]interface{})

		if vRouteTarget, ok := mRouteAction["target"].([]interface{}); ok {
			routeAction.Target = expandAppmeshGatewayRouteTarget(vRouteTarget)
		}

		route.Action = routeAction
	}

	if vRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vRouteMatch) > 0 && vRouteMatch[0] != nil {
		routeMatch := &appmesh.HttpGatewayRouteMatch{}

		mRouteMatch := vRouteMatch[0].(map[string]interface{})

		if vPrefix, ok := mRouteMatch["prefix"].(string); ok && vPrefix != "" {
			routeMatch.Prefix = aws.String(vPrefix)
		}

		route.Match = routeMatch
	}

	return route
}

func flattenAppmeshGatewayRouteSpec(spec *appmesh.GatewayRouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"grpc_route":  flattenAppmeshGrpcGatewayRoute(spec.GrpcRoute),
		"http2_route": flattenAppmeshHttpGatewayRoute(spec.Http2Route),
		"http_route":  flattenAppmeshHttpGatewayRoute(spec.HttpRoute),
	}

	return []interface{}{mSpec}
}

func flattenAppmeshGatewayRouteTarget(routeTarget *appmesh.GatewayRouteTarget) []interface{} {
	if routeTarget == nil {
		return []interface{}{}
	}

	mRouteTarget := map[string]interface{}{}

	if virtualService := routeTarget.VirtualService; virtualService != nil {
		mVirtualService := map[string]interface{}{
			"virtual_service_name": aws.StringValue(virtualService.VirtualServiceName),
		}

		mRouteTarget["virtual_service"] = []interface{}{mVirtualService}
	}

	return []interface{}{mRouteTarget}
}

func flattenAppmeshGrpcGatewayRoute(grpcRoute *appmesh.GrpcGatewayRoute) []interface{} {
	if grpcRoute == nil {
		return []interface{}{}
	}

	mGrpcRoute := map[string]interface{}{}

	if routeAction := grpcRoute.Action; routeAction != nil {
		mRouteAction := map[string]interface{}{
			"target": flattenAppmeshGatewayRouteTarget(routeAction.Target),
		}

		mGrpcRoute["action"] = []interface{}{mRouteAction}
	}

	if routeMatch := grpcRoute.Match; routeMatch != nil {
		mRouteMatch := map[string]interface{}{
			"service_name": aws.StringValue(routeMatch.ServiceName),
		}

		mGrpcRoute["match"] = []interface{}{mRouteMatch}
	}

	return []interface{}{mGrpcRoute}
}

func flattenAppmeshHttpGatewayRoute(httpRoute *appmesh.HttpGatewayRoute) []interface{} {
	if httpRoute == nil {
		return []interface{}{}
	}

	mHttpRoute := map[string]interface{}{}

	if routeAction := httpRoute.Action; routeAction != nil {
		mRouteAction := map[string]interface{}{
			"target": flattenAppmeshGatewayRouteTarget(routeAction.Target),
		}

		mHttpRoute["action"] = []interface{}{mRouteAction}
	}

	if routeMatch := httpRoute.Match; routeMatch != nil {
		mRouteMatch := map[string]interface{}{
			"prefix": aws.StringValue(routeMatch.Prefix),
		}

		mHttpRoute["match"] = []interface{}{mRouteMatch}
	}

	return []interface{}{mHttpRoute}
}
