{{ define "signin" }}

{{if not .Sess.Auth}}

<div>
<p>Please sign-in</p>
<div class="account-buttons">
  <div id="entrance">
    <p><a id="log-in" class="button" href="/login">Log In</a></p>
  </div>
</div>
</div>

{{end}}
{{end}}
