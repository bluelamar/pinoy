{{ define "pagecontent" }}

<h1>Room Status</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

{{if eq .Sess.Role "Manager"}}
    <p><a href="/manager/update_room">Add or Update Room and Set Room Rates</a></p>
{{end}}

<p>Table of rooms</p>
<table>
<tr>
<th>Room Number</th><th>Room Status</th><th>Registration</th><th>Guest Name</th><th>Guest Check in date and time</th>
</tr>
{{range .RoomTable}}
<tr>
<td>{{.Num}}</td><td>{{.Status}}</td><td><a id="registration" class="button" href="/desk/register?room={{.Num}}">Register</a></td><td>{{.GuestInfo}}</td><td>{{.CheckinTime}}</td>
</tr>
{{else}}
No Rooms to report
{{end}}
</table>
{{else if eq .Sess.Role "Staff"}}

  Staff form goes here - TODO

{{end}}

{{ end }}

{{end}}
