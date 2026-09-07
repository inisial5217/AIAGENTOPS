import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ChatInput } from "./chat-input";

describe("ChatInput Component", () => {
  it("renders input field and quick suggestions", () => {
    render(<ChatInput onSendMessage={vi.fn()} isLoading={false} />);

    expect(screen.getByPlaceholderText(/Ask AI DevOps Assistant/i)).toBeDefined();
    expect(screen.getByText("Check pod status in default namespace")).toBeDefined();
  });

  it("submits message on send button click", async () => {
    const handleSend = vi.fn().mockResolvedValue(undefined);
    const { container } = render(<ChatInput onSendMessage={handleSend} isLoading={false} />);

    const textarea = screen.getByPlaceholderText(/Ask AI DevOps Assistant/i);
    fireEvent.change(textarea, { target: { value: "Restart container payment-service" } });

    const sendBtn = container.querySelector("#ai-chat-send-btn") as HTMLButtonElement;
    expect(sendBtn.disabled).toBe(false);
    fireEvent.click(sendBtn);

    await waitFor(() => {
      expect(handleSend).toHaveBeenCalledWith("Restart container payment-service");
    });
  });

  it("populates textarea when quick suggestion is clicked", () => {
    render(<ChatInput onSendMessage={vi.fn()} isLoading={false} />);

    const suggestionBtn = screen.getByText("Show failed docker containers");
    fireEvent.click(suggestionBtn);

    const textarea = screen.getByPlaceholderText(/Ask AI DevOps Assistant/i) as HTMLTextAreaElement;
    expect(textarea.value).toBe("Show failed docker containers");
  });

  it("disables send button when empty or loading", () => {
    const { container, rerender } = render(<ChatInput onSendMessage={vi.fn()} isLoading={false} />);
    const sendBtn = container.querySelector("#ai-chat-send-btn") as HTMLButtonElement;
    expect(sendBtn.disabled).toBe(true);

    rerender(<ChatInput onSendMessage={vi.fn()} isLoading={true} />);
    expect(sendBtn.disabled).toBe(true);
  });
});
