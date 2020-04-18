# pinoy
Pinoy Lodge Application

## Configuration:

**pinoy** will load config from /etc/pinoy/config.json

There is no implementation checked into the repo for the following API:
```
func (cfg *PinoyConfig) DecryptDbPwd() (string, error) 
func (cfg *PinoyConfig) EncryptDbPwd() (string, error) 
```

This is for security reasons so that implementors may use their own secret implementations.
A simple implementation could like this:
```
    package config

    const key = "my secret"

    func (cfg *PinoyConfig) DecryptDbPwd() (string, error) {

	// decrypt base64 crypto to original value
	pwd := Decrypt(key, cfg.DbPwd)
	return pwd, nil
    }

    func (cfg *PinoyConfig) EncryptDbPwd() (string, error) {
	// encrypt base64 crypto to original value
	pwd := Encrypt(key, cfg.DbPwd)
	cfg.DbPwd = pwd

	return cfg.DbPwd, nil
    }
```


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

## Development Environment:

Below is example for setting up CouchDB.

Build the database:
```
./scripts/bld_cdb.sh
```

Run the couchdb:
```
./scripts/run_cdb.sh
```


