{{ define "pagecontent" }}

<h1>Checkout Bell Hop Assignment</h1>

{{if and .Sess .Sess.User}}
<p><h3>{{.Sess.User}} : {{.Sess.Role}} Page</h3></p>

{{if ne .OldCost ""}}
The original charge for the room was {{.MonetarySymbol}} {{.OldCost}}
The over time charge is {{.MonetarySymbol}} {{.OverCost}}
{{end}}
Total charge for the room is {{.MonetarySymbol}} {{.Total}}


<p>Choose Room Attendant</p>
  <form action="/desk/checkout" method="post">
    <table>
    <tr>
    <td>Room</td><td>
    <input readonly placeholder="Room" label="false" spellcheck="false" class="is-sensitive" value="{{.RoomNum}}" name="room_num" id="room_num" />
    </td>
    </tr>
    <tr>
    <td>Checkin Time</td><td>
    <input readonly required placeholder="Checkin Time" label="false" value="{{.CheckinTime}}" name="citime" id="citime" />
    </td>
    </tr>
    <tr>
    <td>Checkout Time</td><td>
    <input readonly required placeholder="Checkout Time" label="false" value="{{.CheckoutTime}}" name="cotime" id="cotime" />
    </td>
    </tr>
    <tr>
    <td>Room Attendant</td>
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
    </tr>
    </table>
{{if ne .OldCost ""}}
    <input type="hidden" id="oldcost" name="oldcost" value="{{.OldCost}}" />
    <input type="hidden" id="overcost" name="overcost" value="{{.OverCost}}" />
{{end}}
    <input type="hidden" id="total" name="total" value="{{.Total}}" />
    <input type="submit" name="commit" value="Submit" />
  </form>

{{end}}
{{end}}
