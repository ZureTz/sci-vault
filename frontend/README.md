# SVC-Frontend

`svc-frontend` is the web client for the `sci-vault` application. Built with [SvelteKit](https://svelte.dev/docs/kit/introduction), [Vite](https://vite.dev/), and [Tailwind CSS](https://tailwindcss.com/), it provides a modern, responsive user interface for exploring and receiving personalized article recommendations powered by the backend microservices.

## Getting Started

Follow the instructions below to set up and run the frontend development server locally.

### Prerequisites

- [Bun](https://bun.sh/) 1.0+ -  A fast all-in-one JavaScript runtime and package manager

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

Optionally, you can open your browser automatically:

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
- `bun run format`: Auto-fix formatting with Prettier and ESLint.

## Tech Stack

- **Framework**: [SvelteKit](https://kit.svelte.dev/) - Modern, lightweight Svelte framework
- **Build Tool**: [Vite](https://vitejs.dev/) - Lightning-fast build tool
- **Styling**: [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- **Language**: [TypeScript](https://www.typescriptlang.org/) - Type safety and developer experience
- **Code Quality**: [ESLint](https://eslint.org/) and [Prettier](https://prettier.io/) - Linting and formatting
- **Icons**: [Lucide Svelte](https://lucide.dev/) - Beautiful icon library
- **UI Utilities**: [clsx](https://github.com/lukeed/clsx) and [tailwind-merge](https://github.com/dcastil/tailwind-merge)

## Project Structure

```text
src/
├── routes/           # SvelteKit pages and layouts
├── lib/
│   ├── components/   # Reusable UI components
│   ├── assets/       # Static assets and images
│   ├── hooks/        # Client-side hooks
│   └── utils.ts      # Utility functions
├── app.d.ts          # TypeScript ambient declarations
└── app.html          # HTML entry point
```

## Deployment

The built application is optimized for deployment. You may need to install an appropriate [SvelteKit adapter](https://kit.svelte.dev/docs/adapters) depending on your target environment (Node.js, serverless, static, etc.).

## Roadmap: API Integration & Docker

Upcoming improvements include:
- Seamless integration with the `svc-gateway` API
- Docker containerization for consistent development and deployment environments
- Enhanced UI components and user experience improvements

## License

This project is licensed under the [LICENSE](../LICENSE) file in the root directory.
