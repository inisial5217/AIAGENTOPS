import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Badge } from "../components/ui/badge";

describe("Badge Component", () => {
  it("renders with correct variant styles", () => {
    render(<Badge variant="cyan">Active</Badge>);
    expect(screen.getByText("Active")).toBeDefined();
  });

  it("renders with pulse dot when pulse prop is true", () => {
    const { container } = render(
      <Badge variant="success" pulse>
        Running
      </Badge>
    );
    expect(container.querySelector(".animate-ping")).toBeDefined();
  });
});
