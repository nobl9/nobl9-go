# Release process

The release process for nobl9-go cannot be discussed without the context of
our platform releases. In the following document, `n9` will refer to our
platform **production** version.

## The process

1. Create release candidate version. \
  Most likely this step is where you can stop.
  Creating an official version release is only necessary if you want our users
  to have immediate access to the released changes.
  In many scenarios it might not be the case.
  Furthermore, creating release candidate version instead of official release
  gives us more time to "sit" on the introduced changes, which in turn
  increases a chance for bug detection and gives room to introduce breaking
  changes which might be necessary to address the uncovered issue.
2. Create official release version (restricted access). \
  Reach out to Foundations domain If you want to go with the official release.

## Merging to main

The single, most important rule to keep in mind when merging to nobl9-go main
branch is that it **MUST** be release ready. You **CANNOT** assume that
a release is going to happen on X date. There might be bugs or features you're
not aware of which will require immediate release.

How should I then account for different scenarios in relation to n9 state?
Below, we're listing the most common scenarios.
If you find this list lacking in any manner, please reach out so that we can
discuss your case and eventually write it down in this list.

### The changes are not required to be in sync with n9

In this scenario, there's nothing blocking you from merging to main.

<!-- markdownlint-disable-next-line MD001 -->
##### Examples

- Bug fix which corrects SDK specific behavior.
- New endpoint which is already available through n9 API.
- Dependency update.
- Documentation enhancement.

### The changes are required to be in sync with n9 (backwards compatible)

This scenario can be divided into two cases, depending on the following
characteristics of the code you've written:

- Visibility, is it exported (available through the public contract of the SDK)
  and what it is its magnitude (amount of exported code)?
- How long is it estimated to stay out of sync with n9?
  When is the version of n9, which supports these changes estimated to land on
  production?

#### Release version of n9 is known

In this scenario, we're most probably talking about small initiatives, which
are often easier to estimate time wise. Specific version of n9 which will
support the feature is most likely determined. The code will sit in nobl9-go
for a short period of time before it is supported by our API.

##### What should I do?

Add the n9 version which will support the changes you've introduced to
`## Release Notes` header.
Example:

```markdown
## Motivation

Add new field to Direct.

## Release Notes

Added new field to Direct called `this`. It allows this and that.
It will be available once Nobl9 platform version 1.69.0 is released.
```

##### Examples

- New field for Prometheus Direct which is not yet handled by n9.
- New endpoint which is yet to be available through n9 API.
- New manifest object kind.

#### Release version of n9 is unknown

In this scenario, we're most probably talking about large initiatives, which
are often harder to estimate time wise. Specific version of n9 which will
support the feature is most likely not yet determined.
The code will sit in nobl9-go for a longer period of time before it is
supported by our API.

##### What should I do?

If the code you've introduced is public,
it will likely sit there for some time.
Annotate each public code element with a `// experimental:` comment.
Write down a short detail, that it's not yet supported, explain if using this
struct/function will be ignored or result in an error.

Example:

```go
// Composite is doing this and that.
// experimental: this feature is not yet supported, applying the object will
// result in an error.
type Composite struct {
    ...
}
```

**NOTE**: Once the n9 version supporting it is released,
remember about removing these annotations.

##### Examples

- Composite SLOs v2 initiative.
- New set of endpoints which is yet to be available through n9 API.
- New manifest object kind.

### Breaking changes

#### The changes have not yet been released

You **MUST NOT** communicate these breaking changes with `## Breaking changes`
header in this scenario. It only makes sense to communicate breaking changes
if they were introduced between official versions (see next section).

#### The changes have been released with a previous official release

Describe the breaking changes under `## Breaking changes` header.

Do these changes need to be in sync with n9 platform?
We shouldn't ever be in this position!
If we get here somehow, we'll most likely need to figure out a custom way of
communicating these changes, which might involve reaching out to the users
directly.

## Ideal vs. real world

Ideally, things which are dependent upon certain changes in n9 would only
be released once n9 has been released.
In practice, while doable, it proves to be a challenge not worth the trouble.
Main reason is that feature branches for both n9 and nobl9-go would have to
be maintained.
While it might seem to be an acceptable cost for nobl9-go, it would prove to
be a real challenge for n9 where a magnitude and frequency of potentially
conflicting changes far outranks nobl9-go environment.

## Release automation details

We're using [Release Drafter](https://github.com/release-drafter/release-drafter)
to automate release notes creation. Drafter also does its best to propose
the next release version based on commit messages from `main` branch.

Release Drafter is also responsible for auto-labeling pull requests.
It checks both title and body of the pull request and adds appropriate labels. \
**NOTE:** The auto-labeling mechanism will not remove labels once they're
created. For example, If you end up changing PR title from `sec:` to `fix:`
you'll have to manually remove `security` label.

On each commit to `main` branch, Release Drafter will update the next release
draft. Once you're ready to create new version, simply publish this draft.

In addition to Release Drafter, we're also running a script which extracts
explicitly listed release notes and breaking changes which are optionally
defined in `## Release Notes` and `## Breaking Changes` headers.
It also performs a cleanup of the PR draft mitigating Release Drafter
shortcomings.
