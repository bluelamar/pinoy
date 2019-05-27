{{ define "pagecontent" }}

<h1>Add or Update Employee</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/upd_staff" method="post">
  <legend>Employee Details</legend>
  <table>
  <tr><td>User Id</td><td>
  <input required placeholder="User-id" id="name" name="name" value="{{.Name}}">
  </td></tr>
  <tr><td>Role</td><td>
    <select value="{{.Role}}" id="role" name="role" >
      {{if eq .Role "Staff"}}
      <option selected>Staff</option>
      {{else}}
      <option value="Staff">Staff</option>
      {{end}}
      {{if eq .Role "Desk"}}
      <option selected>Desk</option>
      {{else}}
      <option value="Desk">Desk</option>
      {{end}}
      {{if eq .Role "Manager"}}
      <option selected>Manager</option>
      {{else}}
      <option value="Manager">Manager</option>
      {{end}}
    </select>
  </td></tr>
  <tr><td>Last Name</td><td>
  <input required placeholder="Last Name" id="last" spellcheck="false" class="is-sensitive" value="{{.Last}}" name="last" />
  </td></tr>
  <tr><td>First Name</td><td>
  <input required placeholder="First Name" id="first" spellcheck="false" class="is-sensitive" value="{{.First}}" name="first" />
  </td></tr>
  <tr><td>Middle Name</td><td>
  <input required placeholder="Middle Name" id="middle" spellcheck="false" class="is-sensitive" value="{{.Middle}}" name="middle" />
  </td></tr>
  <tr><td>Salary</td><td>
  <input required placeholder="Salary" id="salary" spellcheck="false" value="{{.Salary}}" name="salary" />
  </td></tr>
  <tr><td>Password</td><td>
  <input required type="password" placeholder="password" id="pwd" spellcheck="false" class="is-sensitive" value="{{.Pwd}}" name="pwd" />
  </td></tr>

  </table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
