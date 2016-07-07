package jsonptr

func Get(document interface{}, ptr string) (interface{}, error) {
	p, err := New(ptr)
	if err != nil {
		return nil, err
	}
	return p.Get(document)
}

func Has(document interface{}, ptr string) bool {
	p, err := New(ptr)
	if err != nil {
		return false
	}
	return p.Exists(document)
}

func Set(document interface{}, ptr string, val interface{}) error {
	p, err := New(ptr)
	if err != nil {
		return err
	}
	return p.Set(document, val)
}

func Force(document interface{}, ptr string, val interface{}) error {
	p, err := New(ptr)
	if err != nil {
		return err
	}
	return p.Force(document, val)
}
