import {
  CohortBenchmarkResult,
  CohortComparisonResult,
  CohortFilter,
  DataQualityReport,
  DatasetManifest,
  GameCatalogResponse,
  GameRecord,
  HypotheticalProfile,
  ProfileBenchmarkResult,
} from "../types";

const BASE_URL = "/api/v1";

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errMsg = `Request failed: ${res.statusText}`;
    try {
      const errJson = await res.json();
      if (errJson && errJson.error) {
        errMsg = errJson.error;
      }
    } catch {
      // ignore json parse error
    }
    throw new Error(errMsg);
  }
  return res.json();
}

export const api = {
  async getMeta(): Promise<DatasetManifest> {
    const res = await fetch(`${BASE_URL}/meta`);
    return handleResponse<DatasetManifest>(res);
  },

  async getOverview(): Promise<CohortBenchmarkResult> {
    const res = await fetch(`${BASE_URL}/overview`);
    return handleResponse<CohortBenchmarkResult>(res);
  },

  async benchmarkCohort(filter: CohortFilter): Promise<CohortBenchmarkResult> {
    const res = await fetch(`${BASE_URL}/cohorts/benchmark`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(filter),
    });
    return handleResponse<CohortBenchmarkResult>(res);
  },

  async compareCohorts(
    cohorts: CohortFilter[],
  ): Promise<CohortComparisonResult> {
    const res = await fetch(`${BASE_URL}/cohorts/compare`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ cohorts }),
    });
    return handleResponse<CohortComparisonResult>(res);
  },

  async benchmarkProfile(
    profile: HypotheticalProfile,
  ): Promise<ProfileBenchmarkResult> {
    const res = await fetch(`${BASE_URL}/profiles/benchmark`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(profile),
    });
    return handleResponse<ProfileBenchmarkResult>(res);
  },

  async getCatalog(
    params: Record<string, string | number | boolean | undefined>,
  ): Promise<GameCatalogResponse> {
    const query = new URLSearchParams();
    for (const [key, val] of Object.entries(params)) {
      if (val !== undefined && val !== "") {
        query.set(key, String(val));
      }
    }
    const res = await fetch(`${BASE_URL}/games?${query.toString()}`);
    return handleResponse<GameCatalogResponse>(res);
  },

  async getGame(appID: string): Promise<GameRecord> {
    const res = await fetch(`${BASE_URL}/games/${encodeURIComponent(appID)}`);
    return handleResponse<GameRecord>(res);
  },

  async getDataQuality(): Promise<DataQualityReport> {
    const res = await fetch(`${BASE_URL}/data-quality`);
    return handleResponse<DataQualityReport>(res);
  },

  getCohortExportUrl(filter: CohortFilter): string {
    const query = new URLSearchParams();
    if (filter.genres && filter.genres.length > 0) {
      query.set("genre", filter.genres.join(","));
    }
    if (filter.quality_mode) {
      query.set("quality_mode", filter.quality_mode);
    }
    return `${BASE_URL}/export/cohort?${query.toString()}`;
  },

  getGamesExportUrl(
    params: Record<string, string | number | boolean | undefined>,
  ): string {
    const query = new URLSearchParams();
    for (const [key, val] of Object.entries(params)) {
      if (val !== undefined && val !== "") {
        query.set(key, String(val));
      }
    }
    return `${BASE_URL}/export/games?${query.toString()}`;
  },
};
