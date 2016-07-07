package jsonptr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testValues = map[string]interface{}{
	"/skey":    "value",
	"/ikey":    10,
	"/obj/0":   "baz",
	"/obj/foo": "bar",
	"/arr/0":   "0th",
	"/arr/1":   "1st",
	"/arr/2":   "2nd",
	"/arr/3":   "3rd",
	"/arr/5":   "10th", // intentionally skip 4
}

func TestExpand(t *testing.T) {
	result, err := Expand(testValues)
	assert.Nil(t, err)
	m, ok := result.(map[string]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 4, len(m))
	assert.Equal(t, testValues["/skey"], m["skey"])
	assert.Equal(t, testValues["/ikey"], m["ikey"])

	obj, ok := m["obj"].(map[string]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 2, len(obj))
	assert.Equal(t, testValues["/obj/0"], obj["0"])
	assert.Equal(t, testValues["/obj/foo"], obj["foo"])

	arr, ok := m["arr"].(map[string]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 5, len(arr))
	assert.Equal(t, testValues["/arr/0"], arr["0"])
	assert.Equal(t, testValues["/arr/1"], arr["1"])
	assert.Equal(t, testValues["/arr/2"], arr["2"])
	assert.Equal(t, testValues["/arr/3"], arr["3"])
	assert.Equal(t, testValues["/arr/5"], arr["5"])
}

func TestExpandDetectArrays(t *testing.T) {
	e := &Expander{DetectArrays: true}
	result, err := e.Expand(testValues)
	assert.Nil(t, err)
	m, ok := result.(map[string]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 4, len(m))
	assert.Equal(t, testValues["/skey"], m["skey"])
	assert.Equal(t, testValues["/ikey"], m["ikey"])

	obj, ok := m["obj"].(map[string]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 2, len(obj))
	assert.Equal(t, testValues["/obj/0"], obj["0"])
	assert.Equal(t, testValues["/obj/foo"], obj["foo"])

	arr, ok := m["arr"].([]interface{})
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 6, len(arr))
	assert.Equal(t, testValues["/arr/0"], arr[0])
	assert.Equal(t, testValues["/arr/1"], arr[1])
	assert.Equal(t, testValues["/arr/2"], arr[2])
	assert.Equal(t, testValues["/arr/3"], arr[3])
	assert.Nil(t, arr[4])
	assert.Equal(t, testValues["/arr/5"], arr[5])
}

func BenchmarkExpandDefaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Expand(testValues)
	}
}

func BenchmarkExpandDetectArrays(b *testing.B) {
	e := &Expander{DetectArrays: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Expand(testValues)
	}
}
