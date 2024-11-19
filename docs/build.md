# Building the DigitalOcean dcgm-exporter

The DigitalOcean DCGM exporter is a thin wrapper around the [dcgm-exporter](https://github.com/NVIDIA/dcgm-exporter) and has the [same build requirements as the dcgm-exporter](https://github.com/NVIDIA/dcgm-exporter?tab=readme-ov-file#building-from-source).

**Requirements**
- `gcc` for the use of CGO
- DCGM that includes `libdcgm.so`

Use the Makefile to compile the binary `do-dcgm-exporter-linux-amd64` in the `/bin` directory.

```bash
$ make build
```

Test the compiled binary using the `version` command:

```bash
$ bin/do-dcgm-exporter-linux-amd64 version
DigitalOcean GPU Metrics Agent:
		version                     : v0.0.1
		dcgm-exporter version       : 3.3.8-3.6.0
		build date                  : 2024-10-23
		go version                  : go1.23.2
		go compiler                 : gc
		platform                    : linux/amd64
```
