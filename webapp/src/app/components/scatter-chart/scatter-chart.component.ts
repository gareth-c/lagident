import { Component, effect, inject, input, OnDestroy, OnInit } from '@angular/core';
import { max, Subscription } from 'rxjs';
import { HistogramsService } from 'src/app/histograms.service';
import { DateTime } from "luxon";

import { NgxEchartsDirective, provideEchartsCore } from 'ngx-echarts';
import * as echarts from 'echarts/core';
import { EChartsOption } from 'echarts/types/dist/shared';
import { BarChart, ScatterChart } from 'echarts/charts';
import { GridComponent, TitleComponent, VisualMapComponent, TooltipComponent } from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';
echarts.use([BarChart, ScatterChart, GridComponent, TitleComponent, VisualMapComponent, TooltipComponent, CanvasRenderer]);

@Component({
  selector: 'app-scatter-chart',
  imports: [
    NgxEchartsDirective
  ],
  templateUrl: './scatter-chart.component.html',
  styleUrl: './scatter-chart.component.css',
  providers: [
    provideEchartsCore({ echarts }),
  ]
})
export class ScatterChartComponent implements OnInit, OnDestroy {

  public visible = input<boolean>(false);
  public targetUuid = input.required<string>();

  public latencyData: any[] = [];
  public lossData: any[] = [];
  public allTimestampsArray: number[] = [];

  public chartOption: EChartsOption = {};
  public echartsInstance: any;

  public maxLatency: number = 0;

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
    this.latencyData = [];
    this.lossData = [];
    this.allTimestampsArray = [];


    this.subscriptions.add(this.HistogramsService.getTimeseriesByUuid(this.targetUuid()).subscribe((timeseries) => {

      // The server response with two separate arrays for latency and loss time series data.
      // How ever, looks like eCharts only supports one array of time stamps for two series.
      // So we have to merge the timestamps of both series into one array, and fill missing values with null.
      // I could not found much information about this than
      // https://stackoverflow.com/questions/60616117/two-array-series-with-different-time-stamps-and-a-different-number-of-data-point

      const allTimestamps: { [key: number]: number } = {};
      let allTimestampsArray: number[] = [];
      const allLatencyTimestamps: { [key: number]: number } = {};
      const allLossTimestamps: { [key: number]: number } = {};

      const latencyValues: any[] = [];
      const lossValues: any[] = [];

      // Save all timestamps in a dictionary
      timeseries.Latencies.forEach((latency) => {
        allTimestamps[latency.timestamp] = 1; // We are only interested in the keys
        allLatencyTimestamps[latency.timestamp] = latency.latency;

        if (latency.latency > this.maxLatency) {
          this.maxLatency = latency.latency;
        }
      });
      timeseries.Losses.forEach((loss) => {
        if (!allTimestamps[loss.timestamp]) {
          allTimestamps[loss.timestamp] = 1; // We are only interested in the keys
          allLossTimestamps[loss.timestamp] = 1; // a loss i binary (we either have a loss or not)
        }
      });

      // Store data to all timestamps
      // It is important to use an array to keep the order of the timestamps
      allTimestampsArray = Object.keys(allTimestamps).map(Number).sort((a, b) => a - b);

      allTimestampsArray.forEach((timestamp) => {
        // Do we have a latency value for this timestamp?
        if (allLatencyTimestamps[timestamp]) {
          latencyValues.push([String(DateTime.fromSeconds(timestamp).toISO()), allLatencyTimestamps[timestamp]]);
        } else {
          latencyValues.push([String(DateTime.fromSeconds(timestamp).toISO()), null]);
        }

        // Do we have a loss value for this timestamp?
        if (allLossTimestamps[timestamp]) {

          lossValues.push([String(DateTime.fromSeconds(timestamp).toISO()), 1]);
        } else {
          lossValues.push([String(DateTime.fromSeconds(timestamp).toISO()), null]);
        }
      });

      this.latencyData = latencyValues;
      this.lossData = lossValues;
      this.allTimestampsArray = allTimestampsArray;

      this.renderAsScatterChart();
    }
    ));
  }


  private renderAsScatterChart(): void {

    let startTime = undefined;
    let endTime = undefined;
    if (this.allTimestampsArray.length > 0) {
      startTime = String(DateTime.fromSeconds(this.allTimestampsArray[0]).toISO());
    }
    if (this.allTimestampsArray.length > 0) {
      endTime = String(DateTime.fromSeconds(this.allTimestampsArray[this.allTimestampsArray.length - 1]).toISO());
    }


    this.chartOption = {
      title: {
        text: 'Latency scatter',
        left: 'center',
        top: 0
      },
      visualMap: {
        seriesIndex: 0, // Only apply to the latency series
        min: 0,
        max: this.maxLatency,
        dimension: 1,
        orient: 'vertical',
        right: 10,
        top: 'center',
        text: ['HIGH', 'LOW'],
        calculable: true,
        inRange: {
          color: ['#00ff00', '#FF0000']
        },
      },
      tooltip: {
        trigger: 'item',
        axisPointer: {
          type: 'cross'
        },
        formatter: function (params: any) {
          //console.log(params);
          const date = DateTime.fromISO(params.value[0]);
          const formattedDate = date.toFormat('yyyy-MM-dd HH:mm:ss');
          if (params.seriesIndex === 0) {
            // Latency series
            return `
            ${params.marker}
            ${params.seriesName}<br/>
            Time: ${formattedDate}<br/>
            Latency: ${params.value[1]} ms
          `;
          }

          return `
          ${params.marker}
          ${params.seriesName}<br/>
          Time: ${formattedDate}<br>
          Loss detected
        `;
        }
      },
      xAxis: [
        {
          min: startTime,
          max: endTime,
          type: 'time',
          minorTick: {
            show: true
          },
          minorSplitLine: {
            show: true
          },
        },
      ],
      yAxis: [
        {
          type: 'value',
          name: 'Latency',
          minorTick: {
            show: true
          },
          minorSplitLine: {
            show: true
          }
        },
        {
          type: 'value',
          name: 'Loss',
          position: 'right',
          axisLine: {
            lineStyle: {
              color: '#FF0000'
            }
          },
          min: 0,
          max: 1,
          axisTick: {
            show: false
          },
          splitLine: {
            show: false
          },
          axisLabel: {
            show: true,
            formatter: function (value, index) {
              // Only show 0 and 1 as axis labels for Loss. A packet can either be lost or not lost.

              if (value === 0 || value === 1) {
                return value.toString();
              }

              return '';
            }
          }
        }
      ],
      series: [
        {
          name: 'Latency',
          type: 'scatter',
          symbolSize: 5,
          yAxisIndex: 0,
          data: this.latencyData
        },
        {
          name: 'Loss',
          type: 'bar',
          data: this.lossData,
          yAxisIndex: 1, // Use the second y-axis for the bar chart
          itemStyle: {
            color: '#D9342B',
            opacity: 0.3
          }
        }
      ]
    };
  }

}
