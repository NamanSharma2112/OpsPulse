import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/login-page';

// These checks need no account: they cover the email step of sign-in.
test.describe('Sign-in: email validation', () => {
  let login: LoginPage;

  test.beforeEach(async ({ page }) => {
    login = new LoginPage(page);
    await login.goto();
  });

  test('empty email shows "Email is required" and stays on sign-in', async ({ page }) => {
    await login.continueButton.click();

    await expect(page.getByText(/Email is required/)).toBeVisible();
    await expect(page).toHaveURL(/\/signin/);
  });

  test('malformed email shows "Invalid email address"', async ({ page }) => {
    await login.submitEmail('not-an-email');

    await expect(page.getByText(/Invalid email address/)).toBeVisible();
    await expect(page).toHaveURL(/\/signin/);
  });

  test('unregistered email shows "User not found!" and no password step', async ({ page }) => {
    // example.com is reserved for documentation, so no real person owns it.
    await login.submitEmail('qa.fictional.unregistered@example.com');

    await expect(page.getByRole('alert').filter({ hasText: 'User not found!' })).toBeVisible();
    await expect(login.passwordInput).toHaveCount(0);
  });

  test('protected page redirects a logged-out user to sign-in', async ({ page }) => {
    await page.goto('/caseload');

    await expect(page).toHaveURL(/\/signin\?redirectUri=/);
    await expect(page.getByRole('heading', { name: 'Log in to your account' })).toBeVisible();
  });
});
