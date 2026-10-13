# Study Guide: AbleSpace QA Assignment

> **This file is for you, not the reviewers. Don't submit it.** Submit only the `submission/` folder.

---

## 0. What's already done and what's left for you

| Done for you | Your job (needs your account and your own eyes) |
|---|---|
| The structure of every section, in the exact formats the assignment asks for | **Create your one account** at staging.ablespace.io. Use email + password, not Google, so the automation can log in. |
| Real checks on the **public** pages (sign-in, sign-up, 404, redirects) at desktop, 390 px, iPad Mini and iPad Pro, with screenshots in `submission/evidence/public-pages/` | **Re-run those checks yourself** in your own Chrome, so you've seen them with your own eyes and can demo them in the interview. |
| 10 exploratory scenarios (3 already have results) | Explore **after login** and fill in every `_TODO_` result. |
| 8 detailed test cases for the student workflow | **Run them** and fill in Actual Result and Pass/Fail. |
| BUG-03, a real usability bug, written in full | **Find and write BUG-01 and BUG-02.** Post-login bugs are usually more serious; see the bug-hunting list in section 4 of this guide. |
| Compatibility matrix with the public-page results filled in | Test the logged-in app on Chrome, Edge or Firefox, the mobile app, and the emulators. |
| Test summary draft | Put in your real numbers and pick the bug you would fix first. |
| A working Playwright project: 4 tests pass against staging, and the login flow is written | Add your credentials to `.env`, run it, and fix any post-login locator that doesn't match (see section 6). |

**Important:** the reviewers will ask you to walk them through one bug and one test case. Only submit what you have personally seen and can explain. If a result in these documents doesn't match what you see, **change it to what you see.**

---

## 1. Understand the product (5 minutes)

**AbleSpace** is an app for **special-education teachers and therapists** (speech, occupational and physical therapists). Each student with a disability in the US has an **IEP** (Individualized Education Program), a legal document listing goals such as "Mira will read 20 sight words with 80 % accuracy".

Teachers use AbleSpace to:
1. Keep a **caseload**, their list of students.
2. Add **goals** to each student, with a **data type** that sets how progress is measured: Accuracy, Frequency, Duration, Prompting Levels, Task Analysis, Rating Scale, and so on.
3. **Take data** during sessions. For an Accuracy goal you press **+** (correct), **−** (incorrect) or **P** (prompted, meaning correct with help), and the app calculates the %.
4. See **graphs** and generate **progress reports** for parents and schools.

**Why this matters for testing:** the most dangerous bugs are **wrong numbers** (a wrong accuracy %, a lost data point, a date on the wrong day) and **privacy leaks** (student data visible after logout). The product advertises FERPA and HIPAA compliance. Say this in the interview; it shows you think about the user, not just the buttons.

---

## 2. QA concepts you must be able to explain

### Test scenario vs test case
- **Scenario** = *what* to test, in one line. "Check that adding a student validates the name fields." (Section 2)
- **Test case** = *exactly how*: preconditions, data, steps, expected result. Anyone can follow it and get the same result. (Section 3)

### Positive, negative, edge and boundary
- **Positive (happy path):** valid input, should succeed. *TC-01: create a student with a valid name.*
- **Negative:** invalid input, should be **rejected with a clear message**. *TC-03: empty first name.*
- **Edge:** unusual but valid input. *TC-06: `O'Brien`, `José`.*
- **Boundary:** values at the limits. *TC-05: 1 character vs 256 characters.*

### Test design techniques (name these in the interview)
- **Equivalence partitioning:** split inputs into groups that should behave the same, and test one from each. Names: valid / empty / only spaces / too long.
- **Boundary value analysis:** bugs live at edges. If the max is 50, test 49, 50 and 51.
- **Error guessing:** use experience to guess what breaks: apostrophes, double-clicks, refresh mid-form, the Back button after logout.

