# agones-allocator-client

This is a simple cli client for testing Agones allocator endpoints.

## Required Input

The following required flags can be passed, or set as environment variables.

```
      --ca-cert string                   The path the CA cert file in PEM format [AGONES_CA_CERT]
      --cert string                      The path the client cert file in PEM format [AGONES_CLIENT_CERT]
      --key string                       The path to the client key file in PEM format [AGONES_CLIENT_KEY]
```

In addition to these values, you will need to specify **either** `--hosts` or `--hosts-ping`.

### hosts

Hosts can be passed a list (slice) of hosts, like so: `example.com,foo.example.com`. In this scenario, the first host will be used. In the event of retries, the additional hosts will be used.

### hosts-ping

This flag is passed as a map like `--hosts-ping example.com=pingServer.example.com`. The pingServer will be used to determine the preferred host by way of shortest ping time. In the event of retries, the other hosts in the list will be used. In the event that the ping check fails, the host will not be added to the list of possible hosts.

## load-test

This command can be used to run a bunch of simultaneous allocations and connections. See the help for configuration.

NOTE: This currently only supports the Agones simple-udp or simple-tcp server. It makes a connect, says hello, waits, and then says goodbye and EXIT.

## Attribution

Original inspiration for this comes from [the Agones gRPC client example](https://github.com/googleforgames/agones/blob/release-1.6.0/examples/allocator-client/main.go)


## Join the Fairwinds Open Source Community

The goal of the Fairwinds Community is to exchange ideas, influence the open source roadmap, and network with fellow Kubernetes users. [Chat with us on Slack](https:\/\/join.slack.com\/t\/fairwindscommunity\/shared_invite\/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g) or [join the user group](https:\/\/www.fairwinds.com\/open-source-software-user-group) to get involved!


## Other Projects from Fairwinds

Enjoying agones-allocator-client? Check out some of our other projects:
* [Polaris](https://github.com/FairwindsOps/Polaris) - Audit, enforce, and build policies for Kubernetes resources, including over 20 built-in checks for best practices
* [Goldilocks](https://github.com/FairwindsOps/Goldilocks) - Right-size your Kubernetes Deployments by compare your memory and CPU settings against actual usage
* [Pluto](https://github.com/FairwindsOps/Pluto) - Detect Kubernetes resources that have been deprecated or removed in future versions
* [Nova](https://github.com/FairwindsOps/Nova) - Check to see if any of your Helm charts have updates available
* [rbac-manager](https://github.com/FairwindsOps/rbac-manager) - Simplify the management of RBAC in your Kubernetes clusters

Or [check out the full list](https://www.fairwinds.com/open-source-software?utm_source=agones-allocator-client&utm_medium=agones-allocator-client&utm_campaign=agones-allocator-client)
