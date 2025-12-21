import { Component, ComponentFactoryResolver, inject, OnDestroy, OnInit, input, effect } from '@angular/core';
import { max, Subscription } from 'rxjs';
import { HistogramsService } from 'src/app/histograms.service';
import { DateTime } from "luxon";

import { NgxEchartsDirective, provideEchartsCore } from 'ngx-echarts';
import * as echarts from 'echarts/core';
import { EChartsOption } from 'echarts/types/dist/shared';
import { HeatmapChart } from 'echarts/charts';
import { GridComponent, TitleComponent, VisualMapComponent, TooltipComponent } from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';
echarts.use([HeatmapChart, GridComponent, TitleComponent, VisualMapComponent, TooltipComponent, CanvasRenderer]);

@Component({
  selector: 'app-heatmap',
  imports: [
    NgxEchartsDirective
  ],
  templateUrl: './heatmap.component.html',
  styleUrl: './heatmap.component.css',
  providers: [
    provideEchartsCore({ echarts }),
  ]
})
export class HeatmapComponent implements OnInit, OnDestroy {

  public visible = input<boolean>(false);
  public targetUuid = input.required<string>();

  public data: any[] = [];

  public xData: string[] = [];
  public yData: number[] = [];
  public valueData: number[] = [];
  public countData: any[][] = [];
  public largestCount: number = 0;

  public chartOption: EChartsOption = {};
  public echartsInstance: any;

  private subscriptions: Subscription = new Subscription();
  private readonly HistogramsService: HistogramsService = inject(HistogramsService);

  constructor() {
    effect(() => {
      if (this.visible()) {
        this.loadData();
      }
    });
  }

  public ngOnInit(): void {

  }

  public ngOnDestroy(): void {
    this.subscriptions.unsubscribe();
  }

  onChartInit(ec: any) {
    this.echartsInstance = ec;
  }

  private loadData(): void {
    this.subscriptions.add(this.HistogramsService.getHistogramByUuid(this.targetUuid()).subscribe((histogram) => {
      // Convert into heatmap data
      this.countData = [];

      // The heatmap expects the data like [ [xPosIndex, yPosIndex, count], [xPosIndex, yPosIndex, count] ]
      const uniqueDates = [...new Set(histogram.map((value) => value[0]))];
      let uniqueBuckets = [...new Set(histogram.map((value) => value[1]))];

      let dataAndBuchets: any = {};
      histogram.forEach((element) => {
        const time = element[0];
        const msbucket = element[1];
        const count = element[2];

        if (count > this.largestCount)
          this.largestCount = count;

        if (!dataAndBuchets[time]) {
          dataAndBuchets[time] = [];
        }

        dataAndBuchets[time][msbucket] = count;
      });

      let timestampIndex = 0;
      for (let timestamp in dataAndBuchets) {
        uniqueBuckets.forEach((bucket, bucketIndex) => {
          if (dataAndBuchets[timestamp].hasOwnProperty(bucket)) {
            this.countData.push([timestampIndex, bucketIndex, dataAndBuchets[timestamp][bucket]]);
          } else {
            // We push a '-' instead of a 0, because the heatmap will not render the 0 values
            // With a '-' the heatmap will render the empty spaces and the chart is more readable
            this.countData.push([timestampIndex, bucketIndex, '-']);
          }
        });
        timestampIndex++;
      }


      this.xData = uniqueDates;
      this.yData = uniqueBuckets;

      console.log(this.xData);
      console.log(this.yData);
      console.log(this.countData);


      this.renderAsHeatmapChart();
    }
    ));
  }


  private renderAsHeatmapChart(): void {

    console.log(Math.max(...this.yData))
    let labels: string[] = this.yData.map((value) => {
      return value.toFixed(2);
    });


    this.chartOption = {
      tooltip: {
        formatter: (params: any) => {


          const gauge = params.value;
          const dateTime = DateTime.fromISO(this.xData[gauge[0]]);
          const count = gauge[2];

          const msIndex = gauge[1];

          const ms = this.yData[msIndex];


          const seriesName = gauge.seriesName;
          const color = gauge.color;
          const marker = params.marker;

          const html = `<div class="row">
              <div class="col-12">
                  ${dateTime.toFormat('dd.MM.yyyy HH:mm:ss')}
              </div>
              <div class="col-12">
                  ${marker} <span class="float-end bold" style="color:${color};">${ms}ms [${count}]</span>
              </div>
              </div>`;

          return html;
        }
      },
      xAxis: {
        type: 'category',
        data: this.xData,
        axisLabel: {
          interval: 0,
          rotate: 90,
          formatter: function (value) {
            const dateTime = DateTime.fromISO(value);
            return dateTime.toFormat('HH:mm');
          }
        },
        splitArea: {
          show: true
        },
        splitLine: {
          show: true,
        },
        axisTick: {
          show: true,
        },
        minorSplitLine: {
          show: true
        }
      },
      yAxis: {
        type: 'category',
        data: labels,
        min: 0,
        max: Math.max(...this.yData),
        axisLabel: {
          formatter: function (value) {
            return value + 'ms';
          }
        },
        splitArea: {
          show: true
        },
        minorTick: {
          show: true
        },
        minorSplitLine: {
          show: true
        }
      },
      visualMap: {
        min: 0,
        max: this.largestCount,
        calculable: true,
        realtime: true,
        inRange: {
          color: [
            '#313695',
            '#4575b4',
            '#74add1',
            '#abd9e9',
            '#e0f3f8',
            '#ffffbf',
            '#fee090',
            '#fdae61',
            '#f46d43',
            '#d73027',
            '#a50026'
          ]
        }
      },
      series: [
        {
          name: 'Gaussian',
          type: 'heatmap',
          data: this.countData,
          label: {
            show: false
          },
          emphasis: {
            itemStyle: {
              borderColor: '#333',
              borderWidth: 1
            }
          },
          progressive: 1000,
          animation: false
        }
      ]
    };
  }


}
