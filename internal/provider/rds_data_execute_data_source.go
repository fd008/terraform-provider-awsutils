// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure RdsDataExecute implements the DataSource interface.
var _ datasource.DataSource = &RdsDataExecute{}

// NewRdsDataExecute is a helper function to simplify the provider implementation.
func NewRdsDataExecute() datasource.DataSource {
	return &RdsDataExecute{}
}

type RdsDataExecute struct{}

type RdsDataExecuteModel struct {
	ResourceArn           types.String `tfsdk:"resource_arn"`
	SecretArn             types.String `tfsdk:"secret_arn"`
	Database              types.String `tfsdk:"database"`
	SQL                   types.String `tfsdk:"sql"`
	Parameters            types.Map    `tfsdk:"parameters"`
	Region                types.String `tfsdk:"region"`
	ContinueAfter         types.Bool   `tfsdk:"continue_after_timeout"`
	IncludeResultMetadata types.Bool   `tfsdk:"include_result_metadata"`
	Result                types.String `tfsdk:"result"`
}

func (d *RdsDataExecute) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rds_data_execute_statement"
}

func (d *RdsDataExecute) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Executes a SQL statement against an Amazon RDS database.",
		Attributes: map[string]schema.Attribute{
			"resource_arn": schema.StringAttribute{
				Required:    true,
				Description: "The Amazon Resource Name (ARN) of the Aurora Serverless DB cluster or the RDS DB instance.",
			},
			"secret_arn": schema.StringAttribute{
				Required:    true,
				Description: "The ARN of the secret that enables access to the DB cluster or instance.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "The name of the database.",
			},
			"sql": schema.StringAttribute{
				Required:    true,
				Description: "The SQL statement to run.",
			},
			"parameters": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the parameter. Valid types are: array, blob, boolean, double, float, integer, long, null, string, and struct.",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The value of the parameter.",
						},
					},
				},
				Optional:    true,
				Description: "The parameters for the SQL statement.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "The AWS region where the RDS instance is located. If not specified, defaults to the region configured in the provider.",
			},
			"continue_after_timeout": schema.BoolAttribute{
				Optional:    true,
				Description: "A value that indicates whether to continue running the statement after the call times out. By default, the statement stops running when the call times out.",
			},
			"include_result_metadata": schema.BoolAttribute{
				Optional:    true,
				Description: "A value that indicates whether to include metadata in the results.",
			},
			"result": schema.StringAttribute{
				Computed:    true,
				Description: "The result of the SQL execution in JSON format.",
			},
		},
	}
}

func (d *RdsDataExecute) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state RdsDataExecuteModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exec := &awscloud.ExecuteModel{
		ResourceArn:           state.ResourceArn.ValueString(),
		SecretArn:             state.SecretArn.ValueString(),
		Database:              state.Database.ValueString(),
		SQL:                   state.SQL.ValueString(),
		Region:                state.Region.ValueStringPointer(),
		ContinueAfterTimeout:  state.ContinueAfter.ValueBool(),
		IncludeResultMetadata: state.IncludeResultMetadata.ValueBool(),
	}

	tflog.Debug(ctx, "params: "+state.Parameters.String())

	var params = make(map[string]awscloud.ParameterModel, len(state.Parameters.Elements()))

	if !state.Parameters.IsNull() && !state.Parameters.IsUnknown() {

		for k, v := range state.Parameters.Elements() {
			obj := v.(types.Object).Attributes()

			params[k] = awscloud.ParameterModel{
				Type:  obj["type"].(types.String).ValueString(),
				Value: obj["value"].(types.String).ValueString(),
			}
		}
	}

	sqlParams, err := exec.ToParams(params)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting parameters",
			err.Error(),
		)
		return
	}

	exec.Parameters = sqlParams

	result, err := exec.ExecuteStatement()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing statement",
			err.Error(),
		)
		return
	}

	state.Result = types.StringValue(result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
