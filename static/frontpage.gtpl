{{ define "pagecontent" }}

  <h1>Pinoy Front Desk</h1>
{{if and .Sess.Auth .Sess.User}}
  <p><h2>Hello {{.Sess.User}}</h2></p>
  <p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
  <nav>
    <ul>
      <li><a href="/desk/room_status?register=yes">Register Guest</a></li>
      <li><a href="/desk/room_status">Room Status</a></li>
      <li><a href="/desk/food">Food and Drink</a></li>
      {{if eq .Sess.Role "Manager"}}
      <li><a href="/manager/staff">Staff</a></li>
      <li><a href="/manager/upd_room?update=true">Add or Update Room</a></li>
      {{end}}
      <li class="mobile-nav-only"><a href="https://eventlogue.net/">Main Site</a></li>
    </ul>
  </nav>

{{end}}

{{ end }}
{{ end }}
