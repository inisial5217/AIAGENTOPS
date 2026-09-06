import { describe, it, expect, vi, beforeEach } from "vitest";
import { settingsService } from "./settings-service";
import { apiClient } from "../lib/api-client";

vi.mock("../lib/api-client", () => ({
  apiClient: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
  },
}));

describe("settingsService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should fetch combined settings", async () => {
    const mockData = {
      system: {
        id: "sys-1",
        app_name: "CIFO Monitoring Platform",
        retention_days: 30,
        refresh_interval: 10,
        ai_auto_remediation: false,
        ai_analysis_threshold: 0.85,
        session_timeout_mins: 60,
        mfa_enforced: false,
        updated_at: new Date().toISOString(),
      },
      notification: {
        id: "notif-1",
        telegram_enabled: true,
        email_enabled: false,
        critical_alert: true,
        warning_alert: true,
        info_alert: false,
        auto_resolve_alert: true,
        quiet_hours_enabled: false,
        updated_at: new Date().toISOString(),
      },
    };

    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockData } });

    const result = await settingsService.getSettings();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/settings");
    expect(result).toEqual(mockData);
  });

  it("should update settings", async () => {
    const updateReq = {
      system: { app_name: "CIFO Production" },
    };
    const mockUpdated = {
      system: {
        id: "sys-1",
        app_name: "CIFO Production",
      },
    };

    vi.mocked(apiClient.put).mockResolvedValueOnce({ data: { data: mockUpdated } });

    const result = await settingsService.updateSettings(updateReq as any);
    expect(apiClient.put).toHaveBeenCalledWith("/api/v1/settings", updateReq);
    expect(result).toEqual(mockUpdated);
  });

  it("should trigger test notification", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "Dispatched" },
    });

    const result = await settingsService.testNotification();
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/settings/test-notification");
    expect(result.status).toBe("success");
  });

  it("should list users with pagination params", async () => {
    const mockUsers = {
      users: [
        {
          id: "u-1",
          email: "admin@cifo.local",
          name: "Admin User",
          role: "admin",
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      total: 1,
    };

    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockUsers });

    const result = await settingsService.listUsers(20, 0);
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/settings/users", {
      params: { limit: 20, offset: 0 },
    });
    expect(result).toEqual(mockUsers);
  });

  it("should update user role", async () => {
    const updatedUser = {
      id: "u-2",
      email: "dev@cifo.local",
      name: "Dev User",
      role: "admin",
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    vi.mocked(apiClient.put).mockResolvedValueOnce({ data: updatedUser });

    const result = await settingsService.updateUserRole("u-2", "admin");
    expect(apiClient.put).toHaveBeenCalledWith("/api/v1/settings/users/u-2/role", {
      role: "admin",
    });
    expect(result).toEqual(updatedUser);
  });

  it("should deactivate user", async () => {
    const deactivated = {
      id: "u-2",
      is_active: false,
    };

    vi.mocked(apiClient.post).mockResolvedValueOnce({ data: deactivated });

    const result = await settingsService.deactivateUser("u-2");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/settings/users/u-2/deactivate");
    expect(result).toEqual(deactivated);
  });

  it("should reactivate user", async () => {
    const reactivated = {
      id: "u-2",
      is_active: true,
    };

    vi.mocked(apiClient.post).mockResolvedValueOnce({ data: reactivated });

    const result = await settingsService.reactivateUser("u-2");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/settings/users/u-2/reactivate");
    expect(result).toEqual(reactivated);
  });
});
