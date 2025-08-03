// Copyright (c) Fuadul Hasan, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CfResource{}
var _ resource.ResourceWithImportState = &CfResource{}

func NewCfResource() resource.Resource {
	return &CfResource{}
}

// CfResource defines the resource implementation.
type CfResource struct {
	cfg aws.Config
}

// CfResourceModel describes the resource data model.
type CfResourceModel struct {
	Distribution_Id types.String `tfsdk:"distribution_id"`
	Paths           types.List   `tfsdk:"paths"`
	InValidation_Id types.String `tfsdk:"invalidation_id"`
	// Status        types.String `tfsdk:"status"`
	Trigger types.String `tfsdk:"trigger"`
}

func (r *CfResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + "cloudfront_invalidation"
}

func (r *CfResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CloudFront cache invalidation resource",
		Attributes: map[string]schema.Attribute{

			"distribution_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloudfront distribution ID",
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.RequiresReplaceIfConfigured(),
				// },
			},
			"paths": schema.ListAttribute{
				MarkdownDescription: "Cache invalidation paths - defaults to `/*`",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("/*"),
						},
					),
				),
				// PlanModifiers: []planmodifier.List{
				// 	listplanmodifier.RequiresReplaceIfConfigured(),
				// },
			},
			"invalidation_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cloudfront cache invalidation ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			// "status": schema.StringAttribute{
			// 	Computed:            true,
			// 	MarkdownDescription: "Cloudfront cache invalidation status",
			// 	PlanModifiers: []planmodifier.String{
			// 		// stringplanmodifier.UseStateForUnknown(),
			// 		stringplanmodifier.RequiresReplaceIfConfigured(),
			// 	},
			// },
			"trigger": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Trigger cache invalidation. Setting unique value each time will trigger a cache invalidation on apply",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				// Default:             stringdefault.StaticString(time.Now().Format(time.RFC3339)),
			},
		},
	}

}

func (r *CfResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// provider configuration
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(aws.Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *aws.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.cfg = cfg

}

func (r *CfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// state
	var state CfResourceModel
	resp.State.Get(ctx, &state)

	if resp.Diagnostics.HasError() {
		return
	}

	// plan
	var data CfResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// override user
	// if !data.Status.IsNull() {
	// 	data.Status = state.Status
	// }

	if !data.InValidation_Id.IsNull() {
		data.InValidation_Id = state.InValidation_Id
	}

	// Convert data.Paths to []string
	paths := make([]string, 0, len(data.Paths.Elements()))
	data.Paths.ElementsAs(ctx, &paths, true)

	// resp.Diagnostics.AddError("info", fmt.Sprintf("Invalidating cache info...%T", paths))

	// invalidate the cache
	cacheRes, err := awscloud.InvalidateCache(r.cfg, data.Distribution_Id.ValueString(), paths)

	if err != nil {
		resp.Diagnostics.AddError("Error", fmt.Sprint("Unable to invalidate cache...", err.Error()))
		// tflog.Error(ctx, fmt.Sprint("Unable to invalidate cache...", err))
		return
	}

	data.InValidation_Id = types.StringPointerValue(cacheRes.Invalidation.Id)
	// data.Status = types.StringPointerValue(cacheRes.Invalidation.Status)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CfResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CfResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := awscloud.GetDistribution(r.cfg, data.Distribution_Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error", fmt.Sprint("Unable to get distribution...", err))
		tflog.Error(ctx, fmt.Sprint("Unable to get distribution...", err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("cf distribution: %v ", res))
	// data.Status = types.StringPointerValue(res.Distribution.Status)

	// // Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// var data CfResourceModel

	// // Read Terraform plan data into the model
	// resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // Convert data.Paths to []string
	// paths := make([]string, 0, len(data.Paths.Elements()))
	// data.Paths.ElementsAs(ctx, &paths, true)

	// // resp.Diagnostics.AddError("info", fmt.Sprintf("Invalidating cache info...%T", paths))

	// // invalidate the cache
	// cacheRes, err := awscloud.InvalidateCache(r.cfg, data.Distribution_Id.ValueString(), paths)

	// if err != nil {
	// 	resp.Diagnostics.AddError("Error", fmt.Sprint("Unable to invalidate cache...", err, " | ", data.Id.ValueString()))
	// 	tflog.Error(ctx, fmt.Sprint("Unable to invalidate cache...", err))
	// 	return
	// }

	// resp.State.Set(ctx, &cacheRes)

	// data.Validation_Id = types.StringPointerValue(cacheRes.Invalidation.Id)
	// // data.Status = types.StringPointerValue(cacheRes.Invalidation.Status)
	// // data.Status = types.StringValue("Completed")

	// // Save data into Terraform state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CfResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//no-op
}

func (r *CfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
