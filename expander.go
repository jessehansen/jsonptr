package jsonptr

import (
	"fmt"
	"strconv"
)

// Expander contains customizable options for how a document can be expanded.
//
// When DetectArrays is true, then resulting nodes with only "0" or positive
// integer keys are made into []interface{} types, rather than
// map[string]interface.
type Expander struct {
	DetectArrays bool
}

// Expand expands a map with keys containing json pointers into a full json.Marshal-able document
func (e *Expander) Expand(values map[string]interface{}) (interface{}, error) {
	result := map[string]interface{}{}
	for k, v := range values {
		if k == "" {
			return nil, fmt.Errorf("Cannot expand when the key is \"\", set directly instead")
		}
		if err := Force(result, k, v); err != nil {
			return nil, err
		}
	}
	if e.DetectArrays {
		return detectArrays(result), nil
	}
	return result, nil
}

func detectArrays(target map[string]interface{}) interface{} {
	allInts := true
	slice := make([]interface{}, len(target))
	for k, v := range target {
		val := v
		if m, ok := v.(map[string]interface{}); ok {
			val = detectArrays(m)
			target[k] = val
		}
		i, err := strconv.Atoi(k)
		if i > len(slice)-1 {
			sl := make([]interface{}, i+1)
			copy(sl, slice)
			slice = sl
		}
		if err != nil || i < 0 {
			allInts = false
		} else {
			slice[i] = val
		}
	}
	if allInts {
		return slice
	}
	return target
}
