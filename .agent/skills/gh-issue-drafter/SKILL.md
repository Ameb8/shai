---
name: gh-issue-drafter
description: Generate detailed draft GitHub issues for new features in .github/issues-draft/**. Use when the user wants to document a new feature idea or enhancement, following a standard logical template.
---

# GH Issue Drafter

This skill helps you create structured, detailed draft GitHub issues for new features or enhancements.

## Workflow

1.  **Gather Context**: Understand the feature's purpose, requirements, and potential impact.
2.  **Select Template**: Use the `assets/feature-issue-template.md` as a base.
3.  **Refine Content**: Follow the principles in `references/template-guidelines.md` to ensure the draft is high-quality.
4.  **Create File**: Save the draft in `.github/issues-draft/<kebab-case-name>.md`.

## Example Usage

**User**: "Draft an issue for adding a 'dry-run' flag to the query command."

**Action**:
1.  Read `assets/feature-issue-template.md`.
2.  Gather details about the 'dry-run' flag (e.g., skip execution, show what would happen).
3.  Draft the content following `references/template-guidelines.md`.
4.  Write to `.github/issues-draft/dry-run-flag.md`.

## Guidelines

- **File Naming**: Use kebab-case for the filename (e.g., `my-new-feature.md`).
- **Location**: Always place drafts in the `.github/issues-draft/` directory.
- **Completeness**: Ensure all sections of the template are filled with meaningful content.
