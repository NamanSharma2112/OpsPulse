# Section 6: Test Summary

_Draft: replace every placeholder in angle brackets and _TODO_ with your real numbers before you submit. Keep it to about half a page._

**Testing performed.** I tested AbleSpace staging (https://staging.ablespace.io) over about _<X>_ hours, from _<date>_ to _<date>_, using one manually created account and fictional data only. I focused on the teacher's daily path: **sign in → caseload → add, search and edit a student → add an Accuracy goal → take data → check it in the Data tab**. I ran 10 exploratory scenarios (Section 2) and 8 detailed test cases for the student workflow (Section 3): _<n>_ passed and _<n>_ failed. I checked compatibility on Google Chrome _<ver>_, _<Edge/Firefox ver>_, the AbleSpace mobile app on _<device>_, and Chrome device emulation at 390 × 844, iPad Mini and iPad Pro (Section 5). I also automated sign-in validation and the "add a student" flow with Playwright (Section 7).

**Issues found.** I found _<n>_ bugs, written up in full in Section 4, plus _<n>_ minor issues: _<n>_ High, _<n>_ Medium, _<n>_ Low. **I would fix _<BUG-0X>_ first**, because _<reason tied to user impact, e.g. "it silently records the wrong accuracy percentage, and that number goes into IEP progress reports that are shared with parents and schools. Teachers trust the number, so the error can go unnoticed for weeks.">_ For comparison, BUG-03 (the sign-in error that disappears) is low severity but cheap to fix, so it's a good quick win.

**Areas not tested, and why.**
- **Security, penetration and load testing**: excluded by the assignment rules.
- **Sign-up automation and multiple accounts**: rules allow one manual account only, so I didn't test multi-user or team-sharing scenarios (for example, two teachers editing the same student).
- **Google, Microsoft, Clever and ClassLink sign-in**: needs real third-party school accounts.
- **IEP document upload (AI import)**: I would need a realistic IEP file, and a real one would contain real student data.
- **Billing and subscription, and email delivery** (password reset emails): outside a 4–5 hour scope.
- _<anything else you skipped>_

**What I would test next (with 1–2 more hours).**
1. **The data-accuracy edge cases end to end**: logged data → goal graph → generated progress report, using a mix of Accuracy, Frequency and Duration goals. Report numbers must match the raw data exactly, because these documents carry legal weight.
2. **Dates and time zones**: log data at 11:30 PM and change the computer's time zone, then check that the session doesn't move to a different day in the graph or report.
3. **Concurrency and offline use**: the same student open in two tabs, edited in both; and taking data while the network drops (DevTools → Offline). Teachers often take data in classrooms with poor Wi-Fi, so data must not be lost silently.
4. **Accessibility basics**: keyboard-only use of Take Data, visible focus, and screen-reader labels on the "+ / − / P" buttons. Many users are educators of students with disabilities, so accessibility matters for this audience.
