// Copyright github.com/fd008 - All rights reserved
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &AWSUtilsProvider{}
var _ provider.ProviderWithFunctions = &AWSUtilsProvider{}
var _ provider.ProviderWithEphemeralResources = &AWSUtilsProvider{}

type AWSUtilsProvider struct {
	version string
}

type AWSUtilsProviderModel struct {
	Region                 *string  `tfsdk:"region"`
	SharedConfigFiles      []string `tfsdk:"shared_config_files"`
	SharedCredentialsFiles []string `tfsdk:"shared_credentials_files"`
}

func (p *AWSUtilsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsutils"
	resp.Version = p.version
}

func (p *AWSUtilsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "The region to use for AWS requests. If not set, defaults to us-east-1",
				Optional:    true,
			},
			"shared_config_files": schema.ListAttribute{
				Description: "Path to shared config file. If not set, defaults to ~/.aws/config.",
				Optional:    true,
				ElementType: types.StringNull().Type(context.TODO()),
			},
			"shared_credentials_files": schema.ListAttribute{
				Description: "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
				Optional:    true,
				ElementType: types.StringNull().Type(context.TODO()),
			},
		},
	}
}

func (p *AWSUtilsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AWSUtilsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := awscloud.GetConfig(data.Region, data.SharedConfigFiles, data.SharedCredentialsFiles)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring AWS Utils provider",
			err.Error(),
		)
		return
	}

	resp.DataSourceData = cfg
	resp.ResourceData = cfg
}

func (p *AWSUtilsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCfResource,
		NewS3UploadResource,
		NewOpenAPIMergeResource,
	}
}

func (p *AWSUtilsProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *AWSUtilsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewKmsPolicyDataSrouce,
		NewS3PolicyDataSrouce,
	}
}

func (p *AWSUtilsProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewAwsVarFunction,
		ShallowListFunction,
		MergePolicyFunction,
		FileSetFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AWSUtilsProvider{
			version: version,
		}
	}
}
