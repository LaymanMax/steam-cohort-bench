import {
  BarChart3,
  BookOpen,
  Columns,
  Database,
  ExternalLink,
  Filter,
  Library,
  ShieldCheck,
  Sparkles,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { CohortComparisonPage } from "../features/cohort-comparison/CohortComparisonPage";
import { CohortExplorerPage } from "../features/cohort-explorer/CohortExplorerPage";
import { DataQualityPage } from "../features/data-quality/DataQualityPage";
import { GameCatalogPage } from "../features/game-catalog/GameCatalogPage";
import { GameProfilePage } from "../features/game-profile/GameProfilePage";
import { MethodologyPage } from "../features/methodology/MethodologyPage";
import { OverviewPage } from "../features/overview/OverviewPage";
import { api } from "../services/api";
import { DatasetManifest } from "../types";

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>("overview");
  const [manifest, setManifest] = useState<DatasetManifest | null>(null);

  useEffect(() => {
    api
      .getMeta()
      .then((m) => setManifest(m))
      .catch(() => {
        // Backend might still be warming up
      });
  }, []);

  const navItems = [
    { id: "overview", label: "Overview", icon: BarChart3 },
    { id: "explorer", label: "Cohort Explorer", icon: Filter },
    { id: "compare", label: "Compare Cohorts", icon: Columns },
    { id: "profile", label: "Game Profile", icon: Sparkles },
    { id: "catalog", label: "Game Catalog", icon: Library },
    { id: "quality", label: "Data Quality", icon: ShieldCheck },
    { id: "methodology", label: "Methodology", icon: BookOpen },
  ];

  return (
    <div className="app-container">
      {/* Sidebar Navigation */}
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="app-logo">
            <BarChart3 size={20} color="var(--cyan)" />
            <span>steam</span>cohort-bench
          </div>
        </div>

        <ul className="nav-list">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <li key={item.id}>
                <a
                  className={`nav-item ${isActive ? "active" : ""}`}
                  onClick={() => setActiveTab(item.id)}
                >
                  <Icon size={17} />
                  <span>{item.label}</span>
                </a>
              </li>
            );
          })}
        </ul>

        <div className="sidebar-footer">
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.4rem",
              marginBottom: "0.35rem",
            }}
          >
            <Database size={13} color="var(--cyan)" />
            <span>DuckDB + Parquet Analytical Engine</span>
          </div>
          <div style={{ fontSize: "0.7rem", color: "var(--text-muted)" }}>
            v1.0.0 • Open Source MIT
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="main-content">
        <header className="top-bar">
          <div className="top-bar-title">
            {navItems.find((n) => n.id === activeTab)?.label}
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
            <div className="status-pill">
              <span
                style={{
                  width: "6px",
                  height: "6px",
                  borderRadius: "50%",
                  backgroundColor: manifest ? "var(--emerald)" : "var(--amber)",
                }}
              />
              {manifest ? (
                <span>
                  Dataset: {manifest.source_row_count.toLocaleString()} rows
                  (rev: {manifest.revision})
                </span>
              ) : (
                <span>No Dataset Loaded</span>
              )}
            </div>

            <a
              href="https://huggingface.co/datasets/FronkonGames/steam-games-dataset"
              target="_blank"
              rel="noreferrer"
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.3rem",
                fontSize: "0.75rem",
                color: "var(--text-dim)",
                textDecoration: "none",
              }}
            >
              Hugging Face Dataset <ExternalLink size={11} />
            </a>
          </div>
        </header>

        <div>
          {activeTab === "overview" && (
            <OverviewPage manifest={manifest} onNavigate={setActiveTab} />
          )}
          {activeTab === "explorer" && (
            <CohortExplorerPage manifest={manifest} />
          )}
          {activeTab === "compare" && (
            <CohortComparisonPage manifest={manifest} />
          )}
          {activeTab === "profile" && <GameProfilePage manifest={manifest} />}
          {activeTab === "catalog" && <GameCatalogPage manifest={manifest} />}
          {activeTab === "quality" && <DataQualityPage />}
          {activeTab === "methodology" && <MethodologyPage />}
        </div>
      </main>
    </div>
  );
};
export default App;
