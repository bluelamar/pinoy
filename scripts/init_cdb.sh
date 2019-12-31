#!/bin/bash

# run this script once after first creation of the DB

DBUSERNAME=$1
DBPASSWD=$2

# get a session cookie
curl -v -c cdbcookies -H "Accept: application/json" -H "Content-Type: application/x-www-form-urlencoded"  http://localhost:5984/_session -X POST -d "name=$DBUSERNAME&password=$DBPASSWD"

# run in single node mode
# http://docs.couchdb.org/en/stable/setup/single-node.html
curl -v --cookie "cdbcookies" http://localhost:5984/_users -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/_replicator -X PUT
# IGNORE this one: curl -v --cookie "cdbcookies" http://localhost:5984/_global_changes -X PUT

exit 0

# create the db's
curl -v --cookie "cdbcookies" http://localhost:5984/rooms -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/room_status -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/room_rates -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/room_usage -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/room_usage_bkup_b -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/room_usage_bkup_c -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/food -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/food_rates -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/bellhops -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/staff -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/staff_hours -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/staff_hours_bkup_b -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/staff_hours_bkup_c -X PUT
curl -v --cookie "cdbcookies" http://localhost:5984/testxyz -X PUT
#curl -v --cookie "cdbcookies" http://localhost:5984/hop_shift -X PUT

INDEX='{"index":{"fields":[{"UserID":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/staff_hours/_index -X POST -d $INDEX
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/staff_hours_bkup_b/_index -X POST -d $INDEX
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/staff_hours_bkup_c/_index -X POST -d $INDEX

INDEX='{"index":{"fields":[{"UserID":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/bellhops/_index -X POST -d $INDEX
INDEX='{"index":{"fields":[{"TimeStamp":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/bellhops/_index -X POST -d $INDEX
INDEX='{"index":{"fields":[{"Room":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/bellhops/_index -X POST -d $INDEX

INDEX='{"index":{"fields":[{"Room":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/room_usage/_index -X POST -d $INDEX

# create indices
INDEX='{"index":{"fields":[{"Status":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/room_status/_index -X POST -d $INDEX

INDEX='{"index":{"fields":[{"Last":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/staff/_index -X POST -d $INDEX

INDEX='{"index":{"fields":[{"Role":"desc"}]}}'
curl -v -H "Content-Type: application/json" --cookie "cdbcookies" http://localhost:5984/staff/_index -X POST -d $INDEX

# find: res= map[docs:[map[Rate:Large Room RoomNum:16 Status:open _id:16 _rev:1-0450a844bcc9679a730910143e15946c CheckinTime: GuestInfo:] map[Status:open _id:17 _rev:1-a265ca63e36e11a4e5a7ab72304c6d49 CheckinTime: GuestInfo: Rate:Medium Room RoomNum:17] map[CheckinTime: GuestInfo: Rate:Large Room RoomNum:18 Status:open _id:18 _rev:1-0deb62dafd9602c4140e9405ac029443] map[Rate:Small Room RoomNum:19 Status:open _id:19 _rev:1-b250ac6f9757354fc03b397fad28ec41 CheckinTime: GuestInfo:]] bookmark:g1AAAAA0eJzLYWBgYMpgSmHgKy5JLCrJTq2MT8lPzkzJBYkbWoJkOGAyULEsAGSYDY0 warning:no matching index found, create an index to optimize query time]

