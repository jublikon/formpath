# Formpath

## Project Purpose

Formpath is a platform for collecting data from multiple sources, generating analyses, and exposing those analyses through interfaces that can be consumed by applications and agents.

The product includes an agentic application layer that:
- shows user data
- lets users interact with their data
- supports generation of outputs such as training plans, custom visualizations, and other derived artifacts
- enables intelligent workflows on top of the underlying data platform

Primary goals:
- ingest and unify data from multiple providers and user inputs
- generate useful analyses and derived outputs
- expose platform capabilities through stable interfaces
- support agentic product experiences on top of the platform
- ship a product that works end to end locally first, and later on AWS

This product must optimize for:
1. usefulness
2. correctness
3. maintainability
4. privacy
5. product iteration speed

When trade-offs exist, prefer correctness and working end-to-end behavior over cleverness.

---

## Product Principles

- The platform should make data from different sources usable in one coherent system.
- The product should support interaction, not just passive display.
- Derived outputs such as plans, summaries, analyses, and visualizations should be grounded in available data.
- Prefer flexible platform capabilities over one-off hardcoded features.
- Favor stable interfaces and reusable building blocks.
- Prefer a thin working product slice over broad unfinished infrastructure.

---

## Engineering Principles

- Prefer simple, explicit solutions over abstraction-heavy designs.
- Prefer small, reviewable diffs over broad refactors.
- Preserve clean boundaries between ingestion, canonical models, analysis, orchestration, and delivery layers.
- Keep business logic out of transport/framework layers.
- Avoid hidden coupling between services.
- Do not introduce new infrastructure, major dependencies, or schema changes unless clearly justified in the plan.
- Build features as vertical slices, not as disconnected horizontal layers.
- Prefer end-to-end functionality over partial framework scaffolding.
- A feature is more valuable when it works through the full product flow than when multiple internal layers are half-finished.

---

## Development Strategy

This project is local-first during the early stage.

That means:
- everything important should run locally on a developer machine
- local development is the default path
- AWS deployment comes later, once the product works end to end
- early architecture should not depend on cloud-only assumptions unless clearly necessary

Design with a future AWS deployment in mind, but optimize first for:
1. local development speed
2. debuggability
3. end-to-end product learning
4. low operational complexity

Prefer architectures that can evolve cleanly from:
- local single-machine development
- to containerized local environments
- to production deployment on AWS

Do not introduce AWS-specific complexity too early if a simpler local-first solution is sufficient.

---

## Vertical Feature Development

Build new functionality vertically.

A vertical slice should, where relevant, include:
- domain model changes
- ingestion changes
- persistence changes
- analysis logic changes
- orchestration / agent logic changes
- API changes
- UI / presentation changes
- tests
- documentation

Do not implement features as isolated internal pieces that are not usable end to end.

When adding a new feature, prefer:
- one small working slice through the system

over:
- partially building multiple architectural layers without user-visible value

For feature work, optimize for:
- end-to-end usability
- thin but complete implementation
- fast feedback from real product behavior

A thin end-to-end slice is preferred over a broad but incomplete platform build-out.

---

## Expected Working Style

For non-trivial work, always:
1. explore the relevant code and docs first
2. propose a short plan
3. implement in small milestones
4. verify after each milestone
5. summarize what changed, risks, and follow-up work

For larger features, prefer:
- spec -> plan -> vertical slice -> verify -> next slice -> verify -> final summary

Do not jump into large edits without first understanding the existing structure.

If a task is ambiguous, make the smallest safe assumption and state it briefly.

When implementing functionality, prefer delivering one complete end-to-end path before adding breadth.

---

## Definition of Done

A task is not done until:
- the implementation matches the requested behavior
- the new functionality works end to end where relevant
- relevant tests pass
- edge cases are covered or explicitly documented
- user-visible behavior is reachable through the actual product flow
- logs / errors are actionable
- docs are updated when behavior or architecture changed

For platform features:
- the data flow should be testable from input to exposed output

For agentic features:
- the user interaction path should be testable
- the output should be reproducible enough for debugging
- assumptions and failure paths should be visible in code or logs

For feature work:
- the feature must be reachable and testable through the actual product flow
- not only through isolated units or mocked internal code paths

---

## Verification Rules

Claude must always give itself a way to verify its work.

When making changes:
- run the smallest relevant tests first
- then run broader checks if the touched surface area is large
- prefer targeted validation over expensive full-suite runs during iteration
- before finalizing, run the most relevant lint / typecheck / test commands available

If the change is a feature slice:
- verify the end-to-end happy path locally
- verify the new path with realistic local data or fixtures
- verify that the user-facing output is actually reachable

If no tests exist for changed logic:
- add tests if the area is stable enough
- otherwise provide a short manual verification procedure

