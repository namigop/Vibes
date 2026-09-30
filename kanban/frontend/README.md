# Kanban — frontend

Next.js (App Router) + TypeScript + Tailwind v4 client for the Kanban API
described in [`../docs/API.md`](../docs/API.md).

## Running it

```powershell
npm install
npm run dev     # http://localhost:3000
```

The Go backend must be listening on port 8080 (see `../docs/API.md`); the API is
called directly from the browser, so CORS has to allow `http://localhost:3000`.

> On this machine ports 8080 and 3000 are already taken by the `todo-app` stack,
> so the backend was run on **8081** and `.env.local` points there. Stop the
> other stack, or change `BACKEND_PORT` and `NEXT_PUBLIC_API_BASE_URL`
> together — they must agree.

```powershell
npm run lint    # eslint
npm run build   # next build (also type-checks)
npm start       # serve the production build
```

### Cross-layer parity test

```powershell
npm run parity                        # backend must already be running
KANBAN_API=http://localhost:8080/api/v1 npm run parity
```

`test/parity.mts` takes the real `board-state.ts` helpers — the same ones the
drag layer uses for its optimistic update — points them at a live API, and
asserts that the index the UI predicts is the index the server settles on, for
both cross-column and within-column moves. It creates a scratch board and
deletes it afterwards. It is excluded from the app's `tsconfig.json`, so it is
type-checked by neither `next build` nor the app compiler.

It exists because "the card lands one slot off after a cross-column drag" is
the kind of bug that no amount of unit testing on either side catches on its
own — only the two halves running against each other will.

## API base URL

`src/lib/constants.ts` exports `API_BASE_URL`, resolved as:

1. `process.env.NEXT_PUBLIC_API_BASE_URL`
2. otherwise `http://localhost:8080/api/v1`

Because it is a `NEXT_PUBLIC_` variable it is inlined **at build time** — change
`.env.local` (or the deployment env) and rebuild; editing it while `next dev` is
already running will not always be picked up.

## Layout

```
src/
  app/
    layout.tsx              root: fonts, pre-paint theme script, providers
    globals.css             design tokens + the Tailwind v4 `dark:` override
    (app)/layout.tsx        app shell: side nav + top bar
    (app)/page.tsx          redirects to the first board, or "create your first"
    (app)/board/[boardId]/  the board route (board id lives in the URL)
  components/
    layout/                 side nav, top bar, theme toggle, new-board form
    board/                  board view, columns, cards, drag and drop
    card/                   card drawer, priority badge, avatar, due date
    ui/                     button, inputs, menu, confirm dialog, toaster
  hooks/                    use-board (all board state + writes), use-is-client
  lib/                      api client, types, date + state helpers
  providers/                theme, toast, boards list
```

## Notes for the next person

- **Dark mode**: `globals.css` re-declares `@custom-variant dark (&:where(.dark,
  .dark *))`. Tailwind v4 otherwise compiles `dark:` to a `prefers-color-scheme`
  media query and the toggle would do nothing. A blocking script in `<head>`
  applies the class before first paint (`themeInitScript` in
  `src/providers/theme-provider.tsx`); the same resolution logic is repeated
  client-side in `resolveInitialTheme`.
- **Optimistic writes**: `useBoard` mutates local state first and sends the
  request afterwards. Each entity has a write counter, so a slow response can
  never clobber a newer local state, and failures restore the pre-mutation
  snapshot and toast the server's `message`.
- **Dates**: `due_date` is a calendar date (`YYYY-MM-DD`). `src/lib/dates.ts`
  parses it into a local `Date` by hand — `new Date("2026-10-15")` is UTC
  midnight and would display as the 14th west of UTC.
