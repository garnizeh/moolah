---
description: finish-task - Complete a roadmap task and update documentation
---

This workflow automates the finalization of a task, ensuring branches are created and documentation is consistent.

### Workflow Steps

1. **Identify the Task**: Find the relevant task file in `docs/tasks/milestone-01-mvp/`.

2. **Branch Creation**:
   - Format: `task/mvp-[phase]-[slug-name]`
   - Example: `git checkout -b task/mvp-01-technical-foundation`

3. **Update Task Status**:
   - In the task file (e.g., `phase-01-technical-foundation.md`), update:
     - `Status:` to ✅ `done`
     - All relevant sub-tasks to `[x]`

4. **Update Roadmap**:
   - In `docs/tasks/roadmap.md`, find the task and mark it as complete `[x]`.

5. **Commit and Sync**:
   - Stage all changes: `git add .`
   - Commit: `git commit -m "Finish task: [Task Title]"`
   - Note: Do NOT push unless explicitly asked by the USER.

6. **Summary**: Provide the USER with a summary of the files updated and the branch name.
