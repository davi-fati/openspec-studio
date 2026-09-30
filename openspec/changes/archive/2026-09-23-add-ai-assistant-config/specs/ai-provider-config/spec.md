## Purpose

Lets the user configure, select, and health-check one or more AI providers - hosted APIs or local CLI coding agents - that other Studio features (spec generation, specflow implementation) call into.

## ADDED Requirements

### Requirement: Configure a hosted API provider
The system SHALL let the user add a hosted AI API provider by specifying provider type, API key, and model, and SHALL persist this configuration locally.

#### Scenario: Add hosted provider
- **WHEN** the user enters a valid provider type, API key, and model, and saves
- **THEN** the provider appears in the configured providers list and is available for selection

### Requirement: Configure a CLI agent provider
The system SHALL let the user add a local CLI-based agent provider (for example, Claude Code CLI or another agent CLI) by specifying the command/executable path and any required arguments.

#### Scenario: Add CLI provider
- **WHEN** the user enters a valid local executable path for a CLI agent and saves
- **THEN** the provider appears in the configured providers list marked as a CLI-based provider

#### Scenario: CLI executable not found
- **WHEN** the user enters a path that does not resolve to an executable
- **THEN** the Studio shows a validation error and does not save an unresolvable provider silently

### Requirement: Provider health check
The system SHALL let the user run a connectivity/health check against any configured provider and SHALL show whether it succeeded or failed.

#### Scenario: Hosted provider health check
- **WHEN** the user triggers a health check on a hosted API provider
- **THEN** the Studio makes a minimal request to that provider and reports success or a specific failure reason

#### Scenario: CLI provider health check
- **WHEN** the user triggers a health check on a CLI agent provider
- **THEN** the Studio invokes the CLI with a no-op/version command and reports success or failure

### Requirement: Select active provider per use
The system SHALL let the user designate a default provider and, where a feature invokes AI assistance (spec generation, specflow), SHALL let the user choose among configured providers for that specific use.

#### Scenario: Default provider used when none specified
- **WHEN** a feature requests AI assistance without an explicit provider choice
- **THEN** the Studio uses the configured default provider

#### Scenario: Override provider for a single request
- **WHEN** the user explicitly selects a non-default provider for a generation request
- **THEN** that request is routed to the selected provider instead of the default

### Requirement: Local, provider-scoped credential storage
The system SHALL store provider credentials (API keys, CLI paths) locally and SHALL only transmit a credential to the provider it was configured for, never to any other endpoint.

#### Scenario: Credentials stay local
- **WHEN** the user configures a provider's API key
- **THEN** the key is stored in local OS-level secret storage (or an encrypted local file if unavailable) and is not included in any request to a different provider or to Studio telemetry
