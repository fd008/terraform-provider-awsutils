// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
)

type ExecuteModel struct {
	ResourceArn           string
	SecretArn             string
	Database              string
	SQL                   string
	Parameters            []types.SqlParameter
	Region                *string
	ContinueAfterTimeout  bool
	IncludeResultMetadata bool
}

func (e *ExecuteModel) ToParams(params map[string]ParameterModel) ([]types.SqlParameter, error) {
	var sqlParams []types.SqlParameter

	for name, param := range params {
		switch param.Type {
		case "string":
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  &name,
				Value: &types.FieldMemberStringValue{Value: param.Value},
			})
		case "long":
			longValue, err := strconv.ParseInt(param.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid long value for parameter %s: %w", name, err)
			}
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  &name,
				Value: &types.FieldMemberLongValue{Value: longValue},
			})
		case "double":
			doubleValue, err := strconv.ParseFloat(param.Value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid double value for parameter %s: %w", name, err)
			}
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  &name,
				Value: &types.FieldMemberDoubleValue{Value: doubleValue},
			})
		case "boolean":
			boolValue, err := strconv.ParseBool(param.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid boolean value for parameter %s: %w", name, err)
			}
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  &name,
				Value: &types.FieldMemberBooleanValue{Value: boolValue},
			})
		case "json":
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  &name,
				Value: &types.FieldMemberStringValue{Value: param.Value},
			})
		case "blob":
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  &name,
				Value: &types.FieldMemberBlobValue{Value: []byte(param.Value)},
			})
		default:
			return nil, fmt.Errorf("unsupported parameter type %s for parameter %s", param.Type, name)
		}
	}
	return sqlParams, nil
}

func (e *ExecuteModel) ExecuteStatement() (string, error) {
	cfg, err := GetConfig(e.Region, nil, nil)

	if err != nil {
		return "", err
	}

	// Create a new RDS Data Service client
	client := rdsdata.NewFromConfig(cfg)

	var input rdsdata.ExecuteStatementInput

	input.ResourceArn = &e.ResourceArn
	input.SecretArn = &e.SecretArn
	input.Database = &e.Database
	input.Sql = &e.SQL
	input.ContinueAfterTimeout = e.ContinueAfterTimeout || true
	input.IncludeResultMetadata = e.IncludeResultMetadata || true
	input.Parameters = e.Parameters

	// Call the ExecuteStatement operation
	resp, err := client.ExecuteStatement(context.TODO(), &input)
	if err != nil {
		return "", fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	// Return the result of the SQL execution
	return anyToJson(resp)
}
