package utils

// 对切片进行去重 - 泛型版本
func RemoveDuplicate[T comparable](slice []T) []T {
	elementMap := make(map[T]struct{})
	for _, item := range slice {
		elementMap[item] = struct{}{}
	}

	result := make([]T, 0, len(elementMap))
	for item := range elementMap {
		result = append(result, item)
	}
	return result
}

func RemoveElement[T comparable](slice []T, value T) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}
