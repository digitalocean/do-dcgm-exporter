# Prerequisite

1. The DigitalOcean DCGM-Exporter exporter is assumed to run on DigitalOcean droplets.
2. Nvidia drivers must be installed. Verify that the binary `nvidia-smi` is available and can discover GPUs and NVSwitches.
   - When using the default OS base image for GPU droplet (currently named `AI/ML ready`), the NVIDIA drivers are already preinstalled.

```bash
# output for a droplet with a single H100 GPU
root@myGPUDroplet:~# nvidia-smi
Tue Feb 11 21:20:03 2025
+---------------------------------------------------------------------------------------+
| NVIDIA-SMI 535.216.01             Driver Version: 535.216.01   CUDA Version: 12.2     |
|-----------------------------------------+----------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id        Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |         Memory-Usage | GPU-Util  Compute M. |
|                                         |                      |               MIG M. |
|=========================================+======================+======================|
|   0  NVIDIA H100 80GB HBM3          On  | 00000000:00:09.0 Off |                    0 |
| N/A   29C    P0              73W / 700W |      0MiB / 81559MiB |      0%      Default |
|                                         |                      |             Disabled |
+-----------------------------------------+----------------------+----------------------+
```

3. NVIDIA Data Center GPU Manager [(DCGM)](https://developer.nvidia.com/dcgm) must be installed
- Installation of DCGM in a DigitalOcean Droplet usually simply involves the following commands
```bash
$ sudo apt install datacenter-gpu-manager
$ sudo systemctl --now enable nvidia-dcgm

# output for a droplet with a single H100 GPU
root@myGPUDroplet:~# dcgmi discovery -l
1 GPU found.
+--------+----------------------------------------------------------------------+
| GPU ID | Device Information                                                   |
+--------+----------------------------------------------------------------------+
| 0      | Name: NVIDIA H100 80GB HBM3                                          |
|        | PCI Bus ID: 00000000:00:09.0                                         |
|        | Device UUID: GPU-abc                |
+--------+----------------------------------------------------------------------+
0 NvSwitches found.
```

# Manual Installation of the DigitalOcean DCGM-Exporter

Please see [build.md](build.md).

# Installation of the DigitalOcean DCGM-Exporter on Ubuntu/Debian via apt package
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