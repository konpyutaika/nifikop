# How to contribute

NiFiKop is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document outlines some of the conventions on commit message formatting, contact points for developers, and other resources to help get contributions into operator-sdk.

# Email and Chat

- Email: [nifikop][nifikop]
- Slack : [https://slack.konpytika.io](https://join.slack.com/t/konpytika/shared_invite/zt-14md072lv-Jr8mqYoeUrqzfZF~YGUpXA)

## Getting started

- Fork the repository on GitHub
- See the [developer guide](https://konpyutaika.github.io/nifikop/docs/6_contributing/1_developer_guide) for build instructions

## Reporting bugs and creating issues

Reporting bugs is one of the best ways to contribute. However, a good bug report has some very specific qualities, so please read over our short document on [reporting bugs](./doc/dev/reporting_bugs.md) before submitting a bug report. 
This document might contain links to known issues, another good reason to take a look there before reporting a bug.

## Contribution flow

This is a rough outline of what a contributor's workflow looks like:

- Create a topic branch from where to base the contribution. This is usually master.
- Run `make fmt` to fix any linting issues in your changes.
- Make commits of logical units.
- Make sure commit messages are in the proper format (see below).
- Push changes in a topic branch to a personal fork of the repository.
- Submit a pull request to operator-framework/operator-sdk.
- The PR must receive a LGTM from two maintainers found in the MAINTAINERS file.

Thanks for contributing!

### Code style

The coding style suggested by the Golang community is used in NiFiKop. See the [style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

Please follow this style to make NifiKop easy to review, maintain and develop.

### Format of the commit message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
scripts: add the test-cluster command

this uses tmux to setup a test cluster that can easily be killed and started for debugging.

Fixes #38
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 70 characters, the second line is always blank, and other lines should be wrapped at 80 characters. This allows the message to be easier to read on GitHub as well as in various git tools.

### Developer Certificate of Origin

Every commit must be signed using the -s switch in git commit.

Non signed commit will not be accepted/merged.

By signing each commit the developper agrees to the following standard DCO:

```text
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Source: https://developercertificate.org/

## Documentation

If the contribution changes the existing APIs or user interface it must include sufficient documentation to explain the use of the new or updated feature. Likewise the [CHANGELOG][changelog] should be updated with a summary of the change and link to the pull request.

[nifikop]: 
[changelog]: https://github.com/konpyutaika/nifikop/blob/master/CHANGELOG.md