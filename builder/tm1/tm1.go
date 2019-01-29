package tm1

import (
	"bytes"
	"encoding/json"
)

// Dimension defines the structure of a single Dimension entity in the TM1 Server schema
type Dimension struct {
	Name        string
	Hierarchies []*Hierarchy
}

// Hierarchy defines the structure of a single Hierarchy entity in the TM1 Server schema
type Hierarchy struct {
	Name     string
	Elements []Element
	Edges    []Edge            `json:",omitempty"`
	Captions map[string]string `json:"-"`
}

// Element defines the structure of a single Element entity in the TM1 Server schema
// Note: We use this struct for both regular element definitions, for which we only specify
// the elemnets Name, as well as for element entity references, represented by @odata.id
type Element struct {
	Name string
}

// Edge defines the structure of a single Edge entity in the TM1 Server schema
type Edge struct {
	ParentName    string
	ComponentName string
	Weight        float64
}

// CreateDimension creates a new Dimension object, with the specified name, and returns the dimension
func CreateDimension(name string) *Dimension {
	return &Dimension{Name: name}
}

// AddHierarchy creates a new Hierarchy in the specified dimension, with the specified name, and returns the hierarchy
func (dimension *Dimension) AddHierarchy(name string) *Hierarchy {
	dimension.Hierarchies = append(dimension.Hierarchies, &Hierarchy{Name: name})
	hierarchy := dimension.Hierarchies[len(dimension.Hierarchies)-1]
	hierarchy.Captions = make(map[string]string)
	return hierarchy
}

// GetAttributesJSON returns the JSON specification to be passed to the Update action to update the attributes cube associated to the dimension
func (dimension *Dimension) GetAttributesJSON() string {
	return dimension.Hierarchies[0].GetAttributesJSON(dimension.Name)
}

// AddElement creates a new Element in the specified Hierarchy, with the specified name, and returns the element
func (hierarchy *Hierarchy) AddElement(name, caption string) *Element {
	hierarchy.Elements = append(hierarchy.Elements, Element{Name: name})
	if caption != "" {
		hierarchy.Captions[name] = caption
	}
	return &hierarchy.Elements[len(hierarchy.Elements)-1]
}

// AddEdge creates a new Edge in the specified Hierarchy, linking the specified parent and component, and returns the edge
func (hierarchy *Hierarchy) AddEdge(parent string, component string) *Edge {
	hierarchy.Edges = append(hierarchy.Edges, Edge{ParentName: parent, ComponentName: component, Weight: 1.0})
	return &hierarchy.Edges[len(hierarchy.Edges)-1]
}

// GetAttributesJSON returns the JSON specification to be passed to the Update action to update the attributes cube associated to the dimension
func (hierarchy *Hierarchy) GetAttributesJSON(dimensionName string) string {
	var jAttributes bytes.Buffer
	var bFirst = true
	jAttributes.WriteString("[")
	for element, caption := range hierarchy.Captions {
		if bFirst == true {
			bFirst = false
		} else {
			jAttributes.WriteString(",")
		}
		jAttributes.WriteString(`{"Slice@odata.bind":["Dimensions('`)
		jAttributes.WriteString(dimensionName)
		jAttributes.WriteString(`')/Hierarchies('`)
		jAttributes.WriteString(hierarchy.Name)
		jAttributes.WriteString(`')/Elements('`)
		jAttributes.WriteString(element)
		jAttributes.WriteString(`')","Dimensions('}ElementAttributes_`)
		jAttributes.WriteString(dimensionName)
		jAttributes.WriteString(`')/Hierarchies('}ElementAttributes_`)
		jAttributes.WriteString(dimensionName)
		jAttributes.WriteString(`')/Elements('Caption')"],"Value":"`)
		jAttributes.WriteString(caption)
		jAttributes.WriteString(`"}`)
	}
	jAttributes.WriteString("]")
	return jAttributes.String()
}

// CubePost defines the structure of a single Cube entity with the JSON annotations for POSTing (read: creating) one
type CubePost struct {
	Name         string
	DimensionIds []string `json:"Dimensions@odata.bind"`
	Rules        string   `json:",omitempty"`
}

// TransactionLogEntry defines the structure of A single TransactionLogEntry entity
type TransactionLogEntry struct {
	ChangeSetID     string
	TimeStamp       string // should be time.Time but because the seconds are omitted if 0 Go's time parser doesn't like them.
	ReplicationTime string // should be time.Time but because the seconds are omitted if 0 Go's time parser doesn't like them.
	User            string
	Cube            string
	Tuple           []string
	OldValue        json.RawMessage
	NewValue        json.RawMessage
	StatusMessage   string
}

// TransactionLogEntriesResponse defines the structure of an odata compliant response wrapping a TransactionLogEntry collection
type TransactionLogEntriesResponse struct {
	Context               string                `json:"@odata.context"`
	Count                 int                   `json:"@odata.count"`
	TransactionLogEntries []TransactionLogEntry `json:"value"`
	NextLink              string                `json:"@odata.nextLink"`
	DeltaLink             string                `json:"@odata.deltaLink"`
}
