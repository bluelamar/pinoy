{{ define "pagecontent" }}

<h1>Room Usage Summary</h1>

{{if eq .Sess.Role "Manager"}}
<p><h2>Hello {{.Sess.User}}</h2></p>

<p><h3>{{.Sess.Role}} Page</h3></p>

{{if ne .Title "Current Room Usage"}}
  <p><a id="backup" class="pinoylink" href="/manager/report_room_usage">Show Current Room Usage</a></p>
{{else}}
  <p><a id="backup" class="pinoylink" href="/manager/backup_room_usage">Backup Room Usage and Reset</a></p>
{{end}}

<p><a id="backup" class="pinoylink" href="/manager/report_room_usage?bkup=b">Show Previous Backup Room Usage</a></p>

<p><a id="backup" class="pinoylink" href="/manager/report_room_usage?bkup=c">Show Oldest Backup Room Usage</a></p>

<p><h3>{{.Title}}</h3></p>
{{if ne .BackupTime ""}}
<p>Backed up at {{.BackupTime}}</p>
{{end}}

<table>
<tr>
<th>Room Number</th><th>Total Number of Guests</th><th>Total Hours Used</th><th>Number Times Occupied</th><th>Total Cost</th>
</tr>
{{range .RoomUsageList}}
<tr>
<td>{{.RoomNum}}</td><td>{{.TotNumGuests}}</td><td>{{printf "%.2f" .TotHours}}</td><td>{{.NumTimesOccupied}}</td><td>{{printf "%.2f" .TotCost}}</td>
</tr>
{{else}}
No room usage to report
{{end}}
</table>

{{end}}
{{end}}
