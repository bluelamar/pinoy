{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Server Statistics</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}

<p><h3>Stats</h3></p>
<table>
{{range .StatsData}}
<tr>
<td>{{.Name}}</td>
<td>{{.Value}}</td>
</tr>

{{else}}
No service stats to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
