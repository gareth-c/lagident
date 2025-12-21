export interface Target {
    uuid: string,
    name: string,
    address: string
}

export interface Statistics {
    target_uuid: string
    state: string
    sent: number
    recv: number
    last: number
    loss: number
    sum: number
    max: number
    min: {
        Float64: number,
        Valid: boolean
    }
    avg15m: number
    avg6h: number
    avg24h: number
    timestamp: number
}


export interface TargetWithStatistics {
    Target: Target
    Statistics?: Statistics
}