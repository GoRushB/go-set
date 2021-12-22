package utils

type Any = interface{}

type AnyList []Any

func (list AnyList) AsInt() []int {
	res := make([]int, len(list))
	for i := range list {
		res[i] = AsInt(list[i])
	}
	return res
}

func (list AnyList) AsInt64() []int64 {
	res := make([]int64, len(list))
	for i := range list {
		res[i] = AsInt64(list[i])
	}
	return res
}

func (list AnyList) AsString() []string {
	res := make([]string, len(list))
	for i := range list {
		res[i] = AsString(list[i])
	}
	return res
}
