---
description: finish-task - Complete a roadmap task and update documentation
---

This workflow automates the finalization of a task, ensuring branches are created and documentation is consistent.

### Workflow Steps

1. **Verification**:
   // turbo
   - Run `make check-ci` to ensure the project meets all standards.
   - If errors occur, resolve them and re-run `make check-ci`.
   - **Crucial**: Do NOT proceed until `make check-ci` passes with no errors.

2. **Identify the Task**: Find the relevant task file in `docs/tasks/milestone-01-mvp/`.

3. **Branch Creation**:
   - Format: `task/mvp-[phase]-[slug-name]`
   - Example: `git checkout -b task/mvp-01-technical-foundation`

4. **Update Task Status**:
   - In the task file (e.g., `phase-01-technical-foundation.md`), update:
     - `Status:` to ✅ `done`
     - All relevant sub-tasks to `[x]`

5. **Update Roadmap**:
   - In `docs/tasks/roadmap.md`, find the task and mark it as complete `[x]`.

6. **Commit and Sync**:
   - Stage all changes: `git add .`
   - Commit: `git commit -m "Finish task: [Task Title]"`
   - Note: Do NOT push unless explicitly asked by the USER.

7. **Summary**: Provide the USER with a summary of the files updated and the branch name.

