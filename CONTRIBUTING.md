# Contributing

Issues, whether bugs, tasks, or feature requests are essential for keeping agones-allocator-client great. We believe it should be as easy as possible to contribute changes that get things working in your environment. There are a few guidelines that we need contributors to follow so that we can keep on top of things.

## Code of Conduct

This project adheres to a [code of conduct](CODE_OF_CONDUCT.md). Please review this document before contributing to this project.

## Sign the CLA
Before you can contribute, you will need to sign the [Contributor License Agreement](https://cla-assistant.io/fairwindsops/agones-allocator-client).

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/FairwindsOps/agones-allocator-client/labels/good%20first%20issue) if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue, please start a conversation on that issue, and we can help answer any questions as they come up.

## Setting Up Your Development Environment
### Prerequisites
* A properly configured Golang environment with Go 1.13 or higher

### Installation
* Clone the project with `go get github.com/fairwindsops/agones-allocator-client`
* Change into the agones-allocator-client directory which is installed at `$GOPATH/src/github.com/fairwindsops/agones-allocator-client`
* Use `make build` to build the binary locally.
* Use `make test` to run the tests and generate a coverage report.

## Creating a New Issue

If you've encountered an issue that is not already reported, please create an issue that contains the following:

- Clear description of the issue
- Steps to reproduce it
- Appropriate labels

## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Add tests that show the issues have been solved
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review or a work in progress
- Be up to date and/or rebased on the master branch

## Creating a new release

Push a new semver tag. Goreleaser will take care of the rest.

## Pre-commit

This repo contains a pre-commit file for use with [pre-commit](https://pre-commit.com/). Just run `pre-commit install` and you will have the hooks.
