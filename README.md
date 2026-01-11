<p align="center">
  <img src="https://raw.githubusercontent.com/makemore/scaffold/main/frontend/public/scaffold-logo.svg" alt="Scaffold" width="120" height="120">
</p>

<h1 align="center">Scaffold</h1>

<p align="center">
  <strong>Bootstrap any software stack with sensible defaults</strong>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#features">Features</a> •
  <a href="#templates">Templates</a> •
  <a href="#creating-templates">Creating Templates</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <img src="https://img.shields.io/github/v/release/makemore/scaffold?style=flat-square" alt="Release">
  <img src="https://img.shields.io/github/license/makemore/scaffold?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/makemore/scaffold/release.yml?style=flat-square" alt="Build">
</p>

---

## What is Scaffold?

Scaffold is a **source-agnostic, composable project bootstrapper** that lets you:

- 🚀 **Start projects instantly** with production-ready templates
- 🔧 **Compose features** by layering modules on top of base templates
- 🌐 **Use any source** — GitHub, GitLab, local files, or any git repo
- 📦 **Zero dependencies** — single binary, works everywhere

Unlike framework-specific CLIs or heavy platform solutions, Scaffold works with **any technology stack** and pulls templates from **anywhere**.

## Quick Start

```bash
# Using npx (no installation required)
npx @makemore/scaffold myapp

# Or install the CLI globally
brew install makemore/tap/scaffold  # macOS
# Then run:
scaffold init myapp
```

### Interactive Mode

Just run `scaffold init` and follow the prompts:

```
? Select a template:
  ❯ django     - Django REST API with authentication, Cloud Tasks, S3 storage
    nextjs     - Next.js 15 with TypeScript, Tailwind CSS, shadcn/ui

? Project name: my-awesome-app
? Project slug: my_awesome_app
? Description: An awesome new project
? GCP Project ID: my-gcp-project

✓ Project created at ./my-awesome-app
```

### Non-Interactive Mode

Perfect for CI/CD or scripting:

```bash
scaffold init myapp \
  --base django \
  --var project_name="My App" \
  --var project_slug=my_app \
  --var gcp_project=my-gcp-project
```

### Use Any Source

```bash
# GitHub shorthand
scaffold init myapp --base github:org/repo

# GitLab
scaffold init myapp --base gitlab:org/repo

# Any git URL
scaffold init myapp --base git:https://git.company.com/templates/base

# Local path (great for development)
scaffold init myapp --base file:./my-templates/django

# With subdirectory and branch
scaffold init myapp --base github:org/repo//templates/django#v2.0
```

## Features

### 🧩 Composable Modules

Layer features on top of your base template:

```bash
scaffold init myapp \
  --base django \
  --add github:org/modules//celery \
  --add github:org/modules//stripe
```

Each module can add files, modify existing ones, and define its own variables.

### 📝 Smart Variable Substitution

Templates use simple `{{ variable }}` syntax:

```python
# settings.py
PROJECT_NAME = "{{ project_name }}"
```

```yaml
# scaffold.yaml
variables:
  - name: project_name
    description: Name of your project
    required: true
  - name: database
    type: choice
    choices: [postgres, mysql, sqlite]
    default: postgres
```

### 🔄 Directory Renaming

Use `__variable__` in directory names:

```
__project_slug__/
  settings.py
  urls.py
```

Becomes:

```
my_app/
  settings.py
  urls.py
```

### 🎯 Post-Generation Actions

Templates can define actions to run after generation:

```yaml
actions:
  - name: install
    type: command
    command: pip install -r requirements.txt
  - name: migrate
    type: command
    command: python manage.py migrate
  - name: welcome
    type: message
    message: |
      🎉 Your project is ready!

      Run: cd {{ project_slug }} && python manage.py runserver
```

## Templates

### Official Templates

| Template | Description |
|----------|-------------|
| `django` | Django REST API with auth, Cloud Tasks, S3, GCP deployment |
| `nextjs` | Next.js 15 with TypeScript, Tailwind CSS, shadcn/ui |

### Using Custom Templates

