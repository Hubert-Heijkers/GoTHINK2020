package processes

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/hubert-heijkers/GoThink2019/builder/helpers/odata"
	"github.com/hubert-heijkers/GoThink2019/builder/northwind"
	"github.com/hubert-heijkers/GoThink2019/builder/tm1"
)

type employeeDimension struct {
	name                string
	dimension           *tm1.Dimension
	geographyHierarchy  *tm1.Hierarchy
	allGeographyElement *tm1.Element
	countryElement      *tm1.Element
	regionElement       *tm1.Element
	cityElement         *tm1.Element
	generationHierarchy *tm1.Hierarchy
	generationElements  [5]*tm1.Element
}

func (d *employeeDimension) processResponse(responseBody []byte) (int, string) {
	// Unmarshal the JSON response
	res := northwind.EmployeeCollectionResponse{}
	err := json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}

	// Process the collection of employees returned by the data source
	// PRESUMPTIONS:
	//  - NO DUPLICATE REGION NAMES (WHICH WORKS IN THIS EXAMPLE;-)
	//  - REGIONS CAN BE EMPTY IN WHICH CASE CITIES RESIDE UNDER COUNTRY
	//  - THERE ARE, FORTUNATELY FOR THIS EXAMPLE, NO DUPLICATE CITY NAMES EITHER
	for _, employee := range res.Employees {
		// Geography hierarchy
		if d.countryElement == nil || d.countryElement.Name != employee.Country {
			d.countryElement = d.geographyHierarchy.AddElement(employee.Country, "")
			d.geographyHierarchy.AddEdge(d.allGeographyElement.Name, d.countryElement.Name)
			if employee.Region != "" {
				d.regionElement = d.geographyHierarchy.AddElement(employee.Region, "")
				d.geographyHierarchy.AddEdge(d.countryElement.Name, d.regionElement.Name)
			} else {
				d.regionElement = nil
			}
			d.cityElement = d.geographyHierarchy.AddElement(employee.City, "")
			if d.regionElement != nil {
				d.geographyHierarchy.AddEdge(d.regionElement.Name, d.cityElement.Name)
			} else {
				d.geographyHierarchy.AddEdge(d.countryElement.Name, d.cityElement.Name)
			}
		} else if (d.regionElement == nil && employee.Region != "") || (d.regionElement != nil && d.regionElement.Name != employee.Region) {
			if employee.Region != "" {
				d.regionElement = d.geographyHierarchy.AddElement(employee.Region, "")
				d.geographyHierarchy.AddEdge(d.countryElement.Name, d.regionElement.Name)
			} else {
				d.regionElement = nil
			}
			d.cityElement = d.geographyHierarchy.AddElement(employee.City, "")
			if d.regionElement != nil {
				d.geographyHierarchy.AddEdge(d.regionElement.Name, d.cityElement.Name)
			} else {
				d.geographyHierarchy.AddEdge(d.countryElement.Name, d.cityElement.Name)
			}
		} else if d.cityElement == nil || d.cityElement.Name != employee.City {
			d.cityElement = d.geographyHierarchy.AddElement(employee.City, "")
			if d.regionElement != nil {
				d.geographyHierarchy.AddEdge(d.regionElement.Name, d.cityElement.Name)
			} else {
				d.geographyHierarchy.AddEdge(d.countryElement.Name, d.cityElement.Name)
			}
		}
		employeeElement := d.geographyHierarchy.AddElement(strconv.Itoa(employee.ID), employee.LastName+", "+employee.FirstName)
		d.geographyHierarchy.AddEdge(d.cityElement.Name, employeeElement.Name)

		// Generation hierarchy
		employeeElement = d.generationHierarchy.AddElement(strconv.Itoa(employee.ID), "")
		year := employee.BirthDate.Year()
		switch {
		case year <= 1945:
			d.generationHierarchy.AddEdge(d.generationElements[0].Name, employeeElement.Name)
		case year >= 1946 && year <= 1964:
			d.generationHierarchy.AddEdge(d.generationElements[1].Name, employeeElement.Name)
		default:
			d.generationHierarchy.AddEdge(d.generationElements[2].Name, employeeElement.Name)
		}
	}

	// Return the nextLink, if there is one
	return res.Count, res.NextLink
}

// GenerateEmployeeDimension generates, based on the data from the northwind database, the dimension definition for the employees dimension
func GenerateEmployeeDimension(client *odata.Client, datasourceServiceRootURL string, name string) *tm1.Dimension {
	dimEmployees := &employeeDimension{
		name:      name,
		dimension: tm1.CreateDimension(name),
	}
	// Note, a more logical name for the geography hierarchy might be something like, well, 'Geography' but we'll
	// use 'Employee', same name as the dimension, for backwards compatibility, in this case with Architect.
	dimEmployees.geographyHierarchy = dimEmployees.dimension.AddHierarchy(name)
	dimEmployees.allGeographyElement = dimEmployees.geographyHierarchy.AddElement("All", "All Geographies")
	dimEmployees.generationHierarchy = dimEmployees.dimension.AddHierarchy("Generation")
	allGenerationsElement := dimEmployees.generationHierarchy.AddElement("All", "All Generations")
	dimEmployees.generationElements[0] = dimEmployees.generationHierarchy.AddElement("1925-1945", "The Silent Generation (1925-1945)")
	dimEmployees.generationHierarchy.AddEdge(allGenerationsElement.Name, dimEmployees.generationElements[0].Name)
	dimEmployees.generationElements[1] = dimEmployees.generationHierarchy.AddElement("1946-1964", "The Baby Boomers (1946-1964)")
	dimEmployees.generationHierarchy.AddEdge(allGenerationsElement.Name, dimEmployees.generationElements[1].Name)
	dimEmployees.generationElements[2] = dimEmployees.generationHierarchy.AddElement("1965-1979", "Generation X (1965-1979)")
	dimEmployees.generationHierarchy.AddEdge(allGenerationsElement.Name, dimEmployees.generationElements[2].Name)
	/*
		dimEmployees.generationElements[3] = dimEmployees.generationHierarchy.AddElement("1980-1995", "The Millennials (1980-1995)")
		dimEmployees.generationHierarchy.AddEdge(allGenerationsElement.Name, dimEmployees.generationElements[3].Name)
		dimEmployees.generationElements[4] = dimEmployees.generationHierarchy.AddElement("1996-2010", "Generation Z (1996-2010)")
		dimEmployees.generationHierarchy.AddEdge(allGenerationsElement.Name, dimEmployees.generationElements[4].Name)
	*/

	client.IterateCollection(datasourceServiceRootURL, "Employees?$select=EmployeeID,LastName,FirstName,TitleOfCourtesy,City,Region,Country,BirthDate&$orderby=Country%20asc,Region%20asc,City%20asc", dimEmployees.processResponse)
	return dimEmployees.dimension
}
