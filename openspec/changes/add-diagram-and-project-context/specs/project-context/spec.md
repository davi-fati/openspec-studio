## Purpose

Tracks a project's Reviews and Tech Debt alongside its OpenSpec specs, giving day-to-day engineering context a home inside the Studio.

## ADDED Requirements

### Requirement: Create and list review notes
The system SHALL let the user create a review note for a project, optionally linked to a specific spec or change, and list all review notes for that project.

#### Scenario: Create review linked to a change
- **WHEN** the user creates a review note from a change's detail view
- **THEN** the note is saved under the project's context and associated with that change's id

#### Scenario: Create standalone review
- **WHEN** the user creates a review note without selecting a specific spec/change
- **THEN** the note is saved as a project-level review with no artifact link

### Requirement: Create and track tech-debt items
The system SHALL let the user create a tech-debt item with a description, status (e.g., open/in-progress/resolved), and an optional link to the spec/change that would resolve it.

#### Scenario: Create debt item with resolving spec link
- **WHEN** the user creates a tech-debt item and links it to a spec
- **THEN** the item shows that link, and the spec's detail view shows the linked debt item

#### Scenario: Update debt item status
- **WHEN** the user changes a debt item's status to resolved
- **THEN** the change is persisted and reflected wherever the item is listed

### Requirement: Project context persists with the project
The system SHALL persist review notes and tech-debt items inside the project's own directory tree, so they travel with the project rather than being Studio-only state.

#### Scenario: Context survives re-opening the project elsewhere
- **WHEN** the project directory (including its context data) is opened in a different Studio installation
- **THEN** all review notes and tech-debt items are present without re-entry
