# Multi-Level Policy Sources Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement PRD-defined HIGH/MEDIUM/LOW three-tier policy source system with RSS crawler, manual import, and seed data.

**Architecture:** New `rss_crawler.go` and `manual_crawler.go` implement existing `Source` interface. Manager switch-case extended. Seed migration adds MEDIUM (RSS media) and LOW (manual) sources.

**Tech Stack:** Go 1.24/1.25, encoding/xml (stdlib), existing crawler Source interface

---

### Task 1: RSS Crawler Implementation

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/crawler/rss_crawler.go`
- Create: `nsi-platform/services/policy-crawler/internal/crawler/rss_crawler_test.go`

- [ ] **Step 1: Write failing tests for RSS parsing**

Create `rss_crawler_test.go` with test fixtures and tests for RSS 2.0 and Atom parsing.

- [ ] **Step 2: Implement RSS crawler**

Create `rss_crawler.go` implementing `Source` interface with RSS/Atom XML parsing.

- [ ] **Step 3: Run tests, verify pass**

- [ ] **Step 4: Commit**

---

### Task 2: Manual Crawler Implementation

**Files:**
- Create: `nsi-platform/services/policy-crawler/internal/crawler/manual_crawler.go`

- [ ] **Step 1: Implement manual crawler (empty Fetch)**

- [ ] **Step 2: Verify build**

- [ ] **Step 3: Commit**

---

### Task 3: Manager Registration + Admin Import Handler

**Files:**
- Modify: `nsi-platform/services/policy-crawler/internal/crawler/manager.go`
- Modify: `nsi-platform/services/policy-crawler/internal/admin/admin.go`
- Modify: `nsi-platform/services/policy-crawler/cmd/main.go`

- [ ] **Step 1: Add rss/manual cases to manager.go Init()**

- [ ] **Step 2: Add admin source import handler**

- [ ] **Step 3: Wire in main.go**

- [ ] **Step 4: Run tests**

- [ ] **Step 5: Commit**

---

### Task 4: Seed Data Migration

**Files:**
- Create: `nsi-platform/services/policy-crawler/migrations/013_multi_level_sources.sql`

- [ ] **Step 1: Create migration with MEDIUM and LOW sources**

- [ ] **Step 2: Commit**

---

### Task 5: Final Verification

- [ ] **Step 1: go build ./... all services**

- [ ] **Step 2: go test ./... all services**

- [ ] **Step 3: go vet ./...**
