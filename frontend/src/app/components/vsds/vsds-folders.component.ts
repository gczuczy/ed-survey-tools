import { Component, OnInit, OnDestroy } from '@angular/core';
import { interval, Subscription }      from 'rxjs';
import { FormsModule }                 from '@angular/forms';
import { DatePipe }                    from '@angular/common';
import { RouterLink }                  from '@angular/router';
import { HttpErrorResponse }           from '@angular/common/http';
import { VsdsService, VSDSFolder }     from '../../services/vsds.service';
import { ButtonModule }                from 'primeng/button';
import { CardModule }                  from 'primeng/card';
import { InputTextModule }             from 'primeng/inputtext';
import { MessageModule }               from 'primeng/message';
import { TableModule }                 from 'primeng/table';
import { ConfirmDialogModule }         from 'primeng/confirmdialog';
import { TooltipModule }              from 'primeng/tooltip';
import { ConfirmationService }         from 'primeng/api';

@Component({
  selector:    'app-vsds-folders',
  standalone:  true,
  imports:     [FormsModule, DatePipe, RouterLink, ButtonModule, CardModule, InputTextModule, MessageModule, TableModule, ConfirmDialogModule, TooltipModule],
  providers:   [ConfirmationService],
  templateUrl: './vsds-folders.component.html',
  styleUrl:    './vsds-folders.component.scss',
})
export class VsdsFoldersComponent implements OnInit, OnDestroy {
  gcpClientEmail = '';

  folders:    VSDSFolder[] = [];
  loading   = true;
  loadError: string | null = null;

  newFolderUrl = '';
  addLoading   = false;
  addError:    string | null = null;
  addSuccess:  string | null = null;

  deletingId:  number | null = null;
  deleteError: string | null = null;

  processingId:  number | null = null;
  processError:  string | null = null;

  private pollSub: Subscription | null = null;

  constructor(private vsdsService: VsdsService, private confirmationService: ConfirmationService) {}

  ngOnInit(): void {
    this.vsdsService.getConfig().subscribe({
      next: (resp) => { this.gcpClientEmail = resp.data?.gcp_client_email ?? ''; },
    });
    this.loadFolders();
  }

  ngOnDestroy(): void {
    this.stopPolling();
  }

  loadFolders(): void {
    this.loading   = true;
    this.loadError = null;
    this.vsdsService.listFolders().subscribe({
      next: (resp) => {
        this.folders = resp.data ?? [];
        this.loading = false;
        this.updatePolling();
      },
      error: (err: HttpErrorResponse) => {
        this.loadError = err.error?.message ?? 'Failed to load folders';
        this.loading   = false;
      },
    });
  }

  addFolder(): void {
    const url = this.newFolderUrl.trim();
    if (!url) return;

    this.addError   = null;
    this.addSuccess = null;
    this.addLoading = true;

    this.vsdsService.addFolder(url).subscribe({
      next: (resp) => {
        this.folders      = [...this.folders, resp.data!];
        this.newFolderUrl = '';
        this.addLoading   = false;
        this.addSuccess   = `Folder "${resp.data!.name}" added successfully.`;
      },
      error: (err: HttpErrorResponse) => {
        this.addError   = err.error?.message ?? 'Failed to add folder';
        this.addLoading = false;
      },
    });
  }

  canProcess(folder: VSDSFolder): boolean {
    return !folder.received_at || folder.finished_at != null;
  }

  processFolder(folder: VSDSFolder): void {
    if (!this.canProcess(folder)) return;
    this.processingId = folder.id;
    this.processError = null;

    this.vsdsService.processFolder(folder.id).subscribe({
      next: () => {
        this.processingId = null;
        this.loadFolders();
      },
      error: (err: HttpErrorResponse) => {
        this.processError = err.error?.message ?? `Failed to queue processing for "${folder.name}"`;
        this.processingId = null;
      },
    });
  }

  private isAnyProcessing(): boolean {
    return this.folders.some(
      f => !!f.received_at && !f.finished_at
    );
  }

  private updatePolling(): void {
    if (this.isAnyProcessing()) {
      this.startPolling();
    } else {
      this.stopPolling();
    }
  }

  private startPolling(): void {
    if (this.pollSub) return;
    this.pollSub = interval(10000).subscribe(() => {
      this.vsdsService.listFolders().subscribe({
        next: (resp) => {
          this.folders = resp.data ?? [];
          if (!this.isAnyProcessing()) {
            this.stopPolling();
          }
        },
      });
    });
  }

  private stopPolling(): void {
    if (this.pollSub) {
      this.pollSub.unsubscribe();
      this.pollSub = null;
    }
  }

  deleteFolder(folder: VSDSFolder): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to delete folder "${folder.name}"?`,
      header:  'Confirm Deletion',
      icon:    'pi pi-exclamation-triangle',
      accept:  () => {
        this.deletingId  = folder.id;
        this.deleteError = null;

        this.vsdsService.deleteFolder(folder.id).subscribe({
          next: () => {
            this.folders    = this.folders.filter(f => f.id !== folder.id);
            this.deletingId = null;
          },
          error: (err: HttpErrorResponse) => {
            this.deleteError = err.error?.message ?? `Failed to delete folder "${folder.name}"`;
            this.deletingId  = null;
          },
        });
      },
    });
  }
}
