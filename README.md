Backendbauer
============

## What it is

Backendbauer is a backend server with REST API for generating json data for frontend js charts.
It is written in Go and used by Mopinion in production.

## What it needs
* The [Go](http://golang.org) language
* The charts library from Highcharts JS.
http://www.highcharts.com/
* [MyMySQL] (http://github.com/ziutek/mymysql) for Go

## How it works

Only two files are needed to get Backendbauer working:
- server/server.go
- server/config.json

For testing you can also use backendbauer.js

Server.go is the source code. In config.json you can set MySQL databases with custom tables and fields.
All settings are based on the url from which the API call is made, the referer.

### Install
Go to server folder  
`$ cd server`  
build server  
`$ sudo go build -o server server.go`  
`sudo chmod 755 server`  
make symbolic link  
`$ sudo ln -s [backendbauer dir]/server/server /usr/sbin/backendbauer`  
copy service file to /etc/init.d  
`$ sudo cp backendbauer /etc/init.d/`  
`sudo chmod 755 /etc/init.d/backendbauer`  
To start and stop  
`sudo service backendbauer start`  
`sudo service backendbauer stop`

## API

Endpoint:  
`http://[host name]:[port]/data`  

The API has a number of variables in order to get the right data in json format.

- x: data field id as configured in config.json
- y: data field id as configured in config.json
- from_date: start date for query in `YYYY-MM-DD` format  
- to_date: end date for query in `YYYY-MM-DD` format  
- avg: is it the average value (only possible for numerical fields) or the count? `0/1`
- filter: a custom filter that is used to make the query.
Filters can be added to narrow the query down in the following manner:  
`|[field]:[value]` translates to `AND [field] = "[value]"`  
`|[field]![value]` translates to `AND [field] <> "[value]"` 
- chart_type (optional): `line` or `pie`, etc. The response will send the type back. This can be used in some cases.
- series (optional): `0/1` sometimes the ajax js code differs when a request is a series or a chart. The response returns this value.
- jsonp: If `true` and a callback function is specified, the response will add the callback function (needed for jsonp crossdomain/port calls)
- callback: if specified and `jsonp=true` the response will be included in a js callback (needed for jsonp crossdomain/port calls)
- order (optional): `asc` or `desc` if given the query will be ordered ascending or descending respectively on the x variable
- limit (optional): `[number]` for example `10` limits the result to 10 rows
- combined (optional): when `true` the response will include the categories in the data `[['category1','data1']['category2']['data2']]` instead of `['data1','data2','data3']`. Easier to add series after a chart already exists
- name (optional): the name of the y variable in the series. Response returns this name, so it can be used in js manipulation of the highchart object.

## Response

The server responds in json format, with the following fields:

- categories: an array object with the categories to be used in the chart
- data: an array object with the data (and optionally the catagories for series creation)
- x_field_name: the name of the x field that has been defined in the config.json
- y_field_name: the name of the y field that has been defined in the config.json or the `name` variable from the request
- x_labels: show labels of x variable?

### Example response

```{
	"categories": [
		"2012-10-01 00:00:00"
		"2012-10-02 00:00:00"
		"2012-10-03 00:00:00"
		"2012-10-04 00:00:00"
		"2012-10-05 00:00:00"
	],
	"data": [
		["2012-10-01 00:00:00",6]
		["2012-10-02 00:00:00",5.3]
		["2012-10-03 00:00:00",5.8]
		["2012-10-04 00:00:00",5.6]
		["2012-10-05 00:00:00",4.8]
	],
	"x_field_name":"Group",
	"x_labels":true,
	"y_field_name":"Rating over time average"
}```


	
