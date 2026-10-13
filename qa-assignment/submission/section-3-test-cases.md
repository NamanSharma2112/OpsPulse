# Section 3: Detailed Test Cases

## Workflow chosen

**Add Student → Save → View in Caseload → Search → Edit Student**

**Why this workflow:** every goal, data point and progress report belongs to a student, so a student record that is wrong, duplicated or not saved breaks everything that follows. Adding a student is also the first thing a new teacher does after signing up.

**Techniques used:** equivalence partitioning (valid vs invalid names), boundary value analysis (name length), error guessing (spaces, apostrophes, duplicates), and a state check (data must survive a page refresh).

### Common preconditions (all test cases)

- Logged in to https://staging.ablespace.io with my test account (email and password).
- On the **Caseload** tab.
- Test data is fictional. Each case uses a unique last name (`Test01`, `Test02`, and so on) so results don't overlap.

### Summary

| ID | Title | Type | Status |
|---|---|---|---|
| TC-01 | Create a student with only the required fields | Positive | _TODO_ |
| TC-02 | Create a student with all fields; data survives a refresh | Positive | _TODO_ |
| TC-03 | Required fields left empty | Negative | _TODO_ |
| TC-04 | Names that are only spaces, or padded with spaces | Negative / Edge | _TODO_ |
| TC-05 | Name length boundaries (1 character and very long) | Boundary | _TODO_ |
| TC-06 | Real-world special characters in names | Edge | _TODO_ |
| TC-07 | Search the caseload: partial, mixed case, no results | Positive / Negative | _TODO_ |
| TC-08 | Edit a student: Save persists, Cancel discards | Positive | _TODO_ |

---

## TC-01: Create a student with only the required fields

| Field | Description |
|---|---|
| **Test Case ID** | TC-01 |
| **Title** | Create a student with first and last name only, and confirm it appears in the caseload |
| **Type** | Positive (happy path) |
| **Preconditions** | Common preconditions. No student named "QA-Mira Test01" exists. |
| **Test Data** | First name: `QA-Mira` · Last name: `Test01` · All other fields left blank |
| **Steps** | 1. On **Caseload**, click **Add Student**.<br>2. Choose **Add Manually**.<br>3. Enter First name `QA-Mira`.<br>4. Enter Last name `Test01`.<br>5. Leave every other field blank.<br>6. Click **Create**.<br>7. Look at the caseload list or cards. |
| **Expected Result** | The dialog closes and a success message (or the new card) appears. "QA-Mira Test01" shows in the caseload **exactly once**. No error appears for the blank optional fields. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO: Pass / Fail_ |

## TC-02: Create a student with all fields; data survives a refresh

