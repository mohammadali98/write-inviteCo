# Write&InviteCo

Write&InviteCo is a Go-based wedding stationery storefront and admin system for:
- wedding cards
- bid boxes
- related wedding products

The project includes:
- public storefront
- product/category browsing
- card detail pages
- checkout flow
- personalization flow
- order confirmation
- admin product management
- admin order management
- email notifications

## Tech Stack

- Go
- Gin
- PostgreSQL
- Server-rendered HTML templates
- Static assets in `static/`
- Resend for email
- Cloudinary for hosted product images (currently may be manual URL-based)

## Project Purpose

This app is designed to:
- display wedding card and product catalog
- allow customers to browse and place orders
- let admin manage products and orders
- keep checkout/order flow reliable and server-trusted

## High-Level Architecture

There are two main concepts in this project:

### Products
Products are the storefront/UI layer.
They control:
- display name
- description
- category
- image URL / product media
- admin visibility

### Cards
Cards are the business logic layer for checkout-enabled stationery.
They control:
- pricing
- foil options
- inserts
- quantity rules
- checkout behavior

### Important Rule
For wedding cards:
- product is the display layer
- card is the pricing/checkout layer

For bid boxes:
- they are standalone products
- they should not use card checkout logic unless explicitly designed to

## Main User Flows

### Wedding Card Flow
Collection → Product/Card → Checkout → Personalization → Review → Order Confirmation

### Bid Box Flow
Collection → Product → Checkout / Order flow

## Admin Features

- add/edit/delete products
- activate/deactivate products
- link products to cards where needed
- manage orders
- update order status
- review customer info and order details

## Running the App

> Update this section if the exact setup differs in your repo.

### Requirements
- Go installed
- PostgreSQL running
- database created
- env variables configured

### Basic Run
```bash
go mod tidy
go run main.go