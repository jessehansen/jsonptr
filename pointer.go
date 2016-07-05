package jsonptr

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Pointer struct {
	path []string
}

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

func MustConstruct(ptr string) *Pointer {
	p, err := New(ptr)
	if err != nil {
		panic(err)
	}
	return p
}

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

func (p *Pointer) Set(document *interface{}, val interface{}) error {
	return set(p.path, document, val, false)
}

func (p *Pointer) Force(document *interface{}, val interface{}) error {
	return set(p.path, document, val, true)
}

func set(path []string, document *interface{}, val interface{}, force bool) error {
	node := *document
	if len(path) == 0 {
		*document = val
		return nil
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

func (p *Pointer) Path() []string {
	return p.path
}

func (p *Pointer) String() string {
	segments := make([]string, len(p.path))
	copy(segments, p.path)
	for i, seg := range segments {
		segments[i] = strings.Replace(strings.Replace(seg, "~", "~0", -1), "/", "~1", -1)
	}
	return fmt.Sprintf("/%s", strings.Join(segments, "/"))
}

func (p *Pointer) URIFragmentIdent() string {
	segments := make([]string, len(p.path))
	copy(segments, p.path)
	for i, seg := range segments {
		segments[i] = strings.Replace(strings.Replace(url.QueryEscape(seg), "~", "~0", -1), "/", "~1", -1)
	}
	return fmt.Sprintf("/%s", strings.Join(segments, "/"))
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
