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
