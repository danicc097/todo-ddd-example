-- name: GetDailySchedule :one
SELECT
  *
FROM
  daily_schedules
WHERE
  user_id = $1
  AND date = $2;

-- name: UpsertDailySchedule :one
INSERT INTO daily_schedules(user_id, date, max_capacity, version)
  VALUES ($1, $2, $3, 1)
ON CONFLICT (user_id, date)
  DO UPDATE SET
    max_capacity = EXCLUDED.max_capacity,
    version = daily_schedules.version + 1
  WHERE
    daily_schedules.version = sqlc.arg(current_version)
  RETURNING
    *;

-- name: GetScheduleTasks :many
SELECT
  *
FROM
  schedule_tasks
WHERE
  user_id = $1
  AND date = $2;

-- name: BulkUpsertScheduleTasks :exec
INSERT INTO schedule_tasks(user_id, date, todo_id, energy_cost)
SELECT
  UNNEST(sqlc.arg(user_ids)::uuid[]),
  UNNEST(sqlc.arg(dates)::timestamptz[]),
  UNNEST(sqlc.arg(todo_ids)::uuid[]),
  UNNEST(sqlc.arg(energy_costs)::integer[])
ON CONFLICT (user_id,
  date,
  todo_id)
  DO UPDATE SET
    energy_cost = EXCLUDED.energy_cost;

-- name: RemoveMissingTasksFromSchedule :exec
DELETE FROM schedule_tasks
WHERE user_id = $1
  AND date = $2
  AND NOT (todo_id = ANY (sqlc.arg(todo_ids)::uuid[]));

-- name: GetSchedulesByTodoID :many
SELECT
  user_id,
  date
FROM
  schedule_tasks
WHERE
  todo_id = $1;

