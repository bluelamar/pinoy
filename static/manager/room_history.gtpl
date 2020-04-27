{{ define "pagecontent" }}

<h1>Room History Report</h1>

{{if eq .Sess.Role "Manager"}}
<p><h2>Hello {{.Sess.User}}</h2></p>

<p><h3>{{.Sess.Role}} Page</h3></p>

{{if ne .Title "Current Room History"}}
  <p><a id="backup" class="pinoylink" href="/manager/report_room_history">Show Current Room History</a></p>
{{else}}
  <p><a id="backup" class="pinoylink" href="/manager/backup_room_history">Backup Room History and Reset</a></p>
{{end}}

<p><a id="backup" class="pinoylink" href="/manager/report_room_history?bkup=b">Show Previous Backup Room History</a></p>

<p><a id="backup" class="pinoylink" href="/manager/report_room_history?bkup=c">Show Oldest Backup Room History</a></p>

<p><h3>{{.Title}}</h3></p>
{{if ne .BackupTime ""}}
<p>Backed up at {{.BackupTime}}</p>
{{end}}

<table>
<tr>
<th>Room Number</th><th>Desk</th><th>Bell Hop</th><th>Activity</th><th>Time</th>
</tr>
{{range .RoomHistList}}
<tr>
<td>{{.Room}}</td><td>{{.Desk}}</td><td>{{.Bellhop}}</td><td>{{.Activity}}</td><td>{{.Timestamp}}</td>
</tr>
{{else}}
No room history to report
{{end}}
</table>

{{end}}
{{end}}
