import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { Input } from "./input";

describe("Input Component", () => {
  it("renders input with label and placeholder", () => {
    render(
      <Input
        label="Username"
        placeholder="Enter your username"
      />
    );

    expect(screen.getByText("Username")).toBeDefined();
    expect(screen.getByPlaceholderText("Enter your username")).toBeDefined();
  });

  it("handles onChange event and user typing", () => {
    const handleChange = vi.fn();
    render(
      <Input
        placeholder="Type here"
        onChange={handleChange}
      />
    );

    const input = screen.getByPlaceholderText("Type here");
    fireEvent.change(input, { target: { value: "admin" } });

    expect(handleChange).toHaveBeenCalledTimes(1);
    expect((input as HTMLInputElement).value).toBe("admin");
  });

  it("displays error message when provided", () => {
    render(
      <Input
        label="Password"
        error="Password is required"
      />
    );

    expect(screen.getByText("Password is required")).toBeDefined();
  });

  it("respects disabled state", () => {
    render(
      <Input
        placeholder="Disabled input"
        disabled
      />
    );

    const input = screen.getByPlaceholderText("Disabled input") as HTMLInputElement;
    expect(input.disabled).toBe(true);
  });
});
