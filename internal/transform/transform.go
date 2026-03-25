// Package transform is a library to generically transform slices or maps
package transform

// SliceToSlice transforms a T1 slice into T2 slice of the same size
func SliceToSlice[T1 any, T2 any](sliceOf []T1, transform func(T1) (to T2)) (dst []T2) {
	dst = make([]T2, len(sliceOf))
	for i, v := range sliceOf {
		dst[i] = transform(v)
	}

	return dst
}

// MapToMap transforms a K:V1 map to a K:V2 map of the same size
func MapToMap[K comparable, V1 any, V2 any](from map[K]V1, transform func(V1) V2) (to map[K]V2) {
	to = make(map[K]V2)

	for k, v := range from {
		to[k] = transform(v)
	}
	return to
}

// ToFilteredSlice a T slice into the subSliceOf T
func ToFilteredSlice[T any](sliceOf []T, meets func(T) bool) (subSliceOf []T) {
	for _, v := range sliceOf {
		if meets(v) {
			subSliceOf = append(subSliceOf, v)
		}
	}

	return subSliceOf
}

// MapToSlice transforms a K:V1 map into a V2 slice of the same size
func MapToSlice[M map[K]V1, K comparable, V1 any, V2 any](fromMap M, transform func(K, V1) V2) (sliceOf []V2) {
	sliceOf = make([]V2, len(fromMap))

	var i int
	for k, v := range fromMap {
		sliceOf[i] = transform(k, v)
		i++
	}

	return sliceOf
}

// GroupSlice groups by K a V slice into a K:[]V map
func GroupSlice[E any, K comparable, V any](slice []E, f func(E) (k K, vv []V)) map[K][]V {
	m := make(map[K][]V)

	for _, v := range slice {
		if k, vv := f(v); vv != nil {
			m[k] = append(m[k], vv...)
		}
	}

	return m
}
