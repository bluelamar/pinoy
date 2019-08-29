{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Rooms</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}
<p><a id="update_room" class="pinoylink" href="/manager/upd_room?update=add">Add Room</a></p>

<table>
<tr>
<th>Room Number</th><th>Number Beds</th><th>Bed Size</th><th>Rate Class</th><th>Update</th><th>Delete</th>
</tr>
{{range .Rooms}}

<tr>
<td>{{.RoomNum}}</td>
<td>{{.NumBeds}}</td>
<td>{{.BedSize}}</td>
<td>{{.RateClass}}</td>

<td><a id="upd_room" class="button" href="/manager/upd_room?room_num={{.RoomNum}}&update=true">Update Room</a></td>
<td><a id="del_room" class="button" href="/manager/upd_room?room_num={{.RoomNum}}&update=delete">Delete Room</a></td>
</tr>

{{else}}
No rooms to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
