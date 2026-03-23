import { Component, OnInit, OnDestroy } from '@angular/core';
import { FormsModule }                from '@angular/forms';
import { HttpErrorResponse }          from '@angular/common/http';
import { ActivatedRoute, Router }     from '@angular/router';
import { Subscription, forkJoin }     from 'rxjs';
import { AuthService }                from '../../auth/auth.service';
import { BreadcrumbService }          from '../../services/breadcrumb.service';
import { VsdsService, VSDSProject,
         VSDSSheetVariant,
         VSDSSheetVariantCheck,
         ValidationCheckResult,
         ValidationTabResult,
         ValidationResponse }
  from '../../services/vsds.service';
import { ButtonModule }               from 'primeng/button';
import { CardModule }                 from 'primeng/card';
import { InputTextModule }            from 'primeng/inputtext';
import { InputNumberModule }          from 'primeng/inputnumber';
import { MessageModule }              from 'primeng/message';
import { TextareaModule }             from 'primeng/textarea';
import { TabsModule }                 from 'primeng/tabs';
import { TooltipModule }              from 'primeng/tooltip';
import { ConfirmPopupModule }         from 'primeng/confirmpopup';
import { ConfirmationService }        from 'primeng/api';
import { PopoverModule }              from 'primeng/popover';
import { DialogModule }               from 'primeng/dialog';

@Component({
  selector:    'app-vsds-projects',
  standalone:  true,
  imports:     [
    FormsModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    InputNumberModule,
    MessageModule,
    TextareaModule,
    TabsModule,
    TooltipModule,
    ConfirmPopupModule,
    PopoverModule,
    DialogModule,
  ],
  providers:   [ConfirmationService],
  templateUrl: './vsds-projects.component.html',
  styleUrl:    './vsds-projects.component.scss',
})
export class VsdsProjectsComponent implements OnInit, OnDestroy {
  projects:        VSDSProject[] = [];
  loadingProjects  = true;
  loadError:       string | null = null;

  selectedProject: VSDSProject | null = null;

  // Admin: add project
  newProjectName   = '';
  addProjectLoad   = false;
  addProjectError: string | null = null;

  // Admin: add single zsample
  newZSample:      number | null = null;
  addZSampleLoad   = false;
  addZSampleError: string | null = null;

  // Admin: delete zsample
  deletingZSample: number | null = null;
  delZSampleError: string | null = null;

  // Admin: bulk replace zsamples
  bulkZSamples     = '';
  bulkLoad         = false;
  bulkError:       string | null = null;

  // ── Sheet variants ────────────────────────────────────────────
  variants:         VSDSSheetVariant[] = [];
  loadingVariants   = false;
  variantLoadError: string | null = null;

  // which tab is active ('new' | variant.id as number)
  activeVariantTab: string | number = 'new';

  deletingVariantId:  number | null = null;
  deleteVariantError: string | null = null;

  // ── Dialog state ──────────────────────────────────────────────
  dialogVisible   = false;
  dialogMode:       'new' | 'edit' = 'new';
  dialogVariantId:  number | null = null;

  dialogDraft: {
    name:               string;
    header_row:         number;   // 1-indexed display value
    sysname_column:     string;   // spreadsheet letter
    zsample_column:     string;
    syscount_column:    string;
    maxdistance_column: string;
  } = {
    name: '', header_row: 1,
    sysname_column: 'A', zsample_column: 'B',
    syscount_column: 'C', maxdistance_column: 'D',
  };

  // Local check list (id is undefined for new checks).
  dialogChecks: Array<{
    id?:   number;
    col:   number;
    row:   number;
    value: string;
  }> = [];
  // IDs of checks that existed in DB when dialog was opened.
  dialogOriginalCheckIds: Set<number> = new Set();

  dialogNewCheck = { col: 'A', row: 1, value: '' };
  dialogNewCheckError: string | null = null;

  saveDialogLoading = false;
  saveDialogError:  string | null = null;

  // ── Validation state ──────────────────────────────────────────
  validateUrl     = '';
  validateLoading = false;
  validateResult: ValidationResponse | null = null;
  validateError:  string | null = null;
  validateActiveTab = 0;

