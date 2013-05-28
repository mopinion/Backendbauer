/*
Backendbauer
Copyright 2013, Mopinion BV Rotterdam
*/

// object
var Backendbauer = function() {
	// global vars
	var charts, options, x, y, chart_type, avg, colors, title, series, from_date, to_date, filter, div, server_url, from_field, to_field, chart_id, order, standard_filter, benchmark;
	var series = 0;
	var debug = false;
	var jsonp = false;
	var limit = 0;
	var role = 0;
	var combined = false;
	var name = '';
	// methods
	return {
		// load frontend
		load:function(v) {
			// load vars
			standard_filter = v.filter+'|';
			charts = v.charts;
			div = v.container;
			server_url = v.server;
			from_field = v.from_field;
			to_field = v.to_field;
			debug = v.debug;
			jsonp = v.jsonp;
			if (typeof v.combined != "undefined") {
				combined = v.combined;
			}
			// loader
			Backendbauer.loader();
			// load on start
			Backendbauer.render(1);
		},
		// render chart
		render:function(id) {
			chart_id = id;
			for (i in charts) {
				if (charts[i]['id'] == id) {
					var series = charts[i]['series'];
					var x = charts[i]['x'];
					var chart_type = charts[i]['type'];
					var colors = charts[i]['colors'];
					var title = charts[i]['title'];
					var order = charts[i]['order'];
					var limit = charts[i]['limit'];
					var series_last = series.length - 1;
					if (standard_filter == undefined)  {
						standard_filter = '';
					}
					if (typeof charts[i]['role'] == "undefined") {
						role = 0;
					} else {
						role = charts[i]['role'];
					}
					for (j in series) {
						set_filters = standard_filter
						var y = series[j]['y'];
						var avg = series[j]['avg'];
						var benchmark = series[j]['benchmark'];
						var filters = series[j]['filters'];
						if (typeof series[j]['name'] == "undefined") {
							var name = '';
						} else {
							var name = series[j]['name'];
						}
						for (k in filters) {
							var field = filters[k]['field'];
							var value = filters[k]['value'];
							var value_not = filters[k]['not'];
							if (k == 0) {
								var sep = '';
							} else {
								var sep = '|';
							}
							if (value != undefined) {
								set_filters += sep+field+':'+value;
							}
							if (value_not != undefined) {
								set_filters += sep+field+'!'+value_not;
							}
						}
						if (j == 0) {
							var set_series = 0;
						} else {
							var set_series = 1;
						}
						if (set_series == 0) {
							Backendbauer.chart(x,y,chart_type,avg,colors,title,set_series,set_filters,order,limit,name,benchmark);
						} else {
							Backendbauer.series(x,y,chart_type,avg,colors,title,set_series,set_filters,order,limit,name,benchmark);
						}
						//Backendbauer.sleep(5000);
					}
					return false;
				}
			}
		},
		// reload chart
		reload:function() {
			Backendbauer.render(chart_id);
		},
		// get data
		chart:function(set_x,set_y,set_chart_type,set_avg,set_colors,set_title,set_series,set_filters,set_order,set_limit,set_name,set_benchmark) {
			// vars
			if (set_x == undefined) {
				set_x = 1;
			}
			if (set_y == undefined) {
				set_y = 1;
			}
			if (chart_type == undefined) {
				chart_type = 'line';
			}
			if (set_colors == undefined) {
				set_colors = ['#006dcc','#faa732','#da4f49','#5bb75b','#49afcd','#c09853','#468847','#b94a48','#3a87ad','#a9302a','#499249','#2a85a0'];
			}
			if (set_title != undefined) {
				title = set_title;
			}
			if (set_series == undefined) {
				set_series = 0;
			}
			if (set_filters != undefined) {
				filter = set_filters;
			} else {
				filter = '';
			}
			if (set_order != undefined) {
				order = set_order;
			} else {
				order = '';
			}
			if (set_limit != undefined) {
				limit = set_limit;
			}
			if (set_name != undefined) {
				name = set_name;
			}
			if (set_benchmark != undefined) {
				benchmark = set_benchmark;
			} else {
				benchmark = '';
			}
			x = set_x;
			y = set_y;
			chart_type = set_chart_type;
			avg = set_avg;
			colors = set_colors;
			series = set_series;
			// date
			from_date = Backendbauer.dateFormat(document.getElementById(from_field).value);
			to_date = Backendbauer.dateFormat(document.getElementById(to_field).value);
			// highcharts options
			options = {
				chart: {
					renderTo: div,
					type: chart_type,
					marginRight: 130,
					marginBottom: 50,
					zoomType: 'xy',
					borderRadius: 12
				},
				plotOptions: {
					line: {
						allowPointSelect: true,
						cursor: 'help',
						dataLabels: {
							enabled:0
						},
						series: {
							connectNulls: false
						},
						showInLegend: true
					}
				},
				title: {
					text: '',
					x: -20 //center
				},
				subtitle: {
					text: '',
					x: -20
				},
				xAxis: {
					title: {
						text:''
					},
					categories: [],
					labels: {
						enabled:false
					}
				},
				yAxis: {
					title: {
						text: ''
					},
					plotLines: [{
						value: 0,
						width: 1,
						color: '#808080'
					}]
				},
				tooltip: {
					formatter: function() {
						if (this.x != undefined && this.x != '') {
							return '<b>'+ this.series.name +'</b><br/>'+
							this.x +': '+ this.y +'';
						} else {
							return '<b>'+ this.series.name +'</b><br/>'+
							this.y; 
						}
					}
				},
				legend: {
					layout: 'vertical',
					align: 'right',
					verticalAlign: 'top',
					x: -10,
					y: 100,
					borderWidth: 0
				},
				series: [],
				colors:colors
			};
			// y-axis
			options.yAxis.title.text = title;
			if (jsonp == true) {
				Backendbauer.jsonp();
			} else {
				Backendbauer.ajax();
			}
		},
		// load series
		series:function(set_x,set_y,set_chart_type,set_avg,set_colors,set_title,set_series,set_filters,set_order,set_limit,set_name,set_benchmark) {
			// vars
			if (set_x == undefined) {
				set_x = 1;
			}
			if (set_y == undefined) {
				set_y = 1;
			}
			if (chart_type == undefined) {
				chart_type = 'line';
			}
			if (set_colors == undefined) {
				set_colors = ['#006dcc','#faa732','#da4f49','#5bb75b','#49afcd','#c09853','#468847','#b94a48','#3a87ad','#a9302a','#499249','#2a85a0'];
			}
			if (set_title != undefined) {
				title = set_title;
			}
			if (set_series == undefined) {
				set_series = 0;
			}
			if (set_filters != undefined) {
				filter = set_filters;
			} else {
				filter = '';
			}
			if (set_order != undefined) {
				order = set_order;
			} else {
				order = '';
			}
			if (set_limit != undefined) {
				limit = set_limit;
			}
			if (set_name != undefined) {
				name = set_name;
			}
			if (set_benchmark != undefined) {
				benchmark = set_benchmark;
			} else {
				benchmark = '';
			}
			x = set_x;
			y = set_y;
			chart_type = set_chart_type;
			avg = set_avg;
			colors = set_colors;
			series = set_series;
			// chart already loaded?
			if (typeof chart == "undefined") {
				var timer = setTimeout(function() {
					Backendbauer.series(set_x,set_y,set_chart_type,set_avg,set_colors,set_title,set_series,set_filters)
				},1000);
			} else {
				if (typeof timer != "undefined") {
					clearTimeout(timer);
				}
				if (jsonp == true) {
					Backendbauer.jsonp();
				} else {
					Backendbauer.ajax();
				}
			}
		},
		// form data querystring
		query:function() {
			var query = 'x='+x+'&y='+y+'&from_date='+from_date+'&to_date='+to_date+'&avg='+avg+'&filter='+filter+'&chart_type='+chart_type+'&series='+series+'&jsonp='+jsonp+'&order='+order+'&limit='+limit+'&role='+role+'&callback=Backendbauer.place&combined='+combined+'&name='+name+'&benchmark='+benchmark;
			if (debug == true) {
				Backendbauer.log(query);
			}
			return query;
		},
		// get json from server with ajax
		ajax:function() {
			if (typeof jQuery == "undefined") {
				Backendbauer.log("Please load jQuery first...");
				return false;
			}
			$.ajax({
				type: "GET",
				url: document.location.protocol+'//'+server_url+'/data',
				data: Backendbauer.query(),
				dataType: 'json',
				contentType: "application/json; charset=utf-8",
				success: function(data){
					Backendbauer.place(data);
				},
				error : function(data) {
					Backendbauer.log("Oops :-s Something went wrong..." + data);
				}
			});
		},
		// get json from server with jsonp
		jsonp:function() {
			var url = document.location.protocol+'//'+server_url+'/data?'+Backendbauer.query();
			if (document.getElementById('BBjsonp'+series)) {
				document.getElementsByTagName('head')[0].removeChild(document.getElementById('BBjsonp'+series));
			}
			var script = document.createElement('script');
    		script.setAttribute('type', 'text/javascript');
    		script.setAttribute('src', url);
    		script.setAttribute('id', 'BBjsonp'+series);
    		document.getElementsByTagName('head')[0].appendChild(script);
		},
		// place data in chart after jsonp
		place:function(data) {
			// debug response
			Backendbauer.log(data);
			series = data['series'];
			Backendbauer.log('series: '+series);
			if (series == 0) {
				// x-axis
				options.xAxis.categories = data['categories'];
				options.xAxis.title.text = data['x_field_name'];
				options.xAxis.labels.enabled = data['x_labels'];
				// make chart object
				chart = new Highcharts.Chart(options);
			}
			// series
			chart.addSeries({'name':data['y_field_name'],'data':data['data']});
		},
		// switch date format to Dutch
		dateFormat:function(date) {
			if (date == undefined || date == '') {
				return '';
			}
			date = date.split('-');
			date = date[2]+'-'+date[1]+'-'+date[0];
			return date;
		},
		// debugging
		log:function(msg) {
			if (debug == true) {
				try {
					console.log(msg);
				} catch(e) {
					// nothin'
				}
			}
		},
		// sleep
		// thank you, @stoyanstefanov
		sleep:function(milliseconds) {
			var start = new Date().getTime();
  			for (var i = 0; i < 1e7; i++) {
    			if ((new Date().getTime() - start) > milliseconds){
      				break;
    			}
  			}
		},
		loader:function() {
			/*
			document.getElementById(div).innerHTML = '<img src="http://backendbauer.com/assets/img/loader.gif" align="absmiddle" border="0">';
			*/
		}
	};
}();// end Backendbauer

//go...
new Backendbauer.load(vars);