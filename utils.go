package jsonptr

// reference-equality set
type refset map[*interface{}]Pointer

func (r *refset) has(key interface{}) bool {
	_, ok := (*r)[&key]
	return ok
}
func (r *refset) get(key interface{}) Pointer {
	return (*r)[&key]
}

func (r *refset) set(key interface{}, val Pointer) {
	(*r)[&key] = val
}
