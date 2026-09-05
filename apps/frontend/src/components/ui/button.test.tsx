import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { Button } from "./button";

describe("Button Component", () => {
  it("renders with default primary variant", () => {
    render(<Button>Click Me</Button>);
    const btn = screen.getByRole("button", { name: /click me/i });
    expect(btn).toBeDefined();
    expect(btn.className).toContain("bg-[var(--accent-default)]");
  });

  it("handles click events", () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Action</Button>);
    fireEvent.click(screen.getByRole("button", { name: /action/i }));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it("renders loading state with disabled attribute", () => {
    render(<Button isLoading>Submitting</Button>);
    const btn = screen.getByRole("button");
    expect(btn.getAttribute("disabled")).toBeDefined();
  });
});
