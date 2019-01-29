package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/hubert-heijkers/GoThink2018/builder/helpers/odata"
	"github.com/hubert-heijkers/GoThink2018/builder/tm1"
	"github.com/joho/godotenv"
)

// Environment variables
var tm1ServiceRootURL string

// Const defines
const ordersCubeName = "Sales"

// The http client, extended with some odata functions, we'll use throughout.
var client *odata.Client

func processTransactionLogEntries(responseBody []byte) (string, string) {
	// Unmarshal the JSON response
	res := tm1.TransactionLogEntriesResponse{}
	err := json.Unmarshal(responseBody, &res)
	if err != nil {
		log.Fatal(err)
	}

	// Process the entries by simply dumping them in a nicely consumable from to the console
	for _, entry := range res.TransactionLogEntries {
		var out bytes.Buffer
		out.WriteString(entry.TimeStamp)
		out.WriteString(" ")
		out.WriteString(entry.Cube)
		out.WriteString("['")
		for i, element := range entry.Tuple {
			if i > 0 {
				out.WriteString("','")
			}
			out.WriteString(element)
		}
		out.WriteString("']: ")
		out.WriteString(string(entry.OldValue))
		out.WriteString(" => ")
		out.WriteString(string(entry.NewValue))
		fmt.Println(out.String())
	}

	// Return the nextLink and deltaLink, if there any
	return res.NextLink, res.DeltaLink
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tm1ServiceRootURL = os.Getenv("TM1_SERVICE_ROOT_URL")

	// Turn 'Verbose' mode off
	odata.Verbose = false

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

	// Track the collection of transaction log entries. This will query the existing entries and then cause
	// the server to query the delta of the collection (read: just the changes) after a defined duration.
	client.TrackCollection(tm1ServiceRootURL, "TransactionLogEntries?$filter=Cube%20eq%20'"+url.QueryEscape(ordersCubeName)+"'", 1*time.Second, processTransactionLogEntries)
}
