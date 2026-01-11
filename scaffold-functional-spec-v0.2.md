# Modular Source‑Agnostic Project Bootstrapper --- Functional Specification (v0.2)

**Working name:** Scaffold (placeholder)\
**Audience:** Product & engineering (pre‑implementation)\
**Primary goal:** Provide a cross‑platform, zero‑prerequisite CLI that
composes modular templates sourced from any location (git repositories,
local paths, or URLs) into a single project scaffold, with variable
substitution, conflict handling, and post‑generation actions.

**Tagline:** Bootstrap any software stack with sensible defaults.

------------------------------------------------------------------------

## 1. Problem Statement

Existing scaffolding tools typically fall into one of these categories:

-   Single‑template generators (e.g., cookiecutter)
-   Ecosystem‑specific scaffolds (e.g., framework CLIs)
-   Heavy platform solutions (e.g., internal developer portals)

None provide a simple, source‑agnostic, composable module system where
users can: - Choose a base template from any source - Layer optional
feature modules - Resolve conflicts deterministically - Apply variable
substitution once - Run post‑generation setup automatically

This project aims to fill that gap.

### 1.1 Why Sensible Defaults Matter

Projects bootstrapped with sensible defaults and best practices are:

-   **Easier for humans** to understand, maintain, and extend
-   **Better for AI‑assisted development** — AI coding tools perform
    significantly better when working with well‑structured, conventional
    codebases that follow established patterns
-   **More predictable** — consistent structure across projects reduces
    cognitive load and onboarding time
-   **Production‑ready faster** — security, testing, and deployment
    patterns built in from day one

Scaffold bakes in best practices not by being opinionated about
technology choices, but by ensuring that whatever stack you choose
follows that ecosystem's conventions and standards.

------------------------------------------------------------------------

## 2. Non‑Goals (Initial Versions)

-   Not a full project lifecycle manager
-   Not a dependency/package manager
-   Not a build system
-   Not an IDE plugin (CLI first)

------------------------------------------------------------------------

## 3. User Personas

### 3.1 Startup / Agency Engineer

-   Wants to spin up consistent projects fast
-   Uses private templates (git repos, local paths, or internal servers)
-   Wants Django/Next/Infra stacks with feature toggles

### 3.2 Open‑Source Maintainer

-   Publishes reusable modules
-   Wants users to compose their own stacks

### 3.3 Platform Team

-   Wants repeatable, policy‑compliant scaffolds
-   Might add registry/index later

------------------------------------------------------------------------

## 4. Core User Experience

### 4.1 Quick Start

``` bash
npx create-scaffold myapp
```

or

``` bash
scaffold init myapp
```

### 4.2 Interactive Flow

1.  Choose base template
2.  Select optional modules (multi‑select)
3.  Answer all prompts (merged from all modules)
4.  Review generation plan
5.  Generate project
6.  Run post‑generation steps

### 4.3 Non‑Interactive (CI / Scripts)

``` bash
# Using git repository
scaffold init myapp \
  --base git:https://gitlab.com/org/base-template#v1 \
  --add git:git@github.com:org/mod-postgres.git \
  --var project_name=myapp

# Using local path
scaffold init myapp \
  --base file:~/templates/base-template \
  --add file:./modules/mod-postgres

# Using URL (tarball/zip)
scaffold init myapp \
  --base https://example.com/templates/base.tar.gz \
  --add https://internal.company.com/modules/postgres.zip
```

------------------------------------------------------------------------

## 5. Template & Module Model

### 5.1 Source‑Agnostic Distribution

Templates and modules can be fetched from any of the following sources:

#### Git Repositories (any provider)

-   `git:https://github.com/org/repo` — HTTPS clone
-   `git:git@github.com:org/repo.git` — SSH clone
-   `git:https://gitlab.com/org/repo#v1.0` — with tag/branch
-   `git:https://bitbucket.org/org/repo//subdir#main` — subdirectory

Supports GitHub, GitLab, Bitbucket, Gitea, self‑hosted, or any git server.

#### Local File Paths

-   `file:./relative/path` — relative to current directory
-   `file:~/templates/my-template` — home directory expansion
-   `file:/absolute/path/to/template` — absolute path

Useful for development, air‑gapped environments, or local template libraries.

#### URLs (HTTP/HTTPS)

-   `https://example.com/template.tar.gz` — tarball
-   `https://example.com/template.zip` — zip archive
-   `https://internal.server/templates/base` — internal servers

Supports authentication via environment variables or config file.

#### Shorthand Aliases (Optional)

For convenience, common providers can use shorthand:

-   `github:org/repo` → `git:https://github.com/org/repo`
-   `gitlab:org/repo` → `git:https://gitlab.com/org/repo`
-   `bitbucket:org/repo` → `git:https://bitbucket.org/org/repo`

No central registry is required for usage.

### 5.2 Base Templates vs Modules

  Type            Purpose
  --------------- --------------------------------
  Base Template   Provides full project skeleton
  Module          Adds or modifies features

Both use the same internal format.

### 5.3 Module Structure

    module-root/
      module.yaml
      templates/
      patches/
      actions/

### 5.4 Manifest (module.yaml) Responsibilities

