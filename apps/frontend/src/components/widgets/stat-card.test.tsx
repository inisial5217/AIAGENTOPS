import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { StatCard } from "./stat-card";

describe("StatCard Component", () => {
  it("renders metric value and labels", () => {
    render(
      <StatCard
        value="145"
        label="Total Kontainer"
        sublabel="12 Namespaces"
        color="cyan"
      />
    );

    expect(screen.getByText("145")).toBeDefined();
    expect(screen.getByText("Total Kontainer")).toBeDefined();
    expect(screen.getByText("12 Namespaces")).toBeDefined();
  });

  it("renders skeleton placeholder when isLoading is true", () => {
    const { container } = render(
      <StatCard
        value="145"
        label="Total Kontainer"
        isLoading={true}
      />
    );

    // Should render skeleton shimmer elements, not the text
    expect(screen.queryByText("145")).toBeNull();
    expect(container.querySelectorAll(".animate-pulse").length).toBeGreaterThan(0);
  });
});
