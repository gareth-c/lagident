import { Component, inject, OnDestroy, OnInit, ChangeDetectionStrategy, ViewChild, EventEmitter, Output } from '@angular/core';
import { Subscription } from 'rxjs';
import { TargetsService } from '../targets.service';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, NgForm, FormsModule, FormControl } from '@angular/forms';
import { v4 as uuidv4 } from 'uuid';
import { NgIf } from '@angular/common';
import { ToastrService } from 'ngx-toastr';
import { InputGroupModule } from 'primeng/inputgroup';
import { InputGroupAddonModule } from 'primeng/inputgroupaddon';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';

@Component({
    selector: 'app-targets-add',
    imports: [
        FormsModule,
        ReactiveFormsModule,
        NgIf,
        InputGroupModule,
        InputGroupAddonModule,
        CardModule,
        ButtonModule,
        InputTextModule
    ],
    templateUrl: './targets-add.component.html',
    styleUrl: './targets-add.component.css'
})
export class TargetsAddComponent implements OnInit, OnDestroy {

  @ViewChild('formRef') formRef!: NgForm;
  @Output() targetAdded = new EventEmitter<void>();


  public targetForm!: FormGroup;

  private subscriptions: Subscription = new Subscription();
  private readonly TargetsService: TargetsService = inject(TargetsService);
  private readonly ToastrService: ToastrService = inject(ToastrService);

  constructor(private fb: FormBuilder) {
    this.targetForm = this.fb.group({
      uuid: new FormControl(uuidv4(), [Validators.required]),
      name: new FormControl('', [Validators.required, Validators.maxLength(255)]),
      address: new FormControl('', [Validators.required, Validators.maxLength(255)]),
    });
  }

  public ngOnInit(): void {

  }

  public ngOnDestroy(): void {
    this.subscriptions.unsubscribe();
  }

  public onSubmit(): void {
    if (this.targetForm.valid) {
      const target = this.targetForm.value;
      this.subscriptions.add(
        this.TargetsService.addTarget(target).subscribe({
          next: (response) => {
            console.log('Target added successfully', response);
            this.ToastrService.success('Target added successfully');

            this.formRef.resetForm();
            this.targetForm.reset({
              uuid: uuidv4(),
            });

            this.targetAdded.emit();
          },
          error: (error) => {
            console.error('Error adding target', error);
            this.ToastrService.error('Error adding target');
          }
        })
      );
    }
  }

}
