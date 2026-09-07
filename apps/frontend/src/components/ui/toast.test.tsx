import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Toast, ToastTitle, ToastDescription, ToastProvider, ToastViewport } from "./toast";

describe("Toast Component", () => {
  it("renders toast with title and description", () => {
    render(
      <ToastProvider>
        <Toast variant="success">
          <ToastTitle>Operation Successful</ToastTitle>
          <ToastDescription>Data was synced to cluster</ToastDescription>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    );

    expect(screen.getByText("Operation Successful")).toBeDefined();
    expect(screen.getByText("Data was synced to cluster")).toBeDefined();
  });

  it("applies error variant styles", () => {
    const { container } = render(
      <ToastProvider>
        <Toast variant="error">
          <ToastTitle>Failed</ToastTitle>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    );

    const toastRoot = container.querySelector("[data-radix-collection-item]");
    expect(toastRoot?.className).toContain("text-red-300");
  });

  it("applies warning variant styles", () => {
    const { container } = render(
      <ToastProvider>
        <Toast variant="warning">
          <ToastTitle>Warning</ToastTitle>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    );

    const toastRoot = container.querySelector("[data-radix-collection-item]");
    expect(toastRoot?.className).toContain("text-amber-300");
  });
});
