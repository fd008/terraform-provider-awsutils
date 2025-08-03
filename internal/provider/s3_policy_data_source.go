// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &s3PolicyDataSrouce{}
)

// News3PolicyDataSrouce is a helper function to simplify the provider implementation.
func NewS3PolicyDataSrouce() datasource.DataSource {
	return &s3PolicyDataSrouce{}
}

// s3PolicyDataSrouce is the data source implementation.
type s3PolicyDataSrouce struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Policy     types.String `tfsdk:"policy"`
	Region     types.String `tfsdk:"region"`
	Strict     types.Bool   `tfsdk:"strict"`
}

// Metadata returns the data source type name.
func (d *s3PolicyDataSrouce) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_policy"
}

// Schema defines the schema for the data source.
func (d *s3PolicyDataSrouce) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Description of the data source
		Description: "Retrieves the default key policy for a given AWS S3 key.",
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true, // The key ID is a required input
				Description: "The name of the S3 bucket.",
			},
			"region": schema.StringAttribute{
				Optional:    true, // The region is optional
				Description: "The AWS region where the S3 is located. If not specified, defaults to the region configured in the provider.",
			},
			"policy": schema.StringAttribute{
				Computed:    true, // The policy is retrieved and thus computed
				Description: "The S3 key policy in JSON format.",
			},
			"strict": schema.BoolAttribute{
				Optional:    true, // The strict mode is optional
				Description: "If true, will throw an error if the key policy is not found. Defaults to false.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *s3PolicyDataSrouce) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state s3PolicyDataSrouce

	// Get configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If 'strict' is not set by the user, provide the default value.
	if state.Strict.IsNull() || state.Strict.IsUnknown() {
		state.Strict = types.BoolValue(false) // Set the default value to false
	}

	// Get the S3 Key ID from the configuration
	bucketName := state.BucketName.ValueString()
	region := state.Region.ValueString()

	policy, pErr := awscloud.GetS3BucketPolicy(bucketName, &region)

	// If the policy retrieval fails and strict mode is enabled, return an error.
	if pErr != nil && state.Strict.ValueBool() {
		resp.Diagnostics.AddError(
			"Error retrieving S3 key policy",
			"Unable to retrieve the S3 key policy for key ID "+bucketName+": "+pErr.Error(),
		)
		return
	}

	// If the policy retrieval fails and strict mode is not enabled, log a warning.
	if pErr != nil && !state.Strict.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Error retrieving S3 key policy",
			"Unable to retrieve the S3 key policy for key ID "+bucketName+": "+pErr.Error(),
		)

		state.Policy = types.StringValue("")
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	state.Policy = types.StringValue(*policy)

	// Set the new state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

}
