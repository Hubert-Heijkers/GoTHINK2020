package northwind

// Customer defines the structure of A single Customer entity
type Customer struct {
	ID      string `json:"CustomerID"`
	Name    string `json:"CompanyName"`
	City    string
	Region  string
	Country string
}

// CustomerCollectionResponse defines the structure of an odata compliant response wrapping a customer collection
type CustomerCollectionResponse struct {
	Context   string     `json:"@odata.context"`
	Count     int        `json:"@odata.count"`
	Customers []Customer `json:"value"`
	NextLink  string     `json:"@odata.nextLink"`
}
