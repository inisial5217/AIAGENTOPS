import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "./card";

describe("Card Component", () => {
  it("renders card with header, title, content and footer", () => {
    render(
      <Card accentTop="cyan">
        <CardHeader>
          <CardTitle>System Performance</CardTitle>
          <CardDescription>Live CPU and memory telemetry</CardDescription>
        </CardHeader>
        <CardContent>
          <div>CPU: 45%</div>
        </CardContent>
        <CardFooter>
          <span>Updated just now</span>
        </CardFooter>
      </Card>
    );

    expect(screen.getByText("System Performance")).toBeDefined();
    expect(screen.getByText("Live CPU and memory telemetry")).toBeDefined();
    expect(screen.getByText("CPU: 45%")).toBeDefined();
    expect(screen.getByText("Updated just now")).toBeDefined();
  });

  it("applies accentTop border classes correctly", () => {
    const { container } = render(<Card accentTop="emerald" className="custom-class" />);
    const cardEl = container.firstChild as HTMLElement;
    expect(cardEl.className).toContain("border-t-[var(--status-success)]");
    expect(cardEl.className).toContain("custom-class");
  });
});
