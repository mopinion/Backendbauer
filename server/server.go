/*
Backendbauer
Floris Snuif
Copyright 2013, Mopinion BV Rotterdam
mopinionlabs.com
*/
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	auth "github.com/abbot/go-http-auth"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"io"
	"io/ioutil"
	//"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type Backendbauer struct {
	domain           string
	port             int
	mysql_host       string
	mysql_user       string
	mysql_pass       string
	mysql_db         string
	mysql_table      string
	mysql_date_field string
	y_field_name     string
	x_field_name     string
	max_items        int
	query            string
	path             string
	referer          string
	standardFilter   []filterType
	mongo_host       string
	mongo_port       int
	mongo_user       string
	mongo_pass       string
	mongo_database   string
	mongo_coll       string
}

type jsonobject struct {
	Object ObjectType
}

type ObjectType struct {
	Auth  []authType
	Mysql MysqlType
	Xvars []varsType
	Yvars []varsType
}

type authType struct {
	User     string
	Password string
}

type MysqlType struct {
	Servers []serversType
}

type serversType struct {
	Domain         string
	Port           int
	Host           string
	User           string
	Pass           string
	Db             string
	Table          string
	DateField      string
	StandardFilter []filterType
	MaxItems       int
	MongoHost      string
	MongoPort      int
	MongoUser      string
	MongoPass      string
	MongoDatabase  string
}

type varsType struct {
	Id        int
	Name      string
	FieldName string
	Type      string
	Values    []Values
	Join      []JoinType
	Select    string
}

type filterType struct {
	Field string
	Value string
}

type Values struct {
	Input  string
	Output string
}

type JoinType struct {
	Table string
	On    OnType
	Value string
}

type OnType struct {
	Left  string
	Right string
}

// mongoDB
type Result struct {
	Content string
}

func main() {
	// object
	bb := new(Backendbauer)
	// get path
	if runtime.GOOS == "linux" {
		// production path
		bb.path = "/var/www/backendbauer/server/"
	} else {
		// local test path
		bb.path = "./"
	}
	if bb.path == "" {
		_, filename, _, _ := runtime.Caller(1)
		path := strings.Split(filename, "/")
		path = path[0 : len(path)-1]
		path_str := ""
		for _, part := range path {
			path_str += part + "/"
		}
		bb.path = path_str
	}
	// http server root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Computer says no")
	})
	// data location
	settings := bb.settings()
	is_auth := settings.Object.Auth
	if len(is_auth) > 0 {
		authenticator := auth.NewBasicAuthenticator("Please, log in to use Backendbauer", password)
		http.HandleFunc("/data", auth.JustCheck(authenticator, request))
	} else {
		http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
			request(w, r)
		})
	}
	// test chart location
	http.HandleFunc("/chart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Content-Type", "text/html")
		bb.chart(w, r)
	})
	// js file location
	http.HandleFunc("/backendbauer.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Content-Type", "application/javascript")
		bb.js(w, r)
	})
	// port
	args := os.Args
	var port string
	if len(args) >= 2 {
		port = args[1]
	} else {
		port = ""
	}
	if port == "" {
		port = "8888"
	}
	http.ListenAndServe(":"+port, nil)
}

// authentication 
func password(user, realm string) string {
	bb := new(Backendbauer)
	settings := bb.settings()
	auth := settings.Object.Auth
	for _, pair := range auth {
		if pair.User == user {
			return pair.Password
		}
	}
	return ""
}

// load json
func (bb *Backendbauer) settings() jsonobject {
	file, err := ioutil.ReadFile(bb.path + "config.json")
	if err != nil {
		panic(err)
	}
	var jsontype jsonobject
	json.Unmarshal(file, &jsontype)
	return jsontype
}

// get server settings
func (bb *Backendbauer) serverSettings(referer string) {
	settings := bb.settings()
	server_settings := settings.Object.Mysql.Servers
	if referer != "" {
		referer_slice := strings.Split(referer, "/")
		referer = referer_slice[2]
	}
	var colon = regexp.MustCompile(`:`)
	if len(colon.FindAllString(referer, -1)) > 0 {
		referer_slice := strings.Split(referer, ":")
		referer = referer_slice[0]
	}
	for _, server := range server_settings {
		if server.Domain == referer {
			bb.domain = server.Domain
			bb.port = server.Port
			bb.mysql_host = server.Host
			bb.mysql_user = server.User
			bb.mysql_pass = server.Pass
			bb.mysql_db = server.Db
			bb.mysql_table = server.Table
			bb.mysql_date_field = server.DateField
			bb.standardFilter = server.StandardFilter
			bb.max_items = server.MaxItems
			bb.mongo_host = server.MongoHost + ":" + strconv.Itoa(server.MongoPort)
			bb.mongo_user = server.MongoUser
			bb.mongo_pass = server.MongoPass
			bb.mongo_database = server.MongoDatabase
			domain_slice := strings.Split(bb.domain, ".")
			subdom := domain_slice[0]
			bb.mongo_coll = subdom
		}
	}
}

