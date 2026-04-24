# Testing Guide

## Purpose

This file documents the minimum testing expected before approving changes.

## Basic Build Checks

Run:

```bash
go mod tidy
go run main.go
```

## Environment Checks

Verify:
- database config is loaded correctly
- PostgreSQL is reachable
- required environment variables are set
- email configuration is present if email behavior is being tested

## Manual Route Checks

Minimum route checks:
- homepage
- wedding card collection
- bid box collection
- card detail page
- product detail page
- checkout page
- order confirmation page
- admin products page
- admin orders page

## Core Flow Checks

### Wedding Card Flow
- browse wedding card collection
- open a card or linked product
- proceed to checkout
- complete personalization flow if applicable
- submit order
- verify confirmation page

### Bid Box Flow
- browse bid box collection
- open bid box product
- verify correct back link and category behavior
- verify ordering path works correctly

### Admin Product Flow
- create product
- edit product
- delete or disable product
- verify product image behavior
- verify category behavior
- verify linked card behavior for card products

### Admin Order Flow
- view order list
- open order detail
- update order status
- verify status changes are saved correctly

## Security Checks

Verify that tampering with client-side hidden values does not affect:
- price
- total
- currency
- canonical card name
- any trusted server-side order values

## Search and UI Checks

Verify:
- search works correctly
- buttons actually route correctly
- back-to-collection respects category
- icons and interactive elements are usable
- no broken links or dead-end flows exist

## Logging Checks

Verify that customer PII is not unnecessarily dumped into logs, including:
- full address
- phone number
- email
- other sensitive customer fields

## Email Checks

If email is enabled, verify:
- order created email
- order status update email
- admin notification email

## Database Checks

Verify:
- migrations apply successfully
- required tables exist
- expected seed data exists if the app depends on it
- no destructive migration behavior is applied unintentionally

## Before Merge

At minimum, confirm:
- app boots successfully
- database is reachable
- homepage works
- collection pages work
- card/product detail pages work
- checkout flow works
- order flow works
- admin critical actions work
- no obvious security regression exists