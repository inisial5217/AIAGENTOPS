import { create } from "zustand";

interface SidebarState {
  isCollapsed: boolean;
  expandedSections: Record<string, boolean>;
  toggleSidebar: () => void;
  setCollapsed: (collapsed: boolean) => void;
  toggleSection: (section: string) => void;
}

export const useSidebarStore = create<SidebarState>((set) => ({
  isCollapsed: false,
  expandedSections: {
    kubernetes: true,
    docker: false,
  },

  toggleSidebar: () => {
    set((state) => ({ isCollapsed: !state.isCollapsed }));
  },

  setCollapsed: (collapsed: boolean) => {
    set({ isCollapsed: collapsed });
  },

  toggleSection: (section: string) => {
    set((state) => ({
      expandedSections: {
        ...state.expandedSections,
        [section]: !state.expandedSections[section],
      },
    }));
  },
}));
