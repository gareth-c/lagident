import { Component, inject, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { DecimalPipe, NgIf, PercentPipe } from '@angular/common';
import { Subscription } from 'rxjs';
import { Target, TargetWithStatistics } from '../targets.interfaces';
import { TargetsService } from '../targets.service';
import { TargetsAddComponent } from "../targets-add/targets-add.component";
import { ButtonModule } from 'primeng/button';
import { TableModule } from 'primeng/table';
import { TooltipModule } from 'primeng/tooltip';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ConfirmationService } from 'primeng/api';
import { ToggleButtonModule } from 'primeng/togglebutton';
import { FormsModule } from '@angular/forms';
import { CardModule } from 'primeng/card';
import { DialogModule } from 'primeng/dialog';
import { HeatmapComponent } from "../../components/heatmap/heatmap.component";
import { SelectButtonModule } from 'primeng/selectbutton';
import { ScatterChartComponent } from "../../components/scatter-chart/scatter-chart.component";
@Component({
    selector: 'app-targets-index',
    imports: [
        TargetsAddComponent,
        DecimalPipe,
        PercentPipe,
        ButtonModule,
        TableModule,
        TooltipModule,
        NgIf,
        ConfirmDialogModule,
        ToggleButtonModule,
        FormsModule,
        CardModule,
        DialogModule,
        //HeatmapComponent,
        SelectButtonModule,
        ScatterChartComponent
    ],
    providers: [
        ConfirmationService
    ],
    templateUrl: './targets-index.component.html',
    styleUrl: './targets-index.component.css'
})
export class TargetsIndexComponent implements OnInit, OnDestroy {


  public targets: TargetWithStatistics[] = [];
  public autorefresh: boolean = true;
  public dialogVisible: boolean = false;

  private intervalId: any;

  private subscriptions: Subscription = new Subscription();
  private readonly TargetsService: TargetsService = inject(TargetsService);
  private readonly ConfirmationService: ConfirmationService = inject(ConfirmationService);

public data      = [
  { column1: 'Data 1', column2: 'Data 2', column3: 'Data 3',column4: 'Data 3',column5: 'Data 3',column6: 'Data 3' },
  { column1: 'Data 1', column2: 'Data 2', column3: 'Data 3',column4: 'Data 3',column5: 'Data 3',column6: 'Data 3' },
  // More data...
];

  // Dialog variables
  public selectedChart = 'scatter';
  public chartTypes: any[] = [
    {
      label: 'Scatter',
      value: 'scatter',
      icon: 'pi pi-chart-scatter'
    },
    //{
    //  label: 'Heatmap (Beta)',
    //  value: 'heatmap',
    //  icon: 'pi pi-chart-bar'
    //}
  ];
  public selectedTarget?: Target;

  public ngOnInit(): void {
    this.loadTargets();
    this.intervalId = setInterval(() => {
      this.loadTargets();
    }, 15000); // 15 seconds
  }

  public ngOnDestroy(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
    this.subscriptions.unsubscribe();
  }

  public toggleAutorefresh() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    } else {
      this.intervalId = setInterval(() => {
        this.loadTargets();
      }, 15000); // 15 seconds
    }

  }

  public loadTargets() {
    this.subscriptions.add(this.TargetsService.getTargetsWithStatistics().subscribe((data: TargetWithStatistics[]) => {
      this.targets = data;
    }));
  }


  public confirmDelete(target: Target) {
    this.ConfirmationService.confirm({
      message: `Are you sure you want to delete target: "${target.name}"?`,
      accept: () => {
        this.subscriptions.add(this.TargetsService.deleteTarget(target.uuid).subscribe(() => {
          this.loadTargets();
        }));
      }
    });
  }

  public showDialog(target: Target) {
    this.dialogVisible = true;
    this.selectedTarget = target;
  }

}
