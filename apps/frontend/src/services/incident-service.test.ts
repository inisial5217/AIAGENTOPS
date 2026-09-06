import { describe, it, expect, vi, beforeEach } from "vitest";
import { incidentService } from "./incident-service";
import { apiClient } from "../lib/api-client";

vi.mock("../lib/api-client", () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

describe("incidentService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should list incidents with proper query params", async () => {
    const mockData = {
      data: [{ id: "inc-1", title: "ContainerOOMKilled", status: "open" }],
      total: 1,
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: mockData });

    const result = await incidentService.listIncidents({
      status: "open",
      severity: "critical",
      search: "payment",
    });

    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/incidents", {
      params: {
        status: "open",
        severity: "critical",
        search: "payment",
      },
    });
    expect(result).toEqual(mockData);
  });

  it("should fetch incident stats", async () => {
    const mockStats = {
      total: 10,
      open: 3,
      acknowledged: 2,
      investigating: 1,
      resolved: 3,
      closed: 1,
      critical_count: 2,
    };
    vi.mocked(apiClient.get).mockResolvedValueOnce({ data: { data: mockStats } });

    const result = await incidentService.getIncidentStats();
    expect(apiClient.get).toHaveBeenCalledWith("/api/v1/incidents/stats");
    expect(result).toEqual(mockStats);
  });

  it("should acknowledge an incident", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "acknowledged" },
    });

    const result = await incidentService.acknowledgeIncident("inc-123");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/incidents/inc-123/acknowledge");
    expect(result.status).toBe("success");
  });

  it("should resolve an incident", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "resolved" },
    });

    const result = await incidentService.resolveIncident("inc-123");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/incidents/inc-123/resolve");
    expect(result.status).toBe("success");
  });

  it("should close an incident", async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { status: "success", message: "closed" },
    });

    const result = await incidentService.closeIncident("inc-123");
    expect(apiClient.post).toHaveBeenCalledWith("/api/v1/incidents/inc-123/close");
    expect(result.status).toBe("success");
  });
});
