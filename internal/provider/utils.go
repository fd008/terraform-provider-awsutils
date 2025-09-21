// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func decodeAny(ctx context.Context, m any) (value attr.Value, diags diag.Diagnostics) {
	switch v := m.(type) {
	case string:
		value = types.StringValue(v)
	case nil:
		value = types.DynamicNull()
	case float64:
		value = types.NumberValue(big.NewFloat(float64(v)))
	case bool:
		value = types.BoolValue(v)
	case []any:
		return decodeList(ctx, v)
	case map[string]any:
		return decodeMap(ctx, v)
	default:
		diags.Append(diag.NewErrorDiagnostic("failed to decode", fmt.Sprintf("unexpected type: %T for value %#v", v, v)))
	}
	return
}

func decodeList(ctx context.Context, s []any) (attr.Value, diag.Diagnostics) {
	lv := make([]attr.Value, len(s))
	lt := make([]attr.Type, len(s))

	for i, v := range s {
		vv, diags := decodeAny(ctx, v)
		if diags.HasError() {
			return nil, diags
		}
		lv[i] = vv
		lt[i] = vv.Type(ctx)
	}

	return types.TupleValue(lt, lv)
}

func decodeMap(ctx context.Context, m map[string]any) (attr.Value, diag.Diagnostics) {
	mv := make(map[string]attr.Value, len(m))
	mt := make(map[string]attr.Type, len(m))

	for k, v := range m {
		vv, diags := decodeAny(ctx, v)
		if diags.HasError() {
			return nil, diags
		}
		mv[k] = vv
		mt[k] = vv.Type(ctx)
	}

	return types.ObjectValue(mt, mv)
}

func dynamicToGoType(d types.Dynamic) (any, error) {
	if d.IsNull() || d.IsUnknown() {
		return nil, nil
	}
	uv := d.UnderlyingValue()

	switch v := uv.(type) {
	case types.String:
		return v.ValueString(), nil
	case types.Number:
		f, _ := v.ValueBigFloat().Float64()
		return f, nil
	case types.Bool:
		return v.ValueBool(), nil
	case types.Tuple:
		var list []any
		for _, e := range v.Elements() {
			ev, err := dynamicToGoType(e.(types.Dynamic))
			if err != nil {
				return nil, err
			}
			list = append(list, ev)
		}
		return list, nil
	case types.Object:
		m := make(map[string]any)
		for k, e := range v.Attributes() {
			ev, err := dynamicToGoType(e.(types.Dynamic))
			if err != nil {
				return nil, err
			}
			m[k] = ev
		}
		return m, nil
	case types.Map:
		m := make(map[string]any)
		for k, e := range v.Elements() {
			ev, err := dynamicToGoType(e.(types.Dynamic))
			if err != nil {
				return nil, err
			}
			m[k] = ev
		}
		return m, nil
	case types.List:
		var list []any
		for _, e := range v.Elements() {
			ev, err := dynamicToGoType(e.(types.Dynamic))
			if err != nil {
				return nil, err
			}
			list = append(list, ev)
		}
		return list, nil
	case types.Dynamic:
		return dynamicToGoType(v)
	case types.Set:
		var list []any
		for _, e := range v.Elements() {
			ev, err := dynamicToGoType(e.(types.Dynamic))
			if err != nil {
				return nil, err
			}
			list = append(list, ev)
		}
		return list, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}

}
