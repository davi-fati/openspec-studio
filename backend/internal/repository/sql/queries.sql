-- name: ListProjects :many
SELECT id, name, path, last_opened_at FROM projects ORDER BY last_opened_at DESC;

-- name: GetProjectByPath :one
SELECT id, name, path, last_opened_at FROM projects WHERE path = ?;

-- name: GetProjectByID :one
SELECT id, name, path, last_opened_at FROM projects WHERE id = ?;

-- name: UpsertProject :one
INSERT INTO projects (name, path, last_opened_at)
VALUES (?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
    name = excluded.name,
    last_opened_at = excluded.last_opened_at
RETURNING id, name, path, last_opened_at;

-- name: TouchProject :exec
UPDATE projects SET last_opened_at = ? WHERE path = ?;

-- name: ListAIProviders :many
SELECT id, name, kind, model, endpoint, cli_path, cli_args, is_default, created_at FROM ai_providers ORDER BY created_at ASC;

-- name: GetAIProvider :one
SELECT id, name, kind, model, endpoint, cli_path, cli_args, is_default, created_at FROM ai_providers WHERE id = ?;

-- name: CreateAIProvider :one
INSERT INTO ai_providers (name, kind, model, endpoint, cli_path, cli_args, is_default, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, kind, model, endpoint, cli_path, cli_args, is_default, created_at;

-- name: UpdateAIProvider :one
UPDATE ai_providers SET name = ?, model = ?, endpoint = ?, cli_path = ?, cli_args = ?
WHERE id = ?
RETURNING id, name, kind, model, endpoint, cli_path, cli_args, is_default, created_at;

-- name: DeleteAIProvider :exec
DELETE FROM ai_providers WHERE id = ?;

-- name: ClearDefaultAIProvider :exec
UPDATE ai_providers SET is_default = 0;

-- name: SetDefaultAIProvider :exec
UPDATE ai_providers SET is_default = 1 WHERE id = ?;

-- name: ListSpecflows :many
SELECT id, project_id, provider_id, scheduled_at, status, missed, created_at FROM specflows ORDER BY scheduled_at ASC;

-- name: GetSpecflow :one
SELECT id, project_id, provider_id, scheduled_at, status, missed, created_at FROM specflows WHERE id = ?;

-- name: ListDueSpecflows :many
SELECT id, project_id, provider_id, scheduled_at, status, missed, created_at FROM specflows
WHERE status = 'pending' AND missed = 0 AND scheduled_at <= ?;

-- name: CreateSpecflow :one
INSERT INTO specflows (project_id, provider_id, scheduled_at, status, missed, created_at)
VALUES (?, ?, ?, 'pending', 0, ?)
RETURNING id, project_id, provider_id, scheduled_at, status, missed, created_at;

-- name: UpdateSpecflowStatus :exec
UPDATE specflows SET status = ? WHERE id = ?;

-- name: MarkSpecflowMissed :exec
UPDATE specflows SET missed = 1 WHERE id = ? ;

-- name: DeleteSpecflow :exec
DELETE FROM specflows WHERE id = ?;

-- name: DeleteSpecflowItems :exec
DELETE FROM specflow_items WHERE flow_id = ?;

-- name: ListSpecflowItems :many
SELECT id, flow_id, position, change_name, status, log, started_at, finished_at FROM specflow_items
WHERE flow_id = ? ORDER BY position ASC;

-- name: CreateSpecflowItem :one
INSERT INTO specflow_items (flow_id, position, change_name, status)
VALUES (?, ?, ?, 'pending')
RETURNING id, flow_id, position, change_name, status, log, started_at, finished_at;

-- name: UpdateSpecflowItemStatus :exec
UPDATE specflow_items SET status = ?, log = ?, started_at = ?, finished_at = ? WHERE id = ?;

-- name: UpdateSpecflowItemPosition :exec
UPDATE specflow_items SET position = ? WHERE id = ? AND flow_id = ?;
