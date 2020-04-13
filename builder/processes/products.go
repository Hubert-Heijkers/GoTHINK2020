package processes

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/hubert-heijkers/GoTHINK2020/builder/northwind"
	"github.com/hubert-heijkers/GoTHINK2020/common/odata"
	"github.com/hubert-heijkers/GoTHINK2020/common/tm1"
)

type productDimension struct {
	name            string
	dimension       *tm1.Dimension
	hierarchy       *tm1.Hierarchy
	allElement      *tm1.Element
	categoryElement *tm1.Element
}

func (d *productDimension) ProcessResponse(responseBody []byte) (int, string) {
	// Unmarshal the JSON response
	res := northwind.CategoryCollectionResponse{}
	err := json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}

	// Process the collection of products, by category, returned by the data source
	// PRESUMPTIONS:
	//  - NO DUPLICATE CATEGORY OR PRODUCT IDS
	//  - HOWEVER IDS ARE NOT UNIQUE ACROSS CATEGORIES AND PRODUCTS
	if d.hierarchy == nil {
		if d.dimension == nil {
			d.dimension = tm1.CreateDimension(d.name)
		}
		d.hierarchy = d.dimension.AddHierarchy(d.name)
		d.allElement = d.hierarchy.AddElement("All", "All Products")
	}
	for _, category := range res.Categories {
		if d.categoryElement == nil || d.categoryElement.Name != "C-"+category.Name {
			d.categoryElement = d.hierarchy.AddElement("C-"+strconv.Itoa(category.ID), category.Name)
			d.hierarchy.AddEdge(d.allElement.Name, d.categoryElement.Name)
		}
		for _, product := range category.Products {
			productElement := d.hierarchy.AddElement("P-"+strconv.Itoa(product.ID), product.Name)
			d.hierarchy.AddEdge(d.categoryElement.Name, productElement.Name)
		}
	}

	// Return the nextLink, if there is one
	return res.Count, res.NextLink
}

// GenerateProductDimension generates, based on the data from the northwind database, the dimension definition for the products dimension
func GenerateProductDimension(client *odata.Client, datasourceServiceRootURL string, name string) *tm1.Dimension {
	dimProducts := &productDimension{name: name}
	client.IterateCollection(datasourceServiceRootURL, "Categories?$select=CategoryID,CategoryName&$orderby=CategoryName&$expand=Products($select=ProductID,ProductName;$orderby=ProductName)", dimProducts.ProcessResponse)
	return dimProducts.dimension
}
