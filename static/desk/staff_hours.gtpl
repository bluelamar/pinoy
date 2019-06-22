{{ define "pagecontent" }}

<h1>Staff Hours</h1>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
<p><h2>Hello {{.Sess.User}}</h2></p>

<p><h3>{{.Sess.Role}} Page</h3></p>

<table>
<tr>
<th>User Id</th><th>Last Clock-in</th><th>Last Clock-out</th><th>Expected Hours</th><th>Total Hours</th><th>Clock In</th><th>Clock Out</th>
</tr>
{{range .StaffHours}}
<tr>
<td>{{.UserID}}</td><td>{{.LastClockinTime}}</td><td>{{.LastClockoutTime}}</td><td>{{.ExpectedHours}}</td><td>{{.TotalHours}}</td>
<td><a id="upd_staff_hours" class="button" href="/desk/upd_staff_hours?user={{.UserID}}&update=clockin">Clock in</a></td>
<td><a id="upd_staff_hours" class="button" href="/desk/upd_staff_hours?user={{.UserID}}&update=clockout">Clock out</a></td>

</tr>
{{else}}
No staff hours to report
{{end}}
</table>

{{end}}
{{end}}