### Exploratory testing
Testing while learning the app, without a script. Give yourself a mission (a "charter"), for example *"Explore Take Data to find calculation errors"*, for 20–30 minutes, and take notes. Useful questions to ask at every screen:
- **CRUD:** can I Create, Read, Update and Delete it? Does each change show up everywhere?
- **Refresh, Back, twice:** what happens if I refresh, press Back, or click Save twice?
- **Empty, long, weird:** empty field, 300 characters, spaces only, `O'Brien`, emoji.
- **Interrupt:** close the dialog half-way, or go offline (DevTools → Network → Offline).
- **Consistency:** does the same thing look and behave the same on every screen?

### Severity vs priority (very likely interview question)
- **Severity** = technical impact. How badly is the product broken?
- **Priority** = business urgency. How soon should it be fixed?

| Example | Severity | Priority |
|---|---|---|
| Accuracy % is calculated wrongly | **Critical / High** (wrong legal data) | **High** |
| App crashes when a student has 300 goals (rare) | High | Low/Medium (rare) |
| The company logo is misspelled on the login page | Low | **High** (everyone sees it, brand) |
| Error toast disappears too fast (BUG-03) | Low | Medium (every user, cheap fix) |

Severity scale: **Critical** = data loss, crash, or a core feature with no workaround · **High** = a major feature broken, workaround is hard · **Medium** = a feature partly wrong, workaround exists · **Low** = cosmetic or a minor usability issue.

### A good bug report
**One bug per report.** It needs a specific title ("Accuracy shows 0 % when 'P' trials are recorded", not "Data bug"), steps from a clear starting point, expected vs actual, environment, reproducibility, and evidence. The test is whether a developer can reproduce it **without asking you a single question.**

Reproducibility: try it **3 times**. If it's 3/3, write "Always (3/3)". If it's 1/3, write "Sometimes (1/3)" and note anything different between the attempts.

