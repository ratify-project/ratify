# Reviewing Guide
This document covers who may review pull requests for this project, and provides guidance on how to perform code reviews that meet our community standards and code of conduct. All reviewers must read this document and agree to follow the project review guidelines. Reviewers who do not follow these guidelines may have their privileges revoked.

## Attribution
The reviewing guide was created using content and verbiage from [CNCF reviewing guide]( https://contribute.cncf.io/maintainers/github/templates/recommended/reviewing/)

## Values

All reviewers must abide by the [Code of Conduct](CODE_OF_CONDUCT.md) and are also protected by it. A reviewer should not tolerate poor behavior and is encouraged to report any behavior that violates the Code of Conduct. All of our values listed above are distilled from our Code of Conduct.

Below are concrete examples of how it applies to code review specifically:

### Inclusion

Be welcoming and inclusive. You should proactively ensure that the author is successful. While any particular pull request may not ultimately be merged, overall we want people to have a great experience and be willing to contribute again. Answer the questions they didn't know to ask or offer concrete help when they appear stuck.

### Sustainability

Avoid burnout by enforcing healthy boundaries. Here are some examples of how a reviewer is encouraged to act to take care of themselves:

* Authors should meet baseline expectations when submitting a pull request, such as writing tests.
* If your availability changes, you can step down from a pull request and have someone else assigned.
* If interactions with an author are not following code of conduct, close the PR and raise it up with your Code of Conduct committee or point of contact. It's not your job to coax people into behaving.

### Trust

Be trustworthy. During a review, your actions both build and help maintain the trust that the community has placed in this project. Below are examples of ways that we build trust:

* **Transparency** - If a pull request won't be merged, clearly say why and close it. If a pull request won't be reviewed for a while, let the author know so they can set expectations and understand why it's blocked.
* **Integrity** - Put the project's best interests ahead of personal relationships or company affiliations when deciding if a change should be merged.
* **Stability** - Only merge when then change won't negatively impact project stability. It can be tempting to merge a pull request that doesn't meet our quality standards, for example when the review has been delayed, or because we are trying to deliver new features quickly, but regressions can significantly hurt trust in our project.

## Process

### PR Lifecycle

#### 1. PR creation
* Ensure that an issue has been created to track the feature enhancement or bug that is being fixed.
* In the PR description, make sure you've included "Fixes #{issue_number}" e.g. "Fixes #242" so that GitHub knows to link it to an issue.
* To avoid multiple contributors working on the same issue, please add a comment to the issue to let us know you plan to work on it.
* If a significant amount of design is required, please include a proposal in the issue and wait for approval before working on code. If there's anything you're not sure about, please feel free to discuss this in the issue. We'd much rather all be on the same page at the start, so that there's less chance that drastic changes will be needed when your pull request is reviewed.

#### 2. Reviewing/Discussion

* Everyone is encouraged to participate in code reviews.
* It is best to get approval from contributors with knowledge about the code paths being modified. You can identify an appropriate reviewer by looking through the file history to find frequent contributor. Contributors can not be directly added as reviews but can be notified via a comment.
* If your code review has had no activity for 7days a maintainer should assist in reassigning the review.  If a maintainer has not updated your review, please reach out to @project maintainers.

#### 3. Merge or close

* Maintainers are responsible for handling final approval and merging of pull requests. Contributors can approve pull requests and provide input.

## Checklist

Below are a set of common questions that apply to all pull requests:
- [ ] Does the affected code have corresponding tests?
- [ ] Are the changes documented, not just with inline documentation, but also with conceptual documentation such as an overview of a new feature, or task-based documentation like a tutorial? Consider if this change should be announced on your project blog.
- [ ] Does this introduce breaking changes that would require an announcement or bumping the major version?

## Reading List

Reviewers are encouraged to read the following articles for help with common reviewer tasks:

* [The Art of Closing: How to closing an unfinished or rejected pull request](https://blog.jessfraz.com/post/the-art-of-closing/)
* [Kindness and Code Reviews: Improving the Way We Give Feedback](https://product.voxmedia.com/2018/8/21/17549400/kindness-and-code-reviews-improving-the-way-we-give-feedback)
* [Code Review Guidelines for Humans: Examples of good and back feedback](https://phauer.com/2018/code-review-guidelines/#code-reviews-guidelines-for-the-reviewer)
