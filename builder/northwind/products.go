package northwind

// Product defines the structure of A single Product entity
type Product struct {
	ID   int    `json:"ProductID"`
	Name string `json:"ProductName"`
}

// Category defines the structure of A single Category entity
type Category struct {
	ID       int       `json:"CategoryID"`
	Name     string    `json:"CategoryName"`
	Products []Product `json:"Products"`
}

// CategoryCollectionResponse defines the structure of an odata compliant response wrapping a category collection
type CategoryCollectionResponse struct {
	Context    string     `json:"@odata.context"`
	Count      int        `json:"@odata.count"`
	Categories []Category `json:"value"`
	NextLink   string     `json:"@odata.nextLink"`
}
