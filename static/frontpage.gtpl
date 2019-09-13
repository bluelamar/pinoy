{{ define "pagecontent" }}

<script>
function roomStateTimer() {
  var request = new XMLHttpRequest()
  // Open a new connection, using the GET request on the URL endpoint
  request.open('GET', '/desk/room_stati', true);
  request.onload = function () {
    var res = 'Rooms soon to Checkout: ';
    var data = JSON.parse(this.response)
    data.forEach(rs => {
      res = res.concat('[').concat(rs.RoomNum).concat(' at ').concat(rs.CheckoutTime).concat('] ');
      //console.log(res)
    })
    document.getElementById("roomstati").innerHTML = res;
  }
  request.send()
}
var myVar = setInterval(roomStateTimer, 1000 * 300);
roomStateTimer();
</script>

{{if and .Sess.Auth .Sess.User}}
<div id="roomstati" class="roomstati">
</div>
{{end}}

  <h1>Front Desk</h1>
{{if and .Sess.Auth .Sess.User}}
  <p><h2>Hello {{.Sess.User}}</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
  <nav>
    <ul>
      <li><a href="/desk/room_status?register=open" class="pinoylink">Register Guest</a></li>
      <li><a href="/desk/room_status?register=booked" class="pinoylink">Booked Room Status</a></li>
      <li><a href="/desk/food" class="pinoylink">Order Food and Drink</a></li>
      <li><a href="/desk/report_staff_hours" class="pinoylink">Clock-in/Clock-out Staff</a></li>
      {{if eq .Sess.Role "Manager"}}
      <p>Managerial Actions</p>
      <li><a href="/manager/report_room_usage" class="pinoylink">Room Usage Reports</a></li>
      <li><a href="/manager/staff" class="pinoylink">Add or Update Staff</a></li>
      <li><a href="/manager/room_rates" class="pinoylink">Add or Update Room Rates</a></li>
      <li><a href="/manager/rooms" class="pinoylink">Add or Update Rooms</a></li>
      <li><a href="/manager/upd_food" class="pinoylink">Add or Update Food Items</a></li>
      <li><a href="/manager/svc_stats" class="pinoylink">Server Statistics</a></li>
      {{end}}
    </ul>
  </nav>

{{end}}

{{ end }}
{{ end }}
