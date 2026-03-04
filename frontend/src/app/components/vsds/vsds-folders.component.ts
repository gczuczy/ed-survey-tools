import { Component, OnInit }           from '@angular/core';
import { FormsModule }                 from '@angular/forms';
import { DatePipe }                    from '@angular/common';
import { HttpErrorResponse }           from '@angular/common/http';
import { VsdsService, VSDSFolder }     from '../../services/vsds.service';
import { ButtonModule }                from 'primeng/button';
import { CardModule }                  from 'primeng/card';
import { InputTextModule }             from 'primeng/inputtext';
import { MessageModule }               from 'primeng/message';
import { TableModule }                 from 'primeng/table';
import { ConfirmDialogModule }         from 'primeng/confirmdialog';
import { ConfirmationService }         from 'primeng/api';

@Component({
  selector:    'app-vsds-folders',
  standalone:  true,
  imports:     [FormsModule, DatePipe, ButtonModule, CardModule, InputTextModule, MessageModule, TableModule, ConfirmDialogModule],
  providers:   [ConfirmationService],
  templateUrl: './vsds-folders.component.html',
  styleUrl:    './vsds-folders.component.scss',
})
export class VsdsFoldersComponent implements OnInit {
  folders:    VSDSFolder[] = [];
  loading   = true;
  loadError: string | null = null;

  newFolderUrl = '';
  addLoading   = false;
  addError:    string | null = null;
  addSuccess:  string | null = null;

  deletingId:  number | null = null;
  deleteError: string | null = null;

  constructor(private vsdsService: VsdsService, private confirmationService: ConfirmationService) {}

  ngOnInit(): void {
    this.loadFolders();
  }

  loadFolders(): void {
    this.loading   = true;
    this.loadError = null;
    this.vsdsService.listFolders().subscribe({
      next: (resp) => {
        this.folders = resp.data ?? [];
        this.loading = false;
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
