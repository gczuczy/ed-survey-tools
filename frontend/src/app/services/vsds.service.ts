import { Injectable }              from '@angular/core';
import { Observable }              from 'rxjs';
import { ApiService, ApiResponse } from './api.service';

/**
 * Mirrors the backend db.VSDSFolder struct returned by /api/vsds/folders.
 */
export interface VSDSFolder {
  id:           number;
  name:         string;
  gcpid:        string;
  received_at?: string;
  started_at?:  string;
  finished_at?: string;
  in_progress?: number;
  finished?:    number;
  failed?:      number;
}

/**
 * Mirrors db.FolderProcessingSummary (stats block of extraction response).
 */
export interface VSDSExtractionStats {
  received_at:      string;
  started_at?:      string;
  finished_at?:     string;
  documents_total:  number;
  sheets_total:     number;
  sheets_success:   number;
  sheets_failed:    number;
  surveys_ingested: number;
  points_ingested:  number;
  cmdrs_count:      number;
}

/**
 * One failing tab within a document in the extraction summary.
 */
export interface VSDSExtractionSheet {
  id:      number;
  name:    string;
  message: string;
}

/**
 * One document with errors in the extraction summary.
 */
export interface VSDSExtractionDocument {
  id:           number;
  gcpid:        string;
  name:         string;
  content_type: string;
  error_count:  number;
  sheets:       VSDSExtractionSheet[];
}

/**
 * Mirrors FolderExtractionSummaryResp returned by
 * GET /api/vsds/folders/{id}/extraction.
 */
export interface VSDSExtractionSummary {
  folder_name: string;
  stats:       VSDSExtractionStats;
  documents:   VSDSExtractionDocument[];
}

/**
 * Mirrors the backend db.VSDSProject struct returned by /api/vsds/projects.
 */
export interface VSDSProject {
  id:       number;
  name:     string;
  zsamples: number[];
}

/**
 * One cell-value assertion used to fingerprint a sheet variant.
 * col and row are 0-indexed (as stored in the DB).
 * The UI converts col→letter and row→1-indexed for display.
 */
export interface VSDSSheetVariantCheck {
  id:    number;
  col:   number;
  row:   number;
  value: string;
}

/**
 * Mirrors db.DBSheetVariant returned by
 * /api/vsds/projects/{id}/variants.
 * All column/row fields are 0-indexed.
 */
export interface VSDSSheetVariant {
  id:                 number;
  project_id:         number;
  name:               string;
  header_row:         number;
  sysname_column:     number;
  zsample_column:     number;
  syscount_column:    number;
  maxdistance_column: number;
  checks:             VSDSSheetVariantCheck[];
}

/**
 * One check evaluation result returned by the validation endpoint.
 */
export interface ValidationCheckResult {
  col:      number;
  row:      number;
  expected: string;
  actual:   string;
  ok:       boolean;
}

/**
 * One spreadsheet tab with its cell data and check results.
 */
export interface ValidationTabResult {
  name:    string;
  rows:    string[][];
  checks:  ValidationCheckResult[];
  matched: boolean;
}

/**
 * Top-level payload from POST .../variants/validate.
 */
export interface ValidationResponse {
  tabs: ValidationTabResult[];
}

/**
 * Request body for the validation endpoint.
 */
export interface VariantValidationRequest {
  url: string;
  variant: {
    name:               string;
    header_row:         number;
    sysname_column:     number;
    zsample_column:     number;
    syscount_column:    number;
    maxdistance_column: number;
    checks: Array<{ col: number; row: number; value: string }>;
  };
}

/**
 * Mirrors VSDSConfigResponse returned by GET /api/vsds/config.
 */
export interface VSDSConfig {
  gcp_client_email: string;
}

/**
 * Mirrors db.CMDRContribution returned by
 * GET /api/vsds/contribution.
 * coldev_* are omitted (undefined) when there are no multi-point
 * surveys to compute a deviation from.
 */
export interface VSDSContribution {
  surveys:     number;
  points:      number;
  coldev_min?: number;
  coldev_avg?: number;
  coldev_max?: number;
}

/**
 * One errored tab within a document in the user's error list.
 * Mirrors ContribErrorSheetResp.
 */
export interface VSDSContribErrorSheet {
  sheet_name: string;
  message:    string;
}

/**
 * One document with its errored tabs, grouped by processing run.
 * Mirrors ContribErrorDocResp.
 */
export interface VSDSContribErrorDoc {
  doc_id:      number;
  doc_name:    string;
  received_at: string;
  error_count: number;
  sheets:      VSDSContribErrorSheet[];
}

@Injectable({ providedIn: 'root' })
export class VsdsService {

  constructor(private api: ApiService) {}

  getConfig(): Observable<ApiResponse<VSDSConfig>> {
    return this.api.get<ApiResponse<VSDSConfig>>('/api/vsds/config');
  }

