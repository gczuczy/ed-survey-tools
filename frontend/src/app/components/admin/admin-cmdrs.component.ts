import { Component, OnInit }             from '@angular/core';
import { FormsModule }                   from '@angular/forms';
import { HttpErrorResponse }             from '@angular/common/http';
import { AdminService, Cmdr }            from '../../services/admin.service';
import { ButtonModule }                  from 'primeng/button';
import { CardModule }                    from 'primeng/card';
import { CheckboxModule }                from 'primeng/checkbox';
import { ConfirmDialogModule }           from 'primeng/confirmdialog';
import { InputTextModule }               from 'primeng/inputtext';
import { MessageModule }                 from 'primeng/message';
import { TableModule }                   from 'primeng/table';
import { TooltipModule }                 from 'primeng/tooltip';
import { ConfirmationService }           from 'primeng/api';

@Component({
  selector:    'app-admin-cmdrs',
  standalone:  true,
  imports: [
    FormsModule,
    ButtonModule,
    CardModule,
    CheckboxModule,
    ConfirmDialogModule,
    InputTextModule,
    MessageModule,
    TableModule,
    TooltipModule,
  ],
  providers:   [ConfirmationService],
  templateUrl: './admin-cmdrs.component.html',
  styleUrl:    './admin-cmdrs.component.scss',
})
export class AdminCmdrsComponent implements OnInit {
  allCmdrs:    Cmdr[] = [];
  loading      = true;
  loadError:   string | null = null;

  nameFilter          = '';
  onlyWithCustomerId  = false;

  updatingId:   number | null = null;
  updateError:  string | null = null;

  constructor(
    private adminService:        AdminService,
    private confirmationService: ConfirmationService,
  ) {}

  ngOnInit(): void {
    this.adminService.listCmdrs().subscribe({
      next: (cmdrs) => {
        this.allCmdrs = cmdrs;
        this.loading  = false;
      },
      error: (err: HttpErrorResponse) => {
        this.loadError = err.error?.message ?? 'Failed to load commanders';
        this.loading   = false;
      },
    });
  }

  get filteredCmdrs(): Cmdr[] {
    const name = this.nameFilter.trim().toLowerCase();
    return this.allCmdrs.filter(c => {
      if (name && !c.name.toLowerCase().includes(name)) return false;
      if (this.onlyWithCustomerId && !c.customerid) return false;
      return true;
    });
  }

  toggleAdmin(cmdr: Cmdr): void {
    const action  = cmdr.isadmin ? 'revoke admin from' : 'grant admin to';
    const newVal  = !cmdr.isadmin;

    this.confirmationService.confirm({
      message: `Are you sure you want to ${action} CMDR "${cmdr.name}"?`,
      header:  'Confirm Permission Change',
      icon:    'pi pi-exclamation-triangle',
      accept:  () => {
        this.updatingId  = cmdr.id;
        this.updateError = null;

        this.adminService.setCmdrAdmin(cmdr.id, newVal).subscribe({
          next: (updated) => {
            const idx = this.allCmdrs.findIndex(c => c.id === updated.id);
            if (idx !== -1) {
              this.allCmdrs = [
                ...this.allCmdrs.slice(0, idx),
                updated,
                ...this.allCmdrs.slice(idx + 1),
              ];
            }
            this.updatingId = null;
          },
          error: (err: HttpErrorResponse) => {
            this.updateError = err.error?.message
              ?? `Failed to update permissions for "${cmdr.name}"`;
            this.updatingId = null;
          },
        });
      },
    });
  }
}
