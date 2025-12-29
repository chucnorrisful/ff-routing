import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Debugpanel } from './debugpanel';

describe('Debugpanel', () => {
  let component: Debugpanel;
  let fixture: ComponentFixture<Debugpanel>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Debugpanel]
    })
    .compileComponents();

    fixture = TestBed.createComponent(Debugpanel);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
