# Expected Behavior

- Installer with `--version <candidate-tag>` installs runtime that reports the same tag.
- Installer provenance includes requested tag, resolved tag, asset URL, and installed runtime version.
- Homebrew tap install resolves to candidate-aligned formula version when channel is declared in-scope.
- Uninstall/reinstall cycles preserve deterministic version outcomes.
- Any lineage or channel ambiguity is treated as certification failure.
