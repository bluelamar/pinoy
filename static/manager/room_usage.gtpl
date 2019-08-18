{{ define "pagecontent" }}

<h1>Room Usage</h1>

{{if .Sess.Role "Manager"}}
<p><h2>Hello {{.Sess.User}}</h2></p>

<p><h3>{{.Sess.Role}} Page</h3></p>

<p><a id="backup" class="button" href="/manager/backup_room_usage">Backup Room Usage and Reset</a></p>

<p><a id="backup" class="button" href="/manager/report_room_usage?bkup=b">Show Previous Backup Room Usage</a></p>

<p><a id="backup" class="button" href="/manager/report_room_usage?bkup=c">Show Oldest Backup Room Usage</a></p>
{{if ne .Title "Current Room Usage"}}
  <p><a id="backup" class="button" href="/manager/report_room_usage">Show Current Room Usage</a></p>
{{end}}

<p><h3>{{.Title}}</h3></p>
{{if ne .BackupTime ""}}
<p>Backed up at {{.BackupTime}}</p>
{{end}}

<table>
<tr>
<th>Room Number</th><th>Total Number of Guests</th><th>Total Hours Used</th>
</tr>
{{range .RoomUsageList}}
<tr>
<td>{{.RoomNum}}</td><td>{{.TotNumGuests}}</td><td>{{.TotHours}}</td>
</tr>
{{else}}
No room usage to report
{{end}}
</table>

{{end}}
{{end}}
