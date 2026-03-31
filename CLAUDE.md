# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Mackerel custom plugin that collects CPU scheduler latency (run_delay) metrics for KVM virtual machines. It reads `/proc/<pid>/task/<tid>/schedstat` for each qemu-system-x86_64 process to measure scheduling delays, then outputs Mackerel-formatted metrics (per-VM run_delay, mean, and max across all VMs).

Requires `CONFIG_SCHEDSTATS` enabled in the Linux kernel. Runs only on Linux.

## Build and Test

```bash
make build    # go build .
make test     # go test .
```

## Release

Releases are managed by GoReleaser, triggered by pushing a git tag. Semantic versioning tags are created via `git-semv` (installed from Homebrew tap `linyows/git-semv`).

```bash
make release_patch   # bump patch version and push tag
make release_minor   # bump minor version and push tag
make release_major   # bump major version and push tag
```

The GitHub Actions workflow (`.github/workflows/release.yml`) runs GoReleaser on tag push to produce linux/amd64 binaries in deb/rpm formats.

## Architecture

Single-file Go program (`kvm_steal.go`) with no Mackerel plugin helper library — it directly prints tab-separated metrics to stdout in `key\tvalue\ttimestamp` format.

Key flow in `main()`:
1. Check `/proc/self/schedstat` exists (kernel config guard)
2. `collectVMProcesses()` — find all `qemu-system-x86_64` processes via procfs
3. `collectRundelay()` — for each VM, sum `RunDelay` from schedstat across all threads (TIDs)
4. Sleep 1 second, collect again, compute delta to get per-second run_delay
5. `guessVMName()` — extract VM name from qemu `-name` cmdline argument (dots replaced with hyphens)
6. Output per-VM and aggregate (mean/max) metrics

Dependencies:
- `github.com/prometheus/procfs` — process enumeration from /proc
- `github.com/montanaflynn/stats` — mean/max statistical functions