// generate "js file" from js file
func (bb *Backendbauer) js(w http.ResponseWriter, r *http.Request) string {
	var part []byte
	var prefix bool
	var lines []string
	file, err := os.Open(bb.path + "backendbauer.js")
	if err != nil {
		// error
		fmt.Println("error in showing backendbauer.js: ", err)
	}
	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	// json in string form
	text := ""
	for _, line := range lines {
		text += line + "\n"
	}
	fmt.Fprint(w, text)
	return text
}

func request(w http.ResponseWriter, r *http.Request) {
	bb := new(Backendbauer)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Content-Type", "application/x-javascript")
	fmt.Println("Backendbauer server running")
	// check referer domain 
	referer := r.Referer()
	bb.referer = referer
	// server settings
	bb.serverSettings(referer)
	fmt.Println("referer: ", referer)
	fmt.Println("domain: ", bb.domain)
	var needed = regexp.MustCompile(bb.domain)
	if len(needed.FindAllString(referer, -1)) == 0 {
		fmt.Fprint(w, "computer says no")
	}
	// querystring input
	y_field, _ := strconv.Atoi(r.FormValue("y"))
	x_field, _ := strconv.Atoi(r.FormValue("x"))
	from_date := r.FormValue("from_date")
	to_date := r.FormValue("to_date")
	avg, _ := strconv.Atoi(r.FormValue("avg"))
	filter := r.FormValue("filter")
	chart_type := r.FormValue("chart_type")
	series := r.FormValue("series")
	jsonp, _ := strconv.ParseBool(r.FormValue("jsonp"))
	order := r.FormValue("order")
	limit := r.FormValue("limit")
	callback := r.FormValue("callback")
	combined, _ := strconv.ParseBool(r.FormValue("combined"))
	name := r.FormValue("name")
	benchmark, _ := strconv.Atoi(r.FormValue("benchmark"))
	mongo, _ := strconv.ParseBool(r.FormValue("mongo"))
	decimal, _ := strconv.Atoi(r.FormValue("decimal"))
	// round 1 (fight!)
	if decimal == 0 {
		decimal = 1
	}
	//coll := r.FormValue("coll")
	output := ""
	if mongo {
		// mongo
		//output = bb.mongoData()
		output = ""
	} else {
		// mysql
		output = bb.data(y_field, x_field, from_date, to_date, avg, filter, chart_type, series, jsonp, order, limit, callback, combined, name, benchmark, decimal)
	}
	// output
	fmt.Fprint(w, output)
}

func (bb *Backendbauer) data(y_field int, x_field int, from_date string, to_date string, avg int, filter string, chart_type string, series string, jsonp bool, order string, limit string, callback string, combined bool, name string, benchmark int, decimal int) string {
	if y_field == 0 {
		y_field = 1
	}
	if x_field == 0 {
		x_field = 1
	}
	if series == "" {
		series = "0"
	}
	var categories, data, output, group1 string
	categories = "["
	data = "["
	rows, y_field_settings, x_field_settings := bb.connect(y_field, x_field, from_date, to_date, avg, filter, order, limit, benchmark)
	for i, row := range rows {
		// get values from fields
		value := row.Str(1)
		group := row.Str(0)
		// round value if ratio and avg
		if y_field_settings.Type == "ratio" && avg == 1 {
			value_float, _ := strconv.ParseFloat(value, 64)
			value = strconv.FormatFloat(value_float, 'e', decimal, 64)
		}
		// remove time from date
		if x_field_settings.Type == "date" {
			group_split := strings.Split(group, " ")
			group1 = group_split[0]
		} else {
			group1 = group
		}
		// check for mapped value
		group1 = bb.mapValue(group1, x_field)
		if i == 0 {
			categories += `"` + group1 + `"`
			if chart_type == "pie" || combined == true {
				data += `["` + group1 + `",` + value + `]`
			} else {
				data += value
			}
		} else {
			categories += `,"` + group1 + `"`
			if chart_type == "pie" || combined == true {
				data += `,["` + group1 + `",` + value + `]`
			} else {
				data += "," + value
			}
		}
	}
	// show x-axis labels?
	var x_labels string
	if len(rows) <= bb.max_items && len(rows) > 1 {
		x_labels = `,"x_labels":true`
	} else {
		x_labels = `,"x_labels":false`
	}
	// create json
	categories += "]"
	data += "]"
	// y field custom name
	if name == "" {
		name = y_field_settings.Name
	}
	output = `{"categories":` + categories + `,"data":` + data + `,"x_field_name":"` + bb.x_field_name + `","y_field_name":"` + name + `"` + x_labels + `,"series":` + series + `}`
	if jsonp == true && callback != "" {
		output = callback + `(` + output + `);`
	}
	return output
}

