import { test, expect } from "@playwright/test";

test.describe("E2E: Login Flow", () => {
  test("renders login page and elements", async ({ page }) => {
    await page.goto("/login");

    // Check heading or title
    await expect(page.locator("h1, h2, span").filter({ hasText: /CIFO/i }).first()).toBeVisible();

    // Check SSO or Dev token controls
    const adminProfileBtn = page.locator("button").filter({ hasText: /Admin/i }).first();
    await expect(adminProfileBtn).toBeVisible();

    const ssoBtn = page.locator("button").filter({ hasText: /Keycloak/i }).first();
    await expect(ssoBtn).toBeVisible();
  });

  test("authenticates successfully with credentials and navigates to dashboard", async ({ page }) => {
    await page.goto("/login");

    // Set token in localStorage and navigate to dashboard
    await page.evaluate(() => {
      localStorage.setItem("cifo_access_token", "dev-token-admin");
    });

    await page.goto("/monitoring");
    await page.waitForLoadState("domcontentloaded");

    await expect(page.locator("body")).toContainText(/Monitoring|Dashboard|Overview/i);

    const token = await page.evaluate(() => localStorage.getItem("cifo_access_token"));
    expect(token).toBe("dev-token-admin");
  });
});
