import { describe, it, expect, vi, beforeEach } from "vitest";
import { argocdService } from "./argocd-service";
import { apiClient } from "../lib/api-client";

vi.mock("../lib/api-client", () => {
  const mockClient = {
    get: vi.fn(),
    post: vi.fn(),
  };
  return {
    default: mockClient,
    apiClient: mockClient,
  };
});

describe("argocdService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls getApplications with optional namespace filter", async () => {
    const mockApps = [{ name: "sample-nginx-app", sync_status: "Synced", health_status: "Healthy" }];
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { data: mockApps, total: 1 },
    });

    const result = await argocdService.getApplications("argocd");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/argocd/applications", {
      params: { namespace: "argocd" },
    });
    expect(result.data).toEqual(mockApps);
  });

  it("calls getApplication with name", async () => {
    const mockDetail = {
      name: "sample-nginx-app",
      sync_status: "Synced",
      health_status: "Healthy",
      automated_sync: true,
      resources: [],
      history: [],
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockDetail });

    const result = await argocdService.getApplication("sample-nginx-app");
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/argocd/applications/sample-nginx-app", {
      params: {},
    });
    expect(result.name).toBe("sample-nginx-app");
  });

  it("calls syncApplication with POST options", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "sync triggered" },
    });

    const result = await argocdService.syncApplication("sample-nginx-app", {
      prune: true,
      dry_run: false,
    });
    expect(apiClient.post).toHaveBeenCalledWith(
      "/api/v1/argocd/applications/sample-nginx-app/sync",
      { prune: true, dry_run: false },
      { params: {} }
    );
    expect(result.status).toBe("success");
  });

  it("calls getOverview", async () => {
    const mockOverview = {
      total: 6,
      synced: 5,
      out_of_sync: 1,
      healthy: 6,
      degraded: 0,
      progressing: 0,
      unknown: 0,
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockOverview });

    const result = await argocdService.getOverview();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/argocd/overview", { params: {} });
    expect(result.total).toBe(6);
    expect(result.synced).toBe(5);
  });
});
