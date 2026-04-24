# Architecture Overview

## Overview

Write&InviteCo is a server-rendered Go application for wedding stationery and related event products.

The system currently mixes:
- storefront product presentation
- card-based checkout logic
- admin product/order management
- email notification flows

## Core Layers

### 1. Public Storefront
Responsible for:
- homepage
- collections
- product/category browsing
- card detail pages
- search
- checkout-facing pages

### 2. Product Layer
Products control storefront presentation:
- name
- category
- description
- image URL / media
- active/inactive state
- admin management

### 3. Card Layer
Cards control order/checkout behavior for wedding cards:
- price
- quantity rules
- foil options
- inserts
- canonical checkout data

### 4. Order Layer
Responsible for:
- customer capture
- order creation
- checkout submission
- order confirmation
- order status updates
- email triggers

### 5. Admin Layer
Responsible for:
- product CRUD
- product visibility
- card linking
- order listing
- order detail review
- status updates

## Important Relationship: Product vs Card

For wedding cards:

- Product = storefront appearance
- Card = checkout/pricing logic

This means a product may link to a card, and the actual trusted checkout information should come from the linked card/business logic.

## Category Behavior

Expected categories include at least:
- wedding cards
- bid boxes

Additional categories may exist but should be treated carefully unless fully wired.

## Bid Box Rule

Bid boxes should usually be standalone products.
They should not automatically inherit the wedding card checkout rules unless explicitly designed to.

## Image Strategy

Current / intended behavior:
- storefront can use product-specific images
- card flow may use card-linked data
- image handling must stay consistent and not silently switch between unrelated assets

Cloudinary may be used for hosted images.
If manual Cloudinary URLs are used, ensure they are stored consistently and rendered from the correct field.

## Database

PostgreSQL stores:
- cards
- card_images
- customers
- orders
- products
- related product media / admin data as implemented

Migrations live in `db/migrations/`.

## Routing Philosophy

Routes should remain predictable and category-aware.
Back links and collection links should not be hardcoded incorrectly.

## Key Architecture Rule

Do not duplicate pricing/checkout logic across:
- product pages
- checkout templates
- handlers

Server-side business logic should remain centralized and trusted.