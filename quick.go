package jsonptr

// Get returns the value for the specified location in the document.
func Get(document interface{}, ptr string) (interface{}, error) {
	p, err := New(ptr)
	if err != nil {
		return nil, err
	}
	return p.Get(document)
}

// GetBool returns the value for the specified location in the document as a string, or false if not accessible.
func GetBool(document interface{}, ptr string) bool {
	p, err := New(ptr)
	if err != nil {
		return false
	}
	return p.GetBool(document)
}

// GetString returns the value for the specified location in the document as a string, or an empty string if not accessible.
func GetString(document interface{}, ptr string) string {
	p, err := New(ptr)
	if err != nil {
		return ""
	}
	return p.GetString(document)
}

// GetNumber returns the value for the specified location in the document as a string, or 0 if not accessible.
func GetNumber(document interface{}, ptr string) float64 {
	p, err := New(ptr)
	if err != nil {
		return 0
	}
	return p.GetNumber(document)
}

// Has returns a boolean indicating whether the pointer location exists in
// the provided document.
func Has(document interface{}, ptr string) bool {
	p, err := New(ptr)
	if err != nil {
		return false
	}
	return p.Exists(document)
}

// Set sets the specified location in the document to the provided value.
// See also, Pointer.Set
func Set(document interface{}, ptr string, val interface{}) error {
	p, err := New(ptr)
	if err != nil {
		return err
	}
	return p.Set(document, val)
}

// Force sets the specified location in the document to the provided value
// See also, Pointer.Force
func Force(document interface{}, ptr string, val interface{}) error {
	p, err := New(ptr)
	if err != nil {
		return err
	}
	return p.Force(document, val)
}

/*
Flatten compacts the provided json document into a map[string]interface{},
with all keys at the root level. See also Compactor.Flatten
*/
func Flatten(target interface{}) map[string]interface{} {
	c := &Compactor{}
	return c.Flatten(target)
}

/*
List compacts the provided json document into a slice of PointerValues. See
also Compactor.List
*/
func List(target interface{}) []PointerValue {
	c := &Compactor{}
	return c.List(target)
}

// Expand expands a map with keys containing json pointers into a full
// json.Marshal-able document. See also Expander.Expand
func Expand(values map[string]interface{}) (interface{}, error) {
	e := &Expander{}
	return e.Expand(values)
}
