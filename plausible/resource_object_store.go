package plausible

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	// "github.com/hashicorp/terraform/helper/validation"
)

const s3BucketCreationTimeout = 2 * time.Minute

func resourceObjectStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObjectStoreCreate,
		ReadContext:   resourceObjectStoreRead,
		UpdateContext: resourceObjectStoreUpdate,
		DeleteContext: resourceObjectStoreDelete,
		Schema: map[string]*schema.Schema{
			"store_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"store_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 63),
			},
			"store_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"store_name"},
				ValidateFunc:  validation.StringLenBetween(0, 63-resource.UniqueIDSuffixLength),
			},

			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"key_type": {
				Type:    schema.TypeString,
				Default: "simple",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := strings.ToLower(val.(string))
					if v != "simple" && v != "composite" {
						errs = append(errs, fmt.Errorf("%q must be 'simple' or 'composite', not %s", key, v))
					}
					return
				},
			},
			"key_component": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parent": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"terminal": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"regex": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"enum"},
						},
						"enum": {
							Type:          schema.TypeSet,
							Optional:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							ConflictsWith: []string{"regex"},
						},
					},
				},
			},
		},
	}
}

func resourceObjectStoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).s3conn

	// Get the bucket and acl
	var store_name string
	if v, ok := d.GetOk("store_name"); ok {
		store_name = v.(string)
	} else if v, ok := d.GetOk("store_prefix"); ok {
		store_name = resource.PrefixedUniqueId(v.(string))
	} else {
		store_name = resource.UniqueId()
	}
	d.Set("store_name", store_name)

	req := &s3.CreateBucketInput{
		Bucket: aws.String(store_name),
	}

	awsRegion := m.(*AWSClient).region
	if awsRegion != "us-east-1" {
		req.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(awsRegion),
		}
	}

	if err := validateS3BucketName(store_name, awsRegion); err != nil {
		return diag.Errorf("Error validating S3 bucket name: %s", err)
	}

	_, err := conn.CreateBucket(req)
	if err != nil {
		return diag.Errorf("Error creating S3 bucket: %s", err)
	}

	d.SetId(store_name)
	return resourceObjectStoreRead(ctx, d, m)
}

func resourceObjectStoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).s3conn

	input := &s3.HeadBucketInput{
		Bucket: aws.String(d.Id()),
	}

	err := resource.Retry(s3BucketCreationTimeout, func() *resource.RetryError {
		_, err := conn.HeadBucket(input)

		if d.IsNewResource() && isAWSErrRequestFailureStatusCode(err, 404) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return diag.Errorf("error reading S3 Bucket (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: m.(*AWSClient).partition,
		Service:   "s3",
		Resource:  d.Id(),
	}.String()
	d.Set("uri", arn)

	return nil

}

func resourceObjectStoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// conn := m.(*AWSClient).s3conn

	return resourceObjectStoreRead(ctx, d, m)
}

func resourceObjectStoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn := m.(*AWSClient).s3conn

	log.Printf("[DEBUG] S3 Delete Bucket: %s", d.Id())
	_, err := conn.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(d.Id()),
	})

	if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting S3 Bucket (%s): %s", d.Id(), err)
	}

	return nil
}

func validateS3BucketName(value string, region string) error {
	if region != "us-east-1" {
		if (len(value) < 3) || (len(value) > 63) {
			return fmt.Errorf("%q must contain from 3 to 63 characters", value)
		}
		if !regexp.MustCompile(`^[0-9a-z-.]+$`).MatchString(value) {
			return fmt.Errorf("only lowercase alphanumeric characters and hyphens allowed in %q", value)
		}
		if regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`).MatchString(value) {
			return fmt.Errorf("%q must not be formatted as an IP address", value)
		}
		if strings.HasPrefix(value, `.`) {
			return fmt.Errorf("%q cannot start with a period", value)
		}
		if strings.HasSuffix(value, `.`) {
			return fmt.Errorf("%q cannot end with a period", value)
		}
		if strings.Contains(value, `..`) {
			return fmt.Errorf("%q can be only one period between labels", value)
		}
	} else {
		if len(value) > 255 {
			return fmt.Errorf("%q must contain less than 256 characters", value)
		}
		if !regexp.MustCompile(`^[0-9a-zA-Z-._]+$`).MatchString(value) {
			return fmt.Errorf("only alphanumeric characters, hyphens, periods, and underscores allowed in %q", value)
		}
	}
	return nil
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func isAWSErr(err error, code string, message string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == code && strings.Contains(awsErr.Message(), message)
	}
	return false
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.RequestFailure
//  * RequestFailure.StatusCode() matches status code
// It is always preferable to use isAWSErr() except in older APIs (e.g. S3)
// that sometimes only respond with status codes.
func isAWSErrRequestFailureStatusCode(err error, statusCode int) bool {
	var awsErr awserr.RequestFailure
	if errors.As(err, &awsErr) {
		return awsErr.StatusCode() == statusCode
	}
	return false
}
