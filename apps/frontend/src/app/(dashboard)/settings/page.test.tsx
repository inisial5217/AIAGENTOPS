import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import SettingsPage from "./page";
import { settingsService } from "../../../services/settings-service";

// mock settingsService
vi.mock("../../../services/settings-service", () => ({
  settingsService: {
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    testNotification: vi.fn(),
    listUsers: vi.fn(),
    updateUserRole: vi.fn(),
    deactivateUser: vi.fn(),
    reactivateUser: vi.fn(),
  },
}));

// mock authStore
vi.mock("../../../lib/auth", () => ({
  useAuthStore: () => ({
    user: { id: "admin-1", email: "admin@cifo.local", name: "System Admin", role: "admin" },
  }),
}));

describe("SettingsPage", () => {
  const mockSettingsData = {
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
      telegram_bot_token: "test-token",
      telegram_chat_id: "test-chat",
      email_enabled: false,
      email_recipients: "admin@cifo.local",
      critical_alert: true,
      warning_alert: true,
      info_alert: false,
      auto_resolve_alert: true,
      quiet_hours_enabled: false,
      updated_at: new Date().toISOString(),
    },
  };

  const mockUsersData = {
    users: [
      {
        id: "admin-1",
        email: "admin@cifo.local",
        name: "System Admin",
        role: "admin" as const,
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      {
        id: "user-2",
        email: "dev@cifo.local",
        name: "DevOps Engineer",
        role: "devops" as const,
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ],
    total: 2,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(settingsService.getSettings).mockResolvedValue(mockSettingsData);
    vi.mocked(settingsService.listUsers).mockResolvedValue(mockUsersData);
  });

  it("should render header, save button, and default General tab", async () => {
    render(<SettingsPage />);

    await waitFor(() => {
      expect(
        screen.getByText("System Settings & Administration")
      ).toBeDefined();
      expect(screen.getByText("Save Changes")).toBeDefined();
      expect(
        screen.getByText("Platform Profile & Preferences")
      ).toBeDefined();
    });
  });

  it("should switch to Notifications tab and allow triggering test notification", async () => {
    vi.mocked(settingsService.testNotification).mockResolvedValueOnce({
      status: "success",
      message: "Dispatched",
    });

    render(<SettingsPage />);

    await waitFor(() => {
      expect(screen.getByText("Notifications")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Notifications"));

    await waitFor(() => {
      expect(screen.getByText("Telegram Bot Channel")).toBeDefined();
      expect(screen.getByText("Email Dispatch Gateway")).toBeDefined();
    });

    const testBtn = screen.getByText("Test Notification Alert");
    expect(testBtn).toBeDefined();
    fireEvent.click(testBtn);

    await waitFor(() => {
      expect(settingsService.testNotification).toHaveBeenCalled();
    });
  });

  it("should switch to Users tab and display user list for admin", async () => {
    render(<SettingsPage />);

    await waitFor(() => {
      expect(screen.getByText("Users & RBAC")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Users & RBAC"));

    await waitFor(() => {
      expect(
        screen.getByText(/Platform User Directory/)
      ).toBeDefined();
      expect(screen.getByText("dev@cifo.local")).toBeDefined();
    });
  });

  it("should switch to AI Configuration tab and display slider & auto-remediation", async () => {
    render(<SettingsPage />);

    await waitFor(() => {
      expect(screen.getByText("AI Configuration")).toBeDefined();
    });

    fireEvent.click(screen.getByText("AI Configuration"));

    await waitFor(() => {
      expect(
        screen.getByText("Autonomous AI Diagnostic Engine")
      ).toBeDefined();
      expect(
        screen.getByText("AI Service Microservice Connected")
      ).toBeDefined();
      expect(
        screen.getByText("Autonomous Remediation Mode")
      ).toBeDefined();
    });
  });

  it("should switch to Security tab and show sessions list", async () => {
    render(<SettingsPage />);

    await waitFor(() => {
      expect(screen.getByText("Security & Sessions")).toBeDefined();
    });

    fireEvent.click(screen.getByText("Security & Sessions"));

    await waitFor(() => {
      expect(
        screen.getByText("Session Governance & Security Policy")
      ).toBeDefined();
      expect(
        screen.getByText(/Active User Sessions/)
      ).toBeDefined();
    });
  });
});
