import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ApiService } from '../../../service/api.service';
import { AuthService } from '../../../service/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './register.html',
  styleUrl: './register.css',
})
export class Register {
  pwd = '';
  pwdRepeat = '';
  private readonly auth = inject(AuthService);
  
  register(): void {
    if (this.pwd !== this.pwdRepeat) {
      alert('Passwords do not match');
      return;
    }
    this.auth.register(this.pwd)
  }
}