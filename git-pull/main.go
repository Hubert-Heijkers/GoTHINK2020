package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/hubert-heijkers/GoTHINK2020/common/odata"
	"github.com/joho/godotenv"
)

// GitPlanResponse defines the structure of an odata compliant response return a GitPlan
// Note that we are only including those fields that we are interested in!
type GitPlanResponse struct {
	Type       string `json:"@odata.type"`
	ID         string
	Branch     string
	Operations []string
}

// Environment variables
var tm1ServiceRootURL string

// The http client, extended with some odata functions, we'll use throughout.
var client *odata.Client

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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

	// Initialize GIT
	// bind the model to the github.com/hubert-heijkers/tm1-model-northwind repository
	fmt.Println(">> Initialize GIT...")
	resp = client.ExecutePOSTRequest(tm1ServiceRootURL+"GitInit", "application/json", `
	{
		"URL": "https://github.com/Hubert-Heijkers/tm1-model-northwind.git",
		"Deployment": "Development"
	}
	`)
	defer resp.Body.Close()

	// Prepare to pull the head of the master branch
	fmt.Println(">> Pull head from master...")
	resp = client.ExecutePOSTRequest(tm1ServiceRootURL+"GitPull", "application/json", `
	{
		"Branch": "master"
	}
	`)

	// The response contains the details about the git pull plan
	// Read the response body
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	// Unmarshal the git plan details
	gitPullPlan := GitPlanResponse{}
	err = json.Unmarshal(body, &gitPullPlan)
	if err != nil {
		log.Fatal(err)
	}

	// We could obviously implement some logic or ask the consumer for confirmation if he
	// or she whats to actually apply the changes as described in the git pull plan but
	// here we know we do and simply execute the git pull plan to apply all changes.
	fmt.Println(">> Execute the git pull plan...")
	resp = client.ExecutePOSTRequest(tm1ServiceRootURL+"GitPlans('"+gitPullPlan.ID+"')/tm1.Execute", "application/json", `{}`)

	// And we are done!
	fmt.Println(">> Done!")
}
