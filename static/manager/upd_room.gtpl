{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Add or Update Room</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/upd_room" method="post">
  <legend>Room Details</legend>
  <table>
  <tr><td>Room Number</td><td>
  <input required placeholder="Room Number" id="room_num" name="room_num" value="{{.Room.RoomNum}}">
  </td></tr>
  <tr><td>Number of Beds</td><td>
  <input required placeholder="Number of Beds" id="num_beds" type="number" min="1" max="5" value="{{.Room.NumBeds}}" name="num_beds" />
  </td></tr>
  <tr><td>Bed Size</td><td>
  <input required placeholder="Bed Size" id="bed_size" spellcheck="false" class="is-sensitive" value="{{.Room.BedSize}}" name="bed_size" />
  </td></tr>
  <tr><td>Room Rate</td><td>
  {{$rateClass := .Room.RateClass}}
  <select value="{{.Room.RateClass}}" id="room_rate" name="room_rate" >
  {{range $element := .RateClasses}}
    {{if eq $element $rateClass}}
    <option selected>{{$element}}</option>
    {{else}}
    <option value="{{$element}}">{{$element}}</option>
    {{end}}
  {{end}}
  </select>
  </td></tr>
  </table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
