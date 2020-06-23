# agones-allocator-client

This is a simple cli client for testing Agones allocator endpoints.

## Required Input

The following required flags can be passed, or set as environment variables.

```
      --ca-cert string                   The path the CA cert file in PEM format [AGONES_CA_CERT]
      --cert string                      The path the client cert file in PEM format [AGONES_CLIENT_CERT]
      --host string                      The hostname or IP address of the allocator server [AGONES_HOST]
      --key string                       The path to the client key file in PEM format [AGONES_CLIENT_KEY]
```

## load-test

This command can be used to run a bunch of simultaneousallocations and connections. See the help for configuration.

NOTE: This currently only supports the Agones simple-udp server. It makes a connect, says hello, waits, and then says goodbye and EXIT.
