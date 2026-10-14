"""Export the old database read-only; only needed once to preserve legacy data."""
import argparse
import base64
import json
import sqlite3
from pathlib import Path

TABLES = {
    "gamification_users": ("user_id", "state"),
    "accounts": ("email", "state"),
    "billing_purchases": ("id", "state"),
    "billing_tokens": ("user_id", "token"),
    "deleted_demo": ("id",),
    "dictionary_responses": ("cache_key", "source", "body", "fetched_at"),
    "dictionary_entries": ("source", "entry_id", "kind", "body", "fetched_at"),
}

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    with sqlite3.connect(args.source.resolve().as_uri() + "?mode=ro", uri=True) as db:
        db.execute("BEGIN")
        existing = {row[0] for row in db.execute("SELECT name FROM sqlite_master WHERE type='table'")}
        tables = {}
        for table, columns in TABLES.items():
            rows = db.execute(f"SELECT {','.join(columns)} FROM {table}") if table in existing else []
            tables[table] = [
                {column: base64.b64encode(value).decode() if isinstance(value, bytes) else value
                 for column, value in zip(columns, row)}
                for row in rows
            ]
    args.output.parent.mkdir(parents=True, exist_ok=True)
    # Exclusive creation prevents accidentally replacing an existing export.
    with args.output.open("x", encoding="utf-8") as file:
        json.dump({"version": 1, "tables": tables}, file, ensure_ascii=False)
    print("Exported records:", {table: len(rows) for table, rows in tables.items()})

if __name__ == "__main__":
    main()
