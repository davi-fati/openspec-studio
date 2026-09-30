## Purpose

Renders Mermaid diagrams embedded in markdown content wherever the Studio displays proposals, designs, or specs.

## ADDED Requirements

### Requirement: Render Mermaid code fences as diagrams
The system SHALL detect fenced code blocks marked `mermaid` in rendered markdown and render them as diagrams instead of raw text.

#### Scenario: Valid mermaid block renders as diagram
- **WHEN** a markdown document contains a fenced code block with the `mermaid` language tag and valid diagram syntax
- **THEN** the Studio renders it as a diagram in place of the code block

#### Scenario: Invalid mermaid syntax fails gracefully
- **WHEN** a `mermaid` code block contains invalid syntax
- **THEN** the Studio shows a rendering error inline for that block without breaking the rest of the document's rendering

### Requirement: Diagram rendering respects theme
The system SHALL render Mermaid diagrams using colors consistent with the Studio's active light/dark theme.

#### Scenario: Theme switch updates diagram colors
- **WHEN** the user toggles between light and dark theme
- **THEN** already-rendered diagrams update their color scheme to match
