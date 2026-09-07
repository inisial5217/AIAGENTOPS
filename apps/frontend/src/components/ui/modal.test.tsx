import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { Modal } from "./modal";

describe("Modal Component", () => {
  it("renders modal when open", () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()} title="Test Modal" description="Modal description">
        <div>Modal Content Body</div>
      </Modal>
    );

    expect(screen.getByText("Test Modal")).toBeDefined();
    expect(screen.getByText("Modal description")).toBeDefined();
    expect(screen.getByText("Modal Content Body")).toBeDefined();
  });

  it("does not render content when closed", () => {
    render(
      <Modal isOpen={false} onClose={vi.fn()} title="Hidden Modal">
        <div>Hidden Body</div>
      </Modal>
    );

    expect(screen.queryByText("Hidden Modal")).toBeNull();
    expect(screen.queryByText("Hidden Body")).toBeNull();
  });

  it("calls onClose when close button clicked", () => {
    const handleClose = vi.fn();
    render(
      <Modal isOpen={true} onClose={handleClose} title="Closable Modal">
        <div>Content</div>
      </Modal>
    );

    const closeBtn = screen.getByRole("button", { name: /close/i });
    fireEvent.click(closeBtn);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });
});
