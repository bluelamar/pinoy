{{ define "pagecontent" }}

<h1>Add or Update Room</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/add_room" method="post">
  <legend>Room Details</legend>
  <table>
  <tr><td>Room Number</td><td>
  <input required placeholder="Room Number" id="room_num" name="room_num" value="{{.RoomNum}}">
  </td></tr>
  <tr><td>Number of Beds</td><td>
  <input required placeholder="Number of Beds" label="false" type="number" min="1" max="5" value="{{.NumBeds}}" name="num_beds" id="num_beds" />
  </td></tr>
  <tr><td>Bed Size</td><td>
  <input required placeholder="Bed Size" label="false" spellcheck="false" class="is-sensitive"  value="{{.BedSize}}" name="bed_size" id="bed_size" />
  </td></tr></table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
