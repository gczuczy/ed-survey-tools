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
}
