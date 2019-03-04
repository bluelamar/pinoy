{{ define "pagecontent" }}

<h1>Room Rates</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if eq .Sess.Role "Manager"}}
<p><a id="update_room_rates" class="button" href="/manager/upd_room_rate">Add Room Rate</a></p>

<p>Rate Classes</p>
<table>
<tr>
<th>Rate Class</th><th>3 Hour</th><th>6 hour</th><th>Extra Hour</th>
<th>Update Rate</th><th>Delete Rate</th>
</tr>
{{range .RateData}}
<tr>
<td>{{.Class}}</td><td>{{.Hour3}}</td><td>{{.Hour6}}</td><td>{{.Extra}}</td>
<td><a id="upd_rate" class="button" href="/manager/upd_room_rate?rate_class={{.Class}}&hour3={{.Hour3}}&hour6={{.Hour6}}&extra={{.Extra}}&update=true">Update Rate Class</a></td>
<td><a id="del_rate" class="button" href="/manager/upd_room_rate?rate_class={{.Class}}&hour3={{.Hour3}}&hour6={{.Hour6}}&extra={{.Extra}}&update=delete">Delete Rate Class</a></td>
</tr>
{{else}}
No food items to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
