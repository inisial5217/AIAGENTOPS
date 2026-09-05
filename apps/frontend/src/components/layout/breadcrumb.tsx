"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight, Home } from "lucide-react";

export function Breadcrumb() {
  const pathname = usePathname();

  const segments = pathname.split("/").filter(Boolean);

  const getLabel = (segment: string) => {
    switch (segment.toLowerCase()) {
      case "monitoring":
        return "Monitoring";
      case "kubernetes":
        return "Kubernetes";
      case "docker":
        return "Docker";
      case "pods":
        return "Pods";
      case "nodes":
        return "Nodes";
      case "deployments":
        return "Deployments";
      case "services":
        return "Services & Ingress";
      case "namespaces":
        return "Namespaces";
      case "containers":
        return "Containers";
      case "images":
        return "Images";
      case "volumes":
        return "Volumes";
      case "networks":
        return "Networks";
      case "incidents":
        return "Incidents";
      case "settings":
        return "Settings";
      default:
        return segment.charAt(0).toUpperCase() + segment.slice(1);
    }
  };

  return (
    <nav
      aria-label="Breadcrumb"
      className="flex items-center gap-1.5 text-xs text-[var(--text-muted)] font-mono select-none"
    >
      <Link
        href="/monitoring"
        className="flex items-center gap-1 hover:text-[var(--text-primary)] transition-colors"
      >
        <Home className="w-3.5 h-3.5" />
        <span className="sr-only">Home</span>
      </Link>

      {segments.length === 0 && (
        <>
          <ChevronRight className="w-3 h-3 text-[var(--text-muted)]" />
          <span className="text-[var(--text-primary)] font-medium">Monitoring</span>
        </>
      )}

      {segments.map((segment, index) => {
        const href = `/${segments.slice(0, index + 1).join("/")}`;
        const isLast = index === segments.length - 1;
        const label = getLabel(segment);

        return (
          <React.Fragment key={href}>
            <ChevronRight className="w-3 h-3 text-[var(--text-muted)] shrink-0" />
            {isLast ? (
              <span className="text-[var(--text-primary)] font-medium">
                {label}
              </span>
            ) : (
              <Link
                href={href}
                className="hover:text-[var(--text-primary)] transition-colors"
              >
                {label}
              </Link>
            )}
          </React.Fragment>
        );
      })}
    </nav>
  );
}
