---
name: gh-pr-drafter
description: Generate a markdown PR message for merging the current branch into master. Writes the draft to .github/pr-draft/<branch-name>.md. Use when you want to prepare a Pull Request description based on current changes.
---

# GH PR Drafter

This skill helps you generate a structured Pull Request message for merging your current branch into `master`.

## Workflow

1.  **Identify Current Branch**: Determine the current git branch name.
2.  **Analyze Changes**: Compare the current branch with `master` to identify key changes, commits, and impacted files.
    -   `git log master..HEAD --oneline`
    -   `git diff master..HEAD --stat`
3.  **Draft PR Message**: Use `assets/pr-template.md` as a base and fill in the details based on the analysis.
4.  **Save Draft**: Write the resulting markdown to `.github/pr-draft/<branch-name>.md`.

## Example Usage

**User**: "Draft a PR message for this branch."

**Action**:
1.  Run `git branch --show-current` (e.g., returns `feature/add-dry-run`).
2.  Run `git log master..HEAD --oneline` and `git diff master..HEAD --stat`.
3.  Read `assets/pr-template.md`.
4.  Synthesize the summary, key changes, and commits.
5.  Write the PR draft to `.github/pr-draft/feature-add-dry-run.md`.

## Guidelines

- **File Naming**: Use the kebab-case version of the branch name for the filename.
- **Location**: Always place drafts in the `.github/pr-draft/` directory.
- **Assumptions**: Assume current branch and `master` are up to date with origin.
- **Content**: Be concise but descriptive. Highlight the "why" as well as the "what".
