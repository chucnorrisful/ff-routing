import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly baseUrl = 'http://localhost:8080';

  constructor(private http: HttpClient) {}

  private url(path: string): string {
    return `${this.baseUrl}${path.startsWith('/') ? '' : '/'}${path}`;
  }

  login(username: string, password: string) {
    const body = { username, password };
    return this.http.post(this.url('/login'), body);
  }

  register(password: string) {
    const body = { pw: password };
    return this.http.post(this.url('/register'), body);
  }
}