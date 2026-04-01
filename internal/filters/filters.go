// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
)

var AdditionalEvaluators = []gval.Language{
	// jsonpath.PlaceholderExtension() adds support for '$["date"]' syntax, to be
	// able to select fields that have the same name as a function in gval Full language
	jsonpath.PlaceholderExtension(),

	// additional evaluators to handle json.Number type
	getJsonNumberEvaluator("+", func(a, b float64) (any, error) { return a + b, nil }),
	getJsonNumberEvaluator("-", func(a, b float64) (any, error) { return a - b, nil }),
	getJsonNumberEvaluator("*", func(a, b float64) (any, error) { return a * b, nil }),
	getJsonNumberEvaluator("/", func(a, b float64) (any, error) { return a / b, nil }),
	getJsonNumberEvaluator("%", func(a, b float64) (any, error) { return math.Mod(a, b), nil }),
	getJsonNumberEvaluator("**", func(a, b float64) (any, error) { return math.Pow(a, b), nil }),
	getJsonNumberEvaluator(">", func(a, b float64) (any, error) { return a > b, nil }),
	getJsonNumberEvaluator(">=", func(a, b float64) (any, error) { return a >= b, nil }),
	getJsonNumberEvaluator("<", func(a, b float64) (any, error) { return a < b, nil }),
	getJsonNumberEvaluator("<=", func(a, b float64) (any, error) { return a <= b, nil }),
	getJsonNumberEvaluator("==", func(a, b float64) (any, error) { return a == b, nil }),
	getJsonNumberEvaluator("!=", func(a, b float64) (any, error) { return a != b, nil }),
}

func getJsonNumberEvaluator(operator string, baseEvaluator func(a, b float64) (any, error)) gval.Language {
	return gval.InfixOperator(operator, func(a, b any) (any, error) {
		var (
			floatA, floatB float64
			err            error
		)

		// Handle other types the same way InfixOperators from gval do
		var defaultFunc = func(x, y any) bool { return false }
		if operator == "==" {
			defaultFunc = reflect.DeepEqual
		} else if operator == "!=" {
			defaultFunc = func(x, y any) bool { return !reflect.DeepEqual(x, y) }
		}

		switch a := a.(type) {
		case json.Number:
			floatA, err = a.Float64()
			if err != nil {
				return false, err
			}
		case float64:
			floatA = a
		default:
			return defaultFunc(a, b), nil
		}

		switch b := b.(type) {
		case json.Number:
			floatB, err = b.Float64()
			if err != nil {
				return false, err
			}
		case float64:
			floatB = b
		default:
			return defaultFunc(a, b), nil
		}

		return baseEvaluator(floatA, floatB)
	})
}

func FilterLines(values []map[string]any, filters []string) ([]map[string]any, error) {
	var rows []map[string]any

	var evs gval.Evaluables
	for _, filter := range filters {
		normalizedFilter := normalizeRegexFilters(filter)
		evaluator, err := gval.Full(AdditionalEvaluators...).NewEvaluable(normalizedFilter)
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

var (
	regexFilterMatcherDouble = regexp.MustCompile(`(=~|!~)\s*"(.*?)"`)
	regexFilterMatcherSingle = regexp.MustCompile(`(=~|!~)\s*'(.*?)'`)
)

func normalizeRegexFilters(filter string) string {
	withDouble := regexFilterMatcherDouble.ReplaceAllStringFunc(filter, func(match string) string {
		return normalizeRegexMatch(match, regexFilterMatcherDouble, `"`)
	})

	return regexFilterMatcherSingle.ReplaceAllStringFunc(withDouble, func(match string) string {
		return normalizeRegexMatch(match, regexFilterMatcherSingle, `'`)
	})
}

func normalizeRegexMatch(match string, re *regexp.Regexp, quote string) string {
	parts := re.FindStringSubmatch(match)
	if len(parts) != 3 {
		return match
	}

	operator, pattern := parts[1], parts[2]

	if _, err := regexp.Compile(pattern); err == nil {
		return match
	}

	globRegex := globToRegex(pattern)
	return fmt.Sprintf("%s %s%s%s", operator, quote, globRegex, quote)
}

func globToRegex(pattern string) string {
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, "\\*", ".*")
	escaped = strings.ReplaceAll(escaped, "\\?", ".")
	return escaped
}
