package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// Map is a JSON object scalar for answer maps.
type Map map[string]string

// UnmarshalGQL implements graphql.Unmarshaler.
func (m *Map) UnmarshalGQL(v any) error {
	parsed, err := UnmarshalMap(v)
	if err != nil {
		return err
	}
	*m = Map(parsed)
	return nil
}

// MarshalGQL implements graphql.Marshaler.
func (m Map) MarshalGQL(w io.Writer) {
	if m == nil {
		_, _ = w.Write([]byte("null"))
		return
	}
	data, _ := json.Marshal(map[string]string(m))
	_, _ = w.Write(data)
}

// UnmarshalMap decodes a GraphQL map argument.
func UnmarshalMap(v any) (map[string]string, error) {
	switch value := v.(type) {
	case map[string]any:
		out := make(map[string]string, len(value))
		for k, raw := range value {
			s, ok := raw.(string)
			if !ok {
				return nil, fmt.Errorf("map value for %q must be string", k)
			}
			out[k] = s
		}
		return out, nil
	case map[string]string:
		return value, nil
	case nil:
		return map[string]string{}, nil
	default:
		return nil, fmt.Errorf("invalid map type %T", v)
	}
}

// MarshalMap encodes a map for GraphQL.
func MarshalMap(m map[string]string) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		Map(m).MarshalGQL(w)
	})
}
