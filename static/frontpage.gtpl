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

<div id="roomstati" class="roomstati">
</div>

  <h1>Front Desk</h1>
{{if and .Sess.Auth .Sess.User}}
  <p><h2>Hello {{.Sess.User}}</h2></p>

{{if or (eq .Sess.Role "Manager") (eq .Sess.Role "Desk")}}
  <nav>
    <ul>
      <li><a href="/desk/room_status?register=open">Register Guest</a></li>
      <li><a href="/desk/room_status?register=booked">Booked Room Status</a></li>
      <li><a href="/desk/food">Order Food and Drink</a></li>
      <li><a href="/desk/report_staff_hours">Clock-in/Clock-out Staff</a></li>
      {{if eq .Sess.Role "Manager"}}
      <p>Managerial Actions</p>
      <li><a href="/manager/report_room_usage">Room Usage Reports</a></li>
      <li><a href="/manager/staff">Add or Update Staff</a></li>
      <li><a href="/manager/room_rates">Add or Update Room Rates</a></li>
      <li><a href="/manager/rooms">Add or Update Rooms</a></li>
      <li><a href="/manager/food">Add or Update Food Items</a></li>
      <li><a href="/manager/svc_stats">Server Statistics</a></li>
      {{end}}
    </ul>
  </nav>

{{end}}

{{ end }}
{{ end }}
