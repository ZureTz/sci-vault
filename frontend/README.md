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
- **UI Components**: [Bits UI](https://bits-ui.com/) - Headless component primitives
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
│   ├── (dashboard)/          # Authenticated dashboard layout
│   │   ├── +layout.svelte    # Sidebar navigation layout
│   │   ├── +page.svelte      # Home / feed page
│   │   ├── profile/
│   │   │   └── [user_id]/    # User profile page
│   │   └── settings/         # Settings page
│   ├── login/                # Login / registration page
│   └── +layout.svelte        # Root layout
├── lib/
│   ├── api/                  # API client modules (auth, user)
│   ├── components/
│   │   ├── layout/           # Layout components (ThemeToggle, etc.)
│   │   └── ui/               # Reusable UI components (Bits UI based)
│   ├── hooks/                # Client-side Svelte hooks
│   ├── locales/              # i18n translation files (en.json, zh-CN.json)
│   └── assets/               # Static assets
├── app.d.ts                  # TypeScript ambient declarations
└── app.html                  # HTML entry point
```

## Roadmap: Docker

Upcoming improvements include:

- Docker containerization for consistent development and deployment environments

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