For UI or API output changes:
- verify actual output shape and examples, not just implementation assumptions

For agentic behavior:
- verify tool inputs and outputs explicitly
- verify fallback behavior
- verify error handling and partial-failure behavior

---

## Sensitive Data and Privacy

This project handles health-, behavior-, and user-adjacent data.

Always:
- minimize exposure of personal data in logs, fixtures, screenshots, and examples
- avoid storing raw secrets or tokens in code, docs, or test fixtures
- redact identifiers in examples unless synthetic
- use synthetic or anonymized sample data by default
- prefer least-privilege access patterns
- treat wearable, recovery, sleep, HR, HRV, location, subjective input, and similar user data as sensitive

Never hardcode credentials or commit local secrets.

---

## Safety-Critical / Medical Features

This platform may include medical or diagnosis-related functionality.

Treat all such features as safety-critical.

For medical, diagnostic, or health-claim-related functionality:
- prefer explicit logic over opaque behavior where possible
- make data dependencies clear in code
- log and validate important assumptions
- do not silently degrade into unsupported outputs
- handle uncertainty, missing data, and conflicting data explicitly
- add stronger tests than for ordinary product features
- keep reasoning paths and decision boundaries inspectable in code
- document important thresholds, rules, and assumptions
- flag areas that require expert, legal, regulatory, or clinical review

Do not present a feature as reliable, validated, or production-ready unless the implementation and validation justify that claim.

---

## Domain Model Guidance

Key domain idea:
raw provider data -> normalized canonical data -> derived analyses/features -> agent/application outputs -> external interfaces

Keep these concerns separate:
- provider adapters
- canonical domain models
- ingestion pipelines
- feature / signal generation
- analysis logic
- planning / generation logic
- agent orchestration
- explanation / presentation
- delivery / API layer

Examples of likely domains:
- user
- provider connection
- source record
- canonical event / metric
- analysis result
- plan
- visualization
- report
- prompt / agent context
- generated artifact

Do not let provider-specific fields leak deeply into domain logic when a canonical field can be used.

---

## Data and Time Rules

- Store timestamps in UTC internally.
- Convert to user-local time only at presentation boundaries.
- Be explicit about units: seconds, minutes, meters, kilometers, bpm, ms, watts, kcal.
- Do not mix derived and raw values without naming that clearly.
- Use consistent naming for time windows such as 7d, 14d, 28d.
- Be careful with missingness: "unknown", "not synced", and "zero" are different states.
- Ingestion flows must be idempotent where possible.
- Deduplication rules must be explicit and testable.
- Canonicalization rules must be deterministic where possible.

---

## Analysis and Generation Rules

Analysis and generation quality matter more than novelty.

Prefer:
- explicit data transformations
- stable generation pipelines
- bounded behavior
- testable rules and policies
- reproducible outputs where practical
- deterministic fallbacks where possible

Avoid:
- magic constants without explanation
- silent fallback behavior
- hidden rule cascades
- outputs that imply stronger certainty than the implementation supports
- deeply coupled provider-specific analysis logic

Whenever analysis or generation logic changes, document:
- what inputs changed
- what behavior changed
- why this is better
- what surfaces or user paths are affected

---

## Agentic Application Rules

The agentic layer should be built on top of stable platform interfaces.

Prefer:
- clear tool boundaries
- explicit contracts for tool inputs/outputs
- reusable agent actions over one-off prompt logic
- deterministic non-agent fallbacks where sensible
- inspectable prompts, schemas, and orchestration paths

Avoid:
- hiding important business logic only inside prompts
- tightly coupling agent behavior to unstable internal implementation details
- mixing data access, business rules, and presentation in one layer

When adding agentic functionality, define:
- user goal
- available tools/actions
- required data inputs
- expected output shape
- fallback behavior
- verification strategy

---

## API and Service Design

- Keep APIs explicit and versionable.
- Use clear request/response models.
- Validate inputs at boundaries.
- Return structured errors with actionable messages.
- Avoid leaking internal implementation details in public responses.
- Prefer backward-compatible API changes unless explicitly doing a versioned break.

If introducing a new endpoint, tool, or event:
- define purpose
- define consumers
- define schema
- define validation
- define observability

Design APIs and internal services so they can run locally first, and later be deployed on AWS with minimal rewriting.

---

## Observability

Changes should be debuggable in production-like environments.

Prefer:
- structured logs
- stable event names
- clear error messages
- metrics around ingestion success/failure, latency, and generation behavior
- correlation identifiers where appropriate

Do not log sensitive raw payloads unless explicitly required and properly redacted.

Local development should make debugging easy. Prefer setups that expose logs, errors, and state clearly on a single machine.

---

## Testing Guidance

