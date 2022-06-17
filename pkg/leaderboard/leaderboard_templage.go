// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package leaderboard

const leaderboardTmpl = `<html>
<head>
    <title>{{ .Title }} - Leaderboard</title>
		{{ if .DisableCaching }}
    <meta http-equiv="CacheControl" content="no-cache, no-store, must-revalidate"/>
    <meta http-equiv="Pragma" content="no-cache"/>
    <meta http-equiv="Expires" content="0"/>
		{{ end }}
    <link rel="preconnect" href="https://fonts.gstatic.com">
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@300;400;600;700&display=swap" rel="stylesheet">
    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <script type="text/javascript">
        google.charts.load("current", {packages:["corechart"]});
    </script>
    <style>
    body {
       font-family: 'Open Sans', sans-serif;
       background-color: #f7f7fa;
       padding: 1em;
    }

    h1 {
      color: rgba(66,133,244);
      margin-bottom: 0em;
    }

    .subtitle {
      color: rgba(23,90,201);
      font-size: small;
    }

    pre {
        white-space: pre-wrap;
        word-wrap: break-word;
        color: #666;
        font-size: small;
    }

    h2.cli {
       color: #666;
    }

    h2 {
        color: #333;
    }

    .board p {
        font-size: small;
        color: #999;
        text-align: center;
    }


    .board {
        clear: right;
        display: inline-block;
        padding: 0.5em;
        margin: 0.5em;
        background-color: #fff;
    }
    .board:nth-child(4n+3) {
        border: 2px solid rgba(66,133,244,0.25);
        color: rgba(66,133,244);
    }

    .board:nth-child(4n+2) {
        border: 2px solid rgba(219,68,55,0.25);
        color: rgba rgba(219,68,55);
    }

    .board:nth-child(4n+1) {
        border: 2px solid rgba(244,160,0,0.25);
        color: rgba(244,160,0);
    }

    .board:nth-child(4n) {
        border: 2px solid rgba(15,157,88,0.25);
        color: rgba(15,157,88);
    }

    h3 {
        text-align: center;
    }

    </style>
</head>
<body>
    <h1>{{ .Title }}</h1>
    <div class="subtitle">{{.From}} &mdash; {{.Until}}</div>
{{ if not .HideCommand }}
    <h2 class="cli">Command-line</h2>
    <pre>{{.Command}}</pre>
{{ end }}
    {{ range .Categories }}
        <h2>{{ .Title }}</h2>

        {{ range .Charts }}
            <div class="board">
            <h3>{{ .Title }}</h3>
            <p>{{ .Metric }}</p>
            <div id="chart_{{ .ID }}" style="width: 450px; height: 350px;"></div>
            <script type="text/javascript">
                google.charts.setOnLoadCallback(draw{{ .ID}});

                function draw{{.ID}}() {
                    var data = new google.visualization.arrayToDataTable([
                    [{label:'{{.Object}}',type:'string'},{label: '{{.Metric}}', type: 'number'}, { role: 'annotation' }],
                    {{ range .Items }}["{{.Name}}", {{.Count}}, "{{.Count}}"],
                    {{ end }}
                    ]);

                    var options = {
                    axisTitlesPosition: 'none',

                    bars: 'horizontal', // Required for Material Bar Charts.
                    axes: {
                        x: {
                        y: { side: 'top'} // Top x-axis.
                        }
                    },
                    legend: { position: "none" },
                    bar: { groupWidth: "85%" }
                    };

                   var chart = new google.visualization.BarChart(document.getElementById('chart_{{.ID }}'));
                   chart.draw(data, options);
                };
            </script>
            </div>
        {{ end }}
    {{ end}}
</body>
</html>
`
