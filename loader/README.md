This code is part of the TM1 SDK Hands-On Lab

The 'loader' is used to demo how one could load data into a TM1 model, in this case one that was either previously build from that same source using the 'builder' or, created from source using the github.com/hubert-heijkers/tm1-model-northwind GIT repository. The loader code itself uses the REST API from, in this particular case, Go.

The data loaded into the model is sourced from the NorthWind database, hosted on the odata.org.

The Sales cube is being loaded with data coming from the orders that are in the NorthWind database retieved using: 
 - The orders, our data: http://services.odata.org/V4/Northwind/Northwind.svc/Orders?$select=CustomerID,EmployeeID,OrderDate&$expand=Order_Details($select=ProductID,UnitPrice,Quantity)
