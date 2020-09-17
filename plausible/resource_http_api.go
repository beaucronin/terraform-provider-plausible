package plausible

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceHttpApi() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHttpApiCreate,
		ReadContext:   resourceHttpApiRead,
		UpdateContext: resourceHttpApiUpdate,
		DeleteContext: resourceHttpApiDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec_file": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"spec_body": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"uri": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHttpApiCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// var diags diag.Diagnostics

	conn := m.(*AWSClient).apigatewayconn

	// API
	apiInput := &apigateway.CreateRestApiInput{
		// Name: aws.String(d.Get("name").(string)),
		Name: aws.String("temp"),
	}

	gateway, err := conn.CreateRestApi(apiInput)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*gateway.Id)

	specFile := d.Get("spec_file").(string)
	spec, err := ioutil.ReadFile(specFile)
	d.Set("spec_body", string(spec))
	log.Printf("[DEBUG] Initializing API Gateway from OpenAPI spec %s", d.Id())
	_, err = conn.PutRestApi(&apigateway.PutRestApiInput{
		RestApiId: gateway.Id,
		Mode:      aws.String(apigateway.PutModeOverwrite),
		Body:      []byte(spec),
	})

	rest_api_arn := arn.ARN{
		Partition: m.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    m.(*AWSClient).region,
		Resource:  fmt.Sprintf("/restapis/%s", d.Id()),
	}.String()
	d.Set("uri", rest_api_arn)

	// rscs, err := conn.GetResources(&apigateway.GetResourcesInput{
	// 	RestApiId: gateway.Id,
	// })
	// rm := map[string]*string{}
	// for _, item := range rscs.Items {
	// 	rm[*item.Id] = item.Path
	// }
	// if err != nil {
	// 	log.Panic(err)
	// }
	// d.Set("resources", rm)

	// // Deployment
	// deploymentInput := apigateway.CreateDeploymentInput{
	// 	RestApiId: gateway.Id,
	// 	StageName: aws.String("default"),
	// }
	// _, err = conn.CreateDeployment(&deploymentInput)
	// if err != nil {
	// 	return diag.Errorf("Error creating API Gateway Deployment: %s", err)
	// }

	// // Stage
	// stageInput := apigateway.CreateStageInput{
	// 	RestApiId:    gateway.Id,
	// 	StageName:    aws.String("default"),
	// 	DeploymentId: aws.String(d.Get("deployment_id").(string)),
	// }

	// _, err = conn.CreateStage(&stageInput)
	// if err != nil {
	// 	return diag.Errorf("Error creating API Gateway Stage: %s", err)
	// }

	return resourceHttpApiRead(ctx, d, m)
}

func resourceHttpApiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("abc spec_file: %s", d.Get("spec_file"))
	log.Printf("bcd spec_body: %s", d.Get("spec_body"))

	conn := m.(*AWSClient).apigatewayconn
	api, err := conn.GetRestApi(&apigateway.GetRestApiInput{
		RestApiId: aws.String(d.Id()),
	})
	if err != nil {
		return diag.Errorf("Error reading API Gateway")
	}
	d.Set("name", api.Name)

	rscs, err := conn.GetResources(&apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	})
	rm := map[string]*string{}
	for _, item := range rscs.Items {
		rm[*item.Id] = item.Path
	}
	if err != nil {
		log.Panic(err)
	}
	d.Set("resources", rm)

	return diags
}

func resourceHttpApiUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := m.(*AWSClient).apigatewayconn

	if d.HasChange("spec_body") {
		if body, ok := d.GetOk("body"); ok {
			log.Printf("[DEBUG] Updating API Gateway from OpenAPI spec: %s", d.Id())
			_, err := conn.PutRestApi(&apigateway.PutRestApiInput{
				RestApiId: aws.String(d.Id()),
				Mode:      aws.String(apigateway.PutModeOverwrite),
				Body:      []byte(body.(string)),
			})
			if err != nil {
				return diag.Errorf("error updating API Gateway specification: %s", err)
			}
		}
	}
	return diags
}

func resourceHttpApiDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
