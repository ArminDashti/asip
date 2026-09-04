#!/usr/bin/env python3
"""Export asip-local Postgres tables to a D1-compatible SQL dump."""

from __future__ import annotations

import csv
import hashlib
import io
import subprocess
import sys
from pathlib import Path

OUT = Path(__file__).resolve().parents[1] / "dist" / "asip-d1-data.sql"
CONTAINER = "asip-local-postgres-1"
DB_USER = "postgres"
DB_NAME = "as_ip"
BATCH = 200

# table -> (columns, select SQL). Preserve IDs for FK integrity.
TABLES: list[tuple[str, list[str], str]] = [
    (
        "country",
        ["id", "country", "country_code"],
        'SELECT id, country, country_code FROM country ORDER BY id',
    ),
    (
        "origin",
        ["id", "origin", "description"],
        "SELECT id, origin, description FROM origin ORDER BY id",
    ),
    (
        "category",
        ["id", "category", "description"],
        "SELECT id, category, description FROM category ORDER BY id",
    ),
    (
        "role",
        ["id", "role", "description"],
        "SELECT id, role, description FROM role ORDER BY id",
    ),
    (
        "asn",
        [
            "id",
            "asn",
            '"as"',
            "country_id",
            "origin_id",
            "category_id",
            "network_role_id",
            "registered",
            "last_modified",
            "last_announced",
            "prefixes_last_modified",
        ],
        'SELECT id, asn, "as", country_id, origin_id, category_id, network_role_id, '
        "registered, last_modified, last_announced, prefixes_last_modified "
        "FROM asn ORDER BY id",
    ),
    (
        "ipv4_stat",
        [
            "id",
            "asn_id",
            "prefixes",
            "prefixes_aggregated",
            "largest_prefix",
            "total_addresses",
        ],
        "SELECT id, asn_id, prefixes, prefixes_aggregated, largest_prefix, total_addresses "
        "FROM ipv4_stat ORDER BY id",
    ),
    (
        "ipv6_stat",
        [
            "id",
            "asn_id",
            "prefixes",
            "prefixes_aggregated",
            "largest_prefix",
            "total_addresses",
        ],
        "SELECT id, asn_id, prefixes, prefixes_aggregated, largest_prefix, total_addresses "
        "FROM ipv6_stat ORDER BY id",
    ),
    (
        "connectivity",
        ["id", "asn_id", "provider_asn_id"],
        "SELECT id, asn_id, provider_asn_id FROM connectivity ORDER BY id",
    ),
    (
        "asn_ipv4",
        ["id", "asn_id", "cidr", "last_modified", "start_ip", "end_ip"],
        "SELECT id, asn_id, cidr, last_modified, start_ip, end_ip FROM asn_ipv4 ORDER BY id",
    ),
    (
        "asn_ipv6",
        ["id", "asn_id", "cidr", "last_modified"],
        "SELECT id, asn_id, cidr, last_modified FROM asn_ipv6 ORDER BY id",
    ),
    (
        "country_ipv4",
        ["id", "country_id", "cidr", "last_modified", "start_ip", "end_ip"],
        "SELECT id, country_id, cidr, last_modified, start_ip, end_ip "
        "FROM country_ipv4 ORDER BY id",
    ),
    (
        "country_ipv6",
        ["id", "country_id", "cidr", "last_modified"],
        "SELECT id, country_id, cidr, last_modified FROM country_ipv6 ORDER BY id",
    ),
    (
        "sync_state",
        ["id", "last_sync_at"],
        "SELECT id, last_sync_at::text FROM sync_state ORDER BY id",
    ),
]


def sql_literal(value: str | None) -> str:
    if value is None or value == "":
        # COPY WITH NULL '\\N' — empty string is real empty; None only for \\N
        return "NULL"
    return "'" + value.replace("'", "''") + "'"


def copy_csv(select_sql: str) -> csv.reader:
    cmd = [
        "docker",
        "exec",
        "-i",
        CONTAINER,
        "psql",
        "-U",
        DB_USER,
        "-d",
        DB_NAME,
        "-v",
        "ON_ERROR_STOP=1",
        "-c",
        f"COPY ({select_sql}) TO STDOUT WITH (FORMAT csv, NULL '\\N')",
    ]
    proc = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    assert proc.stdout is not None
    return csv.reader(proc.stdout), proc


def write_table(out, table: str, columns: list[str], select_sql: str) -> int:
    reader, proc = copy_csv(select_sql)
    col_list = ", ".join(columns)
    batch: list[str] = []
    count = 0

    def flush() -> None:
        nonlocal batch
        if not batch:
            return
        out.write(f"INSERT INTO {table} ({col_list}) VALUES\n")
        out.write(",\n".join(batch))
        out.write(";\n")
        batch = []

    for row in reader:
        vals = []
        for cell in row:
            if cell == "\\N":
                vals.append("NULL")
            else:
                # numerics without quotes when all digits / optional leading -
                if cell != "" and (
                    cell.isdigit()
                    or (cell[0] == "-" and cell[1:].isdigit())
                ):
                    vals.append(cell)
                else:
                    vals.append(sql_literal(cell))
        batch.append("(" + ", ".join(vals) + ")")
        count += 1
        if len(batch) >= BATCH:
            flush()
            if count % 50000 == 0:
                print(f"  {table}: {count} rows...", flush=True)

    flush()
    stderr = proc.stderr.read() if proc.stderr else ""
    code = proc.wait()
    if code != 0:
        raise RuntimeError(f"COPY failed for {table}: {stderr}")
    return count


def main() -> int:
    OUT.parent.mkdir(parents=True, exist_ok=True)
    print(f"Writing {OUT}", flush=True)
    with OUT.open("w", encoding="utf-8", newline="\n") as out:
        # D1 import rejects BEGIN/COMMIT/SAVEPOINT and PRAGMA transaction APIs.
        # Clear data tables (keep api_request). FK order: children first.
        for table in (
            "country_ipv6",
            "country_ipv4",
            "asn_ipv6",
            "asn_ipv4",
            "connectivity",
            "ipv6_stat",
            "ipv4_stat",
            "asn",
            "role",
            "category",
            "origin",
            "country",
            "sync_state",
        ):
            out.write(f"DELETE FROM {table};\n")

        total = 0
        for table, columns, select_sql in TABLES:
            print(f"Exporting {table}...", flush=True)
            n = write_table(out, table, columns, select_sql)
            print(f"  {table}: {n} rows", flush=True)
            total += n

    data = OUT.read_bytes()
    md5 = hashlib.md5(data).hexdigest()
    print(f"Done: {total} rows, {OUT.stat().st_size} bytes, md5={md5}", flush=True)
    (OUT.parent / "asip-d1-data.md5").write_text(md5 + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
