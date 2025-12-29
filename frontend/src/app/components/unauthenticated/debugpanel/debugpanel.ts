import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../../service/auth.service';


@Component({
  selector: 'app-debugpanel',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './debugpanel.html',
  styleUrl: './debugpanel.css',
})
export class Debugpanel {
  // inject the singleton auth service
  private readonly auth: AuthService = inject(AuthService);

  /** expose the logged‑in flag for the template */
  get isLoggedIn(): boolean {
    return this.auth.isLoggedIn;
  }
}