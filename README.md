# mackerel-plugin-scheduler-latency-kvm

A [Mackerel](https://mackerel.io/) custom plugin that collects CPU scheduler latency (`run_delay`) metrics for KVM virtual machines.

## Overview

This plugin reads `/proc/<pid>/task/<tid>/schedstat` for each `qemu-system-x86_64` process running on the host. It measures the scheduling delay (run_delay) to detect CPU steal time at the hypervisor level.

## Requirements

- Linux with `CONFIG_SCHEDSTATS` enabled in the kernel
- KVM virtual machines running with `qemu-system-x86_64`

## Installation

Download the binary from [Releases](https://github.com/hiboma/mackerel-plugin-scheduler-latency-kvm/releases) (deb/rpm packages available), or build from source:

```bash
go build .
```

## Usage

Add the plugin to your mackerel-agent configuration:

```toml
[plugin.metrics.scheduler-latency-kvm]
command = "/usr/bin/mackerel-plugin-scheduler-latency-kvm"
```

## Metrics

The plugin outputs the following metrics:

| Metric | Description |
|---|---|
| `vm.<name>.run_delay` | Per-VM scheduling delay (percentage) |
| `vm.all.mean_run_delay` | Mean scheduling delay across all VMs |
| `vm.all.max_run_delay` | Max scheduling delay across all VMs |

VM names are extracted from the qemu `-name` cmdline argument. Dots in the name are replaced with hyphens (e.g., `www.example.com` becomes `www-example-com`). If the name cannot be determined, the process PID is used instead.

## How It Works

1. Find all `qemu-system-x86_64` processes via `/proc`
2. For each VM, sum `run_delay` from `schedstat` across all threads (TIDs)
3. Sleep 1 second, collect again, and compute the delta
4. Output per-second run_delay as a percentage

The `run_delay` field in `schedstat` represents the time a task spent waiting on a runqueue. A high value indicates CPU contention at the hypervisor level, which guests observe as steal time.

## Build

```bash
make build    # go build .
make test     # go test .
```

## License

MIT

