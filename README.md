# doktor
**doktor** is a (not yet functional in any way) kubectl plugin for tracing and debugging containers using [eBPF](https://ebpf.io/) capabilities of the Linux Kernel.

**doktor** is completely open source and is being built using [ksniff](https://github.com/eldadru/ksniff) as a sort of template since I am not smart enough to workout how to layout such a project myself (thanks @eldadru!).

## Status

**doktor** is at the earliest possible stage of development and doesn't work for anything just yet. Currently I am just planning out how to bring the benefits of **BPF** based tracing to an easy to use kubectl plugin.

I imagine that users can run commands like the ones below which will select the relevant process from the provided pod and run the bpftrace command against it.

```
$ kubectl doktor some-pod --filter 'tracepoint:raw_syscalls:sys_enter { @[comm] = count(); }'

```


## See also

* https://github.com/cloudflare/ebpf_exporter
* https://github.com/iovisor/bpftrace