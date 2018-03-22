package tools

// custom strings concatenator to avoid separator artefacts of empty params
func Join(a []string, sep string) string {
	result := make([]byte, 0)

	for i, _ := range a {
		if len(a[i]) == 0 {
			continue
		}

		if len(result) > 0 && a[i] != "" {
			result = append(result, []byte(sep)...)
		}

		result = append(result, []byte(a[i])...)
	}

	return string(result)
}

type Slice interface {
	Contains(string) bool
}

type StringSlice []string

func (ss StringSlice) Contains(s string) bool {
	for i, _ := range ss {
		if ss[i] == s {
			return true
		}
	}

	return false
}
