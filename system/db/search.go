package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ponzu-cms/ponzu/system/item"

	"encoding/json"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

var (
	// Search tracks all search indices to use throughout system
	Search map[string]bleve.Index

	// ErrNoSearchIndex is for failed checks for an index in Search map
	ErrNoSearchIndex = errors.New("No search index found for type provided")
)

// Searchable ...
type Searchable interface {
	SearchMapping() (*mapping.IndexMappingImpl, error)
	IndexContent() bool
}

func init() {
	Search = make(map[string]bleve.Index)
}

// MapSearchIndex creates the mapping for a type and tracks the index to be used within
// the system for adding/deleting/checking data
func MapSearchIndex(typeName string) error {
	// type assert for Searchable, get configuration (which can be overridden)
	// by Ponzu user if defines own SearchMapping()
	it, ok := item.Types[typeName]
	if !ok {
		return fmt.Errorf("[search] MapSearchIndex Error: Failed to MapIndex for %s, type doesn't exist", typeName)
	}
	s, ok := it().(Searchable)
	if !ok {
		return fmt.Errorf("[search] MapSearchIndex Error: Item type %s doesn't implement db.Searchable", typeName)
	}

	// skip setting or using index for types that shouldn't be indexed
	if !s.IndexContent() {
		return nil
	}

	mapping, err := s.SearchMapping()
	if err == item.ErrNoSearchMapping {
		return nil
	}
	if err != nil {
		return err
	}

	idxName := typeName + ".index"
	var idx bleve.Index

	// check if index exists, use it or create new one
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	searchPath := filepath.Join(pwd, "search")

	err = os.MkdirAll(searchPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	idxPath := filepath.Join(searchPath, idxName)
	if _, err = os.Stat(idxPath); os.IsNotExist(err) {
		idx, err = bleve.New(idxPath, mapping)
		if err != nil {
			return err
		}
		idx.SetName(idxName)
	} else {
		idx, err = bleve.Open(idxPath)
		if err != nil {
			return err
		}
	}

	// add the type name to the index and track the index
	Search[typeName] = idx

	return nil
}

// UpdateSearchIndex sets data into a content type's search index at the given
// identifier
func UpdateSearchIndex(id string, data interface{}) error {
	// check if there is a search index to work with
	target := strings.Split(id, ":")
	ns := target[0]

	idx, ok := Search[ns]
	if ok {
		// unmarshal json to struct, error if not registered
		it, ok := item.Types[ns]
		if !ok {
			return fmt.Errorf("[search] UpdateSearchIndex Error: type '%s' doesn't exist", ns)
		}

		p := it()
		err := json.Unmarshal(data.([]byte), &p)
		if err != nil {
			return err
		}

		// add data to search index
		return idx.Index(id, p)
	}

	return nil
}

// DeleteSearchIndex removes data from a content type's search index at the
// given identifier
func DeleteSearchIndex(id string) error {
	// check if there is a search index to work with
	target := strings.Split(id, ":")
	ns := target[0]

	idx, ok := Search[ns]
	if ok {
		// add data to search index
		return idx.Delete(id)
	}

	return nil
}

// SearchType conducts a search and returns a set of Ponzu "targets", Type:ID pairs,
// and an error. If there is no search index for the typeName (Type) provided,
// db.ErrNoSearchIndex will be returned as the error
func SearchType(typeName, query string) ([]string, error) {
	idx, ok := Search[typeName]
	if !ok {
		return nil, ErrNoSearchIndex
	}

	q := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequest(q)
	res, err := idx.Search(req)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, hit := range res.Hits {
		results = append(results, hit.ID)
	}

	return results, nil
}
