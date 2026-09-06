import { apiClient } from "../lib/api-client";
import {
  CombinedSettings,
  UpdateSettingsRequest,
  UserAdmin,
} from "../types/settings";

export const settingsService = {
  // getSettings fetches combined system and notification configuration
  async getSettings(): Promise<CombinedSettings> {
    const res = await apiClient.get("/api/v1/settings");
    return res.data.data;
  },

  // updateSettings persists configuration changes
  async updateSettings(req: UpdateSettingsRequest): Promise<CombinedSettings> {
    const res = await apiClient.put("/api/v1/settings", req);
    return res.data.data;
  },

  // testNotification dispatches test Telegram notification
  async testNotification(): Promise<{ status: string; message: string }> {
    const res = await apiClient.post("/api/v1/settings/test-notification");
    return res.data;
  },

  // listUsers fetches paginated user list for administration
  async listUsers(
    limit: number = 50,
    offset: number = 0
  ): Promise<{ users: UserAdmin[]; total: number }> {
    const res = await apiClient.get("/api/v1/settings/users", {
      params: { limit, offset },
    });
    return res.data;
  },

  // updateUserRole updates target user role
  async updateUserRole(id: string, role: string): Promise<UserAdmin> {
    const res = await apiClient.put(`/api/v1/settings/users/${id}/role`, { role });
    return res.data;
  },

  // deactivateUser disables user account
  async deactivateUser(id: string): Promise<UserAdmin> {
    const res = await apiClient.post(`/api/v1/settings/users/${id}/deactivate`);
    return res.data;
  },

  // reactivateUser restores user account
  async reactivateUser(id: string): Promise<UserAdmin> {
    const res = await apiClient.post(`/api/v1/settings/users/${id}/reactivate`);
    return res.data;
  },
};