  listFolders(): Observable<ApiResponse<VSDSFolder[]>> {
    return this.api.get<ApiResponse<VSDSFolder[]>>('/api/vsds/folders');
  }

  addFolder(url: string): Observable<ApiResponse<VSDSFolder>> {
    return this.api.post<ApiResponse<VSDSFolder>>('/api/vsds/folders', { url });
  }

  deleteFolder(id: number): Observable<ApiResponse<null>> {
    return this.api.delete<ApiResponse<null>>(`/api/vsds/folders/${id}`);
  }

  processFolder(id: number): Observable<ApiResponse<null>> {
    return this.api.post<ApiResponse<null>>(`/api/vsds/folders/${id}/process`, null);
  }

  getFolderExtractionSummary(
    id: number,
  ): Observable<ApiResponse<VSDSExtractionSummary>> {
    return this.api.get<ApiResponse<VSDSExtractionSummary>>(
      `/api/vsds/folders/${id}/extraction`,
    );
  }

  listProjects(): Observable<ApiResponse<VSDSProject[]>> {
    return this.api.get<ApiResponse<VSDSProject[]>>('/api/vsds/projects');
  }

  addProject(name: string): Observable<ApiResponse<VSDSProject>> {
    return this.api.put<ApiResponse<VSDSProject>>('/api/vsds/projects', { name });
  }

  getProject(id: number): Observable<ApiResponse<VSDSProject>> {
    return this.api.get<ApiResponse<VSDSProject>>(`/api/vsds/projects/${id}`);
  }

  setZSamples(id: number, zsamples: number[]): Observable<ApiResponse<VSDSProject>> {
    return this.api.post<ApiResponse<VSDSProject>>(`/api/vsds/projects/${id}/zsamples`, zsamples);
  }

  addZSample(id: number, zsample: number): Observable<ApiResponse<VSDSProject>> {
    return this.api.put<ApiResponse<VSDSProject>>(`/api/vsds/projects/${id}/zsamples/${zsample}`, null);
  }

  deleteZSample(
    id: number,
    zsample: number,
  ): Observable<ApiResponse<VSDSProject>> {
    return this.api.delete<ApiResponse<VSDSProject>>(
      `/api/vsds/projects/${id}/zsamples/${zsample}`);
  }

  listVariants(
    projectId: number,
  ): Observable<ApiResponse<VSDSSheetVariant[]>> {
    return this.api.get<ApiResponse<VSDSSheetVariant[]>>(
      `/api/vsds/projects/${projectId}/variants`);
  }

  addVariant(
    projectId: number,
    body: {
      name: string; header_row: number;
      sysname_column: number; zsample_column: number;
      syscount_column: number; maxdistance_column: number;
    },
  ): Observable<ApiResponse<VSDSSheetVariant>> {
    return this.api.put<ApiResponse<VSDSSheetVariant>>(
      `/api/vsds/projects/${projectId}/variants`, body);
  }

  updateVariant(
    projectId: number,
    variantId: number,
    body: {
      name: string; header_row: number;
      sysname_column: number; zsample_column: number;
      syscount_column: number; maxdistance_column: number;
    },
  ): Observable<ApiResponse<VSDSSheetVariant>> {
    return this.api.post<ApiResponse<VSDSSheetVariant>>(
      `/api/vsds/projects/${projectId}/variants/${variantId}`,
      body);
  }

  deleteVariant(
    projectId: number,
    variantId: number,
  ): Observable<ApiResponse<null>> {
    return this.api.delete<ApiResponse<null>>(
      `/api/vsds/projects/${projectId}/variants/${variantId}`);
  }

  addVariantCheck(
    projectId: number,
    variantId: number,
    body: { col: number; row: number; value: string },
  ): Observable<ApiResponse<VSDSSheetVariant>> {
    return this.api.put<ApiResponse<VSDSSheetVariant>>(
      `/api/vsds/projects/${projectId}/variants/${variantId}/checks`,
      body);
  }

  deleteVariantCheck(
    projectId: number,
    variantId: number,
    checkId: number,
  ): Observable<ApiResponse<VSDSSheetVariant>> {
    return this.api.delete<ApiResponse<VSDSSheetVariant>>(
      `/api/vsds/projects/${projectId}/variants/${variantId}` +
      `/checks/${checkId}`);
  }

  validateVariant(
    projectId: number,
    body: VariantValidationRequest,
  ): Observable<ApiResponse<ValidationResponse>> {
    return this.api.post<ApiResponse<ValidationResponse>>(
      `/api/vsds/projects/${projectId}/variants/validate`,
      body);
  }

  getMyContribution(): Observable<ApiResponse<VSDSContribution>> {
    return this.api.get<ApiResponse<VSDSContribution>>(
      '/api/vsds/contribution');
  }

  getMyContributionErrors(): Observable<ApiResponse<VSDSContribErrorDoc[]>> {
    return this.api.get<ApiResponse<VSDSContribErrorDoc[]>>(
      '/api/vsds/contribution/errors');
  }
}
