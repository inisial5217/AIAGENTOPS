import { create } from "zustand";

export interface NotificationItem {
  id: string;
  title: string;
  message: string;
  severity: "info" | "warning" | "error" | "critical";
  timestamp: string;
  read: boolean;
}

interface NotificationState {
  notifications: NotificationItem[];
  unreadCount: number;
  addNotification: (item: Omit<NotificationItem, "id" | "timestamp" | "read">) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
}

const defaultNotifications: NotificationItem[] = [
  {
    id: "notif-1",
    title: "CPU Load Spike",
    message: "node-worker-1 CPU spiked to 99% in production-cifo-1.",
    severity: "critical",
    timestamp: "10:14:19",
    read: false,
  },
  {
    id: "notif-2",
    title: "Connection Timeout",
    message: "target 'payment-gateway' CONNECTION TIMEOUT!",
    severity: "error",
    timestamp: "10:14:18",
    read: false,
  },
  {
    id: "notif-3",
    title: "Docker Sync Completed",
    message: "Docker daemon sync completed with 145 containers.",
    severity: "info",
    timestamp: "10:14:02",
    read: true,
  },
];

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: defaultNotifications,
  unreadCount: 2,

  addNotification: (item) => {
    const newNotif: NotificationItem = {
      ...item,
      id: `notif-${Date.now()}`,
      timestamp: new Date().toLocaleTimeString(),
      read: false,
    };
    set((state) => ({
      notifications: [newNotif, ...state.notifications],
      unreadCount: state.unreadCount + 1,
    }));
  },

  markAsRead: (id: string) => {
    set((state) => {
      const updated = state.notifications.map((n) =>
        n.id === id ? { ...n, read: true } : n
      );
      const unread = updated.filter((n) => !n.read).length;
      return { notifications: updated, unreadCount: unread };
    });
  },

  markAllAsRead: () => {
    set((state) => ({
      notifications: state.notifications.map((n) => ({ ...n, read: true })),
      unreadCount: 0,
    }));
  },

  removeNotification: (id: string) => {
    set((state) => {
      const updated = state.notifications.filter((n) => n.id !== id);
      const unread = updated.filter((n) => !n.read).length;
      return { notifications: updated, unreadCount: unread };
    });
  },

  clearAll: () => {
    set({ notifications: [], unreadCount: 0 });
  },
}));
