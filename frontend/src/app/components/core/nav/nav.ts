import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { AuthService } from '../../../service/auth.service';

interface NavItem {
  label: string;
  route?: string;          // router link – omitted for logout button
  action?: () => void;     // optional click handler (used for logout)
}

@Component({
  selector: 'app-nav',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './nav.html',
  styleUrl: './nav.css'
})
export class Nav {
  /** Auth service – injected via the functional API */
  private readonly auth: AuthService = inject(AuthService);

  /** Title shown on the left side */
  readonly title = 'ff‑frontend';

  /** Items for unauthenticated users */
  private readonly unauthItems: NavItem[] = [
    { label: 'Login',    route: '/login' },
    { label: 'Register', route: '/register' }
  ];

  /** Items for authenticated users */
  private readonly authItems: NavItem[] = [
    { label: 'Home',    route: '/' },
    { label: 'Friends', route: '/friends' },   // placeholder – add a real route later
    {
      label: 'Logout',
      action: () => this.auth.logout()          // calls the service logout()
    }
  ];

  /** Returns the correct list based on auth state */
  get items(): NavItem[] {
    return this.auth.isLoggedIn ? this.authItems : this.unauthItems;
  }

  /** Helper for the template – expose the auth flag */
  get isAuthenticated(): boolean {
    return this.auth.isLoggedIn;
  }
}