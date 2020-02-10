{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Add or Update Food Item</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}

<form action="/manager/upd_food" method="post">
  <legend><b>Add New Item</b></legend>
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

<p><b>Current Items for Sale</b></p>
<table>
<tr>
<th>Item</th><th>Size</th><th>Price</th>
<th>Update Item</th><th>Delete Item</th>
</tr>
{{range .Items}}
<tr>
<td>{{.Item}}</td><td>{{.Size}}</td><td>{{.Price}}</td>
<td><a id="upd_food" class="button" href="/manager/upd_food?item={{.ItemID}}&update=true">Update Item</a></td>
<td><a id="del_food" class="button" href="/manager/upd_food?item={{.ItemID}}&update=delete">Delete Item</a></td>
</tr>
{{else}}
No food items to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
