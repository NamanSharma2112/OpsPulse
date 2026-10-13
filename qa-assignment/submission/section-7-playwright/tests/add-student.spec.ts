import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/login-page';

const email = process.env.ABLESPACE_EMAIL;
const password = process.env.ABLESPACE_PASSWORD;

// Flow: Login -> Add Student (manually) -> verify it appears in the caseload
// and survives a page refresh.
test.describe('Caseload: add a student', () => {
  test.skip(!email || !password, 'Set ABLESPACE_EMAIL and ABLESPACE_PASSWORD in .env');

  test.beforeEach(async ({ page }) => {
    await new LoginPage(page).login(email!, password!);
  });

  test('a new student appears in the caseload and survives a refresh', async ({ page }) => {
    // Unique, clearly fictional name so reruns never collide.
    const firstName = 'QA-Auto';
    const lastName = `Test${Date.now()}`;
    const fullName = `${firstName} ${lastName}`;

    await test.step('open the caseload', async () => {
      await page.goto('/caseload');
      await expect(page.getByRole('button', { name: /add student/i })).toBeVisible();
    });

    await test.step('add the student manually', async () => {
      await page.getByRole('button', { name: /add student/i }).click();
      await page.getByText(/add manually/i).click();

      const form = page.getByRole('dialog');
      await form.getByLabel(/first name/i).fill(firstName);
      await form.getByLabel(/last name/i).fill(lastName);
      await form.getByRole('button', { name: /^create$/i }).click();

      await expect(form).toBeHidden();
    });

    await test.step('the student is listed in the caseload', async () => {
      await expect(page.getByText(fullName)).toBeVisible();
    });

    await test.step('the student is still there after a refresh', async () => {
      await page.reload();
      await expect(page.getByText(fullName)).toBeVisible();
      // Exactly one card: a double-submit would have created a duplicate.
      await expect(page.getByText(fullName)).toHaveCount(1);
    });
  });
});
