import { Component, Input, Output, EventEmitter } from '@angular/core';
import { DialogModule } from 'primeng/dialog';

@Component({
  selector:    'app-privacy-notice',
  standalone:  true,
  imports:     [DialogModule],
  templateUrl: './privacy-notice.component.html',
  styleUrl:    './privacy-notice.component.scss'
})
export class PrivacyNoticeComponent {
  @Input()  visible = false;
  @Output() visibleChange = new EventEmitter<boolean>();
}
