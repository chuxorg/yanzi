# AI Execution Rules

Before executing any development task:
- Ignore changes in all YANZI_LOG.* and issues files.
- Confirm work is on a feature branch created from development branch.
- Confirm no direct commits to development or master.
- After each task:
  - Stage
  - Commit
  - Push
- At phase completion:
  - Create PR to development.
  - Do not create tags before merge.
  - Always follow docs/RELEASE_PROTOCOL.md
  - Always update code comments for documentation
  - Always follow docs/CODE_DOCUMENTATION.md
  - All new and updated code that is testable must be unit tested.
  - All builds and unit tests must pass be for a task is pushed to the remote repo
  - Human will indicate when phase is complete


  
