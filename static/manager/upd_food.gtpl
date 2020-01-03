{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Add or Update Food Item</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/upd_food" method="post">
  <legend>Food Item Details</legend>
  <table>
  <tr><td>Food or Drink Item</td><td>
  <input required placeholder="Food or Drink Item" id="item" name="item" value="{{.FoodData.Item}}">
  </td></tr>
  <tr><td>Price</td><td>
  <input required placeholder="Price" id="item_price" value="{{.FoodData.Price}}" name="item_price" />
  </td></tr>
  <tr><td>Item Size</td><td>
  <input required placeholder="Item Size" id="item_size" spellcheck="false" class="is-sensitive" value="{{.FoodData.Size}}" name="item_size" />
  </td></tr>
  </table>
  <input type="submit" name="commit" value="Submit" />
</form>

{{end}}
{{end}}
{{end}}
