import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import IncidentsPage from "./page";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { incidentService } from "../../../services/incident-service";

// mock incidentService
vi.mock("../../../services/incident-service", () => ({
  incidentService: {
    listIncidents: vi.fn(),
    getIncidentStats: vi.fn(),
    getIncident: vi.fn(),
    acknowledgeIncident: vi.fn(),
    resolveIncident: vi.fn(),
    closeIncident: vi.fn(),
  },
}));

// mock useAuth
vi.mock("../../../hooks/use-auth", () => ({
  useAuth: () => ({
    user: { id: "u-1", name: "DevOps Engineer", role: "devops" },
    logout: vi.fn(),
  }),
}));

// mock useWebSocket
vi.mock("../../../hooks/use-websocket", () => ({
  useWebSocket: () => ({
    status: "connected",
    subscribe: vi.fn(() => vi.fn()),
    sendMessage: vi.fn(),
  }),
}));

describe("IncidentsPage", () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });
  });

  const renderWithClient = (ui: React.ReactElement) => {
    return render(
      <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
    );
  };

  it("should render page header, stats cards, and live indicator", async () => {
    vi.mocked(incidentService.getIncidentStats).mockResolvedValueOnce({
      total: 5,
      open: 2,
      acknowledged: 1,
      investigating: 0,
      resolved: 1,
      closed: 1,
      critical_count: 1,
    });

    vi.mocked(incidentService.listIncidents).mockResolvedValueOnce({
      data: [
        {
          id: "inc-101",
          title: "ContainerCrashLooping",
          description: "Pod crashed repeatedly",
          severity: "critical",
          status: "open",
          source: "Kubernetes",
          resource_id: "payment-gateway",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      total: 1,
    });

    renderWithClient(<IncidentsPage />);

    // verify title and subtitle
    expect(screen.getByText("Incident Management")).toBeDefined();
    expect(screen.getByText(/LIVE ALERTS/i)).toBeDefined();

    // verify KPI titles
    expect(screen.getByText("Total Incidents")).toBeDefined();
    expect(screen.getByText("Open (Action Needed)")).toBeDefined();

    // wait for incident row to appear
    await waitFor(() => {
      expect(screen.getByText("ContainerCrashLooping")).toBeDefined();
      expect(screen.getByText("CRITICAL")).toBeDefined();
      expect(screen.getByText("Ack")).toBeDefined();
    });
  });
});
