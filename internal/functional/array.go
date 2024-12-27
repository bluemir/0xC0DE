package functional

func Map[In any, Out any](in []In, fn func(In) Out) []Out {
	out := []Out{}

	for _, v := range in {
		out = append(out, fn(v))
	}

	return out
}
func Filter[T any](in []T, fn func(T) bool) []T {
	out := []T{}

	for _, v := range in {
		if fn(v) {
			out = append(out, v)
		}
	}

	return out
}
func ToLookupTable[KeyType comparable, ElemType any](in []ElemType, keyFn func(ElemType) KeyType) map[KeyType]ElemType {
	m := map[KeyType]ElemType{}

	for _, elem := range in {
		key := keyFn(elem)
		m[key] = elem
	}

	return m
}
func Some[T any](in []T, fn func(T) bool) bool {
	for _, v := range in {
		if fn(v) {
			return true
		}
	}
	return false
}
func All[T any](in []T, fn func(T) bool) bool {
	for _, v := range in {
		if !fn(v) {
			return false
		}
	}
	return true
}
func Reduce[In any, Out any](in []In, fn func(accumulator *Out, v In) Out, init Out) Out {
	out := init

	for _, v := range in {
		fn(&out, v)
	}

	return out
}
func Flat[T any](in [][]T) []T {
	out := []T{}

	for _, arr := range in {
		out = append(out, arr...)
	}
	return out
}

// arr = Filter(arr, func(v int) bool)
// arr = Map(arr, func(v int) float)
// ---- vs ----
// functional.From(arr).Filter(func(v int) bool).Map(func(v int) float).ToArray()
