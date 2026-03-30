# env-pilot — Guided .env Setup CLI

## Problem

Setting up `.env` files is tedious, error-prone, and repetitive. Developers copy a template, then fill in values one-by-one in a text editor — with no guidance on what each key is, where to get its value, or whether they already have it somewhere else.

## What This Tool Does

A CLI that reads a `.env.example` (or similar template) and walks you through filling in each value interactively, producing a working `.env` file.

## Core Behaviors

### Parsing

- Reads template files: `.env.example`, `.env.sample`, `.env.template`, `.env.dist`, `.env.defaults`
- Auto-detects which template exists in the current directory (or accepts explicit path via flag)
- Extracts for each key:
  - **Key name** (e.g., `DATABASE_URL`)
  - **Default/placeholder value** if present (e.g., `postgres://localhost:5432/mydb`)
  - **Description** from adjacent comment lines above the key
  - **Section** from comment-based groupings (e.g., `# ── Database ──`)
- If a `.env` already exists, reads it to determine which keys are already set

### Output

- Writes to `.env` in the current directory (or path specified via flag)
- Merges with existing `.env` — never overwrites values already set
- Preserves key order from the template
- Does not copy comments or sections into output (clean KEY=VALUE only)

---

## UX Flow

### 1. Launch & Status

On launch, show a summary dashboard:

```
env-pilot — reading from .env.example

  12 keys total
   4 already set in .env
   8 remaining

  [s] Start setup    [r] Review all    [q] Quit
```

If no template is found, show an error with the list of filenames it looked for.

### 2. Key-by-Key Prompt

Each key gets a focused prompt screen:

```
─── Database (2 of 8 remaining) ─────────────────────

  DATABASE_URL
  PostgreSQL connection string for the app database.

  Default: postgres://localhost:5432/mydb

  ▸ Enter value (or ↵ for default, [s]kip, [b]ack):
```

**Elements:**
- **Section header** — from comment groupings in template. Gives context.
- **Progress** — "2 of 8 remaining" so user knows where they are.
- **Key name** — prominent, top of block.
- **Description** — from comment above the key in template. May be multi-line. Omitted if no comment exists.
- **Default** — shown if the template had a value. If the value looks like a placeholder (`your-api-key-here`, `change-me`, `xxx`), label it as "Placeholder" instead of "Default".
- **Input line** — where the user types or pastes.

**Input behaviors:**
- `Enter` with empty input → accept default (if one exists)
- `s` or `skip` → skip this key, move to next
- `b` or `back` → go to previous key
- Typing a value → set it
- `Ctrl+C` → save progress and exit (partial `.env` is valid)

### 3. Sensitive Value Handling

If a key name matches common sensitive patterns (`*_KEY`, `*_SECRET`, `*_TOKEN`, `*_PASSWORD`, `*_CREDENTIAL*`), input is masked:

```
  OPENAI_API_KEY
  API key for OpenAI. Get it at https://platform.openai.com/api-keys

  ▸ Enter value: ••••••••••••••sk-1234

  [r]eveal to verify?
```

- Last 7 characters shown for verification
- `r` toggles full reveal momentarily
- Masked by default, user can override

### 4. Skipped Keys

Skipped keys are tracked separately from "not yet reached" keys. On subsequent runs, only unanswered keys are prompted — skipped keys appear at the end:

```
  ─── Done! ──────────────────────────────────────────

   8 keys set (4 previously + 4 just now)
   2 skipped
   2 optional (blank in template, no value given)

  [v] View skipped    [w] Write & exit    [q] Quit without saving
```

### 5. Review Mode

Accessible from the dashboard (`[r] Review all`) or after completion. Shows all keys with their status:

```
  ─── Review ─────────────────────────────────────────

  Database
    ✓ DATABASE_URL          postgres://localhost:5432/mydb
    ✓ DB_POOL_SIZE          10

  Authentication
    ✓ JWT_SECRET            ••••••••••••abcd
    ⏭ OAUTH_CLIENT_ID       (skipped)
    ⏭ OAUTH_CLIENT_SECRET   (skipped)

  External APIs
    ✓ OPENAI_API_KEY        ••••••••sk-1234
    ○ STRIPE_API_KEY         (not set)
    ✓ SENTRY_DSN            https://abc@sentry.io/123

  [e] Edit a key    [w] Write & exit
```

**Status indicators:**
- `✓` — value set
- `⏭` — explicitly skipped
- `○` — not yet addressed
- Sensitive values masked in review

### 6. Edit Mode

From review, user can type a key name (or number) to jump directly to it and change its value.

---

## CLI Interface

```
env-pilot                    # Auto-detect template, interactive setup
env-pilot --from .env.local  # Explicit template file
env-pilot --out .env.local   # Write to a specific output file
env-pilot --review           # Jump straight to review mode
env-pilot --status           # Print summary (how many set/missing) and exit
```

### Flags

| Flag | Description |
|---|---|
| `--from <file>` | Path to template file. Default: auto-detect. |
| `--out <file>` | Path to output file. Default: `.env`. |
| `--review` | Show review screen and exit. |
| `--status` | Print key counts (set/missing/skipped) and exit. Non-interactive. |
| `--all` | Re-prompt all keys, including already-set ones. |
| `--section <name>` | Only prompt keys in a specific section. |
| `--help` | Show help. |

---

## Design Considerations for the Designer

### Visual Hierarchy (in terminal)

1. **Section name** — subtle, provides grouping context
2. **Key name** — bold/prominent, the anchor of each prompt
3. **Description** — secondary text, softer
4. **Default/placeholder** — distinct from description (maybe dimmed or different color)
5. **Input prompt** — clear active area
6. **Progress indicator** — persistent, unobtrusive

### Color Usage

- Use sparingly. Not all terminals support 256-color or truecolor.
- Suggested palette (adapt to your design):
  - Key names: **bold white**
  - Descriptions: dim/gray
  - Defaults: cyan or blue
  - Status ✓: green
  - Status ⏭/○: yellow/dim
  - Sensitive mask: dim
  - Section headers: bold + underline or a horizontal rule
  - Errors: red

### Terminal Size

- Design for **80-column minimum** width
- Long values (URLs, keys) will wrap — don't fight it, just let them flow
- Description text should soft-wrap at terminal width

### Interaction Model

- This is a **linear wizard with random access** — primarily step-by-step, but user can jump around
- Keep it feeling fast: one keypress to skip, one keypress to go back
- No confirmation dialogs — write on exit, `Ctrl+C` is safe (saves progress)

### Things to Avoid

- Don't make it feel like a form with 30 fields. The one-at-a-time focus is intentional.
- Don't use animation or spinners — this is a fast, local tool.
- Don't require scrolling within a single prompt. If a description is long, truncate with "..." and offer a way to expand.

---

## State Management

The tool needs to track "skipped" status between sessions. Two options:

**Option A: Comment marker in `.env`**
```bash
# env-pilot:skipped
OAUTH_CLIENT_ID=
```

**Option B: Sidecar file `.env.status`**
```json
{ "skipped": ["OAUTH_CLIENT_ID", "OAUTH_CLIENT_SECRET"] }
```

Recommendation: **Option A** — no extra files, and the `.env` remains valid. The comment is a convention the tool reads; everything else ignores it.

---

## Out of Scope (for now)

- Shared key vault / reuse across projects (feature 3)
- Auto-discovery of where to get values (feature 2)
- Validation of values (type checking, URL format, etc.)
- Framework detection (auto-detecting Next.js vs Laravel conventions)
- Team/remote sync
