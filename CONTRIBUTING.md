# Contributing

## Guidelines

Guidelines for contributing.

### How can I get involved?

We have a number of areas where we can accept contributions:

* Work on [missing features and bugs](https://github.com/ewilde/faas-fargate/issues)


### I've found a typo

* A Pull Request is not necessary. Raise an [Issue](https://github.com/ewilde/faas-fargate/issues) and we'll fix it as soon as we can. 

### I have a (great) idea

The maintainers would like to make faas-fargate the best it can be and welcome new contributions that align with the project's goals.
Our time is limited so we'd like to make sure we agree on the proposed work before you spend time doing it.
Saying "no" is hard which is why we'd rather say "yes" ahead of time. You need to raise a proposal.

**Please do not raise a proposal after doing the work - this is counter to the spirit of the project. It is hard to be objective about something which has already been done**

What makes a good proposal?

* Brief summary including motivation/context
* Any design changes
* Pros + Cons
* Effort required up front
* Effort required for CI/CD, release, ongoing maintenance
* Migration strategy / backwards-compatibility
* Mock-up screenshots or examples of how the CLI would work

If you are proposing a new tool or service please do due diligence.
Does this tool already exist? Can we reuse it? For example: a timer / CRON-type scheduler for invoking functions. 

### Paperwork for Pull Requests

Please read this whole guide and make sure you agree to our DCO agreement (included below):

* See guidelines on commit messages (below)
* Sign-off your commits
* Complete the whole template for issues and pull requests
* [Reference addressed issues](https://help.github.com/articles/closing-issues-using-keywords/) in the PR description & commit messages - use 'Fixes #IssueNo' 
* Always give instructions for testing
* Provide us CLI commands and output or screenshots where you can

### Commit messages

The first line of the commit message is the *subject*, this should be followed by a blank line and then a message describing the intent and purpose of the commit. These guidelines are based upon a [post by Chris Beams](https://chris.beams.io/posts/git-commit/).

* When you run `git commit` make sure you sign-off the commit by typing `git commit -s`.
* The commit subject-line should start with an uppercase letter
* The commit subject-line should not exceed 72 characters in length
* The commit subject-line should not end with punctuation (., etc)

When giving a commit body:
* Leave a blank line after the subject-line
* Make sure all lines are wrapped to 72 characters

Here's an example:

```
Add secrets to provider

We need to have the ability to pass secrets to faas-fargate securely.
This commits adds secrets support using  AWS Secrets Manager API.

Resolves #1

Signed-off-by: Edward Wilde <ewilde@gmail.com>
```

If you would like to ammend your commit follow this guide: [Git: Rewriting History](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History)

**Unit testing with Golang**

Please follow style guide on [this blog post](https://blog.alexellis.io/golang-writing-unit-tests/) from [The Go Programming Language](https://www.amazon.co.uk/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)

**I have a question, a suggestion or need help**

Please raise an Issue.

**I need to add a dependency**

We use vendoring for projects written in Go.
This means that we will maintain a copy of the source-code of dependencies within Git.
It allows a repeatable build and isolates change. 

We use Golang's `dep` tool to manage dependencies for Golang projects - https://github.com/golang/dep

## License

This project is licensed under the MIT License.

### Copyright notice

Please add a Copyright notice to new files you add where this is not already present:

```
// Copyright (c) fass-ecs Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
```

### Sign your work

> Note: all of the commits in your PR/Patch must be signed-off.

The sign-off is a simple line at the end of the explanation for a patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](http://developercertificate.org/)):

```
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

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.

* Please sign your commits with `git commit -s` so that commits are traceable.

If you forgot to sign your work and want to fix that, see the following guide: 
[Git: Rewriting History](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History)