  private pendingProjectId: number | null = null;
  private paramSub?:        Subscription;

  constructor(
    public  authService:         AuthService,
    private vsdsService:         VsdsService,
    private confirmationService: ConfirmationService,
    private route:               ActivatedRoute,
    private router:              Router,
    private breadcrumbService:   BreadcrumbService,
  ) {}

  ngOnInit(): void {
    this.paramSub = this.route.paramMap.subscribe(params => {
      const raw = params.get('id');
      if (!raw) return;
      const id = parseInt(raw, 10);
      if (isNaN(id) || id <= 0) return;
      const project = this.projects.find(p => p.id === id);
      if (project) {
        this.applyProject(project);
      } else {
        this.pendingProjectId = id;
      }
    });
    this.loadProjects();
  }

  ngOnDestroy(): void {
    this.paramSub?.unsubscribe();
    this.breadcrumbService.clear();
  }

  get isAdmin(): boolean {
    return this.authService.user?.isadmin ?? false;
  }

  get sortedZSamples(): number[] {
    if (!this.selectedProject) return [];
    return [...this.selectedProject.zsamples].sort((a, b) => a - b);
  }

  loadProjects(): void {
    this.loadingProjects = true;
    this.loadError = null;
    this.vsdsService.listProjects().subscribe({
      next: (resp) => {
        this.projects = resp.data ?? [];
        this.loadingProjects = false;
        if (this.pendingProjectId !== null) {
          const p = this.projects.find(
            p => p.id === this.pendingProjectId);
          this.pendingProjectId = null;
          if (p) this.applyProject(p);
        }
      },
      error: (err: HttpErrorResponse) => {
        this.loadError =
          err.error?.message ?? 'Failed to load projects';
        this.loadingProjects = false;
      },
    });
  }

  selectProject(project: VSDSProject): void {
    this.router.navigate(['/vsds/projects', project.id]);
  }

  private applyProject(project: VSDSProject): void {
    this.breadcrumbService.set(project.name);
    this.selectedProject = project;
    this.addZSampleError = null;
    this.delZSampleError = null;
    this.bulkError       = null;
    this.newZSample      = null;
    this.bulkZSamples    = '';
    if (this.isAdmin) {
      this.loadVariants(project.id);
    }
  }

  addProject(): void {
    const name = this.newProjectName.trim();
    if (name.length < 5) return;
    this.addProjectLoad  = true;
    this.addProjectError = null;
    this.vsdsService.addProject(name).subscribe({
      next: (resp) => {
        this.projects       = [...this.projects, resp.data!];
        this.newProjectName = '';
        this.addProjectLoad = false;
      },
      error: (err: HttpErrorResponse) => {
        this.addProjectError =
          err.error?.message ?? 'Failed to add project';
        this.addProjectLoad  = false;
      },
    });
  }

  addZSample(): void {
    if (this.newZSample === null || !this.selectedProject) return;
    this.addZSampleLoad  = true;
    this.addZSampleError = null;
    this.vsdsService
      .addZSample(this.selectedProject.id, this.newZSample)
      .subscribe({
        next: (resp) => {
          this.updateProject(resp.data!);
          this.newZSample     = null;
          this.addZSampleLoad = false;
        },
        error: (err: HttpErrorResponse) => {
          this.addZSampleError =
            err.error?.message ?? 'Failed to add ZSample';
          this.addZSampleLoad  = false;
        },
      });
  }

  deleteZSample(zsample: number): void {
    if (!this.selectedProject) return;
    this.deletingZSample = zsample;
    this.delZSampleError = null;
    this.vsdsService
      .deleteZSample(this.selectedProject.id, zsample)
      .subscribe({
        next: (resp) => {
          this.updateProject(resp.data!);
          this.deletingZSample = null;
        },
        error: (err: HttpErrorResponse) => {
          this.delZSampleError =
            err.error?.message ?? 'Failed to delete ZSample';
          this.deletingZSample = null;
        },
      });
  }

