# Migration Rules

## Purpose

This file defines safe migration practices for this project.

## Rules

1. Migrations should primarily change schema, not wipe business data.
2. Avoid destructive statements in schema migrations:
    - DELETE FROM orders
    - DELETE FROM cards
    - DROP COLUMN without migration plan
3. Seed data should be separate from destructive reset logic where possible.
4. If a migration changes pricing columns or critical business fields, preserve old data safely.
5. Review every migration before applying on non-empty databases.
6. Down migrations may be destructive, but up migrations should be safe by default.
7. If a migration is dev-only or reset-style, mark that clearly.

## Current Warning

This project has already had migration safety concerns in review.
Do not apply migrations blindly on production/staging-like environments.