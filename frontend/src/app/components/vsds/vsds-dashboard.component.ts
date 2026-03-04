import { Component, OnInit } from '@angular/core';
import { Router }            from '@angular/router';
import { AuthService }       from '../../auth/auth.service';
import { CarouselModule }    from 'primeng/carousel';
import { CardModule }        from 'primeng/card';

interface SubsectionCard {
  label:       string;
  description: string;
  icon:        string;
  route:       string;
}

@Component({
  selector:    'app-vsds-dashboard',
  standalone:  true,
  imports:     [CarouselModule, CardModule],
  templateUrl: './vsds-dashboard.component.html',
  styleUrl:    './vsds-dashboard.component.scss',
})
export class VsdsDashboardComponent implements OnInit {
  subsections: SubsectionCard[] = [];

  constructor(
    public  authService: AuthService,
    private router:      Router,
  ) {}

  ngOnInit(): void {
    const cards: SubsectionCard[] = [];

    if (this.authService.user?.isadmin) {
      cards.push({
        label:       'Folders',
        description: 'Manage Google Drive folders for document processing.',
        icon:        'pi pi-folder-open',
        route:       '/vsds/folders',
      });
    }

    this.subsections = cards;
  }

  navigate(route: string): void {
    this.router.navigate([route]);
  }
}
