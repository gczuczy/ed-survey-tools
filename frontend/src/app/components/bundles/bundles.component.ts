import { Component, OnInit }             from '@angular/core';
import { FormsModule }                   from '@angular/forms';
import { DatePipe }                      from '@angular/common';
import { HttpErrorResponse }             from '@angular/common/http';
import { AuthService }                   from '../../auth/auth.service';
import {
  BundlesService,
  Bundle,
  CreateBundleRequest,
} from '../../services/bundles.service';
import {
  VsdsService,
  VSDSProject,
} from '../../services/vsds.service';
import { ButtonModule }         from 'primeng/button';
import { CardModule }           from 'primeng/card';
import { TableModule }          from 'primeng/table';
import { TagModule }            from 'primeng/tag';
import { CheckboxModule }       from 'primeng/checkbox';
import { SelectModule }         from 'primeng/select';
import { MultiSelectModule }    from 'primeng/multiselect';
import { InputTextModule }      from 'primeng/inputtext';
import { MessageModule }        from 'primeng/message';
import { ConfirmPopupModule }   from 'primeng/confirmpopup';
import { DialogModule }         from 'primeng/dialog';
import { TooltipModule }        from 'primeng/tooltip';
import { ConfirmationService }  from 'primeng/api';

interface SelectOption {
  label: string;
  value: string | number;
}

@Component({
  selector:    'app-bundles',
  standalone:  true,
  imports: [
    FormsModule,
    DatePipe,
    ButtonModule,
    CardModule,
    TableModule,
    TagModule,
    CheckboxModule,
    SelectModule,
    MultiSelectModule,
    InputTextModule,
    MessageModule,
    ConfirmPopupModule,
    DialogModule,
    TooltipModule,
  ],
  providers:   [ConfirmationService],
  templateUrl: './bundles.component.html',
  styleUrl:    './bundles.component.scss',
})
export class BundlesComponent implements OnInit {
  bundles:  Bundle[] = [];
  loading = true;
  loadError: string | null = null;

  bundleBaseUrl: string | null = null;

  // Add-bundle dialog state
  addDialogVisible = false;
  newName          = '';
  newSubtype       = '';
  newAutoRegen     = false;
  newAllProjects   = false;
  newProjects:     number[] = [];
  addLoading     = false;
  addError:      string | null = null;

  subtypeOptions: SelectOption[] = [
    { label: 'Survey Points', value: 'surveypoints' },
    { label: 'Surveys',       value: 'surveys' },
  ];
  projectOptions: SelectOption[] = [];

  // Per-row action state
  generatingId:  number | null = null;
  generateError: string | null = null;
  deletingId:    number | null = null;
  deleteError:   string | null = null;
  patchingId:    number | null = null;

  constructor(
    public  authService:         AuthService,
    private bundlesService:      BundlesService,
    private vsdsService:         VsdsService,
    private confirmationService: ConfirmationService,
  ) {}

  ngOnInit(): void {
    this.bundlesService.getConfig().subscribe({
      next: cfg => { this.bundleBaseUrl = cfg.bundleBaseUrl; },
    });
    this.loadBundles();
    if (this.authService.user?.isadmin) {
      this.loadProjects();
    }
  }

  loadBundles(): void {
    this.loading   = true;
    this.loadError = null;
    this.bundlesService.listBundles().subscribe({
      next: bundles => {
        this.bundles = bundles;
        this.loading = false;
      },
      error: (err: HttpErrorResponse) => {
        this.loadError =
          err.error?.message ?? 'Failed to load bundles';
        this.loading = false;
      },
    });
  }

  private loadProjects(): void {
    this.vsdsService.listProjects().subscribe({
      next: resp => {
        this.projectOptions = (resp.data ?? []).map(
          (p: VSDSProject) => ({ label: p.name, value: p.id })
        );
      },
    });
  }

  projectsDisplay(bundle: Bundle): string {
    if (bundle.allprojects) return 'All';
    if (!bundle.projects || bundle.projects.length === 0) return '—';
    return bundle.projects.join(', ');
  }

  statusSeverity(
    status: string,
  ): 'success' | 'danger' | 'info' | 'secondary' {
    switch (status) {
      case 'ready':      return 'success';
      case 'error':      return 'danger';
      case 'generating':
      case 'queued':     return 'info';
      default:           return 'secondary';
    }
  }

  downloadUrl(bundle: Bundle): string {
    return (this.bundleBaseUrl ?? '') + bundle.filename;
  }

  openAddDialog(): void {
    this.newName        = '';
    this.newSubtype     = '';
    this.newAutoRegen   = false;
    this.newAllProjects = false;
    this.newProjects    = [];
    this.addError       = null;
    this.addDialogVisible = true;
  }

  closeAddDialog(): void {
    this.addDialogVisible = false;
  }

  submitAddBundle(): void {
    const name = this.newName.trim();
    if (!name || !this.newSubtype) return;

    this.addError   = null;
    this.addLoading = true;

    const body: CreateBundleRequest = {
      measurementtype: 'vsds',
      name,
      autoregen: this.newAutoRegen,
      vsds: {
        subtype:     this.newSubtype,
        allprojects: this.newAllProjects,
        projects:    this.newAllProjects ? [] : this.newProjects,
      },
    };

    this.bundlesService.createBundle(body).subscribe({
      next: bundle => {
        this.bundles          = [...this.bundles, bundle];
        this.addLoading       = false;
        this.addDialogVisible = false;
      },
      error: (err: HttpErrorResponse) => {
        this.addError   =
          err.error?.message ?? 'Failed to create bundle';
        this.addLoading = false;
      },
    });
  }

  generateBundle(bundle: Bundle): void {
    this.generatingId  = bundle.id;
    this.generateError = null;
    this.bundlesService.generateBundle(bundle.id).subscribe({
      next: () => {
        this.generatingId = null;
        this.loadBundles();
      },
      error: (err: HttpErrorResponse) => {
        this.generateError =
          err.error?.message ?? `Failed to queue "${bundle.name}"`;
        this.generatingId = null;
      },
    });
  }

  confirmDeleteBundle(event: Event, bundle: Bundle): void {
    this.confirmationService.confirm({
      target:  event.target as EventTarget,
      message: `Delete bundle "${bundle.name}"?`,
      icon:    'pi pi-exclamation-triangle',
      accept:  () => this.doDeleteBundle(bundle),
    });
  }

  private doDeleteBundle(bundle: Bundle): void {
    this.deletingId  = bundle.id;
    this.deleteError = null;
    this.bundlesService.deleteBundle(bundle.id).subscribe({
      next: () => {
        this.bundles    =
          this.bundles.filter(b => b.id !== bundle.id);
        this.deletingId = null;
      },
      error: (err: HttpErrorResponse) => {
        this.deleteError =
          err.error?.message ?? `Failed to delete "${bundle.name}"`;
        this.deletingId = null;
      },
    });
  }

  toggleAutoregen(bundle: Bundle, checked: boolean): void {
    this.patchingId = bundle.id;
    this.bundlesService.patchBundle(
      bundle.id, { autoregen: checked }
    ).subscribe({
      next: updated => {
        const idx = this.bundles.findIndex(b => b.id === bundle.id);
        if (idx >= 0) {
          this.bundles = [
            ...this.bundles.slice(0, idx),
            updated,
            ...this.bundles.slice(idx + 1),
          ];
        }
        this.patchingId = null;
      },
      error: () => { this.patchingId = null; },
    });
  }
}
