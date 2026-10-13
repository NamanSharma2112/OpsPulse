import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page object for https://staging.ablespace.io/signin.
 *
 * Sign-in is a two-step form: enter the email and press Continue; if the
 * account exists, a password step follows.
 */
export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly continueButton: Locator;
  readonly passwordInput: Locator;
  readonly submitPasswordButton: Locator;

  constructor(page: Page) {
    this.page = page;
    // <label for="email">Email</label> is linked to the input, so the label works.
    this.emailInput = page.getByLabel('Email');
    // exact: true so it doesn't also match "Continue with Google" and friends.
    this.continueButton = page.getByRole('button', { name: 'Continue', exact: true });
    // Password step (only shown for a registered email).
    this.passwordInput = page.getByLabel(/^password/i);
    this.submitPasswordButton = page.getByRole('button', { name: /^(log ?in|sign ?in|continue)$/i });
  }

  async goto() {
    await this.page.goto('/signin');
    await expect(this.emailInput).toBeVisible();
  }

  async submitEmail(email: string) {
    await this.emailInput.fill(email);
    await this.continueButton.click();
  }

  /** Full sign-in with an email + password account. */
  async login(email: string, password: string) {
    await this.goto();
    await this.submitEmail(email);
    await expect(this.passwordInput).toBeVisible();
    await this.passwordInput.fill(password);
    await this.submitPasswordButton.click();
    // Signed in once the app leaves /signin.
    await expect(this.page).not.toHaveURL(/\/signin/, { timeout: 30_000 });
  }
}
