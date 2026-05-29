# SVC-Frontend

`svc-frontend` is the web client for the `sci-vault` application. Built with [SvelteKit](https://svelte.dev/docs/kit/introduction) (Svelte 5), [Vite](https://vite.dev/), and [Tailwind CSS v4](https://tailwindcss.com/), it provides a modern, responsive user interface for exploring and receiving personalized article recommendations powered by the backend microservices.

## Getting Started

Follow the instructions below to set up and run the frontend development server locally.

### Prerequisites

- [Bun](https://bun.sh/) 1.0+ - A fast all-in-one JavaScript runtime and package manager

### Install Dependencies

Navigate to the project directory and install the required dependencies using Bun:

```bash
bun install
```

### Run the Development Server

Start the development server with hot module reloading:

```bash
bun run dev
```

Optionally, open your browser automatically:

```bash
bun run dev -- --open
```

The application will be available at `http://localhost:5173/` by default.

## Building for Production

To create an optimized production build:

```bash
bun run build
```

### Preview Production Build

Test your production build locally before deployment:

```bash
bun run preview
```

## Development Commands

- `bun run check`: Run SvelteKit type checking and Svelte component validation.
- `bun run check:watch`: Watch mode for type checking and validation.
- `bun run lint`: Check code formatting and linting issues.
- `bun run format`: Auto-fix formatting with Prettier.

## Tech Stack

- **Framework**: [SvelteKit](https://kit.svelte.dev/) with [Svelte 5](https://svelte.dev/) (Runes API)
- **Build Tool**: [Vite](https://vitejs.dev/) - Lightning-fast build tool
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/) - Utility-first CSS framework
- **UI Components**: [Bits UI](https://bits-ui.com/) - Headless component primitives (shadcn-svelte)
- **Data Tables**: [TanStack Table](https://tanstack.com/table) - Server-side paginated/sorted/filtered list pages
- **Charts**: [LayerChart](https://www.layerchart.com/) - Dashboard charts (uploads, engagement, format distribution)
- **Language**: [TypeScript](https://www.typescriptlang.org/) - Type safety and developer experience
- **HTTP Client**: [Axios](https://axios-http.com/) - Promise-based HTTP client
- **Authentication**: [jwt-decode](https://github.com/auth0/jwt-decode) - JWT token parsing
- **Internationalization**: [svelte-i18n](https://github.com/kaisermann/svelte-i18n) - i18n with `en` and `zh-CN` locales
- **Theme**: [mode-watcher](https://mode-watcher.svecosystem.com/) - Dark/light mode management
- **Notifications**: [svelte-sonner](https://svelte-sonner.vercel.app/) - Toast notifications
- **Icons**: [Lucide Svelte](https://lucide.dev/) - Icon library
- **Code Quality**: [ESLint](https://eslint.org/) and [Prettier](https://prettier.io/)
- **Adapter**: [@sveltejs/adapter-static](https://kit.svelte.dev/docs/adapter-static) - Static site generation

## Project Structure

```text
src/
├── routes/
│   ├── (dashboard)/          # Authenticated pages wrapped by AppSidebar
│   │   ├── +layout.svelte    # Sidebar navigation layout
│   │   ├── +page.svelte      # Home / feed page
│   │   ├── documents/        # Document library + detail
│   │   ├── lab-documents/    # Lab-visible documents
│   │   ├── labs/             # Lab management
│   │   ├── lab-settings/     # Lab settings (owner)
│   │   ├── members/          # Lab members
│   │   ├── collaborators/    # Suggested collaborators
│   │   ├── recommendations/  # Personalised feed
│   │   ├── search/           # Semantic search
│   │   ├── mine/             # Interaction history (views/likes)
│   │   ├── profile/          # User profile pages
│   │   └── settings/         # User settings
│   ├── login/                # Login / registration page
│   ├── welcome/              # Public welcome / landing page
│   └── +layout.svelte        # Root layout
├── lib/
│   ├── api/                  # API client modules (auth, user, document, lab, ...)
│   ├── actions/              # Custom Svelte actions
│   ├── components/
│   │   ├── dashboard/        # Shared dashboard charts
│   │   ├── layout/           # Layout components (ThemeToggle, etc.)
│   │   └── ui/               # Reusable UI components (shadcn-svelte / Bits UI)
│   ├── hooks/                # Client-side Svelte hooks
│   ├── stores/               # Svelte 5 runes stores (lab, nav, search, user)
│   ├── locales/              # i18n translation files (en.json, zh-CN.json)
│   ├── assets/               # Static assets
│   ├── i18n.ts               # i18n setup
│   └── utils.ts              # Shared utilities
├── app.d.ts                  # TypeScript ambient declarations
└── app.html                  # HTML entry point
```

## Docker Deployment

This application comes with a Dockerfile for easy containerization.
To build and start the frontend along with the rest of the application utilizing Docker Compose, run the following from the root directory:

```bash
cd ..
docker compose up -d --build frontend
```

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
