import { Component }     from '@angular/core';
import { Router }        from '@angular/router';
import { CarouselModule } from 'primeng/carousel';
import { CardModule }    from 'primeng/card';

interface SubsectionCard {
  label:       string;
  description: string;
  icon:        string;
  route:       string;
}

@Component({
  selector:    'app-admin-dashboard',
  standalone:  true,
  imports:     [CarouselModule, CardModule],
  templateUrl: './admin-dashboard.component.html',
  styleUrl:    './admin-dashboard.component.scss',
})
export class AdminDashboardComponent {
  subsections: SubsectionCard[] = [
    {
      label:       'Commanders',
      description: 'Manage commander accounts and admin permissions.',
      icon:        'pi pi-users',
      route:       '/admin/cmdrs',
    },
  ];

  constructor(private router: Router) {}

  navigate(route: string): void {
    this.router.navigate([route]);
  }
}
