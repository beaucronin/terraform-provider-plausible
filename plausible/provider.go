package plausible

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				Description: "The region where AWS operations will take place. Examples\n" +
					"are us-east-1, us-west-2, etc.",
				InputDefault: "us-west-2",
			},
			"app_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of this app, which may be used to uniquely identify resources",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"plausible_function":       resourceFunction(),
			"plausible_http_api":       resourceHttpApi(),
			"plausible_object_store":   resourceObjectStore(),
			"plausible_keyvalue_store": resourceKeyValueStore(),
			// "plausible_stream_analytics": resourceStreamAnalytics(),
			// "plausible_file_store": resourceFileStore(),
			// "plausible_publisher":        resourcePublisher(),
			// "plausible_eventbus":         resourceEventBus(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return provider
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	config := AWSConfig{
		AppName:          d.Get("app_name").(string),
		Region:           d.Get("region").(string),
		terraformVersion: terraformVersion,
		CallerName:       "Plausible|AWS Provider",
	}

	return config.Client()
}
