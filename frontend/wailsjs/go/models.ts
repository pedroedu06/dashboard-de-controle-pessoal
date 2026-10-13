export namespace main {
	
	export class DayDTO {
	    date: string;
	    tasks: TaskDTO[];
	
	    static createFrom(source: any = {}) {
	        return new DayDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.tasks = this.convertValues(source["tasks"], TaskDTO);
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
	export class TaskDTO {
	    id: number;
	    title: string;
	    urgency: string;
	    startsAt: string;
	    endsAt: string;
	    done: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TaskDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.urgency = source["urgency"];
	        this.startsAt = source["startsAt"];
	        this.endsAt = source["endsAt"];
	        this.done = source["done"];
	    }
	}
	export class AgendaDTO {
	    overdue: TaskDTO[];
	    days: DayDTO[];
	
	    static createFrom(source: any = {}) {
	        return new AgendaDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overdue = this.convertValues(source["overdue"], TaskDTO);
	        this.days = this.convertValues(source["days"], DayDTO);
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
	
	
	export class TaskInputDTO {
	    title: string;
	    urgency: string;
	    startsAt: string;
	    endsAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskInputDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.urgency = source["urgency"];
	        this.startsAt = source["startsAt"];
	        this.endsAt = source["endsAt"];
	    }
	}

}

