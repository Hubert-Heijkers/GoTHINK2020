package processes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hubert-heijkers/GoThink2019/builder/helpers/odata"
	"github.com/hubert-heijkers/GoThink2019/builder/northwind"
	"github.com/hubert-heijkers/GoThink2019/builder/tm1"
)

type ordersCubeDataLoad struct {
	client            *odata.Client
	tm1ServiceRootURL string
	cubeName          string
	productDimension  *tm1.Dimension
	customerDimension *tm1.Dimension
	employeeDimension *tm1.Dimension
	timeDimension     *tm1.Dimension
	measuresDimension *tm1.Dimension
}

func (c *ordersCubeDataLoad) processOrderData(responseBody []byte) (int, string) {
	// Unmarshal the JSON response
	res := northwind.OrderCollectionResponse{}
	err := json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}

	// Process the collection of orders and convert to a set of cell updates
	// Note that we are using making it easy on ourselves here and not perform
	// pre-aggregation at the cell level and use a 'spreading' command to update
	// the cells to not have to retrieve the data before adding the new value to
	// it before updating the cell again.
	var jCellUpdates bytes.Buffer
	var bFirst = true
	jCellUpdates.WriteString("[")
	for _, order := range res.Orders {
		var jOrderTuple bytes.Buffer
		jOrderTuple.WriteString(`"Dimensions('`)
		jOrderTuple.WriteString(c.customerDimension.Name)
		jOrderTuple.WriteString(`')/Hierarchies('`)
		jOrderTuple.WriteString(c.customerDimension.Hierarchies[0].Name)
		jOrderTuple.WriteString(`')/Elements('`)
		jOrderTuple.WriteString(order.CustomerID)
		jOrderTuple.WriteString(`')","Dimensions('`)
		jOrderTuple.WriteString(c.employeeDimension.Name)
		jOrderTuple.WriteString(`')/Hierarchies('`)
		jOrderTuple.WriteString(c.employeeDimension.Hierarchies[0].Name)
		jOrderTuple.WriteString(`')/Elements('`)
		jOrderTuple.WriteString(strconv.Itoa(order.EmployeeID))
		jOrderTuple.WriteString(`')","Dimensions('`)
		jOrderTuple.WriteString(c.timeDimension.Name)
		jOrderTuple.WriteString(`')/Hierarchies('`)
		jOrderTuple.WriteString(c.timeDimension.Hierarchies[0].Name)
		jOrderTuple.WriteString(`')/Elements('`)
		jOrderTuple.WriteString(fmt.Sprintf("%02d-%02d-%04d", order.Date.Day(), int(order.Date.Month()), order.Date.Year()))
		for _, detail := range order.Details {
			if bFirst == true {
				bFirst = false
			} else {
				jCellUpdates.WriteString(",")
			}
			// Quantity
			jCellUpdates.WriteString(`{"Slice@odata.bind":[`)
			jCellUpdates.Write(jOrderTuple.Bytes())
			jCellUpdates.WriteString(`')","Dimensions('`)
			jCellUpdates.WriteString(c.productDimension.Name)
			jCellUpdates.WriteString(`')/Hierarchies('`)
			jCellUpdates.WriteString(c.productDimension.Hierarchies[0].Name)
			jCellUpdates.WriteString(`')/Elements('P-`)
			jCellUpdates.WriteString(strconv.Itoa(detail.ProductID))
			jCellUpdates.WriteString(`')","Dimensions('`)
			jCellUpdates.WriteString(c.measuresDimension.Name)
			jCellUpdates.WriteString(`')/Hierarchies('`)
			jCellUpdates.WriteString(c.measuresDimension.Hierarchies[0].Name)
			jCellUpdates.WriteString(`')/Elements('Quantity')"],"Value":"+`)
			jCellUpdates.WriteString(strconv.Itoa(detail.Quantity))
			jCellUpdates.WriteString(`"},`)
			// Revenue
			jCellUpdates.WriteString(`{"Slice@odata.bind":[`)
			jCellUpdates.Write(jOrderTuple.Bytes())
			jCellUpdates.WriteString(`')","Dimensions('`)
			jCellUpdates.WriteString(c.productDimension.Name)
			jCellUpdates.WriteString(`')/Hierarchies('`)
			jCellUpdates.WriteString(c.productDimension.Hierarchies[0].Name)
			jCellUpdates.WriteString(`')/Elements('P-`)
			jCellUpdates.WriteString(strconv.Itoa(detail.ProductID))
			jCellUpdates.WriteString(`')","Dimensions('`)
			jCellUpdates.WriteString(c.measuresDimension.Name)
			jCellUpdates.WriteString(`')/Hierarchies('`)
			jCellUpdates.WriteString(c.measuresDimension.Hierarchies[0].Name)
			jCellUpdates.WriteString(`')/Elements('Revenue')"],"Value":"+`)
			jCellUpdates.WriteString(fmt.Sprintf("%f", float64(detail.Quantity)*detail.UnitPrice))
			jCellUpdates.WriteString(`"}`)
		}
	}
	jCellUpdates.WriteString("]")

	fmt.Println(">> Loading order data...")
	resp := c.client.ExecutePOSTRequest(c.tm1ServiceRootURL+"Cubes('"+c.cubeName+"')/tm1.Update", "application/json", jCellUpdates.String())

	// Validate that the update executed successfully (by default an empty response is expected, hence the 204).
	odata.ValidateStatusCode(resp, 204, func() string {
		return "Loading data into cube '" + c.cubeName + "'."
	})
	resp.Body.Close()

	// Return the nextLink, if there is one
	return res.Count, res.NextLink
}

// LoadOrderData loads, based on the data from the northwind database, the order data into our cube
func LoadOrderData(client *odata.Client, datasourceServiceRootURL, tm1ServiceRootURL, cubeName string, productDimension, customerDimension, employeeDimension, timeDimension, measuresDimension *tm1.Dimension) {
	cubeLoadData := &ordersCubeDataLoad{client: client,
		tm1ServiceRootURL: tm1ServiceRootURL,
		cubeName:          cubeName,
		productDimension:  productDimension,
		customerDimension: customerDimension,
		employeeDimension: employeeDimension,
		timeDimension:     timeDimension,
		measuresDimension: measuresDimension,
	}
	client.IterateCollection(datasourceServiceRootURL, "Orders?$select=CustomerID,EmployeeID,OrderDate&$expand=Order_Details($select=ProductID,UnitPrice,Quantity)", cubeLoadData.processOrderData)
}
