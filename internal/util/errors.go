package util

import "strings"

type MultipleError struct {
	Causes []error
}

func (errs MultipleError) Error() string {
	var str strings.Builder
	str.WriteString("multiple error occur. cause:\n")

	for _, err := range errs.Causes {
		str.WriteString(err.Error() + "\n")
	}

	return str.String()
}

func MergeErrors(errs ...error) error {
	ret := &MultipleError{}
	for _, err := range errs {
		if err == nil {
			continue // skip
		}
		ret.Causes = append(ret.Causes, err)
	}
	if len(ret.Causes) == 0 {
		return nil
	}
	return ret
}
