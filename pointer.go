package jsonptr

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Pointer represents a JSON Pointer
type Pointer struct {
	path []string
}

// New returns a new JSON Pointer from the given string. The string can be a pointer, or a URI Fragment encoded pointer.
func New(ptr string) (*Pointer, error) {
	var path []string
	if looksLikeURIFragment(ptr) {
		p, err := decodeURIFragmentIdent(ptr)
		if err != nil {
			return nil, err
		}
		path = p
	} else {
		p, err := decodePointer(ptr)
		if err != nil {
			return nil, err
		}
		path = p
	}
	return &Pointer{path}, nil
}

// MustConstruct returns a new JSON Pointer from the given string, or panics if the pointer is not valid, like regexp.MustCompile.
func MustConstruct(ptr string) *Pointer {
	p, err := New(ptr)
	if err != nil {
		panic(err)
	}
	return p
}

// Get returns the value for the specified location in the document.
func (p *Pointer) Get(document interface{}) (interface{}, error) {
	node := document
	for _, seg := range p.path {
		switch v := node.(type) {
		case map[string]interface{}:
			n, ok := v[seg]
			if !ok {
				return nil, fmt.Errorf("Map had no key when evaluating path segment '%s'", seg)
			}
			node = n
			break
		case []interface{}:
			if seg == "-" {
				return nil, fmt.Errorf("Cannot return '%s' index from JSON array", seg)
			}
			i, err := strconv.Atoi(seg)
			if err != nil {
				return nil, fmt.Errorf("Could not index when evaluating path segment '%s': %v", seg, err)
			}
			if i < 0 || i > len(v)-1 {
				return nil, fmt.Errorf("Slice index %d is out of range (slice len=%d)", i, len(v))
			}
			node = v[i]
			break
		default:
			return nil, fmt.Errorf("Unsupported node type %T when evaluating path segment '%s'", node)
		}
	}
	return node, nil
}

// GetBool returns the value for the specified location in the document as a string, or false if not accessible.
func (p *Pointer) GetBool(document interface{}) bool {
	node, _ := p.Get(document)
	if b, ok := node.(bool); ok {
		return b
	}
	return false
}

// GetString returns the value for the specified location in the document as a string, or an empty string if not accessible.
func (p *Pointer) GetString(document interface{}) string {
	node, _ := p.Get(document)
	if s, ok := node.(string); ok {
		return s
	}
	if node != nil {
		return fmt.Sprintf("%v", node)
	}
	return ""
}

// GetNumber returns the value for the specified location in the document as a string, or 0 if not accessible.
func (p *Pointer) GetNumber(document interface{}) float64 {
	node, _ := p.Get(document)
	if f, ok := node.(float64); ok {
		return f
	}
	return 0
}

/*
Set sets the specified location in the document to the provided value,
returning an error if the value cannot be set. Set requires all segments in
the path to exist except for the final segment, and returns an error if
they do not.

Set cannot set the root pointer ("")

Set will return an error if it encounters a node in the path that is not of the
type map[string]interface{} or []interface{}, or if it cannot index into an
array with the provided path segment.
*/
func (p *Pointer) Set(document interface{}, val interface{}) error {
	return set(p.path, document, val, false)
}

/*
Force sets the specified location in the document to the provided value,
returning an error if the value cannot be set. Force will create new
map[string]interface{} for segments that do not exist in the document

Force cannot set the root pointer ("")

Force will return an error if it encounters a node in the path that is not of the
type map[string]interface{} or []interface{}, or if it cannot index into an
array with the provided path segment.
*/
func (p *Pointer) Force(document interface{}, val interface{}) error {
	return set(p.path, document, val, true)
}