Prioritize tests for:
- provider normalization
- deduplication / idempotency
- canonical model conversion
- analysis logic behavior
- date/time handling
- serialization boundaries
- contract stability
- end-to-end feature slice behavior
- agent tool contracts
- output generation paths
- error and missing-data behavior

Good tests in this project are:
- small
- deterministic
- readable
- scenario-based

Prefer table-driven tests for rules/policies where useful.

For important features, include at least:
- unit tests for critical logic
- one integration or end-to-end test for the main user path

For safety-critical medical features, prefer stronger integration and scenario testing.

---

## Repo Conventions

Assume a monorepo unless the current repository structure clearly says otherwise.

Prefer organizing code by feature/domain where practical, not only by technical layer.

Preferred separation:
- features
- platform / core domain
- shared libraries
- infra
- docs

Within a feature slice, it is acceptable to keep closely related API, service, domain, and persistence code near each other if that improves clarity and delivery speed.

Keep files close to the domain they belong to.
Do not create generic "utils" unless the helper is truly cross-cutting and well named.

---

## Local-First Infrastructure Conventions

During the product-building phase, the system should run locally with minimal setup friction.

Prefer:
- local processes or Docker Compose for multi-service development
- local databases where possible
- local file storage or emulated object storage where useful
- environment variables documented in one place
- reproducible developer setup
- simple bootstrapping commands

Avoid early dependence on:
- cloud-only queues
- cloud-only networking assumptions
- distributed infrastructure that is unnecessary in the local phase
- premature microservice decomposition

Before introducing infrastructure, ask:
- can this be developed and tested locally?
- does this improve end-to-end product delivery now?
- is this needed before first working product validation?

If the answer is no, prefer the simpler local-first option.

---

## Future AWS Migration Guidance

We plan to deploy to AWS later, after a working product exists.

Therefore:
- keep boundaries clean so components can later be extracted or deployed separately
- prefer infrastructure choices that have a clear AWS path later
- avoid deep coupling between business logic and local-only tooling
- keep deployment assumptions explicit
- document what will likely change during AWS migration

Design for future cloud deployment, but do not optimize prematurely for production-scale infra before the product proves itself locally.

---

## Go Conventions

For Go code:
- prefer small packages with clear ownership
- keep interfaces narrow
- define interfaces where they are consumed, not where they are implemented
- return explicit errors; do not swallow them
- wrap errors with useful context
- avoid premature abstraction
- use table-driven tests where appropriate

---

## Python Conventions

For Python code:
- prefer typed functions where practical
- keep data and ML-related logic explicit and testable
- avoid notebook-style patterns in production code
- separate experimentation code from production code
- prefer pure transformation functions over stateful pipelines when possible

---

## Infrastructure Conventions

Assume AWS-first as the future deployment target, but local-first as the current development mode.

- infrastructure changes must be intentional and reviewable
- do not introduce unnecessary managed services early
- prefer simple deployable building blocks
- document environment assumptions
- keep local development possible without full cloud deployment
- do not require AWS for basic feature development unless explicitly necessary

---

## Documentation Rules

When changing behavior, update the relevant docs.
When introducing new concepts, define them once in the right place.

Prefer concise docs that answer:
- what is this?
- why does it exist?
- how is it used?
- how is it verified?

Do not write long prose when a short decision note or example is enough.

Document local setup and end-to-end verification paths clearly.

---

## Commands

If the repository already contains canonical commands, use those instead of inventing new ones.

When available, prefer project-defined commands such as:
- `make dev`
- `make test`
- `make lint`
- `make typecheck`
- `make ci`
- `docker compose up`
- `just dev`
- `task dev`

If there is no top-level task runner yet, prefer targeted language-native commands:
- Go: `go test ./...`
- Go lint: `golangci-lint run`
- Python tests: `pytest`
- Python lint: `ruff check .`
- Python format check: `ruff format --check .`
- Python typing: `mypy .`

Do not run expensive full-repo checks repeatedly if only a small surface changed, unless the task is close to completion.

---

## What To Put In Skills Instead

Do not bloat this file with long procedures.

Move these into `.claude/skills/` when they become repetitive:
- implementing a new provider connector
- adding a new analysis module
- adding a new agent tool
- reviewing a diff before commit
- investigating a production issue
- planning a schema migration
- shipping a vertical feature slice end to end
- preparing a local feature for AWS migration

---

## Communication Style

When reporting back:
- be concrete
- mention files changed
- mention verification performed
- mention assumptions
- mention open risks

Do not claim certainty you do not have.
Do not say something is production-ready unless tests, validation, and operational concerns justify it.

---

## Project-Specific Default

This is a data platform plus agentic application.
Act like a careful senior engineer building a local-first, end-to-end product with strong platform boundaries and complete vertical slices.