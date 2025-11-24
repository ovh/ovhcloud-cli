Thank you for wanting to contribute to this project ❤️

This project accepts contributions. In order to contribute, you should pay attention to a few things:

1. Your code must follow the coding style rules
2. Your code must be fully documented
3. Your code must be tested
4. GitHub Pull Requests

The following sections explain in details the contribution guidelines.

# Table of content

- [Submitting Modifications](#submitting-modifications)
- [Submitting an Issue](#submitting-an-issue)
- [Coding and Documentation Style](#coding-and-documentation-style)
- [Adding a New CLI Feature](#adding-a-new-cli-feature)
    - [1. Service (Business Logic)](#1-service-business-logic)
    - [2. CLI Command](#2-cli-command)
    - [3. Create / Edit Command UX Flags](#3-create--edit-command-ux-flags)
    - [4. Documentation](#4-documentation)
    - [5. Tests](#5-tests)
    - [6. Typical Structure Example](#6-typical-structure-example)
    - [7. Checklist Before Opening a PR](#7-checklist-before-opening-a-pr)
- [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco)

# Submitting Modifications:

The contributions should be submitted through new GitHub Pull Requests.

# Submiting an Issue:

In addition to contributions, we welcome [bug reports, feature requests and documentation error reports](https://github.com/ovh/ovhcloud-cli/issues/new).

# Coding and documentation Style:

- Code must be formatted with `make fmt` command
- Name your commands according to the API endpoint
- If the input body of an API call has more than five parameters or has more than one level of nesting, the corresponding CLI command must have flags '--editor' and '--from-file' to define its parameters.

# Adding a new CLI Feature

This document explains the minimal workflow to add a new feature to the CLI.

## 1. Service (Business Logic)

Place all code that performs HTTP calls and processes responses in a dedicated sub-package under:
`internal/services/<yourservice>`.

Any logic related to printing or formatting output should be handled by the separate `internal/display` package, which must be used for presenting results or errors to the user. This package handles the formatting depending on the output asked by the user (JSON, YAML, …).

## 2. CLI Command

Add the corresponding command in: `internal/cmd/<yourservice>.go`.
Follow existing patterns for command registration and use cohesive, single-purpose files.

## 3. Create / Edit Command UX Flags

For resource creation or edition commands, you MUST support:

- `--editor`       open a temporary spec in $EDITOR
- `--init-file`    create a skeleton file
- `--from-file`    load full request body from a file

Helper functions already exist to add these flags. See the IAM user creation implementation for reference.

## 4. Documentation

After adding or modifying commands, regenerate docs from the repository root:
```bash
make doc
```

Don't commit changed made to `doc/ovhcloud.md` except if they are manual changes.

## 5. Tests

Add tests for the new command(s) in: `internal/cmd/<yourservice>_test.go` Focus on:

- Flag parsing
- Error paths
- Testing the successful execution path by invoking the service layer using mocks

## 6. Typical Structure Example

```
internal/services/policy/        # HTTP + response handling
internal/cmd/policy/             # cobra command definitions
doc/ovhcloud_<generated>.md      # regenerated via make doc
```

## 7. Checklist Before Opening a PR

- Service code isolated in internal/services
- Command code added in internal/cmd
- Supports required create/edit flags (if applicable)
- Docs regenerated (make doc)
- Tests added and passing
- Lint/build succeeds (make build)
- Keep changes small and consistent with existing patterns
- Prefer commands that are limited to a single use-case and do not require excessive HTTP calls

# Developer Certificate of Origin (DCO)
 
To improve tracking of contributions to this project we will use a
process modeled on the modified DCO 1.1 and use a "sign-off" procedure
on patches that are being emailed around or contributed in any other
way.
 
The sign-off is a simple line at the end of the explanation for the
patch, which certifies that you wrote it or otherwise have the right
to pass it on as an open-source patch.  The rules are pretty simple:
if you can certify the below:
 
By making a contribution to this project, I certify that:
 
(a) The contribution was created in whole or in part by me and I have
    the right to submit it under the open source license indicated in
    the file; or
 
(b) The contribution is based upon previous work that, to the best of
    my knowledge, is covered under an appropriate open source License
    and I have the right under that license to submit that work with
    modifications, whether created in whole or in part by me, under
    the same open source license (unless I am permitted to submit
    under a different license), as indicated in the file; or
 
(c) The contribution was provided directly to me by some other person
    who certified (a), (b) or (c) and I have not modified it.
 
(d) The contribution is made free of any other party's intellectual
    property claims or rights.
 
(e) I understand and agree that this project and the contribution are
    public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
 
 
then you just add a line saying
 
    Signed-off-by: Random J Developer <random@example.org>
 
using your real name (sorry, no pseudonyms or anonymous contributions.)