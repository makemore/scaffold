# create-scaffold

Bootstrap any software stack with sensible defaults.

## Quick Start

```bash
# Create a new Django project
npx create-scaffold myapp --base django

# Create a new Next.js project
npx create-scaffold myapp --base nextjs

# Interactive mode
npx create-scaffold myapp
```

## Available Templates

| Template | Description |
|----------|-------------|
| `django` | Django REST API with authentication, Cloud Tasks, S3 storage |
| `nextjs` | Next.js 15 with TypeScript, Tailwind CSS, shadcn/ui |

## Usage

```bash
# Using npx (recommended)
npx create-scaffold <project-name> [options]

# Or install globally
npm install -g create-scaffold
scaffold init <project-name> [options]
```

## Options

```
-b, --base <template>   Base template (django, nextjs, or URL)
-a, --add <module>      Additional modules to layer
-v, --var <key=value>   Template variables
-o, --output <dir>      Output directory
    --no-prompt         Disable interactive prompts
```

## Custom Templates

Use any git repository as a template:

```bash
# GitHub
npx create-scaffold myapp --base github:org/repo

# GitLab
npx create-scaffold myapp --base gitlab:org/repo

# Any git URL
npx create-scaffold myapp --base git:https://git.example.com/repo

# Local path
npx create-scaffold myapp --base file:~/templates/my-template
```

## Documentation

For full documentation, visit: https://github.com/christophercochran/scaffold

## License

MIT

