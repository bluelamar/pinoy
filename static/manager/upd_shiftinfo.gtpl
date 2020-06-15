{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Add or Update Shift</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}

{{range .Shifts}}

<form action="/manager/upd_shiftinfo" method="post">
  <legend>Shift Details - 24 Hour Clock settings: 0 is midnight and 12 is noon</legend>
  <table>
  <tr><td>Shift Number</td><td>
    <input required placeholder="Shift-number" id="shiftid" name="shiftid" type="number" min="1" max="6" value="{{.Shift}}" />
  </td></tr>
  <tr><td>Start Hour</td><td>
    <input required id="starthour" name="starthour" type="number" min="0" max="23" value="{{.StartTime}}" />
  </td></tr>
  <tr><td>End Hour</td><td>
    <input required id="endhour" name="endhour" type="number" min="0" max="23" value="{{.EndTime}}" />
  </td></tr>

  </table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
{{end}}
