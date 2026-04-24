# Checkout Flow

## Purpose

This document defines the trusted checkout behavior for the project.

## Trusted Principle

The client/browser must not be trusted for business-critical order values.

The server must derive canonical order values from the database.

## Current Intended Flow

For wedding cards:

Collection → Card/Product Detail → Checkout → Personalization → Review → Order Confirmation

For bid boxes:

Collection → Product Detail → Checkout / Order flow

## Server-Trusted Fields

The backend should trust:
- card_id
- product_id only where appropriate for lookup
- quantity (after validation)
- customer-entered contact/personalization fields

The backend should NOT trust:
- posted price
- posted total
- posted currency
- posted card_name
- any display-only hidden field that can be tampered with

## Required Server Behavior

On order submission:
1. validate the incoming identifiers
2. fetch canonical product/card info from DB
3. calculate the real price server-side
4. calculate totals server-side
5. create the order using trusted values only

## Checkout Page Rules

The checkout page itself should preferably be rendered from trusted DB data.
It should not display critical values sourced only from raw query params or hidden inputs without backend verification.

## Personalization

Personalization fields may be submitted by the user, but they are content fields, not pricing authority.

## Order Confirmation

Order confirmation must reflect the trusted server-created order, not raw client-submitted display values.

## Security Notes

Any field visible in the browser can be edited by the user.
Therefore:
- hidden inputs are not trusted
- displayed totals are not trusted
- server-side recalculation is mandatory

## Testing Expectations

Before considering checkout safe:
- tampering with frontend price must not change created order price
- tampering with currency must not change trusted order currency
- tampering with card_name must not override DB truth