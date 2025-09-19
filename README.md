Fritzcable exporter
===================

# What
[Prometheus](https://github.com/prometheus/prometheus) exporter for Fritzbox Cable downstream channel error counters

# Why
Mine ends up in a 100% package loss state sometimes, but without detecting a problem by itself, or reconnecting on it's
own. Error counters seem high when I take a look in the UI, but I'd like a history to see if it's really related, or 
just coincidence, because I don't look at the counters in the UI if everything is fine.

There are loads more FritzBox Exporters on GitHub.

I needed one that exports Docsis channel statistics, which none of the TR-064 exporters can expose, because AVM
didn't put them in TR-064 in the first place for whatever reason.

And then I didn't look too hard for an exporter using the UI endpoints because I wanted to play around with Go again a 
bit.

# How
    ❯ go run ./cmd/main.go --help
    Usage of /var/folders/6x/sbngy4696fdbxpg2v3q78jk00000gp/T/go-build1776044078/b001/exe/main:
      -host string
            hostname of IP address of target device (default "fritz.box")
      -listen string
            net/http listen string (default ":8080")
      -password string
            password for device login
      -passwordFile string
            path to file containing the password
      -username string
            username for login (if any)

Wire up to your existing Prometheus installation you no doubt have

# Trigger resync
`/webhook` endpoint accepts Alertmanger formatted messages and trigger a resync of the cable
modem if any of the alerts in the list is in state `firing`.

## Alert example

This makes use of a blackbox_exporter installation to determine if Internet connection is working

Example rule:
```yaml
- name: fritzbox
  rules:
    - alert: FritzBorked
      expr: sum(probe_success{job="probe-icmp",instance!="192.168.0.1"}) == 0 AND (sum(docsis_ready_state{state="ready"}) == 1)
      for: 30s
      labels:
        severity: critical
      annotations:
        summary: FritzBox thinks it's connected, but it isn't
```

Matching Alertmanager config snippets:
```yaml
route:
  ...
  routes:
    - matchers:
        - alertname="FritzBorked"
      receiver: 'fritzcable'
      group_wait: 0s
      group_interval: 0s
      repeat_interval: 180s
...
receivers:
  ...
  - name: 'fritzcable'
    webhook_configs:
      - url: http://localhost:9004/webhook
```