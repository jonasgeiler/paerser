// Package env implements encoding and decoding between environment variable and a typed Configuration.
package env

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/traefik/paerser/parser"
)

// DefaultNamePrefix is the default prefix for environment variable names.
const DefaultNamePrefix = "TRAEFIK_"

// Decode decodes the given environment variables into the given element.
// The operation goes through four stages roughly summarized as:
// - env vars -> map
// - map -> tree of untyped nodes
// - untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// - "typed" nodes -> typed element.
func Decode(environ []string, prefix string, element interface{}) error {
	if err := checkPrefix(prefix); err != nil {
		return err
	}

	vars := make(map[string]string)
	for _, evr := range environ {
		k, v, _ := strings.Cut(evr, "=")
		if strings.HasPrefix(strings.ToUpper(k), prefix) {
			key := strings.ReplaceAll(strings.ToLower(k), "_", ".")
			vars[key] = v
		}
	}

	rootName := strings.ToLower(prefix[:len(prefix)-1])
	return parser.Decode(vars, element, rootName)
}

// Encode encodes the configuration in element into the environment variables represented in the returned Flats.
// The operation goes through three stages roughly summarized as:
// - typed configuration in element -> tree of untyped nodes
// - untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// - "typed" nodes -> environment variables with default values (determined by type/kind).
func Encode(prefix string, element interface{}) ([]parser.Flat, error) {
	if err := checkPrefix(prefix); err != nil {
		return nil, err
	}

	rootName := strings.ToLower(prefix[:len(prefix)-1])

	if element == nil {
		return nil, nil
	}

	etnOpts := parser.EncoderToNodeOpts{OmitEmpty: false, TagName: parser.TagLabel, AllowSliceAsStruct: true}
	node, err := parser.EncodeToNode(element, rootName, etnOpts)
	if err != nil {
		return nil, err
	}

	metaOpts := parser.MetadataOpts{TagName: parser.TagLabel, AllowSliceAsStruct: true}
	err = parser.AddMetadata(element, node, metaOpts)
	if err != nil {
		return nil, err
	}

	flatOpts := parser.FlatOpts{Case: "upper", Separator: "_", TagName: parser.TagLabel}
	return parser.EncodeToFlat(element, node, flatOpts)
}

func checkPrefix(prefix string) error {
	prefixPattern := `^[a-zA-Z0-9][a-zA-Z0-9_]*_$`
	matched, err := regexp.MatchString(prefixPattern, prefix)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("invalid prefix %q, the prefix pattern must match the following pattern: %s", prefix, prefixPattern)
	}

	return nil
}
