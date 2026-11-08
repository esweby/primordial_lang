package evaluator

import "github.com/esweby/primordial_lang/object"

// copyValue deep copies a value being assigned, so that the two names do not
// share a container. Integers, strings and booleans are never changed in
// place, and functions are shared, so they are returned as they are.
func copyValue(obj object.Object) object.Object {
	switch value := obj.(type) {
	case *object.Array:
		return &object.Array{Elements: copyValues(value.Elements)}
	case *object.Slice:
		return &object.Slice{Elements: copyValues(value.Elements)}
	case *object.Tuple:
		return &object.Tuple{Elements: copyValues(value.Elements)}
	case *object.Struct:
		fields := make(map[string]object.Object, len(value.Fields))
		for name, field := range value.Fields {
			fields[name] = copyValue(field)
		}
		return &object.Struct{Name: value.Name, Definition: value.Definition, Fields: fields}
	case *object.Map:
		pairs := make(map[object.HashKey]object.Object, len(value.Pairs))
		originalKeys := make(map[object.HashKey]object.Object, len(value.OriginalKeys))
		for key, pair := range value.Pairs {
			pairs[key] = copyValue(pair)
			originalKeys[key] = value.OriginalKeys[key]
		}
		return &object.Map{
			MapType:      value.MapType,
			Pairs:        pairs,
			Keys:         append([]object.HashKey(nil), value.Keys...),
			OriginalKeys: originalKeys,
		}
	default:
		return obj
	}
}

func copyValues(values []object.Object) []object.Object {
	copies := make([]object.Object, len(values))
	for i, value := range values {
		copies[i] = copyValue(value)
	}
	return copies
}