| Field | Description |
|---|---|
| **Test Case ID** | TC-02 |
| **Title** | Every field entered while creating a student is saved and shown correctly after a page refresh |
| **Type** | Positive (data integrity) |
| **Preconditions** | Common preconditions. |
| **Test Data** | First name: `QA-Leo` · Last name: `Test02` · School: `Fictional Elementary` · Grade: `3` (or "3rd") · Date of birth: `15/06/2017` (use the app's date format) · Teacher: `Ms. Example` · Avatar colour: any non-default colour |
| **Steps** | 1. Click **Add Student**, then **Add Manually**.<br>2. Fill in every field with the test data.<br>3. Click **Create**.<br>4. Open the new student's profile (three-dot menu, then **Edit Student**, or click the card).<br>5. Note each field's value.<br>6. Press **F5** to refresh and open the profile again. |
| **Expected Result** | Every field shows exactly what was entered, before **and** after the refresh. The date of birth is not shifted by one day (a common timezone bug) and the grade is not changed. The avatar colour is the one selected. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |

## TC-03: Required fields left empty

| Field | Description |
|---|---|
| **Test Case ID** | TC-03 |
| **Title** | A student can't be created without a first name or last name |
| **Type** | Negative |
| **Preconditions** | Common preconditions. Count the students in the caseload before starting. |
| **Test Data** | Run A: First name empty, Last name `Test03`<br>Run B: First name `QA-Ava`, Last name empty<br>Run C: both empty |
| **Steps** | 1. Click **Add Student**, then **Add Manually**.<br>2. Enter the Run A data and click **Create**.<br>3. Note what happens, fix the field, then clear the other one (Run B) and click **Create**.<br>4. Clear both (Run C) and click **Create**.<br>5. Close the dialog and count the students again. |
| **Expected Result** | For each run, the dialog stays open and a clear inline message appears next to the empty field (for example "First name is required"). **No student is created**: the count is unchanged and there is no card with a blank name. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |

## TC-04: Names that are only spaces, or padded with spaces

| Field | Description |
|---|---|
| **Test Case ID** | TC-04 |
| **Title** | Whitespace-only names are rejected, and spaces around a valid name are trimmed |
| **Type** | Negative / Edge |
| **Preconditions** | Common preconditions. |
| **Test Data** | Run A: First name `"   "` (3 spaces), Last name `"   "`<br>Run B: First name `"  QA-Noah  "`, Last name `"  Test04  "` (2 spaces before and after) |
| **Steps** | 1. **Add Student**, then **Add Manually**. Enter the Run A data and click **Create**.<br>2. Note the result.<br>3. Enter the Run B data and click **Create**.<br>4. Search the caseload for `QA-Noah`. |
| **Expected Result** | Run A is treated the same as empty: an error appears and no blank-looking student is created. Run B saves as "QA-Noah Test04" with no extra spaces, and searching for `QA-Noah` finds it. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |

## TC-05: Name length boundaries

| Field | Description |
|---|---|
| **Test Case ID** | TC-05 |
| **Title** | Very short and very long names are handled without errors or a broken layout |
| **Type** | Boundary value analysis |
| **Preconditions** | Common preconditions. |
| **Test Data** | Run A: First name `Q` (1 character), Last name `Test05a`<br>Run B: First name of **256 characters** (for example `QA` followed by 254 × `x`; any counter tool works), Last name `Test05b` |
| **Steps** | 1. Create the Run A student and confirm it saves.<br>2. Start creating the Run B student. Paste the 256-character first name and click **Create**.<br>3. If it saves, look at the caseload card, the profile, and the student selector in **Take Data**.<br>4. Repeat step 3 at 390 px wide (DevTools device mode). |
| **Expected Result** | The 1-character name saves. For 256 characters, **either** a maximum length is enforced with a clear message (for example "Max 50 characters"), **or** the name saves and is shown truncated with "…" without overlapping other elements or causing horizontal scrolling. There is never a silent failure or a generic "Something went wrong". |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |

## TC-06: Real-world special characters in names

| Field | Description |
|---|---|
| **Test Case ID** | TC-06 |
| **Title** | Names with apostrophes, hyphens and accented letters save, display and search correctly |
| **Type** | Edge (error guessing) |
| **Preconditions** | Common preconditions. |
| **Test Data** | Student 1: First `Anne-Marie`, Last `O'Brien-Test06`<br>Student 2: First `José`, Last `Núñez-Test06` |
| **Steps** | 1. Create Student 1 and Student 2 with **Add Manually**.<br>2. Check both names in the caseload and on each profile.<br>3. Search for `o'brien`.<br>4. Search for `jose` (no accent) and then `josé`.<br>5. Refresh and repeat step 2. |
| **Expected Result** | Both students save. Names display exactly as typed, with no `&#39;`, `JosÃ©` or other garbled characters, before and after a refresh. Searching `o'brien` finds Student 1. Searching `josé` finds Student 2. Note whether `jose` without the accent also matches; that's an observation, not a requirement. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |

## TC-07: Search the caseload

| Field | Description |
|---|---|
| **Test Case ID** | TC-07 |
| **Title** | Caseload search matches partial and mixed-case names, and shows a clear empty state when nothing matches |
| **Type** | Positive / Negative |
| **Preconditions** | Common preconditions. TC-01 has been run, so "QA-Mira Test01" exists. |
| **Test Data** | `mira` · `MIRA` · `Test0` · `  mira  ` · `zzzz-no-such-student` |
| **Steps** | 1. Type `mira` in the caseload search box.<br>2. Replace it with `MIRA`.<br>3. Replace it with `Test0`.<br>4. Replace it with `  mira  `.<br>5. Replace it with `zzzz-no-such-student`.<br>6. Clear the search box. |
| **Expected Result** | Steps 1, 2 and 4 show QA-Mira Test01 (search ignores case and surrounding spaces). Step 3 shows every Test0x student. Step 5 shows a friendly "No students found" message, not a blank page or spinner. Step 6 brings the full caseload back. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |

## TC-08: Edit a student

| Field | Description |
|---|---|
| **Test Case ID** | TC-08 |
| **Title** | Edits to a student are saved everywhere after Save, and discarded after Cancel |
| **Type** | Positive (data integrity) |
| **Preconditions** | Common preconditions. TC-02 has been run, so "QA-Leo Test02" exists. |
| **Test Data** | Save run: Last name `Test02` → `Test02-Edited`, Grade `3` → `4`<br>Cancel run: First name `QA-Leo` → `SHOULD-NOT-SAVE` |
| **Steps** | 1. On QA-Leo's card, open the three-dot menu and choose **Edit Student**.<br>2. Change the last name and grade as in the test data and click **Save**.<br>3. Check the caseload card and profile.<br>4. Refresh (F5) and check again. Also check the student name in **Take Data** and the **Data** tab.<br>5. Open **Edit Student** again, change the first name to `SHOULD-NOT-SAVE`, and click **Cancel** (or close the dialog).<br>6. Refresh and check the name. |
| **Expected Result** | After Save, "QA-Leo Test02-Edited" with grade 4 shows on the card, the profile, Take Data and the Data tab, both before and after the refresh. No duplicate student is created. After Cancel, the first name is still "QA-Leo" and "SHOULD-NOT-SAVE" appears nowhere. |
| **Actual Result** | _TODO_ |
| **Status** | _TODO_ |
