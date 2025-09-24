// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

// import (
// 	"context"
// 	awscloud "terraform-provider-awsutils/internal/aws_cloud"

// 	"github.com/hashicorp/terraform-plugin-framework/attr"
// 	"github.com/hashicorp/terraform-plugin-framework/datasource"
// 	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/hashicorp/terraform-plugin-log/tflog"

// 	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator" // Import the validator package
// )

// // Ensure RdsDataBatchExecute implements the DataSource interface.
// var _ datasource.DataSource = &RdsDataBatchExecute{}

// // NewRdsDataBatchExecute is a helper function to simplify the provider implementation.
// func NewRdsDataBatchExecute() datasource.DataSource {
// 	return &RdsDataBatchExecute{}
// }

// type RdsDataBatchExecute struct{}

// type RdsDataBatchExecuteModel struct {
// 	ResourceArn           types.String `tfsdk:"resource_arn"`
// 	SecretArn             types.String `tfsdk:"secret_arn"`
// 	Database              types.String `tfsdk:"database"`
// 	SQL                   types.String `tfsdk:"sql"`
// 	Parameters            types.List   `tfsdk:"parameters"`
// 	Region                types.String `tfsdk:"region"`
// 	ContinueAfter         types.Bool   `tfsdk:"continue_after_timeout"`
// 	IncludeResultMetadata types.Bool   `tfsdk:"include_result_metadata"`
// 	Result                types.String `tfsdk:"result"`
// }

// func (d *RdsDataBatchExecute) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
// 	resp.TypeName = req.ProviderTypeName + "_rds_data_batch_execute_statement"
// }

// func (d *RdsDataBatchExecute) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
// 	resp.Schema = schema.Schema{
// 		Description: "Executes a SQL statement against an Amazon RDS database.",
// 		Attributes: map[string]schema.Attribute{
// 			"resource_arn": schema.StringAttribute{
// 				Required:    true,
// 				Description: "The Amazon Resource Name (ARN) of the Aurora Serverless DB cluster or the RDS DB instance.",
// 			},
// 			"secret_arn": schema.StringAttribute{
// 				Required:    true,
// 				Description: "The ARN of the secret that enables access to the DB cluster or instance.",
// 			},
// 			"database": schema.StringAttribute{
// 				Required:    true,
// 				Description: "The name of the database.",
// 			},
// 			"sql": schema.StringAttribute{
// 				Required:    true,
// 				Description: "The SQL statement to run.",
// 			},
// 			"parameters": schema.ListNestedAttribute{
// 				CustomType: types.ListType{
// 					ElemType: types.ObjectType{
// 						AttrTypes: map[string]attr.Type{
// 							"name":     types.StringType,
// 							"typeHint": types.StringType,
// 							"type":     types.StringType,
// 							"value":    types.StringType,
// 						},
// 					},
// 				},
// 				Optional:    true,
// 				Description: "The parameters for the SQL statement.",
// 				Validators: []validator.List{
// 					listvalidator.SizeAtLeast(1),
// 				},
// 			},
// 			"region": schema.StringAttribute{
// 				Optional:    true,
// 				Description: "The AWS region where the RDS instance is located. If not specified, defaults to the region configured in the provider.",
// 			},
// 			"continue_after_timeout": schema.BoolAttribute{
// 				Optional:    true,
// 				Description: "A value that indicates whether to continue running the statement after the call times out. By default, the statement stops running when the call times out.",
// 			},
// 			"include_result_metadata": schema.BoolAttribute{
// 				Optional:    true,
// 				Description: "A value that indicates whether to include metadata in the results.",
// 			},
// 			"result": schema.StringAttribute{
// 				Computed:    true,
// 				Description: "The result of the SQL execution in JSON format.",
// 			},
// 		},
// 	}
// }

// func (d *RdsDataBatchExecute) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
// 	var state RdsDataBatchExecuteModel
// 	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	exec := &awscloud.ExecuteBatchModel{
// 		ResourceArn: state.ResourceArn.ValueString(),
// 		SecretArn:   state.SecretArn.ValueString(),
// 		Database:    state.Database.ValueString(),
// 		SQL:         state.SQL.ValueString(),
// 		Region:      state.Region.ValueStringPointer(),
// 	}

// 	tflog.Debug(ctx, "params: "+state.Parameters.String())

// 	var params = make([]map[string]awscloud.ParameterModel, len(state.Parameters.Elements()))

// 	if !state.Parameters.IsNull() && !state.Parameters.IsUnknown() {
// 		paramList := state.Parameters.Elements()
// 		for _, p := range paramList {
// 			if p.IsNull() || p.IsUnknown() {
// 				continue
// 			}
// 			var obj = make(map[string]awscloud.ParameterModel)
// 			paramMap := p.(types.Object)
// 			for key, attr := range paramMap.Attributes() {
// 				if attr.IsNull() || attr.IsUnknown() {
// 					continue
// 				}
// 				if key == "item" {
// 					itemObj := attr.(types.Object)
// 					var item awscloud.ParameterModel
// 					for itemKey, itemAttr := range itemObj.Attributes() {
// 						if itemAttr.IsNull() || itemAttr.IsUnknown() {
// 							continue
// 						}
// 						switch itemKey {
// 						case "type":
// 							item.Type = itemAttr.(types.String).ValueString()
// 						case "value":
// 							item.Value = itemAttr.(types.String).ValueString()
// 						}
// 					}
// 					obj[key] = item
// 				} else if key == "key" {
// 					obj[key] = awscloud.ParameterModel{
// 						Value: attr.(types.String).ValueString(),
// 					}
// 				}
// 			}
// 			params = append(params, obj)
// 		}
// 		exec.Parameters = params
// 	}

// 	result, err := exec.ExecuteBatchStatement()

// 	if err != nil {
// 		resp.Diagnostics.AddError(
// 			"Error executing SQL statement",
// 			err.Error(),
// 		)
// 		return
// 	}

// 	state.Result = types.StringValue(result)
// 	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
// }
