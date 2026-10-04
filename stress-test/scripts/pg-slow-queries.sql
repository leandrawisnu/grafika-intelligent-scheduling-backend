-- Snapshot slow queries after k6 run (Postgres with pg_stat_statements enabled).
-- Usage: psql "$DATABASE_URL" -f stress-test/scripts/pg-slow-queries.sql

SELECT
  calls,
  round(mean_exec_time::numeric, 2) AS mean_ms,
  round(max_exec_time::numeric, 2) AS max_ms,
  round(total_exec_time::numeric, 2) AS total_ms,
  left(regexp_replace(query, '\s+', ' ', 'g'), 120) AS query_preview
FROM pg_stat_statements
WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
ORDER BY mean_exec_time DESC
LIMIT 25;
