# How to Contribute to Kelp

Please read the [Contribution Guide below](#contribution-guide)

Then please [sign the Contributor License Agreement](https://docs.google.com/forms/d/e/1FAIpQLSc5eppq6GOu9-TnFuRMBh4qIP1ChmZx9lrA6zOTyYiowKiwpA/viewform?usp=sf_link).

## Contribution Guide

Your contributions to Kelp will help improve the [Stellar network](https://www.stellar.org/) and the world’s financial infrastructure, faster.

We want to make it as easy as possible to contribute changes that help the project grow and thrive. There are a few guidelines that we ask contributors to follow so that we can merge your changes quickly.

### Getting Started

* Make sure you have a [GitHub account](https://github.com/signup/free).
* [Create a GitHub issue](https://github.com/stellar/kelp/issues) for your contribution, assuming one does not already exist.
  * Clearly describe the issue including steps to reproduce it (if it is a bug).
* Fork the repository on GitHub and start working on your change.
* When your change is ready then [submit a Pull Request].
* Please [sign the Contributor License Agreement](https://docs.google.com/forms/d/e/1FAIpQLSc5eppq6GOu9-TnFuRMBh4qIP1ChmZx9lrA6zOTyYiowKiwpA/viewform?usp=sf_link) so we can accept your contributions.

#### Minor Changes

For low-impact changes (ex: comments, documentation), it is not always necessary to create a new GitHub issue. In this case, it is appropriate to start the first line of a commit with 'doc' instead of an issue number.

### Finding things to work on

The first place to start is always looking over the [open GitHub issues](https://github.com/search?l=&q=is%3Aopen+is%3Aissue+repo%3Astellar%2Fkelp&type=Issues) for the project you are interested in contributing to. Issues marked with [help wanted](https://github.com/search?l=&q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22+repo%3Astellar%2Fkelp&type=Issues) are usually pretty self-contained and a good place to get started.

Stellar Development Foundation (SDF) also uses these same GitHub issues to keep track of what we are working on. If you see any issues that are assigned to a particular person or have the `in progress` label, that means someone is currently working on that issue.

Of course, feel free to create a new issue if you think something needs to be added or fixed.

### Making Changes

* Create a feature branch _in your fork_ from where you want to base your work.
  * It is most common to base branch on the `master` branch.
  * Please avoid working directly on the `master` branch.
* Follow the code conventions of the existing repo. If you are adding a new file, please follow the Directory Structure in the [README](README.md#directory-structure).
* Make sure you have added the necessary tests for your changes and make sure all tests pass.
* Run your code against the test network to ensure that everything works.
* Update the README and walkthroughs if needed.

### Submitting Changes

* [Sign the Contributor License Agreement](https://docs.google.com/forms/d/e/1FAIpQLSc5eppq6GOu9-TnFuRMBh4qIP1ChmZx9lrA6zOTyYiowKiwpA/viewform?usp=sf_link) so we can accept your contributions.
* All content, comments, and pull requests must follow the [Stellar Community Guidelines](https://www.stellar.org/community-guidelines/).
* See [detailed guide on submitting code-reviewer-friendly PRs](https://mtlynch.io/code-review-love/).
* Follow the [Pull Request Guide for Kelp](.github/pull_request_template.md).
* Push your changes to a feature branch in your fork of the repository.
* Submit a Pull Request.
  * Include a descriptive [commit message](https://github.com/erlang/otp/wiki/Writing-good-commit-messages).
  * Include a link to your Github issue. Changes contributed via a pull request should focus on a single issue at a time.
  * Rebase your local changes against the `master` branch. Resolve any conflicts that arise.

At this point you're waiting on us. We like to at least comment on pull requests within three business days (typically, one business day). We may suggest some changes, improvements or alternatives.

## Additional Resources

* [Contributor License Agreement](https://docs.google.com/forms/d/e/1FAIpQLSc5eppq6GOu9-TnFuRMBh4qIP1ChmZx9lrA6zOTyYiowKiwpA/viewform?usp=sf_link)
* [Explore the Stellar API](https://www.stellar.org/developers/reference/)
* Ask questions on the [Stellar StackExchange](https://stellar.stackexchange.com/); use the `kelp` tag 

This document is inspired by:

* https://github.com/stellar/docs/blob/master/CONTRIBUTING.md
