## Context

Consumed by `add-project-spec-management`'s spec-generation flow now, and by `add-specflow-scheduler`'s implementation runs later. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One provider-invocation abstraction usable for both short text-generation calls (spec drafting) and longer-running agent processes (future implementation runs), without hard-coding a specific vendor.
- Credentials never leave the local machine except to their own provider.

**Non-Goals:**
- Implementing the actual implementation-run orchestration (scheduling, dependency ordering) - that belongs to `add-specflow-scheduler`; this change only defines how a provider is configured and invoked.
- Usage/cost tracking or rate-limit management across providers.

## Decisions

- **One `Provider` interface with two implementations**: `HostedAPIProvider` (HTTP call with API key + model) and `CLIAgentProvider` (spawns the configured executable, streams stdout/stderr). Both expose the same `Invoke(prompt) -> result/stream` shape so callers (spec generation, specflow) don't need to know which kind they're using. Alternative considered: separate code paths per provider type in each caller - rejected, it would duplicate provider-selection logic in every feature that needs AI assistance.
- **Credential storage via OS keychain** (e.g., macOS Keychain, or equivalent per platform) with an encrypted-local-file fallback when no keychain is available, rather than plaintext config - matches user expectation for API keys on a desktop app.
- **Health check is a lightweight, explicit user action**, not automatic background polling, to avoid unexpected API usage/cost or spawning CLI processes without the user's knowledge.

## Risks / Trade-offs

- CLI agent providers run arbitrary local executables the user points at → the Studio trusts whatever path the user configures; no sandboxing beyond what the CLI tool itself provides. Mitigated by requiring explicit user configuration (never auto-discovered/run) and showing the resolved path before saving.
- Hosted API keys stored locally are only as safe as the OS keychain/encrypted fallback → acceptable for a local desktop tool; documented as a security-relevant integration point.
