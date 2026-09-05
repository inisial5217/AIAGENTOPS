"use client";

import { useAuthStore } from "../lib/auth";

export function useAuth() {
  const { user, token, isLoading, error, login, logout, fetchMe, initAuth } =
    useAuthStore();

  const isAuthenticated = !!token && !!user;
  const isAdmin = user?.role === "admin";
  const isDevOps = user?.role === "devops" || isAdmin;
  const isViewer = !!user?.role;

  const loginWithKeycloak = () => {
    const keycloakUrl =
      process.env.NEXT_PUBLIC_KEYCLOAK_URL || "http://localhost:8180";
    const realm = process.env.NEXT_PUBLIC_KEYCLOAK_REALM || "cifo";
    const clientId =
      process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT_ID || "cifo-frontend";
    const redirectUri =
      typeof window !== "undefined"
        ? `${window.location.origin}/login`
        : "http://localhost:3000/login";

    const authUrl = `${keycloakUrl}/realms/${realm}/protocol/openid-connect/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(
      redirectUri
    )}&response_type=token&scope=openid%20profile%20email`;

    if (typeof window !== "undefined") {
      window.location.href = authUrl;
    }
  };

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    login,
    logout,
    fetchMe,
    initAuth,
    isAdmin,
    isDevOps,
    isViewer,
    loginWithKeycloak,
  };
}
