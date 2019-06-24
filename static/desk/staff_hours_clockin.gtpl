{{ define "pagecontent" }}

<h1>Staff Clock-in</h1>

{{if and .Sess .Sess.User}}
<p><h3>{{.Sess.User}} : {{.Sess.Role}} Page</h3></p>

<p>Staff</p>
  <form action="/desk/upd_staff_hours" method="post">
    <table>
    <tr>
    <td>Staff ID</td><td>
    <input required placeholder="Staff" label="false" spellcheck="false" class="is-sensitive" value="{{.Emp.UserID}}" name="staffid" id="staffid" />
    </td>
    </tr>
    <tr>
    <td>Expected Hours to Work</td>
    <td>
<input required placeholder="Number" id="hours" type="number" min="1" max="12" value="8" name="hours" />
    </td>
    </tr>
    </table>
    <input type="submit" name="commit" value="Submit" />
  </form>

{{end}}
{{end}}
