import { Component }    from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector:    'app-vsds-visualization',
  standalone:  true,
  imports:     [RouterOutlet],
  templateUrl: './vsds-visualization.component.html',
  styleUrl:    './vsds-visualization.component.scss',
})
export class VsdsVisualizationComponent {}
