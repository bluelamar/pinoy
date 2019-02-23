{{ define "pagecontent" }}

<h1>Room Registration</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

  <form action="/desk/register" method="post">
    <legend>Check In</legend>
    <table>
    <tr><td>First Name</td><td>
    <input required placeholder="First Name" label="false" spellcheck="false" class="is-sensitive" value="" name="first_name" id="first_name" />
    </td></tr><tr><td>Last Name</td><td>
    <input required placeholder="Last Name" label="false" spellcheck="false" class="is-sensitive"  value="" name="last_name" id="last_name" />
    </td></tr><tr><td>Duration</td><td>
    <input required placeholder="Duration" label="false" spellcheck="false" class="is-sensitive"  value="" name="duration" id="duration" />
    </td></tr><tr><td>Room Number</td><td>
    <input required placeholder="Room Number" id="room_num" name="room_num" value="{{.RoomNum}}">
    </td></tr></table>
    <input type="submit" name="commit" value="Register" />
  </form>

{{end}}
{{end}}
{{end}}
