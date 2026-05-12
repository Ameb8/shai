---
name: feature-designer
description: Transforms feature overviews into detailed technical design plans. Enforces a research-first approach to resolve ambiguities before designing. Provides structured breakdowns with subfeatures and git commit strategies.
---

# Feature Designer

## Overview

The `feature-designer` skill guides you through the process of taking a high-level feature overview and turning it into a concrete, technical implementation plan. It prioritizes clarity and precision, ensuring that all technical unknowns are addressed before code is written.

## Workflow

### 1. Research & Analysis
- **Read the Overview:** Carefully analyze the provided markdown document.
- **Identify Ambiguities:** Look for missing technical details, unclear requirements, or potential architectural conflicts.
- **Ask Questions:** **DO NOT GUESS.** If any part of the overview is underspecified, you MUST ask the user for clarification. Provide a numbered list of specific questions.
- **Wait for Input:** Only proceed to the design phase once you have sufficient information.

### 2. Design Planning
Once ambiguities are resolved, generate the technical design plan using the template in `references/design-template.md`.

### 3. Structure Requirements
- **Subfeature Breakdown:** Divide large features into logical, independent subfeatures that can be implemented and tested incrementally.
- **Git Commit Strategy:** Propose a sequence of atomic commits. Refer to `references/commit-strategy-guide.md` for best practices.

## Resources

- **[design-template.md](references/design-template.md):** The standard structure for technical design plans.
- **[commit-strategy-guide.md](references/commit-strategy-guide.md):** Guidelines for planning a clean git history.
