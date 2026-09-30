## Why

Spec generation and, eventually, spec implementation (specflow) need an AI backend, and different users/teams already have different AI tooling: hosted model APIs, or local coding agents like Claude Code CLI or other agent CLIs. The Studio should not hard-code one provider; it needs a configuration surface where the user chooses and configures the AI provider(s) used for generation and later for implementation.

## What Changes

- Add an AI provider settings section where the user configures one or more AI providers: hosted API providers (e.g., API key + model selection) and CLI-based agent providers (e.g., Claude Code CLI, generic agent CLI) that the Studio invokes as a local process.
- Add a provider selection mechanism used by spec-generation (from `add-project-spec-management`) and later by specflow: which configured provider handles a given generation/implementation request.
- Add credential handling: API keys and CLI paths are stored locally and never transmitted anywhere except to the provider they belong to.
- Add a connectivity/health check per configured provider so the user knows before relying on it whether it is actually reachable/runnable.

## Capabilities

### New Capabilities
- `ai-provider-config`: Configuring, storing, and health-checking one or more AI providers (hosted API or local CLI agent) usable across the Studio.

## Impact

- Depends on `add-studio-foundation` (settings persistence pattern, backend API).
- `add-project-spec-management`'s spec-generation endpoint becomes provider-aware: it SHALL call the currently selected provider rather than a fixed implementation.
- New local secret storage for API keys (OS keychain where available, otherwise encrypted local file).
