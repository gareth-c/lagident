import { inject, Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Target, TargetWithStatistics } from "./targets.interfaces";
import { catchError, map, Observable, of } from "rxjs";
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class TargetsService {


  private readonly http = inject(HttpClient);

  public getTargets(): Observable<Target[]> {
    return this.http.get<Target[]>(`${environment.apiUrl}/api/targets`).pipe(
      map(data => {
        return data;
      })
    )
  }


  public getTargetsWithStatistics(): Observable<TargetWithStatistics[]> {
    return this.http.get<{targets: TargetWithStatistics[]}>(`${environment.apiUrl}/api/statistics`).pipe(
      map(data => {
        return data.targets;
      })
    )
  }


  addTarget(target: Target): Observable<any> {
    return this.http.post<any>(`${environment.apiUrl}/api/targets/add`, target);
  }

  deleteTarget(uuid:string): Observable<any> {
    return this.http.delete<any>(`${environment.apiUrl}/api/targets/${uuid}`);
  }
}
