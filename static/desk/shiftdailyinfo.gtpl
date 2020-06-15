{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Shift Settings</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

{{if or (ne .CurShift .Shift) (ne .CurDayOfYear .DayOfYear)}}
<p><a id="curshift" class="pinoylink" href="/desk/shiftdailyinfo?beforeShift=-1&beforeDay=-1">Show Current Shift</a></p>
{{end}}

<p><a id="prevshift" class="pinoylink" href="/desk/shiftdailyinfo?beforeShift={{.Shift}}&beforeDay={{.DayOfYear}}">Show Previous Shift</a></p>

Shift {{.Shift}} on 
{{.Month}}
{{if ne .Day .LastDay}}
{{.LastDay}} to
{{end}}
{{.Day}}

<p><h3>Shift Room Info</h3></p>

<table>
<tr>
<th>Shift Number</th><th>Type</th><th>Duration</th><th>Description</th><th>Count</th><th>Total</th>
</tr>

{{range .Rooms}}
<tr>
<td>{{.Shift}}</td>
<td>{{.Type}}</td>
<td>{{.Subtype}}</td>
<td>{{.Subtype2}}</td>
<td>{{.Volume}}</td>
<td>{{printf "%.2f" .Total}}</td>
</tr>

{{else}}
No shift room info to report
{{end}}
</table>

<p><h3>Shift Food Info</h3></p>

<table>
<tr>
<th>Shift Number</th><th>Type</th><th>Item</th><th>Description</th><th>Count</th><th>Total</th>
</tr>

{{range .Food}}
<tr>
<td>{{.Shift}}</td>
<td>{{.Type}}</td>
<td>{{.Subtype}}</td>
<td>{{.Subtype2}}</td>
<td>{{.Volume}}</td>
<td>{{printf "%.2f" .Total}}</td>
</tr>

{{else}}
No shift food info to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
