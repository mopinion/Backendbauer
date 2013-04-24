Backendbauer
============

## What it is

Backendbauer is a backend server with REST API for generating json data for frontend js charts.
It is written in Go and used by Mopinion in production.

### Highcharts
The charts library is Highcharts JS.
http://www.highcharts.com/

## How it works

Only two files are needed to get Backendbauer working:
- server/server.go
- server/config.json

For testing you can also use backendbauer.js

Server.go is the source code. In config.json you can set MySQL databases with custom tables and fields.
All settings are based on the url from which the API call is made, the referer.

# API

Endpoint:  
`[host name]:[port]/data`  




	
