---
name: go-comment-style
description: Enforce Go documentation standards and project-specific comment styles for all declarations.
---

# Goal
Standardize all code comments to follow Go conventions and project-specific readability requirements.

# Rules
- **Standard Conventions:** Follow `go doc` and `godoc` standards strictly.
- **The "Name-Start" Rule:** The first word of a doc comment must be the exact name of the symbol.
- **Grammar:** Use complete sentences with proper punctuation.
- **Content:** Explain *intent* (the "why"), not the implementation (the "how").
- **Conciseness:** 2-3 lines for functions/methods, 1-2 lines for structs.
- **Internal Logic:** Add brief (1-line) comments to notable blocks within function bodies.
- **fmt** Run `go fmt <filepath>` after all changes have been implemented

# Examples

## Functions
**Bad:**
// This function calculates the SHA256 hash of a file.
func HashFile(path string) string { ... }

**Good:**
// HashFile returns the SHA256 checksum of the file at the given path.
// It returns an empty string if the file is inaccessible or corrupted.
func HashFile(path string) string { ... }

## Struct/Type/Interface
**Bad:**
// Configuration for the microservice.
type Config struct { ... }

**Good:**
// Config defines the environment settings for the orchestration engine.
type Config struct { ... }

## Internal Logic Block
**Bad:**
// Loop through the items and check if they are valid
for _, item := range items { ... }

**Good:**
// Validate item integrity before batch processing to prevent partial commits.
for _, item := range items { ... }

# Preservation Rules
- Do not modify executable logic or rename identifiers.
- Only improve or add comments.