{{ define "pagecontent" }}

<h1>Staff Hours</h1>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
<p><h2>Hello {{.Sess.User}}</h2></p>

<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}
{{if ne .Title "Current Staff Hours"}}
  <p><a id="backup" class="pinoylink" href="/desk/report_staff_hours">Show Current Staff Hours</a></p>
{{else}}
  <p><a id="backup" class="pinoylink" href="/manager/backup_staff_hours">Backup Staff Hours and Reset</a></p>
{{end}}

  <p><a id="backup" class="pinoylink" href="/desk/report_staff_hours?bkup=b">Show Previous Backup Staff Hours</a></p>

  <p><a id="backup" class="pinoylink" href="/desk/report_staff_hours?bkup=c">Show Oldest Backup Staff Hours</a></p>
{{end}}

<p><h3>{{.Title}}</h3></p>
{{if ne .BackupTime ""}}
<p>Backed up at {{.BackupTime}}</p>
{{end}}

<table>
<tr>
<th>User Id</th><th>Last Clock-in</th><th>Last Clock-out</th><th>Expected Hours</th><th>Total Expected Hours</th><th>Total Hours</th><th>Clock In</th><th>Clock Out</th>
</tr>
{{range .StaffHours}}
<tr>
<td>{{.UserID}}</td><td>{{.LastClockinTime}}</td><td>{{.LastClockoutTime}}</td><td>{{.ExpectedHours}}</td><td>{{.TotalExpectedHours}}</td><td>{{.TotalHours}}</td>
<td><a id="upd_staff_hours" class="button" href="/desk/upd_staff_hours?user={{.UserID}}&update=clockin">Clock in [{{.ClockInCnt}}]</a></td>
<td><a id="upd_staff_hours" class="button" href="/desk/upd_staff_hours?user={{.UserID}}&update=clockout">Clock out [{{.ClockOutCnt}}]</a></td>

</tr>
{{else}}
No staff hours to report
{{end}}
</table>

{{end}}
{{end}}
