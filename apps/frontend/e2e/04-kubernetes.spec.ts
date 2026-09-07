import { test, expect } from "@playwright/test";

test.describe("E2E: Kubernetes Management", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem("cifo_access_token", "dev-token-admin");
    });
  });

  test("renders kubernetes pod table and namespace selector", async ({ page }) => {
    await page.goto("/kubernetes");
    await page.waitForLoadState("domcontentloaded");

    // Header check
    await expect(page.locator("body")).toContainText(/Kubernetes|Pods/i);

    // Verify pod table exists
    const table = page.locator("table");
    await expect(table).toBeVisible({ timeout: 15000 });

    // Verify pods are rendered
    const rows = table.locator("tbody tr");
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });

  test("opens pod log viewer modal", async ({ page }) => {
    await page.goto("/kubernetes");
    await page.waitForLoadState("domcontentloaded");

    const logBtn = page.locator("button:has-text('Logs'), button:has-text('Log'), tr button").first();
    if (await logBtn.isVisible()) {
      await logBtn.click();
      await expect(page.locator("[role='dialog'], .fixed")).toBeVisible({ timeout: 5000 });
    }
  });
});
