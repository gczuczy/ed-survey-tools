import { Component, OnInit }          from '@angular/core';
import { FormsModule }                from '@angular/forms';
import { HttpErrorResponse }          from '@angular/common/http';
import { AuthService }                from '../../auth/auth.service';
import { VsdsService, VSDSProject }   from '../../services/vsds.service';
import { ButtonModule }               from 'primeng/button';
import { CardModule }                 from 'primeng/card';
import { InputTextModule }            from 'primeng/inputtext';
import { InputNumberModule }          from 'primeng/inputnumber';
import { MessageModule }              from 'primeng/message';
import { TextareaModule }             from 'primeng/textarea';

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
  ],
  templateUrl: './vsds-projects.component.html',
  styleUrl:    './vsds-projects.component.scss',
})
export class VsdsProjectsComponent implements OnInit {
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

  constructor(
    public  authService: AuthService,
    private vsdsService: VsdsService,
  ) {}

  ngOnInit(): void {
    this.loadProjects();
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
      },
      error: (err: HttpErrorResponse) => {
        this.loadError = err.error?.message ?? 'Failed to load projects';
        this.loadingProjects = false;
      },
    });
  }

  selectProject(project: VSDSProject): void {
    this.selectedProject = project;
    this.addZSampleError = null;
    this.delZSampleError = null;
    this.bulkError       = null;
    this.newZSample      = null;
    this.bulkZSamples    = '';
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
        this.addProjectError = err.error?.message ?? 'Failed to add project';
        this.addProjectLoad  = false;
      },
    });
  }

  addZSample(): void {
    if (this.newZSample === null || !this.selectedProject) return;
    this.addZSampleLoad  = true;
    this.addZSampleError = null;
    this.vsdsService.addZSample(this.selectedProject.id, this.newZSample).subscribe({
      next: (resp) => {
        this.updateProject(resp.data!);
        this.newZSample     = null;
        this.addZSampleLoad = false;
      },
      error: (err: HttpErrorResponse) => {
        this.addZSampleError = err.error?.message ?? 'Failed to add ZSample';
        this.addZSampleLoad  = false;
      },
    });
  }

  deleteZSample(zsample: number): void {
    if (!this.selectedProject) return;
    this.deletingZSample = zsample;
    this.delZSampleError = null;
    this.vsdsService.deleteZSample(this.selectedProject.id, zsample).subscribe({
      next: (resp) => {
        this.updateProject(resp.data!);
        this.deletingZSample = null;
      },
      error: (err: HttpErrorResponse) => {
        this.delZSampleError = err.error?.message ?? 'Failed to delete ZSample';
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
    this.vsdsService.setZSamples(this.selectedProject.id, values).subscribe({
      next: (resp) => {
        this.updateProject(resp.data!);
        this.bulkZSamples = '';
        this.bulkLoad     = false;
      },
      error: (err: HttpErrorResponse) => {
        this.bulkError = err.error?.message ?? 'Failed to replace ZSamples';
        this.bulkLoad  = false;
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
}
