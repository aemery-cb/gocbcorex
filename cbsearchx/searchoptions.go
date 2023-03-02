package cbsearchx

// SearchHighlightStyle indicates the type of highlighting to use for a search query.
type SearchHighlightStyle string

const (
	// DefaultHighlightStyle specifies to use the default to highlight search result hits.
	DefaultHighlightStyle SearchHighlightStyle = ""

	// HTMLHighlightStyle specifies to use HTML tags to highlight search result hits.
	HTMLHighlightStyle SearchHighlightStyle = "html"

	// AnsiHightlightStyle specifies to use ANSI tags to highlight search result hits.
	AnsiHightlightStyle SearchHighlightStyle = "ansi"
)

// SearchScanConsistency indicates the level of data consistency desired for a search query.
type SearchScanConsistency uint

const (
	searchScanConsistencyNotSet SearchScanConsistency = iota

	// SearchScanConsistencyNotBounded indicates no data consistency is required.
	SearchScanConsistencyNotBounded
)

// SearchHighlightOptions are the options available for search highlighting.
type SearchHighlightOptions struct {
	Style  SearchHighlightStyle
	Fields []string
}

// SearchOptions represents a pending search query.
type SearchOptions struct {
	ScanConsistency SearchScanConsistency
	Limit           uint32
	Skip            uint32
	Explain         bool
	Highlight       *SearchHighlightOptions
	Fields          []string
	Sort            []Sort
	Facets          map[string]Facet
	ConsistentWith  map[string]map[string][]uint64

	// Raw provides a way to provide extra parameters in the request body for the query.
	Raw map[string]interface{}

	DisableScoring bool

	Collections []string

	// If set to true, will include the SearchRowLocations.
	IncludeLocations bool

	// Endpoint overrides internal routing of requests to send a request directly to an endpoint.
	Endpoint string

	OnBehalfOf string
}

func (opts *SearchOptions) toMap(indexName string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	if opts.Limit > 0 {
		data["size"] = opts.Limit
	}

	if opts.Skip > 0 {
		data["from"] = opts.Skip
	}

	if opts.Explain {
		data["explain"] = opts.Explain
	}

	if len(opts.Fields) > 0 {
		data["fields"] = opts.Fields
	}

	if len(opts.Sort) > 0 {
		data["sort"] = opts.Sort
	}

	if opts.Highlight != nil {
		highlight := make(map[string]interface{})
		highlight["style"] = string(opts.Highlight.Style)
		highlight["fields"] = opts.Highlight.Fields
		data["highlight"] = highlight
	}

	if opts.Facets != nil {
		facets := make(map[string]interface{})
		for k, v := range opts.Facets {
			facets[k] = v
		}
		data["facets"] = facets
	}

	if opts.ScanConsistency != 0 && opts.ConsistentWith != nil {
		return nil, InvalidArgumentError{"ScanConsistency and ConsistentWith must be used exclusively"}
	}

	var ctl map[string]interface{}

	if opts.ScanConsistency != searchScanConsistencyNotSet {
		consistency := make(map[string]interface{})

		if opts.ScanConsistency == SearchScanConsistencyNotBounded {
			consistency["level"] = ""
		} else {
			return nil, InvalidArgumentError{"unexpected consistency option"}
		}

		ctl = map[string]interface{}{"consistency": consistency}
	}

	if opts.ConsistentWith != nil {
		consistency := make(map[string]interface{})

		consistency["level"] = "at_plus"
		consistency["vectors"] = opts.ConsistentWith

		if ctl == nil {
			ctl = make(map[string]interface{})
		}
		ctl["consistency"] = consistency
	}

	if ctl != nil {
		data["ctl"] = ctl
	}

	if opts.DisableScoring {
		data["score"] = "none"
	}

	if opts.Raw != nil {
		for k, v := range opts.Raw {
			data[k] = v
		}
	}

	if len(opts.Collections) > 0 {
		data["collections"] = opts.Collections
	}

	if opts.IncludeLocations {
		data["includeLocations"] = true
	}

	return data, nil
}