-   Metadata (name, version, description)
-   Prompt schema
-   Files to render
-   Patch instructions
-   Dependencies on other modules
-   Post‑generation commands

------------------------------------------------------------------------

## 6. Variable & Prompt System

### 6.1 Unified Prompt Pass

All selected modules contribute to a single merged prompt schema.

User is prompted once for all required variables.

### 6.2 Variable Scopes

-   Global: project_name, org, license, language versions
-   Module‑namespaced: postgres.db_name, celery.broker

### 6.3 Prompt Types

-   string
-   boolean
-   select
-   multiselect
-   number

Defaults and computed values supported.

------------------------------------------------------------------------

## 7. Generation Pipeline

### 7.1 Execution Order

1.  Resolve all modules and dependencies
2.  Fetch and unpack repositories
3.  Merge prompt schemas
4.  Collect user input
5.  Render base template files
6.  Apply modules in deterministic order:
    -   Render templates
    -   Apply patches
    -   Perform injections
7.  Run post‑generation actions
8.  Write lockfile

### 7.2 Deterministic Layering

Modules are applied in: 1. Dependency order 2. User‑selected order 3.
Alphabetical tie‑break

------------------------------------------------------------------------

## 8. File Operations

### 8.1 Template Rendering

Files in templates/ are copied to project root with variable
substitution.

### 8.2 Patch Application

Patches apply diffs to existing files.

Conflict strategies: - Fail (default) - Interactive resolve -
Last‑writer wins (configurable)

### 8.3 Injection Operations

Declarative edits such as: - Append to file - Insert after pattern -
Modify JSON/YAML keys

------------------------------------------------------------------------

## 9. Post‑Generation Actions

### 9.1 Supported Actions

-   Run shell commands
-   Initialize git repo
-   Install dependencies
-   Open docs/URLs

### 9.2 Safety Model

Actions require user confirmation unless `--yes` is provided.

------------------------------------------------------------------------

## 10. Lockfile & Reproducibility

### 10.1 scaffold.lock

Stores: - Base template ref - Module refs and versions - All resolved
variable values

### 10.2 Future Use

Enables: - Regeneration - Template updates - Drift detection

------------------------------------------------------------------------

## 11. Update & Regeneration (Future Phase)

Planned command:

``` bash
scaffold update
```

Capabilities: - Re‑fetch module versions - Re‑apply templates - Surface
conflicts - Allow selective upgrades

Not required for initial MVP, but format must support it.

------------------------------------------------------------------------

## 12. Distribution & Installation

### 12.1 Primary Path

-   Prebuilt static binaries:
    -   macOS (arm64, amd64)
    -   Windows
    -   Linux

### 12.2 Bootstrap Path

-   `npx create-scaffold`
-   Downloads correct binary
-   Executes locally

No language runtimes required.

------------------------------------------------------------------------

## 13. Extensibility

### 13.1 Module as Extension Mechanism

New behaviors are introduced via modules, not plugins.

### 13.2 Optional Future Plugin System

Could allow: - Custom actions - Custom prompt types - Custom conflict
resolvers

Out of scope for v1.

------------------------------------------------------------------------

## 14. Security Considerations

-   All remote code requires explicit source refs
-   Source verification via checksums (optional)
-   No silent execution of scripts
-   Clear display of all actions before run
-   Optional sandboxing of actions (future)
-   Authentication credentials never stored in lockfiles

------------------------------------------------------------------------

## 15. MVP Feature Set

Must have:
-   Git repository fetching (any provider)
-   Local file path support
-   Variable rendering
-   Modular layering
-   Patch application
-   Unified prompts
-   Lockfile generation

Nice to have:
-   URL/archive fetching
-   Interactive conflict resolution
-   Registry/index
-   Template updates

------------------------------------------------------------------------

## 16. Success Criteria

-   Can recreate cookiecutter‑style Django template
-   Can compose at least 3 independent modules cleanly
-   Zero runtime dependencies for end user
-   Works on macOS, Windows, Linux

------------------------------------------------------------------------

## 17. Roadmap (Indicative)

### Phase 1 --- Core Engine

-   Repo fetch
-   Manifest parsing
-   Template rendering

### Phase 2 --- Composition

-   Dependency resolution
-   Patch engine
-   Conflict handling

### Phase 3 --- UX Polish

-   Interactive TUI
-   Plan preview
-   Better errors

### Phase 4 --- Updates

-   scaffold update
-   Template version diffing

------------------------------------------------------------------------

## 18. Open Questions

-   Patch format: unified diff vs structured ops?
-   Registry or discovery mechanism?
-   Module trust/signing model?
-   Should modules be allowed to modify other modules?

------------------------------------------------------------------------

## Appendix: Design Principles

-   **Sensible defaults** — best practices baked in, not bolted on
-   **Source‑agnostic** — any git repo, local path, or URL
-   **Composition over monolith** — layer modules, don't fork templates
-   **Deterministic output** — same inputs always produce same outputs
-   **Human‑readable manifests** — YAML configs anyone can understand
-   **AI‑friendly by design** — conventional structures that AI tools navigate well
-   **Minimal required tooling** — single binary, zero runtime dependencies
-   **Works offline** — local sources for air‑gapped environments

------------------------------------------------------------------------

End of Functional Specification v0.2
