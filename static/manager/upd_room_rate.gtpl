{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Add or Update Room Rate</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/upd_room_rate" method="post">
  <legend>Room Rate Details</legend>
  <table>
  <tr><td>Room Rate Class</td><td>
  <input required placeholder="Ex: Small Room" id="rate_class" name="rate_class" value="{{.RateData.RateClass}}">
  </td></tr>

  {{if .RateData.Rates}}
  {{range .RateData.Rates}}
    <td>{{.TUnit}} : {{.Cost}}</td>
  {{end}}
  {{end}}

  <tr><td>New Rate Time Unit</td><td>
  <input required placeholder="Ex: Daily" id="new_rate_time_unit" name="new_rate_time_unit" spellcheck="false" value="" />
  </td></tr>
  <tr><td>New Rate Cost per Time Unit</td><td>
  <input required placeholder="Ex: $50" id="new_rate_cost" name="new_rate_cost" spellcheck="false" value="" />
  </td></tr>
  </table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
