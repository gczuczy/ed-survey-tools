import { Component, OnInit, OnDestroy } from '@angular/core';
import { FormsModule }                from '@angular/forms';
import { HttpErrorResponse }          from '@angular/common/http';
import { ActivatedRoute, Router }     from '@angular/router';
import { Subscription }               from 'rxjs';
import { AuthService }                from '../../auth/auth.service';
import { BreadcrumbService }          from '../../services/breadcrumb.service';
import { VsdsService, VSDSProject,
         VSDSSheetVariant,
         VSDSSheetVariantCheck }      from '../../services/vsds.service';
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

  // ── Sheet variants ────────────────────────────────────────────────
  variants:         VSDSSheetVariant[] = [];
  loadingVariants   = false;
  variantLoadError: string | null = null;

  // which tab is active ('new' | variant.id as number)
  activeVariantTab: string | number = 'new';

  // id of the variant currently being edited, or null
  editingVariantId: number | null = null;

  // mutable draft used while editing an existing variant
  editDraft: {
    name:               string;
    header_row:         number;   // 1-indexed display value
    sysname_column:     string;
    zsample_column:     string;
    syscount_column:    string;
    maxdistance_column: string;
  } = {
    name: '', header_row: 1,
    sysname_column: 'A', zsample_column: 'B',
    syscount_column: 'C', maxdistance_column: 'D',
  };

  // add-new-variant form
  newVariant = {
    name:               '',
    header_row:         1,         // 1-indexed; subtract 1 on submit
    sysname_column:     'A',       // letter; convert on submit
    zsample_column:     'B',
    syscount_column:    'C',
    maxdistance_column: 'D',
  };

  // per-active-variant add-check form
  newCheck = {
    col:   'A',
    row:   1,          // 1-indexed; subtract 1 on submit
    value: '',
  };

  addVariantLoad    = false;
  addVariantError:  string | null = null;

  savingVariantId:  number | null = null;
  saveVariantError: string | null = null;

  deletingVariantId:  number | null = null;
  deleteVariantError: string | null = null;

  addCheckLoad    = false;
  addCheckError:  string | null = null;

  deletingCheckId:  number | null = null;
  deleteCheckError: string | null = null;

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
    this.selectedProject    = project;
    this.addZSampleError    = null;
    this.delZSampleError    = null;
    this.bulkError          = null;
    this.newZSample         = null;
    this.bulkZSamples       = '';
    this.editingVariantId   = null;
    this.addVariantError    = null;
    this.saveVariantError   = null;
    this.deleteVariantError = null;
    this.addCheckError      = null;
    this.deleteCheckError   = null;
    this.newCheck = { col: 'A', row: 1, value: '' };
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

  loadVariants(projectId: number): void {
    this.loadingVariants  = true;
    this.variantLoadError = null;
    this.variants = [];
    this.vsdsService.listVariants(projectId).subscribe({
      next: (resp) => {
        this.variants        = resp.data ?? [];
        this.loadingVariants = false;
        this.activeVariantTab = this.variants.length > 0
          ? this.variants[0].id
          : 'new';
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

  startEdit(variant: VSDSSheetVariant): void {
    this.editingVariantId = variant.id;
    this.saveVariantError = null;
    this.editDraft = {
      name:               variant.name,
      header_row:         variant.header_row + 1,
      sysname_column:     this.colToLetter(variant.sysname_column),
      zsample_column:     this.colToLetter(variant.zsample_column),
      syscount_column:    this.colToLetter(variant.syscount_column),
      maxdistance_column: this.colToLetter(
        variant.maxdistance_column),
    };
  }

  cancelEdit(): void {
    this.editingVariantId = null;
    this.saveVariantError = null;
  }

  saveVariant(variantId: number): void {
    if (!this.selectedProject) return;
    const sc  = this.letterToCol(this.editDraft.sysname_column);
    const zc  = this.letterToCol(this.editDraft.zsample_column);
    const syc = this.letterToCol(this.editDraft.syscount_column);
    const mdc = this.letterToCol(
      this.editDraft.maxdistance_column);
    if (sc < 0 || zc < 0 || syc < 0 || mdc < 0) {
      this.saveVariantError =
        'Column must be a valid letter (A, B, … AA, …)';
      return;
    }
    const hr = this.editDraft.header_row - 1;
    if (hr < 0) {
      this.saveVariantError = 'Header row must be at least 1';
      return;
    }
    this.savingVariantId  = variantId;
    this.saveVariantError = null;
    this.vsdsService.updateVariant(
      this.selectedProject.id, variantId,
      {
        name:               this.editDraft.name.trim(),
        header_row:         hr,
        sysname_column:     sc,
        zsample_column:     zc,
        syscount_column:    syc,
        maxdistance_column: mdc,
      },
    ).subscribe({
      next: (resp) => {
        this.updateVariantInList(resp.data!);
        this.editingVariantId = null;
        this.savingVariantId  = null;
      },
      error: (err: HttpErrorResponse) => {
        this.saveVariantError =
          err.error?.message ?? 'Failed to save variant';
        this.savingVariantId  = null;
      },
    });
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

  addVariant(): void {
    if (!this.selectedProject) return;
    if (!this.newVariant.name.trim()) {
      this.addVariantError = 'Name is required';
      return;
    }
    const sc  = this.letterToCol(this.newVariant.sysname_column);
    const zc  = this.letterToCol(this.newVariant.zsample_column);
    const syc = this.letterToCol(
      this.newVariant.syscount_column);
    const mdc = this.letterToCol(
      this.newVariant.maxdistance_column);
    if (sc < 0 || zc < 0 || syc < 0 || mdc < 0) {
      this.addVariantError =
        'Column must be a valid letter (A, B, … AA, …)';
      return;
    }
    const hr = this.newVariant.header_row - 1;
    if (hr < 0) {
      this.addVariantError = 'Header row must be at least 1';
      return;
    }
    this.addVariantLoad  = true;
    this.addVariantError = null;
    this.vsdsService.addVariant(this.selectedProject.id, {
      name:               this.newVariant.name.trim(),
      header_row:         hr,
      sysname_column:     sc,
      zsample_column:     zc,
      syscount_column:    syc,
      maxdistance_column: mdc,
    }).subscribe({
      next: (resp) => {
        this.variants        = [...this.variants, resp.data!];
        this.addVariantLoad  = false;
        this.activeVariantTab = resp.data!.id;
        this.newVariant = {
          name: '', header_row: 1,
          sysname_column: 'A', zsample_column: 'B',
          syscount_column: 'C', maxdistance_column: 'D',
        };
      },
      error: (err: HttpErrorResponse) => {
        this.addVariantError =
          err.error?.message ?? 'Failed to add variant';
        this.addVariantLoad  = false;
      },
    });
  }

  addCheck(variantId: number): void {
    if (!this.selectedProject) return;
    const col = this.letterToCol(this.newCheck.col);
    if (col < 0) {
      this.addCheckError = 'Column must be a valid letter';
      return;
    }
    const row = this.newCheck.row - 1;
    if (row < 0) {
      this.addCheckError = 'Row must be at least 1';
      return;
    }
    if (!this.newCheck.value.trim()) {
      this.addCheckError = 'Value is required';
      return;
    }
    this.addCheckLoad  = true;
    this.addCheckError = null;
    this.vsdsService.addVariantCheck(
      this.selectedProject.id, variantId,
      { col, row, value: this.newCheck.value.trim() },
    ).subscribe({
      next: (resp) => {
        this.updateVariantInList(resp.data!);
        this.addCheckLoad = false;
        this.newCheck = { col: 'A', row: 1, value: '' };
      },
      error: (err: HttpErrorResponse) => {
        this.addCheckError =
          err.error?.message ?? 'Failed to add check';
        this.addCheckLoad  = false;
      },
    });
  }

  confirmDeleteCheck(
    event: Event,
    variantId: number,
    checkId: number,
  ): void {
    this.confirmationService.confirm({
      target:       event.target as EventTarget,
      message:      'Remove this header check?',
      icon:         'pi pi-exclamation-triangle',
      acceptLabel:  'Remove',
      rejectLabel:  'Cancel',
      accept: () => this.deleteCheck(variantId, checkId),
    });
  }

  deleteCheck(variantId: number, checkId: number): void {
    if (!this.selectedProject) return;
    this.deletingCheckId  = checkId;
    this.deleteCheckError = null;
    this.vsdsService.deleteVariantCheck(
      this.selectedProject.id, variantId, checkId,
    ).subscribe({
      next: (resp) => {
        this.updateVariantInList(resp.data!);
        this.deletingCheckId = null;
      },
      error: (err: HttpErrorResponse) => {
        this.deleteCheckError =
          err.error?.message ?? 'Failed to delete check';
        this.deletingCheckId = null;
      },
    });
  }

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
