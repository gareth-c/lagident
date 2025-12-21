import { Target } from "./targets/targets.interfaces";

export interface TimeseriesResponse {
    Target: Target,
    Latencies: Latency[],
    Losses: Loss[]
}

export interface Latency {
    target_uuid: string,
    timestamp: number, //unix timestamp
    latency: number // latecny value in ms
}

export interface Loss {
    target_uuid: string,
    timestamp: number, //unix timestamp
}