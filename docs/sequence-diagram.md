# OpenSpec Studio — End-to-End Sequence

Covers the full lifecycle the six planned changes implement: opening a project,
authoring a spec, scheduling it in Specflow, and unattended implementation via
a configured AI provider, with real-time updates pushed over SSE.

```mermaid
sequenceDiagram
    actor User
    participant Shell as Tauri Shell
    participant BE as Go Backend (MVC)
    participant FS as Project openspec/
    participant AI as AI Provider (Hosted API / CLI Agent)
    participant FE as React Frontend

    User->>Shell: Launch OpenSpec Studio
    Shell->>BE: Start backend process
    BE-->>Shell: Ready (listening on 127.0.0.1)
    Shell->>FE: Show main window
    FE->>BE: GET /api/projects
    BE-->>FE: Registered projects list
    FE-->>User: Overview tab (default)

    User->>FE: Open Project (choose directory)
    FE->>BE: POST /api/projects
    BE->>FS: Validate / scaffold openspec/ structure
    FS-->>BE: OK
    BE-->>FE: Project registered
    FE->>BE: GET /api/overview
    BE-->>FE: Per-project progress
    FE-->>User: Overview cards update

    User->>FE: New Spec (describe capability)
    FE->>BE: POST /api/specs/generate {description}
    BE->>AI: Invoke(prompt)
    AI-->>BE: Draft spec (requirements + scenarios)
    BE-->>FE: Draft for review
    User->>FE: Edit + Save
    FE->>BE: POST /api/specs {content, provenance}
    BE->>FS: Write specs/<capability>/spec.md (+ metadata)
    FS-->>BE: Written
    BE-->>FE: Spec created
    BE--)FE: SSE update (file changed)
    FE-->>User: Kanban / Specs views refresh

    User->>FE: Build Specflow (select planned changes, order, schedule)
    FE->>BE: POST /api/specflow/flows {items, schedule, provider}
    BE->>BE: Validate deps + planning completeness
    BE-->>FE: Flow scheduled (pending)

    Note over BE: Scheduled time reached (e.g. overnight window)
    BE->>BE: Scheduler triggers flow execution
    loop each flow item, in order
        BE->>AI: Invoke implementation (change N)
        AI-->>BE: Stream logs / result
        BE->>FS: Apply implementation changes
        BE--)FE: SSE status update (running → succeeded/failed)
        FE-->>User: Kanban card + Overview reflect status
    end

    alt item fails
        BE-->>FE: Flow halted, failure surfaced
        User->>FE: Review logs, fix, or cancel
    else all items succeed
        BE-->>FE: Flow completed
        FE-->>User: Overview shows updated progress
    end
```
