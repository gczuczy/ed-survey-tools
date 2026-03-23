import { Injectable }              from '@angular/core';
import { Observable }              from 'rxjs';
import { map }                     from 'rxjs/operators';
import { ApiService, ApiResponse } from './api.service';

/**
 * Mirrors the backend db.User struct returned by /api/cmdrs.
 */
export interface Cmdr {
  id:         number;
  name:       string;
  customerid: number | null;
  isowner:    boolean;
  isadmin:    boolean;
}

@Injectable({ providedIn: 'root' })
export class AdminService {

  constructor(private api: ApiService) {}

  listCmdrs(): Observable<Cmdr[]> {
    return this.api
      .get<ApiResponse<Cmdr[]>>('/api/cmdrs')
      .pipe(map(r => r.data ?? []));
  }

  setCmdrAdmin(id: number, isAdmin: boolean): Observable<Cmdr> {
    return this.api
      .patch<ApiResponse<Cmdr>>(
        `/api/cmdrs/${id}`,
        { isadmin: isAdmin },
      )
      .pipe(map(r => r.data as Cmdr));
  }
}
