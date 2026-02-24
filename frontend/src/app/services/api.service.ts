import { Injectable }              from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable }              from 'rxjs';
import { map }                     from 'rxjs/operators';

export interface ApiResponse<T> {
  status:   string;
  message?: string;
  data?:    T;
}

/**
 * Generic HTTP client service with typed methods.
 * Extend this or use directly in your components.
 */
@Injectable({ providedIn: 'root' })
export class ApiService {

  constructor(private http: HttpClient) {}

  /**
   * Fetches configuration from a backend endpoint.
   * Unwraps the standard API response envelope and returns the data payload.
   */
  getConfig<T>(url: string): Observable<T> {
    return this.http.get<ApiResponse<T>>(url).pipe(
      map(resp => resp.data as T)
    );
  }

  // ── Generic typed HTTP methods ────────────────────────────────────────────

  get<T>(url: string, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.get<T>(url, options);
  }

  post<T>(url: string, body: unknown, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.post<T>(url, body, options);
  }

  put<T>(url: string, body: unknown, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.put<T>(url, body, options);
  }

  delete<T>(url: string, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.delete<T>(url, options);
  }

  patch<T>(url: string, body: unknown, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.patch<T>(url, body, options);
  }
}
