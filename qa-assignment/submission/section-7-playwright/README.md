# Section 7: Playwright Automation

TypeScript Playwright tests for AbleSpace staging (https://staging.ablespace.io).

## What is automated

| Spec | Flow | Needs login? |
|---|---|---|
| `tests/add-student.spec.ts` | **Login → Add Student (manually) → verify it's listed in the caseload → refresh → verify it's still there, exactly once** | Yes |
| `tests/signin-validation.spec.ts` | Sign-in email step: empty email, malformed email, unregistered email, and the redirect when a logged-out user opens `/caseload` | No |

The suite never automates sign-up. It logs in with the one account I created by hand.

## 1. Install

You need **Node.js 18 or newer** (check with `node -v`).

```bash
cd section-7-playwright
npm install
npx playwright install chromium
```

## 2. Set the login credentials

Credentials are **not** in the code. They're read from environment variables, which you can put in a `.env` file. `.env` is listed in `.gitignore`, so it is never committed.

```bash
cp .env.example .env        # Windows PowerShell: copy .env.example .env
```

Then edit `.env`:

```
ABLESPACE_EMAIL=your-test-account@example.com
ABLESPACE_PASSWORD=your-password
```

You can also set them in the shell instead of a file: `export ABLESPACE_EMAIL=...` on macOS or Linux, or `$env:ABLESPACE_EMAIL="..."` in PowerShell.

If the credentials are missing, the login test is **skipped** rather than failed, and the sign-in validation tests still run.

> The account must use **email + password**. Accounts created with "Continue with Google", Microsoft, Clever or ClassLink have no password step to automate.

## 3. Run

```bash
npm test                 # headless, all tests
npm run test:headed      # watch the browser
npm run test:ui          # Playwright UI mode: step through and pick locators
npm run report           # open the HTML report from the last run
```

Run a single file: `npx playwright test tests/add-student.spec.ts`

When a test fails, the HTML report includes a screenshot, a video and a trace (open it with `npx playwright show-trace <path>`).

## How the code is organised

```
section-7-playwright/
├── playwright.config.ts            baseURL, 1 worker, trace/screenshot/video on failure, loads .env
├── pages/login-page.ts             page object for the two-step sign-in form
├── tests/add-student.spec.ts       main flow: login → add student → verify
├── tests/signin-validation.spec.ts negative checks on the email step
├── .env.example                    template for credentials (copy to .env)
└── .gitignore                      keeps .env, node_modules and reports out of git
```

## Design choices

- **Locators a user would recognise.** I used `getByRole` (buttons, dialog, headings) and `getByLabel` (form fields) instead of CSS classes, which change whenever the UI is restyled. For example, `getByRole('button', { name: 'Continue', exact: true })` matches the Continue button but not "Continue with Google".
- **Meaningful assertions.** The tests don't stop at "no error". They check the exact validation message, that a password step does *not* appear for an unknown email, that the new student is visible, that it survives `page.reload()`, and that it appears exactly once (no duplicate from a double submit).
- **No fixed waits.** There's no `waitForTimeout`. Playwright's web-first assertions (`expect(...).toBeVisible()`) retry until the condition holds or the timeout expires, which keeps the tests both fast and stable.
- **Unique test data.** Each run creates `QA-Auto Test<timestamp>`, so runs never collide. These students stay in the test account; delete them by hand when you want a tidy caseload.
- **Gentle on a live system.** One worker, no retries, and only a handful of requests per run.

## Status

- `signin-validation.spec.ts`: 4/4 passing against staging (Chromium 141).
- `add-student.spec.ts`: the post-login locators (`Add Student`, `Add Manually`, `First name`, `Last name`, `Create`) follow AbleSpace's help-centre wording. _TODO: confirm them after the first logged-in run, and fix any that don't match using `npm run test:ui` or `npx playwright codegen https://staging.ablespace.io`._
