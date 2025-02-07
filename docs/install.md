# Manual Installation 

Please see [build.md](build.md).

# Installation on Ubuntu/Debian via apt package
Download the `apt` package in `.deb` format for your OS version from the release page: https://github.com/digitalocean/do-dcgm-exporter/releases.

Install the package from the local filesystem
- sets up `/etc/apt/sources.list.d/do-dcgm-exporter.list` for future package upgrades
- creates `/etc/systemd/system/do-dcgm-exporter.service`
```
$ sudo apt install do-dcgm-exporter_0.0.1_amd64-jammy.deb
```

Then start the `do-dcgm-exporter.service` service

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl start do-dcgm-exporter.service
```

Check if the service is running

```bash
$ sudo systemctl status do-dcgm-exporter
● do-dcgm-exporter.service - DigitalOcean DCGM Exporter
     Loaded: loaded (/etc/systemd/system/do-dcgm-exporter.service; disabled; vendor preset: enabled)
     Active: active (running) since Fri 2025-02-07 22:12:52 UTC; 47s ago
   Main PID: 596393 (do-dcgm-exporte)
      Tasks: 15 (limit: 289778)
     Memory: 32.8M
        CPU: 37ms
     CGroup: /system.slice/do-dcgm-exporter.service
             └─596393 /opt/digitalocean/bin/do-dcgm-exporter

Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Initializing system entities of type: NvSwitch"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Not collecting NvSwitch metrics: no switches to monitor"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Initializing system entities of type: NvLink"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Not collecting NvLink metrics: no switches to monitor"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Not collecting CPU metrics: no fields to watch for devi>
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Not collecting CPU Core metrics: no fields to watch for>
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Pipeline starting"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Starting webserver"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="Listening on" address="[::]:9401"
Feb 07 22:12:52 my-droplet do-dcgm-exporter[596393]: time="2025-02-07T22:12:52Z" level=info msg="TLS is disabled." address="[::]:9401" http2=false
```