  setBulkZSamples(): void {
    if (!this.selectedProject || !this.bulkZSamples.trim()) return;
    const values = this.bulkZSamples
      .split(/[\s,]+/)
      .map(s => parseInt(s.trim(), 10))
      .filter(n => !isNaN(n));
    if (values.length === 0) return;
    this.bulkLoad  = true;
    this.bulkError = null;
    this.vsdsService
      .setZSamples(this.selectedProject.id, values)
      .subscribe({
        next: (resp) => {
          this.updateProject(resp.data!);
          this.bulkZSamples = '';
          this.bulkLoad     = false;
        },
        error: (err: HttpErrorResponse) => {
          this.bulkError =
            err.error?.message ?? 'Failed to replace ZSamples';
          this.bulkLoad  = false;
        },
      });
  }

  loadVariants(projectId: number, selectId?: number): void {
    this.loadingVariants  = true;
    this.variantLoadError = null;
    this.variants = [];
    this.vsdsService.listVariants(projectId).subscribe({
      next: (resp) => {
        this.variants        = resp.data ?? [];
        this.loadingVariants = false;
        if (selectId !== undefined &&
            this.variants.some(v => v.id === selectId)) {
          this.activeVariantTab = selectId;
        } else {
          this.activeVariantTab = this.variants.length > 0
            ? this.variants[0].id
            : 'new';
        }
      },
      error: (err: HttpErrorResponse) => {
        this.variantLoadError =
          err.error?.message ?? 'Failed to load variants';
        this.loadingVariants  = false;
      },
    });
  }

  /** Convert 0-indexed column number to spreadsheet letter(s). */
  colToLetter(n: number): string {
    let result = '';
    let col = n;
    do {
      result = String.fromCharCode(65 + (col % 26)) + result;
      col = Math.floor(col / 26) - 1;
    } while (col >= 0);
    return result;
  }

  /**
   * Convert spreadsheet column letter(s) to 0-indexed number.
   * Returns -1 if the input is not a valid column reference.
   */
  letterToCol(s: string): number {
    const upper = s.trim().toUpperCase();
    if (!/^[A-Z]+$/.test(upper)) return -1;
    let result = 0;
    for (let i = 0; i < upper.length; i++) {
      result = result * 26 + (upper.charCodeAt(i) - 64);
    }
    return result - 1;
  }

  confirmDeleteVariant(event: Event, variantId: number): void {
    this.confirmationService.confirm({
      target:      event.target as EventTarget,
      message:     'Delete this variant and all its checks?',
      icon:        'pi pi-exclamation-triangle',
      acceptLabel: 'Delete',
      rejectLabel: 'Cancel',
      accept: () => this.deleteVariant(variantId),
    });
  }

  deleteVariant(variantId: number): void {
    if (!this.selectedProject) return;
    this.deletingVariantId  = variantId;
    this.deleteVariantError = null;
    this.vsdsService.deleteVariant(
      this.selectedProject.id, variantId,
    ).subscribe({
      next: () => {
        this.variants = this.variants.filter(
          v => v.id !== variantId);
        this.deletingVariantId = null;
        this.activeVariantTab  = this.variants.length > 0
          ? this.variants[0].id
          : 'new';
      },
      error: (err: HttpErrorResponse) => {
        this.deleteVariantError =
          err.error?.message ?? 'Failed to delete variant';
        this.deletingVariantId = null;
      },
    });
  }

  // ── Dialog ────────────────────────────────────────────────────

  openEditDialog(v: VSDSSheetVariant): void {
    this.dialogMode      = 'edit';
    this.dialogVariantId = v.id;
    this.dialogDraft = {
      name:               v.name,
      header_row:         v.header_row + 1,
      sysname_column:     this.colToLetter(v.sysname_column),
      zsample_column:     this.colToLetter(v.zsample_column),
      syscount_column:    this.colToLetter(v.syscount_column),
      maxdistance_column: this.colToLetter(v.maxdistance_column),
    };
    this.dialogChecks = v.checks.map(c => ({
      id: c.id, col: c.col, row: c.row, value: c.value,
    }));
    this.dialogOriginalCheckIds = new Set(v.checks.map(c => c.id));
    this.resetDialogMeta();
    this.dialogVisible = true;
  }

