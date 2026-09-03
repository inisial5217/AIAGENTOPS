import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import HomePage from "./page";

describe("HomePage", () => {
  it("renders command center title", () => {
    render(<HomePage />);
    expect(screen.getByText(/Command Center Initialized/i)).toBeDefined();
  });
});
