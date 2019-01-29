The 'builder' is the code used in the Hands-On Lab 3548 at World of Watson 2016
It's used to demo how one could build a TM1 model using the REST API from, in this particular case, Go.
It uses the NorthWind database, hosted on the odata.org, as the source for building the model.

The model being build contains 5 dimensions, 4 of which are build from data retrieved using the folling requests:
 - Products by Product Category: http://services.odata.org/V4/Northwind/Northwind.svc/Categories?$select=CategoryID,CategoryName&$orderby=CategoryName&$expand=Products($select=ProductID,ProductName;$orderby=ProductName)
 - Customers by Country, Region and City: http://services.odata.org/V4/Northwind/Northwind.svc/Customers?$orderby=Country%20asc,Region%20asc,%20City%20asc&$select=CustomerID,CompanyName,City,Region,Country
 - Employees by Country, Region and City: http://services.odata.org/V4/Northwind/Northwind.svc/Employees?$select=EmployeeID,LastName,FirstName,TitleOfCourtesy,City,Region,Country&$orderby=Country%20asc,Region%20asc,City%20asc 
 - Time span in the orders by looking at the first and last order dates:
   First: http://services.odata.org/V4/Northwind/Northwind.svc/Orders?$select=OrderDate&$orderby=OrderDate%20asc&$top=1
   Last: http://services.odata.org/V4/Northwind/Northwind.svc/Orders?$select=OrderDate&$orderby=OrderDate%20desc&$top=1

The Sales cube is being loaded with data coming from the orders that are in the NorthWind database retieved using: 
 - The orders, our data: http://services.odata.org/V4/Northwind/Northwind.svc/Orders?$select=CustomerID,EmployeeID,OrderDate&$expand=Order_Details($select=ProductID,UnitPrice,Quantity)
