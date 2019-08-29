{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Room Rates</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}
<p><a id="update_room_rates" class="pinoylink" href="/manager/upd_room_rate?update=add">Add Room Rate Class</a></p>

<p><h3>Rate Classes</h3></p>
<table>
{{range .RateData}}

<tr>
<td>{{.RateClass}}</td>
{{range .Rates}}
  <td>{{.TUnit}} : {{.Cost}}</td>
{{end}}

<td><a id="upd_rate" class="button" href="/manager/upd_room_rate?rate_class={{.RateClass}}&update=true">Update Rate</a></td>
<td><a id="del_rate" class="button" href="/manager/upd_room_rate?rate_class={{.RateClass}}&update=delete">Delete Rate</a></td>
</tr>

{{else}}
No room rate items to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