Any git repository with a `scaffold.yaml` can be used as a template:

```bash
# Your company's internal templates
scaffold init myapp --base git:git@github.com:company/templates.git//django

# Community templates
scaffold init myapp --base github:awesome-user/cool-template
```

## Creating Templates

### Basic Structure

```
my-template/
├── scaffold.yaml          # Template manifest
├── README.md
├── __project_slug__/      # Renamed to project slug
│   └── settings.py
└── requirements.txt
```

### scaffold.yaml

```yaml
name: my-template
description: A great template for starting projects
type: base
version: "1.0.0"

variables:
  - name: project_name
    description: Human-readable project name
    required: true

  - name: project_slug
    description: Python package name
    required: true

  - name: author
    description: Author name
    default: Anonymous

  - name: license
    type: choice
    choices:
      - MIT
      - Apache-2.0
      - GPL-3.0
    default: MIT

files:
  exclude:
    - "*.pyc"
    - "__pycache__"
    - ".git"

actions:
  - name: welcome
    type: message
    message: "Project {{ project_name }} created successfully!"
```

### Variable Types

| Type | Description |
|------|-------------|
| `string` | Free-form text input (default) |
| `choice` | Select from predefined options |
| `confirm` | Yes/no boolean |

## Installation

### macOS (Homebrew)

```bash
brew install makemore/tap/scaffold
```

### Download Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/makemore/scaffold/releases).

### Using npx

No installation required:

```bash
npx @makemore/scaffold myapp
```

## CLI Reference

```bash
scaffold init [name] [flags]

Flags:
  -b, --base string      Base template (name, URL, or path)
  -a, --add strings      Additional modules to layer
  -v, --var strings      Variables in key=value format
  -o, --output string    Output directory (default: current directory)
  -y, --yes              Skip confirmation prompts
  -h, --help             Help for init

scaffold list           # List available templates
scaffold version        # Show version
```

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+ (for npx wrapper)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/makemore/scaffold.git
cd scaffold

# Build the CLI
cd cli
go build -o scaffold .

# Run tests
go test ./...
```

### Running Template Tests

```bash
# Test Django template
./tests/templates/test_django.sh --no-docker

# Test Next.js template
./tests/templates/test_nextjs.sh --no-docker
```

## Release Process

### Creating a New Release

1. **Update version** in relevant files if needed

2. **Create and push a tag:**
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. **GitHub Actions will automatically:**
   - Build binaries for all platforms (darwin/linux/windows × amd64/arm64)
   - Create a GitHub release with the binaries attached
   - Generate release notes from commits

4. **Publish the npm wrapper:**

   Option A — Using 2FA (interactive):
   ```bash
   cd create-scaffold
   npm version 0.1.0
   npm publish
   # Enter your 2FA code when prompted
   ```

   Option B — Using a granular access token:
   ```bash
   # Create a token at https://www.npmjs.com/settings/tokens
   # Select "Granular Access Token" with publish permissions
   # Enable "Bypass 2FA for automation"
   npm config set //registry.npmjs.org/:_authToken=YOUR_TOKEN
   npm publish
   ```

### Platform Binaries

Each release includes binaries for:
- `scaffold-darwin-amd64` — macOS Intel
- `scaffold-darwin-arm64` — macOS Apple Silicon
- `scaffold-linux-amd64` — Linux x64
- `scaffold-linux-arm64` — Linux ARM64
- `scaffold-windows-amd64.exe` — Windows x64

## Contributing

We welcome contributions! Here's how to get started:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `cd cli && go test ./...`
5. Commit: `git commit -m 'Add amazing feature'`
6. Push: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Adding a New Template

1. Create a new directory in `templates/`
2. Add a `scaffold.yaml` manifest
3. Add template files with `{{ variable }}` placeholders
4. Add tests in `tests/templates/`
5. Update `templates.yaml` to register the template

## License

MIT License — see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by [cookiecutter](https://github.com/cookiecutter/cookiecutter), [degit](https://github.com/Rich-Harris/degit), and the many framework-specific scaffolding tools that came before.

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/makemore">MakeMore</a>
</p>

