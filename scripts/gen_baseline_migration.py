#!/usr/bin/env python3
"""从 schema_reference.sql 生成 golang-migrate baseline up/down 脚本。"""

from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REF = ROOT / "migrations" / "schema_reference.sql"
UP = ROOT / "migrations" / "000001_baseline.up.sql"
DOWN = ROOT / "migrations" / "000001_baseline.down.sql"

CREATE_TABLE_RE = re.compile(
    r"CREATE TABLE `(\w+)` \([\s\S]*?\) ENGINE=InnoDB[^;]*;",
    re.MULTILINE,
)


def extract_tables(text: str) -> list[tuple[str, str]]:
    tables: list[tuple[str, str]] = []
    for match in CREATE_TABLE_RE.finditer(text):
        tables.append((match.group(1), match.group(0).strip()))
    return tables


def write_up(tables: list[tuple[str, str]]) -> None:
    header = """-- Baseline schema (generated from schema_reference.sql)
-- Regenerate: make migrate-gen-baseline
-- Do NOT run schema_reference.sql on existing databases (contains DROP TABLE).

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

"""
    body = "\n\n".join(ddl for _, ddl in tables)
    footer = "\n\nSET FOREIGN_KEY_CHECKS = 1;\n"
    UP.write_text(header + body + footer, encoding="utf-8")


def write_down(tables: list[tuple[str, str]]) -> None:
    header = """-- Rollback baseline: drop all application tables
-- WARNING: destroys all data. Use in dev/test only.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

"""
    lines = [f"DROP TABLE IF EXISTS `{name}`;" for name, _ in reversed(tables)]
    footer = "\n\nSET FOREIGN_KEY_CHECKS = 1;\n"
    DOWN.write_text(header + "\n".join(lines) + footer, encoding="utf-8")


def main() -> None:
    if not REF.is_file():
        raise SystemExit(f"missing reference file: {REF}")

    text = REF.read_text(encoding="utf-8")
    tables = extract_tables(text)
    if not tables:
        raise SystemExit("no CREATE TABLE statements found in schema_reference.sql")

    write_up(tables)
    write_down(tables)
    print(f"generated {UP.relative_to(ROOT)} ({len(tables)} tables)")
    print(f"generated {DOWN.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
