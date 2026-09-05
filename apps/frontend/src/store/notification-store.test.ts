import { describe, it, expect, beforeEach } from "vitest";
import { useNotificationStore } from "./notification-store";

describe("Notification Store", () => {
  beforeEach(() => {
    useNotificationStore.setState({
      notifications: [
        {
          id: "test-1",
          title: "Alert 1",
          message: "Msg 1",
          severity: "info",
          timestamp: "12:00",
          read: false,
        },
        {
          id: "test-2",
          title: "Alert 2",
          message: "Msg 2",
          severity: "critical",
          timestamp: "12:01",
          read: false,
        },
      ],
      unreadCount: 2,
    });
  });

  it("adds a new notification and increments unread", () => {
    useNotificationStore.getState().addNotification({
      title: "New Incident",
      message: "Pod crashed",
      severity: "error",
    });

    const state = useNotificationStore.getState();
    expect(state.notifications.length).toBe(3);
    expect(state.unreadCount).toBe(3);
    expect(state.notifications[0].title).toBe("New Incident");
  });

  it("marks a notification as read and updates count", () => {
    useNotificationStore.getState().markAsRead("test-1");
    const state = useNotificationStore.getState();
    expect(state.unreadCount).toBe(1);
    expect(state.notifications.find((n) => n.id === "test-1")?.read).toBe(true);
  });

  it("marks all as read", () => {
    useNotificationStore.getState().markAllAsRead();
    const state = useNotificationStore.getState();
    expect(state.unreadCount).toBe(0);
    expect(state.notifications.every((n) => n.read)).toBe(true);
  });

  it("clears all notifications", () => {
    useNotificationStore.getState().clearAll();
    const state = useNotificationStore.getState();
    expect(state.notifications.length).toBe(0);
    expect(state.unreadCount).toBe(0);
  });
});
