package plausible

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePublisher() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePublisherCreate,
		ReadContext:   resourcePublisherRead,
		UpdateContext: resourcePublisherUpdate,
		DeleteContext: resourcePublisherDelete,
		Schema: map[string]*schema.Schema{

			"uri": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourcePublisherCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).snsconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	req := &sns.CreateTopicInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateTopic(req)
	if err != nil {
		return diag.Errorf("Error creating SNS topic: %s", err)
	}
	d.SetId(*output.TopicArn)

	if d.HasChange("arn") {
		_, v := d.GetChange("arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "TopicArn", v, conn); err != nil {
			return diag.Errorf("Error updating ARN for SNS topic: %s", err)
		}
	}
	if d.HasChange("display_name") {
		_, v := d.GetChange("display_name")
		if err := updateAwsSnsTopicAttribute(d.Id(), "DisplayName", v, conn); err != nil {
			return diag.Errorf("Error updating DisplayName for SNS topic: %s", err)
		}
	}

	return resourcePublisherRead(ctx, d, m)
}

func resourcePublisherRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).snsconn

	attributeOutput, err := conn.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: aws.String(d.Id()),
	})
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if attributeOutput.Attributes != nil && len(attributeOutput.Attributes) > 0 {
		d.Set("arn", aws.StringValue(attributeOutput.Attributes["TopicArn"]))
		d.Set("display_name", aws.StringValue(attributeOutput.Attributes["DisplayName"]))
	}

	if _, ok := d.GetOk("name"); !ok {
		arn := d.Get("arn").(string)
		idx := strings.LastIndex(arn, ":")
		if idx > -1 {
			d.Set("name", arn[idx+1:])
		}
	}

	return nil
}

func resourcePublisherUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourcePublisherDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func updateAwsSnsTopicAttribute(topicArn, name string, value interface{}, conn *sns.SNS) error {
	// Ignore an empty policy
	if name == "Policy" && value == "" {
		return nil
	}
	log.Printf("[DEBUG] Updating SNS Topic Attribute: %s", name)

	// Make API call to update attributes
	req := sns.SetTopicAttributesInput{
		TopicArn:       aws.String(topicArn),
		AttributeName:  aws.String(name),
		AttributeValue: aws.String(fmt.Sprintf("%v", value)),
	}

	_, err := conn.SetTopicAttributes(&req)

	return err
}
