package jsonptr

import (
	"strconv"
)

// Compactor contains customizable options for how a document can be compacted.
//
// When AllNodes is false, then only the leaf nodes of the document are
// returned true, then all nodes are returned in the result.
//
// When URIFragment is true, then the keys in Flatten's resulting
// map[string]interface{} are RFC 6901 URL Fragment Identifiers.
type Compactor struct {
	AllNodes, URIFragment bool
}

// PointerValue represents a pointer and it's value. A slice of pointer values
// is returned from Compactor.List
type PointerValue struct {
	Pointer Pointer
	Value   interface{}
}

/*
Flatten compacts the provided json document into a map[string]interface{},
with all keys at the root level.

    // Given doc is unmarshalled from { "foo": { "bar": ["baz"] } }
    c := &jsonptr.Compactor{AllNodes: false, URIFragment: false}
    res := c.Flatten(doc)
    // res would be { "/foo/bar/0": "baz" }
*/
func (c *Compactor) Flatten(document interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	var v visitor
	if c.URIFragment {
		v = func(path []string, val interface{}) {
			p := Pointer{path}
			res[p.URIFragmentIdent()] = val
		}
	} else {
		v = func(path []string, val interface{}) {
			p := Pointer{path}
			res[p.String()] = val
		}
	}
	c.visit(document, v)
	return res
}

/*
List compacts the provided json document into a slice of PointerValues

    // Given doc is unmarshalled from { "foo": { "bar": ["baz"] } }
    c := &jsonptr.Compactor{AllNodes: false, URIFragment: false}
    res := c.List(doc)
    // len(res) == 1
    // res[0].Pointer.String() == "/foo/bar/0"
    // res[0].Value == "baz"
*/
func (c *Compactor) List(document interface{}) []PointerValue {
	res := make([]PointerValue, 0, 8)
	c.visit(document, func(path []string, val interface{}) {
		res = append(res, PointerValue{Pointer{path}, val})
	})
	return res
}

type visitor func([]string, interface{})

type qnode struct {
	obj  interface{}
	path []string
}

func (c *Compactor) visit(target interface{}, visit visitor) {
	q := make([]qnode, 1, 10)
	q[0] = qnode{target, []string{}}
	qcursor := 0

	if c.AllNodes || !canVisitChildren(target) {
		visit([]string{}, target)
	}
	for qcursor < len(q) {
		cursor := q[qcursor]
		qcursor++
		switch v := cursor.obj.(type) {
		case []interface{}:
			for j, it := range v {
				path := childpath(cursor.path, strconv.Itoa(j))
				if canVisitChildren(it) {
					q = append(q, qnode{it, path})
					if !c.AllNodes {
						continue
					}
				}
				visit(path, it)
			}
		case map[string]interface{}:
			for key, it := range v {
				path := childpath(cursor.path, key)
				if canVisitChildren(it) {
					q = append(q, qnode{it, path})
					if !c.AllNodes {
						continue
					}
				}
				visit(path, it)
			}
		}
	}
}

func canVisitChildren(obj interface{}) bool {
	if obj == nil {
		return false
	}
	if _, ok := obj.(map[string]interface{}); ok {
		return true
	}
	if _, ok := obj.([]interface{}); ok {
		return true
	}
	return false
}

func childpath(parent []string, n string) []string {
	res := make([]string, len(parent)+1)
	copy(res, parent)
	res[len(parent)] = n
	return res
}
