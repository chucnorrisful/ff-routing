import { Component, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { Debugpanel } from './components/unauthenticated/debugpanel/debugpanel';
import { Footer } from './components/core/footer/footer';
import { Nav } from './components/core/nav/nav';


@Component({
  selector: 'app-root',
  imports: [RouterOutlet, Debugpanel, Footer, Nav],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('ff-frontend');
}
