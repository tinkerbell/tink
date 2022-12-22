# Hello Contributors!

Thanks for your interest!
We're so glad you're here.
Thanks for helping make Tinkerbell better 😍!

There are many areas we can use contributions - ranging from code, documentation, feature proposals, issue triage, samples, and content creation.

First, please read and understand the code of conduct found [here](https://github.com/tinkerbell/.github/blob/main/CODE_OF_CONDUCT.md).
By participating, you're expected to uphold this code.

## Table of Contents

-   [Choose something to work on](#choose-something-to-work-on)
    -   [Get help](#get-help)
-   [Contributing](#contributing)
    -   [File an issue](#file-an-issue)
    -   [Submit a change](#submit-a-change)
        -   [DCO Sign Off](#DCO-Sign-Off)
        -   [Environment Details](#Environment-Details)
    -   [Code style guide](#code-style-guide)
-   [Understanding code structure](#understanding-code-structure)
    -   [cmd](#cmd)
    -   [db](#db)
    -   [deploy](#deploy)
    -   [grpc-server](#grpc-server)
    -   [protos](#protos)

## Choose something to work on

There are [multiple repositories](https://github.com/tinkerbell) within the Tinkerbell organization.
Each repository has beginner-friendly issues that are a great place to get started on your contributor journey.
For example, a list of issues for Tink repository can be found [here](https://github.com/tinkerbell/tink/issues).
If there is something that you find interesting and would like to work on, go ahead.
You can filter issues with label "[good first issue](https://github.com/tinkerbell/tink/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)", which are relatively self sufficient issues and great for first time contributors.

-   If you are going to pick up an issue, it would be good to add a comment stating the intention.
-   If the contribution is a big change/new feature, please raise an issue and discuss the needs, design in the issue in detail.

### Get help

Do reach out on Slack or Twitter and we are happy to help.

-   Drop by the [Slack channel](https://eqix-metal-community.slack.com).
-   Say "Hi!" on [Twitter](https://twitter.com/tinkerbell_oss).

## Contributing

### File an issue

Not ready to contribute code, but see something that needs work?
While the community encourages everyone to contribute code, it is also appreciated when someone reports an issue.
Issues should be filed under the appropriate Tinkerbell subrepository.
For example, a documentation issue should be opened in [tinkerbell/tinkerbell-docs](https://github.com/tinkerbell/tinkerbell-docs/issues).
Make sure to adhere to the prompted submission guidelines while opening an issue.

### Submit a change

All submissions are more than welcome.
There are [multiple repositories](https://github.com/tinkerbell) within the Tinkerbell organization.
Before you submit a change, you must fork the repository and submit a pull request with the change(s).
Please ensure that you adhere to the prompted submission guidelines while raising a pull request.
We will try to review and provide feedback as soon as possible.

### DCO Sign Off

Please read and understand the DCO found [here](docs/DCO.md).

### Environment Details

Building is handled by `make`, please see the [Makefile](Makefile) for available targets.

#### Nix

This repo's build environment can be reproduced using `nix`.

##### Install Nix

Follow the [Nix installation](https://nixos.org/download.html) guide to setup Nix on your box.

##### Load Dependencies

Loading build dependencies is as simple as running `nix-shell` or using [lorri](https://github.com/nix-community/lorri).
If you have `direnv` installed the included `.envrc` will make that step automatic.

### Code Style Guide

#### Protobuf

Please ensure protobuf related files are generated along with _any_ change to a protobuf file.
In the future CI will enforce this, but for the time being does not.
Handling of protobuf deps and generating the go files are both handled by the [protoc.sh](./protos/protoc.sh) script.
Both `go`, and `protoc` are required by `protoc.sh`.

#### Unit Tests

One must support their proposed changes with unit tests.
As you submit a pull request(PR) the CI generates a code coverage report.
It will help you identify parts of the code that are not yet covered in unit tests.

#### Go

##### Import Groups

There should be two groups of import blocks, one for stdlib and the other for everything else.

## Understanding code structure

This is a nonexhaustive list important packages that happen to cover most of the code base.

### cmd

The `cmd` package is home for three core binaries for Tinkerbell:

-   `tink-server` - the tink API server
-   `tink-worker` - responsible for executing the workload

### deploy

The `deploy` directory contains all the essentials to setup Tinkerbell stack.
You can setup a local stack with `docker-compose` or Vagrant.

### grpc-server

The `grpc-server` exposes a gRPC API that connects everything together.
It has a base server that implements the API.

### protos

The `protos` package contains all the protobuf files used by the gRPC server.
Handling of protobuf deps and generating the go files are both handled [buf] via `make pbfiles`.
The protubuf/grpc files are not generated as part of the build pipeline to keep builds fast.
CI will ensure generated files are up to date.

[buf]: https://buf.build/

### environment variables

Tink Server, CLI, and Worker environment variables are documented [here](docs/ENVVARS.md).
