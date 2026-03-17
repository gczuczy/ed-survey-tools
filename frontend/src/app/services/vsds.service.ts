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

@Injectable({ providedIn: 'root' })
export class VsdsService {

  constructor(private api: ApiService) {}

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

  deleteZSample(id: number, zsample: number): Observable<ApiResponse<VSDSProject>> {
    return this.api.delete<ApiResponse<VSDSProject>>(`/api/vsds/projects/${id}/zsamples/${zsample}`);
  }
}
