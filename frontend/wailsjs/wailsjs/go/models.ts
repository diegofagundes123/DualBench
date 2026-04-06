export namespace main {
	
	export class DriveSummary {
	    path: string;
	    writeMBps: number;
	    readMBps: number;
	    writeBytes: number;
	    readBytes: number;
	    durationMs: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DriveSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.writeMBps = source["writeMBps"];
	        this.readMBps = source["readMBps"];
	        this.writeBytes = source["writeBytes"];
	        this.readBytes = source["readBytes"];
	        this.durationMs = source["durationMs"];
	        this.error = source["error"];
	    }
	}
	export class BenchmarkSummary {
	    drive1: DriveSummary;
	    drive2: DriveSummary;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.drive1 = this.convertValues(source["drive1"], DriveSummary);
	        this.drive2 = this.convertValues(source["drive2"], DriveSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