### Other terms they may ask about
- **Smoke test:** a quick check that the main things work at all (can log in, can open the caseload).
- **Sanity test:** a quick, narrow check after a small fix.
- **Retesting:** checking that a specific fixed bug is really fixed.
- **Regression testing:** checking that the fix didn't break anything else.
- **Verification vs validation:** are we building it right (meets the spec)? vs are we building the right thing (meets the user's need)?
- **Cross-browser testing:** same app, different browsers (Chrome, Edge, Firefox, Safari).
- **Responsive testing:** same app, different screen sizes (phone, tablet, desktop).

---

## 3. Step-by-step plan (about 4–5 hours)

### Step 1: Setup (15 min)
1. Sign up at https://staging.ablespace.io/signup **with email + password**. Use fictional details everywhere.
2. Write down your Chrome version (open `chrome://version`) and your OS. You need these for every bug report.
3. Make a Google Drive folder for evidence screenshots and recordings.

### Step 2: Re-check what's already filled in (15 min)
Open the sign-in page and repeat S-01, S-02 and S-03 and BUG-03 yourself. Take your **own** screenshots (Windows: `Win + Shift + S`; Mac: `Cmd + Shift + 4`). Replace the browser line in BUG-03 with your Chrome version.

For S-03, while logged out, open `https://staging.ablespace.io/caseload` → you are sent to sign-in → **log in** → **do you land on /caseload?** Write the answer in S-03 and M-02. If you land somewhere wrong, that's a real bug: promote M-02 to BUG-01 or BUG-02.

### Step 3: Explore after login (60–90 min) → Section 2
Go through S-04 to S-10. Keep a notes file open and write down **everything** odd. Screenshot or record anything that looks like a bug *immediately*, because some bugs are hard to reproduce later.

### Step 4: Run the test cases (45 min) → Section 3
Follow TC-01 to TC-08 exactly. If a field name or button differs from what's written (for example "Save" instead of "Create"), **update the steps** to match the real app. Then fill in **Actual Result** (what you saw, factually) and **Status**.

### Step 5: Write bugs (45 min) → Section 4
Pick your **2 most important** findings for BUG-01 and BUG-02. Use the template. Put the rest as one-liners under "Other minor issues".

### Step 6: Compatibility (45 min) → Section 5
- Repeat the main flows in **Edge or Firefox**.
- Install the **AbleSpace app** from the Play Store or App Store and repeat sign-in, the caseload, and taking data.
- In Chrome, press `F12`, then `Ctrl + Shift + M` (device mode). Choose **iPad Mini** and **iPad Pro** from the device list, then set a custom size of **390 × 844**. Look for horizontal scrollbars, cut-off buttons and overlapping text.

### Step 7: Summary (15 min) → Section 6
Fill in the numbers. Choose the bug you'd fix first and explain why **in terms of user impact.**

### Step 8: Playwright (optional, 30–60 min) → Section 7
See section 6 of this guide.

### Step 9: Submit
See section 7 of this guide.

---

## 4. Where bugs usually hide in an app like this (check these first)

These are **places to look**, not confirmed bugs. Only report what you actually see.

**Take Data / Accuracy (highest value)**
- 7 "+" and 3 "−" should give 70 %. Is it right?
- How is **P** counted: correct, incorrect, or left out? Is that explained anywhere? Is it consistent between the graph and the Data tab?
- Save with **zero** trials. Do you see `NaN %`, `Infinity`, a crash, or a sensible message?
- **Double-click Save.** Do you get two sessions or data points?
- Record data, **refresh**. Is it still there? Is it in the graph?
- Undo the last trial. Does the % recalculate?

**Students**
- Create two students with the **same name**. Is there a warning? Can you tell them apart?
- Edit a name. Does it change in **every** place (card, profile, Take Data, Data tab, graphs)?
- Very long names: does the card layout break?

**Dates**
- Date of birth in the future: is it accepted?
- Log a session late at night (11:30 PM). Does the graph show the right day?

**Session and navigation**
- Log out, then press **Back**. Can you still see student data? (Privacy: report this as **High**.)
- Refresh on a deep page (a student profile). Do you stay there, or get thrown to the home page?
- Open the app in two tabs, log out in one, then click around in the other.

**Forms in general**
- Close a half-filled dialog. Is there an "unsaved changes" warning, or is the data silently lost?
- Do required fields show **which** field is wrong?
- Is the error message understandable, or a generic "Something went wrong"?

---

## 5. Capturing evidence

- **Screenshot** the bug in its broken state with the URL bar visible.
- **Screen recording** for anything with several steps or timing (like the disappearing toast): use Loom (free), the Windows Game Bar (`Win + G`), or the Mac screen recorder (`Cmd + Shift + 5`).
- Upload to Google Drive. Set sharing to **"Anyone with the link can view"** and paste the link into the Evidence field. **Test the link in an incognito window.**
- Or put the files in `submission/evidence/` and link them with relative paths, as BUG-03 does.

---

## 6. The Playwright part, explained

### What Playwright is
A tool that controls a real browser from code: it opens pages, clicks, types, and **asserts** (checks) that the page shows what it should. QA teams use it for **regression tests**, so the same checks run automatically before every release.

### Key ideas in this code (be ready to explain each)
1. **Locators** find elements. We use the ones a real user would recognise:
   - `page.getByRole('button', { name: 'Continue', exact: true })` finds the button whose visible name is Continue. `exact: true` stops it matching "Continue with Google".
   - `page.getByLabel('Email')` finds the input linked to the "Email" label.
   - These are **stable**: they keep working if a developer changes CSS classes, and they also confirm the page is accessible.
2. **Assertions** check the result: `await expect(locator).toBeVisible()`, `toHaveURL(...)`, `toHaveCount(1)`.
3. **Auto-waiting:** `expect(...)` retries until the condition is true or the timeout expires. That's why there are **no `sleep` / `waitForTimeout` calls**. Fixed sleeps make tests slow and flaky.
4. **Page Object Model** (`pages/login-page.ts`): the sign-in locators and steps live in one class. If the login page changes, you fix one file instead of every test.
5. **Credentials from environment variables:** `process.env.ABLESPACE_EMAIL`, loaded from `.env` by `dotenv`. `.env` is in `.gitignore`, so your password never reaches GitHub.
6. **`test.step`** labels each stage in the report, so a failure tells you *which* stage broke.
7. **Unique data:** `Test${Date.now()}` makes a new name each run, so a rerun never finds an old student by mistake.
8. **`workers: 1`, `retries: 0`:** this is a live system and we must not send lots of requests.

### Running it
```bash
cd submission/section-7-playwright
npm install
npx playwright install chromium
cp .env.example .env     # then put your email and password in .env
npm test
```

### If the add-student test fails on a locator
The sign-in tests were run and pass. The post-login steps (`Add Student`, `Add Manually`, `First name`, `Create`) use the wording from AbleSpace's help centre, but **they couldn't be checked without your account**. If one doesn't match:
1. Run `npx playwright codegen https://staging.ablespace.io`. A browser opens with a recorder.
2. Log in and click through **Add Student → Add Manually → type the names → Create**.
3. The recorder shows the locator for each click (for example `getByRole('button', { name: 'Save' })`). Copy the one you need into `tests/add-student.spec.ts`.
4. Run `npm run test:headed` to watch it pass.

That loop (record, read the locator, choose a stable one, assert) is exactly what an interviewer wants to hear you describe.

---

## 7. Submitting

1. Delete this `STUDY-GUIDE.md` from what you submit (or simply submit only the `submission/` folder).
2. Search the folder for `TODO` and placeholders in angle brackets (`<...>`). Make sure none are left.
3. Check that **no password** appears anywhere (`.env` must not be included).
4. Choose one:
   - **GitHub (recommended):** create a new public repository, e.g. `ablespace-qa-assignment`, and upload the contents of `submission/`. GitHub renders the `.md` tables nicely. Don't upload `node_modules/`; the `.gitignore` already excludes it.
   - **ZIP:** zip `submission/` without `node_modules/`.
   - **Google Drive:** upload the folder and set sharing to "Anyone with the link".
5. Submit the link on the Google Form: https://forms.gle/BGMAaXhXVPyZU7Sd9

---

## 8. Interview preparation

### Walking through a bug (practise this out loud, about 1 minute)
Use **Context → Steps → Expected vs actual → Impact → Severity and priority reasoning**:

> "On the sign-in page, I tried an email that isn't registered, which is a common real-life mistake. I clicked Continue and expected a clear error next to the email field, like the one shown for a badly formatted email. Instead, a red 'User not found!' toast appeared in the top-right corner, away from the form, and disappeared by itself after about five seconds. On mobile it covered the header. I reproduced it three times out of three. The impact is that a returning user who looks away, or a new user who clicked 'Log in' by mistake, gets no lasting feedback and no link to create an account. I rated it **Low severity**, because nothing is broken and they can retry, but **Medium priority**, because every user passes through this screen and the fix is small: reuse the existing inline error style."

(If you find a more serious post-login bug, practise that one too. A data-accuracy bug is the strongest story.)

### Walking through a test case
> "I chose the student workflow because every goal, data point and report belongs to a student, so if the student record is wrong, everything after it is wrong. Take TC-08, editing a student. I changed the last name and grade, saved, and checked it in four places: the card, the profile, Take Data and the Data tab. I checked before and after a refresh, because a common bug is that the screen updates but the server doesn't save, or one screen keeps showing old cached data. Then I made a change and pressed Cancel, to make sure Cancel really discards it."

### Likely questions and short answers
- **Why only 8 test cases?** "Quality over quantity. Each one covers a different risk or technique: positive, negative, boundary, special characters, search, and edit consistency."
- **How did you decide what to test first?** "Risk: how often a teacher uses a feature times how bad a failure would be. Data accuracy and privacy come first."
- **What would you automate next?** "Take Data: record known values and assert the calculated % and that the data point survives a refresh. That's high value and repeated every release."
- **What is a flaky test and how do you avoid it?** "A test that sometimes passes and sometimes fails without any code change. Avoid fixed sleeps, use web-first assertions that auto-wait, use stable role and label locators, and use unique test data."
- **Why not use CSS or XPath selectors?** "They depend on page structure and styling, which developers change often. Role and label locators depend on what the user sees, so they're stabler, and they also check accessibility."
- **What didn't you test?** "Security and load (excluded by the rules), third-party logins, IEP upload (it would need real student documents), and multi-user sharing (one account only)."
