package northwind

import (
	"time"
)

// OrderDetail defines the structure of A single order detail (read: line) entity
type OrderDetail struct {
	ProductID int
	UnitPrice float64
	Quantity  int
}

// Order defines the structure of A single Order entity
type Order struct {
	CustomerID string
	EmployeeID int
	Date       time.Time     `json:"OrderDate"`
	Details    []OrderDetail `json:"Order_Details"`
}

// OrderCollectionResponse defines the structure of an odata compliant response wrapping a order collection
type OrderCollectionResponse struct {
	Context  string  `json:"@odata.context"`
	Count    int     `json:"@odata.count"`
	Orders   []Order `json:"value"`
	NextLink string  `json:"@odata.nextLink"`
}
