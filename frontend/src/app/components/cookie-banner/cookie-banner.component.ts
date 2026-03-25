import { Component, OnInit } from '@angular/core';
import { ButtonModule }     from 'primeng/button';

const ACK_KEY = 'edst-cookie-ack';

@Component({
  selector:    'app-cookie-banner',
  standalone:  true,
  imports:     [ButtonModule],
  templateUrl: './cookie-banner.component.html',
  styleUrl:    './cookie-banner.component.scss'
})
export class CookieBannerComponent implements OnInit {
  visible = false;

  ngOnInit(): void {
    this.visible = !localStorage.getItem(ACK_KEY);
  }

  dismiss(): void {
    localStorage.setItem(ACK_KEY, '1');
    this.visible = false;
  }
}
