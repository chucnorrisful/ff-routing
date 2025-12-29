import { inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { tap, catchError, of } from 'rxjs';
import { ApiService } from './api.service';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly useMock = false;

  private readonly _loggedIn = signal(this.readFromStorage());

  private readonly api = inject(ApiService)

  constructor(
    private router: Router,
  ) {}

  get isLoggedIn(): boolean {
    return this._loggedIn();
  }

  login(user: string, pass: string): void {

    // ---------- MOCK LOGIC ----------
    if (this.useMock) {
      console.log('🔹 Mock login →', user);
      this._loggedIn.set(true);
      this.saveToStorage(true);
      return;
    }
    
    this.api.login(user, pass).subscribe({
      next: res => {
        alert('Login successful!');
        console.log(res)
      },
      error: err => {
        console.error('Register error', err);
        alert('Login failed');
      }
    })
  }
  
  register(pass: string): void {
    this.api.register(pass).subscribe({
      next: res => {
        alert('Registration successful!');
        console.log(res)
      },
      error: err => {
        console.error('Register error', err);
        alert('Registration failed');
      }
    })
  }

  logout(): void {
    this._loggedIn.set(false);
    this.saveToStorage(false);
    this.router.navigate(['/login']);
  }

  private readFromStorage(): boolean {
    const stored = localStorage.getItem('auth');
    return stored === '1';
  }

  private saveToStorage(value: boolean): void {
    if (value) {
      localStorage.setItem('auth', '1');
    } else {
      localStorage.removeItem('auth');
    }
  }
}