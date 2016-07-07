package jsonptr

type JsonReference struct {
	pointer Pointer
}

func (r *JsonReference) String() string {
	return r.pointer.URIFragmentIdent()
}

func (r *JsonReference) Resolve(target interface{}) (interface{}, error) {
	return r.pointer.Get(target)
}
