import { create } from "zustand";
import { apiClient } from "./api";

export interface User {
  id: string;
  email: string;
  name: string;
  role: "admin" | "devops" | "viewer";
  keycloak_id?: string;
  is_active?: boolean;
}

interface AuthState {
  token: string | null;
  user: User | null;
  isLoading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  fetchMe: () => Promise<void>;
  initAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isLoading: false,
  error: null,

  initAuth: async () => {
    if (typeof window === "undefined") return;
    const token = localStorage.getItem("cifo_access_token");
    if (token) {
      set({ token });
      await get().fetchMe();
    }
  },

  login: async (username: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const res = await apiClient.post("/api/v1/auth/login", {
        username,
        password,
      });
      const token = res.data.access_token;
      if (token) {
        localStorage.setItem("cifo_access_token", token);
        set({ token, isLoading: false });
        await get().fetchMe();
        return true;
      }
      set({ isLoading: false, error: "Missing token in response" });
      return false;
    } catch (err: any) {
      const msg =
        err.response?.data?.detail || err.message || "Login failed";
      set({ isLoading: false, error: msg });
      return false;
    }
  },

  fetchMe: async () => {
    try {
      const res = await apiClient.get("/api/v1/auth/me");
      if (res.data?.user) {
        set({ user: res.data.user, error: null });
      }
    } catch (err: any) {
      set({ user: null });
    }
  },

  logout: async () => {
    try {
      await apiClient.post("/api/v1/auth/logout");
    } catch (err) {
      // ignore
    } finally {
      localStorage.removeItem("cifo_access_token");
      set({ token: null, user: null, error: null });
    }
  },
}));
