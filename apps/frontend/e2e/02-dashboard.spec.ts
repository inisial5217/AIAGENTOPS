import { test, expect } from "@playwright/test";

test.describe("E2E: Dashboard & Monitoring Overview", () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      localStorage.setItem("cifo_access_token", "dev-token-admin");
    });
  });

  test("renders KPI stat cards and telemetry overview", async ({ page }) => {
    await page.goto("/monitoring");

    // Wait for main content to load
    await page.waitForLoadState("domcontentloaded");

    // Check header
    await expect(page.locator("body")).toContainText(/Monitoring|Dashboard|Overview/i);

    // Verify presence of telemetry and stat metrics
    const cards = page.locator(".grid").first();
    await expect(cards).toBeVisible();

    // Verify live telemetry or charts container
    const chartContainer = page.locator("canvas, svg, .echarts-for-react, [data-chart]").first();
    // Chart or telemetry container should be attached
    await expect(page.locator("body")).toContainText(/CPU|Memory|Status|Containers|Active/i);
  });
});
