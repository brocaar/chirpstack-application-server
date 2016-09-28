# Frequently asked questions

## transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused

When you see the error `dial tcp 127.0.0.1:8000: getsockopt: connection refused`
in your logs, it means that LoRa App Server can't connect to
[LoRa Server](https://docs.loraserver.io/loraserver/).

## transport: dial tcp 127.0.0.1:8080: getsockopt: connection refused

You should see this log message only a couple of times.
This is a glitch in the current version of the LoRa App Server which will be
resolved in a next version. 
