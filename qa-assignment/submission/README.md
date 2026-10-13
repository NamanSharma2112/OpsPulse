# AbleSpace: QA Assignment (Software Tester)

**Candidate:** _<your name>_
**Application under test:** https://staging.ablespace.io
**Testing window:** _<start date> to <end date>_
**Account used:** one account created manually through the normal sign-up flow. All student and goal data is fictional.

## Where to find each section

| Section | Document |
|---|---|
| 1. Application access | Account created at staging.ablespace.io (see "Environment" below) |
| 2. Exploratory testing | [section-2-test-scenarios.md](section-2-test-scenarios.md) |
| 3. Detailed test cases | [section-3-test-cases.md](section-3-test-cases.md) |
| 4. Bug reports | [section-4-bug-reports.md](section-4-bug-reports.md) |
| 5. Cross-browser and responsive | [section-5-compatibility-report.md](section-5-compatibility-report.md) |
| 6. Test summary | [section-6-test-summary.md](section-6-test-summary.md) |
| 7. Playwright automation (bonus) | [section-7-playwright/](section-7-playwright/README.md) |
| Evidence (screenshots and recordings) | [evidence/](evidence/) |

## Environment

| Item | Value |
|---|---|
| URL | https://staging.ablespace.io |
| Desktop browser 1 | Google Chrome _<version from chrome://version>_ on _<OS>_ |
| Desktop browser 2 | _Microsoft Edge or Firefox <version>_ on _<OS>_ |
| Mobile | AbleSpace app _<version>_ from _Play Store or App Store_ on _<device and OS>_ |
| Tablet emulation | Chrome DevTools device mode: iPad Mini (768 × 1024) and iPad Pro (1024 × 1366) |
| Phone emulation | Chrome DevTools device mode: 390 × 844 |

## Rules I followed

- I created only one account, by hand. I never automated sign-up.
- All data was fictional. Student names follow the pattern `QA-<Name> Test<NN>`.
- I did no security, penetration, injection or load testing. "Special characters" tests used ordinary characters that real names contain (apostrophes, hyphens, accented letters).
- Credentials are kept out of the code. The Playwright project reads them from a git-ignored `.env` file.
