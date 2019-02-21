{{ define "welcome" }}
<html>
  <head>
    {{ template "header" . }}
  </head>
  <body>
<div>
{{if and .Sess .Sess.User}} Hi {{.Sess.User}} {{else}} Hi guest {{end}}
</div>

<div>
Hello {{.Sess.User}}
</div>

{{if and .Sess .Sess.Role }}
  {{if eq .Sess.Role "manager"}}

  my role is {{.Sess.Role}}
<nav>
  <ul>
    <li><a href="/manager/staff">Staff</a></li>
    <li><a href="/manager/room_status">Room Status</a></li>
    <li><a href="/manager/food">Food and Drink</a></li>
    <li class="mobile-nav-only"><a href="https://eventlogue.net/">Main Site</a></li>
  </ul>
</nav>
  {{else if eq .Sess.Role "desk"}}
Desk menu here
  {{else if eq .Sess.Role "staff"}}
Staff menu here
  {{end}}
{{else}}
<div class="account-buttons">
    <div id="entrance">
      <a id="sign-up" class="button" href="/register">Sign Up</a>
      <a id="log-in" class="button" href="/login">Log In</a>
    </div>
</div>

{{end}}
</body>
</html>
{{ end }}
