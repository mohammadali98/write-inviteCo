# Admin Panel Guide

## Purpose

The admin panel is intended to let the business manage:
- products
- product visibility
- product-to-card links
- orders
- order statuses

## Product Types

### Card Products
Use when the item is part of the wedding card flow.

Expected behavior:
- category = card
- linked card is selected
- storefront display comes from product data
- checkout logic comes from linked card behavior

### Bid Box Products
Use when the item is a standalone product.

Expected behavior:
- category = bid_box
- linked card should usually be empty
- do not force card checkout logic unless explicitly intended

## Product Fields

Typical fields include:
- name
- category
- price (display/storefront use depending on current implementation)
- description
- image URL / uploaded media
- linked card
- active/inactive

## Linked Card Meaning

Linked Card tells the app which checkout/pricing logic to use for a wedding-card product.

It should match the correct product/card design and intended checkout behavior.

## Product Management Rules

- keep names clean and consistent
- match image/design to correct linked card
- do not assign random linked cards
- inactive products should not appear publicly
- delete only when really necessary; prefer disable where possible

## Order Management

Admin can:
- view orders
- inspect order details
- update order status

Expected statuses:
- pending
- confirmed
- completed
- cancelled

## Email Behavior

Order-related emails should ideally be triggered from server-side order events:
- order created
- confirmed
- completed
- cancelled

## Image Management

Current image strategy may involve:
- static assets
- Cloudinary URLs
- manual admin image URLs
- product image fields

Be consistent and avoid mixing image sources without intention.

## Safety Rules

- do not trust admin UI alone; backend validation still matters
- category behavior should be explicit
- linked card behavior should be validated server-side