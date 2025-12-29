import { Component, inject } from '@angular/core';
import { AuthService } from '../../service/auth.service';

@Component({
  selector: 'app-main',
  imports: [],
  templateUrl: './main.html',
  styleUrl: './main.css',
})
export class Main {

  private readonly auth = inject(AuthService);

  logOut(): void {
    this.auth.logout()
  }
  
}
