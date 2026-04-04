import { Component, OnInit }    from '@angular/core';
import { Router }               from '@angular/router';
import { forkJoin }             from 'rxjs';
import { DecimalPipe, DatePipe } from '@angular/common';
import { AuthService }          from '../../auth/auth.service';
import { VsdsService,
         VSDSContribution,
         VSDSContribErrorDoc }  from '../../services/vsds.service';
import { CarouselModule }       from 'primeng/carousel';
import { CardModule }           from 'primeng/card';
import { MessageModule }        from 'primeng/message';
import { TableModule }          from 'primeng/table';
import { TagModule }            from 'primeng/tag';
import { ButtonModule }         from 'primeng/button';

interface SubsectionCard {
  label:       string;
  description: string;
  icon:        string;
  route:       string;
}

@Component({
  selector:    'app-vsds-dashboard',
  standalone:  true,
  imports: [
    CarouselModule,
    CardModule,
    MessageModule,
    TableModule,
    TagModule,
    ButtonModule,
    DecimalPipe,
    DatePipe,
  ],
  templateUrl: './vsds-dashboard.component.html',
  styleUrl:    './vsds-dashboard.component.scss',
})
export class VsdsDashboardComponent implements OnInit {
  subsections:    SubsectionCard[]       = [];
  contribution:   VSDSContribution | null = null;
  contribErrors:  VSDSContribErrorDoc[] | null = null;
  contribLoading  = false;

  constructor(
    public  authService: AuthService,
    private router:      Router,
    private vsdsService: VsdsService,
  ) {}

  ngOnInit(): void {
    const cards: SubsectionCard[] = [];

    cards.push({
      label:       'Projects',
      description: 'View and manage survey projects and their ' +
                   'ZSample assignments.',
      icon:        'pi pi-briefcase',
      route:       '/vsds/projects',
    });

    cards.push({
      label:       'Visualization',
      description: 'Interactive 3D and cross-section density plots.',
      icon:        'pi pi-chart-scatter',
      route:       '/vsds/visualization',
    });

    if (this.authService.user?.isadmin) {
      cards.push({
        label:       'Folders',
        description: 'Manage Google Drive folders for document ' +
                     'processing.',
        icon:        'pi pi-folder-open',
        route:       '/vsds/folders',
      });
    }

    this.subsections = cards;

    if (this.authService.user) {
      this.contribLoading = true;
      forkJoin({
        contrib: this.vsdsService.getMyContribution(),
        errors:  this.vsdsService.getMyContributionErrors(),
      }).subscribe({
        next: ({ contrib, errors }) => {
          this.contribution  = contrib.data  ?? null;
          this.contribErrors = errors.data   ?? null;
          this.contribLoading = false;
        },
        error: () => {
          this.contribLoading = false;
        },
      });
    }
  }

  navigate(route: string): void {
    this.router.navigate([route]);
  }
}
