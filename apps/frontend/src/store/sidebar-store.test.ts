import { describe, it, expect, beforeEach } from "vitest";
import { useSidebarStore } from "./sidebar-store";

describe("Sidebar Store", () => {
  beforeEach(() => {
    useSidebarStore.setState({
      isCollapsed: false,
      expandedSections: { kubernetes: true, docker: true },
    });
  });

  it("toggles collapsed state", () => {
    expect(useSidebarStore.getState().isCollapsed).toBe(false);
    useSidebarStore.getState().toggleSidebar();
    expect(useSidebarStore.getState().isCollapsed).toBe(true);
    useSidebarStore.getState().toggleSidebar();
    expect(useSidebarStore.getState().isCollapsed).toBe(false);
  });

  it("toggles sub-navigation accordion sections", () => {
    expect(useSidebarStore.getState().expandedSections.kubernetes).toBe(true);
    useSidebarStore.getState().toggleSection("kubernetes");
    expect(useSidebarStore.getState().expandedSections.kubernetes).toBe(false);
  });
});
