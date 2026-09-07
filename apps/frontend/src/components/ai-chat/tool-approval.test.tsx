import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ToolApproval } from "./tool-approval";
import { ToolCallInfo } from "../../types/ai";

describe("ToolApproval Component", () => {
  const mockToolCall: ToolCallInfo = {
    id: "call-12345",
    name: "restart_container",
    parameters: {
      container_id: "cifo-backend-prod",
      force: true,
    },
    status: "pending",
  };

  it("renders pending tool call with parameters and warning", () => {
    render(
      <ToolApproval
        toolCall={mockToolCall}
        onApprove={vi.fn()}
        onReject={vi.fn()}
      />
    );

    expect(screen.getByText("restart_container")).toBeDefined();
    expect(screen.getByText("cifo-backend-prod")).toBeDefined();
    expect(screen.getByText("pending")).toBeDefined();
    expect(screen.getByText(/User confirmation required/i)).toBeDefined();
    expect(screen.getByRole("button", { name: /Approve & Execute/i })).toBeDefined();
    expect(screen.getByRole("button", { name: /Reject Action/i })).toBeDefined();
  });

  it("triggers onApprove when approve button is clicked", async () => {
    const handleApprove = vi.fn().mockResolvedValue(undefined);
    render(
      <ToolApproval
        toolCall={mockToolCall}
        onApprove={handleApprove}
        onReject={vi.fn()}
      />
    );

    const approveBtn = screen.getByRole("button", { name: /Approve & Execute/i });
    fireEvent.click(approveBtn);

    await waitFor(() => {
      expect(handleApprove).toHaveBeenCalledWith("call-12345");
    });
  });

  it("shows rejection input and calls onReject with custom reason", async () => {
    const handleReject = vi.fn().mockResolvedValue(undefined);
    render(
      <ToolApproval
        toolCall={mockToolCall}
        onApprove={vi.fn()}
        onReject={handleReject}
      />
    );

    const rejectBtn = screen.getByRole("button", { name: /Reject Action/i });
    fireEvent.click(rejectBtn);

    const reasonInput = screen.getByPlaceholderText(/Reason for rejection/i);
    fireEvent.change(reasonInput, { target: { value: "Maintenance window active" } });

    const confirmBtn = screen.getByRole("button", { name: /Confirm Reject/i });
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(handleReject).toHaveBeenCalledWith("call-12345", "Maintenance window active");
    });
  });

  it("renders executed tool call with result", () => {
    const executedToolCall: ToolCallInfo = {
      id: "call-99999",
      name: "scale_deployment",
      parameters: { deployment_name: "order-service", replicas: 3 },
      status: "executed",
      result: "Deployment successfully scaled to 3 replicas",
    };

    render(
      <ToolApproval
        toolCall={executedToolCall}
        onApprove={vi.fn()}
        onReject={vi.fn()}
      />
    );

    expect(screen.getByText("executed")).toBeDefined();
    expect(screen.getByText("Deployment successfully scaled to 3 replicas")).toBeDefined();
    expect(screen.queryByRole("button", { name: /Approve & Execute/i })).toBeNull();
  });

  it("renders rejected tool call status badge", () => {
    const rejectedToolCall: ToolCallInfo = {
      id: "call-88888",
      name: "delete_pod",
      parameters: { pod_name: "rogue-pod" },
      status: "rejected",
    };

    render(
      <ToolApproval
        toolCall={rejectedToolCall}
        onApprove={vi.fn()}
        onReject={vi.fn()}
      />
    );

    expect(screen.getByText("rejected")).toBeDefined();
    expect(screen.queryByRole("button", { name: /Approve & Execute/i })).toBeNull();
  });
});
