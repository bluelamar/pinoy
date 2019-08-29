<html>
<head>
  {{ template "header" . }}
</head>
<body>
  <h1>Pinoy Lodge Sign in</h1>
  <form action="/signin" method="post">
    <legend>Log In</legend>
    <input placeholder="User ID" label="false" spellcheck="false" class="is-sensitive" value="" name="user_id" id="user_id" />
    <input placeholder="Password" label="false" autocomplete="off" class="is-sensitive" type="password" name="user_password" id="user_password" />
    <div class="grbutton">
    <input type="submit" name="commit" value="Login" />
    </div>
  </form>
  <p><b>Forgot password? Contact your manager to have it reset please.</b></p>
</body>
</html>
