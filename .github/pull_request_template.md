# Guidelines for submitting code to Kelp

## Before Submitting a PR

1. Problem Identification
    1. ensure you understand the problem / task.
    1. ask initial questions to understand the task described in the Github issue, which may not be clearly defined.
    1. you can reach the team on keybase or by commenting on the Github issue.
    1. this is not the point to discuss concrete solutions (concrete design or implementation), but only what the problem is and the desired outcome.
    1. ask any questions related to scope of work, breakdown of tasks into subtasks and PRs.
    1. typically this should be covered in the Github issue. If it is not, then add to the issue after discussion with code owner.
    1. author needs to be comfortable with breakdown and scope of tasks. Scope should be small. Think about the overhead on the code reviewer when breaking down the tasks.

1. Research Solution
    1. conduct research on the issue at hand. Identify point of code change and any “fallout” from the change (tip: search the codebase for consumers of the function / object, and their consumers etc., until you get to a part of the code that you are well-versed with).
    1. author needs to become the “expert” on the part of the code being touched before submitting a code change.
    1. If it will take you 2 more days to figure something out by reading the code or experimenting with the codebase then do it.
        1. This is not wasted effort.
        1. It is likely that the code reviewer would use a similar process to figure it out, but this is your job and not the code reviewer’s job.
    1. If the process of becoming an “expert” on a specific area would take you more than 2 days then discuss this with the code reviewer if it is worth it, and proceed accordingly.
    1. Think about what kind of tests you would want. Are you going with a black-box testing approach, or a white-box testing approach? Why?
1. Technical Design Discussion
    1. PRs are not the best medium to conduct design discussions. The team is readily accessible and available so 1-1 meetings can be very effective.
    1. This is the best avenue to discuss your ideas for how to approach the problem.
    1. Prepare any relevant technical documentation for the design discussion and post on the Github issue and ping code owner.
        1. This can be in the form of a google document, flowchart diagram, or in rare cases a code snippet in a Draft PR.
        1. The Draft PR should not be reviewed by the code reviewer as part of the code change but can be used as a space to get feedback
        1. Resulting comments on a Draft PR can capture any guidance from the technical discussion.
        1. All Draft PRs should be discarded and never converted to a non-draft PR to avoid the overhead of comments lingering from the design discussion using the Draft PR.
    1. Schedule technical design discussion to go over the proposed solution.
    1. Include a list of tests you need to add, remove, modify. Consider all possibilities (tip: data related tests would require many test cases).
1. Implementation
    1. implement feature or bug fix. This should be the first line of code that you write for the PR.
    1. Avoid for loops to “autogenerate” test cases, unless absolutely necessary.
1. Pull Request
    1. The first version of the PR should not have any “surprises” compared to what was discussed in the technical design.
    1. See [Handling of PRs](#handling-of-prs) section below for guide on how to get a PR to merged status.
    1. See [detailed guide on submitting code-reviewer-friendly PRs](https://mtlynch.io/code-review-love/). This guide will save a lot of time for both author and reviewer and will help speed up the turnaround time on PRs.

## Handling of PRs

Typically, this is what is expected from the author of a Pull Request:

1. Each PR should correspond to the *smallest independent logical unit* of change.
    1. each PR should be prefixed with "bugfix", "feature", or "refactor".
    1. refactors should never be mixed in with logic changes or bug fixes and vice-versa.
    1. refactors that change variable names across more than 1 function should be a separate PR altogether.
1. Author should add inline commentary on the code changes to "guide" the reviewer on how to read the code change, highlighting key changes made.
    1. this should typically be 2-3 places, as anything bigger is too big of a PR and should be split up.
1. We should aim for the following benchmarks in terms of PR efficiency:
    1. In 60% of the cases the PR should be approved without any changes requested from the reviewer.
    1. In 25% of the cases the reviewer should add 2-3 comments that need to be "fixed up". This is most likely a code style suggestion.
    1. In 10% of the cases the reviewer should request a structural change to the code. These should not be "new ideas".
    1. In <=5% of the cases the PR may be discarded because we have learned something new in the process of implementation.
1. Author makes requested changes.
    1. no "new ideas" to be introduced at this stage.
    1. no structural changes to be introduced beyond what is requested.
    1. focus should be on minimizing "changes" beyond what is requested by reviewer.
    1. person who created a comment / raised a question on the PR is responsible to mark it as resolved once they are satisfied. The developer should not mark questions raised by the reviewer as resolved since the reviewer uses these comments as a way to track what they are looking for in the next iteration of the review (tip: use an emoji such as the checkbox emoji to mark an inline comment as handled or addressed).
1. Reviewer re-reviews the code.
    1. result of the PR should be that the PR has moved to the next stage in the review ladder described above.
    1. this means that a PR will take a maximum of 2 review rounds to get to completion.
    1. back to the previous step for the author to make changes based on the re-review.

## Draft PRs

1. If there is a need to "try out the change" and get feedback, we can use a draft PR, although this should be used sparingly.
1. If a PR starts off in draft state then it should be used more as a "whiteboard for discussion" as opposed to the early stage of a PR.
1. In most cases, Draft PRs should be discarded. To avoid noise from comments, Draft PRs should be resubmitted as fresh standalone PRs instead of converting the draft PR to a regular PR.
