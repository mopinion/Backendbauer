Backendbauer
============

## What it is

Backendbauer is a backend server with REST API for generating json data for frontend js charts.
It is written in Go and used by Mopinion in production.

## What it needs
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
`Go to server folder  
$ cd server  
build server  
$ sudo go build -o server server.go  
make symbolic link  
$ sudo ln -s [backendbauer dir]/server/server /usr/sbin/backendbauer  
copy service file to /etc/init.d  
$ sudo cp backendbauer /etc/init.d/  `

## API

Endpoint:  
`http://[host name]:[port]/data`  




	
