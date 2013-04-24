/*
Backendbauer
Floris Snuif
Mopinion BV Rotterdam
*/
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

/*
type Backendbauerer interface {
	data()
}
*/

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
}

type jsonobject struct {
	Object ObjectType
}

type ObjectType struct {
	Mysql MysqlType
	Xvars []varsType
	Yvars []varsType
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
}

type varsType struct {
	Id        int
	Name      string
	FieldName string
	Type      string
	Values    []Values
	Join      []JoinType
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

func main() {
	// object
	bb := new(Backendbauer)
	// get path
	if runtime.GOOS == "linux" {
		bb.path = "/var/www/backendbauer/"
	} else {
		bb.path = "/conceptables/backendbauer/go/"
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
	// http server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Computer says no")
	})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Content-Type", "application/x-javascript")
		bb.data(w, r)
	})
	http.HandleFunc("/chart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Content-Type", "text/html")
		bb.chart(w, r)
	})
	http.HandleFunc("/backendbauer.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Content-Type", "application/javascript")
		bb.js(w, r)
	})
	http.ListenAndServe(":8888", nil)
}

// load json
func (bb *Backendbauer) settings() jsonobject {
	file, err := ioutil.ReadFile(bb.path + "config.json")
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s\n", string(file))
	var jsontype jsonobject
	json.Unmarshal(file, &jsontype)
	//fmt.Printf("Results: %v\n", jsontype)
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

func (bb *Backendbauer) data(w http.ResponseWriter, r *http.Request) int {
	fmt.Println("Backendbauer server running")
	// check referer domain 
	referer := r.Referer()
	bb.referer = referer
	//remote_addr := r.RemoteAddr
	// server settings
	bb.serverSettings(referer)
	fmt.Println("referer: ", referer)
	fmt.Println("domain: ", bb.domain)
	var needed = regexp.MustCompile(bb.domain)
	if len(needed.FindAllString(referer, -1)) == 0 || referer == "" {
		fmt.Fprint(w, "computer says no")
		return 1
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
	role_id, _ := strconv.Atoi(r.FormValue("role"))
	callback := r.FormValue("callback")
	combined, _ := strconv.ParseBool(r.FormValue("combined"))
	name := r.FormValue("name")
	if y_field == 0 {
		y_field = 1
	}
	if x_field == 0 {
		x_field = 1
	}
	if from_date == "" {
		from_date = "2012-01-01"
	}
	if to_date == "" {
		to_date = "2025-12-31"
	}
	//fmt.Fprint(w, "data_field_id: ", myval)
	//data_field_id, _ := strconv.Atoi(y)
	var categories, data, output, group1 string
	categories = "["
	data = "["
	rows, y_field_settings, x_field_settings := bb.connect(y_field, x_field, from_date, to_date, avg, filter, order, limit, role_id)
	for i, row := range rows {
		// get values from fields
		value := row.Str(1)
		group := row.Str(0)
		// round value if ratio and avg
		if y_field_settings.Type == "ratio" && avg == 1 {
			value_float, _ := strconv.ParseFloat(value, 64)
			value = strconv.FormatFloat(value_float, 'e', 1, 64)
		}
		// remove time from date
		if x_field_settings.Type == "date" {
			group_split := strings.Split(group, " ")
			group1 = group_split[0]
		} else {
			group1 = group
		}
		//fmt.Fprint(w, "\n", group, ": ", value)
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
	if len(rows) <= bb.max_items {
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
	fmt.Fprint(w, output)
	return 0
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
	//fmt.Println("host: ", bb.mysql_host)
	db := mysql.New("tcp", "", bb.mysql_host, bb.mysql_user, bb.mysql_pass, bb.mysql_db)

	err := db.Connect()
	if err != nil {
		panic(err)
	}
	return db
}

func (bb *Backendbauer) connect(y_field int, x_field int, from_date string, to_date string, avg int, filter string, order string, limit string, role_id int) ([]mysql.Row, varsType, varsType) {
	db := bb.DB()
	// vars
	//y_field_str := strconv.Itoa(y_field)
	//x_field_str := strconv.Itoa(x_field)
	// get settings for fields
	y_field_settings := bb.fieldSettings("y", y_field)
	x_field_settings := bb.fieldSettings("x", x_field)
	bb.x_field_name = x_field_settings.Name
	bb.y_field_name = y_field_settings.Name
	var table = regexp.MustCompile("\\.")
	// filter for y var
	var y_field_filter string
	fmt.Println("y_field_settings.FieldName: ", y_field_settings.FieldName)
	fmt.Println("bb.mysql_table: ", bb.mysql_table)
	if y_field_settings.Type == "ratio" {
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
	} else {
		field_select = x_field_settings.FieldName
		field_group = x_field_settings.FieldName
	}
	// join tables
	var join_query string
	join := x_field_settings.Join
	fmt.Println("join: ", join)
	for _, j := range join {
		join_table := j.Table
		join_on_left := j.On.Left
		join_on_right := j.On.Right
		join_value := j.Value
		if join_value != "" {
			field_select = join_value
		}
		if join_table != "" && join_on_left != "" && join_on_right != "" {
			if len(table.FindAllString(y_field_settings.FieldName, -1)) == 0 {
				join_query += ` JOIN ` + join_table + ` ON ` + join_table + `.` + join_on_right + ` = ` + join_on_left
			} else {
				join_query += ` JOIN ` + join_table + ` ON ` + join_on_right + ` = ` + join_on_left
			}
		}
	}
	// count or average?
	var var1 string
	if avg == 1 && y_field_settings.Type == "ratio" {
		var1 = `AVG(` + y_field_settings.FieldName + `)`
	} else {
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
	// extra filters
	filter_slices := strings.Split(filter, "|")
	extra_filter := ""
	var fs_field, fs_value, fs_sign string
	for _, filter_slice := range filter_slices {
		// AND
		var colon = regexp.MustCompile(":")
		if len(colon.FindAllString(filter_slice, -1)) > 0 {
			fs := strings.Split(filter_slice, ":")
			fs_field = fs[0]
			fs_value = fs[1]
			fs_sign = `=`
		}
		// AND NOT
		var exclam = regexp.MustCompile("!")
		if len(exclam.FindAllString(filter_slice, -1)) > 0 {
			fs := strings.Split(filter_slice, "!")
			fs_field = fs[0]
			fs_value = fs[1]
			fs_sign = `<>`
		}
		// table given?
		if len(table.FindAllString(fs_field, -1)) > 0 {
			extra_filter += ` AND ` + fs_field + ` ` + fs_sign + ` "` + fs_value + `"`
		} else {
			extra_filter += ` AND ` + bb.mysql_table + `.` + fs_field + ` ` + fs_sign + ` "` + fs_value + `"`
		}
	}
	// date
	// table given?
	var date_query string
	if len(table.FindAllString(fs_field, -1)) > 0 {
		date_query = `AND ` + bb.mysql_date_field + ` >= "` + from_date + ` 0:00:00"
	  AND ` + bb.mysql_date_field + ` <= "` + to_date + ` 23:59:59"`
	} else {
		date_query = `AND ` + bb.mysql_table + `.` + bb.mysql_date_field + ` >= "` + from_date + ` 0:00:00"
	  AND ` + bb.mysql_table + `.` + bb.mysql_date_field + ` <= "` + to_date + ` 23:59:59"`
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
	// role query (for EFM)
	// role_id
	var role_query string
	if role_id == 0 {
		role_query = ""
	} else {
		role := bb.childQuery(role_id, "")
		role_query = "AND (f.role_id = " + strconv.Itoa(role_id) + " " + role + ")"
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
	 # role query
	 ` + role_query + `
	 GROUP BY ` + field_group + `
	  ` + order_query + `
	  ` + limit_query + `
	 `
	fmt.Println("-->", query, "<--")
	rows, _, err := db.Query(query)
	bb.query = query
	if err != nil {
		panic(err)
	}
	return rows, y_field_settings, x_field_settings
}

func (bb *Backendbauer) fieldSettings(field_type string, field_id int) varsType {
	settings := bb.settings()
	var output varsType
	if field_type == "y" {
		field_settings := settings.Object.Yvars
		//fmt.Println(field_settings)
		for _, field := range field_settings {
			if field.Id == field_id {
				output = field
			}
		}
	} else if field_type == "x" {
		field_settings := settings.Object.Xvars
		//fmt.Println(field_settings)
		for _, field := range field_settings {
			if field.Id == field_id {
				output = field
			}
		}
	}
	return output
}

// roles tree
func (bb *Backendbauer) getRolesTree(parent_id int, children []int) []int {
	db := bb.DB()
	query := `SELECT id FROM
	roles
	WHERE parent_id = ` + strconv.Itoa(parent_id)
	rows, _, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	//fmt.Println("roles")
	//fmt.Println(children)
	for _, row := range rows {
		child_id := row.Int(0)
		//children[i] = child_id
		children = append(children, child_id)
		children = bb.getRolesTree(child_id, children)
	}
	return children
}

func (bb *Backendbauer) childQuery(role_id int, table string) string {
	if table == "" {
		table = "f.role_id"
	}
	var childr []int
	children := bb.getRolesTree(role_id, childr)
	fmt.Println("query")
	fmt.Println(children)
	query := ""
	for _, child := range children {
		query += " OR " + table + " = " + strconv.Itoa(child)
	}
	return query
}

func (bb *Backendbauer) chart(w http.ResponseWriter, r *http.Request) {
	html := `
	<html>
		<head>
			<script type="text/javascript" src="http://backendbauer.com/assets/js/jquery.js"></script>
			<script type="text/javascript" src="http://backendbauer.com/assets/js/highcharts.latest.js"></script>
			<script src="http://code.highcharts.com/modules/exporting.js"></script>
			<title>Backendbauer chart test</title>
		</head>
		<body>
			<script type="text/javascript">
				// chart 1
				(function() {
					vars = {
						'container':'backendbauer_chart',
						'server':'localhost:8888',
						'from_field':'from',
						'to_field':'to',
						'debug':true,
						'jsonp':false,
						'combined':true,
						'charts': [
							{
								'id':1,
								'series':[
									{
										'y':1,
										'name':'Aantal items',
										'avg':0,
										'filters':[
											{
												'field':'fv.data_field_id',
												'value':391
											},
											{
												'field':'fv2.data_field_id',
												'value':392
											},
											{
												'field':'fv3.data_field_id',
												'value':388
											},
											{
												'field':'fv3.value',
												'value':12
											}
										]
									},
									{
										'y':1,
										'avg':0,
										'filters':[
											{
												'field':'fv.data_field_id',
												'value':391
											},
											{
												'field':'fv2.data_field_id',
												'value':392
											},
											{
												'field':'fv3.data_field_id',
												'value':388
											}
										]
									}
								],
								'x':1,
								'type':'area',
								'title':'Testing',
								//'role':217
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
						<option value="1">Test</value>
					</select>
				</div>
				<div style="float:left;padding:10px;">
					Van
					<input type="text" value="01-01-2012" id="from" />
					<br />
					Tot
					<input type="text" value="31-12-2012" id="to" />
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
