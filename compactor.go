package jsonptr

import (
	"strconv"
)

type Compactor struct {
	CheckCycles, OmitNonLeaf, URIFragment bool
}

type PointerValue struct {
	Pointer Pointer
	Value   interface{}
}

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

func (c *Compactor) List(document interface{}) []PointerValue {
	res := make([]PointerValue, 0, 8)
	c.visit(document, func(path []string, val interface{}) {
		res = append(res, PointerValue{Pointer{path}, val})
	})
	return res
}

func Flatten(target interface{}) map[string]interface{} {
	c := &Compactor{}
	return c.Flatten(target)
}

func List(target interface{}) []PointerValue {
	c := &Compactor{}
	return c.List(target)
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

	var distinctObjects *refset
	if c.CheckCycles {
		distinctObjects = &refset{}
	}
	if !c.OmitNonLeaf || !canVisitChildren(target) {
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
					if c.CheckCycles && distinctObjects.has(it) {
						if !c.OmitNonLeaf {
							visit(path, JsonReference{distinctObjects.get(it)})
						}
						continue
					}
					q = append(q, qnode{it, path})
					if c.CheckCycles {
						distinctObjects.set(it, Pointer{path})
					}
					if c.OmitNonLeaf {
						continue
					}
				}
				visit(path, it)
			}
		case map[string]interface{}:
			for key, it := range v {
				path := childpath(cursor.path, key)
				if canVisitChildren(it) {
					if c.CheckCycles && distinctObjects.has(it) {
						if !c.OmitNonLeaf {
							visit(path, JsonReference{distinctObjects.get(it)})
						}
						continue
					}
					q = append(q, qnode{it, path})
					if c.CheckCycles {
						distinctObjects.set(it, Pointer{path})
					}
					if c.OmitNonLeaf {
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
