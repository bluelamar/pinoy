{{ define "pagecontent" }}

<h1>Purchase Item</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

<p>Item Details</p>
  <form action="/desk/purchase" method="post">
    <table>
    <tr>
    <td>Room</td><td>
    <input required placeholder="Room" label="false" spellcheck="false" class="is-sensitive" value="{{.Room}}" name="room_num" id="room_num" />
    </td>
    </tr>
    <tr>
    <td>Item</td><td>
    <input required placeholder="Item" label="false" spellcheck="false" class="is-sensitive"  value="{{.FoodData.Item}}" name="item" id="item" />
    </td>
    </tr>
    <tr>
    <td>Size</td><td>
    <input required placeholder="Size" label="false" spellcheck="false" class="is-sensitive"  value="{{.FoodData.Size}}" name="size" id="size" />
    </td>
    </tr>
    <tr><td>Quantity</td><td>
    <input required placeholder="Quantity" label="false" spellcheck="false" class="is-sensitive"  value="" name="quantity" id="quantity" />
    </td></tr></table>
    <input type="submit" name="commit" value="Purchase" />
  </form>

{{end}}
{{end}}
{{end}}
