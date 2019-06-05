{{ define "pagecontent" }}

<h1>Staff</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}
<p><a id="update_staff" class="button" href="/manager/add_staff">Add New Employee</a></p>

<table>
<tr>
<th>User Id</th><th>Role</th><th>Last</th><th>First</th><th>Middle</th><th>Salary</th><th>Update</th><th>Delete</th>
</tr>
{{range .Staff}}
<tr>
<td>{{.Name}}</td><td>{{.Role}}</td><td>{{.Last}}</td><td>{{.First}}</td><td>{{.Middle}}</td><td>{{.Salary}}</td>
<td><a id="upd_staff" class="button" href="/manager/upd_staff?id={{.Name}}&role={{.Role}}&last={{.Last}}&first={{.First}}&middle={{.Middle}}&salary={{.Salary}}">Update</a></td>
<td><a id="del_staff" class="button" href="/manager/upd_staff?id={{.Name}}&role={{.Role}}&last={{.Last}}&first={{.First}}&middle={{.Middle}}&salary={{.Salary}}&update=delete">Delete</a></td>

</tr>
{{else}}
No staff to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
