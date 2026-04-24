This file defines the working rules for AI coding agents (Codex, Hermes, Bob, etc.) in this repository.

## Primary Rule

Make small, safe, explainable changes.

Do not make broad architectural changes unless explicitly asked.

## Project Truths

- Products are the storefront/UI layer.
- Cards are the pricing/checkout layer for wedding cards.
- Bid boxes are standalone products unless explicitly linked into a different flow.
- Existing working checkout flow should be preserved unless a change is explicitly requested.

## Must Preserve

Do not break:
- collection pages
- card detail pages
- product to card handoff
- checkout flow
- personalization flow
- review page
- order confirmation
- admin order management
- admin product management

## Never Do Without Explicit Approval

- destructive migrations
- deleting real data
- dropping columns/tables carelessly
- rewriting checkout architecture
- replacing working routing behavior
- major refactors across many files
- changing security/business rules silently
- changing product/card mapping behavior without confirming intent

## Security Rules

Never trust client-submitted:
- price
- total
- currency
- card_name
- display-only product data for order integrity

Server must derive:
- canonical card/product info
- price
- total
- trusted order values

Do not log sensitive customer PII in production logs:
- full address
- phone
- email
- full customer payloads

## Migration Rules

- Review every migration before applying
- Separate schema changes from seed/reset logic when possible
- Avoid `DELETE FROM ...` in schema migrations
- Keep migrations safe for non-empty databases unless explicitly marked dev-only

## Change Rules

When making a change:
1. explain what changed
2. explain why it changed
3. list files touched
4. state what was tested
5. state what is still risky or incomplete

## Testing Expectations

For any non-trivial change, verify:
- app builds
- key routes still work
- no obvious regressions in checkout/admin flow

## Preferred Style

- small diffs
- explicit reasoning
- keep logic readable
- avoid duplicate business logic
- avoid moving too much at once

## If Unsure

Do not guess. Inspect the actual repo and report findings before changing core behavior.