import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TargetsIndexComponent } from './targets-index.component';

describe('TargetsIndexComponent', () => {
  let component: TargetsIndexComponent;
  let fixture: ComponentFixture<TargetsIndexComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TargetsIndexComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TargetsIndexComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
