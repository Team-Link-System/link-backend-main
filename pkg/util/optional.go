package util

func GetValueOrDefault[T any](value *T, defaultValue T) T {
	if value == nil {
		return defaultValue
	}
	return *value
}

func GetValueOrDefaultSlice[T any](value *[]T, defaultValue []T) []T {
	if value == nil {
		return defaultValue
	}
	return *value
}

// GetFirstOrEmpty 슬라이스의 첫번째 요소를 반환, 없으면 빈 값 반환
func GetFirstOrEmpty[T any](values []T, defaultValue T) T {
	if len(values) == 0 {
		return defaultValue
	}
	return values[0]
}

// ExtractValuesFromMapSlice 맵 슬라이스에서 특정 키의 값들을 추출
func ExtractValuesFromMapSlice[T any](maps []*map[string]interface{}, key string) []T {
	var result []T

	for _, m := range maps {
		if m == nil {
			continue
		}

		if val, exists := (*m)[key]; exists && val != nil {
			if typedVal, ok := val.(T); ok {
				result = append(result, typedVal)
			}
		}
	}

	return result
}

// FilterNonNil nil이 아닌 값들만 필터링
func FilterNonNil[T any](slice []*T) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if item != nil {
			result = append(result, *item)
		}
	}
	return result
}

// MapSlice 슬라이스의 각 요소를 변환 - 다른 타입으로 변환할때
func MapSlice[T any, R any](slice []T, transform func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = transform(item)
	}
	return result
}
