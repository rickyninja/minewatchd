# minewatchd

Minewatchd tails a minecraft server's latest.log, and sends login & logout notices to a [notification service](https://github.com/rickyninja/httpemaild).

# configuration file

## emails

This is a list of email addresses to notify on matching log events.  Minewatchd makes use of [email to SMS gateway](https://en.wikipedia.org/wiki/SMS_API) for notification.

## `notify_url`

This is the url of an http notification service that will relay the notice.  I don't want to configure a mail server on the minecraft server, so reusing a host that is already setup for mail.

## `log_file`

This is the log file minewatchd will be tailing to match events.

## `time_zone`

This is the time zone the minecraft server's logs are configured to use.

## `muted_users`

This is so you won't get notices when you login to your own server.

## example .minewatchd.yaml config

```
---
emails:
  - foo@example.com
  - bar@example.com
notify_url: http://127.0.0.1:8080/minecraft/notify
log_file: latest.log
time_zone: America/Phoenix
muted_users:
  - ricky_ninja
```