  openNewDialog(): void {
    this.dialogMode      = 'new';
    this.dialogVariantId = null;
    this.dialogDraft = {
      name: '', header_row: 1,
      sysname_column: 'A', zsample_column: 'B',
      syscount_column: 'C', maxdistance_column: 'D',
    };
    this.dialogChecks           = [];
    this.dialogOriginalCheckIds = new Set();
    this.resetDialogMeta();
    this.dialogVisible = true;
  }

  private resetDialogMeta(): void {
    this.dialogNewCheck      = { col: 'A', row: 1, value: '' };
    this.dialogNewCheckError = null;
    this.saveDialogLoading   = false;
    this.saveDialogError     = null;
    this.validateUrl         = '';
    this.validateResult      = null;
    this.validateError       = null;
    this.validateActiveTab   = 0;
  }

  closeDialog(): void {
    this.dialogVisible = false;
  }

  addDialogCheck(): void {
    const col = this.letterToCol(this.dialogNewCheck.col);
    if (col < 0) {
      this.dialogNewCheckError =
        'Column must be a valid letter (A, B, … AA, …)';
      return;
    }
    const row = this.dialogNewCheck.row - 1;
    if (row < 0) {
      this.dialogNewCheckError = 'Row must be at least 1';
      return;
    }
    if (!this.dialogNewCheck.value.trim()) {
      this.dialogNewCheckError = 'Value is required';
      return;
    }
    if (this.dialogChecks.some(c => c.col === col && c.row === row)) {
      this.dialogNewCheckError =
        'A check for this cell already exists';
      return;
    }
    this.dialogChecks = [
      ...this.dialogChecks,
      { col, row, value: this.dialogNewCheck.value.trim() },
    ];
    this.dialogNewCheck      = { col: 'A', row: 1, value: '' };
    this.dialogNewCheckError = null;
  }

  confirmRemoveDialogCheck(event: Event, index: number): void {
    this.confirmationService.confirm({
      target:      event.target as EventTarget,
      message:     'Remove this check?',
      icon:        'pi pi-exclamation-triangle',
      acceptLabel: 'Remove',
      rejectLabel: 'Cancel',
      accept: () => this.removeDialogCheck(index),
    });
  }

  removeDialogCheck(index: number): void {
    this.dialogChecks = this.dialogChecks.filter(
      (_, i) => i !== index);
  }

  saveDialog(): void {
    if (!this.selectedProject) return;

    const sc  = this.letterToCol(this.dialogDraft.sysname_column);
    const zc  = this.letterToCol(this.dialogDraft.zsample_column);
    const syc = this.letterToCol(this.dialogDraft.syscount_column);
    const mdc = this.letterToCol(
      this.dialogDraft.maxdistance_column);
    if (sc < 0 || zc < 0 || syc < 0 || mdc < 0) {
      this.saveDialogError =
        'Column must be a valid letter (A, B, … AA, …)';
      return;
    }
    const hr = this.dialogDraft.header_row - 1;
    if (hr < 0) {
      this.saveDialogError = 'Header row must be at least 1';
      return;
    }
    const name = this.dialogDraft.name.trim();
    if (!name) {
      this.saveDialogError = 'Name is required';
      return;
    }

    this.saveDialogLoading = true;
    this.saveDialogError   = null;

    const projectId = this.selectedProject.id;
    const body = {
      name,
      header_row:         hr,
      sysname_column:     sc,
      zsample_column:     zc,
      syscount_column:    syc,
      maxdistance_column: mdc,
    };

    const save$ = this.dialogMode === 'new'
      ? this.vsdsService.addVariant(projectId, body)
      : this.vsdsService.updateVariant(
          projectId, this.dialogVariantId!, body);

    save$.subscribe({
      next: (resp) => {
        this.doCheckOps(projectId, resp.data!.id);
      },
      error: (err: HttpErrorResponse) => {
        this.saveDialogError =
          err.error?.message ?? 'Failed to save variant';
        this.saveDialogLoading = false;
      },
    });
  }

