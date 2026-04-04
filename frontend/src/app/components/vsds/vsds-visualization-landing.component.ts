import { Component }    from '@angular/core';
import { Router }       from '@angular/router';
import { CardModule }   from 'primeng/card';
import { ButtonModule } from 'primeng/button';

interface VisCard {
  label:       string;
  description: string;
  icon:        string;
  route:       string;
}

@Component({
  selector:    'app-vsds-visualization-landing',
  standalone:  true,
  imports:     [CardModule, ButtonModule],
  templateUrl: './vsds-visualization-landing.component.html',
  styleUrl:    './vsds-visualization-landing.component.scss',
})
export class VsdsVisualizationLandingComponent {
  readonly cards: VisCard[] = [
    {
      label:       'Bowling Pins',
      description: '3D view of survey columns placed on the galaxy map. '
                 + 'Pin radius at each altitude encodes local density.',
      icon:        'pi pi-chart-bar',
      route:       '/vsds/visualization/bowling-pins',
    },
    {
      label:       'Cross-Section',
      description: 'Draw a line on the galaxy map and view a density '
                 + 'cross-section of nearby survey points.',
      icon:        'pi pi-map',
      route:       '/vsds/visualization/crosssection',
    },
  ];

  constructor(private router: Router) {}

  open(route: string): void {
    this.router.navigate([route]);
  }
}
