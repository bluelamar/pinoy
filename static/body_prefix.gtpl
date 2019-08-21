{{ define "body_prefix" }}
<div class="header">
Pinoy Lodge
</div>
<p>

{{if and .Sess .Sess.Auth}}

<div class="account-buttons">
  <div id="entrance">
    <a id="sign-out" class="button litegreen" href="/signout">Sign Out</a>
  </div>
</div>

{{else}}

<div class="account-buttons">
  <div id="entrance">
    <p><a id="log-in" class="button" href="/signin">Sign In</a></p>
  </div>
</div>

{{end}}

<p>
<a class="button" href="/frontpage">Front Desk</a>
</p>

{{$msgLen:=.Sess.Role}}
{{if .Sess}}
{{$msgLen:=len .Sess.Message}}
{{if gt $msgLen 0}}
<p>
<div class="specialmsg">
{{.Sess.Message}}
</div>
</p>
{{end}}
{{end}}
</p>
{{end}}
