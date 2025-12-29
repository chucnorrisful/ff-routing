import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../../service/auth.service';
import { Router } from '@angular/router';
import { ApiService } from '../../../service/api.service';

@Component({
  selector: 'app-login',
  standalone: true,               // <-- make it a standalone component
  imports: [CommonModule],        // <-- needed for built‑in directives
  templateUrl: './login.html',
  styleUrl: './login.css',
})
export class Login {


  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  uid = '';
  password = '';

  onSubmit(): void {
    this.auth.login(this.uid, this.password);
    this.router.navigate(['/']);
  } 
}