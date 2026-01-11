# {{ project_name }}

{{ description }}

## Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) 🎉

## Features

- ⚡️ **Next.js 16** - Latest with Turbopack
- ⚛️ **React 19** - Latest features
- 🎨 **Tailwind CSS v4** - Modern CSS with OKLCH colors
- 🧩 **shadcn/ui** - Pre-configured component library
- 📝 **TypeScript** - Full type safety
- 🎯 **ESLint** - Next.js recommended config

## Adding Components

```bash
# Add shadcn/ui components
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add dialog
```

## Available Scripts

```bash
npm run dev      # Development server
npm run build    # Production build
npm start        # Production server
npm run lint     # Run ESLint
```

## Project Structure

```
{{ project_slug }}/
├── src/
│   ├── app/
│   │   ├── globals.css      # Tailwind + theme
│   │   ├── layout.tsx       # Root layout
│   │   └── page.tsx         # Home page
│   ├── components/ui/       # shadcn/ui components
│   └── lib/utils.ts         # Utilities (cn helper)
├── components.json          # shadcn/ui config
└── README.md
```

## Deployment

### Vercel (Recommended)
```bash
npm install -g vercel
vercel
```

### Other Platforms
```bash
npm run build
npm start
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [shadcn/ui Documentation](https://ui.shadcn.com)
- [React Documentation](https://react.dev)
