{{ define "pagecontent" }}

<h1>Cost and Bell Hop Assignment</h1>

{{if and .Sess .Sess.User}}
<p><h3>{{.Sess.User}} : {{.Sess.Role}} Page</h3></p>

{{if eq .OldCost ""}}
The cost for the room is {{.MonetarySymbol}} {{.Total}}
{{else}}
The new cost for the room is {{.MonetarySymbol}} {{.Total}}
The previous cost was {{.MonetarySymbol}} {{.OldCost}}
{{end}}

{{if eq .Repeat "false"}}
<p>
<div class="specialmsg">
Invalid PIN was specified, please try again:
</div>
</p>
{{end}}

<p>Bell Hop</p>
  <form action="/desk/room_hop" method="post">
    <table>
    <tr>
    <td>Room</td><td>
    <input required placeholder="Room" label="false" spellcheck="false" class="is-sensitive" value="{{.RoomNum}}" name="room_num" id="room_num" />
    </td>
    </tr>
    <tr>
    <td>Checkin Time</td><td>
    <input readonly required placeholder="Checkin Time" label="false" value="{{.CheckinTime}}" name="citime" id="citime" />
    </td>
    </tr>
    <tr>
    <td>Bell Hop PIN</td>
    <td>
    <select id="hopper" name="hopper" >
    {{range $element := .Hoppers}}
      {{if eq $element "none"}}
        <option selected>{{$element}}</option>
      {{else}}
        <option value="{{$element}}">{{$element}}</option>
      {{end}}
    {{end}}
    </select>
    </td>
    <td>
    <input placeholder="Input User ID if not in List" label="false" spellcheck="false" class="is-sensitive" value="" name="user_id" id="user_id" />
    </td>
    <td>
    <input required type="password" placeholder="PIN" label="false" spellcheck="false" class="is-sensitive" value="" name="bell_hop_pin" id="bell_hop_pin" />
    </td>
    </tr>
    </table>
    <input type="hidden" id="oldcost" name="oldcost" value="{{.OldCost}}" />
    <input type="hidden" id="total" name="total" value="{{.Total}}" />
    <input type="hidden" id="repeat" name="repeat" value="{{.Repeat}}" />
    <input type="submit" name="commit" value="Submit" />
  </form>

{{end}}
{{end}}
