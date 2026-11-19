import { AlertTriangle, CheckCircle2, FileText } from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../services/api";
import { DataQualityReport } from "../../types";

export const DataQualityPage: React.FC = () => {
  const [report, setReport] = useState<DataQualityReport | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadQuality();
  }, []);

  const loadQuality = async () => {
    setLoading(true);
    try {
      const res = await api.getDataQuality();
      setReport(res);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  if (loading || !report) {
    return (
      <div className="page-body">
        <h1>Data Quality & Audit</h1>
        <p className="lead">Loading completeness and integrity report...</p>
        <div className="metrics-grid">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="metric-tile skeleton"
              style={{ height: "90px" }}
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="page-body">
      <div style={{ marginBottom: "1.5rem" }}>
        <h1>Data Quality & Integrity Audit</h1>
        <p className="lead" style={{ marginBottom: 0 }}>
          Rigorous visibility into source row counts, missing field rates,
          playtest classification, and normalization anomalies.
        </p>
      </div>

      {/* Discrepancy Reconcilation Alert */}
      <div className="alert alert-info" style={{ marginBottom: "1.75rem" }}>
        <FileText size={20} />
        <div>
          <strong>Provenance Count Reconcilation:</strong>
          <div style={{ marginTop: "0.25rem", fontSize: "0.85rem" }}>
            {report.discrepancy_explanation}
          </div>
          <div
            style={{
              marginTop: "0.5rem",
              display: "flex",
              gap: "1.5rem",
              fontFamily: "var(--font-mono)",
              fontSize: "0.8rem",
            }}
          >
            <span>
              README Stated Count:{" "}
              <strong>{report.readme_reported_count.toLocaleString()}</strong>
            </span>
            <span>
              Imported Parquet Ground Truth:{" "}
              <strong>{report.source_row_count.toLocaleString()}</strong>
            </span>
            <span>
              Unreconciled Variance:{" "}
              <strong>
                {(
                  report.readme_reported_count - report.source_row_count
                ).toLocaleString()}
              </strong>
            </span>
          </div>
        </div>
      </div>

      {/* Audit KPI Tiles */}
      <div className="metrics-grid">
        <div className="metric-tile">
          <div className="metric-label">Actual Source Rows</div>
          <div className="metric-value">
            {report.source_row_count.toLocaleString()}
          </div>
          <div className="metric-sub">
            {report.distinct_app_ids.toLocaleString()} distinct AppIDs
          </div>
        </div>

        <div className="metric-tile">
          <div className="metric-label">Duplicate AppIDs</div>
          <div className="metric-value" style={{ color: "var(--emerald)" }}>
            {report.duplicate_ids}
          </div>
          <div className="metric-sub">Zero primary key collision</div>
        </div>

        <div className="metric-tile">
          <div className="metric-label">Probable Playtests</div>
          <div className="metric-value" style={{ color: "var(--amber)" }}>
            {report.probable_playtest_count.toLocaleString()}
          </div>
          <div className="metric-sub">
            {(
              (report.probable_playtest_count / report.source_row_count) *
              100
            ).toFixed(1)}
            % of dataset
          </div>
        </div>

        <div className="metric-tile">
          <div className="metric-label">Sparse Metadata Titles</div>
          <div className="metric-value" style={{ color: "var(--rose)" }}>
            {report.sparse_records_count.toLocaleString()}
          </div>
          <div className="metric-sub">Lack genres, categories & reviews</div>
        </div>
      </div>

      {/* Field Completeness Distribution */}
      <div className="card" style={{ marginBottom: "1.75rem" }}>
        <h2>Dataset Field Completeness Rates</h2>
        <p
          style={{
            fontSize: "0.85rem",
            color: "var(--text-dim)",
            marginBottom: "1.25rem",
          }}
        >
          Valid observations per field. Incomplete records are visible and
          metric denominators reflect only valid entries.
        </p>

        <div
          style={{ display: "flex", flexDirection: "column", gap: "0.85rem" }}
        >
          {report.field_completeness.map((field) => (
            <div key={field.key} className="dist-item">
              <div className="dist-item-header">
                <span
                  className="dist-item-name"
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.4rem",
                  }}
                >
                  <CheckCircle2
                    size={13}
                    color={
                      field.percentage >= 90 ? "var(--emerald)" : "var(--amber)"
                    }
                  />
                  {field.label}
                </span>
                <span className="dist-item-stats">
                  {field.count.toLocaleString()} valid ({field.percentage}%)
                </span>
              </div>
              <div className="dist-track">
                <div
                  className="dist-fill"
                  style={{
                    width: `${Math.max(2, field.percentage)}%`,
                    backgroundColor:
                      field.percentage >= 90
                        ? "var(--emerald)"
                        : field.percentage >= 50
                          ? "var(--cyan)"
                          : "var(--amber)",
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Detailed Anomaly Breakdown */}
      <div className="charts-grid-2">
        <div className="card">
          <h2>Messy Data Classification Rules</h2>
          <div
            style={{
              fontSize: "0.85rem",
              color: "var(--text-dim)",
              display: "flex",
              flexDirection: "column",
              gap: "0.75rem",
              marginTop: "0.75rem",
            }}
          >
            <div>
              <strong style={{ color: "var(--text-main)" }}>
                Probable Playtest Detection:
              </strong>
              <div>
                Matches titles containing keywords such as <code>playtest</code>
                , <code>closed beta</code>, <code>open beta</code>, or{" "}
                <code>alpha test</code>. Filterable via Quality Mode.
              </div>
            </div>
            <div>
              <strong style={{ color: "var(--text-main)" }}>
                Ownership '0 - 0' Interval:
              </strong>
              <div>
                {report.zero_owners_count.toLocaleString()} games report 0 to 0
                owners. These are explicitly retained as a known zero-state
                rather than silently turned into missing values.
              </div>
            </div>
            <div>
              <strong style={{ color: "var(--text-main)" }}>
                Sparse Records:
              </strong>
              <div>
                {report.sparse_records_count.toLocaleString()} games contain
                empty genre lists and zero reviews. Excluded in "More Complete
                Records" mode.
              </div>
            </div>
          </div>
        </div>

        <div className="card">
          <h2>Ingestion Pipeline Quality Flags</h2>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "0.5rem",
              marginTop: "0.75rem",
            }}
          >
            {report.ingestion_warnings?.map((w, idx) => (
              <div
                key={idx}
                className="alert alert-warning"
                style={{ marginBottom: 0 }}
              >
                <AlertTriangle size={15} />
                <span>{w}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
