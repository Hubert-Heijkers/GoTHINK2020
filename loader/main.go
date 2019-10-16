package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"

	"github.com/hubert-heijkers/GoDAIF2019/builder/northwind"
	"github.com/hubert-heijkers/GoDAIF2019/common/odata"
	"github.com/joho/godotenv"
)

// Environment variables
var datasourceServiceRootURL string
var tm1ServiceRootURL string

// The http client, extended with some odata functions, we'll use throughout.
var client *odata.Client

func processOrderData(responseBody []byte) (int, string) {
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
	// Also note that we know the dimension names so we directly pass them here,
	// in a real world example it might take a bit more setting up and validation
	// then assuming the dimension names and presuming they all have a same named
	// hierarchy (a-typically in cases with multiple, alternate, hierarchies)
	var jCellUpdates bytes.Buffer
	var bFirst = true
	jCellUpdates.WriteString("[")
	for _, order := range res.Orders {
		var jOrderTuple bytes.Buffer
		jOrderTuple.WriteString(`"Dimensions('Customers')/Hierarchies('Customers')/Elements('`)
		jOrderTuple.WriteString(order.CustomerID)
		jOrderTuple.WriteString(`')","Dimensions('Employees')/Hierarchies('Employees')/Elements('`)
		jOrderTuple.WriteString(strconv.Itoa(order.EmployeeID))
		jOrderTuple.WriteString(`')","Dimensions('Time')/Hierarchies('Time')/Elements('`)
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
			jCellUpdates.WriteString(`')","Dimensions('Products')/Hierarchies('Products')/Elements('P-`)
			jCellUpdates.WriteString(strconv.Itoa(detail.ProductID))
			jCellUpdates.WriteString(`')","Dimensions('Measures')/Hierarchies('Measures')/Elements('Quantity')"],"Value":"+`)
			jCellUpdates.WriteString(strconv.Itoa(detail.Quantity))
			jCellUpdates.WriteString(`"},`)
			// Revenue
			jCellUpdates.WriteString(`{"Slice@odata.bind":[`)
			jCellUpdates.Write(jOrderTuple.Bytes())
			jCellUpdates.WriteString(`')","Dimensions('Products')/Hierarchies('Products')/Elements('P-`)
			jCellUpdates.WriteString(strconv.Itoa(detail.ProductID))
			jCellUpdates.WriteString(`')","Dimensions('Measures')/Hierarchies('Measures')/Elements('Revenue')"],"Value":"+`)
			jCellUpdates.WriteString(fmt.Sprintf("%f", float64(detail.Quantity)*detail.UnitPrice))
			jCellUpdates.WriteString(`"}`)
		}
	}
	jCellUpdates.WriteString("]")

	fmt.Println(">> Loading order data...")
	resp := client.ExecutePOSTRequest(tm1ServiceRootURL+"Cubes('Sales')/tm1.Update", "application/json", jCellUpdates.String())

	// Validate that the update executed successfully (by default an empty response is expected, hence the 204).
	odata.ValidateStatusCode(resp, 204, func() string {
		return "Loading data into cube 'Sales'."
	})
	resp.Body.Close()

	// Return the nextLink, if there is one
	return res.Count, res.NextLink
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	datasourceServiceRootURL = os.Getenv("DATASOURCE_SERVICE_ROOT_URL")
	tm1ServiceRootURL = os.Getenv("TM1_SERVICE_ROOT_URL")

	// Create the one and only http client we'll be using, with a cookie jar enabled to keep reusing our session
	client = &odata.Client{}
	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar

	// Validate that the TM1 server is accessable by requesting the version of the server
	req, _ := http.NewRequest("GET", tm1ServiceRootURL+"Configuration/ProductVersion/$value", nil)

	// Since this is our initial request we'll have to provide a user name and
	// password, also conveniently stored in the environment variables, to authenticate.
	// Note: using authentication mode 1, TM1 authentication, which maps to basic
	// authentication in HTTP[S]
	req.SetBasicAuth(os.Getenv("TM1_USER"), os.Getenv("TM1_PASSWORD"))

	// We'll expect text back in this case but we'll simply dump the content out and
	// won't do any content type verification here
	req.Header.Add("Accept", "*/*")

	// Let's execute the request
	resp, err := client.Do(req)
	if err != nil {
		// Execution of the request failed, log the error and terminate
		log.Fatal(err)
	}

	// Validate that the request executed successfully
	odata.ValidateStatusCode(resp, 200, func() string {
		return "Server responded with an unexpected result while asking for its version number."
	})

	// The body simply contains the version number of the server
	version, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// which we'll simply dump to the console
	fmt.Println("Using TM1 Server version", string(version))

	// Note that as a result of this request a TM1SessionId cookie was added to the cookie
	// jar which will automatically be reused on subsequent requests to our TM1 server,
	// and therefore don't need to send the credentials over and over again.

	// Load the data in the cube
	// The load once again uses one of our utility functions, IterateCollection, that
	// iterates the collection and calls back to our processOrderData function.
	// The load itself is based on the data from the northwind database, from which we
	// read the order data once again and this time put the data into our Sales cube.
	client.IterateCollection(datasourceServiceRootURL, "Orders?$select=CustomerID,EmployeeID,OrderDate&$expand=Order_Details($select=ProductID,UnitPrice,Quantity)", processOrderData)

	// And we are done!
	fmt.Println(">> Done!")
}
