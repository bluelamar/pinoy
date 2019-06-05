{{ define "pagecontent" }}

<h1>Room Bell Hop</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

<p>Bell Hop</p>
  <form action="/desk/room_hop" method="post">
    <table>
    <tr>
    <td>Room</td><td>
    <input required placeholder="Room" label="false" spellcheck="false" class="is-sensitive" value="{{.RoomNum}}" name="room_num" id="room_num" />
    </td>
    </tr>
    <tr>
    <td>Bell Hop PIN</td><td>
    <input required type="password" placeholder="PIN" label="false" spellcheck="false" class="is-sensitive"  value="" name="bell_hop_pin" id="bell_hop_pin" />
    </td>
    </tr>
    </table>
    <input type="submit" name="commit" value="Submit" />
  </form>

{{end}}
{{end}}
