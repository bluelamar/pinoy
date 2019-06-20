{{ define "pagecontent" }}

  <h1>Front Desk</h1>
{{if and .Sess.Auth .Sess.User}}
  <p><h2>Hello {{.Sess.User}}</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
  <nav>
    <ul>
      <li><a href="/desk/room_status?register=open">Register Guest</a></li>
      <li><a href="/desk/room_status?register=booked">Booked Room Status</a></li>
      <li><a href="/desk/food">Order Food and Drink</a></li>
      <li><a href="/desk/report_staff_hours">Clock-in/Clock-out Staff</a></li>
      {{if eq .Sess.Role "Manager"}}
      <p>Managerial Actions</p>
      <li><a href="/manager/staff">Add or Update Staff</a></li>
      <li><a href="/manager/room_rates">Add or Update Room Rates</a></li>
      <li><a href="/manager/rooms">Add or Update Rooms</a></li>
      <li><a href="/manager/food">Add or Update Food Items</a></li>
      {{end}}
      <li class="mobile-nav-only"><a href="https://eventlogue.net/">Main Site</a></li>
    </ul>
  </nav>

{{end}}

{{ end }}
{{ end }}
