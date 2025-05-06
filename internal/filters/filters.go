package filters

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
)

var (
	// customEqualFilter is a custom "==" operator added to the basic
	// gval ones that adds support for json.Number comparison
	customEqualFilter = gval.InfixOperator("==", func(a, b interface{}) (interface{}, error) {
		var (
			newA any = a
			newB any = b
			err  error
		)
		if reflect.TypeOf(a).String() == "json.Number" {
			newA, err = a.(json.Number).Float64()
			if err != nil {
				return false, err
			}
		}
		if reflect.TypeOf(b).String() == "json.Number" {
			newB, err = b.(json.Number).Float64()
			if err != nil {
				return false, err
			}
		}

		return reflect.DeepEqual(newA, newB), nil
	})
)

func FilterLines(values []map[string]any, filters []string) ([]map[string]any, error) {
	var rows []map[string]any

	var evs gval.Evaluables
	for _, filter := range filters {
		// jsonpath.PlaceholderExtension() is added to gval.Full to add support for '$["date"]' syntax, to
		// be able to select fields that have the same name as a function in gval Full language.
		evaluator, err := gval.Full(jsonpath.PlaceholderExtension(), customEqualFilter).NewEvaluable(filter)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filter %q: %s", filter, err)
		}

		evs = append(evs, evaluator)
	}

	for _, value := range values {
		skipThisRow := false
		for _, ev := range evs {
			res, err := ev.EvalBool(context.Background(), value)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate filter: %s", err)
			}

			if !res {
				skipThisRow = true
				break
			}
		}

		if !skipThisRow {
			rows = append(rows, value)
		}
	}

	return rows, nil
}
