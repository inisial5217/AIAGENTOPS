import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { useAuth } from "./use-auth";
import { useAuthStore } from "../lib/auth";

describe("useAuth hook", () => {
  beforeEach(() => {
    useAuthStore.setState({
      token: null,
      user: null,
      isLoading: false,
      error: null,
    });
  });

  it("returns unauthenticated state initially", () => {
    const { result } = renderHook(() => useAuth());

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isDevOps).toBe(false);
    expect(result.current.isViewer).toBe(false);
    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
  });

  it("correctly identifies admin role", () => {
    useAuthStore.setState({
      token: "mock-admin-token",
      user: {
        id: "usr-admin",
        name: "Admin User",
        email: "admin@cifo.local",
        role: "admin",
      },
    });

    const { result } = renderHook(() => useAuth());

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.isAdmin).toBe(true);
    expect(result.current.isDevOps).toBe(true);
    expect(result.current.isViewer).toBe(true);
  });

  it("correctly identifies devops role", () => {
    useAuthStore.setState({
      token: "mock-devops-token",
      user: {
        id: "usr-devops",
        name: "DevOps Engineer",
        email: "devops@cifo.local",
        role: "devops",
      },
    });

    const { result } = renderHook(() => useAuth());

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isDevOps).toBe(true);
    expect(result.current.isViewer).toBe(true);
  });

  it("correctly identifies viewer role", () => {
    useAuthStore.setState({
      token: "mock-viewer-token",
      user: {
        id: "usr-viewer",
        name: "Viewer User",
        email: "viewer@cifo.local",
        role: "viewer",
      },
    });

    const { result } = renderHook(() => useAuth());

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.isAdmin).toBe(false);
    expect(result.current.isDevOps).toBe(false);
    expect(result.current.isViewer).toBe(true);
  });

  it("constructs keycloak redirect url and navigates", () => {
    const originalLocation = window.location;
    // @ts-expect-error Mocking window.location
    delete window.location;
    window.location = {
      ...originalLocation,
      origin: "http://localhost:3000",
      href: "",
    } as any;

    const { result } = renderHook(() => useAuth());

    act(() => {
      result.current.loginWithKeycloak();
    });

    expect(window.location.href).toContain("/realms/cifo/protocol/openid-connect/auth");
    expect(window.location.href).toContain("client_id=cifo-frontend");

    window.location = originalLocation;
  });
});
