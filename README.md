Backendbauer
============

## What it is

Backendbauer is a backend server with REST API for generating json data for frontend js charts.
It is written in Go and used by Mopinion.

## What it needs
* The [Go](http://golang.org) language
* The charts library from Highcharts JS.
http://www.highcharts.com/
* [MyMySQL] (http://github.com/ziutek/mymysql) for Go
* [Go http auth](https://github.com/abbot/go-http-auth)

## How it works

Only two files are needed to get Backendbauer working:
- server/server.go
- server/config.json

For testing you can also use server/backendbauer.js

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
or  
`nohup [backendbauer dir]/server/server &`

### MySQL
To test the example execute the `sql/my_database.sql` file on your local MySQL server.
Set the proper username and password in config.json

### Try
`sudo service backendbauer start`   
Go to   
`http://localhost:8888/chart`


## API

Endpoint:  
`https://[username]:[password]@[host name]:[port]/data`  

The API has a number of variables in order to get the right data in json format.

- x: data field id as configured in config.json
- y: data field id as configured in config.json
- from_date: start date for query in `YYYY-MM-DD` format  
- to_date: end date for query in `YYYY-MM-DD` format  
- avg: 0 -> count, 1 -> average or 2 -> percentage (of items that are 1 and not 0) `0/1/2`
- filter: a custom filter that is used to make the query.
Filters can be added to narrow the query down in the following manner:  
`|[field]:[value]` translates to `AND [field] = "[value]"`  
`|[field]![value]` translates to `AND [field] <> "[value]"`  
example:  
`filter=field1:right_value|field2!wrong_value`  
- chart_type (optional): `line` or `pie`, etc. The response will send the type back. This can be used in some cases.
- series (optional): `0/1` sometimes the ajax js code differs when a request is a series or a chart. The response returns this value.
- jsonp: If `true` and a callback function is specified, the response will add the callback function (needed for jsonp crossdomain/port calls)
- callback: if specified and `jsonp=true` the response will be included in a js callback (needed for jsonp crossdomain/port calls)
- order (optional): `asc` or `desc` if given the query will be ordered ascending or descending respectively on the x variable
- limit (optional): `[number]` for example `10` limits the result to 10 rows
- combined (optional): when `true` the response will include the categories in the data `[['category1','data1']['category2']['data2']]` instead of `['data1','data2','data3']`. Easier to add series after a chart already exists
- name (optional): the name of the y variable in the series. Response returns this name, so it can be used in js manipulation of the highchart object.
- benchmark (optional): fixed value to set the y variable to, for benchmarking.

### Example request

```html
http://franz:jawohl@localhost:8888/data?x=1&y=1&from_date=2013-04-01&to_date=2013-04-30&avg=1&filter=my_table.rating!12|my_table.rating!11&chart_type=area&series=0&jsonp=false&order=&limit=0&role=0&callback=Backendbauer.place&combined=true
```

## Response

The server responds in json format, with the following fields:

- categories: an array object with the categories to be used in the chart
- data: an array object with the data (and optionally the catagories for series creation)
- x_field_name: the name of the x field that has been defined in the config.json
- y_field_name: the name of the y field that has been defined in the config.json or the `name` variable from the request
- x_labels: show labels of x variable?

### Example response

```json
{
	"categories": [
		"2012-10-01"
		"2012-10-02"
		"2012-10-03"
		"2012-10-04"
		"2012-10-05"
	],
	"data": [
		["2012-10-01",6]
		["2012-10-02",5.3]
		["2012-10-03",5.8]
		["2012-10-04",5.6]
		["2012-10-05",4.8]
	],
	"x_field_name":"Date",
	"x_labels":true,
	"y_field_name":"Rating over time average"
}
```

## Future development
- more filters such as OR and LIKE
- other databases (such as MongoDB)
- API to easily add data to make fast charts of any process imaginable
	
