package jsonptr

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

const (
	RfcDoc = `{
  "foo": ["bar", "baz"],
  "": 0,
  "a/b": 1,
  "c%d": 2,
  "e^f": 3,
  "g|h": 4,
  "i\\j": 5,
  "k\"l": 6,
  " ": 7,
  "m~n": 8
}`
	DeepDoc = `{
  "foo": {
    "bar": {
      "baz": ["100"]
    }
  },
  "item": "value"
}`
	TypeDoc = `{
	"string": "string",
	"number": 1.2,
	"bool": true,
	"node": { "name": "val" }
}`
)

func makePointer(path []string) *Pointer {
	return &Pointer{path}
}

func TestString(t *testing.T) {
	assert.Equal(t, "", makePointer([]string{}).String())
	assert.Equal(t, "/", makePointer([]string{""}).String())
	assert.Equal(t, "//", makePointer([]string{"", ""}).String())
	assert.Equal(t, "/foo/bar", makePointer([]string{"foo", "bar"}).String())
	assert.Equal(t, "/m~0n", makePointer([]string{"m~n"}).String())
	assert.Equal(t, "/~01", makePointer([]string{"~1"}).String())
	assert.Equal(t, "/~1", makePointer([]string{"/"}).String())
}

func TestURIFragmentIdent(t *testing.T) {
	assert.Equal(t, "#", makePointer([]string{}).URIFragmentIdent())
	assert.Equal(t, "#/", makePointer([]string{""}).URIFragmentIdent())
	assert.Equal(t, "#//", makePointer([]string{"", ""}).URIFragmentIdent())
	assert.Equal(t, "#/foo/bar", makePointer([]string{"foo", "bar"}).URIFragmentIdent())
	assert.Equal(t, "#/m~0n", makePointer([]string{"m~n"}).URIFragmentIdent())
	assert.Equal(t, "#/~01", makePointer([]string{"~1"}).URIFragmentIdent())
	assert.Equal(t, "#/~1", makePointer([]string{"/"}).URIFragmentIdent())
	assert.Equal(t, "#/with+space", makePointer([]string{"with space"}).URIFragmentIdent())
	assert.Equal(t, "#/with%5Ecarat", makePointer([]string{"with^carat"}).URIFragmentIdent())
}

// From RFC 6901:
// The following JSON strings evaluate to the accompanying values:
//  ""           // the whole document
//  "/foo"       ["bar", "baz"]
//  "/foo/0"     "bar"
//  "/"          0
//  "/a~1b"      1
//  "/c%d"       2
//  "/e^f"       3
//  "/g|h"       4
//  "/i\\j"      5
//  "/k\"l"      6
//  "/ "         7
//  "/m~0n"      8
func TestGetRfcPointerCases(t *testing.T) {
	var rfcDocument map[string]interface{}
	json.Unmarshal([]byte(RfcDoc), &rfcDocument)

	assertPointerEvaluatesTo(t, "", rfcDocument, rfcDocument)
	assertPointerEvaluatesTo(t, "/foo", rfcDocument, rfcDocument["foo"])
	assertPointerEvaluatesTo(t, "/foo/0", rfcDocument, "bar")
	assertPointerEvaluatesTo(t, "/", rfcDocument, 0)
	assertPointerEvaluatesTo(t, "/a~1b", rfcDocument, 1)
	assertPointerEvaluatesTo(t, "/c%d", rfcDocument, 2)
	assertPointerEvaluatesTo(t, "/e^f", rfcDocument, 3)
	assertPointerEvaluatesTo(t, "/g|h", rfcDocument, 4)
	assertPointerEvaluatesTo(t, `/i\j`, rfcDocument, 5)
	assertPointerEvaluatesTo(t, `/k"l`, rfcDocument, 6)
	assertPointerEvaluatesTo(t, "/ ", rfcDocument, 7)
	assertPointerEvaluatesTo(t, "/m~0n", rfcDocument, 8)
}

func TestGetDeepPointerCases(t *testing.T) {
	var deepDocument interface{}
	json.Unmarshal([]byte(DeepDoc), &deepDocument)

	assertPointerEvaluatesTo(t, "/foo/bar/baz/0", deepDocument, "100")
}

