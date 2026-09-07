import { describe, it, expect, beforeEach, vi } from "vitest";
import { useAuthStore } from "./auth";
import { apiClient } from "./api";

describe("useAuthStore", () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({
      token: null,
      user: null,
      isLoading: false,
      error: null,
    });
    vi.restoreAllMocks();
  });

  it("handles successful login", async () => {
    vi.spyOn(apiClient, "post").mockResolvedValueOnce({
      data: { access_token: "jwt-test-token" },
    });
    vi.spyOn(apiClient, "get").mockResolvedValueOnce({
      data: {
        user: {
          id: "usr-1",
          name: "Test User",
          email: "test@cifo.local",
          role: "admin",
        },
      },
    });

    const success = await useAuthStore.getState().login("admin", "adminpass");

    expect(success).toBe(true);
    expect(useAuthStore.getState().token).toBe("jwt-test-token");
    expect(useAuthStore.getState().user?.name).toBe("Test User");
    expect(localStorage.getItem("cifo_access_token")).toBe("jwt-test-token");
  });

  it("handles failed login", async () => {
    vi.spyOn(apiClient, "post").mockRejectedValueOnce({
      response: { data: { detail: "Invalid credentials" } },
    });

    const success = await useAuthStore.getState().login("baduser", "wrong");

    expect(success).toBe(false);
    expect(useAuthStore.getState().error).toBe("Invalid credentials");
    expect(useAuthStore.getState().token).toBeNull();
  });

  it("handles logout", async () => {
    useAuthStore.setState({
      token: "existing-token",
      user: { id: "usr-1", name: "User", email: "u@cifo.local", role: "viewer" },
    });
    localStorage.setItem("cifo_access_token", "existing-token");

    vi.spyOn(apiClient, "post").mockResolvedValueOnce({});

    await useAuthStore.getState().logout();

    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
    expect(localStorage.getItem("cifo_access_token")).toBeNull();
  });

  it("initializes auth from localStorage", async () => {
    localStorage.setItem("cifo_access_token", "stored-token");

    vi.spyOn(apiClient, "get").mockResolvedValueOnce({
      data: {
        user: {
          id: "usr-2",
          name: "Stored User",
          email: "stored@cifo.local",
          role: "devops",
        },
      },
    });

    await useAuthStore.getState().initAuth();

    expect(useAuthStore.getState().token).toBe("stored-token");
    expect(useAuthStore.getState().user?.role).toBe("devops");
  });
});