func set(path []string, document interface{}, val interface{}, force bool) error {
	node := document
	if len(path) == 0 {
		return fmt.Errorf("Cannot set root object, set it directly instead")
	}

	for i, seg := range path {
		isLast := i == len(path)-1
		switch v := node.(type) {
		case map[string]interface{}:
			n, ok := v[seg]
			if !ok {
				if !isLast {
					if force {
						n = map[string]interface{}{}
						v[seg] = n
					} else {
						return fmt.Errorf("Map had no key when evaluating path segment '%s'", seg)
					}
				}
			}
			node = n
			if isLast {
				v[seg] = val
				return nil
			}
			break
		case []interface{}:
			if seg == "-" {
				if !isLast {
					if force {
						node = map[string]interface{}{}
						if err := set(path[:i], document, append(v, node), false); err != nil {
							return err
						}
						continue
					} else {
						return fmt.Errorf("Cannot append to JSON array when not forcing")
					}
				} else {
					return set(path[:i], document, append(v, val), false) // set the immediate parent to the appended slice
				}
			}
			i, err := strconv.Atoi(seg)
			if err != nil {
				return fmt.Errorf("Could not index when evaluating path segment '%s': %v", seg, err)
			}
			if i < 0 || (!force && i > len(v)-1) {
				return fmt.Errorf("Slice index %d is out of range (slice len=%d): %v", i, len(v), err)
			}
			if force && i > len(v)-1 {
				sl := make([]interface{}, i+1, i+1)
				copy(sl, v)
				if !isLast {
					v[i] = map[string]interface{}{}
				}
				v = sl
				if err := set(path[:i], document, sl, false); err != nil {
					return err
				}
			}
			if isLast {
				v[i] = val
				return nil
			}
			node = v[i]
			break
		default:
			return fmt.Errorf("Unsupported node type %T when evaluating path segment '%s'", node)
		}
	}
	return fmt.Errorf("Could not set value in path")
}

// Exists returns a boolean indicating whether the pointer location exists in
// the provided document.
func (p *Pointer) Exists(document interface{}) bool {
	node := document
	for _, seg := range p.path {
		switch v := node.(type) {
		case map[string]interface{}:
			n, ok := v[seg]
			if !ok {
				return false
			}
			node = n
			break
		case []interface{}:
			if seg == "-" {
				return false
			}
			i, err := strconv.Atoi(seg)
			if err != nil {
				return false
			}
			if i < 0 || i > len(v)-1 {
				return false
			}
			node = v[i]
			break
		default:
			return false
		}
	}
	return true
}

// Path returns the path segments of the pointer, as a slice of strings.
func (p *Pointer) Path() []string {
	return p.path
}

// String returns the RFC 6901 string representation of the JSON pointer
func (p *Pointer) String() string {
	if len(p.path) == 0 {
		return ""
	}
	segments := make([]string, len(p.path))
	copy(segments, p.path)
	for i, seg := range segments {
		segments[i] = strings.Replace(strings.Replace(seg, "~", "~0", -1), "/", "~1", -1)
	}
	return fmt.Sprintf("/%s", strings.Join(segments, "/"))
}

// URIFragmentIdent returns the RFC 6901 URI Fragment representation of the JSON pointer
func (p *Pointer) URIFragmentIdent() string {
	if len(p.path) == 0 {
		return "#"
	}
	segments := make([]string, len(p.path))
	copy(segments, p.path)
	for i, seg := range segments {
		segments[i] = url.QueryEscape(strings.Replace(strings.Replace(seg, "~", "~0", -1), "/", "~1", -1))
	}
	return fmt.Sprintf("#/%s", strings.Join(segments, "/"))
}

func looksLikeURIFragment(ptr string) bool {
	return strings.HasPrefix(ptr, "#")
}

func unescape(str string) string {
	res, _ := url.QueryUnescape(str)
	if res == "" {
		return str
	}
	return res
}

func decodeURIFragmentIdent(ptr string) ([]string, error) {
	if len(ptr) == 1 {
		return []string{}, nil
	}
	if !strings.HasPrefix(ptr, "#/") {
		return nil, fmt.Errorf("Invalid JSON Pointer syntax")
	}
	segments := strings.Split(ptr, "/")
	result := make([]string, len(segments)-1)
	for i, seg := range segments[1:] {
		result[i] = strings.Replace(strings.Replace(unescape(seg), "~1", "/", -1), "~0", "~", -1)
	}
	return result, nil
}

func decodePointer(ptr string) ([]string, error) {
	if len(ptr) == 0 {
		return []string{}, nil
	}
	if !strings.HasPrefix(ptr, "/") {
		return nil, fmt.Errorf("Invalid JSON Pointer syntax")
	}
	segments := strings.Split(ptr, "/")
	result := make([]string, len(segments)-1)
	for i, seg := range segments[1:] {
		result[i] = strings.Replace(strings.Replace(seg, "~1", "/", -1), "~0", "~", -1)
	}
	return result, nil
}
