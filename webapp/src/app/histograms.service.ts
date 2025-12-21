import { inject, Injectable } from '@angular/core';
import { environment } from '../environments/environment';
import { map, Observable } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { TimeseriesResponse } from './histograms.interface';


@Injectable({
  providedIn: 'root'
})
export class HistogramsService {

  private readonly http = inject(HttpClient);

  constructor() { }

  public getHistogramByUuid(uuid:string): Observable<[time: string, msbucket: number, count: number][]> {
    return this.http.get<{buckets	: [time: string, msbucket: number, count: number][]}>(`${environment.apiUrl}/api/histograms/${uuid}`).pipe(
      map(data => {
        return data.buckets	;
      })
    )
  }

  public getTimeseriesByUuid(uuid:string): Observable<TimeseriesResponse> {
    return this.http.get<{response: TimeseriesResponse}>(`${environment.apiUrl}/api/timeseries/${uuid}`).pipe(
      map(data => {
        return data.response	;
      })
    )
  }
}
