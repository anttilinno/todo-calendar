# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.8: Shipped (see MILESTONES.md)
- v1.9 Fuzzy Date Todos: Shipped 2026-02-12 (see MILESTONES.md)
- v2.0 Settings UX: Shipped 2026-02-12 (see MILESTONES.md)
- v2.1 Priorities: Shipped 2026-02-13 (see MILESTONES.md)
- v2.2 Google Calendar Events: Shipped 2026-02-14 (see MILESTONES.md)
- v2.3 Polybar Status: Shipped 2026-02-23 (see MILESTONES.md)
- v2.4 Auto-Update: In Progress

## Phases

<details>
<summary>v1.9 Fuzzy Date Todos (Phases 27-29) — SHIPPED 2026-02-12</summary>

- [x] Phase 27: Date Precision & Input (2/2 plans) — completed 2026-02-12
- [x] Phase 28: Display & Indicators (2/2 plans) — completed 2026-02-12
- [x] Phase 29: Settings & View Filtering (1/1 plan) — completed 2026-02-12

</details>

<details>
<summary>v2.0 Settings UX (Phase 30) — SHIPPED 2026-02-12</summary>

- [x] Phase 30: Save-on-Close Settings — completed 2026-02-12

</details>

<details>
<summary>v2.1 Priorities (Phases 31-32) — SHIPPED 2026-02-13</summary>

- [x] Phase 31: Priority Data Layer (1/1 plan) — completed 2026-02-13
- [x] Phase 32: Priority UI + Theme (2/2 plans) — completed 2026-02-13

</details>

<details>
<summary>v2.2 Google Calendar Events (Phases 33-35) — SHIPPED 2026-02-14</summary>

- [x] Phase 33: OAuth & Offline Guard (2/2 plans) — completed 2026-02-14
- [x] Phase 34: Event Fetching & Async Integration (2/2 plans) — completed 2026-02-14
- [x] Phase 35: Event Display & Grid (3/3 plans) — completed 2026-02-14

</details>

<details>
<summary>v2.3 Polybar Status (Phases 36-37) — SHIPPED 2026-02-23</summary>

- [x] Phase 36: Status Subcommand (2/2 plans) — completed 2026-02-23
- [x] Phase 37: TUI State File Integration (1/1 plan) — completed 2026-02-23

</details>

### v2.4 Auto-Update (Phases 38-40)

#### Phase 38: Update Engine
**Goal:** Build the core update engine — check GitHub Releases API for latest version, download binary + SHA256 checksum, verify integrity, and atomically replace the binary at configured path
**Requirements:** R1, R2, R3, NR1, NR2, NR3
**Plans:** 2 plans
Plans:
- [ ] 38-01-PLAN.md — GitHub API client, downloader, and SHA256 checksum verifier
- [ ] 38-02-PLAN.md — Atomic binary replacement with permission preservation
**Success criteria:**
1. CheckForUpdate returns latest version tag from GitHub Releases API
2. DownloadRelease fetches binary and checksum files with 5s timeout
3. SHA256 verification rejects tampered binaries
4. Atomic binary replacement preserves original file permissions
5. Network errors and API failures are logged to stderr without blocking app launch

#### Phase 39: Settings Integration
**Goal:** Add auto_update toggle and binary_path display to Config and settings overlay
**Requirements:** R4, R5
**Success criteria:**
1. Config struct has AutoUpdate bool and BinaryPath string fields with TOML tags
2. DefaultConfig sets AutoUpdate=false and BinaryPath=""
3. Settings overlay shows Auto Update cycling toggle (Enabled/Disabled)
4. Settings overlay shows Binary Path as read-only display with resolved effective path

#### Phase 40: Launch Wiring
**Goal:** Wire update engine into main() to run on app launch before TUI start
**Success criteria:**
1. main() calls update check after config load when AutoUpdate is true
2. Update is skipped when version == "dev"
3. BinaryPath resolves from os.Executable() when config value is empty
4. App launches normally regardless of update outcome

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 27. Date Precision & Input | v1.9 | 2/2 | Complete | 2026-02-12 |
| 28. Display & Indicators | v1.9 | 2/2 | Complete | 2026-02-12 |
| 29. Settings & View Filtering | v1.9 | 1/1 | Complete | 2026-02-12 |
| 30. Save-on-Close Settings | v2.0 | 1/1 | Complete | 2026-02-12 |
| 31. Priority Data Layer | v2.1 | 1/1 | Complete | 2026-02-13 |
| 32. Priority UI + Theme | v2.1 | 2/2 | Complete | 2026-02-13 |
| 33. OAuth & Offline Guard | v2.2 | 2/2 | Complete | 2026-02-14 |
| 34. Event Fetching & Async | v2.2 | 2/2 | Complete | 2026-02-14 |
| 35. Event Display & Grid | v2.2 | 3/3 | Complete | 2026-02-14 |
| 36. Status Subcommand | v2.3 | 2/2 | Complete | 2026-02-23 |
| 37. TUI State File Integration | v2.3 | 1/1 | Complete | 2026-02-23 |
| 38. Update Engine | v2.4 | 0/2 | Pending | — |
| 39. Settings Integration | v2.4 | 0/? | Pending | — |
| 40. Launch Wiring | v2.4 | 0/? | Pending | — |
