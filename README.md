# Gohc

Library for making different type of healthchecks with a common interface.

Different type are:
- Http(s): which allow HTTP1, HTTP2 and HTTP3 healthcheck and can send data, check for multiple statuses and check for received data.
- Tcp: which allow TCP healthcheck by trying to connect in tcp send data if set and check for received data if set.
- GRPC: Perform grpc healthcheck defined in https://github.com/grpc/grpc/blob/master/doc/health-checking.md
- Program: Execute a program by passing config (json format) in stdin and check for exit code.

**Note**: All types allow tls support. You can, for example, do tcp+tls test.

## Usage

Go to [examples](./examples) folder to see how to use it.
