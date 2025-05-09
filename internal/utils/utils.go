package utils

func MergeMaps(left, right map[string]any) map[string]any {
	if left == nil {
		return right
	}

	returnLeft := make(map[string]any)

	for key, rightVal := range right {
		if leftVal, present := left[key]; present {
			switch leftVal := leftVal.(type) {
			case map[string]any:
				if rightMap, ok := rightVal.(map[string]any); ok {
					returnLeft[key] = MergeMaps(leftVal, rightMap)
				}

			case string:
				if leftVal == "" {
					returnLeft[key] = rightVal
				}

				// TODO: handle []any instead
				// case []map[string]any:
				// 	if rightArray, ok := rightVal.([]map[string]any); ok && len(leftVal) == len(rightArray) {
				// 		for idx, val := range leftVal {
				// 			MergeMaps(val, rightArray[idx])
				// 		}
				// 	}

			default:
				returnLeft[key] = leftVal
			}

		} else {
			returnLeft[key] = rightVal
		}
	}

	// Copy remaining keys from left that are not in right
	for key, value := range left {
		if _, ok := returnLeft[key]; !ok {
			returnLeft[key] = value
		}
	}

	return returnLeft
}
