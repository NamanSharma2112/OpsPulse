# Section 2: Exploratory Testing, Scenarios Tested

## How I chose what to test

AbleSpace is used by special-education teachers and therapists to track IEP goals for students. A teacher's daily loop is:

> **Sign in → open Caseload → pick a student → take data on a goal → review progress or reports**

I ranked areas by **risk = how often a real user touches it × how bad a failure would be**. The worst failures for this product are:
1. **Wrong or lost progress data.** IEP progress reports are formal documents shared with parents and schools.
2. **Student data shown to the wrong person**, or after logout. The product markets itself as FERPA and HIPAA compliant.
3. **Being blocked from the daily loop**: can't sign in, can't find the student, can't save.

The order of the scenarios below follows that priority.

**Legend:** ✅ Pass · ❌ Fail (bug filed) · ⚠️ Pass with an issue (minor or usability)

## Scenarios

| ID | Area | What I checked | Result | Evidence |
|---|---|---|---|---|
| S-01 | Sign-in and sign-up: email validation | Clicked **Continue** with an empty email, then with `not-an-email`. On sign-up, also submitted a malformed email by pressing **Enter**. | ✅ Inline errors "\*Email is required" and "\*Invalid email address" appear under the field and the form is not submitted. The Enter key triggers the same validation, and sign-in and sign-up behave the same way. | [01](evidence/public-pages/01-signin-empty-email.png), [02](evidence/public-pages/02-signin-invalid-email.png) |
| S-02 | Sign-in: unregistered email | Entered a well-formed but unregistered address (`qa.fictional.unregistered@example.com`). | ⚠️ A red "User not found!" toast appears in the top-right corner, far from the form, and **disappears after about 5 s**. There is no inline error and no link to create an account. On a 390 px screen the toast covers the logo. → **BUG-03** | [03](evidence/public-pages/03-signin-user-not-found-toast-desktop.png), [04](evidence/public-pages/04-signin-toast-gone-after-6s.png), [05](evidence/public-pages/05-signin-user-not-found-toast-mobile-390.png) |
| S-03 | Protected pages, deep links and 404 | While logged out, opened `/caseload`, `/data` and `/goals` directly, plus an invalid URL. After logging in, checked whether I was returned to the page I asked for. | ✅ Logged out: each page redirects to `/signin?redirectUri=%252Fcaseload` and the invalid URL shows a proper "Page not found" page. ⚠️ The redirect URI is **double-encoded** (`%252F` instead of `%2F`). _TODO: after logging in, did you land on /caseload or somewhere else?_ | [06](evidence/public-pages/06-caseload-logged-out-redirects-to-signin.png), [07](evidence/public-pages/07-404-page.png) |
| S-04 | Session: refresh, logout, Back button | Refreshed (F5) on Caseload and on a student profile. Logged out, then pressed the browser **Back** button. Logged out in tab A and used tab B. | _TODO_. Expect: refresh keeps you on the same page and logged in. After logout, Back must **not** show any student data. | |
| S-05 | Add Student: field validation | Saved with first or last name empty, with names of only spaces, with 1 character, with 100+ characters, with real-name special characters (`O'Brien`, `Anne-Marie`, `José`), with a date of birth in the future, and with a duplicate of an existing name. | _TODO_ | |
| S-06 | Edit Student: data consistency | Edited the name and grade, saved, refreshed, then checked the caseload card, the profile, the Data tab and the goal screens. Also edited a field, clicked **Cancel**, and checked that nothing changed. | _TODO_. Expect: the new value shows everywhere after a refresh, and Cancel discards the edit. | |
| S-07 | Caseload search and filter | Searched by partial name, in a different case (`mira` vs `Mira`), with leading or trailing spaces, with an apostrophe, and with no matches. Cleared the search. | _TODO_. Expect: matches regardless of case, a clear "no results" empty state, and the full list back once cleared. | |
| S-08 | Add Goal (Accuracy data type) | Saved with empty goal text, then with very long goal text (2,000+ characters). Typed goal text, switched the data type, and checked whether the text survived. Added an objective under the goal. | _TODO_ | |
| S-09 | Take Data: accuracy calculation | Recorded 7 × "+" and 3 × "−", expecting **70 %**. Checked how "P" (prompted) is counted. Tried saving with **0 trials** (no `NaN`, `Infinity` or crash). Double-clicked **Save** to check for duplicates. Refreshed and checked that the data appears in the graph and Data tab. | _TODO_. This is the highest-value check because the numbers feed IEP progress reports. | |
| S-10 | Data tab: Log Data and delete | Clicked **Log Data** and saved a row with no student or goal selected. Used duplicate and remove on rows. Deleted a data point and checked that the graph and accuracy update. Left the page with an unsaved draft. | _TODO_. Expect: required-field errors, the graph updates after a delete, and a warning before unsaved data is lost. | |

## Notes from exploration

_Write short notes here while you explore: anything odd, slow or confusing, even if it isn't a bug yet. Good bug reports usually start as a line in this list._

- Sign-in is a two-step flow: email, then **Continue**, then the password step. Accounts created with "Continue with Google", Microsoft, Clever or Classlink have no password step.
- Desktop and mobile show the same "User not found!" error in different places: top-right toast on desktop, full-width banner over the header on mobile.
