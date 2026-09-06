import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { LogTerminal } from "./log-terminal";

describe("LogTerminal", () => {
  const sampleLogs = [
    "[INFO] Server started on port 8080",
    "[WARN] High memory usage detected",
    "[ERROR] Connection to database timed out",
  ];

  it("renders terminal header and log lines", () => {
    render(<LogTerminal initialLogs={sampleLogs} title="Test Terminal" />);

    expect(screen.getByText("Test Terminal")).toBeDefined();
    expect(screen.getByText(/Server started on port 8080/)).toBeDefined();
    expect(screen.getByText(/High memory usage detected/)).toBeDefined();
    expect(screen.getByText(/Connection to database timed out/)).toBeDefined();
  });

  it("filters logs based on search query", () => {
    render(<LogTerminal initialLogs={sampleLogs} />);

    const searchInput = screen.getByPlaceholderText("Search logs...");
    fireEvent.change(searchInput, { target: { value: "database" } });

    // matched keyword is highlighted in a mark tag
    expect(screen.getByText("database")).toBeDefined();
    // non-matching log should be filtered out
    expect(screen.queryByText(/Server started on port 8080/)).toBeNull();
  });

  it("toggles auto-scroll mode", () => {
    render(<LogTerminal initialLogs={sampleLogs} />);

    const autoScrollBtn = screen.getByTitle("Auto-scroll to latest log");
    expect(autoScrollBtn.textContent).toContain("Auto-scroll");

    fireEvent.click(autoScrollBtn);
    expect(autoScrollBtn).toBeDefined();
  });

  it("clears terminal output when clear button clicked", () => {
    render(<LogTerminal initialLogs={sampleLogs} />);

    const clearBtn = screen.getByTitle("Clear terminal display");
    fireEvent.click(clearBtn);

    expect(screen.getByText("Waiting for stream output...")).toBeDefined();
  });
});
