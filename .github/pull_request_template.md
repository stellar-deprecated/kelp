# Guidelines for submitting Pull Requests to Kelp

Questions about technical design should be discussed before starting a PR. This can be discussed in the Github issue (preferred), or over IM (keybase / slack).

## Handling of PRs

Typically, this is what is expected from the author of a Pull Request:

1. Each PR should correspond to the *smallest independent logical unit* of change.
    - each PR should be prefixed with "bugfix", "feature", or "refactor".
    - refactors should never be mixed in with logic changes or bug fixes and vice-versa.
    - refactors that change variable names across more than 1 function should be a separate PR altogether.
1. Author should add inline commentary on the code changes to "guide" the reviewer on how to read the code change, highlighting key changes made.
    - this should typically be 2-3 places, as anything bigger is too big of a PR and should be split up.
1. We should aim for the following benchmarks in terms of PR efficiency:
    1. In 60% of the cases the PR should be approved without any changes requested from the reviewer.
    1. In 25% of the cases the reviewer should add 2-3 comments that need to be "fixed up". This is most likely a code style suggestion.
    1. In 10% of the cases the reviewer should request a structural change to the code. These should not be "new ideas".
    1. In <=5% of the cases the PR may be discarded because we have learned something new in the process of implementation.
1. Author makes requested changes.
    - no "new ideas" to be introduced at this stage.
    - no structural changes to be introduced beyond what is requested.
    - focus should be on minimizing "changes" beyond what is requested by reviewer.
1. Reviewer re-reviews the code.
    - result of the PR should be that the PR has moved to the next stage in the review ladder described above.
    - this means that a PR will take a maximum of 2 review rounds to get to completion.

### Draft PRs

- If there is a need to "try out the change" and get feedback, we can use a draft PR, although this should be used sparingly.
- If a PR starts off in draft state then it should be used more as a "whiteboard for discussion" as opposed to the early stage of a PR.
- In most cases, Draft PRs should be discarded. To avoid noise from comments, Draft PRs should be resubmitted as fresh standalone PRs.
- Only in rare cases should Draft PRs be converted to actual PRs by request of the reviewer.


