{{ define "pagecontent" }}

<h1>Room Registration</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

  <form action="/desk/register" method="post">
    <legend>Check In</legend>
    <table>
    <tr><td>First Name</td><td>
    <input required placeholder="First Name" label="false" spellcheck="false" class="is-sensitive" value="" name="first_name" id="first_name" />
    </td></tr>
    <tr><td>Last Name</td><td>
    <input required placeholder="Last Name" label="false" spellcheck="false" class="is-sensitive"  value="" name="last_name" id="last_name" />
    </td></tr>
    <tr><td>Number of Guests</td><td>
    <input required placeholder="Number of Guests" id="num_guests" type="number" min="1" max="5" value="1" name="num_guests" />
    </td></tr>
    <tr><td>Duration</td><td>
    <select id="duration" name="duration" >
    {{range $element := .DurationOptions}}
      <option value="{{$element}}">{{$element}}</option>
    {{end}}
    </select>
    </td></tr>
    <tr><td>Room Number</td><td>
    <input readonly placeholder="Room Number" id="room_num" name="room_num" value="{{.RoomNum}}">
    </td></tr></table>
    <input type="submit" name="commit" value="Register" />
  </form>

{{end}}
{{end}}
{{end}}
