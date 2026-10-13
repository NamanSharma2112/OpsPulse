# Section 4: Bug Reports

Bugs are ordered by importance: severity first, then priority.

**Severity** is how much the bug hurts the product (Critical, High, Medium, Low). **Priority** is how soon I think it should be fixed (High, Medium, Low).

| ID | Title | Severity | Priority |
|---|---|---|---|
| BUG-01 | _TODO_ | _TODO_ | _TODO_ |
| BUG-02 | _TODO_ | _TODO_ | _TODO_ |
| BUG-03 | "User not found!" sign-in error disappears after ~5 s, appears far from the email field, and offers no path to sign up | Low | Medium |

---

## BUG-01: _TODO: short, specific title (what is broken + where)_

| Field | Description |
|---|---|
| **Bug ID** | BUG-01 |
| **Title** | _e.g. "Editing a student's last name doesn't update the name shown in the Data tab until re-login"_ |
| **Severity** | _Critical / High / Medium / Low_ |
| **Priority** | _High / Medium / Low_ |
| **Environment** | https://staging.ablespace.io/_<page>_ · tested _<date>_ |
| **Browser / Device** | _e.g. Google Chrome 141.0.x on Windows 11, 1920 × 1080_ |
| **Reproducibility** | _Always (x/x attempts) / Sometimes (x/y) / Once_ |
| **Preconditions** | _e.g. logged in; a student "QA-Leo Test02" exists with one Accuracy goal_ |
| **Steps to Reproduce** | 1. _start from a clear point, e.g. "Open https://staging.ablespace.io/caseload"_<br>2. _…_<br>3. _…_ |
| **Expected Result** | _what should happen_ |
| **Actual Result** | _what actually happens: exact error text, wrong value, etc._ |
| **Impact** | _who is affected and why it matters (wrong IEP data? blocked task? data loss?)_ |
| **Evidence** | _screenshot or recording link (Google Drive link set to "Anyone with the link can view", or a file in `evidence/`)_ |

## BUG-02: _TODO_

| Field | Description |
|---|---|
| **Bug ID** | BUG-02 |
| **Title** | _TODO_ |
| **Severity** | _TODO_ |
| **Priority** | _TODO_ |
| **Environment** | https://staging.ablespace.io/_<page>_ · tested _<date>_ |
| **Browser / Device** | _TODO_ |
| **Reproducibility** | _TODO_ |
| **Preconditions** | _TODO_ |
| **Steps to Reproduce** | 1. _…_ |
| **Expected Result** | _TODO_ |
| **Actual Result** | _TODO_ |
| **Impact** | _TODO_ |
| **Evidence** | _TODO_ |

## BUG-03: "User not found!" sign-in error disappears after ~5 s, appears far from the email field, and offers no path to sign up

| Field | Description |
|---|---|
| **Bug ID** | BUG-03 |
| **Title** | "User not found!" sign-in error disappears after ~5 s, appears far from the email field, and offers no path to sign up |
| **Severity** | **Low**. It's a usability problem: sign-in still works and the user can retry. |
| **Priority** | **Medium**. Every returning user goes through this screen, a mistyped email is common, and the fix is small (show the error inline, the same way "\*Invalid email address" is already shown). |
| **Environment** | https://staging.ablespace.io/signin · tested 13 Oct 2026 |
| **Browser / Device** | Chromium 141.0.7390.37, desktop viewport 1366 × 900; also 390 × 844 mobile emulation. _Add your own Chrome version and OS after reproducing it yourself._ |
| **Reproducibility** | Always (3/3 attempts) |
| **Preconditions** | Logged out. |
| **Steps to Reproduce** | 1. Open https://staging.ablespace.io/signin.<br>2. In **Email**, type a well-formed address that has no account, e.g. `qa.fictional.unregistered@example.com`.<br>3. Click **Continue**.<br>4. Keep watching the screen for 6 seconds without clicking anything. |
| **Expected Result** | A clear error appears **next to the Email field**, in the same style as the existing "\*Invalid email address" message, and **stays until the user edits the field**. For example: "No account found for this email. Check for typos or create an account", with a link to `/signup`. |
| **Actual Result** | A red toast "User not found!" appears in the **top-right corner** of the page, on the opposite side of the screen from the form, and **disappears on its own after about 5 seconds**. Nothing changes near the Email field. There is no link to create an account. At 390 px wide the toast is a full-width banner that covers the AbleSpace logo and header. A user who looks away (for example, to check which email they used) sees no feedback and may think **Continue** did nothing. |
| **Impact** | Returning users who mistype their email get no lasting explanation. New users who went to "Log in" by mistake are not pointed to "Create account". Error styles are also inconsistent on one screen: format errors are inline, while "not found" is a disappearing toast. |
| **Evidence** | Desktop toast visible at 1 s: [03-signin-user-not-found-toast-desktop.png](evidence/public-pages/03-signin-user-not-found-toast-desktop.png)<br>Same screen at 6 s, toast gone: [04-signin-toast-gone-after-6s.png](evidence/public-pages/04-signin-toast-gone-after-6s.png)<br>Mobile 390 px, toast over the header: [05-signin-user-not-found-toast-mobile-390.png](evidence/public-pages/05-signin-user-not-found-toast-mobile-390.png)<br>For comparison, the inline format error: [02-signin-invalid-email.png](evidence/public-pages/02-signin-invalid-email.png) |

---

## Other minor issues (one line each)

| ID | Issue | Where |
|---|---|---|
| M-01 | The email placeholder is "James.musk@email.com" on Sign-in but "user@example.com" on Sign-up. These are similar screens, so the copy is inconsistent; a neutral example address on both would be cleaner. | /signin vs /signup |
| M-02 | When a logged-out user opens a protected page, the return URL is double-encoded (`/signin?redirectUri=%252Fcaseload`; `%25` is an encoded `%`). _TODO: after logging in, check whether you land on /caseload. If yes, this is cosmetic. If you land on the wrong page or a 404, promote it to a full bug report._ | /caseload, /data, /goals while logged out |
| M-03 | _TODO_ | |
