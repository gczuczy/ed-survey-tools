import { Injectable }              from '@angular/core';
import { Observable }              from 'rxjs';
import { map }                     from 'rxjs/operators';
import { ApiService, ApiResponse } from './api.service';

export interface Bundle {
  id:              number;
  measurementtype: string;
  name:            string;
  filename:        string;
  generated_at:    string | null;
  autoregen:       boolean;
  status:          'pending' | 'queued' | 'generating' | 'ready' | 'error';
  error_message:   string | null;
  // VSDS-specific
  subtype:     string | null;
  allprojects: boolean | null;
  projects:    string[] | null;
}

export interface CreateBundleRequest {
  measurementtype: string;
  name:            string;
  autoregen:       boolean;
  vsds?: {
    subtype:     string;
    allprojects: boolean;
    projects:    number[];
  };
}

export interface PatchBundleRequest {
  name?:      string;
  autoregen?: boolean;
  vsds?: {
    subtype?:     string;
    allprojects?: boolean;
    projects?:    number[];
  };
}

export interface AppConfig {
  bundleBaseUrl: string;
}

@Injectable({ providedIn: 'root' })
export class BundlesService {
  private _bundleBaseUrl: string | null = null;

  constructor(private api: ApiService) {}

  get bundleBaseUrl(): string | null {
    return this._bundleBaseUrl;
  }

  getConfig(): Observable<AppConfig> {
    return this.api.get<ApiResponse<AppConfig>>('/api/config').pipe(
      map(resp => {
        const cfg = resp.data!;
        this._bundleBaseUrl = cfg.bundleBaseUrl;
        return cfg;
      })
    );
  }

  listBundles(): Observable<Bundle[]> {
    return this.api.get<ApiResponse<Bundle[]>>('/api/bundles').pipe(
      map(resp => resp.data ?? [])
    );
  }

  getBundle(id: number): Observable<Bundle> {
    return this.api.get<ApiResponse<Bundle>>(`/api/bundles/${id}`).pipe(
      map(resp => resp.data!)
    );
  }

  createBundle(body: CreateBundleRequest): Observable<Bundle> {
    return this.api.put<ApiResponse<Bundle>>('/api/bundles', body).pipe(
      map(resp => resp.data!)
    );
  }

  deleteBundle(id: number): Observable<void> {
    return this.api.delete<ApiResponse<null>>(
      `/api/bundles/${id}`
    ).pipe(map(() => undefined));
  }

  generateBundle(id: number): Observable<void> {
    return this.api.post<ApiResponse<null>>(
      `/api/bundles/${id}/generate`, null
    ).pipe(map(() => undefined));
  }

  patchBundle(id: number, body: PatchBundleRequest): Observable<Bundle> {
    return this.api.patch<ApiResponse<Bundle>>(
      `/api/bundles/${id}`, body
    ).pipe(map(resp => resp.data!));
  }
}
