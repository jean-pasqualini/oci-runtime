# Test Plan: <feature-technical-name>

## 1. Identification
- Plan ID:           TP-<feature-technical-name>
- Feature:           .agent/requests/feature/<feature-technical-name>.md
- Version / commit:  <git sha or branch at time of writing>
- Author:            <name>
- Date created:      YYYY-MM-DD
- Last executed:     YYYY-MM-DD (or "never")

## 2. Objective
One short paragraph: what this plan verifies end-to-end, and what claim its passing demonstrates.

## 3. Scope
- In scope:    bullet list of behaviours covered
- Out of scope: bullet list of behaviours intentionally not covered (link to a separate plan if one exists)

## 4. References
- Feature request:   .agent/requests/feature/<name>.md
- Related code:      key packages / files exercised
- External specs:    OCI spec sections, RFCs, design docs, tickets

## 5. Test Environment
- Build / launch:    e.g. `go run -tags medium ./cmd/oci-runtime <subcommand>`
- Runtime / OS:      e.g. Linux kernel >= 5.10, container with /app workspace mount
- Required fixtures: e.g. `/app/bundle` is a valid OCI bundle; `/tmp/state` writable and empty
- Required tooling:  e.g. `mount`, `lsns`, `ip netns`

## 6. Test Data
Single source of truth for inputs used across cases — keeps each case short.
- containerNames:  c1, c2, c3, ghost
- root:            /tmp/state
- bundle:          /app/bundle

## 7. Entry Criteria
Must all be true before execution begins. Stop the run if any fails.
- Build is green (`check` subcommand exits 0)
- No residual state under `<root>` for the names in Test Data
- No leftover mounts / namespaces from prior runs

## 8. Exit Criteria
Must all be true to call the plan complete.
- All P0 and P1 cases: Pass
- No open Critical or High defects against the feature
- Execution Log filled for every case (Pass / Fail / Blocked / Skipped, never Not Run)

## 9. Test Cases

Each case is independent and self-contained. Order in the document = recommended execution order, but cases must not depend on earlier cases' side effects unless declared as a Precondition.

---

### TC-001 <short imperative title — "delete removes state dir after run">
- Priority:        P0 | P1 | P2 | P3
- Type:            Smoke | Functional | Negative | Edge | Regression | Integration | Performance | Security
- Precondition:    explicit state required (env vars, prior cases, fixtures). "None" if independent.
- Steps:
  1. <exact command or action>
  2. <exact command or action>
  3. ...
- Expected result: observable outcome — exit code, file presence/absence, error message substring, side effects
- Actual result:   (filled at execution — verbatim observation, including any stderr)

### TC-002 ...
(repeat the block)

---

## 10. Risks & Known Gaps
Things that may invalidate results or that the plan knowingly cannot cover today.
- Risk:  short description -> mitigation
- Gap:   scenario X is unreachable until feature task N lands (link feature memory). Distinguish:
  - **Test gap**: verification path can't run (tooling, env, permissions)
  - **Feature gap**: code path doesn't exist yet (cite the feature task that will close it)

## 11. Execution Log

- TC-XXXX: Summary title of the test case
  Status: NOT_RUN/FAILED_PASSED
  Iteration: how much iteration were needed to pass
  Notes: Additional notes or reserve.
- TC-XXX: Summary title of the test case
  [...]