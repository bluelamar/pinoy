{{ define "pagecontent" }}

<h1>Room Status</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}} : {{.Sess.Role}}</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

{{$role:=.Sess.Role}}

{{if .OpenRoomsOnly}}
<p><h3>Table of <i>Open</i> rooms</h3></p>
{{else}}
<p><h3>Table of <i>Booked</i> rooms</h3></p>
{{end}}
<table>
<tr>
<th>Room Number</th><th>Room Status</th><th>Room Rate</th>
{{if .OpenRoomsOnly}}
  <th>Register</th>
{{else}}
  <th>Guest Name</th><th>Cost ({{.MonetarySymbol}})</th><th>Number of Guests</th><th>Extra Guests</th><th>Hourly Rate per Extra Guest</th><th>Duration</th><th>Overtime count</th>
  <th>Check in time</th><th>Check out time</th><th>Check Out</th><th>Purchase Items</th><th>Update</th>
{{end}}
</tr>
{{range .Rooms}}
<tr>
<td>{{.RoomNum}}</td><td>{{.Status}}</td><td>{{.Rate}}</td>
{{if eq .Status "open"}}
  <td><a id="registration" class="button" href="/desk/register?room={{.RoomNum}}&rate={{.Rate}}&reg=checkin">Register</a></td>
{{else}}
  <td>{{.GuestInfo}}</td><td>{{.Cost}}</td><td>{{.NumGuests}}</td><td>{{.NumExtraGuests}}</td><td>{{.ExtraRate}}</td><td>{{.Duration}}</td><td>{{.Overtime}}</td><td>{{.CheckinTime}}</td><td>{{.CheckoutTime}}</td>
  <td><a id="registration" class="button" href="/desk/register?room={{.RoomNum}}&reg=checkout">Check Out</a></td>
  <td><a id="purchase" class="button" href="/desk/food?room={{.RoomNum}}">Purchase</a></td>
  <td><a id="update" class="button" href="/desk/register?room={{.RoomNum}}&reg=update&rate={{.Rate}}">Update</a></td>
{{end}}

</tr>
{{else}}
No Rooms to report
{{end}}
</table>

{{end}}

{{ end }}

{{end}}
