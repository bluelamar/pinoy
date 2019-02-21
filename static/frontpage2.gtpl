<html>
<head>
  {{ template "header" . }}
</head>
<body>
    <h1>Pinoy Front Desk</h1>
{{if and .Sess.Auth .Sess.User}}
    <p><h2>Hello {{.Sess.User}}</h2></p>
    <p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
  <nav>
    <ul>
      <li><a href="/desk/room_status">Room Status</a></li>
      <li><a href="/desk/food">Food and Drink</a></li>
      <li><a href="/desk/staff">Staff</a></li>
      <li class="mobile-nav-only"><a href="https://eventlogue.net/">Main Site</a></li>
    </ul>
  </nav>

{{else if eq .Sess.Role "Staff"}}

  Staff form goes here

{{end}}

{{else}}

    <p>Please sign-in</p>
    <div class="account-buttons">
        <div id="entrance">
          <p><a id="log-in" class="button" href="/login">Log In</a></p>
        </div>
    </div>
{{ end }}
</body>
</html>
