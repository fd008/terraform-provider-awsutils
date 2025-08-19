// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &S3UploadResource{}
var _ resource.ResourceWithImportState = &S3UploadResource{}

func NewS3UploadResource() resource.Resource {
	return &S3UploadResource{}
}

// S3UploadResource defines the resource implementation.
type S3UploadResource struct {
	cfg aws.Config
}

// S3UploadResourceModel describes the resource data model.
type S3UploadResourceModel struct {
	BucketName    types.String `tfsdk:"bucket_name"`
	DirPath       types.String `tfsdk:"dir_path"`
	ExclusionList types.List   `tfsdk:"exclusion_list"`
	Trigger       types.String `tfsdk:"trigger"`
	Region        types.String `tfsdk:"region"`
	Prefix        types.String `tfsdk:"prefix"`
	MimeMap       types.Map    `tfsdk:"mime_map"`
}

func (r *S3UploadResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + "s3_dir_upload"
}

func (r *S3UploadResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Upload a directory to S3 bucket efficiently, with support for large files and concurrency.",
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "S3 Bucket Name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dir_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Directory path to upload to S3 bucket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "AWS Region where the S3 bucket is located. If not specified, defaults to the region configured in the provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional prefix for the S3 bucket path. If not specified, files will be uploaded to the root of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"exclusion_list": schema.ListAttribute{
				MarkdownDescription: "List of file patterns to exclude from upload. Patterns can include wildcards like `*.tmp` or specific file names.",
				Description:         "List of file patterns to exclude from upload. Patterns can include wildcards like `*.tmp` or specific file names.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringNull().Type(context.TODO()),
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{},
					),
				),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Trigger for the upload operation. Setting this to a new value will trigger the upload operation. If not set, defaults to the current timestamp.",
				Description:         "Trigger for the upload operation. Setting this to a new value will trigger the upload operation. If not set, defaults to the current timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mime_map": schema.MapAttribute{
				MarkdownDescription: "Custom MIME types for specific file extensions. This map allows you to",
				Description:         "Custom MIME types for specific file extensions. This map allows you to define how files are served based on their extensions.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default: mapdefault.StaticValue(
					types.MapValueMust(
						types.StringType,
						map[string]attr.Value{},
					),
				),
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
	}

}

func (r *S3UploadResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *S3UploadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// state
	var plan S3UploadResourceModel // Get the planned state from Terraform

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exList := plan.ExclusionList
	var exSlice []string

	resp.Diagnostics.Append(exList.ElementsAs(ctx, &exSlice, false)...) //
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Region.IsNull() || plan.Region.ValueString() != "" {
		r.cfg.Region = plan.Region.ValueString()
	}

	cfg, err := awscloud.GetConfig(&r.cfg.Region, nil, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring AWS S3 upload",
			"Unable to create AWS configuration: "+err.Error(),
		)
	}

	if plan.Trigger.IsNull() || plan.Trigger.ValueString() == "" {
		now := time.Now()

		plan.Trigger = types.StringValue(fmt.Sprint(now.Unix()))
	}

	mimeMap := make(map[string]string)
	if !plan.MimeMap.IsNull() {
		resp.Diagnostics.Append(plan.MimeMap.ElementsAs(ctx, &mimeMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	awscloud.Upload(cfg, plan.DirPath.ValueString(), plan.BucketName.ValueString(), plan.Prefix.ValueStringPointer(), &exSlice, &mimeMap)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *S3UploadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//no-op
}

func (r *S3UploadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// no-op

}

func (r *S3UploadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//no-op
}

func (r *S3UploadResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
