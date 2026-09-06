"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { dockerService } from "../../../../services/docker-service";
import { Badge } from "../../../../components/ui/badge";
import { Button } from "../../../../components/ui/button";
import { Network, RefreshCw, Search } from "lucide-react";
import { DockerNavTabs } from "../../../../components/layout/docker-nav-tabs";

export default function DockerNetworksPage() {
  const [search, setSearch] = React.useState("");

  const { data, isLoading, isRefetching, refetch } = useQuery({
    queryKey: ["docker", "networks"],
    queryFn: () => dockerService.getNetworks(),
  });

  const networks = data?.data || [];
  const filtered = networks.filter((n) => {
    if (!search) return true;
    return (
      n.name.toLowerCase().includes(search.toLowerCase()) ||
      n.driver.toLowerCase().includes(search.toLowerCase()) ||
      n.id.toLowerCase().includes(search.toLowerCase())
    );
  });

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-[var(--border-subtle)]">
        <div className="text-left">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-[var(--text-primary)]">
              Docker Networks
            </h1>
            <Badge variant="cyan" size="sm">
              SDN Topology
            </Badge>
          </div>
          <p className="text-xs text-[var(--text-muted)] font-mono mt-1">
            Bridge, host, and overlay networks &bull; {networks.length} Total
          </p>
        </div>

        <div className="flex items-center gap-3">
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-muted)]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search networks..."
              className="w-full h-8 pl-8 pr-3 rounded-lg bg-[var(--bg-secondary)] border border-[var(--border-default)] text-xs text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:outline-none focus:border-cyan-500"
            />
          </div>

          <Button
            size="sm"
            variant="outline"
            onClick={() => refetch()}
            isLoading={isRefetching}
            title="Refresh networks"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>

      <DockerNavTabs />

      <div className="bg-[var(--bg-card)] border border-[var(--border-default)] rounded-xl overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-mono">
            <thead className="bg-[var(--bg-secondary)] border-b border-[var(--border-subtle)] text-[var(--text-muted)]">
              <tr>
                <th className="py-3 px-4">Network Name</th>
                <th className="py-3 px-4">Network ID</th>
                <th className="py-3 px-4">Driver</th>
                <th className="py-3 px-4">Scope</th>
                <th className="py-3 px-4">Internal</th>
                <th className="py-3 px-4">Subnet / Gateway</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--border-subtle)]">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-[var(--text-muted)]">
                    Loading networks from Docker Engine...
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-[var(--text-muted)]">
                    No networks found.
                  </td>
                </tr>
              ) : (
                filtered.map((n) => (
                  <tr key={n.id} className="hover:bg-[var(--bg-hover)] transition-colors">
                    <td className="py-3 px-4 font-semibold text-[var(--text-primary)]">
                      <div className="flex items-center gap-2">
                        <Network className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                        <span>{n.name}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-[var(--text-muted)]">{n.id.substring(0, 12)}</td>
                    <td className="py-3 px-4">
                      <Badge variant="neutral" size="sm">
                        {n.driver}
                      </Badge>
                    </td>
                    <td className="py-3 px-4 text-[var(--text-secondary)]">{n.scope}</td>
                    <td className="py-3 px-4 text-[var(--text-muted)]">{n.internal ? "Yes" : "No"}</td>
                    <td className="py-3 px-4 text-[var(--text-muted)]">{n.subnet || n.gateway || "-"}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
