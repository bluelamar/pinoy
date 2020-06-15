{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Shift Settings</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}
<p><a id="update_shifts" class="pinoylink" href="/manager/upd_shiftinfo?update=add">Add Shift</a></p>

<p><h3>Shifts - 24 Hour Clock: 0 is Midnight and 12 is Noon</h3></p>

<table>
<tr>
<th>Shift Number</th><th>Start Hour</th><th> </th><th>End Hour</th><th> </th><th>Update</th><th>Delete</th>
</tr>

{{range .Shifts}}
<tr>
<td>{{.Shift}}</td>
<td>{{.StartTime}}</td>
<td>{{if lt .StartTime 12}}AM{{else}}PM{{end}}</td>
<td>{{.EndTime}}</td>
<td>{{if lt .EndTime 12}}AM{{else}}PM{{end}}</td>

<td><a id="upd_shift" class="button" href="/manager/upd_shiftinfo?shift={{.Shift}}&update=update">Update Shift</a></td>
<td><a id="del_shift" class="button" href="/manager/upd_shiftinfo?shift={{.Shift}}&update=delete">Delete Shift</a></td>
</tr>

{{else}}
No shift info to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
