{{ define "body_prefix" }}

<p>
<a href="/frontpage">Front Desk</a>

{{if and .Sess .Sess.Auth}}

<div class="account-buttons">
  <div id="entrance">
    <a id="sign-out" class="button" href="/signout">Sign Out</a>
  </div>
</div>

{{else}}

<div class="account-buttons">
  <div id="entrance">
    <p><a id="log-in" class="button" href="/signin">Sign In</a></p>
  </div>
</div>

{{end}}
</p>
{{end}}
