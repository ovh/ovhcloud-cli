package filters

import (
	"context"
	"fmt"

	"github.com/PaesslerAG/gval"
)

func FilterLines(values []map[string]any, filters []string) ([]map[string]any, error) {
	var rows []map[string]any

	var evs gval.Evaluables
	for _, filter := range filters {
		evaluator, err := gval.Full().NewEvaluable(filter)
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
