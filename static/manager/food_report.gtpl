{{ define "pagecontent" }}

<h1>Food Report</h1>

{{if eq .Sess.Role "Manager"}}
<p><h2>Hello {{.Sess.User}}</h2></p>

<p><h3>{{.Sess.Role}} Page</h3></p>

{{if ne .Title "Current Food Report"}}
  <p><a id="backup" class="pinoylink" href="/manager/report_food_usage">Show Current Food Report</a></p>
{{else}}
  <p><a id="backup" class="pinoylink" href="/manager/backup_food_usage">Backup Food Report and Reset</a></p>
{{end}}

<p><a id="backup" class="pinoylink" href="/manager/report_food_usage?bkup=b">Show Previous Backup Food Report</a></p>

<p><a id="backup" class="pinoylink" href="/manager/report_food_usage?bkup=c">Show Oldest Backup Food Report</a></p>

<p><h3>{{.Title}}</h3></p>
{{if ne .BackupTime ""}}
<p>Backed up at {{.BackupTime}}</p>
{{end}}

<table>
<tr>
<th>Food Item</th><th>Total Number of Orders</th><th>Total Cost</th>
</tr>
{{range .FoodUsageList}}
<tr>
<td>{{.ItemID}}</td><td>{{.TotOrders}}</td><td>{{printf "%.2f" .TotCost}}</td>
</tr>
{{else}}
No food report
{{end}}
</table>

{{end}}
{{end}}
