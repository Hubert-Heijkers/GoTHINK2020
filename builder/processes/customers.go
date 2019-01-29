package processes

import (
	"encoding/json"
	"log"

	"github.com/hubert-heijkers/GoThink2019/builder/helpers/odata"
	"github.com/hubert-heijkers/GoThink2019/builder/northwind"
	"github.com/hubert-heijkers/GoThink2019/builder/tm1"
)

type customerDimension struct {
	name           string
	dimension      *tm1.Dimension
	hierarchy      *tm1.Hierarchy
	allElement     *tm1.Element
	countryElement *tm1.Element
	regionElement  *tm1.Element
	cityElement    *tm1.Element
}

func (d *customerDimension) processResponse(responseBody []byte) (int, string) {
	// Unmarshal the JSON response
	res := northwind.CustomerCollectionResponse{}
	err := json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}

	// Process the collection of customers returned by the data source
	// PRESUMPTIONS:
	//  - NO DUPLICATE REGION NAMES (WHICH WORKS IN THIS EXAMPLE;-)
	//  - REGIONS CAN BE EMPTY IN WHICH CASE CITIES RESIDE UNDER COUNTRY
	//  - THERE ARE HOWEVER DUPLICATE CITY NAMES SO WE PRE-EMPT THEM WITH COUNTRY (WHICH SUFFICES IN THIS EXAMPLE)
	if d.hierarchy == nil {
		if d.dimension == nil {
			d.dimension = tm1.CreateDimension(d.name)
		}
		d.hierarchy = d.dimension.AddHierarchy(d.name)
		d.allElement = d.hierarchy.AddElement("All", "All Customers")
	}
	for _, customer := range res.Customers {
		if d.countryElement == nil || d.countryElement.Name != customer.Country {
			d.countryElement = d.hierarchy.AddElement(customer.Country, "")
			d.hierarchy.AddEdge(d.allElement.Name, d.countryElement.Name)
			if customer.Region != "" {
				d.regionElement = d.hierarchy.AddElement(customer.Region, "")
				d.hierarchy.AddEdge(d.countryElement.Name, d.regionElement.Name)
			} else {
				d.regionElement = nil
			}
			d.cityElement = d.hierarchy.AddElement(customer.Country+"-"+customer.City, customer.City)
			if d.regionElement != nil {
				d.hierarchy.AddEdge(d.regionElement.Name, d.cityElement.Name)
			} else {
				d.hierarchy.AddEdge(d.countryElement.Name, d.cityElement.Name)
			}
		} else if (d.regionElement == nil && customer.Region != "") || (d.regionElement != nil && d.regionElement.Name != customer.Region) {
			if customer.Region != "" {
				d.regionElement = d.hierarchy.AddElement(customer.Region, "")
				d.hierarchy.AddEdge(d.countryElement.Name, d.regionElement.Name)
			} else {
				d.regionElement = nil
			}
			d.cityElement = d.hierarchy.AddElement(customer.Country+"-"+customer.City, customer.City)
			if d.regionElement != nil {
				d.hierarchy.AddEdge(d.regionElement.Name, d.cityElement.Name)
			} else {
				d.hierarchy.AddEdge(d.countryElement.Name, d.cityElement.Name)
			}
		} else if d.cityElement == nil || d.cityElement.Name != customer.Country+"-"+customer.City {
			d.cityElement = d.hierarchy.AddElement(customer.Country+"-"+customer.City, customer.City)
			if d.regionElement != nil {
				d.hierarchy.AddEdge(d.regionElement.Name, d.cityElement.Name)
			} else {
				d.hierarchy.AddEdge(d.countryElement.Name, d.cityElement.Name)
			}
		}
		customerElement := d.hierarchy.AddElement(customer.ID, customer.Name)
		d.hierarchy.AddEdge(d.cityElement.Name, customerElement.Name)
	}

	// Return the nextLink, if there is one
	return res.Count, res.NextLink
}

// GenerateCustomerDimension generates, based on the data from the northwind database, the dimension definition for the customers dimension
func GenerateCustomerDimension(client *odata.Client, datasourceServiceRootURL string, name string) *tm1.Dimension {
	dimCustomers := &customerDimension{name: name}
	client.IterateCollection(datasourceServiceRootURL, "Customers?$orderby=Country%20asc,Region%20asc,%20City%20asc&$select=CustomerID,CompanyName,City,Region,Country", dimCustomers.processResponse)
	return dimCustomers.dimension
}
