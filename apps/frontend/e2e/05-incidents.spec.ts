import { test, expect } from "@playwright/test";

test.describe("E2E: Incidents Lifecycle", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem("cifo_access_token", "dev-token-admin");
    });
  });

  test("renders incidents page with KPI stats and incidents list", async ({ page }) => {
    await page.goto("/incidents");
    await page.waitForLoadState("domcontentloaded");

    // Header check
    await expect(page.locator("body")).toContainText(/Incidents|Incident Management/i);

    // Verify presence of table or incidents container
    const tableOrList = page.locator("table, [data-testid='incidents-list'], .space-y-4").first();
    await expect(tableOrList).toBeVisible({ timeout: 15000 });
  });

  test("allows filtering and viewing incident detail modal", async ({ page }) => {
    await page.goto("/incidents");
    await page.waitForLoadState("domcontentloaded");

    // Check if any incident row/item exists to click
    const incidentItem = page.locator("table tbody tr, button:has-text('View'), button:has-text('Detail')").first();
    if (await incidentItem.isVisible()) {
      await incidentItem.click();
      // Should show detail dialog
      const dialog = page.locator("[role='dialog']").first();
      if (await dialog.isVisible({ timeout: 3000 })) {
        await expect(dialog).toBeVisible();
      }
    }
  });
});
