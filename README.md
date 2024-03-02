# rant-go

[Rant](https://rant.anton2920.ru) is the simple [X](https://x.com) clone built using an experimental high-performance HTTP server. It allows you to watch my short messages, search for them on client or on server and subscribe to updates using [RSS](https://rant.anton2920.ru/rss).

Server is built using Go and assembly without any packages, except for package `runtime` and its dependencies from Go's standard library. Currently only `freebsd/amd64` is supported.

HTTP server supports only `GET` requests, query parameters can be included but must be parsed by hand. It also supports pipelining, infinite keep-alives and a lot of concurrent connections.

For C version of this server, see https://github.com/anton2920/rant-c.

# Performance

Using [wrk](https://github.com/wg/wrk) I've measured performance of `net/http` and my server using rules of [Plaintext](https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#plaintext) benchmark rules. Results on my `i7 6700K` with 32 GiB RAM, sending 16 pipelined requests with `text/plain` `Hello, world\n` response are following:

```
net/http:

$ ./wrk -t 4 -c 256 -d 15 --script plaintext.lua http://localhost:7070/plaintext -- 16
Running 15s test @ http://localhost:7070/plaintext
  4 threads and 256 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    11.67ms   21.01ms 358.21ms   95.32%
    Req/Sec   151.80k    13.82k  182.68k    73.17%
  9090809 requests in 15.05s, 1.11GB read
Requests/sec: 603914.99
Transfer/sec:     75.45MB

rant:

$ ./wrk -t 5 -c 256 -d 15 --script plaintext.lua http://localhost:7070/plaintext -- 16
Running 15s test @ http://localhost:7070/plaintext
  5 threads and 256 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.44ms   42.67ms 999.54ms   98.25%
    Req/Sec     1.18M    51.48k    1.42M    94.96%
  88366480 requests in 15.10s, 10.70GB read
Requests/sec: 5852009.97
Transfer/sec:    725.52MB
```

Both server and `wrk` were running on one computer. For each server  `wrk` parameters was selected to produce the highest results.

# Copyright

Pavlovskii Anton, 2023-2024 (MIT).
