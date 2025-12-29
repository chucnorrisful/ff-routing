import { Routes } from '@angular/router';

import { Login } from './components/unauthenticated/login/login'
import { Register } from './components/unauthenticated/register/register'

import { authGuard } from './guard/auth.guard'
import { Main } from './components/main/main';
import { Friends } from './components/friends/friends';

export const routes: Routes = [
  // ... other routes can go here ...

  { path: 'login', component: Login },
  { path: 'register', component: Register },
  
  // authenticated
  { path: '', 
    component: Main,
    canActivate: [authGuard]
   },
  { path: 'friends', 
    component: Friends,
    canActivate: [authGuard] },
];