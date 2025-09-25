// Copyright (c) HashiCorp, Inc.

package awscloud

import (
	"encoding/json"
	"fmt"
)

type ParameterModel struct {
	Value string
	Type  string
}

// func toParams(params *map[string]ParameterModel) ([]types.SqlParameter, error) {
// 	if params == nil {
// 		return nil, nil
// 	}

// 	var sqlParams []types.SqlParameter
// 	for name, param := range *params {

// 		switch param.Type {
// 		case "string":
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberStringValue{Value: param.Value},
// 			})
// 		case "long":
// 			longValue, err := strconv.ParseInt(param.Value, 10, 64)
// 			if err != nil {
// 				return nil, fmt.Errorf("invalid long value for parameter %s: %w", name, err)
// 			}
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberLongValue{Value: longValue},
// 			})
// 		case "double":
// 			doubleValue, err := strconv.ParseFloat(param.Value, 64)
// 			if err != nil {
// 				return nil, fmt.Errorf("invalid double value for parameter %s: %w", name, err)
// 			}
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberDoubleValue{Value: doubleValue},
// 			})
// 		case "boolean":
// 			boolValue, err := strconv.ParseBool(param.Value)
// 			if err != nil {
// 				return nil, fmt.Errorf("invalid boolean value for parameter %s: %w", name, err)
// 			}
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberBooleanValue{Value: boolValue},
// 			})
// 		case "json":
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberStringValue{Value: param.Value},
// 			})
// 		case "blob":
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberBlobValue{Value: []byte(param.Value)},
// 			})
// 		case "null":
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberIsNull{Value: true},
// 			})
// 		default:
// 			sqlParams = append(sqlParams, types.SqlParameter{
// 				Name:  &name,
// 				Value: &types.FieldMemberStringValue{Value: param.Value},
// 			})
// 		}
// 	}

// 	return sqlParams, nil

// }

func ToJson(data *string) ([]byte, error) {
	var records []map[string]any
	if data != nil {
		err := json.Unmarshal([]byte(*data), &records)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(records, "", "  ")
	}
	return json.MarshalIndent(records, "", "  ")
}

func anyToJson(v any) (string, error) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error converting to JSON: %v", err), nil
	}
	return string(jsonData), nil
}
