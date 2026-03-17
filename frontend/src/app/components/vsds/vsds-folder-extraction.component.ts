import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute }              from '@angular/router';
import { RouterLink }                  from '@angular/router';
import { DatePipe }                    from '@angular/common';
import { HttpErrorResponse }           from '@angular/common/http';
import {
  VsdsService,
  VSDSExtractionSummary,
} from '../../services/vsds.service';
import { BreadcrumbService }           from '../../services/breadcrumb.service';
import { ButtonModule }         from 'primeng/button';
import { CardModule }           from 'primeng/card';
import { MessageModule }        from 'primeng/message';
import { TableModule }          from 'primeng/table';
import { TagModule }            from 'primeng/tag';

@Component({
  selector:    'app-vsds-folder-extraction',
  standalone:  true,
  imports: [
    DatePipe,
    RouterLink,
    ButtonModule,
    CardModule,
    MessageModule,
    TableModule,
    TagModule,
  ],
  templateUrl: './vsds-folder-extraction.component.html',
  styleUrl:    './vsds-folder-extraction.component.scss',
})
export class VsdsFolderExtractionComponent implements OnInit, OnDestroy {
  loading   = true;
  loadError: string | null = null;
  data:      VSDSExtractionSummary | null = null;


  constructor(
    private route:             ActivatedRoute,
    private vsdsService:       VsdsService,
    private breadcrumbService: BreadcrumbService,
  ) {}

  ngOnInit(): void {
    const id = Number(this.route.snapshot.paramMap.get('id'));
    if (!id || id <= 0) {
      this.loadError = 'Invalid folder ID';
      this.loading   = false;
      return;
    }
    this.vsdsService.getFolderExtractionSummary(id).subscribe({
      next: (resp) => {
        this.data    = resp.data ?? null;
        this.loading = false;
        if (this.data) {
          this.breadcrumbService.set(this.data.folder_name);
        }
      },
      error: (err: HttpErrorResponse) => {
        this.loadError =
          err.error?.message ?? 'Failed to load extraction summary';
        this.loading = false;
      },
    });
  }

  ngOnDestroy(): void {
    this.breadcrumbService.clear();
  }

  get processingDuration(): string {
    const s = this.data?.stats;
    if (!s?.started_at)  return '—';
    if (!s?.finished_at) return 'In progress\u2026';

    const secs = Math.floor(
      (new Date(s.finished_at).getTime() -
       new Date(s.started_at).getTime()) / 1000,
    );
    const h = Math.floor(secs / 3600);
    const m = Math.floor((secs % 3600) / 60);
    const sec = secs % 60;
    return h > 0 ? `${h}h ${m}m ${sec}s` : `${m}m ${sec}s`;
  }

  driveUrl(gcpid: string): string {
    return `https://drive.google.com/open?id=${gcpid}`;
  }
}
