{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Food and Drink</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>

{{if eq .Sess.Role "Manager"}}
<p><a id="update_food" class="button" href="/manager/upd_food">Add Items</a></p>
{{end}}

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

<p>Items for Sale</p>
<table>
<tr>
<th>Item</th><th>Size</th><th>Price</th><th>Purchase</th>
{{if eq .Sess.Role "Manager"}}
<th>Update Item</th><th>Delete Item</th>
{{end}}
</tr>
{{$role:=.Sess.Role}}
{{range .Items}}
<tr>
<td>{{.Item}}</td><td>{{.Size}}</td><td>{{.Price}}</td><td><a id="purchase" class="button" href="/desk/purchase?item={{.ItemID}}">Purchase</a></td>
{{if eq $role "Manager"}}
<td><a id="upd_food" class="button" href="/manager/upd_food?item={{.ItemID}}&update=true">Update Item</a></td>
<td><a id="del_food" class="button" href="/manager/upd_food?item={{.ItemID}}&update=delete">Delete Item</a></td>
{{end}}
</tr>
{{else}}
No food items to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
