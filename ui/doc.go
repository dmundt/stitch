// Package ui provides constructor-based, semantic HTML components for Stitch.
//
// Usage contract:
//  - Build UI through exported constructors and compose using Component values.
//  - Add methods append children in the provided order.
//  - Several constructors shallow-copy component slices, while others retain
//    caller-owned slices. Check constructor docs when mutating shared inputs.
//  - Component HTML output is composed as trusted fragment HTML in nested
//    layouts, so custom Component implementations should emit safe markup.
//
// Choosing layout primitives:
//  - Use Row and Column when working with provider grid span classes.
//  - Use Grid and GridItem for generic item-grid layouts.
//  - Use Section and Article for semantic grouping and titled content regions.
//  - Use Split, SidebarLayout, and AppShell for two-region and shell layouts.
//
// Interaction guidance:
//  - Prefer one request method per Interaction (Get, Post, Put, or Delete).
//  - Keep Target and Select stable and app-controlled to avoid brittle updates.
package ui
