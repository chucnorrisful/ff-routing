import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../service/auth.service';

export const authGuard: CanActivateFn = (route, state) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isLoggedIn) {
    // user is authenticated → allow navigation
    return true;
  }

  // not logged in → redirect to login, preserve intended URL for later use
  router.navigate(['/login'], { queryParams: { returnUrl: state.url } });
  return false;  
  
  return true;
};