// From RFC 6901:
// Given the same example document as above, the following URI fragment
// identifiers evaluate to the accompanying values:
//  #            // the whole document
//  #/foo        ["bar", "baz"]
//  #/foo/0      "bar"
//  #/           0
//  #/a~1b       1
//  #/c%25d      2
//  #/e%5Ef      3
//  #/g%7Ch      4
//  #/i%5Cj      5
//  #/k%22l      6
//  #/%20        7
//  #/m~0n       8
func TestGetRfcFragmentCases(t *testing.T) {
	var rfcDocument map[string]interface{}
	json.Unmarshal([]byte(RfcDoc), &rfcDocument)

	assertPointerEvaluatesTo(t, "#", rfcDocument, rfcDocument)
	assertPointerEvaluatesTo(t, "#/foo", rfcDocument, rfcDocument["foo"])
	assertPointerEvaluatesTo(t, "#/foo/0", rfcDocument, "bar")
	assertPointerEvaluatesTo(t, "#/", rfcDocument, 0)
	assertPointerEvaluatesTo(t, "#/a~1b", rfcDocument, 1)
	assertPointerEvaluatesTo(t, "#/c%25d", rfcDocument, 2)
	assertPointerEvaluatesTo(t, "#/e%5Ef", rfcDocument, 3)
	assertPointerEvaluatesTo(t, "#/g%7Ch", rfcDocument, 4)
	assertPointerEvaluatesTo(t, "#/i%5Cj", rfcDocument, 5)
	assertPointerEvaluatesTo(t, "#/k%22l", rfcDocument, 6)
	assertPointerEvaluatesTo(t, "#/%20", rfcDocument, 7)
	assertPointerEvaluatesTo(t, "#/m~0n", rfcDocument, 8)
}

func TestSetEscaping(t *testing.T) {
	assertSetWorks(t, "#/a~1b")
	assertSetWorks(t, "#/c%25d")
	assertSetWorks(t, "#/e%5Ef")
	assertSetWorks(t, "#/g%7Ch")
	assertSetWorks(t, "#/i%5Cj")
	assertSetWorks(t, "#/k%22l")
	assertSetWorks(t, "#/%20")
	assertSetWorks(t, "#/m~0n")
}

func TestSetArrayMember(t *testing.T) {
	res := doSet(t, "/foo/bar/baz/0", DeepDoc, "200")
	n1 := res.(map[string]interface{})
	n2 := n1["foo"].(map[string]interface{})
	n3 := n2["bar"].(map[string]interface{})
	n4 := n3["baz"].([]interface{})
	assert.Equal(t, 1, len(n4))
	assert.Equal(t, "200", n4[0])
}

func TestSetDashArray(t *testing.T) {
	res := doSet(t, "/foo/bar/baz/-", DeepDoc, "300")
	n1 := res.(map[string]interface{})
	n2 := n1["foo"].(map[string]interface{})
	n3 := n2["bar"].(map[string]interface{})
	n4 := n3["baz"].([]interface{})
	assert.Equal(t, 2, len(n4))
	assert.Equal(t, "300", n4[1])
}

func TestForceWithArrayInTheMiddle(t *testing.T) {
	res := doForce(t, "/foo/-/bar/baz", `{"foo":[]}`, "value")
	n1 := res.(map[string]interface{})
	n2 := n1["foo"].([]interface{})
	assert.Equal(t, 1, len(n2))
	n3 := n2[0].(map[string]interface{})
	n4 := n3["bar"].(map[string]interface{})
	assert.Equal(t, "value", n4["baz"])
}

func TestGetBool(t *testing.T) {
	doc := getDocWithTypes()
	p := MustConstruct("/bool")
	assert.Equal(t, true, p.GetBool(doc))
	p = MustConstruct("/string")
	assert.Equal(t, false, p.GetBool(doc))
	p = MustConstruct("/node")
	assert.Equal(t, false, p.GetBool(doc))
	p = MustConstruct("/not/there")
	assert.Equal(t, false, p.GetBool(doc))
}

func TestGetString(t *testing.T) {
	doc := getDocWithTypes()
	p := MustConstruct("/string")
	assert.Equal(t, "string", p.GetString(doc))
	p = MustConstruct("/bool")
	assert.Equal(t, "true", p.GetString(doc))
	p = MustConstruct("/number")
	assert.Equal(t, "1.2", p.GetString(doc))
	p = MustConstruct("/node")
	assert.Equal(t, "map[name:val]", p.GetString(doc))
	p = MustConstruct("/not/there")
	assert.Equal(t, "", p.GetString(doc))
}

func TestGetNumber(t *testing.T) {
	doc := getDocWithTypes()
	p := MustConstruct("/number")
	assert.Equal(t, 1.2, p.GetNumber(doc))
	p = MustConstruct("/bool")
	assert.Equal(t, 0.0, p.GetNumber(doc))
	p = MustConstruct("/node")
	assert.Equal(t, 0.0, p.GetNumber(doc))
	p = MustConstruct("/not/there")
	assert.Equal(t, 0.0, p.GetNumber(doc))
}

func ExamplePointer_Get() {
	var doc interface{}
	json.Unmarshal([]byte(`{"hello":"world"}`), &doc)

	ptr1, _ := New("/hello")
	hello, err := ptr1.Get(doc)

	fmt.Printf("%v %s", err == nil, hello)
	// Output: true world
}

