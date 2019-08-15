# pinoy
Pinoy Lodge Application

## Configuration:

**pinoy** will load config from /etc/pinoy/config.json

### Logging

**LogOutput** takes 1 of these values: file, stderr, stdout

If **file** is specified, then set **LogFile** to the explicit file path.

Example:
```
  "LogOutput": "file",
  "LogFile": "/var/log/pinoy/pinoy.log",
```

Else set **LogFile** to the empty string **""**

Example: 
```
  "LogOutput": "stderr",
  "LogFile": "",
```

## Log Rotation:

Place logrotate config files in /etc/logrotate.d/

Example: pinoy
```
  /var/log/pinoy/*log {
    rotate 10
    daily
    postrotate
      /usr/bin/killall -HUP pinoy
    endscript
  }
```

Example: couchdb
```
  /var/opt/couchdb/logs/couch.log {
    daily
    rotate 7
    copytruncate
    delaycompress
    compress
    notifempty
    missingok
}
```

Ensure logrotate will run daily by adding script to /etc/cron.daily/

Example: logrotate
```
#!/bin/sh

# Clean non existent log file entries from status file
cd /var/lib/logrotate
test -e status || touch status
head -1 status > status.clean
sed 's/"//g' status | while read logfile date
do
    [ -e "$logfile" ] && echo "\"$logfile\" $date"
done >> status.clean
mv status.clean status

test -x /usr/sbin/logrotate || exit 0
/usr/sbin/logrotate /etc/logrotate.conf
```


