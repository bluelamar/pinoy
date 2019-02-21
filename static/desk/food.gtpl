{{ define "pagecontent" }}

<h1>Food and Drink</h1>

{{if and .Sess .Sess.User}}
<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h2>{{.Sess.Role}} Page</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

<p>Items</p>
<table>
<tr>
<th>Item</th><th>Size</th><th>Price</th>
</tr>
{{range .FoodTable}}
<tr>
<td>{{.Item}}</td><td>{{.Size}}</td><td>{{Price}}</td><td><a id="purchase" class="button" href="/desk/purchase?item={{.Item}}&size={{.Size}}">Purchase</a></td>
</tr>
{{else}}
No food items to report
{{end}}
</table>

{{end}}
{{end}}
{{end}}