func ExamplePointer_Set() {
	var doc interface{}
	json.Unmarshal([]byte(`{"hello":"world"}`), &doc)

	ptr1, _ := New("/foo/bar")
	e1 := ptr1.Set(doc, "baz")

	ptr2, _ := New("/foo")
	e2 := ptr2.Set(doc, "bar")

	out, _ := json.Marshal(doc)
	fmt.Printf("%v %v %s", e1 == nil, e2 == nil, string(out))
	// Output: false true {"foo":"bar","hello":"world"}
}

func ExamplePointer_Force() {
	var doc interface{}
	json.Unmarshal([]byte(`{"hello":"world"}`), &doc)

	ptr, _ := New("/foo/bar")
	err := ptr.Force(doc, "baz")

	out, _ := json.Marshal(doc)
	fmt.Printf("%v %s", err == nil, string(out))
	// Output: true {"foo":{"bar":"baz"},"hello":"world"}
}

func getZips() interface{} {
	zips, err := ioutil.ReadFile("zips.json")
	if err != nil {
		panic(err)
	}
	var doc interface{}
	err = json.Unmarshal(zips, &doc)
	if err != nil {
		panic(err)
	}
	return doc
}

func BenchmarkShallowGet(b *testing.B) {
	doc := getZips()
	ptr := MustConstruct("/item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr.Get(doc)
	}
}

func BenchmarkDeepGet(b *testing.B) {
	doc := getZips()
	ptr := MustConstruct("/foo/bar/baz/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr.Get(doc)
	}
}

func BenchmarkShallowExists(b *testing.B) {
	doc := getZips()
	ptr1 := MustConstruct("/item")
	ptr2 := MustConstruct("/missing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr1.Exists(doc)
		ptr2.Exists(doc)
	}
}

func BenchmarkDeepExists(b *testing.B) {
	doc := getZips()
	ptr1 := MustConstruct("/foo/bar/baz/0")
	ptr2 := MustConstruct("/foo/bar/baz/7")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr1.Exists(doc)
		ptr2.Exists(doc)
	}
}

func BenchmarkShallowSet(b *testing.B) {
	doc := getZips()
	ptr := MustConstruct("/item")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr.Set(doc, "someval")
	}
}

func BenchmarkDeepSet(b *testing.B) {
	doc := getZips()
	ptr := MustConstruct("/foo/bar/baz/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr.Set(doc, "someval")
	}
}

func BenchmarkShallowForce(b *testing.B) {
	doc := getZips()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr := MustConstruct(fmt.Sprintf("/item%d", i))
		ptr.Set(doc, "someval")
	}
}

func BenchmarkDeepForce(b *testing.B) {
	doc := getZips()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr := MustConstruct(fmt.Sprintf("/foo%d/bar%d/item%d", i, i, i))
		ptr.Set(doc, "someval")
	}
}

func getDocWithTypes() interface{} {
	var doc map[string]interface{}
	json.Unmarshal([]byte(TypeDoc), &doc)
	return doc
}

func assertPointerEvaluatesTo(t *testing.T, pointer string, doc interface{}, expected interface{}) {
	ptr, err := New(pointer)
	assert.Nil(t, err, "Pointer construction error")
	assert.NotNil(t, ptr, "Pointer construction returned no instance")
	if ptr == nil {
		return
	}
	actual, err := ptr.Get(doc)
	assert.Nil(t, err, "Pointer evaluation error")

	assert.EqualValues(t, expected, actual, "Pointer: %s", pointer)
}

func assertSetWorks(t *testing.T, pointer string) {
	ptr, err := New(pointer)
	assert.Nil(t, err, "Pointer construction error")
	assert.NotNil(t, ptr, "Pointer construction returned no instance")
	if ptr == nil {
		return
	}

	var doc interface{}
	json.Unmarshal([]byte(RfcDoc), &doc)

	err = ptr.Set(doc, "some value")
	assert.Nil(t, err, "Pointer evaluation error")

	actual, _ := ptr.Get(doc)

	assert.EqualValues(t, "some value", actual, "Pointer: %s", pointer)
}

func doSet(t *testing.T, pointer string, str string, val interface{}) interface{} {
	var doc interface{}
	json.Unmarshal([]byte(str), &doc)

	ptr, err := New(pointer)
	assert.Nil(t, err, "Pointer construction error")
	assert.NotNil(t, ptr, "Pointer construction returned no instance")
	if ptr == nil {
		return nil
	}
	err = ptr.Set(doc, val)
	assert.Nil(t, err, "Pointer evaluation error")
	return doc
}

func doForce(t *testing.T, pointer string, str string, val interface{}) interface{} {
	var doc interface{}
	json.Unmarshal([]byte(str), &doc)

	ptr, err := New(pointer)
	assert.Nil(t, err, "Pointer construction error")
	assert.NotNil(t, ptr, "Pointer construction returned no instance")
	if ptr == nil {
		return nil
	}
	err = ptr.Force(doc, val)
	assert.Nil(t, err, "Pointer evaluation error")
	return doc
}
