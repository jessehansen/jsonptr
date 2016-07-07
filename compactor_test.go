package jsonptr

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

const SampleDoc = `{
  "legumes": [{
    "name": "pinto beans",
    "unit": "lbs",
    "instock": 4
  },{
    "name": "lima beans",
    "unit": "lbs",
    "instock": 21
  },{
    "name": "black eyed peas",
    "unit": "lbs",
    "instock": 13
  },{
    "name": "split peas",
    "unit": "lbs",
    "instock": 8
  }]
}`

var pointerKeys = []string{
	"", "/legumes",
	"/legumes/0", "/legumes/0/name", "/legumes/0/unit", "/legumes/0/instock",
	"/legumes/1", "/legumes/1/name", "/legumes/1/unit", "/legumes/1/instock",
	"/legumes/2", "/legumes/2/name", "/legumes/2/unit", "/legumes/2/instock",
	"/legumes/3", "/legumes/3/name", "/legumes/3/unit", "/legumes/3/instock",
}

var pointerLeafKeys = []string{
	"/legumes/0/name", "/legumes/0/unit", "/legumes/0/instock",
	"/legumes/1/name", "/legumes/1/unit", "/legumes/1/instock",
	"/legumes/2/name", "/legumes/2/unit", "/legumes/2/instock",
	"/legumes/3/name", "/legumes/3/unit", "/legumes/3/instock",
}

var fragmentKeys = []string{
	"#", "#/legumes",
	"#/legumes/0", "#/legumes/0/name", "#/legumes/0/unit", "#/legumes/0/instock",
	"#/legumes/1", "#/legumes/1/name", "#/legumes/1/unit", "#/legumes/1/instock",
	"#/legumes/2", "#/legumes/2/name", "#/legumes/2/unit", "#/legumes/2/instock",
	"#/legumes/3", "#/legumes/3/name", "#/legumes/3/unit", "#/legumes/3/instock",
}

var fragmentLeafKeys = []string{
	"#/legumes/0/name", "#/legumes/0/unit", "#/legumes/0/instock",
	"#/legumes/1/name", "#/legumes/1/unit", "#/legumes/1/instock",
	"#/legumes/2/name", "#/legumes/2/unit", "#/legumes/2/instock",
	"#/legumes/3/name", "#/legumes/3/unit", "#/legumes/3/instock",
}

func TestFlatten(t *testing.T) {
	assertFlattenProducesKeys(t, &Compactor{OmitNonLeaf: false, URIFragment: false}, pointerKeys)
	assertFlattenProducesKeys(t, &Compactor{OmitNonLeaf: true, URIFragment: false}, pointerLeafKeys)
	assertFlattenProducesKeys(t, &Compactor{OmitNonLeaf: false, URIFragment: true}, fragmentKeys)
	assertFlattenProducesKeys(t, &Compactor{OmitNonLeaf: true, URIFragment: true}, fragmentLeafKeys)
}

func assertFlattenProducesKeys(t *testing.T, c *Compactor, keys []string) {
	var sampleDocument map[string]interface{}
	json.Unmarshal([]byte(SampleDoc), &sampleDocument)

	result := c.Flatten(sampleDocument)

	assert.Equal(t, len(keys), len(result), "key count")
	for _, k := range keys {
		_, ok := result[k]
		assert.True(t, ok, "Expected: %s", k)
	}
}
