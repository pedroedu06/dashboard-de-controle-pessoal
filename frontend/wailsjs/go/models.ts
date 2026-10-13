export namespace main {
	
	export class DBStatus {
	    connected: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DBStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.error = source["error"];
	    }
	}

}

