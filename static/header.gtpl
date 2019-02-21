{{ define "header" }}
<title>{{.PgCont.PageTitle}} | Pinoy Lodge</title>

<style>
body { font-family: "proxima-nova", "proxima nova", "helvetica neue", "helvetica", "arial", sans-serif;}
</style>

<link rel="stylesheet" type="text/css" href="/css/base.css">

<meta http-equiv="Content-Type" content="text/html; charset=utf-8">

<meta name="keywords" content="social">
<meta name="description" content="{{.PgCont.PageDescr}}">

<meta name="og:title" content="{{.PgCont.PageTitle}} | Pinoy" />
<meta name="og:description" content="{{.PgCont.PageDescr}}" />
<meta name="og:site_name" content="Pinoy">
<meta name="og:type" content="website">
{{if .Sess}}
<meta name="csrf-param" content="{{.Sess.CsrfParam}}" />
<meta name="csrf-token" content="{{.Sess.CsrfToken}}" />
<meta name="csrf-sessid" content="{{.Sess.SessID}}" />
{{end}}
{{ end }}