func (bb *Backendbauer) mapValue(input string, x_field int) string {
	output := input
	settings := bb.settings()
	Xvars := settings.Object.Xvars
	for _, Xvar := range Xvars {
		if Xvar.Id == x_field {
			values := Xvar.Values
			for _, value := range values {
				if value.Input == input {
					output = value.Output
				}
			}
		}
	}
	return output
}

func (bb *Backendbauer) DB() mysql.Conn {
	db := mysql.New("tcp", "", bb.mysql_host, bb.mysql_user, bb.mysql_pass, bb.mysql_db)
	err := db.Connect()
	if err != nil {
		panic(err)
	}
	return db
}

func (bb *Backendbauer) connect(y_field int, x_field int, from_date string, to_date string, avg int, filter string, order string, limit string, benchmark int) ([]mysql.Row, varsType, varsType) {
	db := bb.DB()
	// get settings for fields
	y_field_settings := bb.fieldSettings("y", y_field)
	x_field_settings := bb.fieldSettings("x", x_field)
	bb.x_field_name = x_field_settings.Name
	bb.y_field_name = y_field_settings.Name
	var table = regexp.MustCompile("\\.")
	// filter for y var
	var y_field_filter string
	if avg == 2 {
		if len(table.FindAllString(y_field_settings.FieldName, -1)) == 0 && len(table.FindAllString(bb.mysql_table, -1)) == 0 {
			y_field_filter = `WHERE ` + bb.mysql_table + `.` + x_field_settings.FieldName + ` > 0`
		} else {
			y_field_filter = `WHERE ` + x_field_settings.FieldName + ` > 0`
		}
	} else if y_field_settings.Type == "ratio" {
		if len(table.FindAllString(y_field_settings.FieldName, -1)) == 0 && len(table.FindAllString(bb.mysql_table, -1)) == 0 {
			y_field_filter = `WHERE ` + bb.mysql_table + `.` + y_field_settings.FieldName + ` > 0`
		} else {
			y_field_filter = `WHERE ` + y_field_settings.FieldName + ` > 0`
		}
	} else {
		if len(table.FindAllString(y_field_settings.FieldName, -1)) == 0 && len(table.FindAllString(bb.mysql_table, -1)) == 0 {
			y_field_filter = `WHERE ` + bb.mysql_table + `.` + y_field_settings.FieldName + ` <> ""`
		} else {
			y_field_filter = `WHERE ` + y_field_settings.FieldName + ` <> ""`
		}
	}
	// date in right format
	var field_select string
	var field_group string
	if x_field_settings.Type == "date" {
		field_select = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%d-%m")`
		field_group = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%Y-%m-%d")`
	} else if x_field_settings.Type == "month" {
		field_select = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%m-%Y")`
		field_group = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%y-%m")`
	} else if x_field_settings.Type == "week" {
		field_select = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%v")`
		field_group = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%v")`
	} else if x_field_settings.Type == "day" {
		field_select = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%Y-%m-%d")`
		field_group = `DATE_FORMAT(` + x_field_settings.FieldName + `, "%Y-%m-%d")`
	} else {
		field_select = x_field_settings.FieldName
		field_group = x_field_settings.FieldName
	}
	// join tables
	var join_query string
	join := x_field_settings.Join
	for _, j := range join {
		join_table := j.Table
		join_on_left := j.On.Left
		join_on_right := j.On.Right
		join_value := j.Value
		if join_value != "" {
			field_select = join_value
		}
		if join_table != "" && join_on_left != "" && join_on_right != "" {
			if len(table.FindAllString(join_on_right, -1)) == 0 {
				join_query += ` JOIN ` + join_table + ` ON ` + join_table + `.` + join_on_right + ` = ` + join_on_left
			} else {
				join_query += ` JOIN ` + join_table + ` ON ` + join_on_right + ` = ` + join_on_left
			}
		}
	}
	// count, average or percentage? or benchmark? or custom select?
	var var1 string
	if benchmark != 0 {
		// benchmark
		var1 = `'` + strconv.Itoa(benchmark) + `'`
	} else if y_field_settings.Type == "custom" {
		// custom y field query
		var1 = y_field_settings.Select
	} else if avg == 1 && y_field_settings.Type == "ratio" {
		// average
		var1 = `AVG(` + y_field_settings.FieldName + `)`
	} else if avg == 2 {
		// percentage of total
		var1 = `count(CASE WHEN ` + y_field_settings.FieldName + ` = 1 THEN 1 END) / count(*) * 100`
	} else {
		// count
		//var1 = `COUNT(DISTINCT(` + y_field_settings.FieldName + `))`
		var1 = `COUNT(` + y_field_settings.FieldName + `)`
	}
	// standard filters
	standardFilters := bb.standardFilter
	standard_filter_query := ""
	for _, standardFilter := range standardFilters {
		// table given?
		if len(table.FindAllString(standardFilter.Field, -1)) == 0 {
			standard_filter_query += ` AND ` + bb.mysql_table + `.` + standardFilter.Field + ` = "` + standardFilter.Value + `"`
		} else {
			standard_filter_query += ` AND ` + standardFilter.Field + ` = "` + standardFilter.Value + `"`
		}
	}
	// extra filter
	big_nr := 9999
	extra_filter := filter
	extra_filter = strings.Replace(extra_filter, ":", " = ", big_nr)
	extra_filter = strings.Replace(extra_filter, "|", " AND ", big_nr)
	extra_filter = strings.Replace(extra_filter, ">:", " >= ", big_nr)
	extra_filter = strings.Replace(extra_filter, "<:", " <= ", big_nr)
	extra_filter = strings.Replace(extra_filter, "/", " OR ", big_nr)
	extra_filter = strings.Replace(extra_filter, `!~`, ` NOT LIKE `, big_nr)
	extra_filter = strings.Replace(extra_filter, "!", " <> ", big_nr)
	extra_filter = strings.Replace(extra_filter, `^`, `"`, big_nr)
	extra_filter = strings.Replace(extra_filter, `~`, ` LIKE `, big_nr)
	extra_filter = strings.Replace(extra_filter, `*`, `%`, big_nr)
	extra_filter = strings.Replace(extra_filter, `$`, `:`, big_nr)
	extra_filter = strings.Replace(extra_filter, `\`, `/`, big_nr)
	// date
	date_query := ""
	if len(table.FindAllString(bb.mysql_date_field, -1)) > 0 {
		if from_date != "" {
			date_query += `AND ` + bb.mysql_date_field + ` >= "` + from_date + ` 0:00:00"`
		}
		if to_date != "" {
			date_query += ` AND ` + bb.mysql_date_field + ` <= "` + to_date + ` 23:59:59"`
		}
	} else {
		if from_date != "" {
			date_query += `AND ` + bb.mysql_table + `.` + bb.mysql_date_field + ` >= "` + from_date + ` 0:00:00"`
		}
		if to_date != "" {
			date_query += ` AND ` + bb.mysql_table + `.` + bb.mysql_date_field + ` <= "` + to_date + ` 23:59:59"`
		}
	}
	// order
	var order_query string
	if order == "desc" {
		order_query = `ORDER BY ` + var1 + ` DESC`
	} else if order == "asc" {
		order_query = `ORDER BY ` + var1 + ` ASC`
	} else {
		order_query = ``
	}
	// limit
	var limit_query string
	if limit == "0" || limit == "" {
		limit_query = ``
	} else {
		limit_query = `LIMIT 0,` + limit
	}
	// query
	query := `
	SELECT ` + field_select + `, ` + var1 + `
	 FROM ` + bb.mysql_table + `
	 # join query
	 ` + join_query + `
	 # y field query
	 ` + y_field_filter + `
	 # standard filter query
	 ` + standard_filter_query + `
	 # extra filter query
	 ` + extra_filter + `
	 # date query
	 ` + date_query + `
	 GROUP BY ` + field_group + `
	  ` + order_query + `
	  ` + limit_query + `
	 `
	rows, _, err := db.Query(query)
	fmt.Println("query: ", query)
	bb.query = query
	if err != nil {
		panic(err)
	}
	// close again
	db.Close()
	return rows, y_field_settings, x_field_settings
}

func (bb *Backendbauer) fieldSettings(field_type string, field_id int) varsType {
	settings := bb.settings()
	var output varsType
	if field_type == "y" {
		field_settings := settings.Object.Yvars
		for _, field := range field_settings {
			if field.Id == field_id {
				output = field
			}
		}
	} else if field_type == "x" {
		field_settings := settings.Object.Xvars
		for _, field := range field_settings {
			if field.Id == field_id {
				output = field
			}
		}
	}
	return output
}

// MongoDB
/*
// data
func (bb *Backendbauer) mongoData() string {
	// db
	db := bb.MGOconnect()
	// collection
	c := db.C(bb.mongo_coll)
	// query
	Results := []Result{}
	err := c.Find(bson.M{}).All(&Results)
	if err != nil {
		fmt.Println("query error")
		panic(err)
	}
	var output string
	for _, row := range rows {
		output += row
	}
	return output
}

// connect to MongoDB
func (bb *Backendbauer) MGOconnect() *mgo.Database {
	//fmt.Println(WF.Mhost + ":" + strconv.Itoa(WF.Mport))
	// connection
	session, err := mgo.Dial(bb.mongo_host)
	if err != nil {
		fmt.Println("session error")
		panic(err)
	}
	//defer session.Close()
	// database
	db := session.DB(bb.mongo_database)
	// login
	err = db.Login(bb.mongo_user, bb.mongo_pass)
	if err != nil {
		fmt.Println("login error")
		panic(err)
	}
	return db
}
*/

// chart js
func (bb *Backendbauer) chart(w http.ResponseWriter, r *http.Request) {
	html := `
	<html>
		<head>
			<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js"></script>
			<script type="text/javascript" src="http://code.highcharts.com/highcharts.js"></script>
			<script src="http://code.highcharts.com/modules/exporting.js"></script>
			<title>Backendbauer chart test</title>
		</head>
		<body>
			<script type="text/javascript">
				// chart 1
				(function() {
					vars = {
						'container':'backendbauer_chart',
						'server':'franz:jawohl@localhost:8888',
						'from_field':'from',
						'to_field':'to',
						'debug':true,
						'jsonp':false,
						'combined':true,
						'on_load':true,
						'filter':'|my_table.rating!12',
						'charts': [
							{
								'id':1,
								'series':[
									{
										'y':1,
										'avg':1,
										'benchmark':6,
										'filter':'|(my_table.rating!11/my_table.rating!^12^)'
									},
									{
										'y':1,
										'avg':1,
										'filter':'|my_table.rating!11'
									}
								],
								'x':1,
								'type':'area',
								'title':'Rating'
							},
							{
								'id':2,
								'series':[
									{
										'y':1,
										'avg':0,
										'filter':'|my_table.rating!11'
									}
								],
								'x':2,
								'type':'bar',
								'title':'Groups'
							},
							{
								'id':3,
								'series':[
									{
										'y':2,
										'avg':2,
										'filter':'|my_table.rating!11'
									}
								],
								'x':1,
								'type':'line',
								'title':'Percentage promoters'
							},
							{
								'id':4,
								'series':[
									{
										'y':3,
										'filter':''
									}
								],
								'x':1,
								'type':'line',
								'title':'Special number over time'
							}
						]
					};
    				var bb = document.createElement('script'); 
    				bb.type = 'text/javascript'; 
    				bb.id = 'backendbauer';
    				bb.async = true;
    				bb.src = document.location.protocol + '//' + document.location.host + '/backendbauer.js';
    				document.getElementsByTagName('head')[0].appendChild(bb);
  				})();

			</script>
			<div style:"width:100%;overflow:hidden;">
				<div style="float:left;padding:10px;">
					<select id="loadChart" onchange="Backendbauer.render(this.value);">
						<option value="1">Rating over time</value>
						<option value="2">Groups</value>
						<option value="3">Percentage</value>
						<option value="4">Special number</value>
					</select>
				</div>
				<div style="float:left;padding:10px;">
					From
					<input type="text" value="01-04-2013" id="from" />
					<br />
					To
					<input type="text" value="30-04-2013" id="to" />
				</div>
			</div>
			<div style:"width:100%;overflow:hidden;">
				<div style="float:left;padding:10px;width:100%">
					<div id="backendbauer_chart" style="min-width: 400px; height: 400px; margin: 0 auto"></div>
				</div>
			</div>
	</body>
</html>
`
	fmt.Fprint(w, html)
}
