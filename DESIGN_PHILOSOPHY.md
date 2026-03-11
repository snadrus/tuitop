# Design Philosophy for AI

When modifying `tuios-fork`, follow these principles:

## Minimal Changes

- Keep changes to tuios-fork as **minimal** as possible
- Every change must include a **PR.md** (or equivalent) description explaining:
  - What changed and why
  - How it integrates with the host application

## No Constants

- **No hardcoded constants** in tuios-fork for theme, colors, or host-specific values
- **All values must be passed in** from the host application (e.g. `main.go`) via:
  - `UserConfig` / `WithUserConfig`
  - Options / functional options
  - Other injection mechanisms

The fork is a library. The host application owns the design decisions; the fork should accept configuration, not define it.
