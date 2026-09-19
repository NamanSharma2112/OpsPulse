#!/bin/sh
# Applies the schema when the Postgres data volume is first created.
#
# The migrations directory is mounted at /migrations rather than directly into
# /docker-entrypoint-initdb.d, because that directory is executed in plain
# alphabetical order — which would interleave the .down.sql files with the
# .up.sql ones. This applies only the up migrations, in filename order.
set -eu

for file in /migrations/*.up.sql; do
	echo "init-db: applying $(basename "$file")"
	psql -v ON_ERROR_STOP=1 \
		--username "$POSTGRES_USER" \
		--dbname "$POSTGRES_DB" \
		--quiet \
		--file "$file"
done

echo "init-db: schema ready"
