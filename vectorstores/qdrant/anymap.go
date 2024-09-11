package qdrant

import (
	"fmt"

	qc "github.com/qdrant/go-client/qdrant"
)

// Converts a map of string to *grpc.Value back to a map of string to any.
// Returns an error if the conversion fails.
func NewAnyMap(valueMap map[string]*qc.Value) (map[string]interface{}, error) {
	inverseMap := make(map[string]interface{})
	for key, val := range valueMap {
		inverseVal, err := NewAnyValue(val)
		if err != nil {
			return nil, err
		}
		inverseMap[key] = inverseVal
	}
	return inverseMap, nil
}

// Converts a *Value back to a generic Go interface.
func NewAnyValue(v *qc.Value) (interface{}, error) {
	switch v.GetKind().(type) {
	case *qc.Value_NullValue:
		return nil, nil // nolint: nilnil
	case *qc.Value_BoolValue:
		return v.GetBoolValue(), nil
	case *qc.Value_IntegerValue:
		return v.GetIntegerValue(), nil
	case *qc.Value_DoubleValue:
		return v.GetDoubleValue(), nil
	case *qc.Value_StringValue:
		return v.GetStringValue(), nil
	case *qc.Value_StructValue:
		structVal := v.GetStructValue()
		return InverseStruct(structVal)
	case *qc.Value_ListValue:
		listVal := v.GetListValue()
		return InverseListValue(listVal)
	default:
		return nil, fmt.Errorf("unknown Value type: %T", v.GetKind())
	}
}

// Converts a ListValue back to a slice of interfaces.
func InverseListValue(lv *qc.ListValue) ([]interface{}, error) {
	res := make([]interface{}, len(lv.GetValues()))
	for i, v := range lv.GetValues() {
		val, err := NewAnyValue(v)
		if err != nil {
			return nil, err
		}
		res[i] = val
	}
	return res, nil
}

// Converts a Struct back to a map of string to any.
func InverseStruct(sv *qc.Struct) (map[string]interface{}, error) {
	res := make(map[string]interface{}, len(sv.GetFields()))
	for k, v := range sv.GetFields() {
		val, err := NewAnyValue(v)
		if err != nil {
			return nil, err
		}
		res[k] = val
	}
	return res, nil
}