  private doCheckOps(
    projectId: number,
    variantId: number,
  ): void {
    const toAdd = this.dialogChecks
      .filter(c => c.id === undefined)
      .map(c => this.vsdsService.addVariantCheck(
        projectId, variantId,
        { col: c.col, row: c.row, value: c.value },
      ));
    const toDelete = [...this.dialogOriginalCheckIds]
      .filter(id => !this.dialogChecks.some(c => c.id === id))
      .map(id => this.vsdsService.deleteVariantCheck(
        projectId, variantId, id));

    const all = [...toAdd, ...toDelete];
    if (all.length === 0) {
      this.finishSave(projectId, variantId);
      return;
    }

    forkJoin(all).subscribe({
      next: () => this.finishSave(projectId, variantId),
      error: (err: HttpErrorResponse) => {
        this.saveDialogError =
          err.error?.message ?? 'Failed to save checks';
        this.saveDialogLoading = false;
      },
    });
  }

  private finishSave(
    projectId: number,
    variantId: number,
  ): void {
    this.saveDialogLoading = false;
    this.dialogVisible     = false;
    this.loadVariants(projectId, variantId);
  }

  // ── Validation ────────────────────────────────────────────────

  validateSheet(): void {
    if (!this.validateUrl.trim() || !this.selectedProject) return;

    const sc  = this.letterToCol(this.dialogDraft.sysname_column);
    const zc  = this.letterToCol(this.dialogDraft.zsample_column);
    const syc = this.letterToCol(
      this.dialogDraft.syscount_column);
    const mdc = this.letterToCol(
      this.dialogDraft.maxdistance_column);
    if (sc < 0 || zc < 0 || syc < 0 || mdc < 0) {
      this.validateError = 'Fix column values before validating';
      return;
    }

    this.validateLoading   = true;
    this.validateResult    = null;
    this.validateError     = null;
    this.validateActiveTab = 0;

    this.vsdsService.validateVariant(this.selectedProject.id, {
      url: this.validateUrl.trim(),
      variant: {
        name:               this.dialogDraft.name.trim() || 'draft',
        header_row:         this.dialogDraft.header_row - 1,
        sysname_column:     sc,
        zsample_column:     zc,
        syscount_column:    syc,
        maxdistance_column: mdc,
        checks: this.dialogChecks.map(c => ({
          col: c.col, row: c.row, value: c.value,
        })),
      },
    }).subscribe({
      next: (resp) => {
        this.validateResult  = resp.data!;
        this.validateLoading = false;
      },
      error: (err: HttpErrorResponse) => {
        this.validateError  =
          err.error?.message ?? 'Validation failed';
        this.validateLoading = false;
      },
    });
  }

  // ── Spreadsheet cell helpers (used from template) ─────────────

  tabColIndices(tab: ValidationTabResult): number[] {
    const n = tab.rows.reduce((m, r) => Math.max(m, r.length), 0);
    return Array.from({ length: n }, (_, i) => i);
  }

  cellCheckResult(
    tab: ValidationTabResult,
    row: number,
    col: number,
  ): ValidationCheckResult | undefined {
    return tab.checks.find(c => c.row === row && c.col === col);
  }

  cellClass(
    tab: ValidationTabResult,
    row: number,
    col: number,
  ): string {
    const r = this.cellCheckResult(tab, row, col);
    if (!r) return '';
    return r.ok ? 'cell-ok' : 'cell-error';
  }

  cellTooltip(
    tab: ValidationTabResult,
    row: number,
    col: number,
  ): string {
    const r = this.cellCheckResult(tab, row, col);
    if (!r) return '';
    const coord = this.colToLetter(r.col) + (r.row + 1);
    const base  = `${coord} = '${r.expected}'`;
    return r.ok ? base : `${base} — found: '${r.actual}'`;
  }

  // ── Private helpers ───────────────────────────────────────────

  private updateProject(project: VSDSProject): void {
    const idx = this.projects.findIndex(p => p.id === project.id);
    if (idx !== -1) {
      const updated = [...this.projects];
      updated[idx]  = project;
      this.projects = updated;
    }
    this.selectedProject = project;
  }

  private updateVariantInList(variant: VSDSSheetVariant): void {
    const idx = this.variants.findIndex(v => v.id === variant.id);
    if (idx !== -1) {
      const updated = [...this.variants];
      updated[idx]  = variant;
      this.variants = updated;
    }
  }
}
