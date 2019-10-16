This code is part of the TM1 SDK Hands-On Lab

The 'builder' is used to demo how one could build a TM1 model using the REST API from, in this particular case, Go.

It uses the NorthWind database, hosted on the odata.org, as the source for building the model.

The model being build contains 5 dimensions, 4 of which are build from data retrieved using the folling requests:
 - Products by Product Category: http://services.odata.org/V4/Northwind/Northwind.svc/Categories?$select=CategoryID,CategoryName&$orderby=CategoryName&$expand=Products($select=ProductID,ProductName;$orderby=ProductName)
 - Customers by Country, Region and City: http://services.odata.org/V4/Northwind/Northwind.svc/Customers?$orderby=Country%20asc,Region%20asc,%20City%20asc&$select=CustomerID,CompanyName,City,Region,Country
 - Employees by Country, Region and City: http://services.odata.org/V4/Northwind/Northwind.svc/Employees?$select=EmployeeID,LastName,FirstName,TitleOfCourtesy,City,Region,Country&$orderby=Country%20asc,Region%20asc,City%20asc 
 - Time span in the orders by looking at the first and last order dates:
   First: http://services.odata.org/V4/Northwind/Northwind.svc/Orders?$select=OrderDate&$orderby=OrderDate%20asc&$top=1
   Last: http://services.odata.org/V4/Northwind/Northwind.svc/Orders?$select=OrderDate&$orderby=OrderDate%20desc&$top=1
