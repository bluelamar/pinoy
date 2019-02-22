{{ define "pagecontent" }}

<h1>Purchase Item</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

  <form action="/desk/purchase" method="post">
    <legend>Purchase</legend>
    <input placeholder="Room" label="false" spellcheck="false" class="is-sensitive" value="{{.Room}}" name="room_num" id="room_num" />
    <input placeholder="Item" label="false" spellcheck="false" class="is-sensitive"  value="" name="item" id="item" />
    <input placeholder="Size" label="false" spellcheck="false" class="is-sensitive"  value="" name="size" id="size" />
    <input placeholder="Quantity" label="false" spellcheck="false" class="is-sensitive"  value="" name="quantity" id="quantity" />
    <input type="submit" name="commit" value="Purchase" />
  </form>

{{end}}
{{end}}
{{end}}
