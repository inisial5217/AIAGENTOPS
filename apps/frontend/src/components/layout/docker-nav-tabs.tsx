"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Container, Layers, HardDrive, Network } from "lucide-react";

export function DockerNavTabs() {
  const pathname = usePathname();

  const tabs = [
    {
      name: "Containers",
      href: "/docker",
      icon: Container,
      isActive: pathname === "/docker" || pathname === "/docker/containers",
    },
    {
      name: "Images",
      href: "/docker/images",
      icon: Layers,
      isActive: pathname === "/docker/images",
    },
    {
      name: "Volumes",
      href: "/docker/volumes",
      icon: HardDrive,
      isActive: pathname === "/docker/volumes",
    },
    {
      name: "Networks",
      href: "/docker/networks",
      icon: Network,
      isActive: pathname === "/docker/networks",
    },
  ];

  return (
    <div className="flex items-center gap-1 p-1 bg-[var(--bg-secondary)] border border-[var(--border-subtle)] rounded-lg text-xs font-mono w-fit">
      {tabs.map((tab) => {
        const Icon = tab.icon;
        return (
          <Link
            key={tab.href}
            href={tab.href}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-md transition-all ${
              tab.isActive
                ? "bg-cyan-500/15 text-cyan-400 font-semibold border border-cyan-500/30 shadow-sm"
                : "text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-card)]"
            }`}
          >
            <Icon className="w-3.5 h-3.5" />
            <span>{tab.name}</span>
          </Link>
        );
      })}
    </div>
  );
}
