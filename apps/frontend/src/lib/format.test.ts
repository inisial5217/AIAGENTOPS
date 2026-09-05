import { describe, it, expect } from "vitest";
import {
  formatBytes,
  formatUptime,
  formatPercentage,
  formatDate,
} from "./format";

describe("Format Utilities", () => {
  it("formats bytes accurately across units", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(1024)).toBe("1 KB");
    expect(formatBytes(1536)).toBe("1.5 KB");
    expect(formatBytes(1048576)).toBe("1 MB");
    expect(formatBytes(1073741824)).toBe("1 GB");
  });

  it("formats uptime in human-readable notation", () => {
    expect(formatUptime(0)).toBe("0s");
    expect(formatUptime(-5)).toBe("0s");
    expect(formatUptime(45)).toBe("0m");
    expect(formatUptime(125)).toBe("2m");
    expect(formatUptime(3665)).toBe("1h 1m");
    expect(formatUptime(90000)).toBe("1d 1h");
  });

  it("formats percentage with precision", () => {
    expect(formatPercentage(50.1234)).toBe("50.1%");
    expect(formatPercentage(50.1234, 2)).toBe("50.12%");
    expect(formatPercentage(150)).toBe("150.0%");
  });

  it("formats date into ISO pattern", () => {
    const d = new Date("2026-09-05T12:00:00Z");
    const res = formatDate(d);
    expect(res).toContain("2026");
  });
});
