{{ define "pagecontent" }}
<html>
<head>
  {{ template "header" . }}
</head>
<body>
<h1>Pinoy Room Status</h1>

<div><p><a href="/frontpage">Front Page</a></p></div>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

{{if eq .Sess.Role "Manager"}}
    <p><a href="/manager/add_room">Add Room</a></p>
{{end}}

Table of rooms TODO

{{else if eq .Sess.Role "Staff"}}

  Staff form goes here - TODO

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
{{end}}
