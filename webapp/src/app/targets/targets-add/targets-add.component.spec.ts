import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TargetsAddComponent } from './targets-add.component';

describe('TargetsAddComponent', () => {
  let component: TargetsAddComponent;
  let fixture: ComponentFixture<TargetsAddComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TargetsAddComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TargetsAddComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
