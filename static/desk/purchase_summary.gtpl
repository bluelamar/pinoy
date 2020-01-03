{{ define "pagecontent" }}

{{if and .Sess .Sess.User}}

<h1>Purchase Summary</h1>

<p><h2>Hello {{.Sess.User}}</h2></p>
<p><h3>{{.Sess.Role}} Page</h3></p>


{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}

<p>Items Bought</p>
<table>
<tr>
<th>Item</th><th>Size</th><th>Price</th><th>Quantity</th>
</tr>
{{range .Items}}
<tr>
<td>{{.Item}}</td><td>{{.Size}}</td><td>{{.Price}}</td><td>{{.Quantity}}</td>
</tr>
{{end}}

<p><b>Total Cost: {{.TotalCost}}</b></p>

<p><a id="purchase" class="button" href="/desk/food">Purchase More Items</a></p>
<p><a id="purchase_done" class="button" href="/desk/purchase_summary?done=true">Completed Purchases</a></p>

{{else}}
No items bought
{{end}}
</table>

{{end}}
{{end}}
