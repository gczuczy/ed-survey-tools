import { Component }           from '@angular/core';
import { AuthService }         from '../../auth/auth.service';
import { ThemeService }        from '../../auth/theme.service';
import { CardModule }          from 'primeng/card';
import { TagModule }           from 'primeng/tag';
import { DividerModule }       from 'primeng/divider';
import { ButtonModule }        from 'primeng/button';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ConfirmationService } from 'primeng/api';

@Component({
  selector:    'app-settings',
  standalone:  true,
  imports:     [
    CardModule, TagModule, DividerModule,
    ButtonModule, ConfirmDialogModule,
  ],
  providers:   [ConfirmationService],
  templateUrl: './settings.component.html',
  styleUrl:    './settings.component.scss'
})
export class SettingsComponent {
  deleteError: string | null = null;

  constructor(
    public  authService:         AuthService,
    public  themeService:        ThemeService,
    private confirmationService: ConfirmationService,
  ) {}

  get themeSeverity(): 'secondary' | 'info' {
    return this.themeService.currentTheme === 'dark'
      ? 'secondary' : 'info';
  }

  confirmDeleteAccount(): void {
    this.confirmationService.confirm({
      message: 'This will unlink your Frontier Developments account. ' +
               'Your CMDR name is retained for survey attribution but ' +
               'you will no longer be able to log in. Are you sure?',
      header:  'Delete Account',
      icon:    'pi pi-exclamation-triangle',
      acceptButtonStyleClass: 'p-button-danger',
      accept:  () => {
        this.deleteError = null;
        this.authService.deleteAccount().subscribe({
          next:  () => this.authService.clearSession(),
          error: () => {
            this.deleteError =
              'Failed to delete account. Please try again.';
          },
        });
      },
    });
  }
}
