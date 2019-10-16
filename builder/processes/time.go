package processes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/hubert-heijkers/GoDAIF2019/builder/northwind"
	"github.com/hubert-heijkers/GoDAIF2019/common/odata"
	"github.com/hubert-heijkers/GoDAIF2019/common/tm1"
)

// GenerateTimeDimension generates, based on the data from the northwind database, the dimension definition for the time dimension
func GenerateTimeDimension(client *odata.Client, datasourceServiceRootURL string, name string) *tm1.Dimension {

	// Grab the orderdate of the FIRST order, by order data, in the system
	resp := client.ExecuteGETRequest(datasourceServiceRootURL + "Orders?$select=OrderDate&$orderby=OrderDate%20asc&$top=1")
	odata.ValidateStatusCode(resp, 200, func() string {
		return "Failed to retrieve first order date."
	})
	defer resp.Body.Close()
	responseBody, _ := ioutil.ReadAll(resp.Body)
	res := northwind.OrderCollectionResponse{}
	err := json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}
	tmBegin := res.Orders[0].Date

	// Grab the orderdate of the LAST order, by order data, in the system
	resp = client.ExecuteGETRequest(datasourceServiceRootURL + "Orders?$select=OrderDate&$orderby=OrderDate%20desc&$top=1")
	odata.ValidateStatusCode(resp, 200, func() string {
		return "Failed to retrieve last order date."
	})
	defer resp.Body.Close()
	responseBody, _ = ioutil.ReadAll(resp.Body)
	res = northwind.OrderCollectionResponse{}
	err = json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}
	tmEnd := res.Orders[0].Date

	// Show the order date range we are going to use to create the time dimension
	fmt.Println("Order date range:", tmBegin.Format(time.ANSIC), "-", tmEnd.Format(time.ANSIC))

	// Build the dimension definition
	dimension := tm1.CreateDimension(name)

	// Build a same named hierarchy, which might be atypical, but otherwise we don't see any in Architec;-)
	hierarchy := dimension.AddHierarchy(name)
	allElement := hierarchy.AddElement("All", "All Years")
	year := tmBegin.Year() - 1
	var month int
	var quarter int
	var yearElement *tm1.Element
	var quarterElement *tm1.Element
	var monthElement *tm1.Element
	// Create elements for every day in the range from the first till last day we have data for
	for iTm := tmBegin; iTm.After(tmEnd) == false; iTm = iTm.AddDate(0, 0, 1) {
		if iTm.Year() != year {
			year = iTm.Year()
			yearElement = hierarchy.AddElement(strconv.Itoa(year), "")
			hierarchy.AddEdge(allElement.Name, yearElement.Name)
			month = int(iTm.Month())
			quarter = (month + 2) / 3
			quarterElement = hierarchy.AddElement(fmt.Sprintf("Q%d-%04d", quarter, year), fmt.Sprintf("Q%d %04d", quarter, year))
			hierarchy.AddEdge(yearElement.Name, quarterElement.Name)
			monthElement = hierarchy.AddElement(fmt.Sprintf("%02d-%04d", month, year), fmt.Sprintf("%s %04d", iTm.Month().String()[:3], year))
			hierarchy.AddEdge(quarterElement.Name, monthElement.Name)
		} else if (int(iTm.Month())+2)/3 != quarter {
			month = int(iTm.Month())
			quarter = (month + 2) / 3
			quarterElement = hierarchy.AddElement(fmt.Sprintf("Q%d-%04d", quarter, year), fmt.Sprintf("Q%d %04d", quarter, year))
			hierarchy.AddEdge(yearElement.Name, quarterElement.Name)
			monthElement = hierarchy.AddElement(fmt.Sprintf("%02d-%04d", month, year), fmt.Sprintf("%s %04d", iTm.Month().String()[:3], year))
			hierarchy.AddEdge(quarterElement.Name, monthElement.Name)
		} else if int(iTm.Month()) != month {
			month = int(iTm.Month())
			monthElement = hierarchy.AddElement(fmt.Sprintf("%02d-%04d", month, year), fmt.Sprintf("%s %04d", iTm.Month().String()[:3], year))
			hierarchy.AddEdge(quarterElement.Name, monthElement.Name)
		}
		dayElement := hierarchy.AddElement(fmt.Sprintf("%02d-%02d-%04d", iTm.Day(), month, year), fmt.Sprintf("%s %s %2d %04d", iTm.Weekday().String()[:3], iTm.Month().String()[:3], iTm.Day(), year))
		hierarchy.AddEdge(monthElement.Name, dayElement.Name)
	}

	// Now lets add year, quarter and month hierarchies
	// Years
	yearHierarchy := dimension.AddHierarchy("Years")
	allYearsElement := yearHierarchy.AddElement("Years", "All Years")
	yearElement = nil
	year = tmBegin.Year() - 1
	// Quarters
	quarterHierarchy := dimension.AddHierarchy("Quarters")
	allQuartersElement := quarterHierarchy.AddElement("Quarters", "All Quarters")
	var quarterElements [4]*tm1.Element
	for i := 0; i < 4; i++ {
		quarterElements[i] = quarterHierarchy.AddElement(fmt.Sprintf("Q%d", i+1), fmt.Sprintf("Quarter %d", i+1))
		quarterHierarchy.AddEdge(allQuartersElement.Name, quarterElements[i].Name)
	}
	// Months
	monthHierarchy := dimension.AddHierarchy("Months")
	allMonthsElement := monthHierarchy.AddElement("Months", "All Months")
	var monthElements [12]*tm1.Element
	for i := 0; i < 12; i++ {
		monthElements[i] = monthHierarchy.AddElement(time.Month(i+1).String(), "")
		monthHierarchy.AddEdge(allMonthsElement.Name, monthElements[i].Name)
	}

	year = tmBegin.Year() - 1
	// Create elements for every day in the range from the first till last day we have data for
	for iTm := tmBegin; iTm.After(tmEnd) == false; iTm = iTm.AddDate(0, 0, 1) {
		if iTm.Year() != year {
			year = iTm.Year()
			yearElement = yearHierarchy.AddElement(strconv.Itoa(year), "")
			yearHierarchy.AddEdge(allYearsElement.Name, yearElement.Name)
		}
		month = int(iTm.Month())
		quarter = (month + 2) / 3
		dayName := fmt.Sprintf("%02d-%02d-%04d", iTm.Day(), month, year)
		yearHierarchy.AddEdge(yearElement.Name, yearHierarchy.AddElement(dayName, "").Name)
		quarterHierarchy.AddEdge(quarterElements[quarter-1].Name, quarterHierarchy.AddElement(dayName, "").Name)
		monthHierarchy.AddEdge(monthElements[month-1].Name, monthHierarchy.AddElement(dayName, "").Name)
	}

	return dimension
}
