import { Injectable }    from '@angular/core';
import { BehaviorSubject } from 'rxjs';

/**
 * Allows components to override the last breadcrumb label with a
 * dynamic value (e.g. an entity name loaded from the API).
 * The navbar subscribes to dynamicLabel$ and uses the value when
 * building the breadcrumb trail for the current route.
 */
@Injectable({ providedIn: 'root' })
export class BreadcrumbService {
  private labelSubject = new BehaviorSubject<string | null>(null);
  readonly dynamicLabel$ = this.labelSubject.asObservable();

  set(label: string): void  { this.labelSubject.next(label); }
  clear(): void             { this.labelSubject.next(null);  }
}
