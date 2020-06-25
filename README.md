# agones-allocator-client

This is a simple cli client for testing Agones allocator endpoints.

## Required Input

The following required flags can be passed, or set as environment variables.

```
      --ca-cert string                   The path the CA cert file in PEM format [AGONES_CA_CERT]
      --cert string                      The path the client cert file in PEM format [AGONES_CLIENT_CERT]
      --key string                       The path to the client key file in PEM format [AGONES_CLIENT_KEY]
```

In addition to this, you will need to specify **either** `--hosts` or `--hosts-ping`.

### hosts

Hosts can be passed a list (slice) of hosts, like so: `example.com,foo.example.com`. In this scenario, the first host will be used. In the event of retries, the additional hosts will be used.

### hosts-ping

This flag is passed as a map like `--hosts-ping example.com=pingServer.example.com`. The pingServer will be used to determine the preferred host by way of shortest ping time. In the event of retries, the other hosts in the list will be used. In the event that the ping check fails, the host will not be added to the list of possible hosts.

## load-test

This command can be used to run a bunch of simultaneousallocations and connections. See the help for configuration.

NOTE: This currently only supports the Agones simple-udp server. It makes a connect, says hello, waits, and then says goodbye and EXIT.

## Attribution

Original inspiration for this comes from [the Agones gRPC client example](https://github.com/googleforgames/agones/blob/release-1.6.0/examples/allocator-client/main.go)
