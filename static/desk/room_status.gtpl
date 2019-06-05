{{ define "pagecontent" }}

<h1>Room Status</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

{{if eq .Sess.Role "Manager"}}
    <p><a href="/manager/upd_room">Add Room</a></p>
{{end}}

{{$role:=.Sess.Role}}

<p>Table of rooms</p>
<table>
<tr>
<th>Room Number</th><th>Room Status</th><th>Room Rate</th><th>Registration</th><th>Guest Name</th>
<th>Guest Check in time</th><th>Purchase Items</th>
{{if eq .Sess.Role "Manager"}}
<th>Update Room</th><th>Delete Room</th>
{{end}}
</tr>
{{range .RoomTable}}
<tr>
<td>{{.Num}}</td><td>{{.Status}}</td><td>{{.Rate}}</td>
<td><a id="registration" class="button" href="/desk/register?room={{.Num}}">Register</a></td>
<td>{{.GuestInfo}}</td><td>{{.CheckinTime}}</td>
<td><a id="purchase" class="button" href="/desk/food?room={{.Num}}">Purchase</a></td>
{{if eq $role "Manager"}}
<td><a id="upd_room" class="button" href="/manager/upd_room?room={{.Num}}&update=true">Update Room</a></td>
<td><a id="del_room" class="button" href="/manager/upd_room?room={{.Num}}&update=delete">Delete Room</a></td>
{{end}}
</tr>
{{else}}
No Rooms to report
{{end}}
</table>

{{end}}

{{ end }}

{{end}}
