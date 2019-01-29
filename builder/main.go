package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/hubert-heijkers/GoThink2019/builder/helpers/odata"
	proc "github.com/hubert-heijkers/GoThink2019/builder/processes"
	"github.com/hubert-heijkers/GoThink2019/builder/tm1"
	"github.com/joho/godotenv"
)

// Environment variables
var datasourceServiceRootURL string
var tm1ServiceRootURL string

// Const defines
const productDimensionName = "Products"
const customerDimensionName = "Customers"
const employeeDimensionName = "Employees"
const timeDimensionName = "Time"
const measuresDimensionName = "Measures"
const ordersCubeName = "Sales"

// The http client, extended with some odata functions, we'll use throughout.
var client *odata.Client

// createDimension is the function that triggers the TM1 server to create the dimension
func createDimension(dimension *tm1.Dimension) *tm1.Dimension {

	// Create a JSON representation for the dimension
	jDimension, _ := json.Marshal(dimension)

	// POST the dimension to the TM1 server
	fmt.Println(">> Create dimension", dimension.Name)
	resp := client.ExecutePOSTRequest(tm1ServiceRootURL+"Dimensions", "application/json", string(jDimension))

	// Validate that the dimension got created successfully
	odata.ValidateStatusCode(resp, 201, func() string {
		return "Failed to create dimension '" + dimension.Name + "'."
	})
	resp.Body.Close()

	// Secondly create an element attribute named 'Caption' of type 'string'
	fmt.Println(">> Create 'Caption' attribute for dimension", dimension.Name)
	resp = client.ExecutePOSTRequest(tm1ServiceRootURL+"Dimensions('"+dimension.Name+"')/Hierarchies('"+dimension.Name+"')/ElementAttributes", "application/json", `{"Name":"Caption","Type":"String"}`)

	// Validate that the element attribute got created successfully as well
	odata.ValidateStatusCode(resp, 201, func() string {
		return "Creating element attribute 'Caption' for dimension '" + dimension.Name + "'."
	})
	resp.Body.Close()

	// Now that the caption attribute exists lets set the captions accordingly for this
	// we'll simply update the }ElementAttributes_DIMENSION cube directly, updating the
	// default value. Note: TM1 Server doesn't support passing the attribute values as
	// part of the dimension definition just yet (should shortly), so for now this is the
	// easiest way around that. Alternatively, one could have updated the attribute
	// values for elements one by one by POSTing to or PATCHing the LocalizedAttributes
	// of the individual elements.
	fmt.Println(">> Set 'Caption' attribute values for elements in dimension", dimension.Name)
	resp = client.ExecutePOSTRequest(tm1ServiceRootURL+"Cubes('}ElementAttributes_"+dimension.Name+"')/tm1.Update", "application/json", dimension.GetAttributesJSON())

	// Validate that the update executed successfully (by default an empty response is expected, hence the 204).
	odata.ValidateStatusCode(resp, 204, func() string {
		return "Setting Caption values for elements in dimension '" + dimension.Name + "'."
	})
	resp.Body.Close()

	// Return the generated dimension
	return dimension
}

// createCube is the function that, given a set of dimension and rules, requests the TM1 server to create the cube
func createCube(name string, dimensions []*tm1.Dimension, rules string) string {

	// Build array of dimension ids representing the dimensions making up the cube
	dimensionIds := make([]string, len(dimensions))
	for i, dim := range dimensions {
		dimensionIds[i] = "Dimensions('" + dim.Name + "')"
	}

	// Create a JSON representation for the cube
	jCube, _ := json.Marshal(tm1.CubePost{Name: name, DimensionIds: dimensionIds, Rules: rules})

	// POST the dimension to the TM1 server
	fmt.Println(">> Create cube", name)
	resp := client.ExecutePOSTRequest(tm1ServiceRootURL+"Cubes", "application/json", string(jCube))

	// Validate that the cube got created successfully
	odata.ValidateStatusCode(resp, 201, func() string {
		return "Failed to create cube '" + name + "'."
	})
	resp.Body.Close()

	// Return the odata.id of the generated cube
	return "Cubes('" + name + "')"
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
	client = &odata.Client{http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}}
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

	// Now let's build some Dimensions. The definition of the dimension is based on data
	// in the NorthWind database, a data source hosted on odata.org which can be queried
	// using its OData complaint REST API.
	var dimensions [5]*tm1.Dimension
	dimensions[0] = createDimension(proc.GenerateProductDimension(client, datasourceServiceRootURL, productDimensionName))
	dimensions[1] = createDimension(proc.GenerateCustomerDimension(client, datasourceServiceRootURL, customerDimensionName))
	dimensions[2] = createDimension(proc.GenerateEmployeeDimension(client, datasourceServiceRootURL, employeeDimensionName))
	dimensions[3] = createDimension(proc.GenerateTimeDimension(client, datasourceServiceRootURL, timeDimensionName))
	dimensions[4] = createDimension(proc.GenerateMeasuresDimension(client, datasourceServiceRootURL, measuresDimensionName))

	// Now that we have all our dimensions, let's create cube
	createCube(ordersCubeName, dimensions[:], "UNDEFVALS;\nSKIPCHECK;\n\n['UnitPrice']=['Revenue']\\['Quantity'];\n\nFEEDERS;\n['Quantity']=>['UnitPrice'];")

	// Load the data in the cube
	proc.LoadOrderData(client, datasourceServiceRootURL, tm1ServiceRootURL, ordersCubeName, dimensions[0], dimensions[1], dimensions[2], dimensions[3], dimensions[4])

	// And we are done!
	fmt.Println(">> Done!")
}
