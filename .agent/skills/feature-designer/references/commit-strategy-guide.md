# Git Commit Strategy Guide

A good commit history makes code reviews easier and helps in tracking down bugs. Follow these patterns when designing a feature implementation:

## 1. Atomic Commits
Each commit should do one thing and do it completely. If you're adding a new endpoint, separate the data model changes from the handler logic if they are large enough.

## 2. Logical Sequencing
1. **Infrastructure/Schema:** Migrations, new dependencies, configuration changes.
2. **Interfaces/Types:** Shared types, interfaces, or protocol buffers.
3. **Core Logic:** Internal packages, services, or business logic.
4. **Integration:** API handlers, controllers, or command-line interfaces.
5. **UI/Front-end:** View changes, components, or client-side logic.
6. **Polishing:** Documentation updates, final linting, or cleanup.

## 3. Commit Message Standards
- Use the imperative mood ("Add feature X" not "Added feature X").
- Keep the subject line under 50 characters.
- Provide more detail in the body if the "why" isn't obvious.

## 4. Examples
- `feat: add user_preferences table migration`
- `feat: implement UserPreferenceStore in internal/db`
- `feat: add GET /api/preferences endpoint`
- `test: add integration tests for preferences API`
- `docs: update API documentation for preferences`
