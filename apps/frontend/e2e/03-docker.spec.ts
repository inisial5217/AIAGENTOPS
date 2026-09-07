import { test, expect } from "@playwright/test";

test.describe("E2E: Docker Management", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem("cifo_access_token", "dev-token-admin");
    });
  });

  test("renders container inventory table and allows filtering", async ({ page }) => {
    await page.goto("/docker");
    await page.waitForLoadState("domcontentloaded");

    // Check page header
    await expect(page.locator("body")).toContainText(/Docker|Containers/i);

    // Verify table exists
    const table = page.locator("table");
    await expect(table).toBeVisible({ timeout: 15000 });

    // Verify rows exist (e.g. cifo-postgres, cifo-vault, cifo-redis, etc.)
    const rows = table.locator("tbody tr");
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);

    // Search filter
    const searchInput = page.locator("input[placeholder*='Search' i], input[type='search']").first();
    if (await searchInput.isVisible()) {
      await searchInput.fill("cifo");
      await page.waitForTimeout(500);
      const filteredCount = await table.locator("tbody tr").count();
      expect(filteredCount).toBeGreaterThan(0);
    }
  });

  test("opens container detail modal upon inspect click", async ({ page }) => {
    await page.goto("/docker");
    await page.waitForLoadState("domcontentloaded");

    const table = page.locator("table");
    await expect(table).toBeVisible({ timeout: 15000 });

    // Click inspect button or container name on first row
    const inspectBtn = page.locator("button:has-text('Inspect'), button:has-text('Detail'), tr button").first();
    if (await inspectBtn.isVisible()) {
      await inspectBtn.click();
      // Check modal appears
      await expect(page.locator("[role='dialog'], .fixed")).toBeVisible({ timeout: 5000 });
    }
  });
});
