{{ define "pagecontent" }}

<h1>Room Registration</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

  <form action="/desk/register" method="post">
    <legend>Check In</legend>
    <input placeholder="First Name" label="false" spellcheck="false" class="is-sensitive" value="" name="first_name" id="first_name" />
    <input placeholder="Last Name" label="false" spellcheck="false" class="is-sensitive"  value="" name="last_name" id="last_name" />
    <input placeholder="Duration" label="false" spellcheck="false" class="is-sensitive"  value="" name="duration" id="duration" />
    <input type="hidden" id="room_num" name="room_num" value="{{.RoomNum}}">
    <input type="submit" name="commit" value="Register" />
  </form>

{{end}}
{{end}}
{{end}}
