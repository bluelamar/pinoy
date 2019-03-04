{{ define "pagecontent" }}

<h1>Add or Update Room Rate</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/upd_room_rate" method="post">
  <legend>Room Rate Details</legend>
  <table>
  <tr><td>Room Rate Class</td><td>
  <input required placeholder="Room Rate Class" id="rate_class" name="rate_class" value="{{.RateData.Class}}">
  </td></tr>
  <tr><td>3 Hour Rate</td><td>
  <input required placeholder="3 Hour Rate" id="hour3" name="hour3" value="{{.RateData.Hour3}}" />
  </td></tr>
  <tr><td>6 Hour Rate</td><td>
  <input required placeholder="6 Hour Rate" id="hour6" name="hour6" spellcheck="false" class="is-sensitive" value="{{.RateData.Hour6}}" />
  </td></tr>
  <tr><td>Extra Hour Rate</td><td>
  <input required placeholder="Extra Hour Rate" id="extra_rate" name="extra_rate" spellcheck="false" class="is-sensitive" value="{{.RateData.Extra}}" />
  </td></tr>
  </table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
