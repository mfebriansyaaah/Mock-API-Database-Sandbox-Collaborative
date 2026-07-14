# Mock Sandbox Console

Web UI for the [Mock API & Database Sandbox](../) Go backend. Provides a visual
console for managing projects, tables, documents, and testing endpoints.

## Stack

- **Vite 5** + **React 18** (JavaScript, not TypeScript)
- **Tailwind CSS 3** for styling
- **React Router 6** for routing
- **Zustand 5** for state management (persisted to `localStorage`)
- **Axios** for HTTP
- **Lucide React** for icons
- **nanoid** for ID generation

## Getting started

```bash
cd web-ui
npm install
npm run dev
```

The dev server runs on <http://localhost:5173>. `vite.config.js` proxies
`/sandbox` and `/hello` to the Go backend on <http://localhost:8080>, so the
UI uses **relative URLs** and the browser never sees a CORS preflight. Make
sure the Go server is running first (`make run` from the project root).

> Because the dev server proxies requests, the configured backend URL is only
> used for non-proxy hops (e.g. a remote deployment). In local dev the default
> empty string is correct.

## Scripts

| Script            | Purpose                          |
| ----------------- | -------------------------------- |
| `npm run dev`     | Start Vite dev server with proxy |
| `npm run build`   | Production build into `dist/`    |
| `npm run preview` | Preview the production build     |

## Configuration

The UI reads its base URL from the **Settings** page (default: empty string
so that the Vite dev proxy is used). The setting is persisted to
`localStorage` under `mocksbx.settings.v1`.

Project and table bookmarks are stored in `localStorage` under
`mocksbx.projects.v1` and request history under `mocksbx.tester.v1`.

The Settings page also supports **export / import** of the full
configuration as a single JSON blob.

## Pages

| Route                                       | Purpose                                                                                  |
| ------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `/`                                         | **Overview** — Live stats: total projects, tables, and documents, with recent activity.  |
| `/projects`                                 | **Projects** — Manage sandbox projects (tracked locally; bookmarks only).                |
| `/projects/:projectId`                      | **Tables** — List and search tables within a project.                                    |
| `/projects/:projectId/:table`               | **Documents** — Browse, search, paginate, create, edit, and delete with bulk operations. |
| `/projects/:projectId/:table/:id`           | **Document Detail** — Read-only view with side-by-side metadata and inline JSON editor.   |
| `/tester`                                   | **REST Tester** — Send arbitrary `GET/POST/PUT/PATCH/DELETE` requests; `Ctrl+Enter` sends.|
| `/logs`                                     | **Access Logs** — Local view of requests issued from the console.                        |
| `/settings`                                 | **Settings** — Backend URL, theme, density, export/import configuration.                 |
| `*`                                         | **404** — Friendly not-found page.                                                       |

## Features

- **Server-side pagination** — `?limit=N&offset=M` (default 25, max 100) is
  sent to the backend; the response envelope `{ data, limit, offset, count,
  nextOffset }` drives the page controls.
- **Quota-aware UX** — Firestore `ResourceExhausted` responses are surfaced
  as a friendly "Backend quota exceeded" banner with a retry button, instead
  of a raw `HTTP 500` stack trace.
- **REST Tester** — Method picker, path autocomplete, JSON body editor, and
  a full response viewer. `Ctrl+Enter` (or `Cmd+Enter` on macOS) sends the
  request.
- **Light / Dark theme** with smooth transitions, persisted per browser.
- **Responsive** — desktop-first with a mobile bottom-nav below the `md`
  breakpoint.
- **Stable React effects** — page-level `useEffect` hooks derive a stable
  signature from their inputs to avoid infinite re-render loops.
